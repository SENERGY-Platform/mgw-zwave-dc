package connector

import (
	"errors"
	"strconv"
	"zwave2mqtt-connector/lib/mgw"
	"zwave2mqtt-connector/lib/zwave2mqtt"
)

//result id with prefix
func (this *Connector) nodeToDeviceInfo(node zwave2mqtt.DeviceInfo) (id string, info mgw.DeviceInfo, err error) {
	id = this.nodeIdToDeviceId(node.NodeId)
	info = mgw.DeviceInfo{
		Name:  node.Name,
		State: mgw.Online,
	}
	if info.Name == "" {
		info.Name = getDefaultName(node)
	}
	var known bool
	typeMappingKey := this.getTypeMappingKey(node)
	info.DeviceType, known = this.deviceTypeMapping[typeMappingKey]
	if !known {
		err = errors.New("no known mapping for " + typeMappingKey)
		return
	}
	return
}

func getDefaultName(node zwave2mqtt.DeviceInfo) string {
	return node.Product + " (" + strconv.FormatInt(node.NodeId, 10) + ")"
}

func (this *Connector) getTypeMappingKey(node zwave2mqtt.DeviceInfo) string {
	return node.ManufacturerId + "." + node.ProductType + "." + node.ProductId
}
