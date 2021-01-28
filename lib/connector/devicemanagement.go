package connector

import (
	"log"
	"runtime/debug"
	"zwave2mqtt-connector/lib/mgw"
	"zwave2mqtt-connector/lib/zwave2mqtt"
)

func (this *Connector) NotifyRefresh() {
	err := this.z2mClient.RequestDeviceInfoUpdate()
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
	}
	if this.updateTicker != nil {
		this.updateTicker.Reset(this.updateTickerDuration)
	}
}

func (this *Connector) DeviceInfoListener(nodes []zwave2mqtt.DeviceInfo, withValues bool, allKnownDevices bool) {
	deviceInfos := map[string]mgw.DeviceInfo{}
	for _, node := range nodes {
		id, info, err := this.nodeToDeviceInfo(node)
		if err != nil {
			log.Println("WARNING: unable to create device info for node", err)
			continue
		}
		err = this.registerDevice(id, info)
		if err != nil {
			log.Println("ERROR: unable to register device", err)
			return
		}
		if withValues {
			for _, value := range node.Values {
				this.ValueEventListener(value)
			}
		}
	}
	if allKnownDevices {
		this.unregisterMissingDevices(deviceInfos)
	}
}

func (this *Connector) registerDevice(id string, info mgw.DeviceInfo) (err error) {
	err = this.mgwClient.SetDevice(id, info)
	if err != nil {
		log.Println("ERROR: unable to send device info to mgw", err)
		return err
	}
	err = this.mgwClient.ListenToDeviceCommands(id, this.CommandHandler)
	if err != nil {
		log.Println("ERROR: unable to subscribe to device commands", err)
		return err
	}
	this.deviceRegisterSet(id, info)
	return nil
}

func (this *Connector) deviceRegisterSet(id string, info mgw.DeviceInfo) {
	this.deviceRegisterMux.Lock()
	defer this.deviceRegisterMux.Unlock()
	this.deviceRegister[id] = info
}

func (this *Connector) deviceRegisterRemove(id string) {
	this.deviceRegisterMux.Lock()
	defer this.deviceRegisterMux.Unlock()
	delete(this.deviceRegister, id)
}

func (this *Connector) deviceRegisterGet(id string) (info mgw.DeviceInfo, ok bool) {
	this.deviceRegisterMux.Lock()
	defer this.deviceRegisterMux.Unlock()
	info, ok = this.deviceRegister[id]
	return
}

func (this *Connector) deviceRegisterGetIds() (ids []string) {
	this.deviceRegisterMux.Lock()
	defer this.deviceRegisterMux.Unlock()
	for id, _ := range this.deviceRegister {
		ids = append(ids, id)
	}
	return
}

func (this *Connector) unregisterMissingDevices(infos map[string]mgw.DeviceInfo) {
	for _, id := range this.deviceRegisterGetIds() {
		info, found := infos[id]
		if !found {
			info.State = mgw.Offline
			err := this.mgwClient.SetDevice(id, info)
			if err != nil {
				log.Println("ERROR: unable to send device info (offline) to mgw", err)
				return
			}
			err = this.mgwClient.StopListenToDeviceCommands(id)
			if err != nil {
				log.Println("WARNING: unable to stop listening to device commands", err)
			}
			this.deviceRegisterRemove(id)
		}
	}
}
