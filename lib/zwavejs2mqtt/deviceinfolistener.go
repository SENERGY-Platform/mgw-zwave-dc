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
	log.Println("subscribe:", this.apiTopic+GetNodesCommandTopic)
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

	log.Println("subscribe:", this.networkEventsTopic+NodeAvailableTopic)
	token := this.mqtt.Subscribe(this.networkEventsTopic+NodeAvailableTopic, 2, func(client paho.Client, message paho.Message) {
		if this.deviceInfoListener != nil {
			if this.debug {
				log.Println("node available event: \n", string(message.Payload()))
			}
			wrapper := NodeAvailableMessage{}
			err := json.Unmarshal(message.Payload(), &wrapper)
			if err != nil {
				log.Println("ERROR: unable to unmarshal getNodes result", err)
				this.ForwardError("unable to unmarshal getNodes result: " + err.Error())
				return
			}

			for _, info := range wrapper.Data {
				if info.Id > 1 {
					deviceInfo := model.DeviceInfo{
						NodeId:         info.Id,
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
