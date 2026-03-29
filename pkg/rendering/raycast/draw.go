//go:build !noebiten

package raycast

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// Draw renders the current view to the screen using DDA raycasting.
func (r *Renderer) Draw(screen *ebiten.Image) {
	r.drawBackground(screen)
	r.drawWalls(screen)
}

// drawBackground fills the ceiling and floor.
func (r *Renderer) drawBackground(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 20, G: 12, B: 28, A: 255})
	floorColor := color.RGBA{R: 40, G: 35, B: 45, A: 255}
	for y := r.Height / 2; y < r.Height; y++ {
		for x := 0; x < r.Width; x++ {
			screen.Set(x, y, floorColor)
		}
	}
}

// drawWalls casts rays and renders wall columns.
func (r *Renderer) drawWalls(screen *ebiten.Image) {
	for x := 0; x < r.Width; x++ {
		r.drawWallColumn(screen, x)
	}
}

// drawWallColumn renders a single vertical wall strip.
func (r *Renderer) drawWallColumn(screen *ebiten.Image, x int) {
	cameraX := 2.0*float64(x)/float64(r.Width) - 1.0
	rayAngle := r.PlayerA + cameraX*(r.FOV/2)
	rayDirX := math.Cos(rayAngle)
	rayDirY := math.Sin(rayAngle)

	distance, wallType := r.castRay(rayDirX, rayDirY)
	distance *= math.Cos(cameraX * (r.FOV / 2)) // Fix fisheye

	if distance < MinWallDistance {
		distance = MinWallDistance
	}
	wallHeight := int(float64(r.Height) / distance)
	if wallHeight > r.Height {
		wallHeight = r.Height
	}

	drawStart := (r.Height - wallHeight) / 2
	drawEnd := drawStart + wallHeight
	wallColor := r.getWallColor(wallType, distance)

	for y := drawStart; y < drawEnd; y++ {
		if y >= 0 && y < r.Height {
			screen.Set(x, y, wallColor)
		}
	}
}
