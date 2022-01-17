package mgw

import "log"

func (this *Client) SendClientError(message string) {
	if !this.mqtt.IsConnected() {
		log.Println("WARNING: mqtt client not connected")
		return
	}
	payload := this.connectorId + ": " + message
	topic := "error/client"
	if this.debug {
		log.Println("DEBUG: publish ", topic, payload)
	}
	token := this.mqtt.Publish(topic, 2, false, payload)
	if token.Wait() && token.Error() != nil {
		log.Println("Error on Client.Publish(): ", token.Error())
	}
	return
}

func (this *Client) SendDeviceError(localDeviceId string, message string) {
	if !this.mqtt.IsConnected() {
		log.Println("WARNING: mqtt client not connected")
		return
	}
	payload := this.connectorId + ": " + message
	topic := "error/device/" + localDeviceId
	if this.debug {
		log.Println("DEBUG: publish ", topic, payload)
	}
	token := this.mqtt.Publish(topic, 2, false, payload)
	if token.Wait() && token.Error() != nil {
		log.Println("Error on Client.Publish(): ", token.Error())
	}
	return
}

func (this *Client) SendCommandError(correlationId string, message string) {
	if !this.mqtt.IsConnected() {
		log.Println("WARNING: mqtt client not connected")
		return
	}
	payload := this.connectorId + ": " + message
	topic := "error/command/" + correlationId
	if this.debug {
		log.Println("DEBUG: publish ", topic, payload)
	}
	token := this.mqtt.Publish(topic, 2, false, message)
	if token.Wait() && token.Error() != nil {
		log.Println("Error on Client.Publish(): ", token.Error())
	}
	return
}
