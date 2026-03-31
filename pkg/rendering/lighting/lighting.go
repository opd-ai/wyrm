// Package lighting implements dynamic lighting for the raycaster.
package lighting

import (
	"image/color"
	"math"
)

// Light type constants.
const (
	TypePoint       = "point"
	TypeDirectional = "directional"
	TypeSpot        = "spot"
	TypeAmbient     = "ambient"
)

// Default lighting parameters.
const (
	DefaultAmbient      = 0.2
	DefaultFalloff      = 2.0  // Quadratic falloff exponent
	DefaultMaxRange     = 10.0 // Maximum effective light range in world units
	DefaultIntensity    = 1.0
	DayNightCyclePeriod = 24.0 // Hours in a full day/night cycle
)

// Light represents a light source in the world.
type Light struct {
	// Type of light (point, directional, spot, ambient)
	Type string
	// Position in world coordinates (for point/spot lights)
	X, Y, Z float64
	// Direction for directional/spot lights (normalized)
	DirX, DirY, DirZ float64
	// Intensity multiplier (0.0-1.0+)
	Intensity float64
	// Color of the light
	Color color.RGBA
	// Maximum range for point/spot lights
	Range float64
	// Falloff exponent (1=linear, 2=quadratic)
	Falloff float64
	// Spot angle in radians (for spot lights)
	SpotAngle float64
	// Enabled flag
	Enabled bool
}

// NewPointLight creates a new point light source.
func NewPointLight(x, y, z, intensity float64, c color.RGBA) *Light {
	return &Light{
		Type:      TypePoint,
		X:         x,
		Y:         y,
		Z:         z,
		Intensity: intensity,
		Color:     c,
		Range:     DefaultMaxRange,
		Falloff:   DefaultFalloff,
		Enabled:   true,
	}
}

// NewDirectionalLight creates a new directional light (like sun/moon).
func NewDirectionalLight(dirX, dirY, dirZ, intensity float64, c color.RGBA) *Light {
	// Normalize direction
	len := math.Sqrt(dirX*dirX + dirY*dirY + dirZ*dirZ)
	if len > 0 {
		dirX /= len
		dirY /= len
		dirZ /= len
	}
	return &Light{
		Type:      TypeDirectional,
		DirX:      dirX,
		DirY:      dirY,
		DirZ:      dirZ,
		Intensity: intensity,
		Color:     c,
		Enabled:   true,
	}
}

// NewAmbientLight creates ambient lighting.
func NewAmbientLight(intensity float64, c color.RGBA) *Light {
	return &Light{
		Type:      TypeAmbient,
		Intensity: intensity,
		Color:     c,
		Enabled:   true,
	}
}

// CalculateAttenuation returns the light attenuation at a given distance.
func (l *Light) CalculateAttenuation(distance float64) float64 {
	if !l.Enabled || l.Type == TypeAmbient || l.Type == TypeDirectional {
		return 1.0
	}
	if distance <= 0 {
		return 1.0
	}
	if distance >= l.Range {
		return 0.0
	}
	// Smooth falloff
	normalizedDist := distance / l.Range
	return math.Pow(1.0-normalizedDist, l.Falloff) * l.Intensity
}

// GetColorAtDistance returns the light color attenuated by distance.
func (l *Light) GetColorAtDistance(distance float64) color.RGBA {
	atten := l.CalculateAttenuation(distance)
	if atten <= 0 {
		return color.RGBA{}
	}
	return color.RGBA{
		R: uint8(float64(l.Color.R) * atten),
		G: uint8(float64(l.Color.G) * atten),
		B: uint8(float64(l.Color.B) * atten),
		A: l.Color.A,
	}
}

// System manages lights and calculates lighting for the scene.
type System struct {
	lights       []*Light
	ambient      *Light
	sun          *Light
	timeOfDay    float64 // 0-24 hours
	indoorMode   bool
	genrePalette GenrePalette
}

// GenrePalette defines genre-specific lighting colors.
type GenrePalette struct {
	Sunlight   color.RGBA
	Moonlight  color.RGBA
	Torchlight color.RGBA
	Magic      color.RGBA
	Ambient    color.RGBA
}

// GenrePalettes maps genre to lighting palette.
var GenrePalettes = map[string]GenrePalette{
	"fantasy": {
		Sunlight:   color.RGBA{255, 250, 220, 255},
		Moonlight:  color.RGBA{150, 170, 200, 255},
		Torchlight: color.RGBA{255, 180, 80, 255},
		Magic:      color.RGBA{150, 180, 255, 255},
		Ambient:    color.RGBA{40, 40, 50, 255},
	},
	"sci-fi": {
		Sunlight:   color.RGBA{255, 255, 255, 255},
		Moonlight:  color.RGBA{180, 190, 210, 255},
		Torchlight: color.RGBA{200, 220, 255, 255}, // LED lights
		Magic:      color.RGBA{100, 200, 255, 255},
		Ambient:    color.RGBA{30, 40, 50, 255},
	},
	"horror": {
		Sunlight:   color.RGBA{220, 200, 180, 255},
		Moonlight:  color.RGBA{130, 140, 160, 255},
		Torchlight: color.RGBA{255, 160, 60, 255},
		Magic:      color.RGBA{180, 100, 180, 255},
		Ambient:    color.RGBA{20, 20, 25, 255},
	},
	"cyberpunk": {
		Sunlight:   color.RGBA{255, 240, 200, 255},
		Moonlight:  color.RGBA{100, 120, 180, 255},
		Torchlight: color.RGBA{255, 100, 200, 255}, // Neon
		Magic:      color.RGBA{100, 255, 255, 255},
		Ambient:    color.RGBA{30, 20, 40, 255},
	},
	"post-apocalyptic": {
		Sunlight:   color.RGBA{255, 220, 180, 255},
		Moonlight:  color.RGBA{180, 180, 160, 255},
		Torchlight: color.RGBA{255, 150, 50, 255},
		Magic:      color.RGBA{100, 255, 100, 255}, // Radiation
		Ambient:    color.RGBA{40, 35, 30, 255},
	},
}

