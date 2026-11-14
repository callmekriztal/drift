package canvas

import (
	"bytes"
	"strconv"
)

// ColorMode selects the output color encoding.
type ColorMode int

const (
	// ColorTrue is 24-bit truecolor (SGR 38;2 / 48;2).
	ColorTrue ColorMode = iota
	// Color256 is xterm-256 nearest-cube mapping.
	Color256
)

var (
	halfBlock = []byte("▀")
	escFG     = []byte("\x1b[38;")
	escBG     = []byte("\x1b[48;")
	escReset  = []byte("\x1b[0m")
)

// Render composes the buffer into ANSI cell rows and appends them to buf.
// Colors are gamma-encoded then quantized with *static, position-seeded
// dither*, so unchanged pixels keep byte-identical encodings across frames and
// the terminal (or Bubble Tea's line differ) can skip them.
func (c *Canvas) Render(mode ColorMode, buf *bytes.Buffer) {
	buf.Reset()

	// Index overlays by cell for O(1) lookup while emitting.
	ovIdx := make(map[int]Overlay, 128)
	for _, o := range c.overlays {
		x := o.Col
		for _, r := range o.Text {
			if x < c.CW {
				ovIdx[o.Row*c.CW+x] = Overlay{Row: o.Row, Col: x, Text: string(r), Color: o.Color, Alpha: o.Alpha}
			}
			x++
		}
	}

	var lastFg, lastBg rgb8
	lastFg.invalid, lastBg.invalid = true, true

	// Overlay rows get a per-row average backdrop: rock-stable quantized
	// colors, so the whole text row emits one escape run and diffs clean.
	avgRow, avgVal := -1, RGB{}
	rowAvg := func(row int) RGB {
		if row == avgRow {
			return avgVal
		}
		var sum RGB
		for x := 0; x < c.W; x++ {
			sum.R += c.Pix[row*c.W+x].R
			sum.G += c.Pix[row*c.W+x].G
			sum.B += c.Pix[row*c.W+x].B
		}
		avgVal = RGB{sum.R / float32(c.W), sum.G / float32(c.W), sum.B / float32(c.W)}
		avgRow = row
		return avgVal
	}

	for cy := 0; cy < c.CH; cy++ {
		lastFg.invalid, lastBg.invalid = true, true
		for cx := 0; cx < c.CW; cx++ {
			ti := (cy*2)*c.W + cx
			bi := (cy*2+1)*c.W + cx
			top, bot := c.Pix[ti], c.Pix[bi]

			ch := halfBlock
			fg, bg := top, bot

			if br, bcol, ok := c.Br.cell(cx, cy); ok {
				ch = []byte(string(br))
				fg = bcol
				bg = Mix(top, bot, 0.5) // braille keeps the scene color as background
			}

			if o, ok := ovIdx[cy*c.CW+cx]; ok {
				ch = []byte(o.Text)
				// A quiet per-row backdrop: the text sits on the row's
				// average tone, so glyphs stay crisp and the whole row
				// emits one stable escape run.
				base := rowAvg(cy * 2)
				fg = Mix(base, o.Color, o.Alpha)
				bg = rowAvg(cy*2 + 1).Scale(0.45)
			}

			fg8 := quantizeD(fg, cx, cy*2, true)
			bg8 := quantizeD(bg, cx, cy*2+1, true)
			if _, isOv := ovIdx[cy*c.CW+cx]; isOv {
				// Overlay rows are deliberately flat: skip the positional
				// dither so each text row emits one stable escape run.
				fg8 = quantizeD(fg, cx, cy*2, false)
				bg8 = quantizeD(bg, cx, cy*2+1, false)
			}

			if fg8 != lastFg {
				writeColor(buf, escFG, fg8, mode)
				lastFg = fg8
			}
			if bg8 != lastBg {
				writeColor(buf, escBG, bg8, mode)
				lastBg = bg8
			}
			buf.Write(ch)
		}
		buf.Write(escReset)
		if cy < c.CH-1 {
			buf.WriteByte('\n')
		}
	}
}

type rgb8 struct {
	R, G, B uint8
	invalid bool
}

// quantize gamma-encodes and rounds a linear color to 8-bit, optionally with
// static position-seeded dither to kill gradient banding.
func quantizeD(col RGB, x, y int, dither bool) rgb8 {
	var d float32
	if dither {
		d = hashUnit(x, y)
	}
	r8 := to8(GammaEncode(col.R), d)
	g8 := to8(GammaEncode(col.G), d)
	b8 := to8(GammaEncode(col.B), d)
	return rgb8{r8, g8, b8, false}
}

func to8(v, dither float32) uint8 {
	v = v + dither*(1.0/255.0)
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	return uint8(v*255.0 + 0.5)
}

// writeColor emits an SGR color sequence in the requested mode.
func writeColor(buf *bytes.Buffer, prefix []byte, c rgb8, mode ColorMode) {
	buf.Write(prefix)
	switch mode {
	case ColorTrue:
		buf.WriteString("2;")
		buf.WriteString(strconv.Itoa(int(c.R)))
		buf.WriteByte(';')
		buf.WriteString(strconv.Itoa(int(c.G)))
		buf.WriteByte(';')
		buf.WriteString(strconv.Itoa(int(c.B)))
		buf.WriteByte('m')
	default:
		buf.WriteString("5;")
		buf.WriteString(strconv.Itoa(cube256(c)))
		buf.WriteByte('m')
	}
}

// cube256 maps an 8-bit RGB triple to the nearest xterm-256 color cube index.
func cube256(c rgb8) int {
	r := (int(c.R)*5 + 127) / 255
	g := (int(c.G)*5 + 127) / 255
	b := (int(c.B)*5 + 127) / 255
	return 16 + 36*r + 6*g + b
}
