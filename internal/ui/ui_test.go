package ui

import (
	"drift/internal/canvas"
	"drift/internal/palette"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func teaKey(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

const frameDt = 1.0 / 24.0

func stepN(m *Model, n int) {
	for i := 0; i < n; i++ {
		m.step(frameDt)
	}
}

// TestHintBarVisibleAndStable checks the hint bar shows within a second and
// that its text renders contiguously (no per-cell escapes inside words).
func TestHintBarVisibleAndStable(t *testing.T) {
	m := New(Options{
		Palette:   palette.NordNight(),
		FPS:       24,
		Hour:      22,
		NoAI:      true,
		ColorMode: canvas.ColorTrue,
	})
	m.width, m.height = 80, 24
	stepN(m, 30)
	view := m.View()
	plain := stripANSI(view)
	rows := strings.Split(plain, "\n")
	if !strings.Contains(rows[len(rows)-1], "b breathe") {
		t.Errorf("hint bar not contiguous in plain last row: %q", rows[len(rows)-1])
	}
	if !strings.Contains(rows[len(rows)-1], "q quit") {
		t.Errorf("hint bar tail missing")
	}
}

func stripANSI(s string) string {
	var b strings.Builder
	inEsc := false
	for _, r := range s {
		switch {
		case r == '\x1b':
			inEsc = true
		case inEsc && r == 'm':
			inEsc = false
		case !inEsc:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// TestHintBarAutoHides verifies the hint bar fades out after ~5s and stays
// gone unless toggled.
func TestHintBarAutoHides(t *testing.T) {
	m := New(Options{
		Palette:   palette.NordNight(),
		FPS:       24,
		Hour:      22,
		NoAI:      true,
		ColorMode: canvas.ColorTrue,
	})
	m.width, m.height = 80, 24
	stepN(m, 30)
	stepN(m, 24*14) // well past the 5s timer + full fade
	if m.hints.alpha > 0.02 {
		t.Fatalf("hint bar should have faded, alpha=%f", m.hints.alpha)
	}
	// Toggle back on with `h`.
	m.Update(teaKey("h"))
	stepN(m, 30)
	if m.hints.alpha < 0.5 {
		t.Fatalf("hints should be visible after toggle, alpha=%f", m.hints.alpha)
	}
}

// TestPaletteCycles cycles `p` through all three palettes.
func TestPaletteCycles(t *testing.T) {
	m := New(Options{
		Palette:   palette.NordNight(),
		FPS:       24,
		Hour:      22,
		NoAI:      true,
		ColorMode: canvas.ColorTrue,
	})
	m.width, m.height = 80, 24
	want := []string{"rose-pine-dawn", "lavender-haze", "nord-night"}
	for _, name := range want {
		m.Update(teaKey("p"))
		if m.pals[m.palIdx].Name != name {
			t.Errorf("after p: got %s want %s", m.pals[m.palIdx].Name, name)
		}
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// TestKeysToggleOverlays exercises every interaction key for panics and
// expected state changes.
func TestKeysToggleOverlays(t *testing.T) {
	m := New(Options{
		Palette:   palette.NordNight(),
		FPS:       24,
		Hour:      22,
		NoAI:      true,
		ColorMode: canvas.ColorTrue,
	})
	m.width, m.height = 80, 24

	if m.breath.on {
		t.Fatal("breathing should start off")
	}
	m.Update(teaKey("b"))
	stepN(m, 20)
	if !m.breath.on || m.breath.alpha < 0.3 {
		t.Fatalf("breathing guide should be fading in")
	}
	m.Update(teaKey("esc"))
	if m.breath.on {
		t.Fatal("esc should dismiss the breathing guide")
	}

	m.Update(teaKey("w"))
	stepN(m, 5)
	text, alpha := m.whisper.Visible()
	if text == "" || alpha <= 0 {
		t.Fatal("`w` should start a whisper immediately")
	}

	m.Update(teaKey("h"))
	if !m.hints.forced || !m.hints.visible {
		t.Fatal("`h` should pin the hint bar")
	}
	m.Update(teaKey("h"))
	if m.hints.forced {
		t.Fatal("second `h` should unpin")
	}
}

// TestWhisperAppears verifies a whisper fades in and renders as text.
func TestWhisperAppears(t *testing.T) {
	m := New(Options{
		Palette:   palette.NordNight(),
		FPS:       24,
		Hour:      22,
		NoAI:      true,
		ColorMode: canvas.ColorTrue,
	})
	m.width, m.height = 80, 24
	m.whisper.FastForwardForTest(1)
	stepN(m, 24*4) // past the 3s fade-in
	_, alpha := m.whisper.Visible()
	if alpha < 0.9 {
		t.Fatalf("whisper should be fully visible, alpha=%f", alpha)
	}
	text, _ := m.whisper.Visible()
	plain := stripANSI(m.View())
	if !strings.Contains(plain, text) {
		t.Errorf("whisper text %q not found in frame", text)
	}
}
