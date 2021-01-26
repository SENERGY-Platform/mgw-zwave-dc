package zwave2mqtt

import (
	"encoding/json"
	"errors"
	paho "github.com/eclipse/paho.mqtt.golang"
	"log"
)

func (this *Client) startValueEventListener() error {
	if !this.mqtt.IsConnected() {
		log.Println("WARNING: mqtt client not connected")
		return errors.New("mqtt client not connected")
	}

	token := this.mqtt.Subscribe(this.valueEventTopic, 2, func(client paho.Client, message paho.Message) {
		if this.valueEventListener != nil {
			if !this.isValueEvent(message) {
				if this.debug {
					log.Println("is not value event: \n", string(message.Payload()))
				}
				return
			}
			if this.debug {
				log.Println("value event: \n", string(message.Payload()))
			}
			result := NodeValue{}
			err := json.Unmarshal(message.Payload(), &result)
			if err != nil {
				log.Println("ERROR: unable to unmarshal getNodes result", err)
				return
			}
			this.valueEventListener(result)
		}
	})
	if token.Wait() && token.Error() != nil {
		log.Println("Error on Subscribe: ", this.apiTopic+GetNodesCommandTopic, token.Error())
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
