package ui

import "drift/internal/canvas"

// canvas lazily allocates the frame buffer at the current cell size and only
// reconfigures it when the terminal size actually changes.
func (m *Model) canvas() *canvas.Canvas {
	if m.cv == nil {
		m.cv = canvas.New(m.width, m.height)
	} else if m.cvW != m.width || m.cvH != m.height {
		m.cv.Resize(m.width, m.height)
	}
	m.cvW, m.cvH = m.width, m.height
	return m.cv
}
