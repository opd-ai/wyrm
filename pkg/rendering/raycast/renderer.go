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
	// DefaultWallHeight is the standard wall height multiplier.
	DefaultWallHeight = 1.0
	// MaxWallHeight is the maximum wall height multiplier.
	MaxWallHeight = 3.0
	// MinWallHeight is the minimum wall height multiplier.
	MinWallHeight = 0.25
	// MaxPitchAngle is the maximum vertical look angle (85 degrees in radians).
	// This prevents the player from looking straight up or down which causes rendering issues.
	MaxPitchAngle = 85.0 * math.Pi / 180.0
)

// CellFlags are bit flags for MapCell properties.
type CellFlags uint16

const (
	// FlagSolid indicates the cell blocks movement.
	FlagSolid CellFlags = 1 << iota
	// FlagPassable indicates the cell can be walked through (e.g., tall grass).
	FlagPassable
	// FlagTransparent indicates the cell renders with alpha blending.
	FlagTransparent
	// FlagClimbable indicates the cell can be climbed over.
	FlagClimbable
	// FlagDestructible indicates the cell can be destroyed.
	FlagDestructible
	// FlagSemiOpaque indicates partial opacity with gap patterns.
	FlagSemiOpaque
)

// MapCell represents a single cell in the world map with extended properties.
// This replaces the simple int wall type with rich data for variable-height
// walls, materials, and partial barriers.
type MapCell struct {
	// WallType is the wall texture index (0=empty, 1-N=wall textures).
	WallType int
	// WallHeight is the height multiplier (0.5=half, 1.0=standard, 3.0=triple).
	WallHeight float64
	// FloorH is the floor elevation (0.0=ground level).
	FloorH float64
	// CeilH is the ceiling height (defaults to WallHeight if 0).
	CeilH float64
	// MaterialID is the index into the MaterialRegistry.
	MaterialID int
	// Flags contains bit flags for cell properties.
	Flags CellFlags
}

// DefaultMapCell returns a MapCell with default values for an empty cell.
func DefaultMapCell() MapCell {
	return MapCell{
		WallType:   0,
		WallHeight: DefaultWallHeight,
		FloorH:     0.0,
		CeilH:      DefaultWallHeight,
		MaterialID: 0,
		Flags:      0,
	}
}

// WallMapCell returns a MapCell configured as a standard wall.
func WallMapCell(wallType int) MapCell {
	return MapCell{
		WallType:   wallType,
		WallHeight: DefaultWallHeight,
		FloorH:     0.0,
		CeilH:      DefaultWallHeight,
		MaterialID: wallType,
		Flags:      FlagSolid,
	}
}

// WallMapCellWithHeight returns a MapCell configured as a wall with custom height.
func WallMapCellWithHeight(wallType int, height float64) MapCell {
	if height < MinWallHeight {
		height = MinWallHeight
	}
	if height > MaxWallHeight {
		height = MaxWallHeight
	}
	return MapCell{
		WallType:   wallType,
		WallHeight: height,
		FloorH:     0.0,
		CeilH:      height,
		MaterialID: wallType,
		Flags:      FlagSolid,
	}
}

// IsEmpty returns true if the cell has no wall.
func (c MapCell) IsEmpty() bool {
	return c.WallType == 0
}

// IsSolid returns true if the cell blocks movement.
func (c MapCell) IsSolid() bool {
	return c.Flags&FlagSolid != 0
}

// IsTransparent returns true if the cell renders with alpha.
func (c MapCell) IsTransparent() bool {
	return c.Flags&FlagTransparent != 0
}

// IsClimbable returns true if the cell can be climbed over.
func (c MapCell) IsClimbable() bool {
	return c.Flags&FlagClimbable != 0
}

// IsDestructible returns true if the cell can be destroyed.
func (c MapCell) IsDestructible() bool {
	return c.Flags&FlagDestructible != 0
}

// EffectiveHeight returns the effective wall height for rendering.
// If CeilH is 0, it defaults to WallHeight.
func (c MapCell) EffectiveHeight() float64 {
	if c.CeilH > 0 {
		return c.CeilH - c.FloorH
	}
	return c.WallHeight
}

// ============================================================
// Transparency and Wall Rendering Utilities
// ============================================================

// getTransparencyForFlags returns the opacity (0.0-1.0) for the given cell flags.
func getTransparencyForFlags(flags CellFlags) float64 {
	if flags&FlagTransparent != 0 {
		return 0.5 // 50% transparent
	}
	if flags&FlagSemiOpaque != 0 {
		return 0.9 // 90% opaque (slight transparency)
	}
	return 1.0 // Fully opaque
}

