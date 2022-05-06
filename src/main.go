package main

import (
	"time"

	"github.com/ysoldak/HeadTracker/src/display"
	"github.com/ysoldak/HeadTracker/src/orientation"
	"github.com/ysoldak/HeadTracker/src/trainer"
)

const (
	angleMax         = 45
	PERIOD           = 20
	BLINK_MAIN_COUNT = 500
	BLINK_WARM_COUNT = 125
	BLINK_PARA_COUNT = 250
	TRACE_COUNT      = 1000
)

var (
	d *display.Display
	t *trainer.Trainer
	o *orientation.Orientation
)

func init() {

	// Orientation
	o = orientation.New()
	o.Configure(PERIOD * time.Millisecond)

	// Bluetooth (FrSKY's PARA trainer protocol)
	t = trainer.New()
	t.Configure()
	go t.Run(PERIOD * time.Millisecond)

	// Display
	d = display.New()
	d.Address = t.Address
	d.Version = version
	d.Configure()
	go d.Run(5 * PERIOD * time.Millisecond)

}

func main() {

	// warm up IMU
	for i := 0; i < 50; i++ {
		o.Update(false)
		time.Sleep(PERIOD * time.Millisecond)
	}

	// record initial orientation
	o.Center()

	// calibrate gyroscope
	iter := 0
	for !o.Stable() {

		o.Update(false)
		d.Paired = t.Paired
		state(iter)

		time.Sleep(time.Millisecond)
		iter++
	}
	d.Stable = true

	// main loop
	iter = 0
	for {

		o.Update(true)
		for i, v := range o.Angles() {
			a := angleToChannel(v, 45)
			t.Channels[i] = a
			d.Channels[i] = a
		}
		d.Paired = t.Paired

		// blink and trace
		state(iter)
		trace(iter)

		// wait
		time.Sleep(PERIOD * time.Millisecond)
		iter += PERIOD

	}

}

// --- Utils -------------------------------------------------------------------

func angleToChannel(angle float32, max float32) uint16 {
	result := uint16(1500 + 500/max*angle)
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
		if t.Paired {
			on(ledB) // on, connected
		} else {
			toggle(ledB) // blink, advertising
		}
	}
}

func trace(iter int) {
	if iter%TRACE_COUNT == 0 { // print out state
		r, p, y := t.Channels[0], t.Channels[1], t.Channels[2]
		rc, pc, yc := o.Offsets()
		println(time.Now().Unix(), ": ", t.Address, " [", r, ",", p, ",", y, "] (", rc, ",", pc, ",", yc, ")")
	}
}
