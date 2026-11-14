package whisper

import (
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Cache persists AI whisper lines under the user cache dir so repeated runs
// build up a pool of gentle lines even when the API is only reachable
// sometimes.
type Cache struct {
	path  string
	mu    sync.Mutex
	Items []Item
}

type Item struct {
	Text string `json:"text"`
	At   int64  `json:"at"` // unix seconds
}

// LoadCache reads the default user cache dir; missing file → empty cache.
func LoadCache() (*Cache, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}
	return NewCache(dir), nil
}

// NewCache builds a cache rooted at dir (drift/whispers.json underneath).
func NewCache(dir string) *Cache {
	c := &Cache{path: filepath.Join(dir, "drift", "whispers.json")}
	data, err := os.ReadFile(c.path)
	if err != nil {
		return c // fresh cache
	}
	_ = json.Unmarshal(data, &c.Items)
	if len(c.Items) > 200 {
		c.Items = c.Items[len(c.Items)-200:]
	}
	return c
}

// Add appends a line, deduped, capped, and best-effort persisted.
func (c *Cache) Add(text string) {
	if text == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, it := range c.Items {
		if it.Text == text {
			return
		}
	}
	c.Items = append(c.Items, Item{Text: text, At: time.Now().Unix()})
	if len(c.Items) > 100 {
		c.Items = c.Items[len(c.Items)-100:]
	}
	if c.path == "" {
		return
	}
	_ = os.MkdirAll(filepath.Dir(c.path), 0o755)
	data, err := json.MarshalIndent(c.Items, "", "  ")
	if err == nil {
		_ = os.WriteFile(c.path, data, 0o644)
	}
}

// Pick returns a random cached line, avoiding `exclude` when possible.
func (c *Cache) Pick(exclude string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.Items) == 0 {
		return "", false
	}
	// Prefer candidates that aren't the line we just showed.
	pool := make([]Item, 0, len(c.Items))
	for _, it := range c.Items {
		if it.Text != exclude {
			pool = append(pool, it)
		}
	}
	if len(pool) == 0 {
		pool = c.Items
	}
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	return pool[rnd.Intn(len(pool))].Text, true
}
