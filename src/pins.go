package main

import "machine"

var (
	pinDebugMain   = machine.P0_02
	pinDebugData   = machine.P0_03
	pinResetCenter = machine.D2
	pinSelectPPM   = machine.D8
	pinOutputPPM   = machine.D10
)

func initPins() {
	pinDebugMain.Configure(machine.PinConfig{Mode: machine.PinOutput})
	pinDebugData.Configure(machine.PinConfig{Mode: machine.PinOutput})
	pinResetCenter.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	pinSelectPPM.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	pinOutputPPM.Configure(machine.PinConfig{Mode: machine.PinOutput})
}
