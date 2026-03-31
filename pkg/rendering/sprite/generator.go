package sprite

import (
	"image/color"
	"math"
	"math/rand"
)

// FrameGenerator is a function type for generating a single animation frame.
type FrameGenerator func(width, height int, state string, frameIdx int) *Sprite

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

// buildAnimation creates an animation with the specified parameters.
func buildAnimation(animName string, looping bool, frameCount int, genFunc FrameGenerator, width, height int, state string) *Animation {
	anim := NewAnimation(animName, looping)
	for i := 0; i < frameCount; i++ {
		frame := genFunc(width, height, state, i)
		anim.AddFrame(frame)
	}
	return anim
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

	// Create a frame generator closure for humanoid frames
	frameGen := func(w, h int, state string, frameIdx int) *Sprite {
		return g.generateHumanoidFrame(w, h, key, rng, state, frameIdx)
	}

	sheet.AddAnimation(buildAnimation(AnimIdle, true, 4, frameGen, width, height, "idle"))
	sheet.AddAnimation(buildAnimation(AnimWalk, true, 8, frameGen, width, height, "walk"))
	sheet.AddAnimation(buildAnimation(AnimAttack, false, 6, frameGen, width, height, "attack"))
	sheet.AddAnimation(buildAnimation(AnimDead, false, 1, frameGen, width, height, "dead"))

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

	g.drawHumanoidLeg(sprite, centerX-legSpacing-legWidth, centerX-legSpacing, top, bottom, width, c, leftLegOffset)
	g.drawHumanoidLeg(sprite, centerX+legSpacing, centerX+legSpacing+legWidth, top, bottom, width, c, rightLegOffset)
}

// drawHumanoidLeg draws one humanoid leg with vertical offset.
func (g *Generator) drawHumanoidLeg(sprite *Sprite, xStart, xEnd, top, bottom, width int, c color.RGBA, yOffset int) {
	for y := top + yOffset; y < bottom; y++ {
		if y < top || y >= bottom {
			continue
		}
		g.drawHorizontalLine(sprite, xStart, xEnd, y, width, c)
	}
}

