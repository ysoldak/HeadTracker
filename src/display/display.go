package display

import (
	"image/color"
	"machine"

	"tinygo.org/x/drivers/ssd1306"
	"tinygo.org/x/tinydraw"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/proggy"
)

var BLACK = color.RGBA{0, 0, 0, 255}
var WHITE = color.RGBA{255, 255, 255, 255}

type Text struct {
	row     byte
	text    string
	altText string
	blink   func() bool
}

type Bar struct {
	value int16
	bidir bool
}

type Display struct {
	device ssd1306.Device

	blinkCount int
	blinkColor color.RGBA

	showBars bool

	Texts []*Text
	Bars  [3]Bar
}

func New() *Display {
	return &Display{
		Texts: []*Text{},
		Bars:  [3]Bar{},
	}
}

func (d *Display) Configure() {
	machine.I2C0.Configure(machine.I2CConfig{
		Frequency: 400 * machine.KHz,
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

func (d *Display) RemoveTextRow(row byte) { // TODO use reslice for efficiency
	result := []*Text{}
	for _, t := range d.Texts {
		if t.row == row {
			d.print(t.row, t.text, BLACK)
			d.print(t.row, t.altText, BLACK)
			continue
		}
		result = append(result, t)
	}
	d.Texts = result
}

func (d *Display) RemoveText(text *Text) { // TODO use reslice for efficiency
	result := []*Text{}
	for _, t := range d.Texts {
		if text == nil || t == text {
			d.print(t.row, t.text, BLACK)
			d.print(t.row, t.altText, BLACK)
			continue
		}
		result = append(result, t)
	}
	d.Texts = result
}

func (d *Display) AddText(row byte, text string) *Text {
	t := &Text{row, text, text, nil}
	d.Texts = append(d.Texts, t)
	d.print(t.row, t.text, WHITE)
	return t
}

func (d *Display) SetTextBlink(text *Text, other string, blink bool) *Text {
	for _, t := range d.Texts {
		if t == text {
			t.altText = other
			t.blink = func() bool { return blink }
		}
	}
	return text
}

func (d *Display) SetTextBlinkFunc(text *Text, other string, blink func() bool) *Text {
	for _, t := range d.Texts {
		if t == text {
			t.altText = other
			t.blink = blink
		}
	}
	return text
}

func (d *Display) SetBar(row byte, value int16, bidir bool) {
	d.showBars = true
	d.Bars[row].value = value
	d.Bars[row].bidir = bidir
}

func (d *Display) Update() {
	d.bars()
	d.blink()
	d.device.Display()
}

func (d *Display) bars() {
	if !d.showBars {
		return
	}
	for row, b := range d.Bars {
		y := int16(row * 5)
		tinydraw.FilledRectangle(&d.device, 13, y, 115, 3, BLACK)
		length := b.value
		if length < 0 {
			tinydraw.FilledRectangle(&d.device, 64+length, y, -length, 3, WHITE)
			if b.bidir {
				tinydraw.FilledRectangle(&d.device, 64, y, -length, 3, WHITE)
			}
		} else {
			tinydraw.FilledRectangle(&d.device, 64, y, length, 3, WHITE)
			if b.bidir {
				tinydraw.FilledRectangle(&d.device, 64-length, y, length, 3, WHITE)
			}
		}
	}
}

func (d *Display) blink() {
	if d.blinkCount > 0 {
		d.blinkCount--
		return
	}
	for _, t := range d.Texts {
		if t.blink != nil && t.blink() {
			if d.blinkColor == WHITE {
				d.print(t.row, t.altText, BLACK) // hide alt text
				d.print(t.row, t.text, WHITE)    // show main text
			} else {
				d.print(t.row, t.text, BLACK)    // hide main text
				d.print(t.row, t.altText, WHITE) // show alt text
			}
		}
	}
	d.blinkColor = toggleColor(d.blinkColor)
	d.blinkCount = 5 // blink every 5th iteration (~500mS)
}

func toggleColor(c color.RGBA) color.RGBA {
	if c == WHITE {
		return BLACK
	}
	return WHITE
}

func (d *Display) print(row byte, text string, c color.RGBA) {
	tinyfont.WriteLineRotated(&d.device, &proggy.TinySZ8pt7b, 14, 12+int16(row)*16, text, c, tinyfont.NO_ROTATION)
}
