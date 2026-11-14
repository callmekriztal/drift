package pond

import (
	"drift/internal/canvas"
	"drift/internal/palette"
	"math"
	"math/rand"
)

type star struct {
	x, y   float64
	b      float64 // base brightness
	tw     float64 // twinkle speed (cycles/sec)
	ph     float64
	bright bool
}

func (p *Pond) genStars(rnd *rand.Rand) {
	n := p.W * p.horizonY / 260
	if n < 8 {
		n = 8
	}
	if n > 220 {
		n = 220
	}
	p.stars = p.stars[:0]
	for i := 0; i < n; i++ {
		p.stars = append(p.stars, star{
			x:      rnd.Float64() * float64(p.W),
			y:      rnd.Float64() * float64(p.horizonY) * 0.92,
			b:      0.25 + rnd.Float64()*0.75,
			tw:     0.02 + rnd.Float64()*0.10,
			ph:     rnd.Float64() * 6.28,
			bright: rnd.Float64() < 0.06,
		})
	}
}

func (p *Pond) sky(c *canvas.Canvas, hour float64) {
	// Gradient: one LUT lerp per row, then a row fill.
	for y := 0; y < p.horizonY; y++ {
		yn := float64(y) / float64(p.horizonY)
		row := p.skyAt(hour, yn)
		slice := c.Pix[y*c.W : y*c.W+c.W]
		for i := range slice {
			slice[i] = row
		}
	}
}

func (p *Pond) renderStars(c *canvas.Canvas, cur palette.Sampled) {
	if cur.StarA <= 0.01 {
		return
	}
	for i := range p.stars {
		s := &p.stars[i]
		a := s.b * (0.42 + 0.58*(0.5+0.5*math.Sin(2*math.Pi*p.t*s.tw+s.ph)))
		a *= cur.StarA
		if a <= 0.03 {
			continue
		}
		x, y := int(s.x), int(s.y)
		c.Blend(x, y, cur.Moon, float32(a*0.85))
		if s.bright && a > 0.30 {
			// four-point glint on the brightest stars
			c.Blend(x-1, y, cur.Moon, float32(a*0.30))
			c.Blend(x+1, y, cur.Moon, float32(a*0.30))
			c.Blend(x, y-1, cur.Moon, float32(a*0.30))
			c.Blend(x, y+1, cur.Moon, float32(a*0.30))
		}
	}
}

// moonPos: the moon glides a slow arc across the sky over the day.
func (p *Pond) moonPos(hour float64) (mx, my, r float64) {
	mx = float64(p.W) * (0.5 + 0.42*math.Sin(2*math.Pi*(hour-2)/24))
	my = float64(p.horizonY) * 0.30
	r = math.Max(2.8, float64(p.W)*0.036)
	return
}

func (p *Pond) moon(c *canvas.Canvas, hour float64, cur palette.Sampled) {
	p.renderStars(c, cur)
	if cur.MoonA <= 0.01 {
		return
	}
	mx, my, r := p.moonPos(hour)
	mc := cur.Moon

	// Three-layer halo: tight core, mid skirt, very wide faint skirt.
	c.Glow(mx, my, r*2.2, 0.40*cur.MoonA, mc)
	c.Glow(mx, my, r*5.0, 0.16*cur.MoonA, mc)
	c.Glow(mx, my, r*9.0, 0.05*cur.MoonA, mc)

	// Disc with terminator shading and a few craters.
	x0, x1 := int(mx-r-1), int(mx+r+2)
	y0, y1 := int(my-r-1), int(my+r+2)
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			dx := float64(x) - mx
			dy := float64(y) - my
			d := math.Sqrt(dx*dx + dy*dy)
			if d > r*1.05 {
				continue
			}
			edge := glslSmooth(r*1.05, r*0.86, d)
			// Lit from the upper left; lower right falls into shadow.
			nd := (dx*0.55 + dy*0.835) / r
			shade := 0.13 * (0.5 + 0.5*nd)
			// Craters (fixed relative positions, subtle).
			nx, ny := dx/r, dy/r
			for _, cr := range [3][2]float64{{0.30, -0.20}, {-0.26, 0.16}, {0.05, 0.38}} {
				cdx, cdy := nx-cr[0], ny-cr[1]
				if cdx*cdx+cdy*cdy < 0.045 {
					shade += 0.06
				}
			}
			b := (1 - math.Min(0.3, shade)) * edge
			c.Blend(x, y, mc, float32(b*cur.MoonA*0.97))
		}
	}
}
