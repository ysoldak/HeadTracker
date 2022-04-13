//go:build xiao_ble
// +build xiao_ble

package main

import (
	"time"

	"tinygo.org/x/tinydraw"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/proggy"
)

func (d *Display) Show() {
	tinyfont.WriteLineRotated(&d.device, &proggy.TinySZ8pt7b, 8, 28, d.addr, WHITE, tinyfont.NO_ROTATION)
	c := WHITE
	count := 5
	for {
		for i := 0; i < 3; i++ {
			d.ShowValue(i)
		}
		if count == 0 {
			tinyfont.WriteLineRotated(&d.device, &proggy.TinySZ8pt7b, 8, 28, "  :  :  :  :  :  ", c, tinyfont.NO_ROTATION)
			if c == WHITE && !paired {
				c = BLACK
			} else {
				c = WHITE
			}
			count = 5
		}
		d.device.Display()
		count--
		time.Sleep(100 * time.Millisecond)
	}
}

func (d *Display) ShowValue(idx int) {
	value := d.values[idx]
	x := int16(64 + (int16(value)-1500)/10)

	y := int16(idx * 5)

	tinydraw.FilledRectangle(&d.device, 13, y, 115, 3, BLACK)
	if x < 64 {
		tinydraw.FilledRectangle(&d.device, x, y, 64-x, 3, WHITE)
	} else {
		tinydraw.FilledRectangle(&d.device, 64, y, x-64, 3, WHITE)
	}
}
