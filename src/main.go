package main

import (
	"time"

	"github.com/ysoldak/HeadTracker/src/display"
	"github.com/ysoldak/HeadTracker/src/orientation"
	"github.com/ysoldak/HeadTracker/src/trainer"
)

const (
	angleMax         = 45
	PERIOD           = 20 * time.Millisecond
	BLINK_MAIN_COUNT = int(500 * time.Millisecond / PERIOD)
	BLINK_WARM_COUNT = int(200 * time.Millisecond / PERIOD)
	BLINK_PARA_COUNT = int(200 * time.Millisecond / PERIOD)
	TRACE_COUNT      = int(1000 * time.Millisecond / PERIOD)
)

var (
	d *display.Display
	t *trainer.Trainer
	o *orientation.Orientation
)

func init() {

	// Orientation
	o = orientation.New()
	o.Configure(PERIOD)

	// Bluetooth (FrSKY's PARA trainer protocol)
	t = trainer.New()
	t.Configure()
	go t.Run(PERIOD)

	// Display
	d = display.New()
	d.Configure(t.Address)
	go d.Run(5 * PERIOD)

}

func main() {

	// warm up
	for i := 0; i < 10; i++ {
		o.Update(false)
		time.Sleep(PERIOD)
	}
	// record initial orientation
	o.Center()
	// stabilize gyroscope
	// for !o.Stable() {
	// 	o.Update(false)
	// 	d.Paired = t.Paired
	// 	state(iter)
	// 	time.Sleep(PERIOD)
	// 	iter++
	// }

	// main loop
	iter := 0
	for {

		// update orientation
		if !o.Stable() {
			o.Update(false)
		} else {
			o.Update(true)
			for i, v := range o.Angles() {
				a := angleToChannel(v, 45)
				t.Channels[i] = a
				d.Channels[i] = a
			}
		}
		d.Paired = t.Paired

		// blink and trace
		state(iter)
		trace(iter)

		// wait
		time.Sleep(PERIOD)
		iter++

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
