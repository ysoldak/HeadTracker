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

	radToMs = 500.0 / math.Pi
)

var (
	d *display.Display
	t trainer.Trainer
	o *orientation.Orientation
)

func init() {

	initLeds()
	initPins()

	// Orientation
	o = orientation.New()
	o.Configure(PERIOD * time.Millisecond)

	// Trainer (Bluetooth or PPM)
	if !pinSelectPPM.Get() { // Low means connected to GND => PPM output requested
		t = trainer.NewPPM(pinOutputPPM) // PPM wire
	} else {
		t = trainer.NewPara()
	}
	t.Configure()
	go t.Run()

	// Display
	d = display.New()
	d.Address = t.Address()
	d.Version = Version
	d.Bluetooth = pinSelectPPM.Get() // High means Bluetooth

	d.Configure()
	go d.Run()

}

func main() {

	// warm up IMU (1 sec)
	for i := 0; i < 50; i++ {
		o.Calibrate()
		time.Sleep(PERIOD * time.Millisecond)
	}

	// record initial orientation
	o.Reset()

	// calibrate gyroscope (until stable)
	iter := 0
	for !o.Stable() {
		o.Calibrate()
		d.Paired = t.Paired()
		state(iter)
		trace(iter)
		time.Sleep(time.Millisecond)
		iter++
	}
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
		state(iter)
		trace(iter)

		// wait
		time.Sleep(PERIOD * time.Millisecond)
		iter += PERIOD

	}

}

// --- Utils -------------------------------------------------------------------

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

func state(iter int) {
	if iter%BLINK_MAIN_COUNT == 0 { // indicate main loop running
		toggle(led)
	}
	if iter%BLINK_WARM_COUNT == 0 { // indicate warm loop running
		if o.Stable() {
			off(ledR) // off, warmed up
		} else {
			toggle(ledR)
		}
	}
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