// drawHorizontalLine draws a horizontal line with bounds checking.
func (g *Generator) drawHorizontalLine(sprite *Sprite, xStart, xEnd, y, width int, c color.RGBA) {
	for x := xStart; x < xEnd; x++ {
		if x >= 0 && x < width {
			sprite.SetPixel(x, y, c)
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

	// Create a frame generator closure for creature frames
	frameGen := func(w, h int, state string, frameIdx int) *Sprite {
		return g.generateCreatureFrame(w, h, key, rng, state, frameIdx)
	}

	sheet.AddAnimation(buildAnimation(AnimIdle, true, 4, frameGen, width, height, "idle"))
	sheet.AddAnimation(buildAnimation(AnimWalk, true, 6, frameGen, width, height, "walk"))

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

	g.drawQuadrupedBody(sprite, bodyLeft, bodyRight, bodyTop, bodyBottom, primary)
	g.drawQuadrupedHead(sprite, width, height, secondary)
	g.drawQuadrupedLegs(sprite, width, height, bodyBottom, primary, state, frameIdx)
}

// drawQuadrupedBody draws the main body of a quadruped creature.
func (g *Generator) drawQuadrupedBody(sprite *Sprite, left, right, top, bottom int, c color.RGBA) {
	for y := top; y < bottom; y++ {
		for x := left; x < right; x++ {
			sprite.SetPixel(x, y, c)
		}
	}
}

// drawQuadrupedHead draws the head of a quadruped creature.
func (g *Generator) drawQuadrupedHead(sprite *Sprite, width, height int, c color.RGBA) {
	headRight := width / 6
	headTop := height / 3
	headBottom := height / 2
	for y := headTop; y < headBottom; y++ {
		for x := 0; x < headRight; x++ {
			sprite.SetPixel(x, y, c)
		}
	}
}

// drawQuadrupedLegs draws the four legs of a quadruped creature with animation.
func (g *Generator) drawQuadrupedLegs(sprite *Sprite, width, height, bodyBottom int, c color.RGBA, state string, frameIdx int) {
	legWidth := width / 10
	legBottom := height
	legOffsets := []int{width / 5, width * 2 / 5, width * 3 / 5, width * 4 / 5}

	animOffset := g.calculateLegAnimOffset(state, frameIdx)

	for i, legX := range legOffsets {
		offset := animOffset
		if i%2 == 1 {
			offset = -offset
		}
		g.drawSingleLeg(sprite, legX, bodyBottom+offset, legBottom, legWidth, width, height, c)
	}
}

// calculateLegAnimOffset calculates the animation offset for leg movement.
func (g *Generator) calculateLegAnimOffset(state string, frameIdx int) int {
	if state == "walk" {
		return int(math.Sin(float64(frameIdx)*math.Pi/3) * 2)
	}
	return 0
}

// drawSingleLeg draws a single leg at the specified position.
func (g *Generator) drawSingleLeg(sprite *Sprite, legX, top, bottom, legWidth, maxWidth, maxHeight int, c color.RGBA) {
	bodyBottom := maxHeight * 2 / 3 // Derive from body proportions
	for y := top; y < bottom; y++ {
		if y >= bodyBottom && y < maxHeight {
			for x := legX; x < legX+legWidth && x < maxWidth; x++ {
				sprite.SetPixel(x, y, c)
			}
		}
	}
}

// drawSerpentine draws a snake-like creature.
func (g *Generator) drawSerpentine(sprite *Sprite, width, height int, primary, secondary color.RGBA, state string, frameIdx int) {
	centerY := height / 2
	bodyHeight := height / 4

	g.drawSerpentineBody(sprite, width, height, centerY, bodyHeight, primary, frameIdx)
	g.drawSerpentineHead(sprite, width, height, centerY, secondary)
}

// drawSerpentineBody draws the sinusoidal body of a serpentine creature.
func (g *Generator) drawSerpentineBody(sprite *Sprite, width, height, centerY, bodyHeight int, c color.RGBA, frameIdx int) {
	for x := 0; x < width; x++ {
		waveOffset := math.Sin(float64(x)/float64(width)*math.Pi*2 + float64(frameIdx)*0.5)
		centerOffset := int(waveOffset * float64(height/6))

		thickness := g.calculateSerpentineThickness(x, width, bodyHeight)

		for y := centerY - bodyHeight/2 + centerOffset; y < centerY+bodyHeight/2+centerOffset; y++ {
			if y >= 0 && y < height {
				if math.Abs(float64(y-centerY-centerOffset)) < thickness/2 {
					sprite.SetPixel(x, y, c)
				}
			}
		}
	}
}

// calculateSerpentineThickness calculates body thickness with tapering at ends.
func (g *Generator) calculateSerpentineThickness(x, width, bodyHeight int) float64 {
	thickness := float64(bodyHeight)
	if x < width/4 {
		thickness *= float64(x) / float64(width/4)
	}
	if x > width*3/4 {
		thickness *= float64(width-x) / float64(width/4)
	}
	return thickness
}

// drawSerpentineHead draws the head of a serpentine creature.
func (g *Generator) drawSerpentineHead(sprite *Sprite, width, height, centerY int, c color.RGBA) {
	headX := 2
	headY := centerY
	headRadius := height / 6
	g.drawCircle(sprite, headX, headY, headRadius, c)
}

// drawCircle draws a filled circle at the specified position.
func (g *Generator) drawCircle(sprite *Sprite, centerX, centerY, radius int, c color.RGBA) {
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx*dx+dy*dy <= radius*radius {
				sprite.SetPixel(centerX+dx, centerY+dy, c)
			}
		}
	}
}

// drawAvian draws a bird-like creature.
func (g *Generator) drawAvian(sprite *Sprite, width, height int, primary, secondary color.RGBA, state string, frameIdx int) {
	centerX := width / 2
	centerY := height / 2
	bodyWidth := width / 3
	bodyHeight := height / 4

	g.drawOvalBody(sprite, centerX, centerY, bodyWidth, bodyHeight, width, primary)
	g.drawCircularHead(sprite, centerX, centerY-bodyHeight, height/8, secondary)
	g.drawAvianWings(sprite, centerX, centerY, bodyWidth, width/3, width, primary, state, frameIdx)
}

// drawOvalBody draws an oval-shaped body.
func (g *Generator) drawOvalBody(sprite *Sprite, centerX, centerY, bodyWidth, bodyHeight, maxWidth int, c color.RGBA) {
	for y := centerY - bodyHeight; y < centerY+bodyHeight; y++ {
		dy := float64(y - centerY)
		xRange := int(math.Sqrt(float64(bodyWidth*bodyWidth) * (1 - (dy*dy)/float64(bodyHeight*bodyHeight))))
		g.drawHorizontalLine(sprite, centerX-xRange, centerX+xRange+1, y, maxWidth, c)
	}
}

// drawCircularHead draws a circular head.
func (g *Generator) drawCircularHead(sprite *Sprite, centerX, centerY, radius int, c color.RGBA) {
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx*dx+dy*dy <= radius*radius {
				sprite.SetPixel(centerX+dx, centerY+dy, c)
			}
		}
	}
}

