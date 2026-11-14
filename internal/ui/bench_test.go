package ui

import (
	"drift/internal/canvas"
	"drift/internal/palette"
	"testing"
)

// BenchmarkFrame measures the full pipeline: scene update + pixel render +
// ANSI encoding at a typical terminal size. Budget: well under 10ms.
func BenchmarkFrame(b *testing.B) {
	m := New(Options{
		Palette:   palette.NordNight(),
		FPS:       24,
		Hour:      22,
		NoAI:      true,
		ColorMode: canvas.ColorTrue,
	})
	m.width, m.height = 100, 30
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		m.step(1.0 / 24.0)
		_ = m.View()
	}
}

func BenchmarkFrameSmall(b *testing.B) {
	m := New(Options{
		Palette:   palette.NordNight(),
		FPS:       24,
		Hour:      22,
		NoAI:      true,
		ColorMode: canvas.ColorTrue,
	})
	m.width, m.height = 60, 20
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		m.step(1.0 / 24.0)
		_ = m.View()
	}
}
