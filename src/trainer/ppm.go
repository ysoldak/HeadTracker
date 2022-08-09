package trainer

import (
	"device/arm"
	"machine"
)

// --- Confgurable ------------------------------------------------------------
var ppmPin = machine.D10

// --- Implementation ---------------------------------------------------------

var ppmFrameLen uint32 = 22500 * sysCyclesPerMicrosecond
var ppmOffLen uint32 = 300 * sysCyclesPerMicrosecond

var sysCyclesPerMicrosecond = machine.CPUFrequency() / 1_000_000

var ppmInstance PPM

type PPM struct {
	curChan  int8
	Channels [8]uint16
}

func NewPPM() *PPM {
	ppmInstance = PPM{
		curChan:  -1,
		Channels: [8]uint16{1500, 1500, 1500, 1500, 1500, 1500, 1500, 1500},
	}
	return &ppmInstance
}

func (ppm *PPM) Configure() {
	ppmPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	ppmPin.Low()
}

func (ppm *PPM) Run() {
	arm.SetupSystemTimer(sysCyclesPerMicrosecond)
}

// --- Interrupt Handler ------------------------------------------------------

//export SysTick_Handler
func timer_isr() {
	// separator
	if ppmPin.Get() {
		ppmPin.Low()
		ppmInstance.curChan++
		if ppmInstance.curChan > 7 {
			ppmInstance.curChan = -1
		}
		arm.SetupSystemTimer(ppmOffLen)
		return
	}
	// regular channel
	if ppmInstance.curChan != -1 {
		ppmPin.High()
		arm.SetupSystemTimer(uint32(ppmInstance.Channels[ppmInstance.curChan])*sysCyclesPerMicrosecond - ppmOffLen)
		return
	}
	// padding
	ppmPin.High()
	sum := uint16(0)
	for _, value := range ppmInstance.Channels {
		sum += value
	}
	arm.SetupSystemTimer(ppmFrameLen - uint32(sum)*sysCyclesPerMicrosecond - 8*ppmOffLen)
}
