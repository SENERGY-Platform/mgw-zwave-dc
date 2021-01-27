package mgw

import (
	"encoding/json"
	"errors"
	"log"
)

func (this *Client) Respond(deviceId string, serviceId string, response Command) error {
	if !this.mqtt.IsConnected() {
		log.Println("WARNING: mqtt client not connected")
		return errors.New("mqtt client not connected")
	}
	topic := "response/" + deviceId + "/" + serviceId
	msg, err := json.Marshal(response)
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
