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

	log.Println("subscribe:", this.deviceStateTopic)
	token := this.mqtt.Subscribe(this.deviceStateTopic, 2, func(client paho.Client, message paho.Message) {
		this.handleDeviceStateMessage(message.Topic(), message.Payload())
		this.handleValueEventMessage(message.Topic(), message.Payload())
	})
	if token.Wait() && token.Error() != nil {
		log.Println("Error on Subscribe: ", this.deviceStateTopic, token.Error())
		this.ForwardError("Error on Subscribe: " + token.Error().Error())
		return token.Error()
	}
	return nil
}

func (this *Client) handleValueEventMessage(topic string, payload []byte) {
	if this.valueEventListener != nil {
		value := NodeValue{}
		err := json.Unmarshal(payload, &value)
		if err != nil || !validValueEvent(value) {
			//is not value event
			return
		}
		this.valueEventListener(transformValue(value))
	}
}

func validValueEvent(value NodeValue) bool {
	return value.Id != "" &&
		value.Value != nil &&
		value.NodeId != 0 &&
		value.NodeId != 1 &&
		value.LastUpdate != 0
}

type DeviceState = string

const DEAD DeviceState = "Dead"
const UNKNOWN DeviceState = "Unknown"
const ALIVE DeviceState = "Alive"
const ASLEEP DeviceState = "Asleep"

func (this *Client) handleDeviceStateMessage(topic string, payload []byte) {
	if this.deviceStateListener != nil && strings.HasSuffix(topic, "/status") {
		msg := DeviceStateMsg{}
		err := json.Unmarshal(payload, &msg)
		if err != nil {
			//is not device status
			return
		}
		if msg.NodeId > 1 {
			if msg.Status == ALIVE || msg.Status == ASLEEP {
				log.Println("device state update: ", msg.NodeId, msg.Status)
				err = this.deviceStateListener(msg.NodeId, true)
			}
			if msg.Status == DEAD {
				log.Println("device state update: ", msg.NodeId, msg.Status)
				err = this.deviceStateListener(msg.NodeId, false)
			}
			if err != nil {
				log.Println("ERROR: unable to update device state:", err)
			}
		}
	}
}
