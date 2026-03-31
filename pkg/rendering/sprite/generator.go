package sprite

import (
	"image/color"
	"math"
	"math/rand"
)

// Generator produces procedural sprites for entities based on their Appearance.
type Generator struct {
	seed  int64
	genre string
	rng   *rand.Rand
}

// NewGenerator creates a new sprite generator for the given genre and seed.
func NewGenerator(genre string, seed int64) *Generator {
	return &Generator{
		seed:  seed,
		genre: genre,
		rng:   rand.New(rand.NewSource(seed)),
	}
}

// GenerateSheet generates a complete sprite sheet from a cache key.
func (g *Generator) GenerateSheet(key SpriteCacheKey) *SpriteSheet {
	// Create seeded RNG for this specific sprite
	rng := rand.New(rand.NewSource(key.Seed))

	switch key.Category {
	case CategoryHumanoid:
		return g.generateHumanoidSheet(key, rng)
	case CategoryCreature:
		return g.generateCreatureSheet(key, rng)
	case CategoryVehicle:
		return g.generateVehicleSheet(key, rng)
	case CategoryObject:
		return g.generateObjectSheet(key, rng)
	case CategoryEffect:
		return g.generateEffectSheet(key, rng)
	default:
		return g.generateHumanoidSheet(key, rng)
	}
}

// generateHumanoidSheet creates sprites for humanoid entities (NPCs, players).
func (g *Generator) generateHumanoidSheet(key SpriteCacheKey, rng *rand.Rand) *SpriteSheet {
	width := int(float64(DefaultSpriteWidth) * key.Scale)
	height := int(float64(DefaultSpriteHeight) * key.Scale)
	if width < 8 {
		width = 8
	}
	if height < 12 {
		height = 12
	}

	sheet := NewSpriteSheet(width, height)

	// Generate idle animation (4 frames with subtle movement)
	idleAnim := NewAnimation(AnimIdle, true)
	for i := 0; i < 4; i++ {
		frame := g.generateHumanoidFrame(width, height, key, rng, "idle", i)
		idleAnim.AddFrame(frame)
	}
	sheet.AddAnimation(idleAnim)

	// Generate walk animation (8 frames)
	walkAnim := NewAnimation(AnimWalk, true)
	for i := 0; i < 8; i++ {
		frame := g.generateHumanoidFrame(width, height, key, rng, "walk", i)
		walkAnim.AddFrame(frame)
	}
	sheet.AddAnimation(walkAnim)

	// Generate attack animation (6 frames, non-looping)
	attackAnim := NewAnimation(AnimAttack, false)
	for i := 0; i < 6; i++ {
		frame := g.generateHumanoidFrame(width, height, key, rng, "attack", i)
		attackAnim.AddFrame(frame)
	}
	sheet.AddAnimation(attackAnim)

	// Generate dead animation (1 frame, non-looping)
	deadAnim := NewAnimation(AnimDead, false)
	deadFrame := g.generateHumanoidFrame(width, height, key, rng, "dead", 0)
	deadAnim.AddFrame(deadFrame)
	sheet.AddAnimation(deadAnim)

	return sheet
}

// generateHumanoidFrame creates a single humanoid sprite frame.
func (g *Generator) generateHumanoidFrame(width, height int, key SpriteCacheKey, rng *rand.Rand, state string, frameIdx int) *Sprite {
	sprite := NewSprite(width, height)
	primary := unpackColor(key.PrimaryColor)
	secondary := unpackColor(key.SecondaryColor)
	accent := unpackColor(key.AccentColor)

	// Calculate body regions
	headTop := 0
	headBottom := height / 6
	torsoTop := headBottom
	torsoBottom := height * 2 / 3
	legsBottom := height

	centerX := width / 2

	// Generate based on state
	switch state {
	case "dead":
		g.drawDeadHumanoid(sprite, width, height, primary, secondary)
	default:
		// Standing pose with animation offset
		animOffset := g.getAnimOffset(state, frameIdx)

		// Draw legs
		g.drawLegs(sprite, centerX, torsoBottom, legsBottom, width, primary, animOffset)

		// Draw torso
		g.drawTorso(sprite, centerX, torsoTop, torsoBottom, width, secondary)

		// Draw head
		g.drawHead(sprite, centerX, headTop, headBottom, width, accent, rng)

		// Add body plan specific details
		g.addBodyPlanDetails(sprite, key.BodyPlan, width, height, accent, rng)
	}

	return sprite
}

