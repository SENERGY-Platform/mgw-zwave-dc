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
	"time"

	"github.com/SENERGY-Platform/device-repository/v2/lib/client"
	"github.com/SENERGY-Platform/models/go/models"
)

const AttributeUsedForZwave = "senergy/zwave-dc"

func (this *DeviceRepo) ListZwaveDeviceTypes() (list []models.DeviceType, err error) {
	age := time.Since(this.lastDtRefresh)
	if age > this.maxCacheDuration {
		err = this.refreshDeviceTypeList()
		if err != nil {
			return nil, err
		}
	}
	return this.getDeviceTypeList(), nil
}

func (this *DeviceRepo) refreshDeviceTypeList() error {
	this.dtMux.Lock()
	defer this.dtMux.Unlock()
	result, err := this.getDeviceTypeListFromDeviceRepository()
	if err != nil {
		this.config.GetLogger().Warn("unable to load device type list from device repository", "error", err)
		return err
	}
	this.deviceTypes = result
	this.lastDtRefresh = time.Now()
	return nil
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

func (this *DeviceRepo) getDeviceTypeList() []models.DeviceType {
	this.dtMux.Lock()
	defer this.dtMux.Unlock()
	return this.deviceTypes
}