// isSemiOpaqueGap determines if a texture coordinate should be transparent
// for semi-opaque barriers like fences or grates.
// Creates a regular pattern of gaps in the wall texture.
func isSemiOpaqueGap(texX, texY float64) bool {
	// Create a 4x4 grid pattern where some cells are gaps
	gridX := int(texX * 8)
	gridY := int(texY * 8)

	// Gaps every other cell in a checkerboard pattern offset
	// This creates a fence-like or grate-like appearance
	return (gridX+gridY)%4 == 0
}

// getSideDarkenFactor returns the darkening factor for a wall side.
func getSideDarkenFactor(side int) float64 {
	if side == 1 {
		return 0.8
	}
	return 1.0
}

// applySideDarkening applies a darkening factor to a color.
func applySideDarkening(c color.RGBA, factor float64) color.RGBA {
	if factor >= 1.0 {
		return c
	}
	return color.RGBA{
		R: uint8(float64(c.R) * factor),
		G: uint8(float64(c.G) * factor),
		B: uint8(float64(c.B) * factor),
		A: c.A,
	}
}

// ============================================================
// Specular Lighting System
// ============================================================

// LightSource represents a light in the scene.
type LightSource struct {
	X, Y, Z   float64    // World position
	Color     color.RGBA // Light color
	Intensity float64    // Light brightness (0.0-1.0+)
	Range     float64    // Maximum effective range
	Specular  float64    // Specular intensity multiplier
}

// LightingConfig configures the lighting calculations.
type LightingConfig struct {
	AmbientColor     color.RGBA // Base ambient light color
	AmbientIntensity float64    // Ambient light level (0.0-1.0)
	SpecularPower    float64    // Specular shininess exponent (higher = sharper highlights)
	FresnelStrength  float64    // Rim lighting intensity (0.0-1.0)
}

// DefaultLightingConfig returns sensible default lighting settings.
func DefaultLightingConfig() LightingConfig {
	return LightingConfig{
		AmbientColor:     color.RGBA{R: 30, G: 30, B: 40, A: 255},
		AmbientIntensity: 0.3,
		SpecularPower:    32.0,
		FresnelStrength:  0.2,
	}
}

// SurfaceProperties describes the material properties for lighting.
type SurfaceProperties struct {
	Roughness float64    // 0.0=mirror, 1.0=matte
	Metalness float64    // 0.0=dielectric, 1.0=metal
	Normal    [3]float64 // Surface normal (X, Y, Z)
}

// CalculateSpecular computes the specular highlight intensity.
// viewDir: direction from surface to viewer (normalized)
// lightDir: direction from surface to light (normalized)
// normal: surface normal (normalized)
// roughness: surface roughness (0.0-1.0)
// power: specular exponent
func CalculateSpecular(viewDir, lightDir, normal [3]float64, roughness, power float64) float64 {
	// Calculate half vector (Blinn-Phong)
	halfX := viewDir[0] + lightDir[0]
	halfY := viewDir[1] + lightDir[1]
	halfZ := viewDir[2] + lightDir[2]

	// Normalize half vector
	halfLen := vecLength(halfX, halfY, halfZ)
	if halfLen < 0.0001 {
		return 0.0
	}
	halfX /= halfLen
	halfY /= halfLen
	halfZ /= halfLen

	// N dot H
	nDotH := normal[0]*halfX + normal[1]*halfY + normal[2]*halfZ
	if nDotH < 0 {
		return 0.0
	}

	// Adjust power based on roughness (rougher = lower power = wider highlight)
	effectivePower := power * (1.0 - roughness*0.9)
	if effectivePower < 1 {
		effectivePower = 1
	}

	// Blinn-Phong specular
	specular := mathPow(nDotH, effectivePower)

	// Scale by inverse roughness (rougher surfaces have weaker specular)
	specular *= (1.0 - roughness*0.8)

	return specular
}

// CalculateFresnelRim computes fresnel rim lighting.
// viewDir: direction from surface to viewer (normalized)
// normal: surface normal (normalized)
// strength: rim effect intensity
func CalculateFresnelRim(viewDir, normal [3]float64, strength float64) float64 {
	// N dot V
	nDotV := normal[0]*viewDir[0] + normal[1]*viewDir[1] + normal[2]*viewDir[2]
	if nDotV < 0 {
		nDotV = 0
	}

	// Fresnel: stronger at grazing angles
	fresnel := 1.0 - nDotV
	fresnel = fresnel * fresnel * fresnel // Schlick-like approximation

	return fresnel * strength
}

