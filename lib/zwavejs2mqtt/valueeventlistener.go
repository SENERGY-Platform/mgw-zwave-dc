package zwavejs2mqtt

import (
	"encoding/json"
	"errors"
	paho "github.com/eclipse/paho.mqtt.golang"
	"log"
)

func (this *Client) startValueEventListener() error {
	if this.networkEventsTopic == "" || this.networkEventsTopic == "-" {
		log.Println("WARNING: no zwave network event topic configured --> no event handling")
		return nil
	}
	if !this.mqtt.IsConnected() {
		log.Println("WARNING: mqtt client not connected")
		return errors.New("mqtt client not connected")
	}

	token := this.mqtt.Subscribe(this.networkEventsTopic+NodeValueEventTopic, 2, func(client paho.Client, message paho.Message) {
		if this.valueEventListener != nil {
			node := NodeInfo{}
			err := json.Unmarshal(message.Payload(), &node)
			if err != nil {
				//is not expected info
				if this.debug {
					log.Println("DEBUG:", this.networkEventsTopic+NodeValueEventTopic, string(message.Payload()))
				}
				return
			}
			for _, value := range node.Values {
				this.valueEventListener(transformValue(value))
			}
		}
	})
	if token.Wait() && token.Error() != nil {
		log.Println("Error on Subscribe: ", this.networkEventsTopic+NodeValueEventTopic, token.Error())
		this.ForwardError("Error on Subscribe: " + token.Error().Error())
		return token.Error()
	}
	return nil
}

func (this *Client) isValueEvent(msg paho.Message) bool {
	temp := map[string]interface{}{}
	err := json.Unmarshal(msg.Payload(), &temp)
	if err != nil {
		return false
	}
	_, ok := temp["value"]
	if !ok {
		return false
	}
	_, ok = temp["value_id"]
	if !ok {
		return false
	}
	return true
}
