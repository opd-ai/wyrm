package sprite

import (
	"image/color"
)

// Sprite dimensions and rendering constants.
const (
	// DefaultSpriteWidth is the base width for humanoid sprites.
	DefaultSpriteWidth = 32
	// DefaultSpriteHeight is the base height for humanoid sprites.
	DefaultSpriteHeight = 48
	// MaxSpriteWidth caps sprite dimensions for memory safety.
	MaxSpriteWidth = 128
	// MaxSpriteHeight caps sprite dimensions for memory safety.
	MaxSpriteHeight = 192
	// AnimFrameRate is the default animation playback rate (frames per second).
	AnimFrameRate = 8.0
)

// Animation state identifiers.
const (
	AnimIdle   = "idle"
	AnimWalk   = "walk"
	AnimRun    = "run"
	AnimAttack = "attack"
	AnimCast   = "cast"
	AnimSneak  = "sneak"
	AnimDead   = "dead"
	AnimSit    = "sit"
	AnimWork   = "work"
)

// Sprite category identifiers.
const (
	CategoryHumanoid = "humanoid"
	CategoryCreature = "creature"
	CategoryVehicle  = "vehicle"
	CategoryObject   = "object"
	CategoryEffect   = "effect"
)

// PixelRGBA is a lightweight RGBA pixel type for sprite rendering.
// Used instead of color.RGBA for explicit struct layout in rendering code.
type PixelRGBA struct {
	R, G, B, A uint8
}

// TestPixelColor is a visible color used for testing sprite rendering.
var TestPixelColor = color.RGBA{R: 255, G: 128, B: 64, A: 255}

// Sprite represents a single frame of procedurally generated pixel data.
// Pixels are stored in row-major order (top-to-bottom, left-to-right).
type Sprite struct {
	// Width is the sprite width in pixels.
	Width int
	// Height is the sprite height in pixels.
	Height int
	// Pixels holds RGBA pixel data in row-major order.
	// Length must equal Width * Height.
	Pixels []color.RGBA
}

// NewSprite creates a new empty sprite with the given dimensions.
// The pixel buffer is allocated but not initialized (transparent black).
func NewSprite(width, height int) *Sprite {
	if width <= 0 {
		width = DefaultSpriteWidth
	}
	if height <= 0 {
		height = DefaultSpriteHeight
	}
	if width > MaxSpriteWidth {
		width = MaxSpriteWidth
	}
	if height > MaxSpriteHeight {
		height = MaxSpriteHeight
	}
	return &Sprite{
		Width:  width,
		Height: height,
		Pixels: make([]color.RGBA, width*height),
	}
}

// GetPixel returns the pixel at (x, y). Returns transparent black if out of bounds.
func (s *Sprite) GetPixel(x, y int) color.RGBA {
	if x < 0 || x >= s.Width || y < 0 || y >= s.Height {
		return color.RGBA{}
	}
	return s.Pixels[y*s.Width+x]
}

// SetPixel sets the pixel at (x, y). Does nothing if out of bounds.
func (s *Sprite) SetPixel(x, y int, c color.RGBA) {
	if x < 0 || x >= s.Width || y < 0 || y >= s.Height {
		return
	}
	s.Pixels[y*s.Width+x] = c
}

// MemorySize returns the approximate memory footprint in bytes.
func (s *Sprite) MemorySize() int64 {
	// Each RGBA pixel is 4 bytes
	return int64(s.Width * s.Height * 4)
}

// Clone creates a deep copy of the sprite.
func (s *Sprite) Clone() *Sprite {
	clone := NewSprite(s.Width, s.Height)
	copy(clone.Pixels, s.Pixels)
	return clone
}

// Fill fills the entire sprite with a single color.
func (s *Sprite) Fill(c color.RGBA) {
	for i := range s.Pixels {
		s.Pixels[i] = c
	}
}

