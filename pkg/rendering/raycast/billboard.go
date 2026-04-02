// Package raycast provides the first-person raycasting renderer.
// This file implements billboard sprite rendering for entities.
package raycast

import (
	"math"
	"sort"

	"github.com/opd-ai/wyrm/pkg/rendering/sprite"
)

// Highlight state constants
const (
	// HighlightNone indicates no highlight effect
	HighlightNone = 0
	// HighlightInRange indicates object is within interaction range
	HighlightInRange = 1
	// HighlightTargeted indicates object is directly targeted by crosshair
	HighlightTargeted = 2
)

// HighlightConfig holds configuration for highlight effects.
type HighlightConfig struct {
	// InRangeColor is the highlight color for objects in interaction range
	InRangeColor sprite.PixelRGBA
	// TargetedColor is the highlight color for directly targeted objects
	TargetedColor sprite.PixelRGBA
	// OutlineWidth is the width of the highlight outline in pixels (screen space)
	OutlineWidth int
	// GlowIntensity controls the brightness of the glow effect (0.0-1.0)
	GlowIntensity float64
	// PulseEnabled enables pulsing animation for highlights
	PulseEnabled bool
	// PulseSpeed controls how fast the highlight pulses (radians per second)
	PulseSpeed float64
}

// DefaultHighlightConfig returns the default highlight configuration.
// defaultHighlightConfig is a package-level default to avoid allocations.
// Use DefaultHighlightConfig() to get a copy if modifications are needed.
var defaultHighlightConfig = &HighlightConfig{
	InRangeColor:  sprite.PixelRGBA{R: 255, G: 255, B: 200, A: 128}, // Soft yellow
	TargetedColor: sprite.PixelRGBA{R: 255, G: 255, B: 100, A: 200}, // Bright yellow
	OutlineWidth:  2,
	GlowIntensity: 0.5,
	PulseEnabled:  true,
	PulseSpeed:    4.0,
}

func DefaultHighlightConfig() *HighlightConfig {
	return &HighlightConfig{
		InRangeColor:  sprite.PixelRGBA{R: 255, G: 255, B: 200, A: 128}, // Soft yellow
		TargetedColor: sprite.PixelRGBA{R: 255, G: 255, B: 100, A: 200}, // Bright yellow
		OutlineWidth:  2,
		GlowIntensity: 0.5,
		PulseEnabled:  true,
		PulseSpeed:    4.0,
	}
}

// HighlightColorForState returns the appropriate highlight color for a state.
func (cfg *HighlightConfig) HighlightColorForState(state int) sprite.PixelRGBA {
	switch state {
	case HighlightInRange:
		return cfg.InRangeColor
	case HighlightTargeted:
		return cfg.TargetedColor
	default:
		return sprite.PixelRGBA{}
	}
}

// ApplyHighlightToPixel applies a highlight effect to a pixel.
// state: 0=none, 1=in range (subtle), 2=targeted (strong)
// isEdge: true if this pixel is on the edge of the sprite (for outline effect)
// pulsePhase: current phase of pulse animation (0.0-2*PI)
func ApplyHighlightToPixel(pixel sprite.PixelRGBA, state int, isEdge bool, pulsePhase float64, cfg *HighlightConfig) sprite.PixelRGBA {
	if state == HighlightNone || pixel.A == 0 {
		return pixel
	}

	if cfg == nil {
		cfg = defaultHighlightConfig
	}

	highlightColor := cfg.HighlightColorForState(state)

	// Calculate pulse intensity if enabled
	intensity := cfg.GlowIntensity
	if cfg.PulseEnabled {
		// Sinusoidal pulse between 0.5 and 1.0 of intensity
		pulseFactor := 0.75 + 0.25*math.Sin(pulsePhase)
		intensity *= pulseFactor
	}

	// Stronger effect for edges (outline effect)
	if isEdge {
		intensity = math.Min(1.0, intensity*1.5)
	}

	// Blend highlight color with original pixel
	blendFactor := float64(highlightColor.A) / 255.0 * intensity
	invBlend := 1.0 - blendFactor

	result := sprite.PixelRGBA{
		R: uint8(float64(pixel.R)*invBlend + float64(highlightColor.R)*blendFactor),
		G: uint8(float64(pixel.G)*invBlend + float64(highlightColor.G)*blendFactor),
		B: uint8(float64(pixel.B)*invBlend + float64(highlightColor.B)*blendFactor),
		A: pixel.A,
	}

	return result
}

