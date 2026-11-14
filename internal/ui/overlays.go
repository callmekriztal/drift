package ui

import (
	"drift/internal/canvas"
	"unicode/utf8"
)

// overlays paints the text layers: whisper line, hint bar, breathing guide.
func (m *Model) overlays() {
	c := m.cv
	m.whisperOverlay(c)
	m.breathOverlay(c)
	m.hintsOverlay(c) // hint bar last: it sits at the very bottom
}

// whisperOverlay: one faint centered line a few rows above the bottom.
func (m *Model) whisperOverlay(c *canvas.Canvas) {
	text, alpha := m.whisper.Visible()
	if text == "" || alpha <= 0.01 {
		return
	}
	col := m.pond.WhisperColor()
	row := c.CH - 3
	cw := utf8.RuneCountInString(text)
	colStart := (c.CW - cw) / 2
	if colStart < 0 {
		colStart = 0
	}
	c.Overlay(canvas.Overlay{
		Row:   row,
		Col:   colStart,
		Text:  text,
		Color: col,
		Alpha: float32(alpha) * 0.85,
	})
}

// hintState drives the auto-hiding control hints at the bottom.
type hintState struct {
	visible bool
	forced  bool // toggled explicitly with `h`
	timer   float64
	alpha   float64
}

func (h *hintState) start() {
	h.visible = true
	h.timer = 5.0
}

func (h *hintState) toggle() {
	h.forced = !h.forced
	h.visible = h.forced
	if h.visible {
		h.timer = 1 << 30 // stay until toggled off
	}
}

func (h *hintState) update(dt float64) {
	if h.visible && !h.forced {
		h.timer -= dt
		if h.timer <= 0 {
			h.visible = false
		}
	}
	target := 0.0
	if h.visible {
		target = 1.0
	}
	// Slow fade both ways; nothing snaps.
	na := h.alpha + (target-h.alpha)*(dt/0.8)
	h.alpha = clamp01(na)
}

func (m *Model) hintsOverlay(c *canvas.Canvas) {
	if m.hints.alpha <= 0.01 {
		return
	}
	text := "b breathe · p palette · w whisper · h hints · q quit"
	col := canvas.RGB{R: 0.44, G: 0.47, B: 0.53}
	cw := utf8.RuneCountInString(text)
	colStart := (c.CW - cw) / 2
	if colStart < 0 {
		colStart = 0
	}
	c.Overlay(canvas.Overlay{
		Row:   c.CH - 1,
		Col:   colStart,
		Text:  text,
		Color: col,
		Alpha: float32(m.hints.alpha) * 0.75,
	})
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
