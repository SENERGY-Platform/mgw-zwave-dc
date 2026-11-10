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
	"testing"

	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/configuration"
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/devicerepo"
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/model"
)

func Test_nodeToDeviceType_unfinished(t *testing.T) {
	c := &Connector{config: configuration.Config{}}
	node := model.DeviceInfo{
		NodeId:         5,
		Manufacturer:   "Acme",
		ManufacturerId: "0x1",
		Product:        "Plug",
		ProductType:    "0x2",
		ProductId:      "0x3",
	}

	dt := c.nodeToDeviceType(node)

	if dt.Name != "zwavejs2mqtt Acme Plug" {
		t.Errorf("unexpected name: %q", dt.Name)
	}
	found := false
	for _, attr := range dt.Attributes {
		if attr.Key == devicerepo.AttributeUnfinished && attr.Value == "true" {
			found = true
		}
	}
	if !found {
		t.Errorf("attribute %v=true missing in %v", devicerepo.AttributeUnfinished, dt.Attributes)
	}
}