// IsPixelOnSpriteEdge determines if a pixel is on the edge of the sprite.
// Used for outline highlight effects.
func IsPixelOnSpriteEdge(frame *sprite.Sprite, texX, texY int, flipH bool) bool {
	if frame == nil {
		return false
	}

	// Check if any adjacent pixel is transparent
	adjacentCoords := [][2]int{
		{texX - 1, texY},
		{texX + 1, texY},
		{texX, texY - 1},
		{texX, texY + 1},
	}

	for _, coord := range adjacentCoords {
		x, y := coord[0], coord[1]

		// Handle horizontal flip
		checkX := x
		if flipH && x >= 0 && x < frame.Width {
			checkX = frame.Width - 1 - x
		}

		// Out of bounds = edge
		if x < 0 || x >= frame.Width || y < 0 || y >= frame.Height {
			return true
		}

		// Adjacent transparent pixel = edge
		adjPixel := frame.GetPixel(checkX, y)
		if adjPixel.A == 0 {
			return true
		}
	}

	return false
}

// SpriteEntity represents an entity to be rendered as a billboard sprite.
// Used to pass entity data from the ECS to the renderer.
type SpriteEntity struct {
	// World position
	X, Y float64
	// Distance from camera (computed during transform)
	Distance float64
	// Transform coordinates in camera space
	TransformX, TransformY float64
	// Screen X position (computed during transform)
	ScreenX int
	// The sprite sheet to render
	Sheet *sprite.SpriteSheet
	// Current animation state
	AnimState string
	// Current animation frame
	AnimFrame int
	// Scale multiplier (1.0 = default size)
	Scale float64
	// Opacity for alpha blending (0.0-1.0)
	Opacity float64
	// Whether to flip horizontally
	FlipH bool
	// Visible flag
	Visible bool

	// Interaction metadata (Phase 5 additions)
	// InteractionType describes what interaction is available (pickup, open, use, etc.)
	InteractionType string
	// InteractionRange is the maximum distance for interaction (in world units)
	InteractionRange float64
	// HighlightState indicates the visual highlight level
	// 0 = no highlight, 1 = in range, 2 = targeted
	HighlightState int
	// IsInteractable indicates if this entity can be interacted with
	IsInteractable bool
	// UseText is the action text shown to the player
	UseText string
	// DisplayName is the entity's name shown on hover/target
	DisplayName string
	// EntityID is the ECS entity ID for interaction callbacks
	EntityID uint64

	// Scale-correct rendering fields (Phase 5)
	// WorldHeight is the entity's height in world units (e.g., 0.3 for potion, 1.8 for human)
	// When non-zero, Scale is computed from this value.
	WorldHeight float64
	// VerticalOffset moves the sprite up/down from ground level in world units
	// Positive values raise the sprite, negative lower it
	// Default 0 means sprite center is at eye level
	VerticalOffset float64
	// GroundLevel indicates if sprite should rest on ground (bottom-aligned)
	// rather than being centered at eye level
	GroundLevel bool
}

// ItemSizeCategory represents standard item size categories for scale-correct rendering.
type ItemSizeCategory int

const (
	// SizeTiny for very small items (coins, rings, keys)
	SizeTiny ItemSizeCategory = iota
	// SizeSmall for small items (potions, scrolls, food)
	SizeSmall
	// SizeMedium for medium items (weapons, tools, books)
	SizeMedium
	// SizeLarge for large items (shields, staves, furniture)
	SizeLarge
	// SizeHuge for huge items (statues, large furniture, vehicles)
	SizeHuge
	// SizeCharacter for character-sized entities (NPCs, creatures)
	SizeCharacter
)

// ItemSizeWorldHeight maps item size categories to world height in units.
var ItemSizeWorldHeight = map[ItemSizeCategory]float64{
	SizeTiny:      0.1,  // ~10cm
	SizeSmall:     0.25, // ~25cm
	SizeMedium:    0.5,  // ~50cm
	SizeLarge:     1.0,  // ~1m
	SizeHuge:      2.0,  // ~2m
	SizeCharacter: 1.8,  // average human height
}