// ApplySpecularToColor adds specular highlight to a surface color.
func ApplySpecularToColor(baseColor color.RGBA, specularIntensity float64, lightColor color.RGBA) color.RGBA {
	if specularIntensity <= 0 {
		return baseColor
	}

	// Specular is added on top (additive blending)
	r := float64(baseColor.R) + specularIntensity*float64(lightColor.R)
	g := float64(baseColor.G) + specularIntensity*float64(lightColor.G)
	b := float64(baseColor.B) + specularIntensity*float64(lightColor.B)

	// Clamp to 255
	return color.RGBA{
		R: clampColorFloat(r),
		G: clampColorFloat(g),
		B: clampColorFloat(b),
		A: baseColor.A,
	}
}

// clampColorFloat clamps a float to [0, 255] and converts to uint8.
func clampColorFloat(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

// vecLength calculates the length of a 3D vector.
func vecLength(x, y, z float64) float64 {
	return math.Sqrt(x*x + y*y + z*z)
}

// mathPow is a simple power function for specular exponents.
func mathPow(base, exp float64) float64 {
	return math.Pow(base, exp)
}

// CalculateDiffuse computes basic diffuse lighting (Lambert).
// lightDir: direction from surface to light (normalized)
// normal: surface normal (normalized)
func CalculateDiffuse(lightDir, normal [3]float64) float64 {
	// N dot L
	nDotL := normal[0]*lightDir[0] + normal[1]*lightDir[1] + normal[2]*lightDir[2]
	if nDotL < 0 {
		nDotL = 0
	}
	return nDotL
}

// CalculateLightAttenuation computes how light falls off with distance.
// distance: distance from surface to light
// lightRange: maximum effective range of the light
func CalculateLightAttenuation(distance, lightRange float64) float64 {
	if distance >= lightRange || lightRange <= 0 {
		return 0.0
	}
	// Inverse square falloff with range cutoff
	normalized := distance / lightRange
	// Smooth falloff curve
	attenuation := 1.0 - normalized*normalized
	if attenuation < 0 {
		attenuation = 0
	}
	return attenuation
}

// CalculateSurfaceLighting computes the full lighting for a surface point.
// Returns a multiplier to apply to the base surface color (diffuse) and specular additive.
func CalculateSurfaceLighting(
	worldPos [3]float64,
	viewerPos [3]float64,
	surface SurfaceProperties,
	light LightSource,
	config LightingConfig,
) (diffuseFactor float64, specularColor color.RGBA) {
	// Direction from surface to viewer
	viewDir := [3]float64{
		viewerPos[0] - worldPos[0],
		viewerPos[1] - worldPos[1],
		viewerPos[2] - worldPos[2],
	}
	viewLen := vecLength(viewDir[0], viewDir[1], viewDir[2])
	if viewLen > 0 {
		viewDir[0] /= viewLen
		viewDir[1] /= viewLen
		viewDir[2] /= viewLen
	}

	// Direction from surface to light
	lightDir := [3]float64{
		light.X - worldPos[0],
		light.Y - worldPos[1],
		light.Z - worldPos[2],
	}
	lightDist := vecLength(lightDir[0], lightDir[1], lightDir[2])
	if lightDist > 0 {
		lightDir[0] /= lightDist
		lightDir[1] /= lightDist
		lightDir[2] /= lightDist
	}

	// Light attenuation
	attenuation := CalculateLightAttenuation(lightDist, light.Range) * light.Intensity
	if attenuation <= 0 {
		return config.AmbientIntensity, color.RGBA{A: 255}
	}

	// Diffuse lighting
	diffuse := CalculateDiffuse(lightDir, surface.Normal)

	// Specular lighting
	specular := CalculateSpecular(viewDir, lightDir, surface.Normal, surface.Roughness, config.SpecularPower)
	specular *= light.Specular * attenuation

	// Apply metalness - metals tint specular with their base color
	specR := float64(light.Color.R)*(1.0-surface.Metalness) + float64(light.Color.R)*surface.Metalness
	specG := float64(light.Color.G)*(1.0-surface.Metalness) + float64(light.Color.G)*surface.Metalness
	specB := float64(light.Color.B)*(1.0-surface.Metalness) + float64(light.Color.B)*surface.Metalness

	specularColor = color.RGBA{
		R: clampColorFloat(specR * specular),
		G: clampColorFloat(specG * specular),
		B: clampColorFloat(specB * specular),
		A: 255,
	}

	// Fresnel rim
	rim := CalculateFresnelRim(viewDir, surface.Normal, config.FresnelStrength)
	specularColor.R = clampColorFloat(float64(specularColor.R) + rim*50)
	specularColor.G = clampColorFloat(float64(specularColor.G) + rim*50)
	specularColor.B = clampColorFloat(float64(specularColor.B) + rim*50)

	// Total diffuse factor
	diffuseFactor = config.AmbientIntensity + diffuse*attenuation
	if diffuseFactor > 1.0 {
		diffuseFactor = 1.0
	}

	return diffuseFactor, specularColor
}

// ApplyLighting applies lighting calculations to a base color.
func ApplyLighting(baseColor color.RGBA, diffuseFactor float64, specularColor color.RGBA) color.RGBA {
	// Apply diffuse factor to base color
	r := float64(baseColor.R)*diffuseFactor + float64(specularColor.R)
	g := float64(baseColor.G)*diffuseFactor + float64(specularColor.G)
	b := float64(baseColor.B)*diffuseFactor + float64(specularColor.B)

	return color.RGBA{
		R: clampColorFloat(r),
		G: clampColorFloat(g),
		B: clampColorFloat(b),
		A: baseColor.A,
	}
}

// Renderer handles first-person raycasting and draws to an Ebitengine image.
type Renderer struct {
	Width       int
	Height      int
	PlayerX     float64
	PlayerY     float64
	PlayerA     float64 // angle in radians
	PlayerZ     float64 // player eye height (default 0.5 = standing)
	PlayerPitch float64 // vertical look angle in radians (clamped ±85°)
	WorldMap    [][]int // 2D map: 0=empty, >0=wall type (legacy)
	// WorldMapCells is the enhanced map with variable heights and materials.
	// When set, this takes precedence over WorldMap for rendering.
	WorldMapCells [][]MapCell
	FOV           float64 // field of view in radians
	Genre         string  // current genre for texture palette
	WallTextures  []*texture.Texture
	FloorTexture  *texture.Texture
	CeilTexture   *texture.Texture
	textureSeed   int64
	// ZBuffer stores the perpendicular distance to walls for each screen column.
	// Used for sprite occlusion testing. Populated during drawWalls().
	ZBuffer []float64
	// Framebuffer stores RGBA pixel data for batch rendering.
	// Allocated once in constructor, reused each frame.
	// Format: width * height * 4 bytes (RGBA)
	Framebuffer []byte
	// visibleSprites is a pre-allocated slice for sorting visible sprites.
	// Reused each frame with [:0] to avoid allocations.
	visibleSprites []*SpriteEntity
	// Skybox handles procedural sky rendering with time-of-day and weather effects.
	// Initialized automatically in NewRenderer. Use SetSkybox to customize.
	Skybox *Skybox
	// HighlightConfig holds the configuration for interaction highlight effects.
	// If nil, DefaultHighlightConfig() is used.
	HighlightConfig *HighlightConfig
	// highlightPulsePhase is the current phase of the highlight pulse animation.
	// Updated via UpdateHighlightPulse() each frame.
	highlightPulsePhase float64
	// contextPool is a pool of reusable SpriteDrawContext objects to reduce allocations.
	// Pre-allocated in NewRenderer, reused in PrepareSpriteDrawContext.
	contextPool []*SpriteDrawContext
	// contextPoolIdx is the current index in the context pool.
	// Reset to 0 at the start of each frame via ResetContextPool().
	contextPoolIdx int
	// LODConfig holds the Level of Detail configuration for performance optimization.
	// If nil, DefaultLODConfig() is used.
	LODConfig *LODConfig
	// QualityConfig holds the rendering quality configuration.
	// If nil, DefaultQualityConfig() is used.
	QualityConfig *QualityConfig
	// accessibilityConfig holds accessibility settings for rendering.
	// If nil, DefaultAccessibilityConfig() is used.
	accessibilityConfig *AccessibilityConfig
	// NormalLighting handles normal map sampling and lighting calculations.
	// If nil, normal map lighting is disabled.
	NormalLighting *NormalLighting
}

// NewRenderer creates a new raycasting renderer.
func NewRenderer(width, height int) *Renderer {
	return NewRendererWithGenre(width, height, "fantasy", 0)
}

// NewRendererWithGenre creates a renderer with specified genre and seed.
func NewRendererWithGenre(width, height int, genre string, seed int64) *Renderer {
	// Validate dimensions to prevent panics from empty framebuffer access
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}

	// Create a simple default world map (legacy format for backward compatibility)
	worldMap := make([][]int, DefaultMapSize)
	worldMapCells := make([][]MapCell, DefaultMapSize)
	for i := range worldMap {
		worldMap[i] = make([]int, DefaultMapSize)
		worldMapCells[i] = make([]MapCell, DefaultMapSize)
		// Create boundary walls
		worldMap[i][0] = 1
		worldMapCells[i][0] = WallMapCell(1)
		worldMap[i][DefaultMapSize-1] = 1
		worldMapCells[i][DefaultMapSize-1] = WallMapCell(1)
		if i == 0 || i == DefaultMapSize-1 {
			for j := range worldMap[i] {
				worldMap[i][j] = 1
				worldMapCells[i][j] = WallMapCell(1)
			}
		}
	}
	// Add some interior walls with varying heights
	worldMap[4][4] = 2
	worldMapCells[4][4] = WallMapCell(2)
	worldMap[4][5] = 2
	worldMapCells[4][5] = WallMapCellWithHeight(2, 1.5) // Taller wall
	worldMap[4][6] = 2
	worldMapCells[4][6] = WallMapCell(2)
	worldMap[8][8] = 3
	worldMapCells[8][8] = WallMapCellWithHeight(3, 2.0) // Double height
	worldMap[8][9] = 3
	worldMapCells[8][9] = WallMapCell(3)
	worldMap[9][8] = 3
	worldMapCells[9][8] = WallMapCellWithHeight(3, 0.5) // Half height

	r := &Renderer{
		Width:          width,
		Height:         height,
		PlayerX:        DefaultPlayerX,
		PlayerY:        DefaultPlayerY,
		PlayerA:        0.0,
		PlayerZ:        0.5, // Default standing eye height
		PlayerPitch:    0.0,
		WorldMap:       worldMap,
		WorldMapCells:  worldMapCells,
		FOV:            DefaultFOV,
		Genre:          genre,
		textureSeed:    seed,
		ZBuffer:        make([]float64, width),
		Framebuffer:    make([]byte, width*height*4),
		visibleSprites: make([]*SpriteEntity, 0, 256),     // Pre-allocate for typical entity count
		contextPool:    make([]*SpriteDrawContext, 0, 64), // Pre-allocate context pool
	}
	// Initialize context pool with reusable objects
	for i := 0; i < 64; i++ {
		r.contextPool = append(r.contextPool, &SpriteDrawContext{})
	}
	r.initTextures()
	r.initSkybox(genre)
	r.NormalLighting = DefaultNormalLighting()
	return r
}

