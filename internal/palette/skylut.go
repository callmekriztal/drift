package palette

import "math"

// SkyLUT precomputes the sky gradient over the full day at fixed hour steps
// and vertical stops. Per-frame sky queries become two table lookups and a
// lerp — no Oklab math per pixel.
const lutSteps = 96 // 15-minute resolution

// SkyLUT: Data[step*Rows + row] holds linear RGB.
type SkyLUT struct {
	Rows int
	Data []canvas_RGB
}

// canvas_RGB aliases the canvas color type through an interface-free wrapper;
// declared here to avoid importing canvas into every consumer signature.
type canvas_RGB = struct{ R, G, B float32 }

// BuildSkyLUT samples the palette across 24h into `rows` vertical stops.
func BuildSkyLUT(p *Palette, rows int) *SkyLUT {
	lut := &SkyLUT{Rows: rows, Data: make([]canvas_RGB, lutSteps*rows)}
	for s := 0; s < lutSteps; s++ {
		hour := float64(s) / float64(lutSteps) * 24
		sm := p.Sample(hour)
		for r := 0; r < rows; r++ {
			t := float64(r) / float64(rows-1)
			// Non-linear vertical mapping: most action lives near the horizon.
			t = math.Pow(t, 1.45)
			c := MixLab(sm.Zenith, sm.Horizon, t).Linear()
			lut.Data[s*rows+r] = canvas_RGB(c)
		}
	}
	return lut
}

// At queries the sky color at a local hour and vertical position, where
// yNorm 0 is the zenith and 1 is the horizon.
func (l *SkyLUT) At(hour, yNorm float64) canvas_RGB {
	if yNorm < 0 {
		yNorm = 0
	}
	if yNorm > 1 {
		yNorm = 1
	}
	h := math.Mod(hour, 24)
	if h < 0 {
		h += 24
	}
	hx := h / 24 * lutSteps
	i0 := int(hx) % lutSteps
	i1 := (i0 + 1) % lutSteps
	fh := hx - float64(int(hx))

	ry := yNorm * float64(l.Rows-1)
	r0 := int(ry)
	r1 := r0 + 1
	if r1 > l.Rows-1 {
		r1 = l.Rows - 1
	}
	fr := ry - float64(r0)

	a := l.Data[i0*l.Rows+r0]
	b := l.Data[i1*l.Rows+r0]
	c := l.Data[i0*l.Rows+r1]
	d := l.Data[i1*l.Rows+r1]
	return canvas_RGB{
		R: float32(lerp(lerp(float64(a.R), float64(b.R), fh), lerp(float64(c.R), float64(d.R), fh), fr)),
		G: float32(lerp(lerp(float64(a.G), float64(b.G), fh), lerp(float64(c.G), float64(d.G), fh), fr)),
		B: float32(lerp(lerp(float64(a.B), float64(b.B), fh), lerp(float64(c.B), float64(d.B), fh), fr)),
	}
}

func lerp(a, b, t float64) float64 { return a + (b-a)*t }