// GetWorldHeightForSize returns the world height for a given size category.
func GetWorldHeightForSize(size ItemSizeCategory) float64 {
	if h, ok := ItemSizeWorldHeight[size]; ok {
		return h
	}
	return ItemSizeWorldHeight[SizeMedium]
}

// ComputeScaleFromWorldHeight calculates the Scale value based on world height.
// baseHeight is the reference height (typically 1.0 for a unit-height sprite).
func ComputeScaleFromWorldHeight(worldHeight, baseHeight float64) float64 {
	if baseHeight <= 0 {
		baseHeight = 1.0
	}
	if worldHeight <= 0 {
		return 1.0
	}
	return worldHeight / baseHeight
}

// SetEntitySize configures an entity for scale-correct rendering based on size category.
// Sets WorldHeight, VerticalOffset, and GroundLevel appropriately.
func (e *SpriteEntity) SetEntitySize(size ItemSizeCategory, grounded bool) {
	e.WorldHeight = GetWorldHeightForSize(size)
	e.GroundLevel = grounded
	if grounded {
		// When grounded, offset so sprite bottom is at floor level
		// Floor is at eye level - 0.5 (half unit below eye level)
		// Sprite center needs to be at floor + worldHeight/2
		e.VerticalOffset = -0.5 + e.WorldHeight/2
	} else {
		e.VerticalOffset = 0
	}
}

// SetCustomWorldHeight sets a custom world height for the entity.
func (e *SpriteEntity) SetCustomWorldHeight(height float64, grounded bool) {
	e.WorldHeight = height
	e.GroundLevel = grounded
	if grounded {
		e.VerticalOffset = -0.5 + height/2
	} else {
		e.VerticalOffset = 0
	}
}

// TransformEntityToScreen computes the screen position and distance for an entity.
// Returns false if the entity is behind the camera or outside the frustum.
func (r *Renderer) TransformEntityToScreen(entity *SpriteEntity) bool {
	if !entity.Visible {
		return false
	}

	// Translate to camera-relative coordinates
	dx := entity.X - r.PlayerX
	dy := entity.Y - r.PlayerY

	// Camera direction vector
	dirX := math.Cos(r.PlayerA)
	dirY := math.Sin(r.PlayerA)

	// Camera plane (perpendicular to direction, scaled by FOV)
	planeMult := math.Tan(r.FOV / 2)
	planeX := -dirY * planeMult
	planeY := dirX * planeMult

	// Transform into camera space using inverse camera matrix
	det := planeX*dirY - dirX*planeY
	// Guard against division by zero (occurs with zero FOV or zero-length direction)
	if math.Abs(det) < 1e-10 {
		return false
	}
	invDet := 1.0 / det
	entity.TransformX = invDet * (dirY*dx - dirX*dy)
	entity.TransformY = invDet * (-planeY*dx + planeX*dy)

	// Skip if behind camera
	if entity.TransformY <= 0 {
		return false
	}

	// Store distance for sorting and fog
	entity.Distance = entity.TransformY

	// Calculate screen X position
	screenW := float64(r.Width)
	entity.ScreenX = int((screenW / 2) * (1 + entity.TransformX/entity.TransformY))

	return true
}

// GetSpriteScreenBounds calculates the screen bounds for a sprite.
// Returns startX, endX, startY, endY, spriteWidth, spriteHeight.
func (r *Renderer) GetSpriteScreenBounds(entity *SpriteEntity, spriteW, spriteH int) (int, int, int, int, int, int) {
	screenH := float64(r.Height)

	// Determine the effective scale
	effectiveScale := entity.Scale
	if entity.WorldHeight > 0 {
		// Use world height for scale-correct rendering
		// WorldHeight is in world units, convert to scale factor
		effectiveScale = entity.WorldHeight
	}

	// Calculate sprite screen height (perspective projection)
	spriteHeight := int(math.Abs(screenH/entity.TransformY) * effectiveScale)
	if spriteHeight <= 0 {
		spriteHeight = 1
	}

	// Calculate sprite screen width maintaining aspect ratio
	aspectRatio := float64(spriteW) / float64(spriteH)
	spriteWidth := int(float64(spriteHeight) * aspectRatio)
	if spriteWidth <= 0 {
		spriteWidth = 1
	}

	// Calculate vertical position
	// By default, sprites are centered on horizon (eye level)
	horizon := int(screenH / 2)
	drawStartY := horizon - spriteHeight/2
	drawEndY := horizon + spriteHeight/2

	// Apply vertical offset if specified
	if entity.VerticalOffset != 0 {
		// Convert world-space offset to screen-space pixels
		// Negative world offset (below eye level) should move sprite DOWN on screen (positive pixel offset)
		offsetPixels := int(-entity.VerticalOffset * screenH / entity.TransformY)
		drawStartY += offsetPixels
		drawEndY += offsetPixels
	}

	// Horizontal position (centered on screen X)
	drawStartX := entity.ScreenX - spriteWidth/2
	drawEndX := entity.ScreenX + spriteWidth/2

	return drawStartX, drawEndX, drawStartY, drawEndY, spriteWidth, spriteHeight
}

