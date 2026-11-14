// Package ui is Drift's Bubble Tea model: the fixed-tick render loop,
// key handling, and the overlay layers (hints, breathing, whisper text).
package ui

import (
	"bytes"
	"drift/internal/canvas"
	"drift/internal/palette"
	"drift/internal/scene/pond"
	"drift/internal/whisper"
	"encoding/json"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type tickMsg struct{}

// Options configures a Model.
type Options struct {
	Palette    *palette.Palette
	FPS        int
	Hour       float64 // <0 → follow wall clock
	NoAI       bool
	ConfigPath string
	ColorMode  canvas.ColorMode
}

// Model is the root Bubble Tea model.
type Model struct {
	opts     Options
	width    int
	height   int
	cv       *canvas.Canvas
	cvW, cvH int
	pond     *pond.Pond
	whisper  *whisper.Manager
	hints    hintState
	breath   breathState

	hourFn  func() float64
	pals    []*palette.Palette
	palIdx  int
	last    time.Time
	frame   bytes.Buffer
	quitted bool
}

// New builds the root model.
func New(opts Options) *Model {
	if opts.FPS < 4 || opts.FPS > 60 {
		opts.FPS = 24
	}
	m := &Model{opts: opts, pals: palette.All()}
	for i, p := range m.pals {
		if p.Name == opts.Palette.Name {
			m.palIdx = i
			break
		}
	}
	if opts.Hour >= 0 {
		h := opts.Hour
		m.hourFn = func() float64 { return h }
	} else {
		m.hourFn = wallHour
	}
	m.pond = pond.New(m.hourFn, opts.Palette)

	var ai *whisper.AI
	if !opts.NoAI {
		ai = whisper.NewAI()
	}
	cache, _ := whisper.LoadCache()
	m.whisper = whisper.New(ai, cache)
	m.whisper.SetHourFn(m.hourFn)
	m.hints.start()
	return m
}

func wallHour() float64 {
	n := time.Now()
	return float64(n.Hour()) + float64(n.Minute())/60.0 + float64(n.Second())/3600.0
}

func (m *Model) Init() tea.Cmd {
	return tickCmd(m.opts.FPS)
}

func tickCmd(fps int) tea.Cmd {
	return tea.Every(time.Second/time.Duration(fps), func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if msg.Width < 4 || msg.Height < 4 {
			return m, nil
		}
		m.width, m.height = msg.Width, msg.Height
		m.pond.Resize(msg.Width, msg.Height)
		return m, nil

	case tickMsg:
		now := time.Now()
		dt := now.Sub(m.last).Seconds()
		m.last = now
		if dt < 0 || dt > 0.25 {
			dt = 1.0 / float64(m.opts.FPS)
		}
		m.step(dt)
		return m, tickCmd(m.opts.FPS)

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitted = true
			return m, tea.Quit
		case "esc":
			if m.breath.on {
				m.breath.toggle()
			}
		case "b":
			m.breath.toggle()
		case "p":
			m.palIdx = (m.palIdx + 1) % len(m.pals)
			m.pond.SetPalette(m.pals[m.palIdx])
			m.saveConfig()
		case "w":
			m.whisper.Now()
		case "h":
			m.hints.toggle()
		}
	}
	return m, nil
}

func (m *Model) step(dt float64) {
	m.pond.Update(dt)
	m.whisper.Update(dt)
	m.hints.update(dt)
	m.breath.update(dt)
}

func (m *Model) saveConfig() {
	if m.opts.ConfigPath == "" {
		return
	}
	_ = os.MkdirAll(dirOf(m.opts.ConfigPath), 0o755)
	data, _ := json.MarshalIndent(map[string]string{"palette": m.pals[m.palIdx].Name}, "", "  ")
	_ = os.WriteFile(m.opts.ConfigPath, data, 0o644)
}

func dirOf(p string) string {
	i := strings.LastIndexByte(p, '/')
	if i < 0 {
		return "."
	}
	return p[:i]
}

// View renders the frame: the scene plus text overlays.
func (m *Model) View() string {
	if m.quitted {
		return ""
	}
	if m.width < 4 || m.height < 4 {
		return ""
	}
	if m.frame.Cap() == 0 {
		m.frame.Grow(1 << 18)
	}
	m.renderFrame()
	return m.frame.String()
}

func (m *Model) renderFrame() {
	m.pond.Render(m.canvas())
	m.overlays()
	m.canvas().Render(m.opts.ColorMode, &m.frame)
}
