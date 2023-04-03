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
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/devicerepo"
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"log"
	"strconv"
	"strings"
)

func (this *Connector) provideDeviceTypeId(node model.DeviceInfo) (result string, err error) {
	var known bool

	if this.nodeDeviceTypeOverwrite != nil {
		result, known = this.nodeDeviceTypeOverwrite[strconv.FormatInt(node.NodeId, 10)]
		if known {
			return result, nil
		}
	}

	typeMappingKey := this.getTypeMappingKey(node)
	result, known = this.deviceTypeMapping[typeMappingKey]
	if known {
		return result, nil
	}

	var usedFallback bool

	result, usedFallback, err = this.devicerepo.FindDeviceTypeId(node)
	if errors.Is(err, model.NoMatchingDeviceTypeFound) {
		log.Println("WARNING: unable to find matching device type", err)
		if !usedFallback {
			if this.config.CreateMissingDeviceTypes {
				log.Println("create device type", err)
				result, err = this.createDeviceType(node)
				if err != nil {
					log.Println("WARNING: unable to create device type", err)
					return result, err
				}
				return result, err
			}
		}

	}
	return result, err
}

func (this *Connector) createDeviceType(node model.DeviceInfo) (string, error) {
	dt, _, err := this.devicerepo.CreateDeviceType(node.GetTypeMappingKey(), this.nodeToDeviceType(node))
	return dt.Id, err
}

func (this *Connector) nodeToDeviceType(node model.DeviceInfo) (result models.DeviceType) {
	result = models.DeviceType{
		Name:          fmt.Sprintf("ZWaveJs2Mqtt %v %v", node.Manufacturer, node.Product),
		Description:   "",
		DeviceClassId: this.config.CreateMissingDeviceTypesWithDeviceClass,
		Attributes: []models.Attribute{
			{Key: devicerepo.AttributeUsedForZwave, Value: "true"},
			{Key: devicerepo.AttributeZwaveTypeMappingKey, Value: node.GetTypeMappingKey()},
		},
		Services: []models.Service{
			{
				LocalId:     "statistics",
				Name:        "statistics",
				Interaction: models.EVENT,
				ProtocolId:  this.config.CreateMissingDeviceTypesWithProtocol,
				Outputs: []models.Content{
					{
						ContentVariable: models.ContentVariable{
							Name: "statistics",
							Type: models.Structure,
							SubContentVariables: []models.ContentVariable{
								{Name: "commandsTX", Type: models.Float},
								{Name: "commandsRX", Type: models.Float},
								{Name: "commandsDroppedRX", Type: models.Float},
								{Name: "commandsDroppedTX", Type: models.Float},
								{Name: "timeoutResponse", Type: models.Float},
								{Name: "rtt", Type: models.Float},
							},
						},
						Serialization:     models.JSON,
						ProtocolSegmentId: this.config.CreateMissingDeviceTypesWithProtocolSegment,
					},
				},
			},
		},
	}
	for _, value := range node.Values {
		var valueType models.Type
		switch strings.ToLower(value.Type) {
		case "number", "float", "float64", "float32", "float16", "double", "double64", "double32":
			valueType = models.Float
		case "int", "integer", "int64", "int32", "duration":
			valueType = models.Float
		case "text", "string":
			valueType = models.String
		case "bool", "boolean", "binary":
			valueType = models.Boolean
		default:
			log.Printf("WARNING: unknown value type %v in value %v", value.Type, value.ValueId)
			continue
		}
		if !value.WriteOnly {
			result.Services = append(result.Services, models.Service{
				LocalId:     value.GetServiceId(true),
				Name:        value.Label,
				Interaction: models.EVENT_AND_REQUEST,
				ProtocolId:  this.config.CreateMissingDeviceTypesWithProtocol,
				Outputs: []models.Content{{
					ContentVariable: models.ContentVariable{
						Name:   "value",
						IsVoid: false,
						Type:   models.Structure,
						SubContentVariables: []models.ContentVariable{
							{
								Name: "value",
								Type: valueType,
							},
							{
								Name:             "lastUpdate",
								Type:             models.Integer,
								FunctionId:       this.config.CreateMissingDeviceTypesLastUpdateFunction,
								CharacteristicId: this.config.CreateMissingDeviceTypesLastUpdateCharacteristic,
							},
							{
								Name:          "value_unit",
								Type:          models.String,
								UnitReference: "value",
							},
							{
								Name:          "lastUpdate_unit",
								Type:          models.String,
								UnitReference: "lastUpdate",
							},
						},
					},
					Serialization:     models.JSON,
					ProtocolSegmentId: this.config.CreateMissingDeviceTypesWithProtocolSegment,
				}},
			})
		}
		if !value.ReadOnly {
			result.Services = append(result.Services, models.Service{
				LocalId:     value.GetServiceId(false),
				Name:        value.Label,
				Interaction: models.EVENT_AND_REQUEST,
				ProtocolId:  this.config.CreateMissingDeviceTypesWithProtocol,
				Inputs: []models.Content{{
					ContentVariable: models.ContentVariable{
						Name: "value",
						Type: valueType,
					},
					Serialization:     models.JSON,
					ProtocolSegmentId: this.config.CreateMissingDeviceTypesWithProtocolSegment,
				}},
			})
		}
	}
	return result
}
