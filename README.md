# drift

> A calm, dreamy ambient scene for a terminal corner — a moonlit pond that
> breathes with your evening. Zero productivity, all quiet.

**demo GIF placeholder** — *(record with the included tape: `vhs demo.tape`)*

---

`drift` renders a living "Moonlit Pond": a sky gradient that follows your local
time, a glowing moon, twinkling stars, silhouetted hills, shimmering water,
bobbing lily pads, wandering fireflies, a shy koi passing underneath, and
foreground reeds swaying in a slow wind. Every few minutes a faint whisper
fades in, holds, and fades out.

Leave it open in a corner of your terminal. It asks nothing of you.

## Install

Requires Go 1.22+.

```sh
go install drift/cmd/drift@latest
```

Or from a checkout of this repository:

```sh
go install ./cmd/drift
# or: go build -o drift ./cmd/drift
```

Then run it — no config needed:

```sh
drift
```

Drift renders in truecolor when your terminal advertises it (`COLORTERM`),
and degrades gracefully to the 256-color cube otherwise.

## Key bindings

| Key     | What it does                                             |
| ------- | -------------------------------------------------------- |
| `b`     | Breathing guide (4-7-8: inhale · hold · exhale)           |
| `p`     | Cycle palette (Rosé Pine Dawn, Nord Night, Lavender Haze) |
| `w`     | New whisper now                                          |
| `h`     | Toggle the hint bar                                      |
| `q` / `Ctrl+C` | Quit                                              |
| `Esc`   | Dismiss the breathing guide                              |

The hint bar shows itself for 5 seconds after launch, then quietly fades.

## Flags

```
--palette   nord-night | rose-pine-dawn | lavender-haze
--fps       animation rate (default 24)
--hour      override the local hour (0..24), e.g. --hour 22 — handy for
            screenshots and demos; the scene otherwise follows wall-clock
            time, so it looks different at dawn than at midnight
--no-ai     disable AI whispers even if DRIFT_API_KEY is set
```

Your palette choice is saved to the user config dir and restored next time.

## AI whispers (optional)

By default whispers come from a built-in set of ~50 handwritten lines. If you
want model-written whispers, point Drift at any OpenAI-compatible endpoint:

```sh
export DRIFT_API_KEY=sk-...
export DRIFT_BASE_URL=https://api.example.com/v1   # optional (default OpenAI)
export DRIFT_MODEL=glm-4.7-flash                   # optional
drift
```

Requests happen in a background goroutine with a short timeout — rendering
never waits on the network. Results are cached on disk under the user cache
dir, sanitized (reasoning blocks, quotes and filler are stripped), and any
failure falls back silently to the built-in set.

## Development

```sh
go test ./...          # color math, canvas conversion, whisper sanitizer
go build ./cmd/framedump && go run ./cmd/framedump --hour 22
                       # render a frame to PNG without a terminal (dev tool)
```

## License

MIT.
# drift
