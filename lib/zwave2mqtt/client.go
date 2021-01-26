package zwave2mqtt

import (
	"encoding/json"
	"errors"
	paho "github.com/eclipse/paho.mqtt.golang"
	"log"
	"strconv"
	"strings"
	"zwave2mqtt-connector/lib/configuration"
)

type DeviceInfoListener func(nodes []DeviceInfo)
type ValueEventListener func(value NodeValue)

const GetNodesCommandTopic = "/getNodes"
const NodeAvailableTopic = "/node_available"

type Client struct {
	mqtt               paho.Client
	valueEventTopic    string
	apiTopic           string
	networkEventsTopic string
	debug              bool
	deviceInfoListener DeviceInfoListener
	valueEventListener ValueEventListener
}

func New(config configuration.Config) (*Client, error) {
	options := paho.NewClientOptions().
		SetPassword(config.ZwaveMqttPw).
		SetUsername(config.ZwaveMqttUser).
		SetAutoReconnect(true).
		SetCleanSession(false).
		SetClientID(config.ConnectorId).
		AddBroker(config.ZwaveMqttBroker)

	mqtt := paho.NewClient(options)
	if token := mqtt.Connect(); token.Wait() && token.Error() != nil {
		log.Println("Error on MqttStart.Connect(): ", token.Error())
		return nil, token.Error()
	}
	client := &Client{
		mqtt:               mqtt,
		valueEventTopic:    config.ZvaveValueEventTopic,
		apiTopic:           config.ZwaveMqttApiTopic,
		networkEventsTopic: config.ZwaveNetworkEventsTopic,
		debug:              config.Debug,
	}
	return client, client.startDefaultListener()
}

func (this *Client) SetGetDeviceInfoListener(listener DeviceInfoListener) {
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

func (this *Client) LoadNodes() error {
	return this.SendZwayCommand(GetNodesCommandTopic, []interface{}{})
}

func (this *Client) SetValue(nodeId int64, classId int64, instanceId int64, index int64, value interface{}) error {
	return this.SendZwayCommand("setValue", []interface{}{nodeId, classId, instanceId, index, value})
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
	return this.SendZwayCommand("setValue", args)
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
