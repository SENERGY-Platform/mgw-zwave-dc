package mgw

import (
	"encoding/json"
	"errors"
	"log"
)

func (this *Client) MarshalAndSendEvent(deviceId string, serviceId string, value interface{}) error {
	msg, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return this.SendEvent(deviceId, serviceId, msg)
}

func (this *Client) SendEvent(deviceId string, serviceId string, msg []byte) error {
	if !this.mqtt.IsConnected() {
		log.Println("WARNING: mqtt client not connected")
		return errors.New("mqtt client not connected")
	}
	topic := "event/" + deviceId + "/" + serviceId
	if this.debug {
		log.Println("DEBUG: publish ", topic, string(msg))
	}
	token := this.mqtt.Publish(topic, 2, false, string(msg))
	if token.Wait() && token.Error() != nil {
		log.Println("Error on Client.Publish(): ", token.Error())
		return token.Error()
	}
	return nil
}
