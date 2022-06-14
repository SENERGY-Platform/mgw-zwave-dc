package zwavejs2mqtt

import (
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/model"
	"strconv"
	"strings"
)

type ResultWrapper struct {
	Success bool          `json:"success"`
	Message string        `json:"message"`
	Args    []interface{} `json:"args"`
}

type NodeInfoResultWrapper struct {
	ResultWrapper
	Result []NodeInfo `json:"result"`
}

/*
{
   "data":[
      5,
      {
         "manufacturer":"Danfoss",
         "manufacturerid":"0x0002",
         "product":"Devolo Home Control Radiator Thermostat",
         "producttype":"0x0005",
         "productid":"0x0175",
         "type":"Setpoint Thermostat",
         "name":"",
         "loc":""
      }
   ]
}
*/
type NodeAvailableMessage struct {
	Data []NodeInfo `json:"data"`
}

/*
{
	"id":2,
	"name":"Lampe_Ingo",
	"loc":"",
	"values":{},
	"groups":[],
	"neighbors":[],
	"ready":true,
	"available":true,
	"hassDevices":{},
	"failed":false,
	"inited":true,
	"hexId":"0x0371-0x0003-0x0002",
	"dbLink":"https://devices.zwave-js.io/?jumpTo=0x0371:0x0003:0x0002:2.5",
	"manufacturerId":881,
	"productId":2,
	"productType":3,
	"deviceConfig":{},
	"productLabel":"ZWA002",
	"productDescription":"Bulb 6 Multi-Color",
	"manufacturer":"Aeotec Ltd.",
	"firmwareVersion":"2.5",
	"protocolVersion":3,
	"zwavePlusVersion":1,
	"zwavePlusNodeType":0,
	"zwavePlusRoleType":5,
	"nodeType":1,
	"endpointsCount":0,
	"endpointIndizes":[],
	"isSecure":"unknown",
	"supportsSecurity":false,
	"supportsBeaming":true,
	"isControllerNode":false,
	"isListening":true,
	"isFrequentListening":false,
	"isRouting":true,
	"keepAwake":false,
	"maxDataRate":100000,
	"deviceClass":{},
	"deviceId":"881-2-3",
	"status":"Alive",
	"interviewStage":"Complete",
	"statistics":{},
	"lastActive":1652876439240
}
*/
type NodeInfo struct {
	Id                 int64                `json:"id"`
	DeviceId           string               `json:"deviceId"`
	Manufacturer       string               `json:"manufacturer"`
	ManufacturerId     int64                `json:"manufacturerId"`
	ProductDescription string               `json:"productDescription"`
	ProductType        int64                `json:"productType"`
	ProductId          int64                `json:"productId"`
	Name               string               `json:"name"`
	Values             map[string]NodeValue `json:"values"`
}

/*
Values ids unique strings have changed, in Z2M valueIds were identified by <nodeId>/<commandClass>/<endpoint>/<index> now they are <nodeId>/<commandClass>/<endpoint>/<property>/<propertyKey?> where property and propertyKey (can be undefined) can be both numbers or strings based on the value. So essentially if you are using Home Assistant or MQTT functions all topics will change, here you can see how we have translated some valueids o
{
   "id":"2-38-0-targetValue",
   "nodeId":2,
   "commandClass":38,
   "commandClassName":"Multilevel Switch",
   "endpoint":0,
   "property":"targetValue",
   "propertyName":"targetValue",
   "type":"number",
   "readable":true,
   "writeable":true,
   "label":"Target value",
   "stateless":false,
   "commandClassVersion":2,
   "min":0,
   "max":99,
   "list":false,
   "lastUpdate":1652869774330
}
*/

type NodeValue struct {
	Id           string      `json:"id"`
	NodeId       int64       `json:"nodeId"`
	CommandClass int64       `json:"commandClass"`
	Endpoint     int64       `json:"endpoint"`
	Type         string      `json:"type"`
	Label        string      `json:"label"`
	Readable     bool        `json:"readable"`
	Writeable    bool        `json:"writeable"`
	Values       interface{} `json:"values"`
	Value        interface{} `json:"value"`
	LastUpdate   int64       `json:"lastUpdate"`
}

func transformValues(values map[string]NodeValue) (result map[string]model.Value) {
	result = map[string]model.Value{}
	for key, value := range values {
		result[key] = transformValue(value)
	}
	return result
}

/*
"id":"2-38-0-targetValue",
   "nodeId":2,
   "commandClass":38,
   "commandClassName":"Multilevel Switch",
   "endpoint":0,
   "property":"targetValue",
   "propertyName":"targetValue",
   "type":"number",
   "readable":true,
   "writeable":true,
   "label":"Target value",
   "stateless":false,
   "commandClassVersion":2,
   "min":0,
   "max":99,
   "list":false,
   "lastUpdate":1652869774330
*/
func transformValue(value NodeValue) (result model.Value) {
	return model.Value{
		ValueId:           value.Id,
		NodeId:            value.NodeId,
		ClassId:           value.CommandClass,
		Type:              value.Type,
		Instance:          value.Endpoint,
		Label:             value.Label,
		ReadOnly:          !value.Writeable,
		WriteOnly:         !value.Readable,
		Values:            value.Values,
		Value:             value.Value,
		LastUpdate:        value.LastUpdate,
		ComputedServiceId: strings.ReplaceAll(strings.TrimPrefix(value.Id, strconv.FormatInt(value.NodeId, 10)+"-"), "/", "_"),
	}
}

type DeviceStateMsg struct {
	Status string `json:"status"`
	NodeId int64  `json:"nodeId"`
}
