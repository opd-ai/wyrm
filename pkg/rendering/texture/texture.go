// Package texture provides procedural texture generation.
package texture

import (
	"image/color"
)

// Texture represents a procedurally generated texture.
type Texture struct {
	Width  int
	Height int
	Pixels []color.RGBA
}

// Generate creates a procedural texture of the given size.
func Generate(width, height int) *Texture {
	pixels := make([]color.RGBA, width*height)
	for i := range pixels {
		pixels[i] = color.RGBA{R: 64, G: 64, B: 64, A: 255}
	}
	return &Texture{Width: width, Height: height, Pixels: pixels}
}
