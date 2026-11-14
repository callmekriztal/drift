package whisper

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSanitizePlain(t *testing.T) {
	got := Sanitize("  the pond remembers the moon  \n")
	if got != "the pond remembers the moon" {
		t.Fatalf("got %q", got)
	}
}

func TestSanitizeStripsReasoningBlock(t *testing.T) {
	in := "<think>\nthe model rambles about ponds here, a long reasoning trace.\n</think>\nstill water keeps the sky"
	got := Sanitize(in)
	if got != "still water keeps the sky" {
		t.Fatalf("got %q", got)
	}
}

func TestSanitizeStripsDanglingTag(t *testing.T) {
	in := "gentle line one\n<cortex_dump>truncated reasoning without a close tag"
	got := Sanitize(in)
	if !strings.HasPrefix(got, "gentle line one") {
		t.Fatalf("got %q", got)
	}
}

func TestSanitizeStripsQuotesAndMarkdown(t *testing.T) {
	cases := [][2]string{
		{`"a calm line"`, "a calm line"},
		{"**bold calm line**", "bold calm line"},
		{"`code calm`", "code calm"},
		{"Sure — here is one: soft light falls", "soft light falls"},
	}
	for _, c := range cases {
		if got := Sanitize(c[0]); got != c[1] {
			t.Errorf("sanitize(%q) = %q, want %q", c[0], got, c[1])
		}
	}
}

func TestSanitizeRejectsLongAndEmpty(t *testing.T) {
	long := strings.Repeat("word ", 30) // 150 chars
	if got := Sanitize(long); got != "" {
		t.Fatalf("long line should be rejected, got %q", got)
	}
	if got := Sanitize(""); got != "" {
		t.Fatalf("empty should be empty")
	}
	if got := Sanitize("   \n\n  "); got != "" {
		t.Fatalf("blank should be empty")
	}
	if got := Sanitize("a"); got != "" {
		t.Fatalf("single char should be rejected")
	}
}

func TestSanitizeCollapsesWhitespace(t *testing.T) {
	got := Sanitize("the   night\tis  wide")
	if got != "the night is wide" {
		t.Fatalf("got %q", got)
	}
}

func TestTimeOfDayName(t *testing.T) {
	cases := map[float64]string{
		0: "night", 6: "dawn", 9: "day", 12: "day", 18: "dusk", 22: "night", 4.9: "night",
	}
	for h, want := range cases {
		if got := TimeOfDayName(h); got != want {
			t.Errorf("TimeOfDayName(%v) = %q, want %q", h, got, want)
		}
	}
}

func TestCacheRoundTrip(t *testing.T) {
	dir := t.TempDir()
	c := NewCache(dir)
	c.Add("one calm line")
	c.Add("two calm lines")
	c.Add("one calm line") // dedup

	// Reload from disk.
	c2 := NewCache(dir)
	if len(c2.Items) != 2 {
		t.Fatalf("expected 2 cached items, got %d", len(c2.Items))
	}
	s, ok := c2.Pick("none")
	if !ok || (s != "one calm line" && s != "two calm lines") {
		t.Fatalf("pick failed: %q %v", s, ok)
	}
	// Exclude is honored when possible.
	for i := 0; i < 10; i++ {
		if s, _ := c2.Pick("one calm line"); s == "one calm line" && len(c2.Items) > 1 {
			t.Fatalf("exclude not honored: %q", s)
		}
	}
}

func TestCacheEmptyDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "does-not-exist")
	c := NewCache(dir)
	if _, ok := c.Pick(""); ok {
		t.Fatal("empty cache should not pick")
	}
	c.Add("survives missing dir") // MkdirAll inside Add
	c2 := NewCache(dir)
	if len(c2.Items) != 1 {
		t.Fatalf("expected persistence, got %d items", len(c2.Items))
	}
	_ = os.Remove(filepath.Join(dir, "drift", "whispers.json"))
}

func TestFallbackSetQuality(t *testing.T) {
	if len(Fallback) < 45 {
		t.Fatalf("fallback set should have ~50 lines, has %d", len(Fallback))
	}
	for _, line := range Fallback {
		if n := len([]rune(line)); n > 80 || n < 10 {
			t.Fatalf("fallback line out of range (%d runes): %q", n, line)
		}
		if strings.ContainsAny(line, "<>") {
			t.Fatalf("fallback line has markup: %q", line)
		}
	}
}
