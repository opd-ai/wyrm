//go:build !noebiten

package raycast

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/wyrm/pkg/rendering/sprite"
)

// Draw renders the current view to the framebuffer using DDA raycasting.
// After calling Draw(), use UploadFramebuffer() to copy to the ebiten.Image.
func (r *Renderer) Draw(screen *ebiten.Image) {
	r.ClearFramebuffer()
	r.drawSky()
	r.drawFloorCeiling()
	r.drawWalls()
	screen.WritePixels(r.Framebuffer)
}

// drawSky renders procedural sky pixels above the horizon line.
// Uses the Skybox renderer for genre-appropriate sky colors, sun/moon, and weather.
func (r *Renderer) drawSky() {
	if r.Skybox == nil || r.Skybox.IsIndoor() {
		return // No sky rendering when indoors or skybox not initialized
	}

	horizonY := r.getHorizonLine()

	// Only render sky above horizon
	for y := 0; y < horizonY; y++ {
		// Calculate vertical position in sky (0=top of sky, 1=horizon)
		skyY := float64(y) / float64(horizonY)

		for x := 0; x < r.Width; x++ {
			// Calculate horizontal position (0=left, 1=right)
			skyX := float64(x) / float64(r.Width)

			// Get sky color at this position
			skyColor := r.Skybox.GetSkyColorAt(skyX, skyY)
			r.SetPixelColor(x, y, skyColor)
		}
	}
}

// DrawSpritesToScreen renders billboard sprites to an ebiten.Image.
// This is a convenience method that handles the conversion between the
// pixel buffer-based DrawSprites and ebiten's Image type.
func (r *Renderer) DrawSpritesToScreen(entities []*SpriteEntity, screen *ebiten.Image) {
	if len(entities) == 0 {
		return
	}

	// Reuse pre-allocated slice with [:0] pattern
	r.visibleSprites = r.visibleSprites[:0]
	for _, e := range entities {
		if r.TransformEntityToScreen(e) {
			r.visibleSprites = append(r.visibleSprites, e)
		}
	}

	if len(r.visibleSprites) == 0 {
		return
	}

	// Sort back-to-front
	SortSpritesByDistance(r.visibleSprites)

	// Draw each sprite directly to the framebuffer
	for _, e := range r.visibleSprites {
		r.drawSpriteToFramebuffer(e)
	}

	// Upload the modified framebuffer to the screen
	screen.WritePixels(r.Framebuffer)
}

// drawSpriteToFramebuffer draws a single sprite to the framebuffer.
func (r *Renderer) drawSpriteToFramebuffer(entity *SpriteEntity) {
	ctx := r.PrepareSpriteDrawContext(entity)
	if ctx == nil || ctx.CurrentFrame == nil {
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
		r.drawSpriteColumnToFramebuffer(screenX, ctx)
	}
}

// drawSpriteColumnToFramebuffer draws a single sprite column to the framebuffer.
func (r *Renderer) drawSpriteColumnToFramebuffer(screenX int, ctx *SpriteDrawContext) {
	if !r.isValidSpriteColumn(screenX, ctx) {
		return
	}

	texX := r.calculateSpriteTexX(screenX, ctx)
	if texX < 0 || texX >= ctx.SpriteWidth {
		return
	}

	r.drawSpriteColumnPixels(screenX, texX, ctx)
}

// isValidSpriteColumn checks if the column should be drawn.
func (r *Renderer) isValidSpriteColumn(screenX int, ctx *SpriteDrawContext) bool {
	if screenX < 0 || screenX >= r.Width {
		return false
	}
	return ctx.Distance <= r.GetZBufferAt(screenX)
}

// calculateSpriteTexX computes the texture X coordinate for a screen column.
func (r *Renderer) calculateSpriteTexX(screenX int, ctx *SpriteDrawContext) int {
	return (screenX - ctx.StartX) * ctx.SpriteWidth / ctx.ScreenSpriteWidth
}