// initSkybox initializes the skybox renderer with genre-appropriate settings.
func (r *Renderer) initSkybox(genre string) {
	r.Skybox = NewSkybox()
	r.Skybox.SetGenre(genre)
}

// SetSkybox sets a custom skybox renderer.
func (r *Renderer) SetSkybox(skybox *Skybox) {
	r.Skybox = skybox
}

// GetSkybox returns the current skybox renderer.
func (r *Renderer) GetSkybox() *Skybox {
	return r.Skybox
}

// getHorizonLine returns the Y coordinate of the horizon line.
// Affected by PlayerPitch: looking up moves horizon down, looking down moves it up.
func (r *Renderer) getHorizonLine() int {
	halfHeight := r.Height / 2

	// Calculate pitch offset: positive pitch (look up) moves horizon down
	// MaxPitchOffset is half the screen height (can look at pure sky or pure floor)
	maxPitchOffset := float64(halfHeight)
	pitchOffset := (r.PlayerPitch / MaxPitchAngle) * maxPitchOffset

	horizonY := halfHeight + int(pitchOffset)

	// Clamp to screen bounds
	if horizonY < 0 {
		horizonY = 0
	}
	if horizonY > r.Height {
		horizonY = r.Height
	}

	return horizonY
}

// SetPixel writes a pixel to the framebuffer at the specified coordinates.
// Bounds checking is performed to prevent out-of-bounds writes.
func (r *Renderer) SetPixel(x, y int, red, green, blue, alpha uint8) {
	if x < 0 || x >= r.Width || y < 0 || y >= r.Height {
		return
	}
	idx := (y*r.Width + x) * 4
	r.Framebuffer[idx] = red
	r.Framebuffer[idx+1] = green
	r.Framebuffer[idx+2] = blue
	r.Framebuffer[idx+3] = alpha
}