// getAnimOffset calculates leg/arm offset for walking animation.
func (g *Generator) getAnimOffset(state string, frameIdx int) float64 {
	switch state {
	case "walk":
		// Sinusoidal leg swing
		return math.Sin(float64(frameIdx) * math.Pi / 4)
	case "idle":
		// Subtle breathing motion
		return math.Sin(float64(frameIdx)*math.Pi/2) * 0.1
	default:
		return 0
	}
}

// drawLegs draws humanoid legs with animation offset.
func (g *Generator) drawLegs(sprite *Sprite, centerX, top, bottom, width int, c color.RGBA, animOffset float64) {
	legWidth := width / 6
	legSpacing := width / 8
	leftLegOffset := int(animOffset * 2)
	rightLegOffset := -leftLegOffset

	// Left leg
	for y := top + leftLegOffset; y < bottom; y++ {
		if y < top || y >= bottom {
			continue
		}
		for x := centerX - legSpacing - legWidth; x < centerX-legSpacing; x++ {
			if x >= 0 && x < width {
				sprite.SetPixel(x, y, c)
			}
		}
	}

	// Right leg
	for y := top + rightLegOffset; y < bottom; y++ {
		if y < top || y >= bottom {
			continue
		}
		for x := centerX + legSpacing; x < centerX+legSpacing+legWidth; x++ {
			if x >= 0 && x < width {
				sprite.SetPixel(x, y, c)
			}
		}
	}
}

// drawTorso draws the humanoid torso.
func (g *Generator) drawTorso(sprite *Sprite, centerX, top, bottom, width int, c color.RGBA) {
	torsoWidth := width / 3
	for y := top; y < bottom; y++ {
		// Slight taper from shoulders to waist
		progress := float64(y-top) / float64(bottom-top)
		currentWidth := int(float64(torsoWidth) * (1.0 - progress*0.2))
		for x := centerX - currentWidth; x < centerX+currentWidth; x++ {
			if x >= 0 && x < width {
				sprite.SetPixel(x, y, c)
			}
		}
	}
}

// drawHead draws the humanoid head.
func (g *Generator) drawHead(sprite *Sprite, centerX, top, bottom, width int, c color.RGBA, rng *rand.Rand) {
	headWidth := width / 4
	headHeight := bottom - top

	// Circular head
	for y := top; y < bottom; y++ {
		dy := float64(y-top) - float64(headHeight)/2
		radius := float64(headWidth)
		xRange := int(math.Sqrt(radius*radius - dy*dy))
		for x := centerX - xRange; x <= centerX+xRange; x++ {
			if x >= 0 && x < width {
				sprite.SetPixel(x, y, c)
			}
		}
	}
}

// drawDeadHumanoid draws a fallen humanoid figure.
func (g *Generator) drawDeadHumanoid(sprite *Sprite, width, height int, primary, secondary color.RGBA) {
	// Horizontal body (lying down)
	bodyTop := height / 2
	bodyBottom := bodyTop + height/8
	for y := bodyTop; y < bodyBottom; y++ {
		for x := width / 6; x < width*5/6; x++ {
			sprite.SetPixel(x, y, primary)
		}
	}
	// Head on one end
	headCenterX := width / 6
	headRadius := height / 12
	for y := bodyTop - headRadius; y < bodyTop+headRadius; y++ {
		for x := headCenterX - headRadius; x < headCenterX+headRadius; x++ {
			if x >= 0 && x < width && y >= 0 && y < height {
				sprite.SetPixel(x, y, secondary)
			}
		}
	}
}

