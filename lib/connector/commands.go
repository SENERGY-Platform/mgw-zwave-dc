package connector

import (
	"encoding/json"
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/mgw"
	"log"
)

//expects ids from mgw (with prefixes and suffixes)
func (this *Connector) CommandHandler(deviceId string, serviceId string, command mgw.Command) {
	if this.isGetServiceId(serviceId) {
		this.handleGetCommand(deviceId, serviceId, command)
	} else {
		this.handleSetCommand(deviceId, serviceId, command)
	}
}

//expects ids from mgw (with prefixes and suffixes)
func (this *Connector) handleGetCommand(deviceId string, serviceId string, command mgw.Command) {
	value, known := this.getValue(deviceId, serviceId)
	if known {
		temp, err := json.Marshal(value)
		if err != nil {
			log.Println("ERROR: unable to marshal saved value to send as response", deviceId, serviceId, err)
			//TODO
			return
		}
		command.Data = string(temp)
		err = this.mgwClient.Respond(deviceId, serviceId, command)
		if err != nil {
			log.Println("ERROR: unable to send response to mgw", err)
			return
		}
	} else {
		log.Println("WARNING: no value saved to send as response", deviceId, serviceId)
		//TODO
		return
	}
}

//expects ids from mgw (with prefixes and suffixes)
func (this *Connector) handleSetCommand(deviceId string, serviceId string, command mgw.Command) {
	valueId := this.removeDeviceIdPrefix(deviceId) + "-" + serviceId
	var value interface{}
	err := json.Unmarshal([]byte(command.Data), &value)
	if err != nil {
		log.Println("ERROR: unable to Unmarshal command data to z2m value\n    ", err)
		return
	}
	err = this.z2mClient.SetValueByValueId(valueId, value)
	if err != nil {
		log.Println("ERROR: unable to send value to z2m\n    ", err)
		return
	}
	command.Data = ""
	err = this.mgwClient.Respond(deviceId, serviceId, command)
	if err != nil {
		log.Println("ERROR: unable to send response to mgw", err)
		return
	}
}
