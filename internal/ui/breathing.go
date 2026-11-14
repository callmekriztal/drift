package ui

import (
	"drift/internal/canvas"
	"math"
	"unicode/utf8"

	"github.com/charmbracelet/harmonica"
)

// breathState implements the 4-7-8 breathing guide: a soft ring that expands
// for 4s (inhale), holds for 7s, and contracts for 8s (exhale), with the
// radius driven by a critically-damped spring tuned to each phase length.
type breathState struct {
	on    bool
	alpha float64
	phase int // 0 inhale, 1 hold, 2 exhale
	pt    float64
	pos   float64
	vel   float64
}

var breathDur = [3]float64{4, 7, 8}
var breathName = [3]string{"inhale", "hold", "exhale"}

func (b *breathState) toggle() {
	b.on = !b.on
	if b.on {
		b.phase = 0
		b.pt = 0
		b.vel = 0
		b.pos = 0 // spring starts collapsed and blooms in
	}
}

func (b *breathState) update(dt float64) {
	target := 0.0
	if b.on {
		target = 1.0
	}
	b.alpha += (target - b.alpha) * clamp01(dt/0.9)
	if !b.on {
		if b.alpha <= 0.01 {
			b.alpha = 0
		}
		return
	}
	b.pt += dt
	if b.pt >= breathDur[b.phase] {
		b.phase = (b.phase + 1) % 3
		b.pt = 0
	}
	// Angular frequency from the settling time of this phase.
	omega := 4.0 / (0.95 * breathDur[b.phase])
	spring := harmonica.NewSpring(dt, omega, 0.95)
	b.pos, b.vel = spring.Update(b.pos, b.vel, b.target())
}

func (b *breathState) target() float64 {
	if b.phase == 2 {
		return 0 // exhale collapses the ring
	}
	return 1 // normalized radius; render scales it
}

func (m *Model) breathOverlay(c *canvas.Canvas) {
	if m.breath.alpha <= 0.01 {
		return
	}
	cx := float64(c.W) / 2
	cy := float64(c.H) / 2
	minDim := math.Min(float64(c.W), float64(c.H))
	maxR := minDim * 0.24
	r := 0.10*minDim + m.breath.pos*(maxR-0.10*minDim)
	rx := r
	ry := r * 2 // terminal cells are ~1:2, so pixels stretch vertically

	fade := m.breath.alpha
	ringCol := canvas.RGB{R: 0.72, G: 0.76, B: 0.86}

	x0, x1 := int(cx-rx-3), int(cx+rx+3)
	y0, y1 := int(cy-ry-3), int(cy+ry+3)
	for y := y0; y < y1; y++ {
		if y < 0 || y >= c.H {
			continue
		}
		dyn := (float64(y) - cy) / ry
		for x := x0; x < x1; x++ {
			if x < 0 || x >= c.W {
				continue
			}
			dxn := (float64(x) - cx) / rx
			dn := math.Sqrt(dxn*dxn + dyn*dyn)
			// Soft ring band around dn == 1, plus a whisper of inner light.
			band := math.Exp(-math.Pow((dn-1)/0.075, 2))
			inner := math.Exp(-math.Pow(dn/0.55, 2)) * 0.08
			a := (band*0.22 + inner) * fade
			if a > 0.01 {
				c.Add(x, y, ringCol, float32(a))
			}
		}
	}

	// Phase label just below the ring.
	label := breathName[m.breath.phase]
	cw := utf8.RuneCountInString(label)
	colStart := int(cx)/2 - cw/2
	if colStart < 0 {
		colStart = 0
	}
	row := (int(cy+ry) + 2) / 2
	if row >= c.CH-1 {
		row = c.CH - 2
	}
	c.Overlay(canvas.Overlay{
		Row:   row,
		Col:   colStart,
		Text:  label,
		Color: canvas.RGB{R: 0.62, G: 0.66, B: 0.76},
		Alpha: float32(fade) * 0.55,
	})
}
