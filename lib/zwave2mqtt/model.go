package zwave2mqtt

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
	NodeId         int64                `json:"node_id"`
	DeviceId       string               `json:"device_id"`
	Manufacturer   string               `json:"manufacturer"`
	ManufacturerId string               `json:"manufacturerid"`
	Product        string               `json:"product"`
	ProductType    string               `json:"producttype"`
	ProductId      string               `json:"productid"`
	Type           string               `json:"type"`
	Name           string               `json:"name"`
	Values         map[string]NodeValue `json:"values"`
}

type NodeValue struct {
	ValueId    string      `json:"value_id"`
	NodeId     int64       `json:"node_id"`
	ClassId    int64       `json:"class_id"`
	Type       string      `json:"type"`
	Genre      string      `json:"genre"`
	Instance   int64       `json:"instance"`
	Index      int64       `json:"index"`
	Label      string      `json:"label"`
	Units      string      `json:"units"`
	Help       string      `json:"help"`
	ReadOnly   bool        `json:"read_only"`
	WriteOnly  bool        `json:"write_only"`
	IsPolled   bool        `json:"is_polled"`
	Values     interface{} `json:"values"`
	Value      interface{} `json:"value"`
	LastUpdate int64       `json:"lastUpdate"`
}

type DeviceInfo struct {
	NodeId         int64
	Name           string
	Manufacturer   string
	ManufacturerId string
	Product        string
	ProductType    string
	ProductId      string
	Type           string
	Values         map[string]NodeValue
}

func (this DeviceInfo) IsValid() bool {
	if this.NodeId == 0 || this.NodeId == 1 {
		return false
	}
	if this.Manufacturer == "" {
		return false
	}
	if this.ManufacturerId == "" {
		return false
	}
	if this.Product == "" {
		return false
	}
	if this.ProductId == "" {
		return false
	}
	if this.ProductType == "" {
		return false
	}
	if this.Type == "" {
		return false
	}
	return true
}