// drawSpriteColumnPixels renders all pixels in a sprite column.
func (r *Renderer) drawSpriteColumnPixels(screenX, texX int, ctx *SpriteDrawContext) {
	for screenY := ctx.StartY; screenY < ctx.EndY; screenY++ {
		if screenY < 0 || screenY >= r.Height {
			continue
		}

		texY := (screenY - ctx.StartY) * ctx.SpriteHeight / ctx.ScreenSpriteHeight
		if texY < 0 || texY >= ctx.SpriteHeight {
			continue
		}

		pixel := GetSpritePixel(ctx.CurrentFrame, texX, texY, ctx.FlipH)
		if pixel.A == 0 {
			continue
		}

		pixel = r.applyPixelEffects(pixel, ctx)
		if pixel.A > 0 {
			r.BlendPixel(screenX, screenY, pixel.R, pixel.G, pixel.B, pixel.A)
		}
	}
}

// applyPixelEffects applies fog and opacity to a pixel.
func (r *Renderer) applyPixelEffects(pixel sprite.PixelRGBA, ctx *SpriteDrawContext) sprite.PixelRGBA {
	pixel = r.ApplyFogToColor(pixel, ctx.Distance)
	if ctx.Opacity < 1.0 {
		pixel = ApplyOpacity(pixel, ctx.Opacity)
	}
	return pixel
}

// drawFloorCeiling renders textured floor and ceiling using raycasting.
// When skybox is active, only the floor is rendered; sky replaces the ceiling.
func (r *Renderer) drawFloorCeiling() {
	halfHeight := r.Height / 2
	horizonY := r.getHorizonLine()
	renderCeiling := r.Skybox == nil || r.Skybox.IsIndoor()

	// Start from horizonY + 1 to avoid rendering the horizon row twice
	// (once as floor and once as ceiling via the ceilY = Height - y - 1 mirror).
	// This prevents a 1-pixel overlap artifact at the exact center for odd heights.
	startY := horizonY + 1
	if startY >= r.Height {
		return
	}

	// Hoist FOV ray direction calculation out of loop - depends only on PlayerA and FOV
	// which are constant for the frame. Avoids ~360 redundant trig calls per frame.
	rayDirX0, rayDirY0, rayDirX1, rayDirY1 := r.calculateFOVRayDirections()

	for y := startY; y < r.Height; y++ {
		rowDistance := r.calculateRowDistance(y, halfHeight)
		floorStepX, floorStepY := r.calculateFloorStep(rowDistance, rayDirX0, rayDirY0, rayDirX1, rayDirY1)
		floorX, floorY := r.calculateFloorStart(rowDistance, rayDirX0, rayDirY0)

		r.renderFloorCeilingRow(y, horizonY, rowDistance, floorX, floorY, floorStepX, floorStepY, renderCeiling)
	}
}

// calculateFOVRayDirections computes ray directions for leftmost and rightmost columns.
func (r *Renderer) calculateFOVRayDirections() (rayDirX0, rayDirY0, rayDirX1, rayDirY1 float64) {
	rayDirX0 = math.Cos(r.PlayerA - r.FOV/2)
	rayDirY0 = math.Sin(r.PlayerA - r.FOV/2)
	rayDirX1 = math.Cos(r.PlayerA + r.FOV/2)
	rayDirY1 = math.Sin(r.PlayerA + r.FOV/2)
	return rayDirX0, rayDirY0, rayDirX1, rayDirY1
}

// calculateRowDistance computes the horizontal distance from camera to floor for a row.
func (r *Renderer) calculateRowDistance(y, halfHeight int) float64 {
	p := y - halfHeight
	// Guard against division by zero when y equals halfHeight (horizon line)
	if p == 0 {
		return FogDistance * 2 // Return maximum distance for horizon row
	}
	posZ := 0.5 * float64(r.Height)
	rowDistance := posZ / float64(p)
	if rowDistance < 0 {
		rowDistance = 0
	}
	if rowDistance > FogDistance*2 {
		rowDistance = FogDistance * 2
	}
	return rowDistance
}

// calculateFloorStep computes the step values for each column.
func (r *Renderer) calculateFloorStep(rowDistance, rayDirX0, rayDirY0, rayDirX1, rayDirY1 float64) (stepX, stepY float64) {
	stepX = rowDistance * (rayDirX1 - rayDirX0) / float64(r.Width)
	stepY = rowDistance * (rayDirY1 - rayDirY0) / float64(r.Width)
	return stepX, stepY
}

// calculateFloorStart computes the starting floor position for a row.
func (r *Renderer) calculateFloorStart(rowDistance, rayDirX0, rayDirY0 float64) (floorX, floorY float64) {
	floorX = r.PlayerX + rowDistance*rayDirX0
	floorY = r.PlayerY + rowDistance*rayDirY0
	return floorX, floorY
}

