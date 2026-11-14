package canvas

import (
	"math"
	"strconv"
	"strings"
)

// RGB is a color in linear light space (channels nominally 0..1, may exceed 1
// transiently for additive glow; everything is clamped at quantization time).
type RGB struct {
	R, G, B float32
}

// RGBf builds an RGB from float components in 0..1 (sRGB).
func RGBf(r, g, b float64) RGB {
	return srgbToLinear(RGB{float32(r), float32(g), float32(b)})
}

// Hex parses a "#rrggbb" (or "rrggbb") string into linear-light RGB.
func Hex(s string) RGB {
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return RGB{}
	}
	r, _ := strconv.ParseUint(s[0:2], 16, 32)
	g, _ := strconv.ParseUint(s[2:4], 16, 32)
	b, _ := strconv.ParseUint(s[4:6], 16, 32)
	return srgbToLinear(RGB{
		float32(r) / 255,
		float32(g) / 255,
		float32(b) / 255,
	})
}

func srgbToLinear(c RGB) RGB {
	return RGB{srgbToLinear1(c.R), srgbToLinear1(c.G), srgbToLinear1(c.B)}
}

func srgbToLinear1(v float32) float32 {
	if v <= 0.04045 {
		return v / 12.92
	}
	return float32(math.Pow(float64(v+0.055)/1.055, 2.4))
}

// GammaEncode converts linear light to the sRGB 0..1 ramp used by terminals.
func GammaEncode(v float32) float32 {
	if v <= 0.0031308 {
		return 12.92 * v
	}
	return float32(1.055*math.Pow(float64(v), 1.0/2.4) - 0.055)
}

func (c RGB) Clamp() RGB {
	return RGB{clamp01(c.R), clamp01(c.G), clamp01(c.B)}
}

func clamp01(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func (c RGB) Scale(s float32) RGB { return RGB{c.R * s, c.G * s, c.B * s} }

// Mul treats o as a per-channel multiplier (for tints).
func (c RGB) Mul(o RGB) RGB { return RGB{c.R * o.R, c.G * o.G, c.B * o.B} }

func (c RGB) Add(o RGB) RGB { return RGB{c.R + o.R, c.G + o.G, c.B + o.B} }

// Mix linearly interpolates a→b by t.
func Mix(a, b RGB, t float32) RGB {
	return RGB{
		a.R + (b.R-a.R)*t,
		a.G + (b.G-a.G)*t,
		a.B + (b.B-a.B)*t,
	}
}

// hashUnit returns a deterministic value in [-0.5, 0.5] for a pixel position.
// Used for static dithering: seeded by position so it never changes between
// frames, which keeps quantized cells byte-identical and lets the terminal
// skip unchanged lines.
func hashUnit(x, y int) float32 {
	h := uint32(x)*374761393 + uint32(y)*668265263 + 0x9E3779B9
	h ^= h >> 13
	h *= 1274126177
	h ^= h >> 16
	return float32(h&0xFFFF)/65535.0 - 0.5
}