// NewSystem creates a new lighting system.
func NewSystem(genre string) *System {
	palette, ok := GenrePalettes[genre]
	if !ok {
		palette = GenrePalettes["fantasy"]
	}

	s := &System{
		lights:       make([]*Light, 0, 32),
		timeOfDay:    12.0, // Noon
		indoorMode:   false,
		genrePalette: palette,
	}

	// Default ambient light
	s.ambient = NewAmbientLight(DefaultAmbient, palette.Ambient)

	// Default sun
	s.sun = NewDirectionalLight(0.5, -0.8, 0.2, 1.0, palette.Sunlight)

	return s
}

// SetGenre updates the lighting palette for a genre.
func (s *System) SetGenre(genre string) {
	palette, ok := GenrePalettes[genre]
	if ok {
		s.genrePalette = palette
		s.ambient.Color = palette.Ambient
		s.sun.Color = palette.Sunlight
	}
}

// SetTimeOfDay sets the time (0-24 hours).
func (s *System) SetTimeOfDay(hours float64) {
	s.timeOfDay = math.Mod(hours, DayNightCyclePeriod)
	if s.timeOfDay < 0 {
		s.timeOfDay += DayNightCyclePeriod
	}
	s.updateSunPosition()
}

// GetTimeOfDay returns the current time (0-24 hours).
func (s *System) GetTimeOfDay() float64 {
	return s.timeOfDay
}

// AdvanceTime moves time forward by dt hours.
func (s *System) AdvanceTime(dt float64) {
	s.SetTimeOfDay(s.timeOfDay + dt)
}

// updateSunPosition calculates sun direction and intensity based on time.
func (s *System) updateSunPosition() {
	// Convert time to angle where 12=noon (sun highest), 0=midnight (sun lowest)
	// angle: 0 at noon, π at midnight (below horizon)
	angle := (s.timeOfDay - 12) * math.Pi / 12

	// Sun direction (rises east, sets west)
	s.sun.DirX = math.Sin(angle)
	s.sun.DirY = math.Cos(angle) // +Y is down in screen space, so cos gives us height
	s.sun.DirZ = 0.1

	// Sun height: 1 at noon, -1 at midnight
	height := math.Cos(angle)

	if height < 0 {
		// Sun below horizon - use moonlight
		// Intensity fades as sun goes lower
		s.sun.Intensity = 0.3 * math.Max(0, 1+height)
		s.sun.Color = s.genrePalette.Moonlight
	} else {
		// Daytime - intensity based on height
		s.sun.Intensity = 0.5 + 0.5*height
		s.sun.Color = s.genrePalette.Sunlight
	}

	// Adjust ambient based on time
	if height < 0 {
		// Night - darker ambient
		s.ambient.Intensity = DefaultAmbient * 0.5
	} else {
		// Day - brighter ambient at noon
		s.ambient.Intensity = DefaultAmbient * (0.8 + 0.4*height)
	}
}

// SetIndoorMode enables/disables indoor lighting mode.
// In indoor mode, sun/moon light is reduced and ambient is lower.
func (s *System) SetIndoorMode(indoor bool) {
	s.indoorMode = indoor
}

// IsIndoor returns whether indoor mode is active.
func (s *System) IsIndoor() bool {
	return s.indoorMode
}

// AddLight adds a light to the system.
func (s *System) AddLight(l *Light) {
	if l != nil {
		s.lights = append(s.lights, l)
	}
}

// RemoveLight removes a light from the system.
func (s *System) RemoveLight(l *Light) {
	for i, light := range s.lights {
		if light == l {
			s.lights = append(s.lights[:i], s.lights[i+1:]...)
			return
		}
	}
}

// ClearLights removes all lights (except ambient and sun).
func (s *System) ClearLights() {
	s.lights = s.lights[:0]
}

// LightCount returns the number of point/spot lights.
func (s *System) LightCount() int {
	return len(s.lights)
}

// Lights returns the current lights.
func (s *System) Lights() []*Light {
	return s.lights
}

// GetAmbient returns the ambient light.
func (s *System) GetAmbient() *Light {
	return s.ambient
}

// GetSun returns the sun/moon directional light.
func (s *System) GetSun() *Light {
	return s.sun
}

