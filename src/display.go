package main

import (
	"image/color"
	"machine"

	"tinygo.org/x/drivers/ssd1306"
)

var BLACK = color.RGBA{0, 0, 0, 255}
var WHITE = color.RGBA{255, 255, 255, 255}

type Display struct {
	addr   string
	values [3]uint16
	device ssd1306.Device
}

func (d *Display) Configure(addr string) {
	d.addr = addr

	machine.I2C0.Configure(machine.I2CConfig{
		Frequency: machine.TWI_FREQ_400KHZ,
		SDA:       machine.SDA0_PIN,
		SCL:       machine.SCL0_PIN,
	})
	d.device = ssd1306.NewI2C(machine.I2C0)
	d.device.Configure(ssd1306.Config{
		Address: ssd1306.Address_128_32,
		Width:   128,
		Height:  32,
	})
	d.device.ClearDisplay()
}

func (d *Display) Set(idx byte, value uint16) {
	d.values[idx] = value
}

// --------------------------------------

var display *Display

func displaySetup() {
	display = &Display{
		values: [3]uint16{1500, 1500, 1500},
	}
	display.Configure(paraAddress)
	go display.Show()
}

func displaySet(idx byte, value uint16) {
	display.Set(idx, value)
}
