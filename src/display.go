package main

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers/ssd1306"
	"tinygo.org/x/tinydraw"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/proggy"
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

func (d *Display) Show() {
	tinyfont.WriteLineRotated(&d.device, &proggy.TinySZ8pt7b, 120, 0, d.addr, WHITE, tinyfont.ROTATION_180)
	c := WHITE
	count := 5
	for {
		for i := 0; i < 3; i++ {
			d.ShowValue(i)
		}
		if count == 0 {
			tinyfont.WriteLineRotated(&d.device, &proggy.TinySZ8pt7b, 120, 0, "  :  :  :  :  :  ", c, tinyfont.ROTATION_180)
			if c == WHITE && !paired {
				c = BLACK
			} else {
				c = WHITE
			}
			count = 5
		}
		count--
		time.Sleep(100 * time.Millisecond)
	}
}

func (d *Display) ShowValue(idx int) {
	value := d.values[idx]
	x := 128 - int16(64+(int16(value)-1500)/10)

	y := 28 - int16(idx*5)

	tinydraw.FilledRectangle(&d.device, 13, y, 115, 3, BLACK)
	if x < 64 {
		tinydraw.FilledRectangle(&d.device, x, y, 64-x, 3, WHITE)
	} else {
		tinydraw.FilledRectangle(&d.device, 64, y, x-64, 3, WHITE)
	}

	d.device.Display()
}

func (d *Display) Set(idx byte, value uint16) {
	d.values[idx] = value
}

// --------------------------------------

var display *Display

func displaySetup() {
	display = &Display{}
	display.Configure(paraAddress)
	go display.Show()
}

func displaySet(idx byte, value uint16) {
	display.Set(idx, value)
}