// CalculateLightingAt returns the total lighting contribution at a world position.
// Returns an intensity multiplier (0.0-1.0+) and color.
func (s *System) CalculateLightingAt(x, y, z float64) (float64, color.RGBA) {
	var totalR, totalG, totalB float64
	var totalIntensity float64

	// Add ambient and sun contributions
	totalR, totalG, totalB, totalIntensity = s.addGlobalLighting(totalR, totalG, totalB, totalIntensity)

	// Add point/directional light contributions
	totalR, totalG, totalB, totalIntensity = s.addLocalLighting(x, y, z, totalR, totalG, totalB, totalIntensity)

	// Normalize and finalize
	return s.finalizeLighting(totalR, totalG, totalB, totalIntensity)
}

// addGlobalLighting adds ambient and sun/moon contributions.
func (s *System) addGlobalLighting(r, g, b, intensity float64) (float64, float64, float64, float64) {
	// Ambient contribution
	if s.ambient != nil && s.ambient.Enabled {
		amb := s.ambient.Intensity
		if s.indoorMode {
			amb *= 0.7 // Darker indoors
		}
		r += float64(s.ambient.Color.R) * amb
		g += float64(s.ambient.Color.G) * amb
		b += float64(s.ambient.Color.B) * amb
		intensity += amb
	}

	// Sun/moon contribution (reduced indoors)
	if s.sun != nil && s.sun.Enabled && !s.indoorMode {
		sunIntensity := s.sun.Intensity
		r += float64(s.sun.Color.R) * sunIntensity
		g += float64(s.sun.Color.G) * sunIntensity
		b += float64(s.sun.Color.B) * sunIntensity
		intensity += sunIntensity
	}

	return r, g, b, intensity
}

// addLocalLighting adds point and directional light contributions.
func (s *System) addLocalLighting(x, y, z, r, g, b, intensity float64) (float64, float64, float64, float64) {
	for _, l := range s.lights {
		if !l.Enabled {
			continue
		}

		switch l.Type {
		case TypePoint:
			r, g, b, intensity = s.addPointLight(l, x, y, z, r, g, b, intensity)
		case TypeDirectional:
			r += float64(l.Color.R) * l.Intensity
			g += float64(l.Color.G) * l.Intensity
			b += float64(l.Color.B) * l.Intensity
			intensity += l.Intensity
		}
	}
	return r, g, b, intensity
}

// addPointLight calculates and adds a point light's contribution.
func (s *System) addPointLight(l *Light, x, y, z, r, g, b, intensity float64) (float64, float64, float64, float64) {
	dx := x - l.X
	dy := y - l.Y
	dz := z - l.Z
	dist := math.Sqrt(dx*dx + dy*dy + dz*dz)
	atten := l.CalculateAttenuation(dist)
	if atten > 0 {
		r += float64(l.Color.R) * atten
		g += float64(l.Color.G) * atten
		b += float64(l.Color.B) * atten
		intensity += atten
	}
	return r, g, b, intensity
}

// finalizeLighting normalizes colors and clamps values.
func (s *System) finalizeLighting(r, g, b, intensity float64) (float64, color.RGBA) {
	// Normalize color by intensity
	if intensity > 0 {
		r /= intensity
		g /= intensity
		b /= intensity
	}

	// Clamp values
	r = clampFloat(r, 0, 255)
	g = clampFloat(g, 0, 255)
	b = clampFloat(b, 0, 255)
	intensity = clampFloat(intensity, 0, 2.0) // Cap at 2x brightness

	return intensity, color.RGBA{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
		A: 255,
	}
}

// clampFloat clamps a value between min and max.
func clampFloat(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// ApplyLighting modifies a pixel color based on lighting at a world position.
func (s *System) ApplyLighting(c color.RGBA, worldX, worldY, worldZ float64) color.RGBA {
	intensity, lightColor := s.CalculateLightingAt(worldX, worldY, worldZ)

	// Multiply pixel color by light color and intensity
	r := float64(c.R) * float64(lightColor.R) / 255.0 * intensity
	g := float64(c.G) * float64(lightColor.G) / 255.0 * intensity
	b := float64(c.B) * float64(lightColor.B) / 255.0 * intensity

	// Clamp
	if r > 255 {
		r = 255
	}
	if g > 255 {
		g = 255
	}
	if b > 255 {
		b = 255
	}

	return color.RGBA{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
		A: c.A,
	}
}

// IsDaytime returns true if the sun is up.
func (s *System) IsDaytime() bool {
	return s.timeOfDay >= 6 && s.timeOfDay < 18
}

// IsNight returns true if it's nighttime (6PM-6AM).
func (s *System) IsNight() bool {
	return !s.IsDaytime()
}

// CreateTorch creates a torch-colored point light.
func (s *System) CreateTorch(x, y, z float64) *Light {
	return NewPointLight(x, y, z, 0.8, s.genrePalette.Torchlight)
}

// CreateMagicLight creates a magic-colored point light.
func (s *System) CreateMagicLight(x, y, z, intensity float64) *Light {
	return NewPointLight(x, y, z, intensity, s.genrePalette.Magic)
}
