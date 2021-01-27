package connector

import (
	"log"
	"zwave2mqtt-connector/lib/zwave2mqtt"
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
