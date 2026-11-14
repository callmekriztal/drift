package pond

import (
	"drift/internal/canvas"
	"drift/internal/noise"
	"drift/internal/palette"
	"drift/internal/particles"
	"math"
	"math/rand"
)

func (p *Pond) genHills(rnd *rand.Rand) {
	p.hillFar = make([]int16, p.W)
	p.hillNear = make([]int16, p.W)
	for x := 0; x < p.W; x++ {
		fh := 4 + 13*noise.V1(float64(x)*0.012, p.seed+21) + 5*noise.V1(float64(x)*0.05, p.seed+22)
		nh := 1 + 6*noise.V1(float64(x)*0.02+40, p.seed+31)
		p.hillFar[x] = int16(clampf(fh, 0, float64(p.horizonY)-1))
		p.hillNear[x] = int16(clampf(nh, 0, float64(p.horizonY)-2))
	}
	// Conifer silhouettes on the near ridge.
	nTrees := p.W / 22
	if nTrees < 2 {
		nTrees = 2
	}
	for i := 0; i < nTrees; i++ {
		tx := rnd.Intn(p.W)
		th := 4 + rnd.Float64()*8
		tw := 1.5 + rnd.Float64()*2.5
		for dx := -int(tw + 1); dx <= int(tw+1); dx++ {
			x := (tx + dx + p.W) % p.W
			fall := math.Max(0, 1-math.Abs(float64(dx))/tw)
			h := th * math.Pow(fall, 1.25)
			if int16(h) > p.hillNear[x] {
				p.hillNear[x] = int16(h)
			}
		}
	}
}

func (p *Pond) hills(c *canvas.Canvas, cur palette.Sampled) {
	// Gentle parallax: the near ridge drifts a little faster than the far one.
	offF := (int(p.t*0.35)%p.W + p.W) % p.W
	offN := (int(p.t*0.80)%p.W + p.W) % p.W
	for x := 0; x < p.W; x++ {
		if h := p.hillFar[(x+offF)%p.W]; h > 0 {
			top := p.horizonY - int(h)
			for y := top; y < p.horizonY; y++ {
				c.Set(x, y, cur.HillFar)
			}
		}
	}
	for x := 0; x < p.W; x++ {
		if h := p.hillNear[(x+offN)%p.W]; h > 0 {
			top := p.horizonY - int(h)
			for y := top; y < p.horizonY; y++ {
				c.Set(x, y, cur.HillNear)
			}
		}
	}
}

func (p *Pond) genLilies(rnd *rand.Rand) {
	n := p.W / 26
	if n < 3 {
		n = 3
	}
	if n > 8 {
		n = 8
	}
	p.lilies = p.lilies[:0]
	for i := 0; i < n; i++ {
		rx := 2.5 + rnd.Float64()*math.Max(3, float64(p.W)*0.022)
		p.lilies = append(p.lilies, lily{
			fx:     0.06 + 0.88*rnd.Float64(),
			fy:     0.22 + 0.72*rnd.Float64(),
			rx:     rx,
			ry:     rx * 0.36,
			ph:     rnd.Float64() * 6.28,
			flower: rnd.Float64() < 0.4,
			fph:    rnd.Float64() * 6.28,
		})
	}
}

type lily struct {
	fx, fy float64
	rx, ry float64
	ph     float64
	flower bool
	fph    float64
}

func (p *Pond) renderLilies(c *canvas.Canvas, cur palette.Sampled) {
	if len(p.lilies) == 0 {
		return
	}
	wh := p.H - p.horizonY
	mx, my, _ := p.moonPos(p.hourFn())
	pad := canvas.Mix(cur.HillNear, cur.HillFar, 0.45).Mul(canvas.RGB{R: 0.60, G: 1.0, B: 0.74})
	for i := range p.lilies {
		l := &p.lilies[i]
		bob := math.Sin(p.t*0.4+l.ph) * 1.2
		sway := math.Sin(p.t*0.27+l.ph*1.7) * 1.0
		cx := l.fx*float64(p.W) + sway
		cy := float64(p.horizonY) + l.fy*float64(wh) + bob

		x0, x1 := int(cx-l.rx-1), int(cx+l.rx+2)
		y0, y1 := int(cy-l.ry-1), int(cy+l.ry+2)
		for y := y0; y < y1; y++ {
			if y < p.horizonY || y >= p.H {
				continue
			}
			dyn := (float64(y) - cy) / l.ry
			for x := x0; x < x1; x++ {
				if x < 0 || x >= p.W {
					continue
				}
				dxn := (float64(x) - cx) / l.rx
				e := dxn*dxn + dyn*dyn
				if e > 1 {
					continue
				}
				alpha := float32(glslSmooth(1.0, 0.82, e)) * 0.93
				c.Blend(x, y, pad, alpha)
				// Faint sheen so pads stay readable at night.
				if cur.MoonA > 0.05 {
					c.Add(x, y, cur.Moon, float32(0.045*cur.MoonA)*alpha)
				}
				// Moonlit rim on the pad edge facing the moon.
				if cur.MoonA > 0.05 && e > 0.72 {
					nx, ny := dxn, dyn*0.5
					nl := math.Sqrt(nx*nx + ny*ny)
					if nl > 0.001 {
						tmx, tmy := mx-cx, my-cy
						tl := math.Sqrt(tmx*tmx + tmy*tmy)
						rdot := (nx*tmx + ny*tmy) / (nl * tl)
						if rdot > 0 {
							c.Add(x, y, cur.Moon, float32(rdot*cur.MoonA*0.15*glslSmooth(1.0, 0.6, e)))
						}
					}
				}
			}
		}
		// A pale blossom breathing on one pad.
		if l.flower {
			fx := cx + l.rx*0.25
			fy := cy - l.ry*0.35
			petal := canvas.Mix(cur.Moon, canvas.RGBf(1.0, 0.72, 0.78), 0.45)
			fa := 0.55 + 0.18*math.Sin(p.t*0.5+l.fph)
			c.Glow(fx, fy, 2.2, 0.30*fa, petal)
			c.Blend(int(fx), int(fy), petal, float32(fa))
		}
	}
}