// SortSpritesByDistance sorts entities back-to-front for correct rendering.
func SortSpritesByDistance(entities []*SpriteEntity) {
	sort.Slice(entities, func(i, j int) bool {
		return entities[i].Distance > entities[j].Distance
	})
}

// IsSpriteColumnVisible checks if a sprite column is visible (not behind a wall).
func (r *Renderer) IsSpriteColumnVisible(screenX int, spriteDistance float64) bool {
	if screenX < 0 || screenX >= r.Width {
		return false
	}
	return spriteDistance < r.GetZBufferAt(screenX)
}

// ApplyFogToColor applies distance-based fog to a color.
func (r *Renderer) ApplyFogToColor(c sprite.PixelRGBA, distance float64) sprite.PixelRGBA {
	fogFactor := 1.0 - distance/FogDistance
	if fogFactor < MinFogFactor {
		fogFactor = MinFogFactor
	}
	if fogFactor > 1.0 {
		fogFactor = 1.0
	}
	return sprite.PixelRGBA{
		R: uint8(float64(c.R) * fogFactor),
		G: uint8(float64(c.G) * fogFactor),
		B: uint8(float64(c.B) * fogFactor),
		A: c.A,
	}
}

// ApplyOpacity applies an opacity multiplier to a color's alpha channel.
func ApplyOpacity(c sprite.PixelRGBA, opacity float64) sprite.PixelRGBA {
	return sprite.PixelRGBA{
		R: c.R,
		G: c.G,
		B: c.B,
		A: uint8(float64(c.A) * opacity),
	}
}

// SpriteDrawContext holds the computed values needed to draw a sprite.
// Pre-computed once per sprite, used for each column.
type SpriteDrawContext struct {
	StartX, EndX       int
	StartY, EndY       int
	SpriteWidth        int
	SpriteHeight       int
	ScreenSpriteWidth  int
	ScreenSpriteHeight int
	Distance           float64
	Opacity            float64
	FlipH              bool
	CurrentFrame       *sprite.Sprite
	// Highlight state (Phase 5)
	HighlightState int // 0=none, 1=in range, 2=targeted
	IsInteractable bool
}

// PrepareSpriteDrawContext computes all values needed to draw a sprite.
// Returns nil if the sprite should not be drawn (invisible, behind camera, etc.).
func (r *Renderer) PrepareSpriteDrawContext(entity *SpriteEntity) *SpriteDrawContext {
	if entity == nil || entity.Sheet == nil || !entity.Visible {
		return nil
	}

	// Get the current animation frame
	anim := entity.Sheet.GetAnimation(entity.AnimState)
	if anim == nil {
		// Fall back to idle
		anim = entity.Sheet.GetAnimation(sprite.AnimIdle)
		if anim == nil {
			return nil
		}
	}

	frame := anim.GetFrame(entity.AnimFrame)
	if frame == nil {
		return nil
	}

	// Compute screen bounds
	startX, endX, startY, endY, screenW, screenH := r.GetSpriteScreenBounds(entity, frame.Width, frame.Height)

	// Frustum culling: skip if entirely off screen
	if endX < 0 || startX >= r.Width || endY < 0 || startY >= r.Height {
		return nil
	}

	// Get a context from the pool to avoid allocation
	ctx := r.getPooledContext()
	ctx.StartX = startX
	ctx.EndX = endX
	ctx.StartY = startY
	ctx.EndY = endY
	ctx.SpriteWidth = frame.Width
	ctx.SpriteHeight = frame.Height
	ctx.ScreenSpriteWidth = screenW
	ctx.ScreenSpriteHeight = screenH
	ctx.Distance = entity.Distance
	ctx.Opacity = entity.Opacity
	ctx.FlipH = entity.FlipH
	ctx.CurrentFrame = frame
	ctx.HighlightState = entity.HighlightState
	ctx.IsInteractable = entity.IsInteractable
	return ctx
}

