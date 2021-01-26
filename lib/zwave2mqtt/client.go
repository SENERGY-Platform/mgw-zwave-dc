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

const GetNodesCommandTopic = "/getNodes"
const NodeAvailableTopic = "/node_available"

type Client struct {
	mqtt               paho.Client
	apiTopic           string
	networkEventsTopic string
	debug              bool
	deviceInfoListener DeviceInfoListener
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
		apiTopic:           config.ZwaveMqttApiTopic,
		networkEventsTopic: config.ZwaveNetworkEventsTopic,
		debug:              config.Debug,
	}
	return client, client.startDefaultListener()
}

func (this *Client) SetGetDeviceInfoListener(listener DeviceInfoListener) {
	this.deviceInfoListener = listener
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
	return nil
}

func (this *Client) startNodeCommandListener() error {
	if !this.mqtt.IsConnected() {
		log.Println("WARNING: mqtt client not connected")
		return errors.New("mqtt client not connected")
	}

	token := this.mqtt.Subscribe(this.apiTopic+GetNodesCommandTopic, 2, func(client paho.Client, message paho.Message) {
		if this.deviceInfoListener != nil {
			if this.debug {
				log.Println("getNodes response: \n", string(message.Payload()))
			}
			result := NodeInfoResultWrapper{}
			err := json.Unmarshal(message.Payload(), &result)
			if err != nil {
				log.Println("ERROR: unable to unmarshal getNodes result", err)
				return
			}
			deviceInfos := []DeviceInfo{}
			for _, node := range result.Result {
				deviceInfo := DeviceInfo{
					NodeId:         node.NodeId,
					Name:           node.Name,
					Manufacturer:   node.Manufacturer,
					ManufacturerId: node.ManufacturerId,
					Product:        node.Product,
					ProductType:    node.ProductType,
					ProductId:      node.DeviceId,
					Type:           node.Type,
				}
				if deviceInfo.IsValid() {
					deviceInfos = append(deviceInfos, deviceInfo)
				} else if this.debug {
					log.Println("IGNORE:", deviceInfo)
				}
			}
			this.deviceInfoListener(deviceInfos)
		}
	})
	if token.Wait() && token.Error() != nil {
		log.Println("Error on Subscribe: ", this.apiTopic+GetNodesCommandTopic, token.Error())
		return token.Error()
	}
	return nil
}

func (this *Client) startNodeEventListener() error {
	if !this.mqtt.IsConnected() {
		log.Println("WARNING: mqtt client not connected")
		return errors.New("mqtt client not connected")
	}

	token := this.mqtt.Subscribe(this.networkEventsTopic+NodeAvailableTopic, 2, func(client paho.Client, message paho.Message) {
		if this.deviceInfoListener != nil {
			if this.debug {
				log.Println("node available event: \n", string(message.Payload()))
			}
			wrapper := NodeAvailableMessageWrapper{}
			err := json.Unmarshal(message.Payload(), &wrapper)
			if err != nil {
				log.Println("ERROR: unable to unmarshal getNodes result", err)
				return
			}
			if len(wrapper.Data) < 2 {
				err = errors.New("unexpected node available event value")
				log.Println(err, message.Payload())
				return
			}
			nodeIdF, ok := wrapper.Data[0].(float64)
			if !ok {
				err = errors.New("unexpected node available event value (unable to cast nodeId)")
				log.Println(err, message.Payload())
				return
			}
			temp, err := json.Marshal(wrapper.Data[1])
			if err != nil {
				log.Println("ERROR: unable to normalize node available event value", err)
				return
			}
			info := NodeInfo{}
			err = json.Unmarshal(temp, &info)
			if err != nil {
				log.Println("ERROR: unable to normalize node available event value (2)", err)
				return
			}
			deviceInfo := DeviceInfo{
				NodeId:         int64(nodeIdF),
				Name:           info.Name,
				Manufacturer:   info.Manufacturer,
				ManufacturerId: info.ManufacturerId,
				Product:        info.Product,
				ProductType:    info.ProductType,
				ProductId:      info.ProductId,
				Type:           info.Type,
			}
			if deviceInfo.IsValid() {
				this.deviceInfoListener([]DeviceInfo{deviceInfo})
			} else if this.debug {
				log.Println("IGNORE:", deviceInfo)
			}
		}
	})
	if token.Wait() && token.Error() != nil {
		log.Println("Error on Subscribe: ", this.apiTopic+GetNodesCommandTopic, token.Error())
		return token.Error()
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
