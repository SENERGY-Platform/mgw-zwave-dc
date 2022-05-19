package connector

import (
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/model"
	"log"
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
	}
}

func (this *Connector) eventShouldBeSend(id string) bool {
	if this.eventsForUnregisteredDevices {
		return true
	}
	_, ok := this.deviceRegisterGet(id)
	return ok
}
