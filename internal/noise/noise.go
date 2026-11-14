package noise

import "math"

// Hash returns a deterministic value in [0,1) for an integer lattice point.
func Hash(i int, seed uint32) float32 {
	h := uint32(i)*2654435761 ^ seed
	h ^= h >> 16
	h *= 0x7FEB352D
	h ^= h >> 15
	h *= 0x846CA68B
	h ^= h >> 16
	return float32(h&0xFFFFFF) / 16777216.0
}

// Hash2 returns a deterministic value in [0,1) for a 2D lattice point.
func Hash2(x, y int, seed uint32) float32 {
	h := uint32(x)*374761393 + uint32(y)*668265263 ^ seed
	h ^= h >> 13
	h *= 1274126177
	h ^= h >> 16
	return float32(h&0xFFFFFF) / 16777216.0
}

func smooth(t float64) float64 {
	// Quintic fade: C2-continuous, no visible lattice creases.
	return t * t * t * (t*(t*6-15) + 10)
}

// V1 is smooth 1D value noise.
func V1(x float64, seed uint32) float64 {
	i0 := math.Floor(x)
	f := smooth(x - i0)
	a := float64(Hash(int(i0), seed))
	b := float64(Hash(int(i0)+1, seed))
	return a + (b-a)*f
}

// V2 is smooth 2D value noise.
func V2(x, y float64, seed uint32) float64 {
	ix, iy := math.Floor(x), math.Floor(y)
	fx, fy := smooth(x-ix), smooth(y-iy)
	a := float64(Hash2(int(ix), int(iy), seed))
	b := float64(Hash2(int(ix)+1, int(iy), seed))
	c := float64(Hash2(int(ix), int(iy)+1, seed))
	d := float64(Hash2(int(ix)+1, int(iy)+1, seed))
	return (a + (b-a)*fx) + ((c+(d-c)*fx)-(a+(b-a)*fx))*fy
}

// EaseOutCubic: fast start, gentle settle.
func EaseOutCubic(t float64) float64 {
	t = clamp(t, 0, 1)
	return 1 - math.Pow(1-t, 3)
}

// EaseInOutSine: symmetric soft ease.
func EaseInOutSine(t float64) float64 {
	t = clamp(t, 0, 1)
	return 0.5 - 0.5*math.Cos(math.Pi*t)
}

// Smoothstep clamps and eases t in [0,1].
func Smoothstep(t float64) float64 {
	t = clamp(t, 0, 1)
	return t * t * (3 - 2*t)
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
