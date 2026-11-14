package palette

import (
	"drift/internal/canvas"
	"math"
	"testing"
)

func TestOklabRoundTrip(t *testing.T) {
	for _, hex := range []string{"#88C0D0", "#2E3440", "#FAF4ED", "#BF616A", "#010203", "#FEDCBA"} {
		orig := canvas.Hex(hex)
		back := LabFromLinear(orig).Linear()
		const eps = 0.004
		if math.Abs(float64(orig.R-back.R)) > eps ||
			math.Abs(float64(orig.G-back.G)) > eps ||
			math.Abs(float64(orig.B-back.B)) > eps {
			t.Fatalf("oklab round trip failed for %s: %+v → %+v", hex, orig, back)
		}
	}
}

func TestOklabInterpolationStaysInGamut(t *testing.T) {
	a := LabFromLinear(canvas.Hex("#1D2740"))
	b := LabFromLinear(canvas.Hex("#C08255"))
	for i := 0; i <= 10; i++ {
		c := MixLab(a, b, float64(i)/10).Linear()
		if c.R < -0.01 || c.G < -0.01 || c.B < -0.01 {
			t.Fatalf("interpolation went out of gamut: %+v", c)
		}
	}
}

func TestSkyLUTSampleBoundaries(t *testing.T) {
	lut := BuildSkyLUT(NordNight(), 64)
	for _, probe := range [][2]float64{{0, 0}, {0, 1}, {12, 0.5}, {23.99, 1}} {
		c := lut.At(probe[0], probe[1])
		if math.IsNaN(float64(c.R)) || math.IsNaN(float64(c.G)) || math.IsNaN(float64(c.B)) {
			t.Fatalf("NaN from LUT at %v", probe)
		}
		if c.R < 0 || c.G < 0 || c.B < 0 {
			t.Fatalf("negative sky color at %v: %+v", probe, c)
		}
	}
	// Zenith must be darker than the horizon at night.
	night := lut.At(22, 0)
	horizon := lut.At(22, 1)
	if night.R+night.G+night.B >= horizon.R+horizon.G+horizon.B {
		t.Fatalf("night zenith should be darker than horizon")
	}
}

func TestSkyLUTDayNightCycle(t *testing.T) {
	lut := BuildSkyLUT(NordNight(), 64)
	brightness := func(h float64) float64 {
		c := lut.At(h, 0.5)
		return float64(c.R + c.G + c.B)
	}
	night := brightness(1)
	day := brightness(13)
	if day <= night {
		t.Fatalf("day sky (%f) should be brighter than night (%f)", day, night)
	}
	// No sudden jumps between adjacent 15-min samples.
	prev := brightness(0)
	for h := 0.25; h <= 24; h += 0.25 {
		cur := brightness(h)
		if math.Abs(cur-prev) > 0.35 {
			t.Fatalf("sky jumped %.3f → %.3f around hour %.2f", prev, cur, h)
		}
		prev = cur
	}
}

func TestSampleWrap(t *testing.T) {
	p := NordNight()
	a := p.Sample(24.5)
	b := p.Sample(0.5)
	if math.Abs(a.MoonA-b.MoonA) > 1e-9 {
		t.Fatalf("wrap failed: %+v vs %+v", a, b)
	}
}

func TestPaletteLookup(t *testing.T) {
	for _, name := range []string{"nord-night", "Nord Night", "nord", "rose-pine-dawn", "lavender-haze", "haze"} {
		if Get(name) == nil {
			t.Fatalf("palette %q should resolve", name)
		}
	}
	if Get("nope") != nil {
		t.Fatal("unknown palette should be nil")
	}
}

func TestAllPalettesCycleSanity(t *testing.T) {
	for _, p := range All() {
		prev := p.Sample(0)
		for h := 0.5; h < 24; h += 0.5 {
			s := p.Sample(h)
			d := math.Abs(s.MoonA-prev.MoonA) + math.Abs(s.StarA-prev.StarA)
			if d > 0.55 {
				t.Fatalf("%s: hard cut at hour %.1f (delta %.2f)", p.Name, h, d)
			}
			prev = s
		}
	}
}
