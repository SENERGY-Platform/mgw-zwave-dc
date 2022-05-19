package zwave2mqtt

import "github.com/SENERGY-Platform/mgw-zwave-dc/lib/model"

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
type NodeAvailableMessageWrapper struct {
	Data []interface{} `json:"data"`
}

type NodeAvailableInfo struct {
	Manufacturer   string `json:"manufacturer"`
	ManufacturerId string `json:"manufacturerid"`
	Product        string `json:"product"`
	ProductType    string `json:"producttype"`
	ProductId      string `json:"productid"`
	Type           string `json:"type"`
	Name           string `json:"name"`
	Loc            string `json:"loc"`
}

/*
{
      "node_id":5,
      "device_id":"2-373-5",
      "manufacturer":"Danfoss",
      "manufacturerid":"0x0002",
      "product":"Devolo Home Control Radiator Thermostat",
      "producttype":"0x0005",
      "productid":"0x0175",
      "type":"Setpoint Thermostat",
      "name":"",
      "values":{ ... }
}
*/
type NodeInfo struct {
	NodeId         int64                  `json:"node_id"`
	DeviceId       string                 `json:"device_id"`
	Manufacturer   string                 `json:"manufacturer"`
	ManufacturerId string                 `json:"manufacturerid"`
	Product        string                 `json:"product"`
	ProductType    string                 `json:"producttype"`
	ProductId      string                 `json:"productid"`
	Type           string                 `json:"type"`
	Name           string                 `json:"name"`
	Values         map[string]model.Value `json:"values"`
}
