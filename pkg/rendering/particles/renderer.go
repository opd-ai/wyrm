// Package particles renderer integration.
package particles

import (
	"image/color"
	"math"
)

// Renderer draws particles to a pixel buffer.
type Renderer struct {
	width  int
	height int
}

// NewRenderer creates a new particle renderer.
func NewRenderer(width, height int) *Renderer {
	return &Renderer{
		width:  width,
		height: height,
	}
}

// SetDimensions updates the renderer dimensions.
func (r *Renderer) SetDimensions(width, height int) {
	r.width = width
	r.height = height
}

// Draw renders all particles in the system to the pixel buffer.
// pixels must be width*height*4 bytes in RGBA row-major format.
func (r *Renderer) Draw(system *System, pixels []byte) {
	if system == nil || len(pixels) < r.width*r.height*4 {
		return
	}

	for _, p := range system.Particles() {
		if !p.Active {
			continue
		}
		r.drawParticle(p, pixels)
	}
}

// drawParticle renders a single particle.
func (r *Renderer) drawParticle(p *Particle, pixels []byte) {
	// Convert normalized coords to screen coords using rounding for accuracy
	screenX := int(math.Round(p.X * float64(r.width)))
	screenY := int(math.Round(p.Y * float64(r.height)))
	size := int(math.Round(p.Size))
	if size < 1 {
		size = 1
	}

	// Get color with alpha
	c := GetColor(p)
	if c.A == 0 {
		return
	}

	// Draw based on particle type
	switch p.Type {
	case TypeRain:
		r.drawRainDrop(screenX, screenY, size, c, pixels)
	case TypeSnow:
		r.drawSnowFlake(screenX, screenY, size, c, pixels)
	case TypeSparks, TypeFire:
		r.drawGlow(screenX, screenY, size, c, pixels)
	default:
		r.drawCircle(screenX, screenY, size/2, c, pixels)
	}
}

// drawPixelBlend draws a pixel with alpha blending.
func (r *Renderer) drawPixelBlend(x, y int, c color.RGBA, pixels []byte) {
	if x < 0 || x >= r.width || y < 0 || y >= r.height {
		return
	}
	if c.A == 0 {
		return
	}

	idx := (y*r.width + x) * 4
	if idx < 0 || idx+3 >= len(pixels) {
		return
	}

	// Alpha blending
	alpha := float64(c.A) / 255.0
	invAlpha := 1.0 - alpha

	pixels[idx] = uint8(float64(c.R)*alpha + float64(pixels[idx])*invAlpha)
	pixels[idx+1] = uint8(float64(c.G)*alpha + float64(pixels[idx+1])*invAlpha)
	pixels[idx+2] = uint8(float64(c.B)*alpha + float64(pixels[idx+2])*invAlpha)
	pixels[idx+3] = 255
}

// drawCircle draws a filled circle.
func (r *Renderer) drawCircle(cx, cy, radius int, c color.RGBA, pixels []byte) {
	if radius < 1 {
		r.drawPixelBlend(cx, cy, c, pixels)
		return
	}

	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx*dx+dy*dy <= radius*radius {
				r.drawPixelBlend(cx+dx, cy+dy, c, pixels)
			}
		}
	}
}

// drawRainDrop draws a vertical line for rain.
func (r *Renderer) drawRainDrop(x, y, length int, c color.RGBA, pixels []byte) {
	for i := 0; i < length; i++ {
		// Fade along length
		alpha := uint8(float64(c.A) * float64(length-i) / float64(length))
		dropColor := color.RGBA{c.R, c.G, c.B, alpha}
		r.drawPixelBlend(x, y+i, dropColor, pixels)
	}
}

// drawSnowFlake draws a small star shape for snow.
func (r *Renderer) drawSnowFlake(x, y, size int, c color.RGBA, pixels []byte) {
	r.drawPixelBlend(x, y, c, pixels)
	if size >= 2 {
		r.drawPixelBlend(x-1, y, c, pixels)
		r.drawPixelBlend(x+1, y, c, pixels)
		r.drawPixelBlend(x, y-1, c, pixels)
		r.drawPixelBlend(x, y+1, c, pixels)
	}
	if size >= 3 {
		r.drawPixelBlend(x-1, y-1, c, pixels)
		r.drawPixelBlend(x+1, y-1, c, pixels)
		r.drawPixelBlend(x-1, y+1, c, pixels)
		r.drawPixelBlend(x+1, y+1, c, pixels)
	}
}

// drawGlow draws a glowing particle with falloff.
func (r *Renderer) drawGlow(cx, cy, size int, c color.RGBA, pixels []byte) {
	if size < 1 {
		size = 1
	}
	for dy := -size; dy <= size; dy++ {
		for dx := -size; dx <= size; dx++ {
			dist := dx*dx + dy*dy
			maxDist := size * size
			if dist <= maxDist {
				// Quadratic falloff
				intensity := 1.0 - float64(dist)/float64(maxDist)
				alpha := uint8(float64(c.A) * intensity * intensity)
				glowColor := color.RGBA{c.R, c.G, c.B, alpha}
				r.drawPixelBlend(cx+dx, cy+dy, glowColor, pixels)
			}
		}
	}
}

// WeatherPreset creates a preset for common weather effects.
type WeatherPreset struct {
	Type      string
	Intensity float64 // 0.0-1.0
	Direction float64 // Wind direction in radians (0 = right, π/2 = down)
}

// CreateWeatherEmitters creates emitters for a weather preset.
func CreateWeatherEmitters(preset *WeatherPreset, seed int64) []*Emitter {
	if preset == nil {
		return nil
	}

	// Number of emitters based on intensity
	count := int(preset.Intensity*10) + 1
	emitters := make([]*Emitter, count)

	for i := 0; i < count; i++ {
		e := NewEmitter(preset.Type, seed+int64(i))
		// Spread emitters across top of screen
		e.X = float64(i) / float64(count)
		e.Y = 0.0
		// Adjust rate based on intensity
		e.Rate *= preset.Intensity
		emitters[i] = e
	}

	return emitters
}

// CombatEffect creates a combat-related particle effect.
type CombatEffect struct {
	Type  string
	X, Y  float64
	Count int
	Seed  int64
}

// SpawnCombatEffect spawns a combat particle effect.
func (s *System) SpawnCombatEffect(effect *CombatEffect) {
	if effect == nil {
		return
	}
	s.SpawnBurst(effect.Type, effect.X, effect.Y, effect.Count)
}
