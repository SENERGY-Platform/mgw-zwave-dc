package tests

import (
	"context"
	"encoding/json"
	"log"
	"testing"
	"time"
	"zwave2mqtt-connector/lib/configuration"
	"zwave2mqtt-connector/lib/zwave2mqtt"
)

func TestGetNodes(t *testing.T) {
	config, err := configuration.Load("./resources/config.json")
	if err != nil {
		t.Error(err)
		return
	}
	client, err := zwave2mqtt.New(config)
	if err != nil {
		t.Error(err)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	client.SetGetDeviceInfoListener(func(nodes []zwave2mqtt.DeviceInfo) {
		temp, err := json.Marshal(nodes)
		log.Println(err, string(temp))
		cancel()
	})
	err = client.LoadNodes()
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
	client, err := zwave2mqtt.New(config)
	if err != nil {
		t.Error(err)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	client.SetGetDeviceInfoListener(func(nodes []zwave2mqtt.DeviceInfo) {
		temp, err := json.Marshal(nodes)
		log.Println(err, string(temp))
		cancel()
	})
	<-ctx.Done()
	time.Sleep(1 * time.Second)
}

func TestSetValue(t *testing.T) {
	config, err := configuration.Load("./resources/config.json")
	if err != nil {
		t.Error(err)
		return
	}
	client, err := zwave2mqtt.New(config)
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