// renderFloorCeilingRow renders a single row of floor and (optionally) ceiling pixels.
// renderCeiling should be false when skybox is active outdoors.
func (r *Renderer) renderFloorCeilingRow(y, horizonY int, rowDistance, floorX, floorY, floorStepX, floorStepY float64, renderCeiling bool) {
	for x := 0; x < r.Width; x++ {
		texX := floorX - math.Floor(floorX)
		texY := floorY - math.Floor(floorY)

		floorColor := r.GetFloorTextureColor(texX, texY, rowDistance)
		r.SetPixelColor(x, y, floorColor)

		// Mirror floor Y to get ceiling Y position
		ceilY := r.Height - y - 1
		if renderCeiling && ceilY >= 0 && ceilY < horizonY {
			ceilColor := r.GetCeilingTextureColor(texX, texY, rowDistance)
			r.SetPixelColor(x, ceilY, ceilColor)
		}

		floorX += floorStepX
		floorY += floorStepY
	}
}

// drawWalls casts rays and renders wall columns, populating the ZBuffer.
func (r *Renderer) drawWalls() {
	// Ensure ZBuffer is sized correctly
	if len(r.ZBuffer) != r.Width {
		r.ZBuffer = make([]float64, r.Width)
	}
	for x := 0; x < r.Width; x++ {
		r.drawWallColumn(x)
	}
}

// drawWallColumn renders a single vertical wall strip with texture mapping.
func (r *Renderer) drawWallColumn(x int) {
	cameraX := 2.0*float64(x)/float64(r.Width) - 1.0
	rayAngle := r.PlayerA + cameraX*(r.FOV/2)
	rayDirX := math.Cos(rayAngle)
	rayDirY := math.Sin(rayAngle)

	distance, wallType, wallX, side, cellWallHeight, cellFlags := r.castRayWithTexCoordAndFlags(rayDirX, rayDirY)
	distance *= math.Cos(cameraX * (r.FOV / 2)) // Fix fisheye
	distance = clampDistance(distance)

	// Store distance in ZBuffer for sprite occlusion
	r.ZBuffer[x] = distance

	// Calculate wall height with variable height multiplier
	wallHeight := calculateWallHeightWithMultiplier(r.Height, distance, cellWallHeight)
	drawStart, drawEnd := calculateDrawBounds(r.Height, wallHeight, cellWallHeight, r.PlayerZ)

	// Texture X coordinate (0-1 range)
	texX := wallX - math.Floor(wallX)

	// Check if wall is transparent
	isTransparent := cellFlags&FlagTransparent != 0
	isSemiOpaque := cellFlags&FlagSemiOpaque != 0

	if isTransparent || isSemiOpaque {
		r.renderTransparentWallStrip(x, drawStart, drawEnd, wallHeight, wallType, texX, distance, side, cellFlags)
	} else {
		r.renderWallStrip(x, drawStart, drawEnd, wallHeight, wallType, texX, distance, side)
	}
}

// clampDistance ensures distance is within valid range.
func clampDistance(distance float64) float64 {
	if distance < MinWallDistance {
		return MinWallDistance
	}
	return distance
}

// calculateWallHeight computes wall height from distance (standard height).
func calculateWallHeight(screenHeight int, distance float64) int {
	return calculateWallHeightWithMultiplier(screenHeight, distance, DefaultWallHeight)
}

// calculateWallHeightWithMultiplier computes wall height with a height multiplier.
func calculateWallHeightWithMultiplier(screenHeight int, distance, heightMultiplier float64) int {
	wallHeight := int(float64(screenHeight) / distance * heightMultiplier)
	maxHeight := int(float64(screenHeight) * MaxWallHeight)
	if wallHeight > maxHeight {
		wallHeight = maxHeight
	}
	return wallHeight
}

// calculateDrawBounds determines where to start and end drawing the wall column.
// Takes into account the wall height multiplier and player eye height.
func calculateDrawBounds(screenHeight, wallHeight int, cellWallHeight, playerZ float64) (drawStart, drawEnd int) {
	// Calculate horizon line (can be offset by player pitch later)
	horizon := screenHeight / 2

	// Calculate where wall top and bottom should be based on player eye height
	// Player at Z=0.5 sees standard walls centered
	// Walls taller than 1.0 extend above and below the standard view
	eyeOffset := (playerZ - 0.5) * float64(wallHeight)

	// For variable height walls, adjust draw bounds
	// A wall with height 2.0 should extend twice as far from center
	halfWall := wallHeight / 2

	drawStart = horizon - halfWall + int(eyeOffset)
	drawEnd = horizon + halfWall + int(eyeOffset)

	return drawStart, drawEnd
}