// addBodyPlanDetails adds occupation/role specific visual elements.
func (g *Generator) addBodyPlanDetails(sprite *Sprite, bodyPlan string, width, height int, accent color.RGBA, rng *rand.Rand) {
	switch bodyPlan {
	case "guard":
		// Add helmet and spear silhouette
		g.addHelmet(sprite, width, height, accent)
		g.addSpear(sprite, width, height, accent)
	case "merchant":
		// Add wide robe silhouette
		g.addRobe(sprite, width, height, accent)
	case "healer":
		// Add staff
		g.addStaff(sprite, width, height, accent)
	case "smith", "blacksmith":
		// Add hammer
		g.addHammer(sprite, width, height, accent)
	}
}

// addHelmet adds a helmet to the sprite head area.
func (g *Generator) addHelmet(sprite *Sprite, width, height int, c color.RGBA) {
	centerX := width / 2
	top := 0
	helmetWidth := width / 3
	helmetHeight := height / 8
	for y := top; y < helmetHeight; y++ {
		for x := centerX - helmetWidth/2; x < centerX+helmetWidth/2; x++ {
			if x >= 0 && x < width {
				sprite.SetPixel(x, y, c)
			}
		}
	}
}

// addSpear adds a spear silhouette to the right side.
func (g *Generator) addSpear(sprite *Sprite, width, height int, c color.RGBA) {
	spearX := width * 3 / 4
	for y := 0; y < height*3/4; y++ {
		sprite.SetPixel(spearX, y, c)
	}
	// Spear tip
	for i := 0; i < 3; i++ {
		sprite.SetPixel(spearX-1+i, 0, c)
		sprite.SetPixel(spearX-1+i, 1, c)
	}
}

// addRobe adds a wide robe silhouette.
func (g *Generator) addRobe(sprite *Sprite, width, height int, c color.RGBA) {
	centerX := width / 2
	robeTop := height / 4
	robeBottom := height * 3 / 4
	for y := robeTop; y < robeBottom; y++ {
		progress := float64(y-robeTop) / float64(robeBottom-robeTop)
		robeWidth := int(float64(width/4) * (1.0 + progress*0.5))
		for x := centerX - robeWidth; x < centerX+robeWidth; x++ {
			if x >= 0 && x < width {
				existing := sprite.GetPixel(x, y)
				if existing.A == 0 {
					sprite.SetPixel(x, y, c)
				}
			}
		}
	}
}

// addStaff adds a staff silhouette to the side.
func (g *Generator) addStaff(sprite *Sprite, width, height int, c color.RGBA) {
	staffX := width / 4
	for y := height / 8; y < height*7/8; y++ {
		sprite.SetPixel(staffX, y, c)
	}
	// Staff top orb
	orbY := height / 8
	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			if dx*dx+dy*dy <= 4 {
				sprite.SetPixel(staffX+dx, orbY+dy, c)
			}
		}
	}
}

// addHammer adds a hammer silhouette.
func (g *Generator) addHammer(sprite *Sprite, width, height int, c color.RGBA) {
	handleX := width * 3 / 4
	// Handle
	for y := height / 4; y < height*2/3; y++ {
		sprite.SetPixel(handleX, y, c)
	}
	// Head
	headY := height / 4
	for dx := -3; dx <= 3; dx++ {
		for dy := -2; dy <= 2; dy++ {
			sprite.SetPixel(handleX+dx, headY+dy, c)
		}
	}
}

