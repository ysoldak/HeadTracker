package main

import (
	"math"
	"runtime"
	"strconv"
	"time"

	"github.com/ysoldak/HeadTracker/src/display"
	"github.com/ysoldak/HeadTracker/src/orientation"
	"github.com/ysoldak/HeadTracker/src/trainer"
)

var Version string

const (
	PERIOD           = 20     // 20000us -- budget for main loop to ensure stable timing for sensor fusion
	DISPLAY_COUNT    = 100    // update display every 100ms, with offset of one period to avoid clashing with tracing
	FLASH_COUNT      = 30_000 // try dump state to flash every 30 seconds
	BLINK_MAIN_COUNT = 500    // main loop indicator
	BLINK_WARM_COUNT = 100    // warm up / calibration indicator
	BLINK_PARA_COUNT = 200    // para (bluetooth) state indicator
	TRACE_COUNT      = 1_000  // tracing to serial output, every 1 second
)

const (
	radToMs = 512.0 / math.Pi
)

const flashStoreTreshold = 100_000

var (
	d *display.Display
	t Trainer
	i *orientation.IMU
	o *orientation.Orientation
	f *Flash
	h *BluetoothCallbackHandler
)

type Trainer interface {
	Start() string
	SetChannel(num int, v uint16)
}

var (
	tickPeriod *time.Ticker
)

var state struct {
	address    string
	channels   [3]uint16
	connected  bool
	deviceName string
}

func init() {

	initLeds()
	initPins()
	initExtras()

	// Orientation
	i = orientation.NewIMU()
	o = orientation.New(i)
	err := o.Configure(PERIOD * time.Millisecond)
	if err != nil {
		for {
			println("IMU configuration error:", err.Error())
			time.Sleep(1 * time.Second)
		}
	}

	state.channels = [3]uint16{1500, 1500, 1500}
	state.address = "--:--:--:--:--:--"

	// Display
	d = display.New()
	d.Configure()

	f = NewFlash()

	tickPeriod = time.NewTicker(PERIOD * time.Millisecond)

}

func main() {

	batVolts, err := batteryVoltage()
	batString := ""
	if err == nil {
		batString = strconv.FormatFloat(batVolts, 'f', 2, 64) + "V"
	}
	d.AddText(0, "Head Tracker "+batString)
	d.AddText(1, Version+" @ysoldak")
	d.Update()

	// warm up IMU (1 sec)
	for i := 0; i < 50; i++ {
		o.Calibrate()
		<-tickPeriod.C
	}

	loadState()

	// record initial orientation
	o.Reset()

	// calibrate gyroscope (until stable)
	waitText := "Calibrating"
	if !f.IsEmpty() {
		waitText = "Loading" // secondary calibration is short, just show "Loading" in that case
	}
	d.RemoveText(nil)
	d.SetTextBlinkFunc(d.AddText(1, waitText+"   "), waitText+"...", func() bool { return true })
	d.Update()
	prev := [3]int32{0, 0, 0}
	directions := [3]int32{1, 1, 1}
	maxCorrection := int32(2_000_000)
	iter := uint16(0)
	stopTime := time.Now().Add(3 * time.Second)
	for {

		for i, v := range o.Calibrate() {
			if v == 0 && f.IsEmpty() {
				v = maxCorrection // for better visualisation at start
			}
			max := abs(1000*v) / maxCorrection
			value := prev[i] + directions[i]
			if value < 0 {
				value = 0
				directions[i] *= -1
			}
			if value > max {
				value = max
				directions[i] *= -1
			}
			prev[i] = value
			d.SetBar(byte(i), int16(value)/20, true)
		}

		blinkMain(iter)
		blinkCalibration(iter)
		printState(iter)

		time.Sleep(time.Millisecond)
		iter++
		iter %= 10_000

		if iter%DISPLAY_COUNT == 0 {
			d.Update()
		}

		if time.Now().After(stopTime) && !f.IsEmpty() { // when had some calibration already, force it if was not able to find better quickly
			o.SetOffsets(f.gyrCalOffsets)
			o.SetStable(true)
		}
		if o.Stable() {
			break
		}
	}

	// indicate calibration is over
	off(ledR)

	// store calibration values (0 to force store now)
	storeState(0)

	// Trainer (Bluetooth or PPM)
	if !pinSelectPPM.Get() { // Low means connected to GND => PPM output requested
		t = trainer.NewPPM(pinOutputPPM) // PPM wire
		state.connected = true
	} else {
		t = trainer.NewPara(state.deviceName, &BluetoothCallbackHandler{})
		state.connected = false
	}
	state.address = t.Start()

	// switch display to normal mode
	d.RemoveText(nil)
	d.AddText(1, state.address)
	if pinSelectPPM.Get() { // high means Bluetooth
		d.SetTextBlinkFunc(d.AddText(1, "  :  :  :  :  :  "), "", func() bool { return !state.connected })
	}
	d.Update()

	// main loop
	iter = 0
	for range tickPeriod.C {

		pinDebugMain.Set(!pinDebugMain.Get())

		// check for reset request
		if !pinResetCenter.Get() || (iter%400 == 0 && i.ReadTap()) { // Button pressed OR [double] tap registered (shall not read register more frequently than double tap duration)
			o.Reset()
			println("Orientation reset via pin or double tap")
		}

		// update orientation, every 20ms (~2360us)
		pinDebugData.High()
		o.Update()
		pinDebugData.Low()

		// set channels, every 20ms (~300us)
		for i, a := range o.Angles() {
			state.channels[i] = angleToChannel(a)
			t.SetChannel(i, state.channels[i])
			d.SetBar(byte(i), int16(1500-state.channels[i])/10, false)
		}

		// update display, every 100ms (~15000us)
		updateDisplay(iter + PERIOD) // slow (when display is connected, shall not clash with anything else, so offset by one period)

		// handle state, period and performance varies
		blinkMain(iter)  // very fast
		blinkPara(iter)  // very fast
		storeState(iter) // very slow (~85300us, can affect sensor fusion if executed too often; as it is so slow no point to offset it)
		printState(iter) // fast (~1500us)

		iter += PERIOD
		iter %= 60_000
	}

}

