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

package model

import (
	"net/url"
	"strconv"
	"strings"
)

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

func (this DeviceInfo) GetTypeMappingKey() string {
	return this.ManufacturerId + "." + this.ProductType + "." + this.ProductId
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

func (this Value) GetServiceId(get bool) string {
	serviceId := this.ComputedServiceId

	if serviceId == "" {
		//legacy, if Z2mClient didnt calculate service id
		serviceId = strconv.FormatInt(this.ClassId, 10) +
			"-" + strconv.FormatInt(this.Instance, 10) +
			"-" + strconv.FormatInt(this.Index, 10)
	}

	serviceId = EncodeLocalId(serviceId)
	if get {
		serviceId = serviceId + ":get"
	}
	return serviceId
}

const escapedChars = "+#/" // % is implicitly escaped because the encoded values contain a %

func EncodeLocalId(raw string) (encoded string) {
	encoded = strings.ReplaceAll(raw, "%", url.QueryEscape("%"))
	for _, char := range escapedChars {
		encoded = strings.ReplaceAll(encoded, string(char), url.QueryEscape(string(char)))
	}
	return
}

func DecodeLocalId(encoded string) (decoded string) {
	decoded = encoded
	for _, char := range escapedChars {
		decoded = strings.ReplaceAll(decoded, url.QueryEscape(string(char)), string(char))
	}
	decoded = strings.ReplaceAll(decoded, url.QueryEscape("%"), "%")
	return
}
