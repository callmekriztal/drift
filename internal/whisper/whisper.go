// Package whisper provides Drift's text layer: a scheduler that fades one
// gentle line in and out, an offline fallback set, an optional
// OpenAI-compatible client, and a disk cache.
package whisper

import (
	"math/rand"
	"sync"
	"time"
)

type phase int

const (
	phaseHidden phase = iota
	phaseIn           // fading in
	phaseHold         // holding
	phaseOut          // fading out
)

const (
	fadeInSeconds  = 3.0
	holdSeconds    = 8.0
	fadeOutSeconds = 4.0
)

// Manager owns the current whisper line and its fade cycle.
type Manager struct {
	fallback []string
	ai       *AI
	cache    *Cache
	rnd      *rand.Rand

	mu          sync.Mutex
	line        string
	last        string
	phase       phase
	timer       float64 // seconds remaining in current phase
	alpha       float64
	nextIn      float64 // seconds until next cycle when hidden
	lastAIfetch time.Time
	fetching    bool
	results     chan string
	hourFn      func() float64
}

// New builds a manager. When ai is non-nil, fetched lines are cached and
// rotate into the mix.
func New(ai *AI, cache *Cache) *Manager {
	m := &Manager{
		fallback: Fallback,
		ai:       ai,
		cache:    cache,
		rnd:      rand.New(rand.NewSource(time.Now().UnixNano())),
		results:  make(chan string, 1),
		hourFn:   func() float64 { return 22 },
	}
	m.nextIn = 4 + m.rnd.Float64()*6 // first whisper arrives soon but not instantly
	return m
}

// Update advances the fade cycle; call once per frame. It never blocks.
func (m *Manager) Update(dt float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Drain a finished AI fetch (non-blocking).
	select {
	case text := <-m.results:
		m.fetching = false
		if text != "" && m.cache != nil {
			m.cache.Add(text)
		}
	default:
	}

	switch m.phase {
	case phaseHidden:
		m.nextIn -= dt
		if m.nextIn <= 0 {
			m.startLocked()
		}
	case phaseIn:
		m.alpha += dt / fadeInSeconds
		if m.alpha >= 1 {
			m.alpha = 1
			m.phase = phaseHold
			m.timer = holdSeconds
		}
	case phaseHold:
		m.timer -= dt
		if m.timer <= 0 {
			m.phase = phaseOut
		}
	case phaseOut:
		m.alpha -= dt / fadeOutSeconds
		if m.alpha <= 0 {
			m.alpha = 0
			m.phase = phaseHidden
			m.nextIn = 60 + m.rnd.Float64()*60
		}
	}

	// Keep a fresh AI line warming in the background (throttled).
	if m.ai != nil && !m.fetching && time.Since(m.lastAIfetch) > 5*time.Minute {
		m.fetch()
	}
}

// Visible returns the current text and fade alpha, if any.
func (m *Manager) Visible() (string, float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.phase == phaseHidden || m.alpha <= 0 {
		return "", 0
	}
	return m.line, m.alpha
}

// Now starts a whisper cycle immediately (the `w` key).
func (m *Manager) Now() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.phase == phaseHidden || m.phase == phaseOut {
		m.startLocked()
	}
}

func (m *Manager) startLocked() {
	m.line = m.pickLocked()
	m.last = m.line
	m.phase = phaseIn
	m.alpha = 0
}

// pickLocked prefers a cached AI line, then the built-in set, never repeating.
func (m *Manager) pickLocked() string {
	if m.cache != nil {
		if s, ok := m.cache.Pick(m.last); ok {
			return s
		}
	}
	for tries := 0; tries < 8; tries++ {
		s := m.fallback[m.rnd.Intn(len(m.fallback))]
		if s != m.last {
			return s
		}
	}
	return m.fallback[0]
}

// fetch spawns a background request; results land on m.results. Rendering
// never waits on it.
func (m *Manager) fetch() {
	m.fetching = true
	m.lastAIfetch = time.Now()
	ai := m.ai
	res := m.results
	timeOfDay := TimeOfDayName(m.hourFn())
	go func() {
		text, err := ai.Fetch(timeOfDay)
		if err != nil {
			res <- ""
			return
		}
		res <- text
	}()
}

// FastForwardForTest jumps the fade cycle ahead — used by tests.
func (m *Manager) FastForwardForTest(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nextIn = 0
	_ = d
}

// SetHourFn lets the manager phrase AI requests with the right time of day.
func (m *Manager) SetHourFn(fn func() float64) {
	if fn != nil {
		m.mu.Lock()
		m.hourFn = fn
		m.mu.Unlock()
	}
}
