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
	"encoding/json"
	"errors"
	paho "github.com/eclipse/paho.mqtt.golang"
	"log"
)

func (this *Client) startValueEventListener() error {
	if !this.mqtt.IsConnected() {
		log.Println("WARNING: mqtt client not connected")
		return errors.New("mqtt client not connected")
	}

	log.Println("subscribe:", this.valueEventTopic)
	token := this.mqtt.Subscribe(this.valueEventTopic, 2, func(client paho.Client, message paho.Message) {
		this.handleValueEventMessage(message.Topic(), message.Payload())
	})
	if token.Wait() && token.Error() != nil {
		log.Println("Error on Subscribe: ", this.valueEventTopic, token.Error())
		this.ForwardError("Error on Subscribe: " + token.Error().Error())
		return token.Error()
	}

	return nil
}

func (this *Client) handleValueEventMessage(topic string, payload []byte) {
	if this.valueEventListener != nil {
		value := NodeValue{}
		err := json.Unmarshal(payload, &value)
		if err != nil || !validValueEvent(value) {
			//is not value event
			return
		}
		this.valueEventListener(transformValue(value))
	}
}

func validValueEvent(value NodeValue) bool {
	return value.Id != "" &&
		value.Value != nil &&
		value.NodeId != 0 &&
		value.NodeId != 1 &&
		value.LastUpdate != 0
}