// SetPixelColor writes a color.RGBA to the framebuffer at the specified coordinates.
func (r *Renderer) SetPixelColor(x, y int, c color.RGBA) {
	r.SetPixel(x, y, c.R, c.G, c.B, c.A)
}

// BlendPixel alpha-blends a pixel onto the existing framebuffer content.
func (r *Renderer) BlendPixel(x, y int, red, green, blue, alpha uint8) {
	if x < 0 || x >= r.Width || y < 0 || y >= r.Height {
		return
	}
	if alpha == 255 {
		r.SetPixel(x, y, red, green, blue, 255)
		return
	}
	if alpha == 0 {
		return
	}
	idx := (y*r.Width + x) * 4
	a := float64(alpha) / 255.0
	invA := 1.0 - a
	r.Framebuffer[idx] = uint8(float64(red)*a + float64(r.Framebuffer[idx])*invA)
	r.Framebuffer[idx+1] = uint8(float64(green)*a + float64(r.Framebuffer[idx+1])*invA)
	r.Framebuffer[idx+2] = uint8(float64(blue)*a + float64(r.Framebuffer[idx+2])*invA)
	r.Framebuffer[idx+3] = 255
}

// BlendPixelColor alpha-blends a color.RGBA onto the existing framebuffer content.
func (r *Renderer) BlendPixelColor(x, y int, c color.RGBA) {
	r.BlendPixel(x, y, c.R, c.G, c.B, c.A)
}

