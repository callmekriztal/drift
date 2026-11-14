// drift — a calm, dreamy ambient scene for a terminal corner. No
// productivity purpose; only a moonlit pond and permission to breathe.
package main

import (
	"drift/internal/canvas"
	"drift/internal/palette"
	"drift/internal/ui"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	var (
		palName = flag.String("palette", "", "palette: nord-night, rose-pine-dawn, lavender-haze")
		fps     = flag.Int("fps", 24, "animation frames per second (4..60)")
		hour    = flag.Float64("hour", -1, "override local hour (0..24), e.g. --hour 22")
		noAI    = flag.Bool("no-ai", false, "disable AI whispers even if DRIFT_API_KEY is set")
	)
	flag.Parse()

	pal := resolvePalette(*palName)
	opts := ui.Options{
		Palette:    pal,
		FPS:        *fps,
		Hour:       *hour,
		NoAI:       *noAI,
		ConfigPath: configPath(),
		ColorMode:  detectColorMode(),
	}

	prog := tea.NewProgram(ui.New(opts), tea.WithAltScreen())
	if _, err := prog.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "drift:", err)
		os.Exit(1)
	}
}

// resolvePalette: --palette flag wins, then the saved config, then Nord Night.
func resolvePalette(flagName string) *palette.Palette {
	if flagName != "" {
		p := palette.Get(flagName)
		if p == nil {
			fmt.Fprintf(os.Stderr, "drift: unknown palette %q (nord-night, rose-pine-dawn, lavender-haze)\n", flagName)
			os.Exit(1)
		}
		return p
	}
	if path := configPath(); path != "" {
		if data, err := os.ReadFile(path); err == nil {
			var cfg struct {
				Palette string `json:"palette"`
			}
			if json.Unmarshal(data, &cfg) == nil && cfg.Palette != "" {
				if p := palette.Get(cfg.Palette); p != nil {
					return p
				}
			}
		}
	}
	return palette.NordNight()
}

func configPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "drift", "config.json")
}

// detectColorMode: truecolor when the terminal advertises it (COLORTERM),
// otherwise degrade gracefully to the 256-color cube.
func detectColorMode() canvas.ColorMode {
	ct := strings.ToLower(os.Getenv("COLORTERM"))
	if ct == "truecolor" || ct == "24bit" || ct == "yes" {
		return canvas.ColorTrue
	}
	// Well-known terminals that always do truecolor.
	term := strings.ToLower(os.Getenv("TERM"))
	for _, t := range []string{"kitty", "alacritty", "wezterm", "ghostty", "foot", "contour"} {
		if strings.Contains(term, t) {
			return canvas.ColorTrue
		}
	}
	if os.Getenv("DRIFT_TRUECOLOR") != "" {
		return canvas.ColorTrue
	}
	return canvas.Color256
}
