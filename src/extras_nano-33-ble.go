//go:build nano_33_ble

package main

import "errors"

var errNoBattery = errors.New("no battery")

func initExtras() {
	// nop
}

func batteryVoltage() (float64, error) {
	return 0, errNoBattery
}