// GetSpritePixel retrieves a pixel from the sprite with optional horizontal flip.
// texX and texY are texture coordinates (0 to sprite width/height - 1).
func GetSpritePixel(frame *sprite.Sprite, texX, texY int, flipH bool) sprite.PixelRGBA {
	if frame == nil {
		return sprite.PixelRGBA{}
	}
	if flipH {
		texX = frame.Width - 1 - texX
	}
	c := frame.GetPixel(texX, texY)
	return sprite.PixelRGBA{R: c.R, G: c.G, B: c.B, A: c.A}
}

// DrawSpriteColumn draws a single vertical column of a sprite.
// screenX is the screen column to draw to.
// ctx contains pre-computed drawing parameters.
// pixels is the destination pixel buffer (row-major, RGBA).
func (r *Renderer) DrawSpriteColumn(screenX int, ctx *SpriteDrawContext, pixels []byte) {
	if ctx == nil || ctx.CurrentFrame == nil {
		return
	}

	// Bounds check
	if screenX < 0 || screenX >= r.Width {
		return
	}

	// Z-buffer test: skip if this column is behind a wall
	// Use > (not >=) to avoid z-fighting when sprites are exactly at wall distance
	if ctx.Distance > r.GetZBufferAt(screenX) {
		return
	}

	// Calculate texture X coordinate
	texX := (screenX - ctx.StartX) * ctx.SpriteWidth / ctx.ScreenSpriteWidth
	if texX < 0 || texX >= ctx.SpriteWidth {
		return
	}

	// Draw each pixel in the column
	r.drawColumnPixels(screenX, texX, ctx, pixels)
}

// drawColumnPixels draws all pixels in a sprite column.
func (r *Renderer) drawColumnPixels(screenX, texX int, ctx *SpriteDrawContext, pixels []byte) {
	for screenY := ctx.StartY; screenY < ctx.EndY; screenY++ {
		if screenY < 0 || screenY >= r.Height {
			continue
		}

		pixel := r.getSpritePixelAt(screenX, screenY, texX, ctx)
		if pixel.A == 0 {
			continue
		}

		r.writePixelToBuffer(screenX, screenY, pixel, pixels)
	}
}

// getSpritePixelAt gets a processed sprite pixel at the given position.
func (r *Renderer) getSpritePixelAt(screenX, screenY, texX int, ctx *SpriteDrawContext) sprite.PixelRGBA {
	// Calculate texture Y coordinate
	texY := (screenY - ctx.StartY) * ctx.SpriteHeight / ctx.ScreenSpriteHeight
	if texY < 0 || texY >= ctx.SpriteHeight {
		return sprite.PixelRGBA{}
	}

	// Get pixel from sprite
	pixel := GetSpritePixel(ctx.CurrentFrame, texX, texY, ctx.FlipH)
	if pixel.A == 0 {
		return pixel
	}

	// Apply highlight effect if enabled (with LOD check)
	if ctx.HighlightState > HighlightNone && ctx.IsInteractable {
		// Skip expensive edge detection at distance (LOD optimization)
		lodFlags := r.GetLODRenderFlags(ctx.Distance)
		if lodFlags.RenderHighlights {
			isEdge := IsPixelOnSpriteEdge(ctx.CurrentFrame, texX, texY, ctx.FlipH)
			pixel = ApplyHighlightToPixel(pixel, ctx.HighlightState, isEdge, r.highlightPulsePhase, r.HighlightConfig)
		}
	}

	// Apply fog and opacity
	pixel = r.ApplyFogToColor(pixel, ctx.Distance)
	if ctx.Opacity < 1.0 {
		pixel = ApplyOpacity(pixel, ctx.Opacity)
	}

	return pixel
}

