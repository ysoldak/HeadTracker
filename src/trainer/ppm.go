package trainer

// PPM (Pulse-position modulation) wired trainer link
// http://flyingeinstein.com/index.php/articles/58-ppm-explained

import (
	"device/nrf"
	"machine"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

const ppmFrameLength = 22500
const ppmSpacerLength = 300

const (
	ppmTimerLowEndOfChannelsIdx    = 3
	ppmTimerLowEndOfFrameSpacerIdx = 4
	ppmTimerLowEndOfFrameIdx       = 5
)
const (
	ppmUpdateInterruptId        = nrf.IRQ_TIMER3
	ppmUpdateInterruptCondition = nrf.TIMER_INTENSET_COMPARE3
)
const (
	ppmTimerLowShorts  = nrf.TIMER_SHORTS_COMPARE5_CLEAR // 5th compare register is end of frame
	ppmTimerHighShorts = nrf.TIMER_SHORTS_COMPARE0_CLEAR | nrf.TIMER_SHORTS_COMPARE0_STOP
)

const (
	ppmTimerPrescaler            = 2
	ppmTimerCyclesPerMicrosecond = (16 >> ppmTimerPrescaler) * 0.995 // = 16Mhz / 2^prescaler / 1'000'000 * inaccuracy coefficient
)

var ppmTimerLow *nrf.TIMER_Type = nrf.TIMER3
var ppmTimerHigh *nrf.TIMER_Type = nrf.TIMER4

var ppmInstance PPM

type PPM struct {
	pin      machine.Pin
	channels [3]uint16
}

func NewPPM(pin machine.Pin) *PPM {
	ppmInstance = PPM{
		pin:      pin,
		channels: [3]uint16{1500, 1500, 1500},
	}
	return &ppmInstance
}

func (ppm *PPM) Configure() {
	ppm.pin.Low()
	configurePin()
	configureTimers()
	configurePpi()
}

func (ppm *PPM) Run() {
	ppmTimerLow.TASKS_START.Set(1)
}

func (ppm *PPM) Paired() bool {
	return true
}

func (ppm *PPM) Address() string {
	return "    PPM OUTPUT"
}

func (ppm *PPM) Channels() []uint16 {
	return ppm.channels[:]
}

func (ppm *PPM) SetChannel(n int, v uint16) {
	ppm.channels[n] = v
}

// --- Configure --------------------------------------------------------------

func configurePin() {
	// Configure a GPIOTE channel.
	nrf.GPIOTE.CONFIG[0].Set(
		(nrf.GPIOTE_CONFIG_MODE_Task << nrf.GPIOTE_CONFIG_MODE_Pos) |
			(uint32(ppmInstance.pin) << nrf.GPIOTE_CONFIG_PSEL_Pos) |
			(nrf.GPIOTE_CONFIG_POLARITY_None << nrf.GPIOTE_CONFIG_POLARITY_Pos) |
			(nrf.GPIOTE_CONFIG_OUTINIT_Low << nrf.GPIOTE_CONFIG_OUTINIT_Pos))
}

func configureTimers() {

	// Timer that pulls pin low
	ppmTimerLow.TASKS_STOP.Set(1)
	ppmTimerLow.PRESCALER.Set(ppmTimerPrescaler)
	ppmTimerLow.SHORTS.Set(ppmTimerLowShorts)                // reset channel timer on frame's end
	ppmTimerLow.BITMODE.Set(nrf.TIMER_BITMODE_BITMODE_32Bit) // low prescaler => more precision but large counters

	// Spacers at the ends of channels
	offset := uint16(0)
	for i, v := range ppmInstance.channels {
		offset += v
		ppmTimerLow.CC[i].Set(microToCount(offset - ppmSpacerLength))
	}
	// All channels are sent => an interrupt fires that updates timer counter registers with new values
	ppmTimerLow.CC[ppmTimerLowEndOfChannelsIdx].Set(microToCount(offset))
	// Spacer at the end of frame
	ppmTimerLow.CC[ppmTimerLowEndOfFrameSpacerIdx].Set(microToCount(ppmFrameLength - ppmSpacerLength))
	// End of frame
	ppmTimerLow.CC[ppmTimerLowEndOfFrameIdx].Set(microToCount(ppmFrameLength))

	// Trigger update values handler
	ppmTimerLow.INTENSET.Set(ppmUpdateInterruptCondition)
	itr := interrupt.New(ppmUpdateInterruptId, updateDelays)
	itr.SetPriority(0x01)
	itr.Enable()

	// ------------------------

	// Timer that pulls pin high
	ppmTimerHigh.TASKS_STOP.Set(1)
	ppmTimerHigh.PRESCALER.Set(ppmTimerPrescaler)
	ppmTimerHigh.SHORTS.Set(ppmTimerHighShorts)               // reset and stop spacer timer automatically
	ppmTimerHigh.BITMODE.Set(nrf.TIMER_BITMODE_BITMODE_32Bit) // low prescaler => more precision but large counters

	// Spacer length
	ppmTimerHigh.CC[0].Set(microToCount(ppmSpacerLength))

}

func configurePpi() {

	// Pull pin down after each channel and start spacer timer to pull it back up again
	for i := range ppmInstance.channels {
		configurePpiChannel(i*2, &ppmTimerLow.EVENTS_COMPARE[i], &nrf.GPIOTE.TASKS_CLR[0])
		configurePpiChannel(i*2+1, &ppmTimerLow.EVENTS_COMPARE[i], &ppmTimerHigh.TASKS_START)
	}

	// Pull pin down near frame's end and start spacer timer to pull it back up again
	configurePpiChannel(6, &ppmTimerLow.EVENTS_COMPARE[ppmTimerLowEndOfFrameSpacerIdx], &nrf.GPIOTE.TASKS_CLR[0])
	configurePpiChannel(7, &ppmTimerLow.EVENTS_COMPARE[ppmTimerLowEndOfFrameSpacerIdx], &ppmTimerHigh.TASKS_START)

	// Pull pin up on spacer timer event
	configurePpiChannel(8, &ppmTimerHigh.EVENTS_COMPARE[0], &nrf.GPIOTE.TASKS_SET[0])

}

func configurePpiChannel(n int, event, task *volatile.Register32) {
	nrf.PPI.CHENSET.Set(1 << n)
	nrf.PPI.CH[n].EEP.Set(uint32(uintptr(unsafe.Pointer(event))))
	nrf.PPI.CH[n].TEP.Set(uint32(uintptr(unsafe.Pointer(task))))
}

// --- Interrupt Handler ------------------------------------------------------

func updateDelays(itr interrupt.Interrupt) {
	if ppmTimerLow.EVENTS_COMPARE[ppmTimerLowEndOfChannelsIdx].Get() != 0 {
		ppmTimerLow.EVENTS_COMPARE[ppmTimerLowEndOfChannelsIdx].Set(0)
		offset := uint16(0)
		for i, v := range ppmInstance.channels {
			offset += v
			ppmTimerLow.CC[i].Set(microToCount(offset - ppmSpacerLength))
		}
	}
}

func microToCount(micro uint16) uint32 {
	return uint32(float64(micro) * ppmTimerCyclesPerMicrosecond)
}
