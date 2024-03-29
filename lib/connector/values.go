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

// expects ids from mgw (with prefixes and suffixes)
func (this *Connector) saveValue(deviceId string, serviceId string, value interface{}) {
	this.valueStoreMux.Lock()
	defer this.valueStoreMux.Unlock()
	this.valueStore[deviceId+"-"+serviceId] = value
}

// expects ids from mgw (with prefixes and suffixes)
func (this *Connector) getValue(deviceId string, serviceId string) (value interface{}, known bool) {
	this.valueStoreMux.Lock()
	defer this.valueStoreMux.Unlock()
	value, known = this.valueStore[deviceId+"-"+serviceId]
	return
}