// ClearFramebuffer zeros the framebuffer to prepare for a new frame.
func (r *Renderer) ClearFramebuffer() {
	for i := range r.Framebuffer {
		r.Framebuffer[i] = 0
	}
}

// GetPixel reads a pixel from the framebuffer at the specified coordinates.
// Returns (0,0,0,0) if coordinates are out of bounds.
func (r *Renderer) GetPixel(x, y int) (red, green, blue, alpha uint8) {
	if x < 0 || x >= r.Width || y < 0 || y >= r.Height {
		return 0, 0, 0, 0
	}
	idx := (y*r.Width + x) * 4
	return r.Framebuffer[idx], r.Framebuffer[idx+1], r.Framebuffer[idx+2], r.Framebuffer[idx+3]
}

// GetFramebuffer returns the raw framebuffer slice for direct access.
// Use this for WritePixels() upload to ebiten.Image.
func (r *Renderer) GetFramebuffer() []byte {
	return r.Framebuffer
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

		if !r.isValidMapPosition(mapX, mapY) && !r.isValidMapCellPosition(mapX, mapY) {
			return false, side, sideDistX, sideDistY, mapX, mapY
		}
		if r.HasWall(mapX, mapY) {
			return true, side, sideDistX, sideDistY, mapX, mapY
		}
	}
	return false, side, sideDistX, sideDistY, mapX, mapY
}

// isValidMapPosition checks if coordinates are within map bounds.
func (r *Renderer) isValidMapPosition(x, y int) bool {
	if x < 0 || x >= len(r.WorldMap) {
		return false
	}
	return y >= 0 && y < len(r.WorldMap[x])
}

// isValidMapCellPosition checks if coordinates are within WorldMapCells bounds.
func (r *Renderer) isValidMapCellPosition(x, y int) bool {
	if r.WorldMapCells == nil {
		return false
	}
	if x < 0 || x >= len(r.WorldMapCells) {
		return false
	}
	return y >= 0 && y < len(r.WorldMapCells[x])
}

// GetMapCell returns the MapCell at the given position.
// If WorldMapCells is set, it returns from that; otherwise, it converts
// from the legacy WorldMap format.
func (r *Renderer) GetMapCell(x, y int) MapCell {
	// Try enhanced map first
	if r.isValidMapCellPosition(x, y) {
		return r.WorldMapCells[x][y]
	}
	// Fallback to legacy map conversion
	if r.isValidMapPosition(x, y) {
		wallType := r.WorldMap[x][y]
		if wallType == 0 {
			return DefaultMapCell()
		}
		return WallMapCell(wallType)
	}
	// Out of bounds
	return DefaultMapCell()
}

// HasWall checks if there is a wall at the given position using either map format.
func (r *Renderer) HasWall(x, y int) bool {
	if r.isValidMapCellPosition(x, y) {
		return r.WorldMapCells[x][y].WallType > 0
	}
	if r.isValidMapPosition(x, y) {
		return r.WorldMap[x][y] > 0
	}
	return false
}

