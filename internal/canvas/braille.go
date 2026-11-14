package canvas

// Braille is a fine-resolution overlay for particles. The dot grid lives at
// 2×2 dots per buffer pixel (each buffer pixel is one terminal column wide and
// two buffer rows tall; a braille cell covers exactly one buffer cell with a
// 2×4 dot matrix).
//
// Compositing rule: a braille cell replaces the half-block glyph, so its
// background is the *average of the two underlying scene pixels* — fireflies
// glow into the scene instead of punching dark holes in it.
type Braille struct {
	DW, DH int // dot grid: 2× buffer pixels
	A      []float32
	C      []RGB
}

func (b *Braille) Resize(wPixels, hPixels int) {
	dw, dh := wPixels*2, hPixels*2
	b.DW, b.DH = dw, dh
	n := dw * dh
	if cap(b.A) >= n {
		b.A = b.A[:n]
		b.C = b.C[:n]
	} else {
		b.A = make([]float32, n)
		b.C = make([]RGB, n)
	}
	b.Clear()
}

func (b *Braille) Clear() {
	for i := range b.A {
		b.A[i] = 0
	}
}

// Dot plots a single dot at buffer-pixel coords (continuous), clamped in.
func (b *Braille) Dot(px, py float64, alpha float32, col RGB) {
	if alpha <= 0 {
		return
	}
	dx := int(px * 2)
	dy := int(py * 2)
	if dx < 0 || dy < 0 || dx >= b.DW || dy >= b.DH {
		return
	}
	i := dy*b.DW + dx
	if alpha > b.A[i] {
		b.A[i] = alpha
		b.C[i] = col
	}
}

// dotBits maps (dotCol 0..1, dotRow 0..3) → braille bit.
// Unicode braille layout:
//
//	1 4      0x01 0x08
//	2 5      0x02 0x10
//	3 6      0x04 0x20
//	7 8      0x40 0x80
var dotBits = [4][2]rune{
	{0x01, 0x08},
	{0x02, 0x10},
	{0x04, 0x20},
	{0x40, 0x80},
}

// cell returns the braille rune and averaged foreground color for a buffer
// cell (cx = pixel column, cy = buffer-pixel row covering two buffer rows).
// ok is false when no dot is set.
func (b *Braille) cell(cx, cy int) (r rune, col RGB, ok bool) {
	dox := cx * 2
	doy := cy * 4
	var bits rune
	var wsum float32
	var acc RGB
	for row := 0; row < 4; row++ {
		for colm := 0; colm < 2; colm++ {
			di := (doy+row)*b.DW + dox + colm
			a := b.A[di]
			if a <= 0 {
				continue
			}
			bits |= dotBits[row][colm]
			acc.R += b.C[di].R * a
			acc.G += b.C[di].G * a
			acc.B += b.C[di].B * a
			wsum += a
		}
	}
	if bits == 0 {
		return 0, RGB{}, false
	}
	if wsum > 0 {
		acc.R /= wsum
		acc.G /= wsum
		acc.B /= wsum
	}
	return 0x2800 | bits, acc, true
}
