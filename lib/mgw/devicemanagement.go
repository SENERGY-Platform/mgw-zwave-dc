package mgw

import (
	"encoding/json"
	"errors"
	"log"
)

func (this *Client) SetDevice(deviceId string, info DeviceInfo) error {
	return this.SendDeviceUpdate(DeviceInfoUpdate{
		Method:   "set",
		DeviceId: deviceId,
		Data:     info,
	})
}

func (this *Client) SendDeviceUpdate(info DeviceInfoUpdate) error {
	if !this.mqtt.IsConnected() {
		log.Println("WARNING: mqtt client not connected")
		return errors.New("mqtt client not connected")
	}
	topic := DeviceManagerTopic + "/" + this.connectorId
	msg, err := json.Marshal(info)
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
