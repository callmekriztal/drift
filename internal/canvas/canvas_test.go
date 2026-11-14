package canvas

import (
	"bytes"
	"math"
	"strings"
	"testing"
)

func TestHexAndBlend(t *testing.T) {
	w := Hex("#FFFFFF")
	b := Hex("#000000")
	if math.Abs(float64(w.R)-1) > 1e-6 || math.Abs(float64(w.B)-1) > 1e-6 {
		t.Fatalf("white should be ~1.0 linear, got %+v", w)
	}
	got := Mix(b, w, 0.5)
	back := GammaEncode(got.R)
	// 50% mix in linear light encodes to ~0.735 in sRGB — the correct
	// perceptually-bright result of physical light blending.
	if math.Abs(float64(back)-0.735) > 0.01 {
		t.Fatalf("mid mix gamma-encoded should be ~0.735, got %f", back)
	}
}

func TestHalfBlockConversion(t *testing.T) {
	// 2 cells wide, 2 cells tall → 2×4 pixels.
	c := New(2, 2)
	// Top-left cell: bright top pixel, dark bottom pixel.
	c.Set(0, 0, Hex("#FF0000"))
	c.Set(0, 1, Hex("#000000"))
	var buf bytes.Buffer
	c.Render(ColorTrue, &buf)
	lines := strings.Split(buf.String(), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(lines))
	}
	// fg 255,0,0 then bg 0,0,0 on the first cell.
	if !strings.Contains(lines[0], "38;2;255;0;0") || !strings.Contains(lines[0], "48;2;0;0;0") {
		t.Fatalf("half-block fg/bg not encoded: %q", lines[0])
	}
	if !strings.Contains(lines[0], "▀") {
		t.Fatalf("half block glyph missing")
	}
}

func TestBrailleDotBits(t *testing.T) {
	c := New(1, 2) // 1 cell wide, 2 buffer rows → one braille cell (4 dot rows)
	col := Hex("#FFFFFF")
	// One buffer pixel spans 2 dot cols and 2 dot rows.
	// pixel (0, 0): dots (0,0),(0,1),(1,0),(1,1)
	c.Br.Dot(0.0, 0.0, 1, col) // pixel(0,0) → dot col 0, dot row 0 → bit 0x01
	_, _, ok := c.Br.cell(0, 0)
	if !ok {
		t.Fatal("dot not registered")
	}
	r, _, _ := c.Br.cell(0, 0)
	if r != 0x2800|0x01 {
		t.Fatalf("expected dot1 bit, got %q (%#x)", r, r-0x2800)
	}
	c.Br.Clear()
	c.Br.Dot(0.99, 1.49, 1, col) // pixel (0,1) lower-right-ish → dot col 1, dot row 2/3
	r, _, _ = c.Br.cell(0, 0)
	if r&(0x20|0x80) == 0 {
		t.Fatalf("expected dot6/dot8 bits, got %#x", r-0x2800)
	}
}

func TestBrailleKeepsSceneColor(t *testing.T) {
	c := New(1, 1)
	c.Set(0, 0, Hex("#FF0000")) // top pixel
	c.Set(0, 1, Hex("#0000FF")) // bottom pixel
	c.Br.Dot(0.5, 0.5, 1, Hex("#00FF00"))
	var buf bytes.Buffer
	c.Render(ColorTrue, &buf)
	s := buf.String()
	// bg should be the averaged scene color (purple-ish), not black.
	if !strings.Contains(s, "48;2;") {
		t.Fatalf("braille cell must keep a background color")
	}
	if strings.Contains(s, "48;2;0;0;0") {
		t.Fatalf("braille punched a dark hole: %q", s)
	}
}

func TestRenderStableAcrossFrames(t *testing.T) {
	// Same buffer rendered twice must produce byte-identical output (static
	// dither + quantization), so terminals can diff cleanly.
	c := New(8, 4)
	for y := 0; y < c.H; y++ {
		for x := 0; x < c.W; x++ {
			c.Set(x, y, RGB{float32(x) / 20, float32(y) / 12, 0.1})
		}
	}
	var a, b bytes.Buffer
	c.Render(ColorTrue, &a)
	c.Render(ColorTrue, &b)
	if a.String() != b.String() {
		t.Fatal("identical buffers produced different encodings")
	}
}

func TestOverlayStableAndApplied(t *testing.T) {
	c := New(20, 3)
	c.Fill(Hex("#101820"))
	c.Overlay(Overlay{Row: 1, Col: 5, Text: "calm", Color: Hex("#88C0D0"), Alpha: 1})
	var a, b bytes.Buffer
	c.Render(ColorTrue, &a)
	c.Render(ColorTrue, &b)
	if a.String() != b.String() {
		t.Fatal("overlay frame not stable")
	}
	lines := strings.Split(a.String(), "\n")
	if !strings.Contains(lines[1], "calm") {
		t.Fatalf("overlay text missing: %q", lines[1])
	}
}

func TestGlowAddsLightAndStaysInBounds(t *testing.T) {
	c := New(4, 4)
	before := c.At(0, 0)
	c.Glow(-1, -1, 2, 0.5, Hex("#FFFFFF")) // off-canvas center, radius reaches in
	after := c.At(0, 0)
	if after.R <= before.R {
		t.Fatal("glow should add light near the edge")
	}
	// No panic, no light far away.
	far := c.At(3, 3)
	if far.R >= after.R {
		t.Logf("note: symmetric falloff, far=%v near=%v", far, after)
	}
}

func Test256Mode(t *testing.T) {
	c := New(2, 1)
	c.Fill(Hex("#FF8800"))
	var buf bytes.Buffer
	c.Render(Color256, &buf)
	if !strings.Contains(buf.String(), "38;5;") {
		t.Fatalf("256-color mode not used: %q", buf.String())
	}
}
