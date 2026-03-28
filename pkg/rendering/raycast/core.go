// Package raycast provides the first-person raycasting renderer.
package raycast

import (
	"image/color"
	"math"
)

// Rendering constants for the raycaster.
const (
	// DefaultMapSize is the default size of the test world map.
	DefaultMapSize = 16
	// DefaultFOV is the default field of view (60 degrees in radians).
	DefaultFOV = math.Pi / 3
	// DefaultPlayerX is the default player starting X position.
	DefaultPlayerX = 8.0
	// DefaultPlayerY is the default player starting Y position.
	DefaultPlayerY = 8.0
	// MaxRaySteps is the maximum DDA steps before giving up.
	MaxRaySteps = 64
	// MaxRayDistance is returned when no wall is hit.
	MaxRayDistance = 100.0
	// FogDistance is the distance at which fog reaches maximum.
	FogDistance = 16.0
	// MinFogFactor prevents walls from becoming completely black.
	MinFogFactor = 0.2
	// MinWallDistance prevents division by zero in wall height calculation.
	MinWallDistance = 0.1
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
	worldMap := make([][]int, DefaultMapSize)
	for i := range worldMap {
		worldMap[i] = make([]int, DefaultMapSize)
		// Create boundary walls
		worldMap[i][0] = 1
		worldMap[i][DefaultMapSize-1] = 1
		if i == 0 || i == DefaultMapSize-1 {
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
		PlayerX:  DefaultPlayerX,
		PlayerY:  DefaultPlayerY,
		PlayerA:  0.0,
		WorldMap: worldMap,
		FOV:      DefaultFOV,
	}
}

// calculateDeltaDist computes the ray step lengths for DDA.
func calculateDeltaDist(rayDirX, rayDirY float64) (deltaDistX, deltaDistY float64) {
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
	return deltaDistX, deltaDistY
}

// calculateSideDist computes initial side distances and step directions for DDA.
func calculateSideDist(playerX, playerY float64, mapX, mapY int, rayDirX, rayDirY, deltaDistX, deltaDistY float64) (sideDistX, sideDistY float64, stepX, stepY int) {
	if rayDirX < 0 {
		stepX = -1
		sideDistX = (playerX - float64(mapX)) * deltaDistX
	} else {
		stepX = 1
		sideDistX = (float64(mapX) + 1.0 - playerX) * deltaDistX
	}
	if rayDirY < 0 {
		stepY = -1
		sideDistY = (playerY - float64(mapY)) * deltaDistY
	} else {
		stepY = 1
		sideDistY = (float64(mapY) + 1.0 - playerY) * deltaDistY
	}
	return sideDistX, sideDistY, stepX, stepY
}

// castRay performs DDA raycasting and returns distance and wall type.
func (r *Renderer) castRay(rayDirX, rayDirY float64) (float64, int) {
	mapX := int(r.PlayerX)
	mapY := int(r.PlayerY)

	deltaDistX, deltaDistY := calculateDeltaDist(rayDirX, rayDirY)
	sideDistX, sideDistY, stepX, stepY := calculateSideDist(
		r.PlayerX, r.PlayerY, mapX, mapY,
		rayDirX, rayDirY, deltaDistX, deltaDistY,
	)

	// DDA loop
	hit := false
	side := 0 // 0 = NS wall, 1 = EW wall

	for i := 0; i < MaxRaySteps && !hit; i++ {
		if sideDistX < sideDistY {
			sideDistX += deltaDistX
			mapX += stepX
			side = 0
		} else {
			sideDistY += deltaDistY
			mapY += stepY
			side = 1
		}

		if mapX >= 0 && mapX < len(r.WorldMap) && mapY >= 0 && mapY < len(r.WorldMap[0]) {
			if r.WorldMap[mapX][mapY] > 0 {
				hit = true
			}
		} else {
			break
		}
	}

	if !hit {
		return MaxRayDistance, 0
	}

	// Calculate perpendicular distance
	var perpWallDist float64
	if side == 0 {
		perpWallDist = sideDistX - deltaDistX
	} else {
		perpWallDist = sideDistY - deltaDistY
	}

	wallType := 0
	if mapX >= 0 && mapX < len(r.WorldMap) && mapY >= 0 && mapY < len(r.WorldMap[0]) {
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
	fogFactor := 1.0 - (distance / FogDistance)
	if fogFactor < MinFogFactor {
		fogFactor = MinFogFactor
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
