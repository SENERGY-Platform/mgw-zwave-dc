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
	"fmt"
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/mgw"
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/model"
	"log"
	"runtime/debug"
)

func (this *Connector) NotifyRefresh() {
	err := this.z2mClient.RequestDeviceInfoUpdate()
	if err != nil {
		log.Println("ERROR:", err)
		this.mgwClient.SendClientError(err.Error())
		debug.PrintStack()
	}
	if this.updateTicker != nil && this.updateTickerDuration != 0 {
		this.updateTicker.Reset(this.updateTickerDuration)
	}
}

func (this *Connector) SetDeviceState(nodeId int64, online bool) error {
	deviceId := this.nodeIdToDeviceId(nodeId)
	info, ok := this.deviceRegisterGet(deviceId)
	if !ok {
		return fmt.Errorf("unknown device %v", nodeId)
	}
	info.State = mgw.Offline
	if online {
		info.State = mgw.Online
	}
	this.deviceRegisterSet(deviceId, info)
	return this.mgwClient.SetDevice(deviceId, info)
}

func (this *Connector) DeviceInfoListener(nodes []model.DeviceInfo, huskIds []int64, withValues bool, allKnownDevices bool) {
	deviceInfos := map[string]mgw.DeviceInfo{}
	for _, node := range nodes {
		id, info, err := this.nodeToDeviceInfo(node)
		if err != nil {
			log.Println("WARNING: unable to create device info for node", err)
			continue
		}
		err = this.registerDevice(id, info)
		if err != nil {
			log.Println("ERROR: unable to register device", err)
			this.mgwClient.SendClientError("unable to register device: " + err.Error())
			return
		}
		deviceInfos[id] = info
		if withValues {
			for _, value := range node.Values {
				this.ValueEventListener(value)
			}
		}
		this.sendStatistics(node)
	}
	isSetToOfflineOrDeleted := map[string]bool{}
	if allKnownDevices {
		isSetToOfflineOrDeleted = this.unregisterMissingDevices(deviceInfos)
	}
	if this.husksShouldBeDeleted {
		this.sendDeleteForHusks(huskIds, isSetToOfflineOrDeleted)
	}
}

func (this *Connector) registerDevice(id string, info mgw.DeviceInfo) (err error) {
	err = this.mgwClient.SetDevice(id, info)
	if err != nil {
		log.Println("ERROR: unable to send device info to mgw", err)
		return err
	}
	err = this.mgwClient.ListenToDeviceCommands(id, this.CommandHandler)
	if err != nil {
		log.Println("ERROR: unable to subscribe to device commands", err)
		return err
	}
	this.deviceRegisterSet(id, info)
	return nil
}

func (this *Connector) deviceRegisterSet(id string, info mgw.DeviceInfo) {
	this.deviceRegisterMux.Lock()
	defer this.deviceRegisterMux.Unlock()
	this.deviceRegister[id] = info
}

func (this *Connector) deviceRegisterRemove(id string) {
	this.deviceRegisterMux.Lock()
	defer this.deviceRegisterMux.Unlock()
	delete(this.deviceRegister, id)
}

func (this *Connector) deviceRegisterGet(id string) (info mgw.DeviceInfo, ok bool) {
	this.deviceRegisterMux.Lock()
	defer this.deviceRegisterMux.Unlock()
	info, ok = this.deviceRegister[id]
	return
}

func (this *Connector) deviceRegisterGetIds() (ids []string) {
	this.deviceRegisterMux.Lock()
	defer this.deviceRegisterMux.Unlock()
	for id, _ := range this.deviceRegister {
		ids = append(ids, id)
	}
	return
}

func (this *Connector) deviceRegisterGetAll() (result map[string]mgw.DeviceInfo) {
	result = map[string]mgw.DeviceInfo{}
	this.deviceRegisterMux.Lock()
	defer this.deviceRegisterMux.Unlock()
	for key, value := range this.deviceRegister {
		result[key] = value
	}
	return
}

func (this *Connector) unregisterMissingDevices(infos map[string]mgw.DeviceInfo) (handled map[string]bool) {
	handled = map[string]bool{}
	for id, info := range this.deviceRegisterGetAll() {
		_, found := infos[id]
		if !found {
			info.State = mgw.Offline
			if this.deleteMissingDevices {
				log.Println("WARNING: remove missing device:", id)
				err := this.mgwClient.RemoveDevice(id)
				if err != nil {
					log.Println("ERROR: unable to send device info (delete) to mgw", err)
					this.mgwClient.SendClientError("unable to send device info (delete) to mgw: " + err.Error())
					return
				}
			} else {
				log.Println("WARNING: set missing device offline:", id)
				err := this.mgwClient.SetDevice(id, info)
				if err != nil {
					log.Println("ERROR: unable to send device info (offline) to mgw", err)
					this.mgwClient.SendClientError("unable to send device info (offline) to mgw: " + err.Error())
					return
				}
			}

			err := this.mgwClient.StopListenToDeviceCommands(id)
			if err != nil {
				log.Println("WARNING: unable to stop listening to device commands", err)
				this.mgwClient.SendClientError("unable to stop listening to device commands: " + err.Error())
			}
			this.deviceRegisterRemove(id)
			handled[id] = true
		}
	}
	return
}

func (this *Connector) sendDeleteForHusks(huskIds []int64, alreadyHandled map[string]bool) {
	for _, huskId := range huskIds {
		deviceId := this.nodeIdToDeviceId(huskId)
		if !alreadyHandled[deviceId] {
			err := this.mgwClient.RemoveDevice(deviceId)
			if err != nil {
				log.Println("ERROR: unable to delete husk in mgw", err)
				return
			}
		}
	}
}
