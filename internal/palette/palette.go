// Package palette holds Drift's color palettes and time-of-day interpolation.
// Each palette is a set of keyframes over the local hour; sampling lerps
// between keyframes in Oklab so dawn and dusk stay smooth and pretty.
package palette

import (
	"drift/internal/canvas"
)

// Key is a palette snapshot at a given local hour.
type Key struct {
	Hour     float64
	Zenith   string // sky top
	Horizon  string // sky bottom (at the waterline)
	HillFar  string // distant ridge silhouette
	HillNear string // near ridge + tree silhouettes
	Water    string // reflection tint multiplier
	Moon     string // moon body + glow color
	Reed     string // foreground reed silhouette
	Whisper  string // whisper text tint
	MoonA    float64
	StarA    float64
	FlyA     float64 // firefly presence
}

// Palette is a named time-of-day family.
type Palette struct {
	Name string
	Keys []Key // sorted by Hour, must span 0..24
}

func (p *Palette) normalize() {
	for i := range p.Keys {
		if p.Keys[i].Hour >= 24 {
			p.Keys[i].Hour -= 24
		}
	}
	sortKeys(p.Keys)
}

// Sample returns the palette state at a local hour (0..24).
func (p *Palette) Sample(hour float64) Sampled {
	keys := p.Keys
	n := len(keys)
	hour = wrapHour(hour)

	// Find segment [keys[i], keys[i+1]] (with wraparound to first key +24).
	seg := -1
	for i := 0; i < n-1; i++ {
		if hour >= keys[i].Hour && hour < keys[i+1].Hour {
			seg = i
			break
		}
	}
	var a, b *Key
	var span, rel float64
	if seg < 0 {
		// Wrap: last key → first key (+24).
		a, b = &keys[n-1], &keys[0]
		span = 24 - a.Hour + b.Hour
		rel = hour + 24 - a.Hour
		if span <= 0 {
			span = 24
		}
	} else {
		a, b = &keys[seg], &keys[seg+1]
		span = b.Hour - a.Hour
		rel = hour - a.Hour
	}
	t := smooth(rel / span)
	return mixKeys(a, b, t)
}

func mixKeys(a, b *Key, t float64) Sampled {
	tf := float32(t)
	return Sampled{
		Zenith:   MixLab(LabFromLinear(canvas.Hex(a.Zenith)), LabFromLinear(canvas.Hex(b.Zenith)), t),
		Horizon:  MixLab(LabFromLinear(canvas.Hex(a.Horizon)), LabFromLinear(canvas.Hex(b.Horizon)), t),
		HillFar:  canvas.Mix(canvas.Hex(a.HillFar), canvas.Hex(b.HillFar), tf),
		HillNear: canvas.Mix(canvas.Hex(a.HillNear), canvas.Hex(b.HillNear), tf),
		Water:    canvas.Mix(canvas.Hex(a.Water), canvas.Hex(b.Water), tf),
		Moon:     canvas.Mix(canvas.Hex(a.Moon), canvas.Hex(b.Moon), tf),
		Reed:     canvas.Mix(canvas.Hex(a.Reed), canvas.Hex(b.Reed), tf),
		Whisper:  canvas.Mix(canvas.Hex(a.Whisper), canvas.Hex(b.Whisper), tf),
		MoonA:    a.MoonA + (b.MoonA-a.MoonA)*t,
		StarA:    a.StarA + (b.StarA-a.StarA)*t,
		FlyA:     a.FlyA + (b.FlyA-a.FlyA)*t,
	}
}

// Sampled is one instant of a palette: everything a frame needs.
type Sampled struct {
	Zenith, Horizon    Lab
	HillFar            canvas.RGB
	HillNear           canvas.RGB
	Water              canvas.RGB // multiplier applied to reflected sky
	Moon               canvas.RGB
	Reed               canvas.RGB
	Whisper            canvas.RGB
	MoonA, StarA, FlyA float64
}

// MixSampled blends two sampled palettes (used for palette crossfades).
func MixSampled(a, b Sampled, t float64) Sampled {
	tf := float32(t)
	return Sampled{
		Zenith:   MixLab(a.Zenith, b.Zenith, t),
		Horizon:  MixLab(a.Horizon, b.Horizon, t),
		HillFar:  canvas.Mix(a.HillFar, b.HillFar, tf),
		HillNear: canvas.Mix(a.HillNear, b.HillNear, tf),
		Water:    canvas.Mix(a.Water, b.Water, tf),
		Moon:     canvas.Mix(a.Moon, b.Moon, tf),
		Reed:     canvas.Mix(a.Reed, b.Reed, tf),
		Whisper:  canvas.Mix(a.Whisper, b.Whisper, tf),
		MoonA:    a.MoonA + (b.MoonA-a.MoonA)*t,
		StarA:    a.StarA + (b.StarA-a.StarA)*t,
		FlyA:     a.FlyA + (b.FlyA-a.FlyA)*t,
	}
}

func wrapHour(h float64) float64 {
	h = h - 24.0*float64(int(h/24.0))
	if h < 0 {
		h += 24
	}
	return h
}

func smooth(t float64) float64 {
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

func sortKeys(k []Key) {
	for i := 1; i < len(k); i++ {
		for j := i; j > 0 && k[j].Hour < k[j-1].Hour; j-- {
			k[j], k[j-1] = k[j-1], k[j]
		}
	}
}
