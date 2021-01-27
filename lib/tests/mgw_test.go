package tests

import (
	"context"
	paho "github.com/eclipse/paho.mqtt.golang"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"
	"zwave2mqtt-connector/lib/configuration"
	"zwave2mqtt-connector/lib/mgw"
	"zwave2mqtt-connector/lib/tests/docker"
)

func TestMgwDeviceManagement(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	config, err := configuration.Load("./resources/config.json")
	if err != nil {
		t.Error(err)
		return
	}
	mqttPort, _, err := docker.Mqtt(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}
	config.MgwMqttBroker = "tcp://localhost:" + mqttPort
	config.MgwMqttUser = ""
	config.MgwMqttPw = ""
	config.MgwMqttClientId = "test-mgw-connector-" + strconv.Itoa(rand.Int())

	mqttclient := paho.NewClient(paho.NewClientOptions().
		SetAutoReconnect(true).
		SetCleanSession(false).
		SetClientID("test-connection-" + strconv.Itoa(rand.Int())).
		AddBroker(config.MgwMqttBroker))
	if token := mqttclient.Connect(); token.Wait() && token.Error() != nil {
		log.Println("Error on Mqtt.Connect(): ", token.Error())
		t.Error(err)
		return
	}
	defer mqttclient.Disconnect(0)

	deviceManagerMsgCount := 0
	log.Println("subscribe to " + "device-manager/device/" + config.ConnectorId)
	token := mqttclient.Subscribe("device-manager/device/"+config.ConnectorId, 2, func(_ paho.Client, message paho.Message) {
		deviceManagerMsgCount = deviceManagerMsgCount + 1
		expectedMessage := `{"method":"set","device_id":"test-device-id","data":{"name":"test","state":"online","device_type":"test-device-type-id"}}`
		if string(message.Payload()) != expectedMessage {
			t.Error(string(message.Payload()))
			return
		}
	})
	if token.Wait() && token.Error() != nil {
		log.Println("ERROR: device management subscription: ", token.Error())
		t.Error(err)
		return
	}

	deviceManagerLwCount := 0
	log.Println("subscribe to " + "device-manager/device/" + config.ConnectorId + "/lw")
	token = mqttclient.Subscribe("device-manager/device/"+config.ConnectorId+"/lw", 2, func(_ paho.Client, message paho.Message) {
		deviceManagerLwCount = deviceManagerLwCount + 1
	})
	if token.Wait() && token.Error() != nil {
		log.Println("ERROR: device management lwt subscription: ", token.Error())
		t.Error(err)
		return
	}

	refreshNotifyCount := 0
	clientCtx, stopClient := context.WithCancel(ctx)
	client, err := mgw.New(config, clientCtx, func() {
		log.Println("LOG: refresh notify")
		refreshNotifyCount = refreshNotifyCount + 1
	})
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

	token = mqttclient.Publish("device-manager/refresh", 2, false, "1")
	if token.Wait() && token.Error() != nil {
		log.Println("ERROR: device management refresh: ", token.Error())
		t.Error(err)
		return
	}

	time.Sleep(1 * time.Second)

	if deviceManagerMsgCount != 1 {
		t.Error(deviceManagerMsgCount)
		return
	}

	// on connect and on request
	if refreshNotifyCount != 2 {
		t.Error(refreshNotifyCount)
		return
	}

	stopClient()
	time.Sleep(1 * time.Second)
	if deviceManagerLwCount != 1 {
		t.Error(deviceManagerLwCount)
		return
	}
}

func TestMgwCommand(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	config, err := configuration.Load("./resources/config.json")
	if err != nil {
		t.Error(err)
		return
	}
	mqttPort, _, err := docker.Mqtt(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}
	config.MgwMqttBroker = "tcp://localhost:" + mqttPort
	config.MgwMqttUser = ""
	config.MgwMqttPw = ""
	config.MgwMqttClientId = "test-mgw-connector-" + strconv.Itoa(rand.Int())

	mqttclient := paho.NewClient(paho.NewClientOptions().
		SetAutoReconnect(true).
		SetCleanSession(false).
		SetClientID("test-connection-" + strconv.Itoa(rand.Int())).
		AddBroker(config.MgwMqttBroker))
	if token := mqttclient.Connect(); token.Wait() && token.Error() != nil {
		log.Println("Error on Mqtt.Connect(): ", token.Error())
		t.Error(err)
		return
	}
	defer mqttclient.Disconnect(0)

	client, err := mgw.New(config, ctx, nil)
	if err != nil {
		t.Error(err)
		return
	}

	commandCount := 0
	err = client.ListenToDeviceCommands("test-device-id", func(deviceId string, serviceId string, command mgw.Command) {
		commandCount = commandCount + 1
		if deviceId != "test-device-id" {
			t.Error(deviceId)
			return
		}
		if serviceId != "test-service-id" {
			t.Error(serviceId)
			return
		}
		if command.CommandId != "command-id" {
			t.Error(command.CommandId)
			return
		}
		if command.Data != "{\"power\": false}" {
			t.Error(command.Data)
			return
		}
		command.Data = "{\"power\": true}"
		err := client.Respond(deviceId, serviceId, command)
		if err != nil {
			t.Error(err)
			return
		}
	})
	if err != nil {
		t.Error(err)
		return
	}

	responseCount := 0
	token := mqttclient.Subscribe("response/test-device-id/test-service-id", 2, func(_ paho.Client, message paho.Message) {
		responseCount = responseCount + 1
		expectedMessage := `{"command_id":"command-id","data":"{\"power\": true}"}`
		if string(message.Payload()) != expectedMessage {
			t.Error(string(message.Payload()))
			return
		}
	})
	if token.Wait() && token.Error() != nil {
		log.Println("ERROR: device management subscription: ", token.Error())
		t.Error(err)
		return
	}

	token = mqttclient.Publish("command/test-device-id/test-service-id", 2, false, `{"command_id": "command-id", "data": "{\"power\": false}"}`)
	if token.Wait() && token.Error() != nil {
		log.Println("ERROR: device management refresh: ", token.Error())
		t.Error(err)
		return
	}

	time.Sleep(1 * time.Second)

	if commandCount != 1 {
		t.Error(commandCount)
		return
	}

	if responseCount != 1 {
		t.Error(responseCount)
		return
	}
}
