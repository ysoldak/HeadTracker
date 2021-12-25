package main

import (
	"machine"
	"time"
)

const (
	angleMax = 45
	PERIOD   = 50 * time.Millisecond
)

var imuChan = make(chan [3]float32, 1)
var paraChan = make(chan [3]uint16, 1)

func main() {
	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	led.Low()

	println("starting")

	imuSetup()
	go imuWork(imuChan)

	paraSetup()
	go paraWork(paraChan)

	// now := time.Now().UnixMilli()
	for {
		angles := <-imuChan
		values := [3]uint16{
			toChannel(angles[0]),
			toChannel(angles[1]),
			toChannel(angles[2]),
		}
		paraChan <- values
		// fmt.Printf("%10d | %8.3f %8.3f %8.3f | %4d %4d %4d\r\n", time.Now().UnixMilli(), angles[0], angles[1], angles[2], values[0], values[1], values[2])
	}

}

// func updateFake() {
// 	newValue := ((channels[0]-1000)+1)%1000 + 1000
// 	for i := 0; i < 3; i++ {
// 		paraSet(byte(i), newValue)
// 	}
// 	paraSend()
// }

// func update() {
// 	var pitch, roll, yaw float64

// 	// if time.Since(conTime) < 5*time.Second { // give it 5 sec to settle, record start angle
// 	// 	imuWarmup()
// 	// 	pitch = 0
// 	// 	roll = 0
// 	// 	yaw = 0
// 	// } else {
// 	// 	pitch, roll, yaw = imuAngles()
// 	// }

// 	paraSet(0, toChannel(pitch))
// 	paraSet(1, toChannel(roll))
// 	paraSet(2, toChannel(yaw))
// 	paraSend()

// }

func toChannel(angle float32) uint16 {
	result := uint16(1500 + 500/angleMax*angle)
	if result < 988 {
		return 988
	}
	if result > 2012 {
		return 2012
	}
	return result
}

func must(action string, err error) {
	if err != nil {
		for {
			println("failed to " + action + ": " + err.Error())
			time.Sleep(time.Second)
		}
	}
}
