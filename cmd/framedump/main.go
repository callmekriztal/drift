// Command framedump renders pond frames offscreen to PNG files for visual
// tuning. It is a developer tool and not part of the drift binary.
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"strings"
	"time"

	"drift/internal/canvas"
	"drift/internal/palette"
	"drift/internal/scene/pond"
)

func main() {
	hour := flag.Float64("hour", 22, "local hour (0..24)")
	w := flag.Int("w", 96, "width in cells")
	h := flag.Int("h", 30, "height in cells")
	pal := flag.String("palette", "nord-night", "palette name")
	frames := flag.Int("frames", 1, "number of frames to simulate")
	step := flag.Float64("step", 0.0, "seconds of simulated time between frames")
	out := flag.String("out", "frame.png", "output png (frames get -N suffix)")
	flag.Parse()

	p := palette.Get(*pal)
	if p == nil {
		fmt.Fprintln(os.Stderr, "unknown palette:", *pal)
		os.Exit(1)
	}
	fn := func() float64 { return *hour }
	s := pond.New(fn, p)
	c := canvas.New(*w, *h)
	s.Resize(*w, *h)

	for i := 0; i < *frames; i++ {
		s.Update(*step)
	}
	// Warm up a bit so fade-in doesn't hide the scene.
	for i := 0; i < 60; i++ {
		s.Update(1.0 / 24.0)
	}
	s.Render(c)

	img := image.NewRGBA(image.Rect(0, 0, c.W, c.H))
	for y := 0; y < c.H; y++ {
		for x := 0; x < c.W; x++ {
			col := c.Pix[y*c.W+x]
			img.SetRGBA(x, y, color.RGBA{
				R: to8(canvas.GammaEncode(col.R)),
				G: to8(canvas.GammaEncode(col.G)),
				B: to8(canvas.GammaEncode(col.B)),
				A: 255,
			})
		}
	}
	name := *out
	if *frames > 1 {
		name = fmt.Sprintf("%s", strings.TrimSuffix(*out, ".png")) + ".png"
	}
	f, err := os.Create(name)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		panic(err)
	}
	fmt.Println("wrote", name, "at", time.Now().Format(time.TimeOnly), "hour", *hour)
}

func to8(v float32) uint8 {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	if math.IsNaN(float64(v)) {
		return 0
	}
	return uint8(v*255 + 0.5)
}
