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

package devicerepo

import (
	"fmt"
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/configuration"
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/devicerepo/fallback"
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"slices"
	"strings"
	"sync"
	"time"
)

type DeviceRepo struct {
	config                    configuration.Config
	auth                      Auth
	fallback                  fallback.Fallback
	deviceTypes               []models.DeviceType
	minCacheDuration          time.Duration
	maxCacheDuration          time.Duration
	lastDtRefresh             time.Time
	lastDtRefreshUsedFallback bool
	dtMux                     sync.Mutex
	createdDt                 map[string]models.DeviceType
}

type Auth interface {
	EnsureAccess(config configuration.Config) (token string, err error)
}

func New(config configuration.Config, auth Auth) (*DeviceRepo, error) {
	minCacheDuration, err := time.ParseDuration(config.MinCacheDuration)
	if err != nil {
		return nil, err
	}
	maxCacheDuration, err := time.ParseDuration(config.MaxCacheDuration)
	if err != nil {
		return nil, err
	}
	f, err := fallback.NewFallback(config.FallbackFile)
	if err != nil {
		return nil, err
	}
	return &DeviceRepo{
		auth:             auth,
		config:           config,
		fallback:         f,
		minCacheDuration: minCacheDuration,
		maxCacheDuration: maxCacheDuration,
		createdDt:        map[string]models.DeviceType{},
	}, nil
}

func (this *DeviceRepo) getToken() (string, error) {
	return this.auth.EnsureAccess(this.config)
}

func (this *DeviceRepo) FindDeviceTypeId(device model.DeviceInfo) (dtId string, usedFallback bool, err error) {
	deviceTypes, err := this.ListZwaveDeviceTypes()
	if err != nil {
		return "", this.getLastDtRefreshUsedFallback(), err
	}
	deviceType, ok := this.getMatchingDeviceType(deviceTypes, device)
	if !ok && time.Since(this.lastDtRefresh) > this.minCacheDuration {
		err = this.refreshDeviceTypeList()
		if err != nil {
			return "", this.getLastDtRefreshUsedFallback(), err
		}
		deviceTypes, err = this.ListZwaveDeviceTypes()
		if err != nil {
			return "", this.getLastDtRefreshUsedFallback(), err
		}
		deviceType, ok = this.getMatchingDeviceType(deviceTypes, device)
	}
	if !ok {
		return "", this.getLastDtRefreshUsedFallback(), fmt.Errorf("%w: mapping-key=%v", model.NoMatchingDeviceTypeFound, device.GetTypeMappingKey())
	}
	return deviceType.Id, this.getLastDtRefreshUsedFallback(), nil
}

const AttributeZwaveTypeMappingKey = "senergy/zwave-type-mapping-key"

func (this *DeviceRepo) getMatchingDeviceType(devicetypes []models.DeviceType, device model.DeviceInfo) (models.DeviceType, bool) {
	deviceTypeKey := device.GetTypeMappingKey()
	for _, dt := range devicetypes {
		attrMap := map[string][]string{}
		for _, attr := range dt.Attributes {
			attrMap[attr.Key] = append(attrMap[attr.Key], strings.TrimSpace(attr.Value))
		}
		if keys, keyIsSet := attrMap[AttributeZwaveTypeMappingKey]; keyIsSet && slices.Contains(keys, strings.TrimSpace(deviceTypeKey)) {
			return dt, true
		}
	}
	return models.DeviceType{}, false
}
