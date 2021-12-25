package main

import "machine"

var debugPin = machine.D2
var debugValue = false

func debugToggle() {
	if debugValue {
		debugPin.High()
	} else {
		debugPin.Low()
	}
	debugValue = !debugValue
}