// generateCreatureSheet creates sprites for creature entities.
func (g *Generator) generateCreatureSheet(key SpriteCacheKey, rng *rand.Rand) *SpriteSheet {
	width := int(float64(DefaultSpriteWidth) * key.Scale * 1.5)
	height := int(float64(DefaultSpriteHeight) * key.Scale)
	if width < 16 {
		width = 16
	}
	if height < 16 {
		height = 16
	}

	sheet := NewSpriteSheet(width, height)

	// Generate idle animation
	idleAnim := NewAnimation(AnimIdle, true)
	for i := 0; i < 4; i++ {
		frame := g.generateCreatureFrame(width, height, key, rng, "idle", i)
		idleAnim.AddFrame(frame)
	}
	sheet.AddAnimation(idleAnim)

	// Generate walk animation
	walkAnim := NewAnimation(AnimWalk, true)
	for i := 0; i < 6; i++ {
		frame := g.generateCreatureFrame(width, height, key, rng, "walk", i)
		walkAnim.AddFrame(frame)
	}
	sheet.AddAnimation(walkAnim)

	return sheet
}

// generateCreatureFrame creates a single creature sprite frame.
func (g *Generator) generateCreatureFrame(width, height int, key SpriteCacheKey, rng *rand.Rand, state string, frameIdx int) *Sprite {
	sprite := NewSprite(width, height)
	primary := unpackColor(key.PrimaryColor)
	secondary := unpackColor(key.SecondaryColor)

	// Generate based on body plan
	switch key.BodyPlan {
	case "quadruped":
		g.drawQuadruped(sprite, width, height, primary, secondary, state, frameIdx)
	case "serpentine":
		g.drawSerpentine(sprite, width, height, primary, secondary, state, frameIdx)
	case "avian":
		g.drawAvian(sprite, width, height, primary, secondary, state, frameIdx)
	default:
		// Default to quadruped
		g.drawQuadruped(sprite, width, height, primary, secondary, state, frameIdx)
	}

	return sprite
}

// drawQuadruped draws a four-legged creature.
func (g *Generator) drawQuadruped(sprite *Sprite, width, height int, primary, secondary color.RGBA, state string, frameIdx int) {
	// Body
	bodyTop := height / 3
	bodyBottom := height * 2 / 3
	bodyLeft := width / 6
	bodyRight := width * 5 / 6

	for y := bodyTop; y < bodyBottom; y++ {
		for x := bodyLeft; x < bodyRight; x++ {
			sprite.SetPixel(x, y, primary)
		}
	}

	// Head
	headRight := width / 6
	headTop := height / 3
	headBottom := height / 2
	for y := headTop; y < headBottom; y++ {
		for x := 0; x < headRight; x++ {
			sprite.SetPixel(x, y, secondary)
		}
	}

	// Legs (4)
	legWidth := width / 10
	legTop := bodyBottom
	legBottom := height
	legOffsets := []int{width / 5, width * 2 / 5, width * 3 / 5, width * 4 / 5}

	animOffset := 0
	if state == "walk" {
		animOffset = int(math.Sin(float64(frameIdx)*math.Pi/3) * 2)
	}

	for i, legX := range legOffsets {
		offset := animOffset
		if i%2 == 1 {
			offset = -offset
		}
		for y := legTop + offset; y < legBottom; y++ {
			if y >= legTop && y < height {
				for x := legX; x < legX+legWidth && x < width; x++ {
					sprite.SetPixel(x, y, primary)
				}
			}
		}
	}
}

