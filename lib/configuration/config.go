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

package configuration

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ConnectorId                  string            `json:"connector_id"`
	DeviceIdPrefix               string            `json:"device_id_prefix"`
	ZwaveMqttBroker              string            `json:"zwave_mqtt_broker"`
	ZwaveMqttUser                string            `json:"zwave_mqtt_user" config:"secret"`
	ZwaveMqttPw                  string            `json:"zwave_mqtt_pw" config:"secret"`
	ZwaveMqttClientId            string            `json:"zwave_mqtt_client_id"`
	MgwMqttBroker                string            `json:"mgw_mqtt_broker"`
	MgwMqttUser                  string            `json:"mgw_mqtt_user" config:"secret"`
	MgwMqttPw                    string            `json:"mgw_mqtt_pw" config:"secret"`
	MgwMqttClientId              string            `json:"mgw_mqtt_client_id"`
	ZwaveController              string            `json:"zwave_controller"`
	ZwaveMqttDeviceStateTopic    string            `json:"zwave_mqtt_device_state_topic"` //used in zwavejs2mqtt
	ZvaveValueEventTopic         string            `json:"zvave_value_event_topic"`       //used in zwave2mqtt
	ZwaveMqttApiTopic            string            `json:"zwave_mqtt_api_topic"`
	ZwaveNetworkEventsTopic      string            `json:"zwave_network_events_topic"`
	UpdatePeriod                 string            `json:"update_period"`
	InitialUpdateRequestDelay    Duration          `json:"initial_update_request_delay"`
	Debug                        bool              `json:"debug"`
	DeviceTypeMapping            map[string]string `json:"device_type_mapping"`
	DeleteMissingDevices         bool              `json:"delete_missing_devices"`
	DeleteHusks                  bool              `json:"delete_husks"`
	EventsForUnregisteredDevices bool              `json:"events_for_unregistered_devices"`
	NodeDeviceTypeOverwrite      map[string]string `json:"node_device_type_overwrite"`

	AuthEndpoint             string  `json:"auth_endpoint"`
	AuthClientId             string  `json:"auth_client_id" config:"secret"`
	AuthExpirationTimeBuffer float64 `json:"auth_expiration_time_buffer"`
	AuthUsername             string  `json:"auth_username" config:"secret"`
	AuthPassword             string  `json:"auth_password" config:"secret"`

	DeviceManagerUrl    string `json:"device_manager_url"`
	DeviceRepositoryUrl string `json:"device_repository_url"`
	FallbackFile        string `json:"fallback_file"`
	MinCacheDuration    string `json:"min_cache_duration"`
	MaxCacheDuration    string `json:"max_cache_duration"`

	CreateMissingDeviceTypes                         bool   `json:"create_missing_device_types"`
	CreateMissingDeviceTypesWithDeviceClass          string `json:"create_missing_device_types_with_device_class"`
	CreateMissingDeviceTypesWithProtocol             string `json:"create_missing_device_types_with_protocol"`
	CreateMissingDeviceTypesWithProtocolSegment      string `json:"create_missing_device_types_with_protocol_segment"`
	CreateMissingDeviceTypesLastUpdateFunction       string `json:"create_missing_device_types_last_update_function"`
	CreateMissingDeviceTypesLastUpdateCharacteristic string `json:"create_missing_device_types_last_update_characteristic"`
}

// loads config from json in location and used environment variables (e.g ZookeeperUrl --> ZOOKEEPER_URL)
func Load(location string) (config Config, err error) {
	file, err := os.Open(location)
	if err != nil {
		return config, err
	}
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		return config, err
	}
	err = handleEnvironmentVars(&config, os.Getenv)
	return config, err
}

var camel = regexp.MustCompile("(^[^A-Z]*|[A-Z]*)([A-Z][^A-Z]+|$)")

func fieldNameToEnvName(s string) string {
	var a []string
	for _, sub := range camel.FindAllStringSubmatch(s, -1) {
		if sub[1] != "" {
			a = append(a, sub[1])
		}
		if sub[2] != "" {
			a = append(a, sub[2])
		}
	}
	return strings.ToUpper(strings.Join(a, "_"))
}

