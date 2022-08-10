package display

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers/ssd1306"
)

var BLACK = color.RGBA{0, 0, 0, 255}
var WHITE = color.RGBA{255, 255, 255, 255}

type Display struct {
	device ssd1306.Device

	Bluetooth  bool
	blinkCount int
	blinkColor color.RGBA

	Paired   bool
	Stable   bool
	Address  string
	Version  string
	Channels [3]uint16
}

func New() *Display {
	return &Display{
		Paired:   false,
		Address:  "B1:6B:00:B5:BA:BE",
		Channels: [3]uint16{1500, 1500, 1500},
	}
}

func (d *Display) Configure() {
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

func (d *Display) Run() {
	d.showAddress()
	clearVersion := true
	for {
		if d.Stable {
			if clearVersion {
				d.showVersion(BLACK)
				clearVersion = false
			}
			for i := 0; i < 3; i++ {
				d.showValue(i)
			}
		} else {
			d.showVersion(WHITE)
		}
		if d.Bluetooth {
			d.showPaired()
		}
		d.device.Display()
		time.Sleep(100 * time.Millisecond)
	}
}
