package main

import "machine"

var (
	pinResetCenter = machine.D2
	pinSelectPPM   = machine.D8
	pinOutputPPM   = machine.D10
)

func initPins() {
	pinResetCenter.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	pinSelectPPM.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	pinOutputPPM.Configure(machine.PinConfig{Mode: machine.PinOutput})
}
