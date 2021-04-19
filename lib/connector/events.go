package connector

import (
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/zwave2mqtt"
	"log"
)

func (this *Connector) ValueEventListener(nodeValue zwave2mqtt.NodeValue) {
	deviceId, serviceId, value, err := this.parseNodeValueAsMgwEvent(nodeValue)
	if err != nil {
		log.Println("ERROR: unable to create device-id and service-id for node-value", err)
		return
	}
	this.saveValue(deviceId, serviceId, value)
	err = this.mgwClient.SendEvent(deviceId, serviceId, value)
	if err != nil {
		log.Println("ERROR: unable to send event", deviceId, serviceId, err)
		return
	}
}
