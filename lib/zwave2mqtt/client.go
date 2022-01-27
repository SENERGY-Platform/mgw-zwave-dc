package zwave2mqtt

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/configuration"
	paho "github.com/eclipse/paho.mqtt.golang"
	"log"
	"strconv"
	"strings"
	"time"
)

type DeviceInfoListener func(nodes []DeviceInfo, huskIds []int64, withValues bool, allKnownDevices bool)
type ValueEventListener func(value NodeValue)

const GetNodesCommandTopic = "/getNodes"
const NodeAvailableTopic = "/node_available"

type Client struct {
	mqtt               paho.Client
	debug              bool
	valueEventTopic    string
	apiTopic           string
	networkEventsTopic string
	deviceInfoListener DeviceInfoListener
	valueEventListener ValueEventListener
	forwardErrorMsg    func(msg string)
}

func New(config configuration.Config, ctx context.Context) (*Client, error) {
	client := &Client{
		valueEventTopic:    config.ZvaveValueEventTopic,
		apiTopic:           config.ZwaveMqttApiTopic,
		networkEventsTopic: config.ZwaveNetworkEventsTopic,
		debug:              config.Debug,
	}
	options := paho.NewClientOptions().
		SetPassword(config.ZwaveMqttPw).
		SetUsername(config.ZwaveMqttUser).
		SetAutoReconnect(true).
		SetCleanSession(true).
		SetClientID(config.ZwaveMqttClientId).
		AddBroker(config.ZwaveMqttBroker).
		SetWriteTimeout(10 * time.Second).
		SetConnectionLostHandler(func(_ paho.Client, err error) {
			log.Println("connection to zwave2mqtt broker lost")
		}).
		SetOnConnectHandler(func(_ paho.Client) {
			log.Println("connected to zwave2mqtt broker")
			err := client.startDefaultListener()
			if err != nil {
				log.Fatal("FATAL: ", err)
			}
		})

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

func (this *Client) ForwardError(msg string) {
	if this.forwardErrorMsg != nil {
		this.forwardErrorMsg(msg)
	}
}

func (this *Client) SetErrorForwardingFunc(f func(msg string)) {
	this.forwardErrorMsg = f
}

func (this *Client) SetDeviceInfoListener(listener DeviceInfoListener) {
	this.deviceInfoListener = listener
}

func (this *Client) SetValueEventListener(listener ValueEventListener) {
	this.valueEventListener = listener
}

func (this *Client) startDefaultListener() error {
	err := this.startNodeCommandListener()
	if err != nil {
		return err
	}
	err = this.startNodeEventListener()
	if err != nil {
		return err
	}
	err = this.startValueEventListener()
	if err != nil {
		return err
	}
	return nil
}

func (this *Client) RequestDeviceInfoUpdate() error {
	return this.SendZwayCommand(GetNodesCommandTopic, []interface{}{})
}

func (this *Client) SetValue(nodeId int64, classId int64, instanceId int64, index int64, value interface{}) error {
	return this.SendZwayCommand("/setValue", []interface{}{nodeId, classId, instanceId, index, value})
}

func (this *Client) SetValueByValueId(valueId string, value interface{}) error {
	args := []interface{}{}
	for _, v := range strings.Split(valueId, "-") {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return err
		}
		args = append(args, id)
	}
	args = append(args, value)
	return this.SendZwayCommand("/setValue", args)
}

func (this *Client) SendZwayCommand(command string, args []interface{}) error {
	if !this.mqtt.IsConnected() {
		log.Println("WARNING: mqtt client not connected")
		return errors.New("mqtt client not connected")
	}
	topic := this.apiTopic + command + "/set"
	payload := map[string]interface{}{
		"args": args,
	}
	msg, err := json.Marshal(payload)
	if this.debug {
		log.Println("DEBUG: publish ", topic, string(msg))
	}
	token := this.mqtt.Publish(topic, 2, false, string(msg))
	if token.Wait() && token.Error() != nil {
		log.Println("Error on Client.Publish(): ", token.Error())
		return token.Error()
	}
	return err
}