// drawSerpentine draws a snake-like creature.
func (g *Generator) drawSerpentine(sprite *Sprite, width, height int, primary, secondary color.RGBA, state string, frameIdx int) {
	centerY := height / 2
	bodyHeight := height / 4

	for x := 0; x < width; x++ {
		// Sinusoidal body
		waveOffset := math.Sin(float64(x)/float64(width)*math.Pi*2 + float64(frameIdx)*0.5)
		centerOffset := int(waveOffset * float64(height/6))

		for y := centerY - bodyHeight/2 + centerOffset; y < centerY+bodyHeight/2+centerOffset; y++ {
			if y >= 0 && y < height {
				// Taper at ends
				thickness := float64(bodyHeight)
				if x < width/4 {
					thickness *= float64(x) / float64(width/4)
				}
				if x > width*3/4 {
					thickness *= float64(width-x) / float64(width/4)
				}
				if math.Abs(float64(y-centerY-centerOffset)) < thickness/2 {
					sprite.SetPixel(x, y, primary)
				}
			}
		}
	}

	// Head (at left end)
	headX := 2
	headY := centerY
	headRadius := height / 6
	for dy := -headRadius; dy <= headRadius; dy++ {
		for dx := -headRadius; dx <= headRadius; dx++ {
			if dx*dx+dy*dy <= headRadius*headRadius {
				sprite.SetPixel(headX+dx, headY+dy, secondary)
			}
		}
	}
}

// drawAvian draws a bird-like creature.
func (g *Generator) drawAvian(sprite *Sprite, width, height int, primary, secondary color.RGBA, state string, frameIdx int) {
	centerX := width / 2
	centerY := height / 2

	// Body (oval)
	bodyWidth := width / 3
	bodyHeight := height / 4
	for y := centerY - bodyHeight; y < centerY+bodyHeight; y++ {
		dy := float64(y - centerY)
		xRange := int(math.Sqrt(float64(bodyWidth*bodyWidth) * (1 - (dy*dy)/float64(bodyHeight*bodyHeight))))
		for x := centerX - xRange; x <= centerX+xRange; x++ {
			if x >= 0 && x < width {
				sprite.SetPixel(x, y, primary)
			}
		}
	}

	// Head
	headRadius := height / 8
	headY := centerY - bodyHeight
	for dy := -headRadius; dy <= headRadius; dy++ {
		for dx := -headRadius; dx <= headRadius; dx++ {
			if dx*dx+dy*dy <= headRadius*headRadius {
				sprite.SetPixel(centerX+dx, headY+dy, secondary)
			}
		}
	}

	// Wings
	wingSpan := width / 3
	wingAngle := 0.0
	if state == "walk" || state == "idle" {
		wingAngle = math.Sin(float64(frameIdx)*math.Pi/2) * 0.3
	}

	for i := 0; i < wingSpan; i++ {
		wingY := centerY - int(float64(i)*wingAngle)
		sprite.SetPixel(centerX-bodyWidth-i, wingY, primary)
		sprite.SetPixel(centerX+bodyWidth+i, wingY, primary)
	}
}

// generateVehicleSheet creates sprites for vehicle entities.
func (g *Generator) generateVehicleSheet(key SpriteCacheKey, rng *rand.Rand) *SpriteSheet {
	width := int(float64(DefaultSpriteWidth) * key.Scale * 1.5)
	height := int(float64(DefaultSpriteHeight) * key.Scale)

	sheet := NewSpriteSheet(width, height)

	// Vehicles typically have a single idle frame
	idleAnim := NewAnimation(AnimIdle, true)
	frame := g.generateVehicleFrame(width, height, key, rng)
	idleAnim.AddFrame(frame)
	sheet.AddAnimation(idleAnim)

	return sheet
}

// generateVehicleFrame creates a vehicle sprite.
func (g *Generator) generateVehicleFrame(width, height int, key SpriteCacheKey, rng *rand.Rand) *Sprite {
	sprite := NewSprite(width, height)
	primary := unpackColor(key.PrimaryColor)
	secondary := unpackColor(key.SecondaryColor)

	// Simple vehicle body (rectangle with wheels)
	bodyTop := height / 4
	bodyBottom := height * 3 / 4
	bodyLeft := width / 8
	bodyRight := width * 7 / 8

	// Body
	for y := bodyTop; y < bodyBottom; y++ {
		for x := bodyLeft; x < bodyRight; x++ {
			sprite.SetPixel(x, y, primary)
		}
	}

	// Wheels (circles at bottom corners)
	wheelRadius := height / 8
	wheels := []struct{ x, y int }{
		{bodyLeft + wheelRadius, bodyBottom},
		{bodyRight - wheelRadius, bodyBottom},
	}

	for _, w := range wheels {
		for dy := -wheelRadius; dy <= wheelRadius; dy++ {
			for dx := -wheelRadius; dx <= wheelRadius; dx++ {
				if dx*dx+dy*dy <= wheelRadius*wheelRadius {
					sprite.SetPixel(w.x+dx, w.y+dy, secondary)
				}
			}
		}
	}

	return sprite
}

