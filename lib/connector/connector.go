package connector

import (
	"context"
	"errors"
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/configuration"
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/mgw"
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/model"
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/zwave2mqtt"
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/zwavejs2mqtt"
	"log"
	"net/url"
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

type Connector struct {
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
}

func New(config configuration.Config, ctx context.Context) (result *Connector, err error) {
	result = &Connector{
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

	return result, nil
}

// returns ids for mgw (with prefixes and suffixes) and the value
func (this *Connector) parseNodeValueAsMgwEvent(nodeValue model.Value) (deviceId string, serviceId string, value interface{}, err error) {
	rawDeviceId := strconv.FormatInt(nodeValue.NodeId, 10)
	rawServiceId := nodeValue.ComputedServiceId

	if rawServiceId == "" {
		//legacy, if Z2mClient didnt calculate service id
		rawServiceId = strconv.FormatInt(nodeValue.ClassId, 10) +
			"-" + strconv.FormatInt(nodeValue.Instance, 10) +
			"-" + strconv.FormatInt(nodeValue.Index, 10)
	}

	rawServiceId = encodeLocalId(rawServiceId)

	deviceId = this.addDeviceIdPrefix(rawDeviceId)
	serviceId = this.addGetServiceSuffix(rawServiceId)
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

func (this *Connector) addGetServiceSuffix(rawServiceId string) string {
	return rawServiceId + ":get"
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

const escapedChars = "+#/" // % is implicitly escaped because the encoded values contain a %

func encodeLocalId(raw string) (encoded string) {
	encoded = strings.ReplaceAll(raw, "%", url.QueryEscape("%"))
	for _, char := range escapedChars {
		encoded = strings.ReplaceAll(encoded, string(char), url.QueryEscape(string(char)))
	}
	return
}

func decodeLocalId(encoded string) (decoded string) {
	decoded = encoded
	for _, char := range escapedChars {
		decoded = strings.ReplaceAll(decoded, url.QueryEscape(string(char)), string(char))
	}
	decoded = strings.ReplaceAll(decoded, url.QueryEscape("%"), "%")
	return
}
