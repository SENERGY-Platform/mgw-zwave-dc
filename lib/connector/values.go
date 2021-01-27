package connector

//expects ids from mgw (with prefixes and suffixes)
func (this *Connector) saveValue(deviceId string, serviceId string, value interface{}) {
	this.valueStoreMux.Lock()
	defer this.valueStoreMux.Unlock()
	this.valueStore[deviceId+"-"+serviceId] = value
}

//expects ids from mgw (with prefixes and suffixes)
func (this *Connector) getValue(deviceId string, serviceId string) (value interface{}, known bool) {
	this.valueStoreMux.Lock()
	defer this.valueStoreMux.Unlock()
	value, known = this.valueStore[deviceId+"-"+serviceId]
	return
}