// calculateWallDistance computes the perpendicular wall distance.
func (r *Renderer) calculateWallDistance(side int, sideDistX, sideDistY, deltaDistX, deltaDistY float64, mapX, mapY int) (float64, int) {
	var perpWallDist float64
	if side == 0 {
		perpWallDist = sideDistX - deltaDistX
	} else {
		perpWallDist = sideDistY - deltaDistY
	}

	cell := r.GetMapCell(mapX, mapY)
	return perpWallDist, cell.WallType
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
// Also populates WorldMapCells for the enhanced rendering pipeline.
func (r *Renderer) SetWorldMap(heightMap []float64, mapSize int, wallThreshold float64) {
	if mapSize <= 0 {
		return
	}
	worldMap := createEmptyWorldMap(mapSize)
	worldMapCells := createEmptyWorldMapCells(mapSize)
	populateWorldMap(worldMap, heightMap, mapSize, wallThreshold)
	populateWorldMapCells(worldMapCells, heightMap, mapSize, wallThreshold)
	r.WorldMap = worldMap
	r.WorldMapCells = worldMapCells
}

// createEmptyWorldMap allocates a 2D map grid of the given size.
func createEmptyWorldMap(mapSize int) [][]int {
	worldMap := make([][]int, mapSize)
	for i := range worldMap {
		worldMap[i] = make([]int, mapSize)
	}
	return worldMap
}

// createEmptyWorldMapCells allocates a 2D MapCell grid of the given size.
func createEmptyWorldMapCells(mapSize int) [][]MapCell {
	worldMapCells := make([][]MapCell, mapSize)
	for i := range worldMapCells {
		worldMapCells[i] = make([]MapCell, mapSize)
		for j := range worldMapCells[i] {
			worldMapCells[i][j] = DefaultMapCell()
		}
	}
	return worldMapCells
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

// populateWorldMapCells converts heightmap values to MapCells with height data.
func populateWorldMapCells(worldMapCells [][]MapCell, heightMap []float64, mapSize int, wallThreshold float64) {
	for y := 0; y < mapSize; y++ {
		for x := 0; x < mapSize; x++ {
			idx := y*mapSize + x
			if idx < len(heightMap) {
				worldMapCells[y][x] = heightToMapCell(heightMap[idx], wallThreshold)
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

// heightToMapCell converts a height value to a MapCell with appropriate properties.
func heightToMapCell(height, wallThreshold float64) MapCell {
	if height <= wallThreshold {
		return DefaultMapCell()
	}
	wallType := heightToWallType(height, wallThreshold)
	// Calculate wall height based on terrain height - higher terrain = taller walls
	// Scale from 0.5 (just above threshold) to 2.0 (maximum height)
	normalizedHeight := (height - wallThreshold) / (1.0 - wallThreshold)
	wallHeight := MinWallHeight + normalizedHeight*(2.0-MinWallHeight)
	if wallHeight > MaxWallHeight {
		wallHeight = MaxWallHeight
	}
	return WallMapCellWithHeight(wallType, wallHeight)
}

// SetWorldMapDirect sets the world map directly (for pre-computed maps).
// Also creates corresponding WorldMapCells with default heights.
func (r *Renderer) SetWorldMapDirect(worldMap [][]int) {
	r.WorldMap = worldMap
	// Create corresponding WorldMapCells
	if len(worldMap) > 0 && len(worldMap[0]) > 0 {
		r.WorldMapCells = convertWorldMapToMapCells(worldMap)
	}
}

// convertWorldMapToMapCells converts a legacy WorldMap to WorldMapCells.
func convertWorldMapToMapCells(worldMap [][]int) [][]MapCell {
	mapCells := make([][]MapCell, len(worldMap))
	for x := range worldMap {
		mapCells[x] = make([]MapCell, len(worldMap[x]))
		for y := range worldMap[x] {
			if worldMap[x][y] == 0 {
				mapCells[x][y] = DefaultMapCell()
			} else {
				mapCells[x][y] = WallMapCell(worldMap[x][y])
			}
		}
	}
	return mapCells
}

// SetWorldMapCells sets the enhanced world map directly.
// Also updates the legacy WorldMap for backward compatibility.
func (r *Renderer) SetWorldMapCells(worldMapCells [][]MapCell) {
	r.WorldMapCells = worldMapCells
	// Create corresponding legacy WorldMap
	if len(worldMapCells) > 0 && len(worldMapCells[0]) > 0 {
		r.WorldMap = convertMapCellsToWorldMap(worldMapCells)
	}
}

// SetWorldMapWithWallHeights sets the world map from chunk data including wall heights.
// This method allows using pre-computed wall heights from chunk generation.
func (r *Renderer) SetWorldMapWithWallHeights(heightMap, wallHeights []float64, mapSize int, wallThreshold float64) {
	if mapSize <= 0 {
		return
	}
	worldMap := createEmptyWorldMap(mapSize)
	worldMapCells := createEmptyWorldMapCells(mapSize)
	populateWorldMap(worldMap, heightMap, mapSize, wallThreshold)
	populateWorldMapCellsWithWallHeights(worldMapCells, heightMap, wallHeights, mapSize, wallThreshold)
	r.WorldMap = worldMap
	r.WorldMapCells = worldMapCells
}

// populateWorldMapCellsWithWallHeights creates MapCells using pre-computed wall heights.
func populateWorldMapCellsWithWallHeights(worldMapCells [][]MapCell, heightMap, wallHeights []float64, mapSize int, wallThreshold float64) {
	for y := 0; y < mapSize; y++ {
		for x := 0; x < mapSize; x++ {
			idx := y*mapSize + x
			height := 0.0
			if idx < len(heightMap) {
				height = heightMap[idx]
			}
			wallHeight := DefaultWallHeight
			if wallHeights != nil && idx < len(wallHeights) {
				wallHeight = wallHeights[idx]
			}
			worldMapCells[y][x] = heightToMapCellWithWallHeight(height, wallThreshold, wallHeight)
		}
	}
}

// heightToMapCellWithWallHeight creates a MapCell using a pre-computed wall height.
func heightToMapCellWithWallHeight(height, wallThreshold, wallHeight float64) MapCell {
	if height <= wallThreshold {
		return DefaultMapCell()
	}
	wallType := heightToWallType(height, wallThreshold)
	// Clamp wall height to valid range
	if wallHeight < MinWallHeight {
		wallHeight = MinWallHeight
	}
	if wallHeight > MaxWallHeight {
		wallHeight = MaxWallHeight
	}
	return MapCell{
		WallType:   wallType,
		WallHeight: wallHeight,
		FloorH:     0.0,
		CeilH:      wallHeight,
		MaterialID: wallType,
		Flags:      FlagSolid,
	}
}

// convertMapCellsToWorldMap converts WorldMapCells to a legacy WorldMap.
func convertMapCellsToWorldMap(mapCells [][]MapCell) [][]int {
	worldMap := make([][]int, len(mapCells))
	for x := range mapCells {
		worldMap[x] = make([]int, len(mapCells[x]))
		for y := range mapCells[x] {
			worldMap[x][y] = mapCells[x][y].WallType
		}
	}
	return worldMap
}

// GetPlayerAngle returns the player's current facing angle in radians.
func (r *Renderer) GetPlayerAngle() float64 {
	return r.PlayerA
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

// UpdateHighlightPulse advances the highlight pulse animation.
// dt is the time delta in seconds since the last frame.
// Call this once per frame to animate highlight effects.
func (r *Renderer) UpdateHighlightPulse(dt float64) {
	cfg := r.HighlightConfig
	if cfg == nil {
		cfg = DefaultHighlightConfig()
	}
	if cfg.PulseEnabled {
		r.highlightPulsePhase += cfg.PulseSpeed * dt
		// Wrap to prevent float overflow on long sessions
		if r.highlightPulsePhase > 2*math.Pi*1000 {
			r.highlightPulsePhase -= 2 * math.Pi * 1000
		}
	}
}

// SetHighlightConfig sets the highlight effect configuration.
func (r *Renderer) SetHighlightConfig(cfg *HighlightConfig) {
	r.HighlightConfig = cfg
}

// GetHighlightPulsePhase returns the current highlight pulse phase.
// Useful for synchronizing other UI effects with the highlight pulse.
func (r *Renderer) GetHighlightPulsePhase() float64 {
	return r.highlightPulsePhase
}

// ResetContextPool resets the context pool index to 0.
// Call this at the start of each frame before rendering sprites.
func (r *Renderer) ResetContextPool() {
	r.contextPoolIdx = 0
}

// getPooledContext returns a reusable SpriteDrawContext from the pool.
// If the pool is exhausted, allocates a new context and adds it to the pool.
func (r *Renderer) getPooledContext() *SpriteDrawContext {
	if r.contextPoolIdx >= len(r.contextPool) {
		// Pool exhausted, grow it
		ctx := &SpriteDrawContext{}
		r.contextPool = append(r.contextPool, ctx)
	}
	ctx := r.contextPool[r.contextPoolIdx]
	r.contextPoolIdx++
	return ctx
}

// GetLODConfig returns the LOD configuration, using the default if not set.
func (r *Renderer) GetLODConfig() *LODConfig {
	if r.LODConfig == nil {
		return defaultLODConfig
	}
	return r.LODConfig
}

// SetLODConfig sets the LOD configuration.
func (r *Renderer) SetLODConfig(cfg *LODConfig) {
	r.LODConfig = cfg
}

// GetLODLevel returns the LOD level for a given distance.
func (r *Renderer) GetLODLevel(distance float64) LODLevel {
	return r.GetLODConfig().GetLODLevel(distance)
}

// GetLODRenderFlags returns the render flags for a given distance.
func (r *Renderer) GetLODRenderFlags(distance float64) LODRenderFlags {
	return r.GetLODConfig().GetRenderFlags(distance)
}

// GetQualityConfig returns the quality configuration, using the default if not set.
func (r *Renderer) GetQualityConfig() *QualityConfig {
	if r.QualityConfig == nil {
		return DefaultQualityConfig()
	}
	return r.QualityConfig
}

// SetQualityConfig sets the rendering quality configuration.
func (r *Renderer) SetQualityConfig(cfg *QualityConfig) {
	r.QualityConfig = cfg
	// Also update LOD config based on quality
	if cfg != nil && r.LODConfig != nil {
		cfg.ApplyToLODConfig(r.LODConfig)
	}
}

// SetQualityPreset sets the rendering quality from a preset.
func (r *Renderer) SetQualityPreset(preset QualityPreset) {
	cfg := NewQualityConfig(preset)
	r.SetQualityConfig(cfg)
}
