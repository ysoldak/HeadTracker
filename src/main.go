package main

import (
	"math"
	"runtime"
	"time"

	"github.com/ysoldak/HeadTracker/src/display"
	"github.com/ysoldak/HeadTracker/src/orientation"
	"github.com/ysoldak/HeadTracker/src/trainer"
)

var Version string

const (
	PERIOD           = 20
	BLINK_MAIN_COUNT = 500
	BLINK_WARM_COUNT = 125
	BLINK_PARA_COUNT = 250
	TRACE_COUNT      = 1000
	MEMORY_COUNT     = -1 // disabled
)

const (
	radToMs = 500.0 / math.Pi
)

const flashStoreTreshold = 10_000

var (
	d *display.Display
	t trainer.Trainer
	o *orientation.Orientation
	f *Flash
)

func init() {

	initLeds()
	initPins()
	initExtras()

	// Orientation
	o = orientation.New()
	o.Configure(PERIOD * time.Millisecond)

	// Trainer (Bluetooth or PPM)
	if !pinSelectPPM.Get() { // Low means connected to GND => PPM output requested
		t = trainer.NewPPM(pinOutputPPM) // PPM wire
	} else {
		t = trainer.NewPara()
	}

	// Display
	d = display.New()
	d.Version = Version

	d.Configure()
	go d.Run()

	f = &Flash{}

}

func main() {

	// warm up IMU (1 sec)
	for i := 0; i < 50; i++ {
		o.Calibrate()
		time.Sleep(PERIOD * time.Millisecond)
	}

	flashLoad()

	// record initial orientation
	o.Reset()

	// calibrate gyroscope (until stable)
	iter := 0
	for {
		o.Calibrate()
		d.Paired = t.Paired()
		stateMain(iter)
		stateCalibration(iter)
		trace(iter)
		time.Sleep(time.Millisecond)
		iter++
		if iter > 3000 && !f.IsEmpty() { // when had some calibration already, force it if was not able to find better quickly
			o.SetOffsets(f.roll, f.pitch, f.yaw)
			o.SetStable(true)
		}
		if o.Stable() {
			break
		}
	}
	// indicate calibration is over
	off(ledR)

	// store calibration values
	flashStore()

	// enable trainer after flash operations (bluetooth conflicts with flash)
	t.Configure()
	go t.Run()

	// switch display to normal mode
	d.Address = t.Address()
	d.Bluetooth = pinSelectPPM.Get() // high means Bluetooth
	d.Stable = true

	// main loop
	iter = 0
	for {

		if !pinResetCenter.Get() { // Low means button pressed => shall reset center
			o.Reset()
			continue
		}

		o.Update()
		for i, v := range o.Angles() {
			a := angleToChannel(v)
			t.SetChannel(i, a)
			d.Channels[i] = a
		}
		d.Paired = t.Paired()

		// blink and trace
		stateMain(iter)
		statePara(iter)
		trace(iter)
		memory(iter)

		// wait
		time.Sleep(PERIOD * time.Millisecond)
		iter += PERIOD
		iter %= 1_000_000
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

// --- Flash ----

func flashLoad() {
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
	o.SetOffsets(f.roll, f.pitch, f.yaw) // zeroes at worst
}

// Store only when difference is large enough
func flashStore() {
	roll, pitch, yaw := o.Offsets()
	if abs(f.roll-roll) > flashStoreTreshold || abs(f.pitch-pitch) > flashStoreTreshold || abs(f.yaw-yaw) > flashStoreTreshold {
		f.roll, f.pitch, f.yaw = roll, pitch, yaw
		err := f.Store()
		if err != nil {
			println(time.Now().Unix(), err.Error())
		}
	}
}

func abs(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}

// --- Logging ----

func stateMain(iter int) {
	if iter%BLINK_MAIN_COUNT == 0 { // indicate main loop running
		toggle(led)
	}
}

func stateCalibration(iter int) {
	if iter%BLINK_WARM_COUNT == 0 { // indicate warm loop running
		toggle(ledR)
	}
}

func statePara(iter int) {
	if iter%BLINK_PARA_COUNT == 0 { // indicate para (bluetooth) state
		if t.Paired() {
			on(ledB) // on, connected
		} else {
			toggle(ledB) // blink, advertising
		}
	}
}

func trace(iter int) {
	if iter%TRACE_COUNT == 0 { // print out state
		channels := t.Channels()
		r, p, y := channels[0], channels[1], channels[2]
		rc, pc, yc := o.Offsets()
		println(time.Now().Unix(), ": ", t.Address(), " | ", Version, " [", r, ",", p, ",", y, "] (", rc, ",", pc, ",", yc, ")")
	}
}

// to debug memory usage or force GC if needed, disabled by default
func memory(iter int) {
	if MEMORY_COUNT < 0 {
		return
	}
	if iter%MEMORY_COUNT == 0 {
		// runtime.GC()
		ms := runtime.MemStats{}
		runtime.ReadMemStats(&ms)
		println("Used: ", ms.HeapInuse, " Free: ", ms.HeapIdle, " Meta: ", ms.GCSys)
	}
}
