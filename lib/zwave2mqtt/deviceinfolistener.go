package zwave2mqtt

import (
	"encoding/json"
	"errors"
	paho "github.com/eclipse/paho.mqtt.golang"
	"log"
)

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
			wrapper := NodeInfoResultWrapper{}
			err := json.Unmarshal(message.Payload(), &wrapper)
			if err != nil {
				log.Println("ERROR: unable to unmarshal getNodes wrapper", err)
				return
			}
			deviceInfos := []DeviceInfo{}
			for _, node := range wrapper.Result {
				deviceInfo := DeviceInfo{
					NodeId:         node.NodeId,
					Name:           node.Name,
					Manufacturer:   node.Manufacturer,
					ManufacturerId: node.ManufacturerId,
					Product:        node.Product,
					ProductType:    node.ProductType,
					ProductId:      node.DeviceId,
					Type:           node.Type,
					Values:         node.Values,
				}
				if deviceInfo.IsValid() {
					deviceInfos = append(deviceInfos, deviceInfo)
				} else if this.debug {
					log.Println("IGNORE:", deviceInfo)
				}
			}
			this.deviceInfoListener(deviceInfos, true, true)
		}
	})
	if token.Wait() && token.Error() != nil {
		log.Println("Error on Subscribe: ", this.apiTopic+GetNodesCommandTopic, token.Error())
		return token.Error()
	}
	return nil
}

func (this *Client) startNodeEventListener() error {
	if this.networkEventsTopic == "" || this.networkEventsTopic == "-" {
		log.Println("WARNING: no zwave network event topic configured --> no live device availability check")
		return nil
	}
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
				this.deviceInfoListener([]DeviceInfo{deviceInfo}, false, false)
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