// generateObjectSheet creates sprites for static objects.
func (g *Generator) generateObjectSheet(key SpriteCacheKey, rng *rand.Rand) *SpriteSheet {
	width := int(float64(DefaultSpriteWidth) * key.Scale)
	height := int(float64(DefaultSpriteHeight) * key.Scale)

	sheet := NewSpriteSheet(width, height)

	// Objects have a single frame
	idleAnim := NewAnimation(AnimIdle, false)
	frame := g.generateObjectFrame(width, height, key, rng)
	idleAnim.AddFrame(frame)
	sheet.AddAnimation(idleAnim)

	return sheet
}

// generateObjectFrame creates an object sprite.
func (g *Generator) generateObjectFrame(width, height int, key SpriteCacheKey, rng *rand.Rand) *Sprite {
	sprite := NewSprite(width, height)
	primary := unpackColor(key.PrimaryColor)

	// Simple rectangle for generic objects
	margin := 2
	for y := margin; y < height-margin; y++ {
		for x := margin; x < width-margin; x++ {
			sprite.SetPixel(x, y, primary)
		}
	}

	return sprite
}

// generateEffectSheet creates sprites for visual effects.
func (g *Generator) generateEffectSheet(key SpriteCacheKey, rng *rand.Rand) *SpriteSheet {
	width := int(float64(DefaultSpriteWidth) * key.Scale)
	height := int(float64(DefaultSpriteHeight) * key.Scale)

	sheet := NewSpriteSheet(width, height)

	// Effects animate
	effectAnim := NewAnimation(AnimIdle, true)
	for i := 0; i < 8; i++ {
		frame := g.generateEffectFrame(width, height, key, rng, i)
		effectAnim.AddFrame(frame)
	}
	sheet.AddAnimation(effectAnim)

	return sheet
}

// generateEffectFrame creates an effect sprite frame.
func (g *Generator) generateEffectFrame(width, height int, key SpriteCacheKey, rng *rand.Rand, frameIdx int) *Sprite {
	sprite := NewSprite(width, height)
	primary := unpackColor(key.PrimaryColor)

	// Pulsing glow effect
	centerX := width / 2
	centerY := height / 2
	maxRadius := min(width, height) / 2

	phase := float64(frameIdx) * math.Pi / 4
	intensity := (math.Sin(phase) + 1) / 2

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			dx := x - centerX
			dy := y - centerY
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist < float64(maxRadius) {
				alpha := uint8(255 * intensity * (1 - dist/float64(maxRadius)))
				c := primary
				c.A = alpha
				sprite.SetPixel(x, y, c)
			}
		}
	}

	return sprite
}

// unpackColor converts a packed RGBA uint32 to color.RGBA.
func unpackColor(packed uint32) color.RGBA {
	return color.RGBA{
		R: uint8((packed >> 24) & 0xFF),
		G: uint8((packed >> 16) & 0xFF),
		B: uint8((packed >> 8) & 0xFF),
		A: uint8(packed & 0xFF),
	}
}

// packColor converts color.RGBA to packed RGBA uint32.
func packColor(c color.RGBA) uint32 {
	return uint32(c.R)<<24 | uint32(c.G)<<16 | uint32(c.B)<<8 | uint32(c.A)
}

// min returns the smaller of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
