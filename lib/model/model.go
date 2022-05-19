package model

type DeviceInfo struct {
	NodeId         int64
	Name           string
	Manufacturer   string
	ManufacturerId string
	Product        string
	ProductType    string
	ProductId      string
	Values         map[string]Value
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
	return true
}

func (this DeviceInfo) IsHusk() bool {
	if this.NodeId == 0 || this.NodeId == 1 {
		return false
	}
	if this.Manufacturer != "" {
		return false
	}
	if this.ManufacturerId != "" {
		return false
	}
	if this.Product != "" {
		return false
	}
	if this.ProductId != "" {
		return false
	}
	if this.ProductType != "" {
		return false
	}
	return true
}

type Value struct {
	ComputedServiceId string      `json:"computedServiceId"`
	ValueId           string      `json:"value_id"`
	NodeId            int64       `json:"node_id"`
	ClassId           int64       `json:"class_id"`
	Type              string      `json:"type"`
	Instance          int64       `json:"instance"`
	Index             int64       `json:"index"`
	Label             string      `json:"label"`
	ReadOnly          bool        `json:"read_only"`
	WriteOnly         bool        `json:"write_only"`
	Values            interface{} `json:"values"`
	Value             interface{} `json:"value"`
	LastUpdate        int64       `json:"lastUpdate"`
}
