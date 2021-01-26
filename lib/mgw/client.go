package mgw

import (
	paho "github.com/eclipse/paho.mqtt.golang"
	"log"
	"sync"
	"zwave2mqtt-connector/lib/configuration"
)

const DeviceManagerTopic = "device-manager/device/"

type Client struct {
	mqtt                         paho.Client
	debug                        bool
	connectorId                  string
	subscriptions                []Subscription
	subscriptionsMux             sync.Mutex
	deviceManagerRefreshNotifier func()
}

func New(config configuration.Config, refreshNotifier func()) (*Client, error) {
	client := &Client{
		connectorId:                  config.ConnectorId,
		debug:                        config.Debug,
		deviceManagerRefreshNotifier: refreshNotifier,
	}
	lwt := "device-manager/device/" + config.ConnectorId + "/lw"
	options := paho.NewClientOptions().
		SetPassword(config.MgwMqttPw).
		SetUsername(config.MgwMqttUser).
		SetAutoReconnect(true).
		SetCleanSession(true).
		SetClientID(config.MgwMqttClientId).
		AddBroker(config.MgwMqttBroker).
		SetOnConnectHandler(func(_ paho.Client) {
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

	return client, nil
}

func (this *Client) NotifyDeviceManagerRefresh(f func()) {
	this.deviceManagerRefreshNotifier = f
}
