package main

import "machine"

var pinResetCenter = machine.D2
var pinOutputPPM = machine.D8

func initPins() {
	pinResetCenter.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	pinOutputPPM.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
}
