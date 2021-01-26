package mgw

import paho "github.com/eclipse/paho.mqtt.golang"

type State string

const Online State = "online"
const Offline State = "offline"

// {"name": "airSensor", "state": "online", "device_type": "urn:infai:ses:device-type:a8cbd322-9d8c-4f4c-afec-ae4b7986b6ed"}
type DeviceInfo struct {
	Name       string `json:"name"`
	State      State  `json:"state"`
	DeviceType string `json:"device_type"`
}

// {"method": "set", "device_id": "x2gHd2fwdxjUR_lmee3bLw-5ecf7fb0bf6d", "data": {"name": "airSensor", "state": "online", "device_type": "urn:infai:ses:device-type:a8cbd322-9d8c-4f4c-afec-ae4b7986b6ed"}}
type DeviceInfoUpdate struct {
	Method   string     `json:"method"`
	DeviceId string     `json:"device_id"`
	Data     DeviceInfo `json:"data"`
}

// {"command_id": "senergy-connector-client-connector-b8b6f84c-6be2-4b87-aca2-a4696bf069ae", "data": ""}
type Command struct {
	CommandId string `json:"command_id"`
	Data      string `json:"data"`
}

type DeviceCommandHandler func(deviceId string, serviceId string, command Command)

type Subscription struct {
	Topic   string
	Handler paho.MessageHandler
}