// writePixelToBuffer writes a pixel to the buffer with alpha blending.
func (r *Renderer) writePixelToBuffer(screenX, screenY int, pixel sprite.PixelRGBA, pixels []byte) {
	idx := (screenY*r.Width + screenX) * 4
	if idx < 0 || idx+3 >= len(pixels) {
		return
	}

	if pixel.A < 255 {
		// Alpha blending
		alpha := float64(pixel.A) / 255.0
		invAlpha := 1.0 - alpha
		pixels[idx] = uint8(float64(pixel.R)*alpha + float64(pixels[idx])*invAlpha)
		pixels[idx+1] = uint8(float64(pixel.G)*alpha + float64(pixels[idx+1])*invAlpha)
		pixels[idx+2] = uint8(float64(pixel.B)*alpha + float64(pixels[idx+2])*invAlpha)
		pixels[idx+3] = 255
	} else {
		pixels[idx] = pixel.R
		pixels[idx+1] = pixel.G
		pixels[idx+2] = pixel.B
		pixels[idx+3] = 255
	}
}

// DrawSprite draws a complete sprite to the pixel buffer.
// pixels is the destination buffer in RGBA row-major format (len = width*height*4).
func (r *Renderer) DrawSprite(entity *SpriteEntity, pixels []byte) {
	ctx := r.PrepareSpriteDrawContext(entity)
	if ctx == nil {
		return
	}

	// Clamp X bounds to screen
	startX := ctx.StartX
	if startX < 0 {
		startX = 0
	}
	endX := ctx.EndX
	if endX > r.Width {
		endX = r.Width
	}

	// Get LOD-based column skip interval for performance optimization
	lod := r.GetLODLevel(ctx.Distance)
	columnSkip := ColumnSkipForLOD(lod)

	// Draw each visible column (with LOD-based skipping)
	for screenX := startX; screenX < endX; screenX += columnSkip {
		r.DrawSpriteColumn(screenX, ctx, pixels)
		// If skipping columns, duplicate pixels to adjacent columns for visual continuity
		if columnSkip > 1 {
			r.duplicateSpriteColumn(screenX, columnSkip, endX, ctx, pixels)
		}
	}
}

// duplicateSpriteColumn copies a rendered sprite column to adjacent columns.
// Used for LOD-based rendering where not every column is computed.
func (r *Renderer) duplicateSpriteColumn(sourceX, skip, endX int, ctx *SpriteDrawContext, pixels []byte) {
	for dx := 1; dx < skip && sourceX+dx < endX; dx++ {
		destX := sourceX + dx
		if destX >= r.Width {
			break
		}
		// Copy each pixel in the column
		for screenY := ctx.StartY; screenY < ctx.EndY; screenY++ {
			if screenY < 0 || screenY >= r.Height {
				continue
			}
			srcIdx := (screenY*r.Width + sourceX) * 4
			dstIdx := (screenY*r.Width + destX) * 4
			if srcIdx+3 < len(pixels) && dstIdx+3 < len(pixels) {
				pixels[dstIdx] = pixels[srcIdx]
				pixels[dstIdx+1] = pixels[srcIdx+1]
				pixels[dstIdx+2] = pixels[srcIdx+2]
				pixels[dstIdx+3] = pixels[srcIdx+3]
			}
		}
	}
}

// DrawSprites draws all visible sprites to the pixel buffer.
// Entities are sorted back-to-front before rendering.
func (r *Renderer) DrawSprites(entities []*SpriteEntity, pixels []byte) {
	if len(entities) == 0 || len(pixels) == 0 {
		return
	}

	// Reset context pool for this frame
	r.ResetContextPool()

	// Get max render distance from LOD config
	lodCfg := r.GetLODConfig()
	maxRenderDist := lodCfg.LowThreshold * 2.0 // Don't render beyond 2x the low LOD threshold

	// Reuse pre-allocated slice with [:0] pattern
	r.visibleSprites = r.visibleSprites[:0]
	for _, e := range entities {
		// Early distance cull before expensive transform
		dx := e.X - r.PlayerX
		dy := e.Y - r.PlayerY
		distSq := dx*dx + dy*dy
		if distSq > maxRenderDist*maxRenderDist {
			continue
		}

		if r.TransformEntityToScreen(e) {
			r.visibleSprites = append(r.visibleSprites, e)
		}
	}

	// Sort back-to-front
	SortSpritesByDistance(r.visibleSprites)

	// Draw each sprite
	for _, e := range r.visibleSprites {
		r.DrawSprite(e, pixels)
	}
}

