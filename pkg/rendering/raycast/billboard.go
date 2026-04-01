// Package raycast provides the first-person raycasting renderer.
// This file implements billboard sprite rendering for entities.
package raycast

import (
	"math"
	"sort"

	"github.com/opd-ai/wyrm/pkg/rendering/sprite"
)

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
	invDet := 1.0 / (planeX*dirY - dirX*planeY)
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

	// Calculate sprite screen height (perspective projection)
	spriteHeight := int(math.Abs(screenH/entity.TransformY) * entity.Scale)
	if spriteHeight <= 0 {
		spriteHeight = 1
	}

	// Calculate sprite screen width maintaining aspect ratio
	aspectRatio := float64(spriteW) / float64(spriteH)
	spriteWidth := int(float64(spriteHeight) * aspectRatio)
	if spriteWidth <= 0 {
		spriteWidth = 1
	}

	// Vertical position (centered on horizon)
	drawStartY := int(screenH/2) - spriteHeight/2
	drawEndY := int(screenH/2) + spriteHeight/2

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

	return &SpriteDrawContext{
		StartX:             startX,
		EndX:               endX,
		StartY:             startY,
		EndY:               endY,
		SpriteWidth:        frame.Width,
		SpriteHeight:       frame.Height,
		ScreenSpriteWidth:  screenW,
		ScreenSpriteHeight: screenH,
		Distance:           entity.Distance,
		Opacity:            entity.Opacity,
		FlipH:              entity.FlipH,
		CurrentFrame:       frame,
	}
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
	if ctx.Distance >= r.GetZBufferAt(screenX) {
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

	// Draw each visible column
	for screenX := startX; screenX < endX; screenX++ {
		r.DrawSpriteColumn(screenX, ctx, pixels)
	}
}

// DrawSprites draws all visible sprites to the pixel buffer.
// Entities are sorted back-to-front before rendering.
func (r *Renderer) DrawSprites(entities []*SpriteEntity, pixels []byte) {
	if len(entities) == 0 || len(pixels) == 0 {
		return
	}

	// Reuse pre-allocated slice with [:0] pattern
	r.visibleSprites = r.visibleSprites[:0]
	for _, e := range entities {
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
