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
	PERIOD           = 10
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
	i *orientation.IMU
	o *orientation.Orientation
	f *Flash
)

var (
	tickPeriod *time.Ticker
)

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

	// Trainer (Bluetooth or PPM)
	if !pinSelectPPM.Get() { // Low means connected to GND => PPM output requested
		t = trainer.NewPPM(pinOutputPPM) // PPM wire
	} else {
		t = trainer.NewPara()
	}

	// Display
	d = display.New()
	d.Configure()

	f = &Flash{}

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

	flashLoad()

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

		stateMain(iter)
		stateCalibration(iter)
		trace(iter)

		time.Sleep(time.Millisecond)
		iter++
		iter %= 10_000

		if iter%10 == 0 {
			d.Update()
		}

		if time.Now().After(stopTime) && !f.IsEmpty() { // when had some calibration already, force it if was not able to find better quickly
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
	t.Configure("HT " + Version)
	t.Start()

	// switch display to normal mode
	d.RemoveText(nil)
	d.AddText(1, t.Address())
	if pinSelectPPM.Get() { // high means Bluetooth
		d.SetTextBlinkFunc(d.AddText(1, "  :  :  :  :  :  "), "", func() bool { return !t.Paired() })
	}
	d.Update()

	// main loop
	iter = 0
	offLedRIter := -1
	for range tickPeriod.C {

		pinDebugMain.Set(!pinDebugMain.Get())

		if !pinResetCenter.Get() || (iter%400 == 0 && i.ReadTap() || t.ResetRequested()) { // Button pressed OR [double] tap registered (shall not read register more frequently than double tap duration)
			o.Reset()
			on(ledR)
			println("HT", Version, "|", t.Address(), "| [ Orientation reset  ]")
			offLedRIter = (int(iter) + 500) % 10_000 // keep LED on for 500 ms
		}
		if offLedRIter >= 0 && int(iter) == offLedRIter {
			off(ledR)
			offLedRIter = -1
		}

		o.Update()

		pinDebugData.High()
		if iter%20 == 0 {
			for i, a := range o.Angles() {
				c := angleToChannel(a)
				t.SetChannel(i, c)
				d.SetBar(byte(i), int16(1500-c)/10, false)
			}
			t.Update()
			d.Update()
		}
		pinDebugData.Low()

		// blink and trace
		stateMain(iter)
		statePara(iter)
		trace(iter)

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

var ms = runtime.MemStats{}

func trace(iter uint16) {
	if iter%TRACE_COUNT == 0 { // print out state
		channels := t.Channels()
		r, p, y := channels[0], channels[1], channels[2]
		rc, pc, yc := o.Offsets()
		runtime.ReadMemStats(&ms)
		println("HT", Version, "|", t.Address(), "| [", r, ",", p, ",", y, "] (", rc, ",", pc, ",", yc, ")", ms.HeapInuse)
	}
}