// preparations for docker
func handleEnvironmentVars(config *Config, getEnv func(key string) string) (err error) {
	configValue := reflect.Indirect(reflect.ValueOf(config))
	configType := configValue.Type()
	for index := 0; index < configType.NumField(); index++ {
		fieldName := configType.Field(index).Name
		fieldConfig := configType.Field(index).Tag.Get("config")
		envName := fieldNameToEnvName(fieldName)
		envValue := getEnv(envName)
		if envValue != "" {
			loggedEnvValue := envValue
			if strings.Contains(fieldConfig, "secret") {
				loggedEnvValue = "***"
			}
			fmt.Println("use environment variable: ", envName, " = ", loggedEnvValue)
			if field := configValue.FieldByName(fieldName); field.Kind() == reflect.Struct && field.CanInterface() {
				fieldPtrInterface := field.Addr().Interface()
				setter, setterOk := fieldPtrInterface.(interface{ SetString(string) })
				errSetter, errSetterOk := fieldPtrInterface.(interface{ SetString(string) error })
				if setterOk {
					setter.SetString(envValue)
				}
				if errSetterOk {
					err = errSetter.SetString(envValue)
					if err != nil {
						return fmt.Errorf("invalid env variable %v=%v: %w", envName, envValue, err)
					}
				}
			}
			if configValue.FieldByName(fieldName).Kind() == reflect.Int64 || configValue.FieldByName(fieldName).Kind() == reflect.Int {
				i, err := strconv.ParseInt(envValue, 10, 64)
				if err != nil {
					return fmt.Errorf("invalid env variable %v=%v: %w", envName, envValue, err)
				}
				configValue.FieldByName(fieldName).SetInt(i)
			}
			if configValue.FieldByName(fieldName).Kind() == reflect.String {
				configValue.FieldByName(fieldName).SetString(envValue)
			}
			if configValue.FieldByName(fieldName).Kind() == reflect.Bool {
				b, err := strconv.ParseBool(envValue)
				if err != nil {
					return fmt.Errorf("invalid env variable %v=%v: %w", envName, envValue, err)
				}
				configValue.FieldByName(fieldName).SetBool(b)
			}
			if configValue.FieldByName(fieldName).Kind() == reflect.Float64 {
				f, err := strconv.ParseFloat(envValue, 64)
				if err != nil {
					return fmt.Errorf("invalid env variable %v=%v: %w", envName, envValue, err)
				}
				configValue.FieldByName(fieldName).SetFloat(f)
			}
			if configValue.FieldByName(fieldName).Kind() == reflect.Slice {
				val := []string{}
				for _, element := range strings.Split(envValue, ",") {
					val = append(val, strings.TrimSpace(element))
				}
				configValue.FieldByName(fieldName).Set(reflect.ValueOf(val))
			}
			if configValue.FieldByName(fieldName).Kind() == reflect.Map {
				value := map[string]string{}
				for _, element := range strings.Split(envValue, ",") {
					keyVal := strings.Split(element, ":")
					key := strings.TrimSpace(keyVal[0])
					val := strings.TrimSpace(keyVal[1])
					value[key] = val
				}
				configValue.FieldByName(fieldName).Set(reflect.ValueOf(value))
			}
		}
	}
	return nil
}

type Duration struct {
	dur time.Duration
}

func (this *Duration) GetDuration() time.Duration {
	return this.dur
}

func (this *Duration) SetDuration(dur time.Duration) {
	this.dur = dur
}

func (this *Duration) SetString(str string) error {
	if str == "" {
		return nil
	}
	duration, err := time.ParseDuration(str)
	if err != nil {
		return err
	}
	this.SetDuration(duration)
	return nil
}

func (this *Duration) UnmarshalJSON(bytes []byte) (err error) {
	var str string
	err = json.Unmarshal(bytes, &str)
	if err != nil {
		return err
	}
	return this.SetString(str)
}
