/*
 * Copyright (c) 2023 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package connector

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/configuration"
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/devicerepo"
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/devicerepo/auth"
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/mgw"
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/model"
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/zwave2mqtt"
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/zwavejs2mqtt"
	"github.com/SENERGY-Platform/models/go/models"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Z2mClient interface {
	SetErrorForwardingFunc(clientError func(message string))
	SetValueEventListener(listener func(nodeValue model.Value))
	SetDeviceInfoListener(listener func(nodes []model.DeviceInfo, huskIds []int64, withValues bool, allKnownDevices bool))
	RequestDeviceInfoUpdate() error
	SetValueByValueId(id string, value interface{}) error
	SetDeviceStatusListener(state func(nodeId int64, online bool) error)
}

type DeviceRepo interface {
	FindDeviceTypeId(device model.DeviceInfo) (dtId string, usedFallback bool, err error)
	CreateDeviceTypeWithDistinctAttributes(key string, dt models.DeviceType, attributeKeys []string) (result models.DeviceType, code int, err error)
}

type Connector struct {
	config                       configuration.Config
	mgwClient                    *mgw.Client
	z2mClient                    Z2mClient
	deviceRegister               map[string]mgw.DeviceInfo
	deviceRegisterMux            sync.Mutex
	valueStore                   map[string]interface{}
	valueStoreMux                sync.Mutex
	connectorId                  string
	deviceIdPrefix               string
	deviceTypeMapping            map[string]string
	updateTicker                 *time.Ticker
	updateTickerDuration         time.Duration
	deleteMissingDevices         bool
	husksShouldBeDeleted         bool
	eventsForUnregisteredDevices bool
	nodeDeviceTypeOverwrite      map[string]string
	devicerepo                   DeviceRepo
}

func New(config configuration.Config, ctx context.Context) (result *Connector, err error) {
	result = &Connector{
		config:                       config,
		deviceRegister:               map[string]mgw.DeviceInfo{},
		valueStore:                   map[string]interface{}{},
		connectorId:                  config.ConnectorId,
		deviceIdPrefix:               config.DeviceIdPrefix,
		deviceTypeMapping:            config.DeviceTypeMapping,
		deleteMissingDevices:         config.DeleteMissingDevices,
		husksShouldBeDeleted:         config.DeleteHusks,
		eventsForUnregisteredDevices: config.EventsForUnregisteredDevices,
		nodeDeviceTypeOverwrite:      config.NodeDeviceTypeOverwrite,
	}

	result.devicerepo, err = devicerepo.New(config, &auth.Auth{})
	if err != nil {
		return nil, err
	}

	switch config.ZwaveController {
	case "":
		fallthrough
	case "zwave2mqtt":
		result.z2mClient, err = zwave2mqtt.New(config, ctx)
	case "zwavejs2mqtt":
		result.z2mClient, err = zwavejs2mqtt.New(config, ctx)
	default:
		err = errors.New("unknown zwave controller")
	}
	if err != nil {
		return nil, err
	}

	result.mgwClient, err = mgw.New(config, ctx, result.NotifyRefresh)
	if err != nil {
		return nil, err
	}
	result.z2mClient.SetErrorForwardingFunc(result.mgwClient.SendClientError)
	result.z2mClient.SetValueEventListener(result.ValueEventListener)
	result.z2mClient.SetDeviceInfoListener(result.DeviceInfoListener)
	result.z2mClient.SetDeviceStatusListener(result.SetDeviceState)

	if config.UpdatePeriod != "" && config.UpdatePeriod != "-" {
		result.updateTickerDuration, err = time.ParseDuration(config.UpdatePeriod)
		if err != nil {
			log.Println("ERROR: unable to parse update period as duration")
			result.mgwClient.SendClientError("unable to parse update period as duration")
			return nil, err
		}
		result.updateTicker = time.NewTicker(result.updateTickerDuration)

		go func() {
			<-ctx.Done()
			result.updateTicker.Stop()
		}()
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case <-result.updateTicker.C:
					log.Println("send periodical update request to z2m", result.z2mClient.RequestDeviceInfoUpdate())
				}
			}
		}()
	}

	log.Println("initial update request", result.z2mClient.RequestDeviceInfoUpdate())

	go func() {
		timer := time.NewTimer(config.InitialUpdateRequestDelay.GetDuration())
		defer func() {
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
		}()
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			log.Println("delayed initial update request", result.z2mClient.RequestDeviceInfoUpdate())
		}
	}()

	return result, nil
}

// returns ids for mgw (with prefixes and suffixes) and the value
func (this *Connector) parseNodeValueAsMgwEvent(nodeValue model.Value) (deviceId string, serviceId string, value interface{}, err error) {
	serviceId = nodeValue.GetServiceId(true)
	rawDeviceId := strconv.FormatInt(nodeValue.NodeId, 10)
	deviceId = this.addDeviceIdPrefix(rawDeviceId)
	value = ValueWithTimestamp{
		Value:      nodeValue.Value,
		LastUpdate: nodeValue.LastUpdate,
	}
	return
}

type ValueWithTimestamp struct {
	Value      interface{} `json:"value"`
	LastUpdate int64       `json:"lastUpdate"`
}

func (this *Connector) isGetServiceId(serviceId string) bool {
	return strings.HasSuffix(serviceId, ":get")
}

func (this *Connector) nodeIdToDeviceId(nodeId int64) string {
	return this.addDeviceIdPrefix(strconv.FormatInt(nodeId, 10))
}

func (this *Connector) addDeviceIdPrefix(rawDeviceId string) string {
	return this.deviceIdPrefix + ":" + rawDeviceId
}

func (this *Connector) removeDeviceIdPrefix(deviceId string) string {
	return strings.Replace(deviceId, this.deviceIdPrefix+":", "", 1)
}

type Duration struct {
	dur time.Duration
}

func (this *Duration) GetDuration() time.Duration {
	return this.dur
}

func (this *Duration) SetDuration(dur time.Duration) {
	this.dur = dur
}

func (this *Duration) SetString(str string) error {
	if str == "" {
		return nil
	}
	duration, err := time.ParseDuration(str)
	if err != nil {
		return err
	}
	this.SetDuration(duration)
	return nil
}

func (this *Duration) UnmarshalJSON(bytes []byte) (err error) {
	var str string
	err = json.Unmarshal(bytes, &str)
	if err != nil {
		return err
	}
	return this.SetString(str)
}
