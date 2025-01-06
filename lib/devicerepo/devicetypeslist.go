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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/models/go/models"
	"log"
	"time"
)

const AttributeUsedForZwave = "senergy/zwave-dc"
const DtFallbackKey = "device-types"

func (this *DeviceRepo) ListZwaveDeviceTypes() (list []models.DeviceType, err error) {
	age := time.Since(this.lastDtRefresh)
	if (this.lastDtRefreshUsedFallback && age > this.minCacheDuration) || age > this.maxCacheDuration {
		err = this.refreshDeviceTypeList()
		if err != nil {
			return nil, err
		}
	}
	return this.getDeviceTypeList(), nil
}

func (this *DeviceRepo) getLastDtRefreshUsedFallback() bool {
	this.dtMux.Lock()
	defer this.dtMux.Unlock()
	return this.lastDtRefreshUsedFallback
}

func (this *DeviceRepo) refreshDeviceTypeList() error {
	this.dtMux.Lock()
	defer this.dtMux.Unlock()
	result, err := this.getDeviceTypeListFromDeviceRepository()
	if err == nil {
		this.deviceTypes = result
		this.lastDtRefresh = time.Now()
		this.lastDtRefreshUsedFallback = false
		err = this.fallback.Set(DtFallbackKey, this.deviceTypes)
		if err != nil {
			log.Println("WARNING: unable to store device-types in fallback file")
		}
		return nil
	} else {
		log.Println("WARNING: use fallback file to load device type list")
		result, err = this.getDeviceTypeListFromFallback()
		if err != nil {
			return err
		}
		this.deviceTypes = result
		this.lastDtRefresh = time.Now()
		this.lastDtRefreshUsedFallback = true
		return nil
	}
}

func (this *DeviceRepo) getDeviceTypeListFromDeviceRepository() (result []models.DeviceType, err error) {
	token, err := this.getToken()
	if err != nil {
		return result, err
	}
	result, _, err, _ = client.NewClient(this.config.DeviceRepositoryUrl, nil).ListDeviceTypesV3(token, client.DeviceTypeListOptions{
		Limit:         9999,
		Offset:        0,
		SortBy:        "name.asc",
		AttributeKeys: []string{AttributeUsedForZwave},
	})
	if err != nil {
		return result, err
	}
	return result, nil
}

func (this *DeviceRepo) getDeviceTypeListFromFallback() (result []models.DeviceType, err error) {
	value, fallbackerr := this.fallback.Get(DtFallbackKey)
	if fallbackerr != nil {
		log.Println("ERROR: unable to load fallback", fallbackerr)
		return result, errors.Join(err, fallbackerr)
	}
	var ok bool
	result, ok = value.([]models.DeviceType)
	if !ok {
		err = jsonCast(value, &result)
		if err != nil {
			err = fmt.Errorf("fallback file does not contain expected format: %w", err)
			log.Println("ERROR:", err)
			return result, err
		}
		this.fallback.Set(DtFallbackKey, result)
	}
	return result, nil
}

func jsonCast(in interface{}, out interface{}) error {
	temp, err := json.Marshal(in)
	if err != nil {
		return err
	}
	err = json.Unmarshal(temp, out)
	return err
}

func (this *DeviceRepo) getDeviceTypeList() []models.DeviceType {
	this.dtMux.Lock()
	defer this.dtMux.Unlock()
	return this.deviceTypes
}