// renderWallStrip draws a vertical strip of wall pixels to the framebuffer.
func (r *Renderer) renderWallStrip(x, drawStart, drawEnd, wallHeight, wallType int, texX, distance float64, side int) {
	sideDarken := getSideDarkenFactor(side)

	for y := drawStart; y < drawEnd; y++ {
		if y < 0 || y >= r.Height {
			continue
		}
		texY := float64(y-drawStart) / float64(wallHeight)
		wallColor := r.GetWallTextureColor(wallType, texX, texY, distance)
		wallColor = applySideDarkening(wallColor, sideDarken)
		r.SetPixelColor(x, y, wallColor)
	}
}

// renderTransparentWallStrip draws a vertical wall strip with alpha blending.
// Used for transparent barriers like fences, glass, or partially destroyed walls.
func (r *Renderer) renderTransparentWallStrip(x, drawStart, drawEnd, wallHeight, wallType int, texX, distance float64, side int, flags CellFlags) {
	sideDarken := getSideDarkenFactor(side)

	// Determine transparency level based on flags
	transparency := getTransparencyForFlags(flags)

	for y := drawStart; y < drawEnd; y++ {
		if y < 0 || y >= r.Height {
			continue
		}
		texY := float64(y-drawStart) / float64(wallHeight)

		// For semi-opaque walls (like fences), check if pixel should be solid
		if flags&FlagSemiOpaque != 0 {
			if isSemiOpaqueGap(texX, texY) {
				continue // Skip this pixel, leaving background visible
			}
		}

		wallColor := r.GetWallTextureColor(wallType, texX, texY, distance)
		wallColor = applySideDarkening(wallColor, sideDarken)

		// Apply transparency
		wallColor.A = uint8(float64(255) * transparency)

		// Use alpha blending
		r.BlendPixelColor(x, y, wallColor)
	}
}

// castRayWithTexCoord performs DDA raycasting and returns texture coordinate info.
// Returns: (perpWallDist, wallType, wallX, side, wallHeight)
func (r *Renderer) castRayWithTexCoord(rayDirX, rayDirY float64) (float64, int, float64, int, float64) {
	dist, wallType, wallX, side, height, _ := r.castRayWithTexCoordAndFlags(rayDirX, rayDirY)
	return dist, wallType, wallX, side, height
}

// castRayWithTexCoordAndFlags performs DDA raycasting and returns texture coordinate info plus flags.
// Returns: (perpWallDist, wallType, wallX, side, wallHeight, cellFlags)
func (r *Renderer) castRayWithTexCoordAndFlags(rayDirX, rayDirY float64) (float64, int, float64, int, float64, CellFlags) {
	mapX := int(r.PlayerX)
	mapY := int(r.PlayerY)

	deltaDistX, deltaDistY := calculateDeltaDist(rayDirX, rayDirY)
	sideDistX, sideDistY, stepX, stepY := calculateSideDist(
		r.PlayerX, r.PlayerY, mapX, mapY,
		rayDirX, rayDirY, deltaDistX, deltaDistY,
	)

	hit, side, sideDistX, sideDistY, mapX, mapY := r.performDDA(sideDistX, sideDistY, deltaDistX, deltaDistY, stepX, stepY, mapX, mapY)

	if !hit {
		return MaxRayDistance, 0, 0.0, 0, DefaultWallHeight, 0
	}

	// Calculate perpendicular wall distance
	var perpWallDist float64
	if side == 0 {
		perpWallDist = sideDistX - deltaDistX
	} else {
		perpWallDist = sideDistY - deltaDistY
	}

	// Calculate wall X coordinate (where on the wall the ray hit)
	var wallX float64
	if side == 0 {
		wallX = r.PlayerY + perpWallDist*rayDirY
	} else {
		wallX = r.PlayerX + perpWallDist*rayDirX
	}

	// Get cell data including wall height and flags
	cell := r.GetMapCell(mapX, mapY)

	return perpWallDist, cell.WallType, wallX, side, cell.WallHeight, cell.Flags
}