// FlipHorizontal returns a new sprite flipped horizontally.
func (s *Sprite) FlipHorizontal() *Sprite {
	flipped := NewSprite(s.Width, s.Height)
	for y := 0; y < s.Height; y++ {
		for x := 0; x < s.Width; x++ {
			flipped.SetPixel(s.Width-1-x, y, s.GetPixel(x, y))
		}
	}
	return flipped
}

// Animation holds the frames for a single animation state.
type Animation struct {
	// Name is the animation state identifier (e.g., "idle", "walk").
	Name string
	// Frames are the sprite frames for this animation.
	Frames []*Sprite
	// FrameDuration is seconds per frame (1.0 / framerate).
	FrameDuration float64
	// Loop indicates whether the animation loops.
	Loop bool
}

// NewAnimation creates a new animation with default frame duration.
func NewAnimation(name string, loop bool) *Animation {
	return &Animation{
		Name:          name,
		Frames:        nil,
		FrameDuration: 1.0 / AnimFrameRate,
		Loop:          loop,
	}
}

// AddFrame adds a sprite frame to the animation.
func (a *Animation) AddFrame(frame *Sprite) {
	a.Frames = append(a.Frames, frame)
}

// GetFrame returns the frame at the given index, clamped to valid range.
// Returns nil if no frames exist.
func (a *Animation) GetFrame(index int) *Sprite {
	if len(a.Frames) == 0 {
		return nil
	}
	if a.Loop {
		index = index % len(a.Frames)
	} else if index >= len(a.Frames) {
		index = len(a.Frames) - 1
	}
	if index < 0 {
		index = 0
	}
	return a.Frames[index]
}

// FrameCount returns the number of frames in this animation.
func (a *Animation) FrameCount() int {
	return len(a.Frames)
}

// Duration returns the total duration of the animation in seconds.
func (a *Animation) Duration() float64 {
	return float64(len(a.Frames)) * a.FrameDuration
}

// SpriteSheet holds all animations for a single entity appearance.
type SpriteSheet struct {
	// Animations maps state name to animation data.
	Animations map[string]*Animation
	// BaseWidth is the sprite width at scale 1.0.
	BaseWidth int
	// BaseHeight is the sprite height at scale 1.0.
	BaseHeight int
}

// NewSpriteSheet creates a new empty sprite sheet.
func NewSpriteSheet(width, height int) *SpriteSheet {
	return &SpriteSheet{
		Animations: make(map[string]*Animation),
		BaseWidth:  width,
		BaseHeight: height,
	}
}

// AddAnimation adds an animation to the sprite sheet.
func (ss *SpriteSheet) AddAnimation(anim *Animation) {
	if anim != nil {
		ss.Animations[anim.Name] = anim
	}
}

// GetAnimation returns the animation for the given state name.
// Returns the idle animation if the requested state doesn't exist.
// Returns nil if no animations exist.
func (ss *SpriteSheet) GetAnimation(stateName string) *Animation {
	if anim, ok := ss.Animations[stateName]; ok {
		return anim
	}
	// Fall back to idle
	if anim, ok := ss.Animations[AnimIdle]; ok {
		return anim
	}
	// Return any animation if no idle
	for _, anim := range ss.Animations {
		return anim
	}
	return nil
}

// GetFrame returns a specific frame from a specific animation.
func (ss *SpriteSheet) GetFrame(stateName string, frameIndex int) *Sprite {
	anim := ss.GetAnimation(stateName)
	if anim == nil {
		return nil
	}
	return anim.GetFrame(frameIndex)
}

// MemorySize returns the approximate memory footprint in bytes.
func (ss *SpriteSheet) MemorySize() int64 {
	var total int64
	for _, anim := range ss.Animations {
		for _, frame := range anim.Frames {
			total += frame.MemorySize()
		}
	}
	return total
}

// AnimationNames returns all animation state names in this sheet.
func (ss *SpriteSheet) AnimationNames() []string {
	names := make([]string, 0, len(ss.Animations))
	for name := range ss.Animations {
		names = append(names, name)
	}
	return names
}
