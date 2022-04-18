//go:build xiao_ble
// +build xiao_ble

package display

import (
	"tinygo.org/x/tinydraw"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/proggy"
)

func (d *Display) showAddress() {
	tinyfont.WriteLineRotated(&d.device, &proggy.TinySZ8pt7b, 8, 28, d.Address, WHITE, tinyfont.NO_ROTATION)
}

func (d *Display) showValue(idx int) {
	x := int16(64 + (int16(d.Channels[idx])-1500)/10)
	y := int16(idx * 5)
	tinydraw.FilledRectangle(&d.device, 13, y, 115, 3, BLACK)
	if x < 64 {
		tinydraw.FilledRectangle(&d.device, x, y, 64-x, 3, WHITE)
	} else {
		tinydraw.FilledRectangle(&d.device, 64, y, x-64, 3, WHITE)
	}
}

func (d *Display) showPaired() {
	if d.blinkCount > 0 {
		d.blinkCount--
	}
	tinyfont.WriteLineRotated(&d.device, &proggy.TinySZ8pt7b, 8, 28, "  :  :  :  :  :  ", d.blinkColor, tinyfont.NO_ROTATION)
	if d.blinkColor == WHITE && !d.Paired {
		d.blinkColor = BLACK
	} else {
		d.blinkColor = WHITE
	}
	d.blinkCount = 5
}
