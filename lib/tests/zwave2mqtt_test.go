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
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/configuration"
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/model"
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/zwave2mqtt"
	"log"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestGetNodes(t *testing.T) {
	t.Skip("manual test with connection to real zwave2mqtt broker")
	config, err := configuration.Load("./resources/config.json")
	if err != nil {
		t.Error(err)
		return
	}
	config.ZwaveMqttClientId = config.ZwaveMqttClientId + strconv.Itoa(rand.Int())
	client, err := zwave2mqtt.New(config, context.Background())
	if err != nil {
		t.Error(err)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	client.SetDeviceInfoListener(func(nodes []model.DeviceInfo, huskIds []int64, _ bool, _ bool) {
		temp, err := json.Marshal(nodes)
		log.Println(err, string(temp))
		cancel()
	})
	err = client.RequestDeviceInfoUpdate()
	if err != nil {
		t.Error(err)
		return
	}
	<-ctx.Done()
	if err := ctx.Err(); err != nil && err != context.Canceled {
		t.Error(err)
	}
}

func TestNodesAvailableEvent(t *testing.T) {
	t.Skip("expects manually thrown available events")
	config, err := configuration.Load("./resources/config.json")
	if err != nil {
		t.Error(err)
		return
	}
	config.ZwaveMqttClientId = config.ZwaveMqttClientId + strconv.Itoa(rand.Int())
	client, err := zwave2mqtt.New(config, context.Background())
	if err != nil {
		t.Error(err)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	client.SetDeviceInfoListener(func(nodes []model.DeviceInfo, huskIds []int64, _ bool, _ bool) {
		temp, err := json.Marshal(nodes)
		log.Println(err, string(temp))
		cancel()
	})
	<-ctx.Done()
	time.Sleep(1 * time.Second)
}

func TestValueEvents(t *testing.T) {
	t.Skip("expects manually thrown available events and manual test stop")
	config, err := configuration.Load("./resources/config.json")
	if err != nil {
		t.Error(err)
		return
	}
	config.ZwaveMqttClientId = config.ZwaveMqttClientId + strconv.Itoa(rand.Int())
	client, err := zwave2mqtt.New(config, context.Background())
	if err != nil {
		t.Error(err)
		return
	}
	ctx := context.Background()
	client.SetValueEventListener(func(value model.Value) {
		temp, err := json.Marshal(value)
		log.Println(err, string(temp))
	})
	<-ctx.Done()
}

func TestSetValue(t *testing.T) {
	t.Skip("manual test with connection to real zwave2mqtt broker")
	config, err := configuration.Load("./resources/config.json")
	if err != nil {
		t.Error(err)
		return
	}
	config.ZwaveMqttClientId = config.ZwaveMqttClientId + strconv.Itoa(rand.Int())
	client, err := zwave2mqtt.New(config, context.Background())
	if err != nil {
		t.Error(err)
		return
	}
	err = client.SetValueByValueId("5-67-1-1", 18)
	if err != nil {
		t.Error(err)
		return
	}
}
