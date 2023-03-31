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

package zwavejs2mqtt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
type DeviceStateListener = func(nodeId int64, online bool) error

const GetNodesCommandTopic = "/getNodes"
const NodeAvailableTopic = "/node_alive"

type Client struct {
	mqtt                paho.Client
	debug               bool
	apiTopic            string
	networkEventsTopic  string
	deviceStateTopic    string
	deviceInfoListener  DeviceInfoListener
	valueEventListener  ValueEventListener
	deviceStateListener DeviceStateListener
	forwardErrorMsg     func(msg string)
}

func New(config configuration.Config, ctx context.Context) (*Client, error) {
	log.Println("start zwavejs2mqtt client")
	client := &Client{
		deviceStateTopic:   config.ZwaveMqttDeviceStateTopic,
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

func (this *Client) SetDeviceStatusListener(listener func(nodeId int64, online bool) error) {
	this.deviceStateListener = listener
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
	err = this.startDeviceStateListener()
	if err != nil {
		return err
	}
	return nil
}

func (this *Client) RequestDeviceInfoUpdate() error {
	return this.SendZwayCommand(GetNodesCommandTopic, []interface{}{})
}

type ValueID struct {
	NodeId       int64       `json:"nodeId"`
	CommandClass int64       `json:"commandClass"`
	Endpoint     int64       `json:"endpoint"`
	Property     interface{} `json:"property"`
	PropertyKey  interface{} `json:"propertyKey,omitempty"`
}

// 2-38-0-targetValue
// <nodeId>/<commandClass>/<endpoint>/<property>/<propertyKey?>
func parseValueId(id string) (valueId ValueID, err error) {
	for i, part := range strings.Split(id, "-") {
		switch i {
		case 0:
			valueId.NodeId, err = strconv.ParseInt(part, 10, 64)
			if err != nil {
				return valueId, fmt.Errorf("unable to format value id %v: %w", id, err)
			}
		case 1:
			valueId.CommandClass, err = strconv.ParseInt(part, 10, 64)
			if err != nil {
				return valueId, fmt.Errorf("unable to format value id %v: %w", id, err)
			}
		case 2:
			valueId.Endpoint, err = strconv.ParseInt(part, 10, 64)
			if err != nil {
				return valueId, fmt.Errorf("unable to format value id %v: %w", id, err)
			}
		case 3:
			valueId.Property = part
		case 4:
			valueId.PropertyKey, err = strconv.ParseInt(part, 10, 64)
			if err != nil {
				err = nil
				valueId.PropertyKey = part
			}
		}
	}
	return valueId, nil
}

func (this *Client) SetValueByValueId(valueId string, value interface{}) error {
	valueIdObject, err := parseValueId(valueId)
	if err != nil {
		return err
	}
	args := []interface{}{valueIdObject, value}
	return this.SendZwayCommand("/writeValue", args)
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