// VisibleSpriteCount returns the number of sprites that would be visible
// after transform and culling. Useful for debugging and metrics.
func (r *Renderer) VisibleSpriteCount(entities []*SpriteEntity) int {
	count := 0
	for _, e := range entities {
		// Save original transform values
		origTX, origTY := e.TransformX, e.TransformY

		if r.TransformEntityToScreen(e) {
			ctx := r.PrepareSpriteDrawContext(e)
			if ctx != nil {
				count++
			}
		}

		// Restore (in case caller reuses entities)
		e.TransformX, e.TransformY = origTX, origTY
	}
	return count
}

// ============================================================
// Interaction Targeting System (Phase 5)
// ============================================================

// TargetingResult holds the result of an interaction raycast.
type TargetingResult struct {
	// HasTarget is true if a targetable entity was found.
	HasTarget bool
	// TargetEntity is the entity being targeted (nil if no target).
	TargetEntity *SpriteEntity
	// Distance is the distance to the targeted entity.
	Distance float64
	// ScreenX is the screen X position of the target center.
	ScreenX int
	// ScreenY is the screen Y position of the target center.
	ScreenY int
	// IsWithinRange is true if the target is within interaction range.
	IsWithinRange bool
}

// TargetingConfig configures the interaction targeting system.
type TargetingConfig struct {
	// CrosshairTolerance is the radius (in pixels) around the crosshair
	// that counts as "aimed at" an entity. Default is 30.
	CrosshairTolerance int
	// MaxTargetDistance is the maximum distance to target entities.
	// Entities further than this are not targetable. Default is 10.0.
	MaxTargetDistance float64
	// PreferCloser when true, prefers closer entities when multiple
	// are in the crosshair tolerance zone. Default is true.
	PreferCloser bool
}

// defaultTargetingConfig is a package-level default to avoid allocations.
// Use DefaultTargetingConfig() to get a copy if modifications are needed.
var defaultTargetingConfig = &TargetingConfig{
	CrosshairTolerance: 30,
	MaxTargetDistance:  10.0,
	PreferCloser:       true,
}

// DefaultTargetingConfig returns the default targeting configuration.
func DefaultTargetingConfig() *TargetingConfig {
	return &TargetingConfig{
		CrosshairTolerance: 30,
		MaxTargetDistance:  10.0,
		PreferCloser:       true,
	}
}

// FindTargetedEntity performs a raycast from the crosshair to find which
// interactable entity (if any) the player is looking at.
// entities should already have been transformed via TransformEntityToScreen.
// Returns the targeting result.
func (r *Renderer) FindTargetedEntity(entities []*SpriteEntity, cfg *TargetingConfig) TargetingResult {
	if cfg == nil {
		cfg = defaultTargetingConfig
	}

	// Reset context pool for targeting pass
	r.ResetContextPool()

	result := TargetingResult{
		HasTarget: false,
	}

	// Crosshair is at screen center
	crosshairX := r.Width / 2
	crosshairY := r.Height / 2

	var bestTarget *SpriteEntity
	bestDistance := cfg.MaxTargetDistance + 1

	for _, e := range entities {
		if e == nil || !e.IsInteractable || !e.Visible {
			continue
		}

		// Check if entity is in front of camera and within targeting distance
		if e.Distance <= 0 || e.Distance > cfg.MaxTargetDistance {
			continue
		}

		// Get entity's screen bounds
		ctx := r.PrepareSpriteDrawContext(e)
		if ctx == nil {
			continue
		}

		// Calculate screen center of the entity
		entityCenterX := (ctx.StartX + ctx.EndX) / 2
		entityCenterY := (ctx.StartY + ctx.EndY) / 2

		// Check if crosshair is within tolerance of entity bounds
		// First check if within expanded bounding box
		expandedStartX := ctx.StartX - cfg.CrosshairTolerance
		expandedEndX := ctx.EndX + cfg.CrosshairTolerance
		expandedStartY := ctx.StartY - cfg.CrosshairTolerance
		expandedEndY := ctx.EndY + cfg.CrosshairTolerance

		if crosshairX < expandedStartX || crosshairX > expandedEndX ||
			crosshairY < expandedStartY || crosshairY > expandedEndY {
			continue
		}

		// Within tolerance - check if this is the best target
		if cfg.PreferCloser {
			if e.Distance < bestDistance {
				bestTarget = e
				bestDistance = e.Distance
				result.ScreenX = entityCenterX
				result.ScreenY = entityCenterY
			}
		} else {
			// Prefer the one closest to crosshair center
			dx := float64(crosshairX - entityCenterX)
			dy := float64(crosshairY - entityCenterY)
			distToCenter := math.Sqrt(dx*dx + dy*dy)
			if bestTarget == nil || distToCenter < bestDistance {
				bestTarget = e
				bestDistance = distToCenter
				result.ScreenX = entityCenterX
				result.ScreenY = entityCenterY
			}
		}
	}

	if bestTarget != nil {
		result.HasTarget = true
		result.TargetEntity = bestTarget
		result.Distance = bestTarget.Distance
		result.IsWithinRange = bestTarget.InteractionRange <= 0 ||
			bestTarget.Distance <= bestTarget.InteractionRange
	}

	return result
}

