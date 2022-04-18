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

	// warmup
	for i := 0; i < 10; i++ {
		o.Update(false)
		time.Sleep(PERIOD)
	}

	// main loop
	iter := 0
	for {

		// update orientation
		o.Update(true)
		r := angleToChannel(o.Roll, 45)
		p := angleToChannel(o.Pitch, 45)
		y := angleToChannel(o.Yaw, 45)

		// update trainer
		t.Channels[0], t.Channels[1], t.Channels[2] = r, p, y

		// update display
		d.Channels[0], d.Channels[1], d.Channels[2] = r, p, y
		d.Paired = t.Paired

		// blink and trace
		state(iter)
		trace(r, p, y, iter)

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
	if iter%BLINK_PARA_COUNT == 0 { // indicate para (bluetooth) state
		if t.Paired {
			on(ledB) // on, connected
		} else {
			toggle(ledB) // blink, advertising
		}
	}
}

func trace(r, p, y uint16, iter int) {
	if iter%TRACE_COUNT == 0 { // print out state
		rc, pc, yc := o.Calibration()
		println(time.Now().Unix(), ": ", t.Address, " [", r, ",", p, ",", y, "] (", rc, ",", pc, ",", yc, ")")
	}
}