// --- Utils -------------------------------------------------------------------

// --- Core ----

func angleToChannel(angle float64) uint16 {
	result := uint16(1500 + angle*radToMs)
	if result < 988 {
		return 988
	}
	if result > 2012 {
		return 2012
	}
	return result
}

// --- Display ----

// Update display, slow operation when display is connected (~15000us)
func updateDisplay(iter uint16) {
	if iter%DISPLAY_COUNT != 0 {
		return
	}
	pinDebugData.High()
	defer pinDebugData.Low()
	d.Update()
}

// --- State ----

func loadState() {
	// reset calibration data when button is pressed
	resetGyrCalOffsets := !pinResetCenter.Get()
	// wait until button is released
	for !pinResetCenter.Get() {
		time.Sleep(10 * time.Millisecond)
	}
	// clear data on flash, "f" object is empty at this point
	if resetGyrCalOffsets {
		err := f.Store()
		if err != nil {
			println(time.Now().Unix(), err.Error())
		}
	}

	// load calibration data, can be empty
	err := f.Load()
	if err != nil {
		println(time.Now().Unix(), err.Error())
	}

	// set offsets, they are either actual previous calibration result or zeroes inially and in case of error
	o.SetOffsets(f.gyrCalOffsets) // zeroes at worst

	// set name
	n := 0
	for n < len(f.deviceName) && f.deviceName[n] != 0 {
		n++
	}
	state.deviceName = string(f.deviceName[:n])
}

// Store current configuration & calibration to flash (~85300us)
// The operation is slow and flash has limited number of write cycles,
// so only do this when difference is large enough and not too often.
func storeState(iter uint16) {
	if iter%FLASH_COUNT != 0 {
		return
	}

	pinDebugData.High()
	defer pinDebugData.Low()

	mustStore := false
	for i := range o.Offsets() {
		if abs(f.gyrCalOffsets[i]-o.Offsets()[i]) > flashStoreTreshold {
			f.gyrCalOffsets[i] = o.Offsets()[i]
			mustStore = true
		}
	}

	newName := false
	for i := range len(f.deviceName) {
		b := byte(0)
		if i < len(state.deviceName) {
			b = byte(state.deviceName[i])
		}
		if f.deviceName[i] != b {
			newName = true
		}
		if newName {
			f.deviceName[i] = b
			mustStore = true
		}
	}

	if !mustStore {
		return
	}

	err := f.Store()
	if err != nil {
		println("Flash error:", err.Error())
	}
}

func abs(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}

// --- Logging ----

// indicate main loop running
func blinkMain(iter uint16) {
	if iter%BLINK_MAIN_COUNT == 0 {
		toggle(led)
	}
}

// indicate warm loop running
func blinkCalibration(iter uint16) {
	if iter%BLINK_WARM_COUNT == 0 {
		toggle(ledR)
	}
}

// indicate para (bluetooth) state
func blinkPara(iter uint16) {
	if iter%BLINK_PARA_COUNT != 0 {
		return
	}
	if !pinSelectPPM.Get() { // PPM mode
		off(ledB)
		return
	}
	if state.connected {
		on(ledB) // on, connected
	} else {
		toggle(ledB) // blink, advertising
	}
}

var ms = runtime.MemStats{}

// Print out state (~1500us)
func printState(iter uint16) {
	if iter%TRACE_COUNT != 0 {
		return
	}
	pinDebugData.High()
	ch0, ch1, ch2 := state.channels[0], state.channels[1], state.channels[2]
	cal := o.Offsets()
	runtime.ReadMemStats(&ms)
	println(state.deviceName, Version, "|", state.address, "| [", ch0, ",", ch1, ",", ch2, "] (", cal[0], ",", cal[1], ",", cal[2], ")", ms.HeapInuse)
	pinDebugData.Low()
}
