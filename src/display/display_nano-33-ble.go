//go:build nano_33_ble

package display

import (
	"tinygo.org/x/tinydraw"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/proggy"
)

func (d *Display) showText() {
	for i := range d.Text {
		if d.Text[i] == "" {
			continue
		}
		tinyfont.WriteLineRotated(&d.device, &proggy.TinySZ8pt7b, 114, 18-int16(i)*18, d.Text[i], WHITE, tinyfont.ROTATION_180)
	}
}

func (d *Display) showValue(idx int) {
	if d.Text[0] != "" {
		return
	}
	x := (int16(d.Channels[idx]) - 1500) / 10
	y := 28 - int16(idx*5)
	tinydraw.FilledRectangle(&d.device, 13, y, 115, 3, BLACK)
	if !d.Stable {
		if x < 0 {
			x = -x
		}
		tinydraw.FilledRectangle(&d.device, 64-x-1, y, x*2+2, 3, WHITE)
		return
	}
	if x < 0 {
		tinydraw.FilledRectangle(&d.device, 64+x, y, -x, 3, WHITE)
	} else {
		tinydraw.FilledRectangle(&d.device, 64, y, x, 3, WHITE)
	}
}

func (d *Display) showPaired() {
	if d.blinkCount > 0 {
		d.blinkCount--
		return
	}
	tinyfont.WriteLineRotated(&d.device, &proggy.TinySZ8pt7b, 114, 0, "  :  :  :  :  :  ", d.blinkColor, tinyfont.ROTATION_180)
	if d.blinkColor == WHITE && !d.Paired {
		d.blinkColor = BLACK
	} else {
		d.blinkColor = WHITE
	}
	d.blinkCount = 5
}
