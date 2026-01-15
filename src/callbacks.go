package main

import (
	"machine"
	"time"
)

type BluetoothCallbackHandler struct {
}

func (b *BluetoothCallbackHandler) OnConnect() {
	println("Bluetooth connected")
	state.connected = true
}

func (b *BluetoothCallbackHandler) OnDisconnect() {
	println("Bluetooth disconnected")
	state.connected = false
}

func (b *BluetoothCallbackHandler) OnOrientationReset() {
	println("Orientation reset via Bluetooth command")
	o.Reset()
}

func (b *BluetoothCallbackHandler) OnFactoryReset() {
	println("Factory reset via Bluetooth command")
	f = NewFlash() // reset flash object
	f.Store()      // store default flash data
	time.Sleep(1 * time.Second)
	machine.CPUReset()
}

func (b *BluetoothCallbackHandler) OnReboot() {
	println("Reboot via Bluetooth command")
	time.Sleep(1 * time.Second)
	machine.CPUReset()
}

func (b *BluetoothCallbackHandler) OnDeviceNameChange(name string) {
	println("Device name changed to", name)
	state.deviceName = name
}

func (b *BluetoothCallbackHandler) OnAxisMappingChange(mapping [3]byte) {
	println("Axis mapping changed to", mapping[0], mapping[1], mapping[2])
	state.axisMapping = mapping
}
