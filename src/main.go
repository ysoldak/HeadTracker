package main

import (
	"machine"
	"time"

	"github.com/tracktum/go-ahrs"
)

const (
	angleMax = 45
	PERIOD   = 50 * time.Millisecond
)

func main() {
	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	led.Low()

	debugPin.Configure(machine.PinConfig{Mode: machine.PinOutput})

	// IMU
	imu := NewIMU()
	imu.Configure()

	// Orientation
	fusion := ahrs.NewMadgwick(0.05, float64(time.Second/PERIOD))

	// Bluetooth
	paraSetup()

	// Main loop
	initial := [3]float32{0, 0, 0}
	current := [3]float32{0, 0, 0}
	warmup := time.Now().Add(5 * time.Second)

	for {

		now := time.Now()

		gx, gy, gz, ax, ay, az, mx, my, mz := imu.Read()
		debugToggle()

		var q [4]float64
		if now.Before(warmup) {
			q = orientationToQuaternion(ax, ay, az, mx, my, mz)
			initial[0], initial[1], initial[2] = quaternionToAngles(q)
			fusion.Quaternions = q
		} else {
			q = fusion.Update9D(
				gx*degToRad, gy*degToRad, gz*degToRad,
				ax, ay, az,
				mx, my, mz,
			)
		}

		current[0], current[1], current[2] = quaternionToAngles(q)
		for i := byte(0); i < 3; i++ {
			angle := angleMinusAngle(current[i], initial[i])
			value := angleToChannel(angle, 45)
			paraSet(i, value)
		}
		paraSend()

		sleep := PERIOD - time.Since(now)
		if sleep > 0 {
			time.Sleep(sleep)
		}

	}

}
