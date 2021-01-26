package tests

import (
	"math/rand"
	"strconv"
	"testing"
	"zwave2mqtt-connector/lib/configuration"
	"zwave2mqtt-connector/lib/mgw"
)

func TestMgwSetDevice(t *testing.T) {
	t.Skip("TODO")
	config, err := configuration.Load("./resources/config.json")
	if err != nil {
		t.Error(err)
		return
	}
	config.MgwMqttClientId = config.MgwMqttClientId + strconv.Itoa(rand.Int())
	client, err := mgw.New(config, nil)
	if err != nil {
		t.Error(err)
		return
	}
	err = client.SetDevice("test-device-id", mgw.DeviceInfo{
		Name:       "test",
		State:      mgw.Online,
		DeviceType: "test-device-type-id",
	})
	if err != nil {
		t.Error(err)
		return
	}
}
