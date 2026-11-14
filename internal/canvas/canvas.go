package canvas

// Canvas holds a W×H pixel buffer in linear light. Pixels map 1:1 to terminal
// columns and 2:1 to terminal rows (half-block ▀ rendering gives 2× vertical
// resolution: one cell shows the pixel above it as foreground and below as
// background).
type Canvas struct {
	W, H int // pixels: W == cells wide, H == 2× cells tall
	CW   int // cells wide
	CH   int // cells tall

	Pix []RGB

	Br Braille

	overlays []Overlay
}

// New builds a canvas for a wCells×hCells terminal area.
func New(wCells, hCells int) *Canvas {
	c := &Canvas{}
	c.Resize(wCells, hCells)
	return c
}

// Resize rebuilds the buffer for a new cell size, reusing backing arrays.
func (c *Canvas) Resize(wCells, hCells int) {
	if wCells < 1 {
		wCells = 1
	}
	if hCells < 1 {
		hCells = 1
	}
	c.CW, c.CH = wCells, hCells
	c.W, c.H = wCells, hCells*2
	n := c.W * c.H
	if cap(c.Pix) >= n {
		c.Pix = c.Pix[:n]
	} else {
		c.Pix = make([]RGB, n)
	}
	c.Br.Resize(c.W, c.H)
	c.overlays = c.overlays[:0]
}

func (c *Canvas) Clear() {
	for i := range c.Pix {
		c.Pix[i] = RGB{}
	}
	c.Br.Clear()
}

// At returns the pixel at integer coords (bounds-clamped).
func (c *Canvas) At(x, y int) RGB {
	x = clampi(x, 0, c.W-1)
	y = clampi(y, 0, c.H-1)
	return c.Pix[y*c.W+x]
}

func (c *Canvas) Set(x, y int, col RGB) {
	if x < 0 || y < 0 || x >= c.W || y >= c.H {
		return
	}
	c.Pix[y*c.W+x] = col
}

// Blend does source-over compositing: dst = dst*(1-a) + col*a.
func (c *Canvas) Blend(x, y int, col RGB, a float32) {
	if a <= 0 || x < 0 || y < 0 || x >= c.W || y >= c.H {
		return
	}
	if a > 1 {
		a = 1
	}
	i := y*c.W + x
	p := &c.Pix[i]
	p.R += (col.R - p.R) * a
	p.G += (col.G - p.G) * a
	p.B += (col.B - p.B) * a
}

// Add adds light: dst += col*a (clamped later at quantization).
func (c *Canvas) Add(x, y int, col RGB, a float32) {
	if a <= 0 || x < 0 || y < 0 || x >= c.W || y >= c.H {
		return
	}
	if a > 1 {
		a = 1
	}
	i := y*c.W + x
	p := &c.Pix[i]
	p.R += col.R * a
	p.G += col.G * a
	p.B += col.B * a
}

func (c *Canvas) Fill(col RGB) {
	for i := range c.Pix {
		c.Pix[i] = col
	}
}

// ScaleAll multiplies every pixel (used for the launch fade-in).
func (c *Canvas) ScaleAll(f float32) {
	for i := range c.Pix {
		c.Pix[i].R *= f
		c.Pix[i].G *= f
		c.Pix[i].B *= f
	}
}

// Glow splats a soft additive radial falloff around a point.
func (c *Canvas) Glow(cx, cy, radius, alpha float64, col RGB) {
	if alpha <= 0 || radius <= 0 {
		return
	}
	x0 := int(cx - radius)
	x1 := int(cx+radius) + 1
	y0 := int(cy - radius)
	y1 := int(cy+radius) + 1
	for y := y0; y < y1; y++ {
		if y < 0 || y >= c.H {
			continue
		}
		dy := float64(y) - cy
		for x := x0; x < x1; x++ {
			if x < 0 || x >= c.W {
				continue
			}
			dx := float64(x) - cx
			d2 := (dx*dx + dy*dy) / (radius * radius)
			if d2 >= 1 {
				continue
			}
			fall := 1 - d2
			fall *= fall // quartic: tight core, soft skirt
			a := float32(alpha * fall)
			c.Add(x, y, col, a)
		}
	}
}

// Overlay is a text line painted over the final cell grid. Color is in linear
// light; Alpha is text opacity over whatever the cell already shows.
type Overlay struct {
	Row, Col int
	Text     string
	Color    RGB
	Alpha    float32
}

func (c *Canvas) Overlay(o Overlay) { c.overlays = append(c.overlays, o) }

// ClearOverlays drops all text overlays (call once per frame before re-adding).
func (c *Canvas) ClearOverlays() { c.overlays = c.overlays[:0] }

func clampi(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
