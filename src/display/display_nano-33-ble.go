//go:build nano_33_ble
// +build nano_33_ble

package display

import (
	"image/color"

	"tinygo.org/x/tinydraw"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/proggy"
)

func (d *Display) showAddress() {
	tinyfont.WriteLineRotated(&d.device, &proggy.TinySZ8pt7b, 120, 0, d.Address, WHITE, tinyfont.ROTATION_180)
}

func (d *Display) showVersion(color color.RGBA) {
	tinyfont.WriteLineRotated(&d.device, &proggy.TinySZ8pt7b, 80, 18, d.Version, color, tinyfont.ROTATION_180)
}

func (d *Display) showValue(idx int) {
	x := 128 - int16(64+(int16(d.Channels[idx])-1500)/10)
	y := 28 - int16(idx*5)
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
		return
	}
	tinyfont.WriteLineRotated(&d.device, &proggy.TinySZ8pt7b, 120, 0, "  :  :  :  :  :  ", d.blinkColor, tinyfont.ROTATION_180)
	if d.blinkColor == WHITE && !d.Paired {
		d.blinkColor = BLACK
	} else {
		d.blinkColor = WHITE
	}
	d.blinkCount = 5
}
