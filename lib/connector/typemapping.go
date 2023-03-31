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
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/mgw"
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/model"
	"strconv"
)

// result id with prefix
func (this *Connector) nodeToDeviceInfo(node model.DeviceInfo) (id string, info mgw.DeviceInfo, err error) {
	id = this.nodeIdToDeviceId(node.NodeId)
	info = mgw.DeviceInfo{
		Name:  node.Name,
		State: mgw.Online,
	}
	if info.Name == "" {
		info.Name = getDefaultName(node)
	}
	info.DeviceType, err = this.provideDeviceTypeId(node)
	return
}

func getDefaultName(node model.DeviceInfo) string {
	return node.Product + " (" + strconv.FormatInt(node.NodeId, 10) + ")"
}

func (this *Connector) getTypeMappingKey(node model.DeviceInfo) string {
	return node.ManufacturerId + "." + node.ProductType + "." + node.ProductId
}
