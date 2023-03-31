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

package connector

import (
	"encoding/json"
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/mgw"
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/model"
	"log"
)

// expects ids from mgw (with prefixes and suffixes)
func (this *Connector) CommandHandler(deviceId string, serviceId string, command mgw.Command) {
	if this.isGetServiceId(serviceId) {
		this.handleGetCommand(deviceId, serviceId, command)
	} else {
		this.handleSetCommand(deviceId, serviceId, command)
	}
}

// expects ids from mgw (with prefixes and suffixes)
func (this *Connector) handleGetCommand(deviceId string, serviceId string, command mgw.Command) {
	value, known := this.getValue(deviceId, serviceId)
	if known {
		temp, err := json.Marshal(value)
		if err != nil {
			log.Println("ERROR: unable to marshal saved value to send as response", deviceId, serviceId, err)
			this.mgwClient.SendCommandError(command.CommandId, "unable to marshal saved value to send as response: "+err.Error())
			return
		}
		command.Data = string(temp)
		err = this.mgwClient.Respond(deviceId, serviceId, command)
		if err != nil {
			log.Println("ERROR: unable to send response to mgw", err)
			this.mgwClient.SendCommandError(command.CommandId, "unable to send response to mgw: "+err.Error())
			return
		}
	} else {
		log.Println("WARNING: no value saved to send as response", deviceId, serviceId)
		this.mgwClient.SendCommandError(command.CommandId, "no value saved to send as response")
		return
	}
}

// expects ids from mgw (with prefixes and suffixes)
func (this *Connector) handleSetCommand(deviceId string, serviceId string, command mgw.Command) {
	valueId := this.removeDeviceIdPrefix(deviceId) + "-" + model.DecodeLocalId(serviceId)
	var value interface{}
	err := json.Unmarshal([]byte(command.Data), &value)
	if err != nil {
		log.Println("ERROR: unable to Unmarshal command data to z2m value\n    ", err)
		this.mgwClient.SendCommandError(command.CommandId, "unable to Unmarshal command data to z2m value: "+err.Error())
		return
	}
	err = this.z2mClient.SetValueByValueId(valueId, value)
	if err != nil {
		log.Println("ERROR: unable to send value to z2m\n    ", err)
		this.mgwClient.SendCommandError(command.CommandId, "unable to send value to z2m: "+err.Error())
		return
	}
	command.Data = ""
	err = this.mgwClient.Respond(deviceId, serviceId, command)
	if err != nil {
		log.Println("ERROR: unable to send response to mgw", err)
		this.mgwClient.SendCommandError(command.CommandId, "unable to send response to mgw: "+err.Error())
		return
	}
}
