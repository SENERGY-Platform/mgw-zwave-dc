package tests

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"strconv"
	"testing"
	"time"
	"zwave2mqtt-connector/lib/configuration"
	"zwave2mqtt-connector/lib/zwave2mqtt"
)

func TestGetNodes(t *testing.T) {
	t.Skip("manual test with connection to real zwave2mqtt broker")
	config, err := configuration.Load("./resources/config.json")
	if err != nil {
		t.Error(err)
		return
	}
	config.ZwaveMqttClientId = config.ZwaveMqttClientId + strconv.Itoa(rand.Int())
	client, err := zwave2mqtt.New(config, context.Background())
	if err != nil {
		t.Error(err)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	client.SetDeviceInfoListener(func(nodes []zwave2mqtt.DeviceInfo, _ bool, _ bool) {
		temp, err := json.Marshal(nodes)
		log.Println(err, string(temp))
		cancel()
	})
	err = client.RequestDeviceInfoUpdate()
	if err != nil {
		t.Error(err)
		return
	}
	<-ctx.Done()
	if err := ctx.Err(); err != nil && err != context.Canceled {
		t.Error(err)
	}
}

func TestNodesAvailableEvent(t *testing.T) {
	t.Skip("expects manually thrown available events")
	config, err := configuration.Load("./resources/config.json")
	if err != nil {
		t.Error(err)
		return
	}
	config.ZwaveMqttClientId = config.ZwaveMqttClientId + strconv.Itoa(rand.Int())
	client, err := zwave2mqtt.New(config, context.Background())
	if err != nil {
		t.Error(err)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	client.SetDeviceInfoListener(func(nodes []zwave2mqtt.DeviceInfo, _ bool, _ bool) {
		temp, err := json.Marshal(nodes)
		log.Println(err, string(temp))
		cancel()
	})
	<-ctx.Done()
	time.Sleep(1 * time.Second)
}

func TestValueEvents(t *testing.T) {
	t.Skip("expects manually thrown available events and manual test stop")
	config, err := configuration.Load("./resources/config.json")
	if err != nil {
		t.Error(err)
		return
	}
	config.ZwaveMqttClientId = config.ZwaveMqttClientId + strconv.Itoa(rand.Int())
	client, err := zwave2mqtt.New(config, context.Background())
	if err != nil {
		t.Error(err)
		return
	}
	ctx := context.Background()
	client.SetValueEventListener(func(value zwave2mqtt.NodeValue) {
		temp, err := json.Marshal(value)
		log.Println(err, string(temp))
	})
	<-ctx.Done()
}

func TestSetValue(t *testing.T) {
	t.Skip("manual test with connection to real zwave2mqtt broker")
	config, err := configuration.Load("./resources/config.json")
	if err != nil {
		t.Error(err)
		return
	}
	config.ZwaveMqttClientId = config.ZwaveMqttClientId + strconv.Itoa(rand.Int())
	client, err := zwave2mqtt.New(config, context.Background())
	if err != nil {
		t.Error(err)
		return
	}
	err = client.SetValueByValueId("5-67-1-1", 18)
	if err != nil {
		t.Error(err)
		return
	}
}
