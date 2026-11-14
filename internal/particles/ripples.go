// Package particles holds reusable particle systems: expanding ripples and
// fireflies on Lissajous paths.
package particles

import "math"

// Ripple is one expanding ring on the water.
type Ripple struct {
	X, Y   float64
	MaxR   float64
	Life   float64
	Age    float64
	Active bool
}

// Ripples is a fixed-capacity pool.
type Ripples struct {
	items []Ripple
}

func NewRipples(capacity int) *Ripples {
	return &Ripples{items: make([]Ripple, capacity)}
}

// Spawn activates a ripple, evicting the oldest if the pool is full.
func (r *Ripples) Spawn(x, y, maxR, life float64) {
	oldest := 0
	for i := range r.items {
		if !r.items[i].Active {
			r.items[i] = Ripple{X: x, Y: y, MaxR: maxR, Life: life, Active: true}
			return
		}
		if r.items[i].Age > r.items[oldest].Age {
			oldest = i
		}
	}
	r.items[oldest] = Ripple{X: x, Y: y, MaxR: maxR, Life: life, Active: true}
}

func (r *Ripples) Update(dt float64) {
	for i := range r.items {
		it := &r.items[i]
		if !it.Active {
			continue
		}
		it.Age += dt
		if it.Age >= it.Life {
			it.Active = false
		}
	}
}

// Each visits active ripples with their eased radius and fading strength
// (u: 0→1 over lifetime, radius: easeOutCubic, alpha: linear fade).
func (r *Ripples) Each(fn func(x, y, radius, alpha float64)) {
	for i := range r.items {
		it := &r.items[i]
		if !it.Active {
			continue
		}
		u := it.Age / it.Life
		radius := it.MaxR * (1 - math.Pow(1-u, 3))
		alpha := 1 - u
		fn(it.X, it.Y, radius, alpha)
	}
}
