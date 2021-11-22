package main

import (
	"machine"
	"time"

	"tinygo.org/x/bluetooth"
)

var conStatus = make(chan bool, 1)

func main() {
	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	led.Low()

	blue := machine.LED_BLUE
	blue.Configure(machine.PinConfig{Mode: machine.PinOutput})
	blue.High()

	println("starting")

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
				paraBoot()
				println("Connected")
			} else {
				blue.High()
				sendAfter = time.Time{}
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
			newValue := ((channels[0]-1000)+1)%1000 + 1000
			for i := 0; i < 3; i++ {
				paraSet(byte(i), newValue)
			}
			paraSend()
			time.Sleep(40 * time.Millisecond)
		}
	}

}

func must(action string, err error) {
	if err != nil {
		for {
			println("failed to " + action + ": " + err.Error())
			time.Sleep(time.Second)
		}
	}
}
