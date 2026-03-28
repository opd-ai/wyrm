// Package raycast provides the first-person raycasting renderer.
package raycast

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// Renderer handles first-person raycasting and draws to an Ebitengine image.
type Renderer struct {
	Width    int
	Height   int
	PlayerX  float64
	PlayerY  float64
	PlayerA  float64 // angle in radians
	WorldMap [][]int // 2D map: 0=empty, >0=wall type
	FOV      float64 // field of view in radians
}

// NewRenderer creates a new raycasting renderer.
func NewRenderer(width, height int) *Renderer {
	// Create a simple default world map
	worldMap := make([][]int, 16)
	for i := range worldMap {
		worldMap[i] = make([]int, 16)
		// Create boundary walls
		worldMap[i][0] = 1
		worldMap[i][15] = 1
		if i == 0 || i == 15 {
			for j := range worldMap[i] {
				worldMap[i][j] = 1
			}
		}
	}
	// Add some interior walls
	worldMap[4][4] = 2
	worldMap[4][5] = 2
	worldMap[4][6] = 2
	worldMap[8][8] = 3
	worldMap[8][9] = 3
	worldMap[9][8] = 3

	return &Renderer{
		Width:    width,
		Height:   height,
		PlayerX:  8.0,
		PlayerY:  8.0,
		PlayerA:  0.0,
		WorldMap: worldMap,
		FOV:      math.Pi / 3, // 60 degrees
	}
}

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

// castRay performs DDA raycasting and returns distance and wall type.
func (r *Renderer) castRay(rayDirX, rayDirY float64) (float64, int) {
	mapX := int(r.PlayerX)
	mapY := int(r.PlayerY)

	// Length of ray from one side to next
	var deltaDistX, deltaDistY float64
	if rayDirX == 0 {
		deltaDistX = 1e30
	} else {
		deltaDistX = math.Abs(1 / rayDirX)
	}
	if rayDirY == 0 {
		deltaDistY = 1e30
	} else {
		deltaDistY = math.Abs(1 / rayDirY)
	}

	var stepX, stepY int
	var sideDistX, sideDistY float64

	if rayDirX < 0 {
		stepX = -1
		sideDistX = (r.PlayerX - float64(mapX)) * deltaDistX
	} else {
		stepX = 1
		sideDistX = (float64(mapX) + 1.0 - r.PlayerX) * deltaDistX
	}
	if rayDirY < 0 {
		stepY = -1
		sideDistY = (r.PlayerY - float64(mapY)) * deltaDistY
	} else {
		stepY = 1
		sideDistY = (float64(mapY) + 1.0 - r.PlayerY) * deltaDistY
	}

	// DDA loop
	hit := false
	side := 0 // 0 = NS wall, 1 = EW wall
	maxSteps := 64

	for i := 0; i < maxSteps && !hit; i++ {
		if sideDistX < sideDistY {
			sideDistX += deltaDistX
			mapX += stepX
			side = 0
		} else {
			sideDistY += deltaDistY
			mapY += stepY
			side = 1
		}

		// Check map bounds and wall hit
		if mapX >= 0 && mapX < len(r.WorldMap) &&
			mapY >= 0 && mapY < len(r.WorldMap[0]) {
			if r.WorldMap[mapX][mapY] > 0 {
				hit = true
			}
		} else {
			break
		}
	}

	if !hit {
		return 100.0, 0
	}

	// Calculate perpendicular distance
	var perpWallDist float64
	if side == 0 {
		perpWallDist = sideDistX - deltaDistX
	} else {
		perpWallDist = sideDistY - deltaDistY
	}

	wallType := 0
	if mapX >= 0 && mapX < len(r.WorldMap) &&
		mapY >= 0 && mapY < len(r.WorldMap[0]) {
		wallType = r.WorldMap[mapX][mapY]
	}

	return perpWallDist, wallType
}

// getWallColor returns a color based on wall type and distance.
func (r *Renderer) getWallColor(wallType int, distance float64) color.RGBA {
	// Base colors for different wall types
	var baseR, baseG, baseB uint8
	switch wallType {
	case 1:
		baseR, baseG, baseB = 128, 64, 64 // Red-brown
	case 2:
		baseR, baseG, baseB = 64, 128, 64 // Green
	case 3:
		baseR, baseG, baseB = 64, 64, 128 // Blue
	default:
		baseR, baseG, baseB = 100, 100, 100 // Grey
	}

	// Apply distance fog
	fogFactor := 1.0 - (distance / 16.0)
	if fogFactor < 0.2 {
		fogFactor = 0.2
	}
	if fogFactor > 1.0 {
		fogFactor = 1.0
	}

	return color.RGBA{
		R: uint8(float64(baseR) * fogFactor),
		G: uint8(float64(baseG) * fogFactor),
		B: uint8(float64(baseB) * fogFactor),
		A: 255,
	}
}

// SetPlayerPos updates the player position.
func (r *Renderer) SetPlayerPos(x, y, angle float64) {
	r.PlayerX = x
	r.PlayerY = y
	r.PlayerA = angle
}
