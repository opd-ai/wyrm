//go:build !noebiten

package raycast

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// Draw renders the current view to the screen using DDA raycasting.
func (r *Renderer) Draw(screen *ebiten.Image) {
	// Fill background (ceiling and floor)
	screen.Fill(color.RGBA{R: 20, G: 12, B: 28, A: 255})

	// Draw floor (bottom half)
	floorColor := color.RGBA{R: 40, G: 35, B: 45, A: 255}
	for y := r.Height / 2; y < r.Height; y++ {
		for x := 0; x < r.Width; x++ {
			screen.Set(x, y, floorColor)
		}
	}

	// Cast rays for each screen column
	for x := 0; x < r.Width; x++ {
		// Calculate ray direction
		cameraX := 2.0*float64(x)/float64(r.Width) - 1.0
		rayAngle := r.PlayerA + cameraX*(r.FOV/2)

		rayDirX := math.Cos(rayAngle)
		rayDirY := math.Sin(rayAngle)

		// Perform DDA
		distance, wallType := r.castRay(rayDirX, rayDirY)

		// Fix fisheye effect
		distance *= math.Cos(cameraX * (r.FOV / 2))

		// Calculate wall height
		if distance < 0.1 {
			distance = 0.1
		}
		wallHeight := int(float64(r.Height) / distance)
		if wallHeight > r.Height {
			wallHeight = r.Height
		}

		// Calculate draw bounds
		drawStart := (r.Height - wallHeight) / 2
		drawEnd := drawStart + wallHeight

		// Get wall color based on type and distance
		wallColor := r.getWallColor(wallType, distance)

		// Draw wall strip
		for y := drawStart; y < drawEnd; y++ {
			if y >= 0 && y < r.Height {
				screen.Set(x, y, wallColor)
			}
		}
	}
}
