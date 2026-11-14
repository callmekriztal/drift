package pond

import (
	"drift/internal/canvas"
	"drift/internal/noise"
	"drift/internal/palette"
	"math"
)

// water fills the pond: a compressed mirror of the sky, darkened and tinted,
// animated by horizontal wind streaks and broken moonlight.
func (p *Pond) water(c *canvas.Canvas, hour float64, cur palette.Sampled) {
	wt := p.horizonY
	wh := p.H - wt
	if wh <= 0 {
		return
	}
	mx, _, _ := p.moonPos(hour)
	moonA := cur.MoonA * p.fade

	for dy := 0; dy < wh; dy++ {
		y := wt + dy
		depthN := float64(dy) / float64(wh)
		// Mirror the sky above the waterline, slightly compressed.
		sy := float64(wt) - float64(dy)*0.85
		if sy < 0 {
			sy = 0
		}
		refl := p.skyAt(hour, sy/float64(wt)).Mul(cur.Water)
		dark := 0.46 - 0.22*depthN // nearer water is darker
		base := refl.Scale(float32(dark))

		// Per-row shimmer phase: streaks drift slowly sideways.
		rp := float64(noise.Hash(y, p.seed)) * 9.0
		// Glint column under the moon, widening and wobbling with depth.
		wob := (noise.V1(float64(y)*0.5+p.t*0.35, p.seed+11) - 0.5) * (1.5 + 4*depthN)
		gx := mx + wob
		sigma := 1.1 + 5*depthN

		for x := 0; x < p.W; x++ {
			n := noise.V1(float64(x)*0.045+rp+p.t*0.025, p.seed+7)
			f := 1.0 + 0.16*(n-0.5)*2
			col := base.Scale(float32(f))
			// Gentle wind streaks: smooth noise creates long horizontal flecks.
			if n > 0.72 {
				extra := (n - 0.72) / 0.28
				col = col.Add(refl.Scale(float32(extra * 0.16)))
			}
			// broken moonlight column
			if moonA > 0.02 {
				dg := float64(x) - gx
				g := math.Exp(-dg * dg / (2 * sigma * sigma))
				brk := 0.45 + 0.75*noise.V1(float64(x)*0.22+float64(y)*1.3+p.t*0.06, p.seed+13)
				g *= brk * (1 - 0.55*depthN) * (1 - 0.7*(n-0.5)*2)
				if g > 0.01 {
					col = col.Add(cur.Moon.Scale(float32(g * 0.42 * moonA)))
				}
			}
			c.Set(x, y, col)
		}
	}

	p.moonReflection(c, hour, cur)
}

// moonReflection paints the mirrored moon disc: a wobbling, streak-broken
// blob of light deepening toward the viewer.
func (p *Pond) moonReflection(c *canvas.Canvas, hour float64, cur palette.Sampled) {
	if cur.MoonA <= 0.02 {
		return
	}
	wt := p.horizonY
	wh := p.H - wt
	mx, my, mr := p.moonPos(hour)
	my2 := 2*float64(wt) - my
	rv := mr * 1.5 // vertically squashed disc

	y0, y1 := int(my2-rv), int(my2+rv)+1
	for y := y0; y < y1; y++ {
		if y < wt || y >= p.H {
			continue
		}
		dyn := (float64(y) - my2) / rv
		if math.Abs(dyn) > 1 {
			continue
		}
		halfw := mr * math.Sqrt(1-dyn*dyn) * 1.15
		depthFade := 1 - float64(y-wt)/float64(wh)
		wob := (noise.V1(float64(y)*0.7+p.t*0.5, p.seed+17) - 0.5) * (1 + 3*depthFade)
		cx := mx + wob
		a0 := 0.40 * cur.MoonA * (0.35 + 0.65*depthFade) * p.fade
		for x := int(cx - halfw); x <= int(cx+halfw); x++ {
			if x < 0 || x >= p.W {
				continue
			}
			dxn := (float64(x) - cx) / math.Max(1, halfw)
			brk := noise.V1(float64(x)*0.3+float64(y)*1.1+p.t*0.08, p.seed+19)
			if brk < 0.42 {
				continue
			}
			edge := glslSmooth(1.0, 0.5, math.Abs(dxn))
			c.Add(x, y, cur.Moon, float32(a0*edge*brk*0.8))
		}
	}
}

func (p *Pond) renderRipples(c *canvas.Canvas, hour float64, cur palette.Sampled) {
	light := canvas.Mix(cur.Moon, p.skyAt(hour, 0.75), 0.55).Scale(0.45)
	p.ripples.Each(func(x, y, radius, alpha float64) {
		a0 := 0.26 * alpha
		squash := 0.38
		x0, x1 := int(x-radius-2), int(x+radius+2)
		y0, y1 := int(y-radius*squash-2), int(y+radius*squash+2)
		for yy := y0; yy < y1; yy++ {
			if yy < p.horizonY || yy >= p.H {
				continue
			}
			dyn := (float64(yy) - y) / squash
			for xx := x0; xx < x1; xx++ {
				if xx < 0 || xx >= p.W {
					continue
				}
				dx := float64(xx) - x
				d := math.Sqrt(dx*dx + dyn*dyn)
				if d > radius+1.5 {
					continue
				}
				band := math.Exp(-math.Pow((d-radius)/1.35, 2))
				if band < 0.02 {
					continue
				}
				c.Add(xx, yy, light, float32(a0*band))
			}
		}
	})
}
