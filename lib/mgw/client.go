package mgw

import (
	"context"
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/configuration"
	paho "github.com/eclipse/paho.mqtt.golang"
	"log"
	"sync"
	"time"
)

const DeviceManagerTopic = "device-manager/device"

type Client struct {
	mqtt                         paho.Client
	debug                        bool
	connectorId                  string
	subscriptions                map[string]paho.MessageHandler
	subscriptionsMux             sync.Mutex
	deviceManagerRefreshNotifier func()
}

func New(config configuration.Config, ctx context.Context, refreshNotifier func()) (*Client, error) {
	client := &Client{
		connectorId:                  config.ConnectorId,
		debug:                        config.Debug,
		deviceManagerRefreshNotifier: refreshNotifier,
		subscriptions:                map[string]paho.MessageHandler{},
	}
	lwt := "device-manager/device/" + config.ConnectorId + "/lw"
	options := paho.NewClientOptions().
		SetPassword(config.MgwMqttPw).
		SetUsername(config.MgwMqttUser).
		SetAutoReconnect(true).
		SetCleanSession(true).
		SetClientID(config.MgwMqttClientId).
		AddBroker(config.MgwMqttBroker).
		SetWriteTimeout(10*time.Second).
		SetOrderMatters(false).
		SetResumeSubs(true).
		SetConnectionLostHandler(func(_ paho.Client, err error) {
			log.Println("connection to mgw broker lost")
		}).
		SetOnConnectHandler(func(_ paho.Client) {
			log.Println("connected to mgw broker")
			err := client.initSubscriptions()
			if err != nil {
				log.Fatal("FATAL: ", err)
			}
			if client.deviceManagerRefreshNotifier != nil {
				client.deviceManagerRefreshNotifier()
			}
		}).SetWill(lwt, "offline", 2, false)

	client.mqtt = paho.NewClient(options)
	if token := client.mqtt.Connect(); token.Wait() && token.Error() != nil {
		log.Println("Error on MqttStart.Connect(): ", token.Error())
		return nil, token.Error()
	}

	go func() {
		<-ctx.Done()
		client.mqtt.Disconnect(0)
	}()

	return client, nil
}

func (this *Client) NotifyDeviceManagerRefresh(f func()) {
	this.deviceManagerRefreshNotifier = f
}
