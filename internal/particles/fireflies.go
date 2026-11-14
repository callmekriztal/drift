package particles

import "math"

const (
	TrailN     = 9
	trailEvery = 0.13 // seconds between trail samples
)

// Firefly drifts on a slow Lissajous path with a breathing glow.
type Firefly struct {
	X, Y   float64
	B      float64 // current glow brightness 0..1
	cx, cy float64
	ax, ay float64
	wx, wy float64
	p1, p2 float64
	pw, ph float64 // pulse angular frequency & phase
	t      float64
	trail  [2 * TrailN]float64 // x,y pairs, newest first
	tt     float64
}

// TrailAt returns the j-th trail position (j=0 newest), ok=false when no
// sample has been recorded there yet.
func (fl *Firefly) TrailAt(j int) (x, y float64, ok bool) {
	x, y = fl.trail[j*2], fl.trail[j*2+1]
	return x, y, x != 0 || y != 0
}

// Fireflies is a swarm with independent slow paths.
type Fireflies struct {
	Flies []Firefly
}

func NewFireflies(n int) *Fireflies {
	return &Fireflies{Flies: make([]Firefly, n)}
}

// SetBounds scatters firefly anchor points inside a region (pixel coords).
func (f *Fireflies) SetBounds(x0, y0, x1, y1 float64, rnd func() float64) {
	for i := range f.Flies {
		fl := &f.Flies[i]
		fl.cx = x0 + (x1-x0)*rnd()
		fl.cy = y0 + (y1-y0)*rnd()
		fl.ax = 4 + (x1-x0)*0.06*rnd()
		fl.ay = 2 + (y1-y0)*0.05*rnd()
		fl.wx = 0.11 + 0.10*rnd()
		fl.wy = 0.09 + 0.10*rnd()
		fl.p1 = rnd() * 6.28
		fl.p2 = rnd() * 6.28
		fl.pw = 6.28 / (2.8 + 2.4*rnd())
		fl.ph = rnd() * 6.28
		fl.t = rnd() * 100
	}
}

func (f *Fireflies) Update(dt float64) {
	for i := range f.Flies {
		fl := &f.Flies[i]
		fl.t += dt

		x := fl.cx + fl.ax*math.Sin(fl.wx*fl.t+fl.p1) + fl.ax*0.35*math.Sin(fl.wx*2.33*fl.t+fl.p1*1.6)
		y := fl.cy + fl.ay*math.Sin(fl.wy*fl.t+fl.p2) + fl.ay*0.30*math.Sin(fl.wy*1.71*fl.t+fl.p2*0.8)

		pulse := math.Sin(fl.pw*fl.t + fl.ph)
		if pulse < 0 {
			pulse = 0
		}
		fl.B = pulse * pulse * pulse // soft square: mostly-off with gentle bloom
		fl.X, fl.Y = x, y

		fl.tt += dt
		if fl.tt >= trailEvery {
			fl.tt = 0
			for j := TrailN - 1; j > 0; j-- {
				fl.trail[j*2] = fl.trail[(j-1)*2]
				fl.trail[j*2+1] = fl.trail[(j-1)*2+1]
			}
			fl.trail[0], fl.trail[1] = x, y
		}
	}
}
