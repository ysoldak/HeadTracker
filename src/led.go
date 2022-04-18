package main

import "machine"

const (
	led  = machine.LED
	ledR = machine.LED_RED
	ledG = machine.LED_GREEN
	ledB = machine.LED_BLUE
)

func init() {
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	ledR.Configure(machine.PinConfig{Mode: machine.PinOutput})
	ledG.Configure(machine.PinConfig{Mode: machine.PinOutput})
	ledB.Configure(machine.PinConfig{Mode: machine.PinOutput})

	led.Low()   // off
	ledR.High() // off
	ledG.High() // off
	ledB.High() // off
}

func toggle(pin machine.Pin) {
	if pin.Get() {
		pin.Low()
	} else {
		pin.High()
	}
}

func off(pin machine.Pin) {
	if pin == led {
		pin.Low()
	} else {
		pin.High()
	}
}

func on(pin machine.Pin) {
	if pin == led {
		pin.High()
	} else {
		pin.Low()
	}
}
