package pond

import (
	"drift/internal/particles"
	"math/rand"
)

// genGeometry rebuilds all procedurally placed content for the current
// buffer size. It is seeded, so resizing rescales the same scene rather
// than reshuffling it.
func (p *Pond) genGeometry() {
	rnd := rand.New(rand.NewSource(int64(p.seed)))

	p.genStars(rnd)
	p.genHills(rnd)
	p.genLilies(rnd)
	p.genReeds(rnd)

	// Fireflies hover over the water, mostly in the lower two thirds.
	wh := p.H - p.horizonY
	p.flies.SetBounds(
		float64(p.W)*0.05, float64(p.horizonY)-float64(wh)*0.15,
		float64(p.W)*0.95, float64(p.H)*0.92,
		rnd.Float64,
	)

	// Koi starts resting; next glide will target the new region.
	p.koi.state = koiHidden
	p.koi.placed = false
	p.koi.cool = 4 + rnd.Float64()*10

	// Don't let a resize strand an off-screen ripple.
	p.ripples = particles.NewRipples(6)
}
