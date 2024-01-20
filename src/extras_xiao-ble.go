//go:build xiao_ble

// The battery charging current is selectable as 50mA or 100mA,
// where you can set Pin13 as high or low to change it to 50mA or 100mA.
//
// The low current charging current is at the input model set up as HIGH LEVEL
// and the high current charging current is at the output model set up as LOW LEVEL.
//
// In this project, we ensure the high current charging is enabled by default.
//
// See https://wiki.seeedstudio.com/XIAO_BLE/#battery-charging-current
package main

import "machine"

var (
	pinChargeCurrent = machine.P0_13
	pinRead          = machine.P0_14
	pinVoltage       = machine.P0_31

	adc machine.ADC
)

func initExtras() {
	pinChargeCurrent.Configure(machine.PinConfig{Mode: machine.PinOutput})
	pinChargeCurrent.Low() // enable charging at high current, 100mA

	// Shall keep this low while reading voltage
	pinRead.Configure(machine.PinConfig{Mode: machine.PinOutput})

	// Battery sensor pin
	adc = machine.ADC{Pin: pinVoltage}
	adc.Configure(machine.ADCConfig{})
}

func batteryVoltage() (float64, error) {
	pinRead.Low()
	return float64(adc.Get()) / 7050, nil
}
