package main

import (
	"math"
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
)

const (
	radToMs = 512.0 / math.Pi
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
	d.Configure()
	go d.Run()

	f = &Flash{}

}

func main() {

	d.AddText(0, "Head Tracker")
	d.AddText(1, Version+" @ysoldak")

	// warm up IMU (1 sec)
	for i := 0; i < 50; i++ {
		o.Calibrate()
		time.Sleep(PERIOD * time.Millisecond)
	}

	flashLoad()

	// record initial orientation
	o.Reset()

	// calibrate gyroscope (until stable)
	d.RemoveText(nil)
	d.SetTextBlink(d.AddText(1, "Calibrating   "), "Calibrating...", true)
	prev := [3]int32{0, 0, 0}
	directions := [3]int32{1, 1, 1}
	maxCorrection := int32(2_000_000)
	iter := uint16(0)
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

		stateMain(iter)
		stateCalibration(iter)
		trace(iter)

		time.Sleep(time.Millisecond)
		iter++
		iter %= 10_000

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
	d.RemoveText(nil)
	d.AddText(1, t.Address())
	if pinSelectPPM.Get() { // high means Bluetooth
		d.SetTextBlinkFunc(d.AddText(1, "  :  :  :  :  :  "), "", func() bool { return !t.Paired() })
	}

	// main loop
	iter = 0
	for {

		if !pinResetCenter.Get() { // Low means button pressed => shall reset center
			o.Reset()
			continue
		}

		o.Update()
		for i, a := range o.Angles() {
			c := angleToChannel(a)
			t.SetChannel(i, c)
			d.SetBar(byte(i), int16(1500-c)/10, false)
		}

		// blink and trace
		stateMain(iter)
		statePara(iter)
		trace(iter)

		// wait
		time.Sleep(PERIOD * time.Millisecond)
		iter += PERIOD
		iter %= 10_000
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

func stateMain(iter uint16) {
	if iter%BLINK_MAIN_COUNT == 0 { // indicate main loop running
		toggle(led)
	}
}

func stateCalibration(iter uint16) {
	if iter%BLINK_WARM_COUNT == 0 { // indicate warm loop running
		toggle(ledR)
	}
}

func statePara(iter uint16) {
	if iter%BLINK_PARA_COUNT == 0 { // indicate para (bluetooth) state
		if t.Paired() {
			on(ledB) // on, connected
		} else {
			toggle(ledB) // blink, advertising
		}
	}
}

func trace(iter uint16) {
	if iter%TRACE_COUNT == 0 { // print out state
		channels := t.Channels()
		r, p, y := channels[0], channels[1], channels[2]
		rc, pc, yc := o.Offsets()
		println(time.Now().Unix(), ": ", t.Address(), " | ", Version, " [", r, ",", p, ",", y, "] (", rc, ",", pc, ",", yc, ")")
	}
}