// UpdateEntityHighlightStates updates the highlight state of entities
// based on their distance and targeting status.
// Call this each frame after FindTargetedEntity to update visual highlights.
func (r *Renderer) UpdateEntityHighlightStates(entities []*SpriteEntity, targetResult TargetingResult) {
	for _, e := range entities {
		if e == nil || !e.IsInteractable {
			continue
		}

		// Check if this is the targeted entity
		if targetResult.HasTarget && targetResult.TargetEntity == e {
			e.HighlightState = HighlightTargeted
			continue
		}

		// Check if within interaction range
		if e.InteractionRange > 0 && e.Distance > 0 && e.Distance <= e.InteractionRange {
			e.HighlightState = HighlightInRange
			continue
		}

		// Not targeted and not in range
		e.HighlightState = HighlightNone
	}
}

// GetInteractionPrompt returns the interaction text for a targeting result.
// Returns empty string if no target or no interaction available.
func GetInteractionPrompt(result TargetingResult) string {
	if !result.HasTarget || result.TargetEntity == nil {
		return ""
	}

	e := result.TargetEntity
	if !result.IsWithinRange {
		return ""
	}

	// Build prompt from entity data
	if e.UseText != "" {
		if e.DisplayName != "" {
			return e.UseText + " " + e.DisplayName
		}
		return e.UseText
	}

	// Default prompts based on interaction type
	switch e.InteractionType {
	case "pickup":
		if e.DisplayName != "" {
			return "Take " + e.DisplayName
		}
		return "Take"
	case "open":
		if e.DisplayName != "" {
			return "Open " + e.DisplayName
		}
		return "Open"
	case "use":
		if e.DisplayName != "" {
			return "Use " + e.DisplayName
		}
		return "Use"
	case "talk":
		if e.DisplayName != "" {
			return "Talk to " + e.DisplayName
		}
		return "Talk"
	case "read":
		if e.DisplayName != "" {
			return "Read " + e.DisplayName
		}
		return "Read"
	case "examine":
		if e.DisplayName != "" {
			return "Examine " + e.DisplayName
		}
		return "Examine"
	default:
		if e.DisplayName != "" {
			return "Interact with " + e.DisplayName
		}
		return "Interact"
	}
}

// IsEntityAtScreenPosition checks if an entity's screen bounds contain
// the given screen coordinates. Useful for mouse-based targeting.
func (r *Renderer) IsEntityAtScreenPosition(e *SpriteEntity, screenX, screenY int) bool {
	if e == nil || !e.Visible || e.Distance <= 0 {
		return false
	}

	ctx := r.PrepareSpriteDrawContext(e)
	if ctx == nil {
		return false
	}

	return screenX >= ctx.StartX && screenX <= ctx.EndX &&
		screenY >= ctx.StartY && screenY <= ctx.EndY
}

// FindEntityAtScreenPosition finds the closest interactable entity at a
// given screen position. Useful for mouse clicks or touch input.
func (r *Renderer) FindEntityAtScreenPosition(entities []*SpriteEntity, screenX, screenY int) *SpriteEntity {
	var closest *SpriteEntity
	closestDist := math.MaxFloat64

	for _, e := range entities {
		if e == nil || !e.IsInteractable || !e.Visible {
			continue
		}

		if r.IsEntityAtScreenPosition(e, screenX, screenY) {
			if e.Distance < closestDist {
				closest = e
				closestDist = e.Distance
			}
		}
	}

	return closest
}
