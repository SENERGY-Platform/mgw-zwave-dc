package zwavejs2mqtt

import (
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/model"
	paho "github.com/eclipse/paho.mqtt.golang"
	"log"
	"strconv"
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
				this.ForwardError("unable to unmarshal getNodes wrapper: " + err.Error())
				return
			}
			deviceInfos := []model.DeviceInfo{}
			huskIds := []int64{}
			for _, node := range wrapper.Result {
				deviceInfo := model.DeviceInfo{
					NodeId:         node.Id,
					Name:           node.Name,
					Manufacturer:   node.Manufacturer,
					ManufacturerId: strconv.FormatInt(node.ManufacturerId, 10),
					Product:        node.ProductDescription,
					ProductType:    strconv.FormatInt(node.ProductType, 10),
					ProductId:      strconv.FormatInt(node.ProductId, 10),
					Values:         transformValues(node.Values),
				}
				if deviceInfo.IsValid() {
					deviceInfos = append(deviceInfos, deviceInfo)
				} else if deviceInfo.IsHusk() {
					huskIds = append(huskIds, deviceInfo.NodeId)
				} else if this.debug {
					log.Println("IGNORE:", deviceInfo)
				}
			}
			this.deviceInfoListener(deviceInfos, huskIds, true, true)
		}
	})
	if token.Wait() && token.Error() != nil {
		log.Println("Error on Subscribe: ", this.apiTopic+GetNodesCommandTopic, token.Error())
		this.ForwardError("Error on Subscribe: " + token.Error().Error())
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
				this.ForwardError("unable to unmarshal getNodes result: " + err.Error())
				return
			}
			if len(wrapper.Data) < 2 {
				err = errors.New("unexpected node available event value")
				log.Println(err, message.Payload())
				this.ForwardError(err.Error())
				return
			}
			nodeIdF, ok := wrapper.Data[0].(float64)
			if !ok {
				err = errors.New("unexpected node available event value (unable to cast nodeId)")
				log.Println(err, message.Payload())
				this.ForwardError(err.Error())
				return
			}
			temp, err := json.Marshal(wrapper.Data[1])
			if err != nil {
				log.Println("ERROR: unable to normalize node available event value", err)
				this.ForwardError("unable to normalize node available event value: " + err.Error())
				return
			}
			info := NodeInfo{}
			err = json.Unmarshal(temp, &info)
			if err != nil {
				log.Println("ERROR: unable to normalize node available event value (2)", err)
				this.ForwardError("unable to normalize node available event value (2): " + err.Error())
				return
			}
			deviceInfo := model.DeviceInfo{
				NodeId:         int64(nodeIdF),
				Name:           info.Name,
				Manufacturer:   info.Manufacturer,
				ManufacturerId: strconv.FormatInt(info.ManufacturerId, 10),
				Product:        info.ProductDescription,
				ProductType:    strconv.FormatInt(info.ProductType, 10),
				ProductId:      strconv.FormatInt(info.ProductId, 10),
			}
			if deviceInfo.IsValid() {
				this.deviceInfoListener([]model.DeviceInfo{deviceInfo}, []int64{}, false, false)
			} else if this.debug {
				log.Println("IGNORE:", deviceInfo)
			}
		}
	})
	if token.Wait() && token.Error() != nil {
		log.Println("Error on Subscribe: ", this.apiTopic+GetNodesCommandTopic, token.Error())
		this.ForwardError("Error on Subscribe: " + token.Error().Error())
		return token.Error()
	}
	return nil
}