func (p *Pond) genReeds(rnd *rand.Rand) {
	n := p.W / 6
	if n < 6 {
		n = 6
	}
	if n > 34 {
		n = 34
	}
	p.reeds = p.reeds[:0]
	for i := 0; i < n; i++ {
		var fx float64
		if rnd.Float64() < 0.62 {
			// clustered along the edges
			if rnd.Float64() < 0.5 {
				fx = 0.02 + 0.15*rnd.Float64()
			} else {
				fx = 0.83 + 0.15*rnd.Float64()
			}
		} else {
			fx = 0.25 + 0.5*rnd.Float64()
		}
		p.reeds = append(p.reeds, reed{
			fx:      fx,
			h:       float64(p.H) * (0.16 + 0.34*rnd.Float64()),
			ph:      rnd.Float64() * 6.28,
			ph2:     rnd.Float64() * 6.28,
			lean:    (rnd.Float64() - 0.5) * float64(p.W) * 0.03,
			w:       0.8 + rnd.Float64()*1.4,
			cattail: rnd.Float64() < 0.3,
		})
	}
}

type reed struct {
	fx      float64
	h       float64
	ph, ph2 float64
	lean    float64
	w       float64
	cattail bool
}

func (p *Pond) renderReeds(c *canvas.Canvas, cur palette.Sampled) {
	gust := 0.5 + 0.5*math.Sin(p.t*0.055+1.3)
	for i := range p.reeds {
		r := &p.reeds[i]
		bx := r.fx * float64(p.W)
		sway := (math.Sin(p.t*0.5+r.ph) + 0.45*math.Sin(p.t*0.23+r.ph2)) *
			r.h * 0.055 * (1 + 0.8*gust)
		p0x, p0y := bx, float64(p.H)-1
		p1x, p1y := bx+r.lean*0.25, float64(p.H)-1-r.h*0.55
		p2x, p2y := bx+r.lean+sway, float64(p.H)-1-r.h
		steps := 20
		for s := 0; s <= steps; s++ {
			u := float64(s) / float64(steps)
			x, y := quadBezier(p0x, p0y, p1x, p1y, p2x, p2y, u)
			w := r.w * (1 - 0.4*u)
			c.Blend(int(x), int(y), cur.Reed, 0.92)
			if w > 1.3 {
				c.Blend(int(x)+1, int(y), cur.Reed, 0.55)
			}
			if r.cattail && u > 0.60 && u < 0.80 {
				// seed head: a soft dark capsule
				c.Blend(int(x)+1, int(y), cur.Reed, 0.8)
				c.Blend(int(x)-1, int(y), cur.Reed, 0.8)
				c.Blend(int(x), int(y)-1, cur.Reed, 0.8)
				c.Blend(int(x), int(y)+1, cur.Reed, 0.8)
			}
		}
	}
}

func quadBezier(x0, y0, x1, y1, x2, y2, u float64) (float64, float64) {
	v := 1 - u
	x := v*v*x0 + 2*v*u*x1 + u*u*x2
	y := v*v*y0 + 2*v*u*y1 + u*u*y2
	return x, y
}

// renderFlies: fireflies glow into the scene via additive splats and leave
// fine braille trails.
func (p *Pond) renderFlies(c *canvas.Canvas, cur palette.Sampled) {
	fa := cur.FlyA * p.fade
	if fa <= 0.02 {
		return
	}
	flyCol := canvas.RGBf(0.70, 0.90, 0.42)
	for i := range p.flies.Flies {
		fl := &p.flies.Flies[i]
		b := fl.B * fa
		if b <= 0.03 {
			continue
		}
		c.Glow(fl.X, fl.Y, 1.9, 0.30*b, flyCol)
		c.Br.Dot(fl.X, fl.Y, float32(math.Min(1, b*1.6)), flyCol)
		for j := 0; j < particles.TrailN; j++ {
			tx, ty, ok := fl.TrailAt(j)
			if !ok {
				continue
			}
			a := b * 0.35 * (1 - float64(j)/float64(particles.TrailN))
			c.Br.Dot(tx, ty, float32(a), flyCol)
		}
	}
}
