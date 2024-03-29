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

package zwave2mqtt

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/configuration"
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/model"
	paho "github.com/eclipse/paho.mqtt.golang"
	"log"
	"strconv"
	"strings"
	"time"
)

type DeviceInfoListener = func(nodes []model.DeviceInfo, huskIds []int64, withValues bool, allKnownDevices bool)
type ValueEventListener = func(value model.Value)

const GetNodesCommandTopic = "/getNodes"
const NodeAvailableTopic = "/node_available"

type Client struct {
	mqtt               paho.Client
	debug              bool
	valueEventTopic    string
	apiTopic           string
	networkEventsTopic string
	deviceInfoListener DeviceInfoListener
	valueEventListener ValueEventListener
	forwardErrorMsg    func(msg string)
}

func New(config configuration.Config, ctx context.Context) (*Client, error) {
	log.Println("start zwave2mqtt client")
	client := &Client{
		valueEventTopic:    config.ZvaveValueEventTopic,
		apiTopic:           config.ZwaveMqttApiTopic,
		networkEventsTopic: config.ZwaveNetworkEventsTopic,
		debug:              config.Debug,
	}
	options := paho.NewClientOptions().
		SetPassword(config.ZwaveMqttPw).
		SetUsername(config.ZwaveMqttUser).
		SetAutoReconnect(true).
		SetCleanSession(true).
		SetClientID(config.ZwaveMqttClientId).
		AddBroker(config.ZwaveMqttBroker).
		SetWriteTimeout(10 * time.Second).
		SetOrderMatters(false).
		SetConnectionLostHandler(func(_ paho.Client, err error) {
			log.Println("connection to zwave2mqtt broker lost")
		}).
		SetOnConnectHandler(func(_ paho.Client) {
			log.Println("connected to zwave2mqtt broker")
			err := client.startDefaultListener()
			if err != nil {
				log.Fatal("FATAL: ", err)
			}
		})

	client.mqtt = paho.NewClient(options)
	if token := client.mqtt.Connect(); token.Wait() && token.Error() != nil {
		log.Println("Error on MqttStart.Connect(): ", token.Error())
		return nil, token.Error()
	}

	go func() {
		<-ctx.Done()
		client.mqtt.Disconnect(0)
	}()

	return client, nil
}

func (this *Client) SetDeviceStatusListener(_ func(nodeId int64, online bool) error) {}

func (this *Client) ForwardError(msg string) {
	if this.forwardErrorMsg != nil {
		this.forwardErrorMsg(msg)
	}
}

func (this *Client) SetErrorForwardingFunc(f func(msg string)) {
	this.forwardErrorMsg = f
}

func (this *Client) SetDeviceInfoListener(listener DeviceInfoListener) {
	this.deviceInfoListener = listener
}

func (this *Client) SetValueEventListener(listener ValueEventListener) {
	this.valueEventListener = listener
}

func (this *Client) startDefaultListener() error {
	err := this.startNodeCommandListener()
	if err != nil {
		return err
	}
	err = this.startNodeEventListener()
	if err != nil {
		return err
	}
	err = this.startValueEventListener()
	if err != nil {
		return err
	}
	return nil
}

func (this *Client) RequestDeviceInfoUpdate() error {
	return this.SendZwayCommand(GetNodesCommandTopic, []interface{}{})
}

func (this *Client) SetValue(nodeId int64, classId int64, instanceId int64, index int64, value interface{}) error {
	return this.SendZwayCommand("/setValue", []interface{}{nodeId, classId, instanceId, index, value})
}

func (this *Client) SetValueByValueId(valueId string, value interface{}) error {
	args := []interface{}{}
	for _, v := range strings.Split(valueId, "-") {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return err
		}
		args = append(args, id)
	}
	args = append(args, value)
	return this.SendZwayCommand("/setValue", args)
}

func (this *Client) SendZwayCommand(command string, args []interface{}) error {
	if !this.mqtt.IsConnected() {
		log.Println("WARNING: mqtt client not connected")
		return errors.New("mqtt client not connected")
	}
	topic := this.apiTopic + command + "/set"
	payload := map[string]interface{}{
		"args": args,
	}
	msg, err := json.Marshal(payload)
	if this.debug {
		log.Println("DEBUG: publish ", topic, string(msg))
	}
	token := this.mqtt.Publish(topic, 2, false, string(msg))
	if token.Wait() && token.Error() != nil {
		log.Println("Error on Client.Publish(): ", token.Error())
		return token.Error()
	}
	return err
}
