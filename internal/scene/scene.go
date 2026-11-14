// Package scene defines the Scene contract. A scene is a living view that
// updates in fixed timesteps and paints into a shared pixel canvas.
package scene

import "drift/internal/canvas"

// Scene is implemented by every ambient view in Drift. dt is seconds since
// the previous update; wCells/hCells are terminal cell dimensions.
type Scene interface {
	Update(dt float64)
	Render(c *canvas.Canvas)
	Resize(wCells, hCells int)
}
