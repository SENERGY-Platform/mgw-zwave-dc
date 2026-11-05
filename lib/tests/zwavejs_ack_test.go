/*
 * Copyright (c) 2023 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tests

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/SENERGY-Platform/mgw-zwave-dc/lib"
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/configuration"
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/mgw"
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/tests/docker"
	paho "github.com/eclipse/paho.mqtt.golang"
)

const ackTestApiTopic = "zwave/_CLIENTS/ZWAVE_GATEWAY-Zwavejs2Mqtt/api"

// fakeZwaveJsUi answers writeValue requests like zwave-js-ui: the response goes to the request topic
// without "/set" and echoes the request payload as origin. answer == nil or a nil result means no response.
func fakeZwaveJsUi(t *testing.T, client paho.Client, answer func(origin map[string]interface{}) map[string]interface{}) (requests chan map[string]interface{}) {
	requests = make(chan map[string]interface{}, 10)
	token := client.Subscribe(ackTestApiTopic+"/writeValue/set", 2, func(_ paho.Client, message paho.Message) {
		origin := map[string]interface{}{}
		if err := json.Unmarshal(message.Payload(), &origin); err != nil {
			t.Error(err)
			return
		}
		requests <- origin
		if answer == nil {
			return
		}
		resp := answer(origin)
		if resp == nil {
			return
		}
		resp["origin"] = origin
		resp["args"] = origin["args"]
		msg, _ := json.Marshal(resp)
		client.Publish(ackTestApiTopic+"/writeValue", 2, false, string(msg))
	})
	if token.Wait() && token.Error() != nil {
		t.Fatal(token.Error())
	}
	return requests
}

func TestZwavejsCommandAck(t *testing.T) {
	config, err := configuration.Load("../../config.json")
	if err != nil {
		t.Fatal(err)
	}
	config.ZwaveController = "zwavejs2mqtt"
	config.ZwaveMqttApiTopic = ackTestApiTopic
	config.ZwaveMqttDeviceStateTopic = "zwave/#"
	config.ZwaveNetworkEventsTopic = "-"
	config.DeviceIdPrefix = "test-prefix"
	config.ConnectorId = "test-connector-id"
	config.UpdatePeriod = ""
	config.CommandAckTimeout = "1s"

	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mgwPort, _, err := docker.Mqtt(ctx, wg)
	if err != nil {
		t.Fatal(err)
	}
	config.MgwMqttBroker = "tcp://localhost:" + mgwPort
	config.MgwMqttClientId = "ack-connector-mgw"
	zwavePort, _, err := docker.Mqtt(ctx, wg)
	if err != nil {
		t.Fatal(err)
	}
	config.ZwaveMqttBroker = "tcp://localhost:" + zwavePort
	config.ZwaveMqttClientId = "ack-connector-zwave"

	c, err := lib.New(config, ctx)
	if err != nil {
		t.Fatal(err)
	}

	connect := func(id string, broker string) paho.Client {
		client := paho.NewClient(paho.NewClientOptions().SetClientID(id).AddBroker(broker))
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			t.Fatal(token.Error())
		}
		t.Cleanup(func() { client.Disconnect(0) })
		return client
	}
	mgwClient := connect("ack-test-mgw", config.MgwMqttBroker)
	zwaveClient := connect("ack-test-zwave", config.ZwaveMqttBroker)

	messages := make(chan string, 10)
	subscribe := func(topic string) {
		token := mgwClient.Subscribe(topic, 2, func(_ paho.Client, message paho.Message) {
			messages <- message.Topic() + " " + string(message.Payload())
		})
		if token.Wait() && token.Error() != nil {
			t.Fatal(token.Error())
		}
	}
	subscribe("response/test-prefix:3/67-1-1")
	subscribe("error/command/+")

	next := func(timeout time.Duration) (string, bool) {
		select {
		case m := <-messages:
			return m, true
		case <-time.After(timeout):
			return "", false
		}
	}

	var mode struct {
		sync.Mutex
		answer func(origin map[string]interface{}) map[string]interface{}
	}
	requests := fakeZwaveJsUi(t, zwaveClient, func(origin map[string]interface{}) map[string]interface{} {
		mode.Lock()
		defer mode.Unlock()
		if mode.answer == nil {
			return nil
		}
		return mode.answer(origin)
	})
	setAnswer := func(f func(origin map[string]interface{}) map[string]interface{}) {
		mode.Lock()
		defer mode.Unlock()
		mode.answer = f
	}
	// paho runs the on-connect handler (which subscribes to the response topic) asynchronously,
	// give it time before the first request
	time.Sleep(time.Second)

	send := func(commandId string) {
		c.CommandHandler("test-prefix:3", "67-1-1", mgw.Command{CommandId: commandId, Data: "42"})
	}

	t.Run("success", func(t *testing.T) {
		setAnswer(func(map[string]interface{}) map[string]interface{} {
			return map[string]interface{}{"success": true, "message": "Success zwave api call", "result": map[string]interface{}{"status": 255}}
		})
		go send("cmd-success")
		req := <-requests
		if req["requestId"] == nil || req["requestId"] == "" {
			t.Errorf("request carries no requestId: %v", req)
		}
		m, ok := next(5 * time.Second)
		if !ok || m != `response/test-prefix:3/67-1-1 {"command_id":"cmd-success","data":""}` {
			t.Errorf("unexpected message %q (received=%v)", m, ok)
		}
		if m, ok := next(200 * time.Millisecond); ok {
			t.Errorf("unexpected additional message %q", m)
		}
	})

	t.Run("failure status", func(t *testing.T) {
		setAnswer(func(map[string]interface{}) map[string]interface{} {
			return map[string]interface{}{"success": true, "result": map[string]interface{}{"status": 2, "message": "node dead"}}
		})
		go send("cmd-fail")
		<-requests
		m, ok := next(5 * time.Second)
		if !ok || !strings.HasPrefix(m, "error/command/cmd-fail ") || !strings.Contains(m, "status 2") || !strings.Contains(m, "node dead") {
			t.Errorf("unexpected message %q (received=%v)", m, ok)
		}
		if m, ok := next(200 * time.Millisecond); ok {
			t.Errorf("unexpected additional message %q", m)
		}
	})

	t.Run("success false", func(t *testing.T) {
		setAnswer(func(map[string]interface{}) map[string]interface{} {
			return map[string]interface{}{"success": false, "message": "Node 3 not found"}
		})
		go send("cmd-false")
		<-requests
		m, ok := next(5 * time.Second)
		if !ok || !strings.HasPrefix(m, "error/command/cmd-false ") || !strings.Contains(m, "Node 3 not found") {
			t.Errorf("unexpected message %q (received=%v)", m, ok)
		}
	})

	t.Run("no answer", func(t *testing.T) {
		setAnswer(nil)
		start := time.Now()
		go send("cmd-timeout")
		<-requests
		m, ok := next(5 * time.Second)
		if !ok || !strings.HasPrefix(m, "error/command/cmd-timeout ") || !strings.Contains(m, "no response") {
			t.Errorf("unexpected message %q (received=%v)", m, ok)
		}
		if d := time.Since(start); d < 900*time.Millisecond {
			t.Errorf("error after %v, before the timeout", d)
		}
	})

	t.Run("answer for another request id is ignored", func(t *testing.T) {
		setAnswer(func(origin map[string]interface{}) map[string]interface{} {
			origin["requestId"] = "someone-else"
			return map[string]interface{}{"success": true, "result": map[string]interface{}{"status": 255}}
		})
		go send("cmd-other")
		<-requests
		m, ok := next(5 * time.Second)
		if !ok || !strings.HasPrefix(m, "error/command/cmd-other ") || !strings.Contains(m, "no response") {
			t.Errorf("unexpected message %q (received=%v)", m, ok)
		}
	})
}

func TestZwavejsCommandAckDisabled(t *testing.T) {
	// with the wait disabled the command is answered right after sending, even if zwave-js-ui never answers
	config, err := configuration.Load("../../config.json")
	if err != nil {
		t.Fatal(err)
	}
	config.ZwaveController = "zwavejs2mqtt"
	config.ZwaveMqttApiTopic = ackTestApiTopic
	config.ZwaveMqttDeviceStateTopic = "zwave/#"
	config.ZwaveNetworkEventsTopic = "-"
	config.DeviceIdPrefix = "test-prefix"
	config.ConnectorId = "test-connector-id"
	config.UpdatePeriod = ""
	config.CommandAckTimeout = "-"

	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mgwPort, _, err := docker.Mqtt(ctx, wg)
	if err != nil {
		t.Fatal(err)
	}
	config.MgwMqttBroker = "tcp://localhost:" + mgwPort
	config.MgwMqttClientId = "ack-off-connector-mgw"
	zwavePort, _, err := docker.Mqtt(ctx, wg)
	if err != nil {
		t.Fatal(err)
	}
	config.ZwaveMqttBroker = "tcp://localhost:" + zwavePort
	config.ZwaveMqttClientId = "ack-off-connector-zwave"

	c, err := lib.New(config, ctx)
	if err != nil {
		t.Fatal(err)
	}
	mgwClient := paho.NewClient(paho.NewClientOptions().SetClientID("ack-off-test-mgw").AddBroker(config.MgwMqttBroker))
	if token := mgwClient.Connect(); token.Wait() && token.Error() != nil {
		t.Fatal(token.Error())
	}
	defer mgwClient.Disconnect(0)
	got := make(chan string, 1)
	token := mgwClient.Subscribe("response/test-prefix:3/67-1-1", 2, func(_ paho.Client, message paho.Message) {
		got <- string(message.Payload())
	})
	if token.Wait() && token.Error() != nil {
		t.Fatal(token.Error())
	}
	go c.CommandHandler("test-prefix:3", "67-1-1", mgw.Command{CommandId: "cmd-off", Data: "42"})
	select {
	case m := <-got:
		if m != `{"command_id":"cmd-off","data":""}` {
			t.Error(m)
		}
	case <-time.After(5 * time.Second):
		t.Error("no immediate response")
	}
}

func TestInvalidCommandAckTimeoutFailsStart(t *testing.T) {
	config, err := configuration.Load("../../config.json")
	if err != nil {
		t.Fatal(err)
	}
	config.CommandAckTimeout = "soon"
	if _, err := lib.New(config, context.Background()); err == nil {
		t.Error("expected error")
	}
}
