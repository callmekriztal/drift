// Package pond implements Drift's v1 scene: a moonlit pond at night,
// alive with hills, water, lily pads, fireflies, a shy koi and reeds.
package pond

import (
	"drift/internal/canvas"
	"drift/internal/palette"
	"drift/internal/particles"
	"math"
	"math/rand"
)

// Pond is the moonlit-pond scene.
type Pond struct {
	seed     uint32
	rnd      *rand.Rand // gameplay rng (ripple spawns, koi wanderings)
	W, H     int        // buffer pixels
	horizonY int        // waterline pixel row

	t      float64
	fade   float64 // launch fade-in 0..1
	hourFn func() float64

	pal     *palette.Palette
	cur     palette.Sampled
	lut     *palette.SkyLUT
	lutRows int

	// palette crossfade
	lutFrom, lutTo *palette.SkyLUT
	from, to       palette.Sampled
	blendT         float64 // <0 idle; 0..1 crossfading

	stars      []star
	hillFar    []int16
	hillNear   []int16
	lilies     []lily
	reeds      []reed
	ripples    *particles.Ripples
	flies      *particles.Fireflies
	koi        koiState
	rippleTime float64
}

// New builds the scene. hourFn supplies the local hour (0..24) each frame;
// pal is the starting palette.
func New(hourFn func() float64, pal *palette.Palette) *Pond {
	p := &Pond{
		seed:   0x5EED,
		rnd:    rand.New(rand.NewSource(0x5EEDC4FE)),
		hourFn: hourFn,
		fade:   0,
		blendT: -1,
	}
	p.pal = pal
	p.lutRows = 64
	p.lut = palette.BuildSkyLUT(pal, p.lutRows)
	p.cur = pal.Sample(hourFn())

	p.ripples = particles.NewRipples(6)
	p.flies = particles.NewFireflies(9)
	p.koi.state = koiHidden
	p.koi.cool = 5 + p.rnd.Float64()*15
	p.rippleTime = 2.5 + p.rnd.Float64()*6
	return p
}

// SetPalette starts a slow crossfade to another palette family.
func (p *Pond) SetPalette(pal *palette.Palette) {
	if pal == nil || pal.Name == p.pal.Name {
		return
	}
	hour := p.hourFn()
	p.lutFrom = p.lut
	p.lutTo = palette.BuildSkyLUT(pal, p.lutRows)
	p.from = p.cur
	p.to = pal.Sample(hour)
	p.blendT = 0
	p.pal = pal
}

func (p *Pond) Resize(wCells, hCells int) {
	if wCells < 1 || hCells < 1 {
		return // no sane geometry yet; keep the old one
	}
	p.W, p.H = wCells, hCells*2
	p.horizonY = int(float64(p.H) * 0.52)
	if p.horizonY < 4 {
		p.horizonY = 4
	}
	p.genGeometry()
}

func (p *Pond) Update(dt float64) {
	p.t += dt
	if p.fade < 1 {
		p.fade = math.Min(1, p.fade+dt/1.5)
	}

	if p.blendT >= 0 {
		p.blendT += dt / 1.2
		if p.blendT >= 1 {
			p.blendT = -1
			p.lut = p.lutTo
			p.cur = p.to
		} else {
			p.cur = palette.MixSampled(p.from, p.to, smoothstep(p.blendT))
		}
	} else {
		p.cur = p.pal.Sample(p.hourFn())
	}

	// Occasional ripple: something fell off a reed.
	p.rippleTime -= dt
	if p.rippleTime <= 0 {
		p.rippleTime = 8 + p.rnd.Float64()*12
		p.spawnRipple()
	}

	p.ripples.Update(dt)
	p.flies.Update(dt)
	p.koi.update(dt, p)
}

// Render paints the full frame, back to front.
func (p *Pond) Render(c *canvas.Canvas) {
	if c.W != p.W || c.H != p.H {
		p.Resize(c.CW, c.CH)
	}
	c.ClearOverlays()
	hour := p.hourFn()
	cur := p.cur

	p.sky(c, hour)
	p.moon(c, hour, cur)
	p.hills(c, cur)
	p.water(c, hour, cur)
	p.koi.render(c, cur, p)
	p.renderRipples(c, hour, cur)
	p.renderLilies(c, cur)
	p.renderFlies(c, cur)
	p.renderReeds(c, cur)

	if p.fade < 1 {
		f := float32(smoothstep(p.fade))
		c.ScaleAll(f)
	}
}

// WhisperColor exposes the palette's whisper tint for the text overlay.
func (p *Pond) WhisperColor() canvas.RGB { return p.cur.Whisper }

// skyAt is the LUT-sampled sky color (blended during palette crossfades).
func (p *Pond) skyAt(hour, yNorm float64) canvas.RGB {
	if p.blendT >= 0 {
		k := smoothstep(p.blendT)
		return canvas.Mix(p.lutFrom.At(hour, yNorm), p.lutTo.At(hour, yNorm), float32(k))
	}
	return p.lut.At(hour, yNorm)
}

func (p *Pond) spawnRipple() {
	wh := p.H - p.horizonY
	x := (0.08 + 0.84*p.rnd.Float64()) * float64(p.W)
	y := float64(p.horizonY) + (0.25+0.6*p.rnd.Float64())*float64(wh)
	if len(p.lilies) > 0 && p.rnd.Float64() < 0.5 {
		l := &p.lilies[p.rnd.Intn(len(p.lilies))]
		x = l.fx * float64(p.W)
		y = float64(p.horizonY) + l.fy*float64(wh)
	}
	maxR := 5 + p.rnd.Float64()*float64(p.W)*0.05
	p.ripples.Spawn(x, y, maxR, 3.5+p.rnd.Float64()*2.5)
}

func smoothstep(t float64) float64 {
	t = math.Min(1, math.Max(0, t))
	return t * t * (3 - 2*t)
}

// glslSmooth is a signed smoothstep between arbitrary edges.
func glslSmooth(edge0, edge1, x float64) float64 {
	t := (x - edge0) / (edge1 - edge0)
	return smoothstep(t)
}

func clampf(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
