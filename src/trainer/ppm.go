package trainer

// PPM (Pulse-position modulation) wired trainer link

import (
	"device/arm"
	"machine"
)

var ppmFrameLen uint32 = 22500 * sysCyclesPerMicrosecond
var ppmOffLen uint32 = 300 * sysCyclesPerMicrosecond

var sysCyclesPerMicrosecond = machine.CPUFrequency() / 1_000_000

var ppmInstance PPM

type PPM struct {
	pin      machine.Pin
	curChan  int8
	channels [8]uint16
}

func NewPPM(pin machine.Pin) *PPM {
	ppmInstance = PPM{
		pin:      pin,
		curChan:  -1,
		channels: [8]uint16{1500, 1500, 1500, 1500, 1500, 1500, 1500, 1500},
	}
	return &ppmInstance
}

func (ppm *PPM) Configure() {
	ppm.pin.Low()
}

func (ppm *PPM) Run() {
	arm.SetupSystemTimer(sysCyclesPerMicrosecond)
}

func (ppm *PPM) Paired() bool {
	return true
}

func (ppm *PPM) Address() string {
	return "    PPM OUTPUT"
}

func (ppm *PPM) Channels() [8]uint16 {
	return ppm.channels
}

func (ppm *PPM) SetChannel(n int, v uint16) {
	ppm.channels[n] = v
}

// --- Interrupt Handler ------------------------------------------------------

//export SysTick_Handler
func timer_isr() {
	// separator
	if ppmInstance.pin.Get() {
		ppmInstance.pin.Low()
		ppmInstance.curChan++
		if ppmInstance.curChan > 7 {
			ppmInstance.curChan = -1
		}
		arm.SetupSystemTimer(ppmOffLen)
		return
	}
	// regular channel
	if ppmInstance.curChan != -1 {
		ppmInstance.pin.High()
		arm.SetupSystemTimer(uint32(ppmInstance.channels[ppmInstance.curChan])*sysCyclesPerMicrosecond - ppmOffLen)
		return
	}
	// padding
	ppmInstance.pin.High()
	sum := uint16(0)
	for _, value := range ppmInstance.channels {
		sum += value
	}
	arm.SetupSystemTimer(ppmFrameLen - uint32(sum)*sysCyclesPerMicrosecond - 8*ppmOffLen)
}
