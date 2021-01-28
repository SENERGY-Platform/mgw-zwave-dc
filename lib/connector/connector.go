package connector

import (
	"context"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
	"zwave2mqtt-connector/lib/configuration"
	"zwave2mqtt-connector/lib/mgw"
	"zwave2mqtt-connector/lib/zwave2mqtt"
)

type Connector struct {
	mgwClient            *mgw.Client
	z2mClient            *zwave2mqtt.Client
	deviceRegister       map[string]mgw.DeviceInfo
	deviceRegisterMux    sync.Mutex
	valueStore           map[string]interface{}
	valueStoreMux        sync.Mutex
	connectorId          string
	deviceTypeMapping    map[string]string
	updateTicker         *time.Ticker
	updateTickerDuration time.Duration
}

func New(config configuration.Config, ctx context.Context) (result *Connector, err error) {
	result = &Connector{
		deviceRegister:    map[string]mgw.DeviceInfo{},
		valueStore:        map[string]interface{}{},
		connectorId:       config.ConnectorId,
		deviceTypeMapping: config.DeviceTypeMapping,
	}
	result.z2mClient, err = zwave2mqtt.New(config, ctx)
	if err != nil {
		return nil, err
	}
	result.mgwClient, err = mgw.New(config, ctx, result.NotifyRefresh)
	if err != nil {
		return nil, err
	}
	result.z2mClient.SetValueEventListener(result.ValueEventListener)
	result.z2mClient.SetDeviceInfoListener(result.DeviceInfoListener)

	if config.UpdatePeriod != "" && config.UpdatePeriod != "-" {
		result.updateTickerDuration, err = time.ParseDuration(config.UpdatePeriod)
		if err != nil {
			log.Println("ERROR: unable to parse update period as duration")
			return nil, err
		}
		result.updateTicker = time.NewTicker(result.updateTickerDuration)
		go func() {
			log.Println("send periodical update request to z2m", result.z2mClient.RequestDeviceInfoUpdate())
		}()
	}

	return result, nil
}

//returns ids for mgw (with prefixes and suffixes) and the value
func (this *Connector) parseNodeValueAsMgwEvent(nodeValue zwave2mqtt.NodeValue) (deviceId string, serviceId string, value interface{}, err error) {
	rawDeviceId := strconv.FormatInt(nodeValue.NodeId, 10)
	rawServiceId := strconv.FormatInt(nodeValue.ClassId, 10) +
		"-" + strconv.FormatInt(nodeValue.Instance, 10) +
		"-" + strconv.FormatInt(nodeValue.Index, 10)
	deviceId = this.addDeviceIdPrefix(rawDeviceId)
	serviceId = this.addGetServiceSuffix(rawServiceId)
	value = nodeValue.Value
	return
}

func (this *Connector) addGetServiceSuffix(rawServiceId string) string {
	return rawServiceId + ":get"
}

func (this *Connector) isGetServiceId(serviceId string) bool {
	return strings.HasSuffix(serviceId, ":get")
}

func (this *Connector) addDeviceIdPrefix(rawDeviceId string) string {
	return this.connectorId + ":" + rawDeviceId
}

func (this *Connector) removeDeviceIdPrefix(deviceId string) string {
	return strings.Replace(deviceId, this.connectorId+":", "", 1)
}
