// Package raycast provides the first-person raycasting renderer.
package raycast

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// Renderer handles first-person raycasting and draws to an Ebitengine image.
type Renderer struct {
	Width  int
	Height int
}

// NewRenderer creates a new raycasting renderer.
func NewRenderer(width, height int) *Renderer {
	return &Renderer{Width: width, Height: height}
}

// Draw renders the current view to the screen.
func (r *Renderer) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 20, G: 12, B: 28, A: 255})
}
