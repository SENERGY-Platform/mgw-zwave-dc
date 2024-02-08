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
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/model"
	"log"
	"strconv"
)

func (this *Connector) ValueEventListener(nodeValue model.Value) {
	deviceId, serviceId, value, err := this.parseNodeValueAsMgwEvent(nodeValue)
	if err != nil {
		log.Println("ERROR: unable to create device-id and service-id for node-value", err)
		this.mgwClient.SendClientError("unable to create device-id and service-id for node-value: " + err.Error())
		return
	}
	if this.eventShouldBeSend(deviceId) {
		this.saveValue(deviceId, serviceId, value)
		err = this.mgwClient.MarshalAndSendEvent(deviceId, serviceId, value)
		if err != nil {
			log.Println("ERROR: unable to send event", deviceId, serviceId, err)
			this.mgwClient.SendClientError("unable to send event: " + err.Error())
			return
		}
	} else if this.config.Debug {
		log.Printf("DEBUG: ignore event for %v because the device is not registered\n", deviceId)
	}
}

func (this *Connector) eventShouldBeSend(id string) bool {
	if this.eventsForUnregisteredDevices {
		return true
	}
	_, ok := this.deviceRegisterGet(id)
	return ok
}

func (this *Connector) sendStatistics(node model.DeviceInfo) {
	rawDeviceId := strconv.FormatInt(node.NodeId, 10)
	deviceId := this.addDeviceIdPrefix(rawDeviceId)
	err := this.mgwClient.MarshalAndSendEvent(deviceId, "statistics", node.Statistics)
	if err != nil {
		log.Println("ERROR: unable to send event", deviceId, "statistics", err)
		this.mgwClient.SendClientError("unable to send event: " + err.Error())
		return
	}
}