// drawAvianWings draws bird wings with optional animation.
func (g *Generator) drawAvianWings(sprite *Sprite, centerX, centerY, bodyWidth, wingSpan, maxWidth int, c color.RGBA, state string, frameIdx int) {
	wingAngle := g.calculateWingAngle(state, frameIdx)
	for i := 0; i < wingSpan; i++ {
		wingY := centerY - int(float64(i)*wingAngle)
		sprite.SetPixel(centerX-bodyWidth-i, wingY, c)
		sprite.SetPixel(centerX+bodyWidth+i, wingY, c)
	}
}

// calculateWingAngle returns the wing angle based on animation state.
func (g *Generator) calculateWingAngle(state string, frameIdx int) float64 {
	if state == "walk" || state == "idle" {
		return math.Sin(float64(frameIdx)*math.Pi/2) * 0.3
	}
	return 0.0
}

// generateSingleFrameSheet is a helper to create sprite sheets with one animation frame.
func (g *Generator) generateSingleFrameSheet(width, height int, key SpriteCacheKey, rng *rand.Rand, loop bool, frameGenerator func(int, int, SpriteCacheKey, *rand.Rand) *Sprite) *SpriteSheet {
	sheet := NewSpriteSheet(width, height)
	idleAnim := NewAnimation(AnimIdle, loop)
	frame := frameGenerator(width, height, key, rng)
	idleAnim.AddFrame(frame)
	sheet.AddAnimation(idleAnim)
	return sheet
}

// generateVehicleSheet creates sprites for vehicle entities.
func (g *Generator) generateVehicleSheet(key SpriteCacheKey, rng *rand.Rand) *SpriteSheet {
	width := int(float64(DefaultSpriteWidth) * key.Scale * 1.5)
	height := int(float64(DefaultSpriteHeight) * key.Scale)
	return g.generateSingleFrameSheet(width, height, key, rng, true, g.generateVehicleFrame)
}

// generateVehicleFrame creates a vehicle sprite.
func (g *Generator) generateVehicleFrame(width, height int, key SpriteCacheKey, rng *rand.Rand) *Sprite {
	sprite := NewSprite(width, height)
	primary := unpackColor(key.PrimaryColor)
	secondary := unpackColor(key.SecondaryColor)

	bounds := g.calculateVehicleBounds(width, height)
	g.fillVehicleBody(sprite, bounds, primary)
	g.drawVehicleWheels(sprite, bounds, height, secondary)

	return sprite
}

// vehicleBounds defines the rectangular bounds of a vehicle body.
type vehicleBounds struct {
	top, bottom, left, right int
}

// calculateVehicleBounds computes the vehicle body dimensions.
func (g *Generator) calculateVehicleBounds(width, height int) vehicleBounds {
	return vehicleBounds{
		top:    height / 4,
		bottom: height * 3 / 4,
		left:   width / 8,
		right:  width * 7 / 8,
	}
}

// fillVehicleBody fills the rectangular body of the vehicle.
func (g *Generator) fillVehicleBody(sprite *Sprite, bounds vehicleBounds, color color.RGBA) {
	for y := bounds.top; y < bounds.bottom; y++ {
		for x := bounds.left; x < bounds.right; x++ {
			sprite.SetPixel(x, y, color)
		}
	}
}

// drawVehicleWheels draws circular wheels on the vehicle.
func (g *Generator) drawVehicleWheels(sprite *Sprite, bounds vehicleBounds, height int, wheelColor color.RGBA) {
	wheelRadius := height / 8
	wheels := []struct{ x, y int }{
		{bounds.left + wheelRadius, bounds.bottom},
		{bounds.right - wheelRadius, bounds.bottom},
	}

	for _, w := range wheels {
		g.drawFilledCircle(sprite, w.x, w.y, wheelRadius, wheelColor)
	}
}

// drawFilledCircle draws a filled circle on a sprite.
func (g *Generator) drawFilledCircle(sprite *Sprite, cx, cy, radius int, c color.RGBA) {
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx*dx+dy*dy <= radius*radius {
				sprite.SetPixel(cx+dx, cy+dy, c)
			}
		}
	}
}

// generateObjectSheet creates sprites for static objects.
func (g *Generator) generateObjectSheet(key SpriteCacheKey, rng *rand.Rand) *SpriteSheet {
	width := int(float64(DefaultSpriteWidth) * key.Scale)
	height := int(float64(DefaultSpriteHeight) * key.Scale)
	return g.generateSingleFrameSheet(width, height, key, rng, false, g.generateObjectFrame)
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
