package main

import (
	"machine"
	"time"

	"github.com/tracktum/go-ahrs"
)

const (
	angleMax    = 45
	PERIOD      = 20 * time.Millisecond
	BLINK_COUNT = int(500 * time.Millisecond / PERIOD)
	TRACE_COUNT = int(1000 * time.Millisecond / PERIOD)
	PARA_COUNT  = int(200 * time.Millisecond / PERIOD)
)

const (
	led  = machine.LED
	ledR = machine.LED_RED
	ledB = machine.LED_BLUE
)

func main() {

	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	ledR.Configure(machine.PinConfig{Mode: machine.PinOutput})
	ledB.Configure(machine.PinConfig{Mode: machine.PinOutput})

	led.Low()   // off
	ledR.High() // off
	ledB.High() // off

	//debugPin.Configure(machine.PinConfig{Mode: machine.PinOutput})

	// IMU
	imu := NewIMU()
	imu.Configure()

	ledR.Low() // on, imu initialised, warmup starts

	// Orientation
	fusion := ahrs.NewMadgwick(0.01, float64(time.Second/PERIOD))

	// Bluetooth
	paraSetup()

	// Display
	displaySetup()

	// Main loop
	initial := [3]float32{0, 0, 0}
	current := [3]float32{0, 0, 0}
	warmup := time.Now().Add(60 * time.Second)
	warmed := false

	counter := 0
	for {

		if !warmed {
			warmed = time.Now().After(warmup)
		} else {
			ledR.High() // off, warmup finished
		}

		// read sensors
		gx, gy, gz, ax, ay, az, err := imu.Read(!warmed)

		// record initial orientation
		if counter == 0 {
			q := orientationToQuaternion(ax, ay, az, 1, 0, 0) // assume N since we don't have mag
			initial[0], initial[1], initial[2] = quaternionToAngles(q)
			fusion.Quaternions = q
		}

		// main logic
		if warmed && err == nil {
			q := fusion.Update6D(
				gx*degToRad, gy*degToRad, gz*degToRad,
				ax, ay, az,
			)
			current[0], current[1], current[2] = quaternionToAngles(q)
			for i := byte(0); i < 3; i++ {
				angle := angleMinusAngle(current[i], initial[i])
				value := angleToChannel(angle, 45)
				paraSet(i, value)
				displaySet(i, value)
			}
		}

		// push data
		paraSend()

		// indicate status
		if counter%BLINK_COUNT == 0 { // indicate main loop running
			togglePin(led)
		}
		if counter%TRACE_COUNT == 0 { // print out state
			println(time.Now().Unix(), ": ", paraAddress, " [", channels[0], ",", channels[1], ",", channels[2], "] (", imu.gyrCal.offset[0], ",", imu.gyrCal.offset[1], ",", imu.gyrCal.offset[2], ")")
		}
		if counter%PARA_COUNT == 0 { // indicate para (bluetooth) state
			if paired {
				ledB.Low() // on, connected
			} else {
				togglePin(ledB) // blink, advertising
			}
		}

		counter++

		time.Sleep(PERIOD)

	}

}

func togglePin(pin machine.Pin) {
	if pin.Get() {
		pin.Low()
	} else {
		pin.High()
	}
}
