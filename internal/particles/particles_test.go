package particles

import (
	"math"
	"testing"
)

func TestRipplesLifecycle(t *testing.T) {
	r := NewRipples(2)
	r.Spawn(10, 10, 5, 2)
	called := 0
	r.Each(func(x, y, radius, alpha float64) { called++ })
	if called != 1 {
		t.Fatalf("expected 1 active ripple, got %d", called)
	}
	r.Update(1.0)
	var rad float64
	r.Each(func(_, _, radius, _ float64) { rad = radius })
	// easeOutCubic(0.5) = 0.875 → radius 4.375
	if math.Abs(rad-5*0.875) > 1e-9 {
		t.Fatalf("radius not eased correctly: %f", rad)
	}
	r.Update(1.0)
	called = 0
	r.Each(func(x, y, radius, alpha float64) { called++ })
	if called != 0 {
		t.Fatalf("ripple should have expired, %d active", called)
	}
}

func TestRipplesPoolEvictsOldest(t *testing.T) {
	r := NewRipples(1)
	r.Spawn(0, 0, 1, 10)
	r.Update(5)
	r.Spawn(1, 1, 1, 1) // pool full → evicts
	count := 0
	r.Each(func(x, y, radius, alpha float64) { count++ })
	if count != 1 {
		t.Fatalf("pool should hold 1, got %d", count)
	}
}

func TestFirefliesStayInFlight(t *testing.T) {
	f := NewFireflies(4)
	f.SetBounds(0, 10, 100, 90, func() float64 { return 0.5 })
	for i := 0; i < 200; i++ {
		f.Update(0.1)
		for i := range f.Flies {
			fl := &f.Flies[i]
			if fl.X < -50 || fl.X > 150 || fl.Y < -30 || fl.Y > 130 {
				t.Fatalf("fly escaped bounds: %v %v", fl.X, fl.Y)
			}
		}
	}
}

func TestFireflyPulses(t *testing.T) {
	f := NewFireflies(1)
	f.SetBounds(0, 0, 10, 10, func() float64 { return 0.5 })
	sawBright, sawDark := false, false
	for i := 0; i < 600; i++ {
		f.Update(0.05)
		b := f.Flies[0].B
		if b > 0.9 {
			sawBright = true
		}
		if b < 0.01 {
			sawDark = true
		}
	}
	if !sawBright || !sawDark {
		t.Fatalf("fly should pulse: bright=%v dark=%v", sawBright, sawDark)
	}
}
