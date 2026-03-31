// Package raycast provides the first-person raycasting renderer.
package raycast

import (
	"image/color"
	"math"

	"github.com/opd-ai/wyrm/pkg/rendering/texture"
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
	Width        int
	Height       int
	PlayerX      float64
	PlayerY      float64
	PlayerA      float64 // angle in radians
	WorldMap     [][]int // 2D map: 0=empty, >0=wall type
	FOV          float64 // field of view in radians
	Genre        string  // current genre for texture palette
	WallTextures []*texture.Texture
	FloorTexture *texture.Texture
	CeilTexture  *texture.Texture
	textureSeed  int64
	// ZBuffer stores the perpendicular distance to walls for each screen column.
	// Used for sprite occlusion testing. Populated during drawWalls().
	ZBuffer []float64
}

// NewRenderer creates a new raycasting renderer.
func NewRenderer(width, height int) *Renderer {
	return NewRendererWithGenre(width, height, "fantasy", 0)
}

// NewRendererWithGenre creates a renderer with specified genre and seed.
func NewRendererWithGenre(width, height int, genre string, seed int64) *Renderer {
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

	r := &Renderer{
		Width:       width,
		Height:      height,
		PlayerX:     DefaultPlayerX,
		PlayerY:     DefaultPlayerY,
		PlayerA:     0.0,
		WorldMap:    worldMap,
		FOV:         DefaultFOV,
		Genre:       genre,
		textureSeed: seed,
		ZBuffer:     make([]float64, width),
	}
	r.initTextures()
	return r
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
// This is a simpler version used for testing; the full renderer uses castRayWithTexCoord.
func (r *Renderer) castRay(rayDirX, rayDirY float64) (float64, int) {
	mapX := int(r.PlayerX)
	mapY := int(r.PlayerY)

	deltaDistX, deltaDistY := calculateDeltaDist(rayDirX, rayDirY)
	sideDistX, sideDistY, stepX, stepY := calculateSideDist(
		r.PlayerX, r.PlayerY, mapX, mapY,
		rayDirX, rayDirY, deltaDistX, deltaDistY,
	)

	hit, side, sideDistX, sideDistY, mapX, mapY := r.performDDA(sideDistX, sideDistY, deltaDistX, deltaDistY, stepX, stepY, mapX, mapY)

	if !hit {
		return MaxRayDistance, 0
	}

	return r.calculateWallDistance(side, sideDistX, sideDistY, deltaDistX, deltaDistY, mapX, mapY)
}

// performDDA executes the DDA algorithm to find wall intersections.
// Returns: hit, side, updated sideDistX, updated sideDistY, mapX, mapY
func (r *Renderer) performDDA(sideDistX, sideDistY, deltaDistX, deltaDistY float64, stepX, stepY, mapX, mapY int) (bool, int, float64, float64, int, int) {
	side := 0
	for i := 0; i < MaxRaySteps; i++ {
		if sideDistX < sideDistY {
			sideDistX += deltaDistX
			mapX += stepX
			side = 0
		} else {
			sideDistY += deltaDistY
			mapY += stepY
			side = 1
		}

		if !r.isValidMapPosition(mapX, mapY) {
			return false, side, sideDistX, sideDistY, mapX, mapY
		}
		if r.WorldMap[mapX][mapY] > 0 {
			return true, side, sideDistX, sideDistY, mapX, mapY
		}
	}
	return false, side, sideDistX, sideDistY, mapX, mapY
}

// isValidMapPosition checks if coordinates are within map bounds.
func (r *Renderer) isValidMapPosition(x, y int) bool {
	return x >= 0 && x < len(r.WorldMap) && y >= 0 && y < len(r.WorldMap[0])
}

// calculateWallDistance computes the perpendicular wall distance.
func (r *Renderer) calculateWallDistance(side int, sideDistX, sideDistY, deltaDistX, deltaDistY float64, mapX, mapY int) (float64, int) {
	var perpWallDist float64
	if side == 0 {
		perpWallDist = sideDistX - deltaDistX
	} else {
		perpWallDist = sideDistY - deltaDistY
	}

	wallType := 0
	if r.isValidMapPosition(mapX, mapY) {
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

// SetWorldMap sets the world map from chunk heightmap data.
// The threshold determines height values that become walls.
func (r *Renderer) SetWorldMap(heightMap []float64, mapSize int, wallThreshold float64) {
	if mapSize <= 0 {
		return
	}
	worldMap := createEmptyWorldMap(mapSize)
	populateWorldMap(worldMap, heightMap, mapSize, wallThreshold)
	r.WorldMap = worldMap
}

// createEmptyWorldMap allocates a 2D map grid of the given size.
func createEmptyWorldMap(mapSize int) [][]int {
	worldMap := make([][]int, mapSize)
	for i := range worldMap {
		worldMap[i] = make([]int, mapSize)
	}
	return worldMap
}

// populateWorldMap converts heightmap values to wall types.
func populateWorldMap(worldMap [][]int, heightMap []float64, mapSize int, wallThreshold float64) {
	for y := 0; y < mapSize; y++ {
		for x := 0; x < mapSize; x++ {
			idx := y*mapSize + x
			if idx < len(heightMap) {
				worldMap[y][x] = heightToWallType(heightMap[idx], wallThreshold)
			}
		}
	}
}

// heightToWallType converts a height value to a wall type.
func heightToWallType(height, wallThreshold float64) int {
	if height <= wallThreshold {
		return 0 // No wall
	}
	if height > 0.8 {
		return 3 // High wall (blue)
	}
	if height > 0.6 {
		return 2 // Medium wall (green)
	}
	return 1 // Low wall (red-brown)
}

// SetWorldMapDirect sets the world map directly (for pre-computed maps).
func (r *Renderer) SetWorldMapDirect(worldMap [][]int) {
	r.WorldMap = worldMap
}

// TextureSize is the size of generated wall textures.
const TextureSize = 64

// initTextures generates procedural textures for walls, floor, and ceiling.
func (r *Renderer) initTextures() {
	// Generate 4 wall textures (one per wall type 1-3 plus default)
	r.WallTextures = make([]*texture.Texture, 4)
	for i := 0; i < 4; i++ {
		texSeed := r.textureSeed + int64(i)*1000
		r.WallTextures[i] = texture.GenerateWithSeed(TextureSize, TextureSize, texSeed, r.Genre)
	}

	// Generate floor and ceiling textures
	r.FloorTexture = texture.GenerateWithSeed(TextureSize, TextureSize, r.textureSeed+10000, r.Genre)
	r.CeilTexture = texture.GenerateWithSeed(TextureSize, TextureSize, r.textureSeed+20000, r.Genre)
}

// SetGenre updates the renderer genre and regenerates textures.
func (r *Renderer) SetGenre(genre string, seed int64) {
	if r.Genre == genre && r.textureSeed == seed {
		return
	}
	r.Genre = genre
	r.textureSeed = seed
	r.initTextures()
}

// GetWallTextureColor samples a color from the wall texture at the given coordinates.
func (r *Renderer) GetWallTextureColor(wallType int, texX, texY, distance float64) color.RGBA {
	if wallType < 0 || wallType >= len(r.WallTextures) || r.WallTextures[wallType] == nil {
		return r.getWallColor(wallType, distance)
	}

	tex := r.WallTextures[wallType]
	// Wrap texture coordinates
	tx := int(texX*float64(tex.Width)) % tex.Width
	ty := int(texY*float64(tex.Height)) % tex.Height
	if tx < 0 {
		tx += tex.Width
	}
	if ty < 0 {
		ty += tex.Height
	}

	idx := ty*tex.Width + tx
	if idx < 0 || idx >= len(tex.Pixels) {
		return r.getWallColor(wallType, distance)
	}

	baseColor := tex.Pixels[idx]
	return applyDistanceFog(baseColor, distance)
}

// sampleTextureColor samples a color from a texture at the given coordinates.
// Returns fallback color if texture is nil or coordinates are out of bounds.
func sampleTextureColor(tex *texture.Texture, texX, texY, distance float64, fallback color.RGBA) color.RGBA {
	if tex == nil {
		return fallback
	}

	tx := int(texX*float64(tex.Width)) % tex.Width
	ty := int(texY*float64(tex.Height)) % tex.Height
	if tx < 0 {
		tx += tex.Width
	}
	if ty < 0 {
		ty += tex.Height
	}

	idx := ty*tex.Width + tx
	if idx < 0 || idx >= len(tex.Pixels) {
		return fallback
	}

	return applyDistanceFog(tex.Pixels[idx], distance)
}

// GetFloorTextureColor samples a color from the floor texture.
func (r *Renderer) GetFloorTextureColor(texX, texY, distance float64) color.RGBA {
	return sampleTextureColor(r.FloorTexture, texX, texY, distance, color.RGBA{R: 40, G: 35, B: 45, A: 255})
}

// GetCeilingTextureColor samples a color from the ceiling texture.
func (r *Renderer) GetCeilingTextureColor(texX, texY, distance float64) color.RGBA {
	return sampleTextureColor(r.CeilTexture, texX, texY, distance, color.RGBA{R: 20, G: 12, B: 28, A: 255})
}

// applyDistanceFog applies fog based on distance to a color.
func applyDistanceFog(c color.RGBA, distance float64) color.RGBA {
	fogFactor := 1.0 - (distance / FogDistance)
	if fogFactor < MinFogFactor {
		fogFactor = MinFogFactor
	}
	if fogFactor > 1.0 {
		fogFactor = 1.0
	}

	return color.RGBA{
		R: uint8(float64(c.R) * fogFactor),
		G: uint8(float64(c.G) * fogFactor),
		B: uint8(float64(c.B) * fogFactor),
		A: 255,
	}
}

// GetZBuffer returns a copy of the current z-buffer.
// The z-buffer contains the perpendicular distance to walls for each screen column.
// Use this for sprite occlusion testing: sprites should only be drawn where
// their distance is less than the z-buffer value at that column.
func (r *Renderer) GetZBuffer() []float64 {
	if r.ZBuffer == nil {
		return nil
	}
	result := make([]float64, len(r.ZBuffer))
	copy(result, r.ZBuffer)
	return result
}

// GetZBufferAt returns the z-buffer distance at a specific screen column.
// Returns MaxRayDistance if x is out of bounds.
func (r *Renderer) GetZBufferAt(x int) float64 {
	if x < 0 || x >= len(r.ZBuffer) {
		return MaxRayDistance
	}
	return r.ZBuffer[x]
}
