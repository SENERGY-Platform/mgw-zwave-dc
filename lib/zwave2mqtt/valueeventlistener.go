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
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/model"
	paho "github.com/eclipse/paho.mqtt.golang"
)

func (this *Client) startValueEventListener() error {
	if !this.mqtt.IsConnected() {
		slog.Warn("mqtt client not connected")
		return errors.New("mqtt client not connected")
	}

	token := this.mqtt.Subscribe(this.valueEventTopic, 2, func(client paho.Client, message paho.Message) {
		if this.valueEventListener != nil {
			if !this.isValueEvent(message) {
				slog.Debug("is not value event", "topic", message.Topic(), "payload", string(message.Payload()))
				return
			}
			slog.Debug("value event", "topic", message.Topic(), "payload", string(message.Payload()))
			result := model.Value{}
			err := json.Unmarshal(message.Payload(), &result)
			if err != nil {
				slog.Error("unable to unmarshal getNodes result", "error", err)
				this.ForwardError("unable to unmarshal getNodes result: " + err.Error())
				return
			}
			this.valueEventListener(result)
		}
	})
	if token.Wait() && token.Error() != nil {
		slog.Error("Error on Subscribe", "topic", this.valueEventTopic, "error", token.Error())
		this.ForwardError("Error on Subscribe: " + token.Error().Error())
		return token.Error()
	}
	return nil
}

func (this *Client) isValueEvent(msg paho.Message) bool {
	temp := map[string]interface{}{}
	err := json.Unmarshal(msg.Payload(), &temp)
	if err != nil {
		return false
	}
	_, ok := temp["value"]
	if !ok {
		return false
	}
	_, ok = temp["value_id"]
	if !ok {
		return false
	}
	return true
}
