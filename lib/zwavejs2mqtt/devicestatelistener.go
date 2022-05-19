package zwavejs2mqtt

import (
	"encoding/json"
	"errors"
	paho "github.com/eclipse/paho.mqtt.golang"
	"log"
	"strings"
)

func (this *Client) startDeviceStateListener() error {
	if !this.mqtt.IsConnected() {
		log.Println("WARNING: mqtt client not connected")
		return errors.New("mqtt client not connected")
	}

	token := this.mqtt.Subscribe(this.deviceStateTopic, 2, func(client paho.Client, message paho.Message) {
		if this.deviceStateListener != nil && strings.HasSuffix(message.Topic(), "/status") {
			msg := DeviceStateMsg{}
			err := json.Unmarshal(message.Payload(), &msg)
			if err != nil {
				//is not device status
				return
			}
			if this.debug {
				log.Println("device state update: \n", string(message.Payload()))
			}
			if msg.NodeId > 1 {
				err = this.deviceStateListener(msg.NodeId, msg.Status == "Alive")
				if this.debug {
					log.Println("device state update result:", err)
				}
			}
		}
	})
	if token.Wait() && token.Error() != nil {
		log.Println("Error on Subscribe: ", this.deviceStateTopic, token.Error())
		this.ForwardError("Error on Subscribe: " + token.Error().Error())
		return token.Error()
	}
	return nil
}
