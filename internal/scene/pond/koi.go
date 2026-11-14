package pond

import (
	"drift/internal/canvas"
	"drift/internal/palette"
	"math"
)

const (
	koiHidden = iota
	koiGlide
)

// koiState: one shy koi that glides between rest spots, darkening the water
// only slightly. It is a shadow, not a fish.
type koiState struct {
	state        int
	x, y         float64
	fromX, fromY float64
	tx, ty       float64
	t, dur       float64
	cool         float64
	ph           float64
	placed       bool
}

func (k *koiState) update(dt float64, p *Pond) {
	wh := p.H - p.horizonY
	switch k.state {
	case koiHidden:
		k.cool -= dt
		if k.cool <= 0 {
			if !k.placed {
				k.x = (0.2 + 0.6*p.rnd.Float64()) * float64(p.W)
				k.y = float64(p.horizonY) + (0.3+0.5*p.rnd.Float64())*float64(wh)
				k.placed = true
			}
			k.fromX, k.fromY = k.x, k.y
			k.tx = (0.15 + 0.7*p.rnd.Float64()) * float64(p.W)
			k.ty = float64(p.horizonY) + (0.25+0.6*p.rnd.Float64())*float64(wh)
			k.t = 0
			k.dur = 9 + p.rnd.Float64()*9
			k.state = koiGlide
		}
	case koiGlide:
		k.t += dt
		u := k.t / k.dur
		if u >= 1 {
			k.x, k.y = k.tx, k.ty
			k.state = koiHidden
			k.cool = 15 + p.rnd.Float64()*30
			return
		}
		e := noise_EaseInOutSine(u)
		k.x = k.fromX + (k.tx-k.fromX)*e
		k.y = k.fromY + (k.ty-k.fromY)*e
	}
}

func (k *koiState) render(c *canvas.Canvas, cur palette.Sampled, p *Pond) {
	if k.state != koiGlide {
		return
	}
	u := k.t / k.dur
	env := math.Sin(math.Pi * u) // fade in and out at each end of the glide
	rx := math.Max(4.5, float64(p.W)*0.028)
	ry := rx * 0.30

	dirX, dirY := k.tx-k.fromX, k.ty-k.fromY
	dl := math.Sqrt(dirX*dirX + dirY*dirY)
	if dl < 0.001 {
		return
	}
	dirX, dirY = dirX/dl, dirY/dl
	perpX, perpY := -dirY, dirX
	tailSway := math.Sin(p.t*1.1+k.ph) * rx * 0.6

	// tail sits behind the body, swaying side to side
	tailX := k.x - dirX*rx*1.5 + perpX*tailSway*0.5
	tailY := k.y - dirY*rx*1.5 + perpY*tailSway*0.5
	trx, try := rx*0.85, rx*0.16

	x0 := int(math.Min(k.x, tailX) - rx - 2)
	x1 := int(math.Max(k.x, tailX) + rx + 2)
	y0 := int(math.Min(k.y, tailY) - ry - 3)
	y1 := int(math.Max(k.y, tailY) + ry + 3)

	for y := y0; y < y1; y++ {
		if y < p.horizonY+2 || y >= p.H-1 {
			continue
		}
		for x := x0; x < x1; x++ {
			if x < 0 || x >= p.W {
				continue
			}
			bodyE := ellDist(float64(x)-k.x, float64(y)-k.y, rx, ry)
			tailE := ellDist(float64(x)-tailX, float64(y)-tailY, trx, try)
			var a float64
			if bodyE < 1 {
				a += math.Pow(1-bodyE, 1.4) * 0.55
			}
			if tailE < 1 {
				a += math.Pow(1-tailE, 1.4) * 0.40
			}
			a *= env * 0.5
			if a < 0.015 {
				continue
			}
			// darken the water beneath: blend toward a dimmer version of itself
			px := c.At(x, y)
			c.Blend(x, y, px.Scale(0.45), float32(a))
		}
	}
}

func ellDist(dx, dy, rx, ry float64) float64 {
	return math.Sqrt((dx*dx)/(rx*rx) + (dy*dy)/(ry*ry))
}

// local ease helpers (kept scene-local to avoid a noise dependency cycle here)
func noise_EaseInOutSine(t float64) float64 {
	t = math.Min(1, math.Max(0, t))
	return 0.5 - 0.5*math.Cos(math.Pi*t)
}
