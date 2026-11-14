package palette

import (
	"drift/internal/canvas"
	"math"
)

// Lab is a color in Oklab space (Björn Ottosson's perceptual color space).
// Sky gradients are interpolated here so transitions stay perceptually even
// and never go muddy.
type Lab struct{ L, A, B float64 }

// Linear sRGB → Oklab (works on linear light directly, no gamma step).
func LabFromLinear(c canvas.RGB) Lab {
	l := 0.4122214708*float64(c.R) + 0.5363325363*float64(c.G) + 0.0514459929*float64(c.B)
	m := 0.2119034982*float64(c.R) + 0.6806995451*float64(c.G) + 0.1073969566*float64(c.B)
	s := 0.0883024619*float64(c.R) + 0.2817188376*float64(c.G) + 0.6299787005*float64(c.B)
	l_, m_, s_ := math.Cbrt(l), math.Cbrt(m), math.Cbrt(s)
	return Lab{
		L: 0.2104542553*l_ + 0.7936177850*m_ - 0.0040720468*s_,
		A: 1.9779984951*l_ - 2.4285922050*m_ + 0.4505937099*s_,
		B: 0.0259040371*l_ + 0.7827717662*m_ - 0.8086757660*s_,
	}
}

// Oklab → linear sRGB.
func (l Lab) Linear() canvas.RGB {
	l_ := l.L + 0.3963377774*l.A + 0.2158037573*l.B
	m_ := l.L - 0.1055613458*l.A - 0.0638541728*l.B
	s_ := l.L - 0.0894841775*l.A - 1.2914855480*l.B
	lCub, mCub, sCub := l_*l_*l_, m_*m_*m_, s_*s_*s_
	return canvas.RGB{
		R: lmsToLin(4.0767416621*lCub - 3.3077115913*mCub + 0.2309699292*sCub),
		G: lmsToLin(-1.2684380046*lCub + 2.6097574011*mCub - 0.3413193965*sCub),
		B: lmsToLin(-0.0041960863*lCub - 0.7034186147*mCub + 1.7076147010*sCub),
	}
}

func lmsToLin(v float64) float32 {
	if v < 0 {
		v = 0
	}
	return float32(v)
}

func MixLab(a, b Lab, t float64) Lab {
	return Lab{a.L + (b.L-a.L)*t, a.A + (b.A-a.A)*t, a.B + (b.B-a.B)*t}
}
