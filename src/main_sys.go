package main

import (
	"machine"
	"time"
)

// implements trainer.SystemHandler interface
type SystemHandler struct {
	name           string
	channelsConfig [3]byte // offset, order, enabled
}

func NewSystemHandler() *SystemHandler {
	return &SystemHandler{
		name: "HT",
	}
}

func (sh *SystemHandler) OrientationReset() {
	println("MAIN: Orientation reset")
	o.Reset() // FIXME: also blink red led
}

func (sh *SystemHandler) FactoryReset() {
	println("MAIN: Factory reset requested, resetting flash data and restarting...")
	f = NewFlash() // reset flash object
	f.Store()      // store empty object
	time.Sleep(1 * time.Second)
	machine.CPUReset() // restart device
}

func (sh *SystemHandler) Reboot() {
	println("MAIN: Reboot requested, restarting...")
	time.Sleep(1 * time.Second)
	machine.CPUReset()
}

func (sh *SystemHandler) Name() string {
	return sh.name
}

func (sh *SystemHandler) SetName(newName string) {
	if newName != sh.name {
		println("FLASH: Device name write:", newName)
		sh.name = newName
		f.SetName(newName)
		err := f.Store()
		if err != nil {
			println("FLASH: store error:", err.Error())
		}
	}
}

func (sh *SystemHandler) ChannelsConfig() (byte, byte, byte) {
	return sh.channelsConfig[0], sh.channelsConfig[1], sh.channelsConfig[2]
}

func (sh *SystemHandler) SetChannelsConfig(offset byte, order byte, enabled byte) {
	println("FLASH: Channels configuration write: offset =", offset, " order =", order, " enabled =", enabled)
	sh.channelsConfig[0] = offset
	sh.channelsConfig[1] = order
	sh.channelsConfig[2] = enabled
	f.channels[0] = offset
	f.channels[1] = order
	f.channels[2] = enabled
	err := f.Store()
	if err != nil {
		println("FLASH: store error:", err.Error())
	}
}
