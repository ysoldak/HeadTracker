package main

import (
	"machine"
	"time"

	"tinygo.org/x/bluetooth"
)

const (
	angleMax = 45
	PERIOD   = 50 * time.Millisecond
)

var conStatus = make(chan bool, 1)

var conTime time.Time

func main() {
	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	led.Low()

	blue := machine.LED_BLUE
	blue.Configure(machine.PinConfig{Mode: machine.PinOutput})
	blue.High()

	println("starting")

	imuSetup()
	paraSetup()

	ble.SetConnectHandler(func(device bluetooth.Addresser, connected bool) {
		conStatus <- connected
	})

	c := false
	address := ""
	var sendAfter time.Time
	for {
		select {
		case c = <-conStatus:
			if c {
				blue.Low()
				sendAfter = time.Now().Add(1 * time.Second)
				conTime = time.Now()
				paraBoot()
				println("Connected")
			} else {
				blue.High()
				sendAfter = time.Time{}
				conTime = time.Time{}
				println("Disconnected")
			}
		default:
			if !c || sendAfter.IsZero() || !time.Now().After(sendAfter) {
				if len(address) == 0 {
					addr, _ := ble.Address()
					address = addr.MAC.String()
				}
				println("Advertising as Hello /", address)
				time.Sleep(1000 * time.Millisecond)
				continue
			}
			now := time.Now()
			update()
			sleep := PERIOD - time.Since(now)
			if sleep > 0 {
				time.Sleep(sleep)
			}
		}
	}

}

func updateFake() {
	newValue := ((channels[0]-1000)+1)%1000 + 1000
	for i := 0; i < 3; i++ {
		paraSet(byte(i), newValue)
	}
	paraSend()
}

func update() {
	var pitch, roll, yaw float64

	if time.Since(conTime) < 5*time.Second { // give it 5 sec to settle, record start angle
		imuWarmup()
		pitch = 0
		roll = 0
		yaw = 0
	} else {
		pitch, roll, yaw = imuAngles()
	}

	paraSet(0, toChannel(pitch))
	paraSet(1, toChannel(roll))
	paraSet(2, toChannel(yaw))
	paraSend()

}

func toChannel(angle float64) uint16 {
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
