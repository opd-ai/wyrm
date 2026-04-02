// Package raycast provides first-person raycasting rendering.
package raycast

import (
	"image/color"
	"math"

	"github.com/opd-ai/wyrm/pkg/rendering/texture"
)

// NormalLighting handles normal map sampling and lighting calculations.
type NormalLighting struct {
	// SunDirection is the normalized sun direction vector (x, y, z).
	// Z points out of the wall surface (towards viewer when positive).
	SunDirection [3]float64

	// SunIntensity is the intensity of sunlight (0.0-1.0).
	SunIntensity float64

	// AmbientLight is the ambient light level (0.0-1.0).
	AmbientLight float64

	// SunColor is the color of sunlight (for specular highlights).
	SunColor color.RGBA

	// SpecularEnabled controls whether specular highlights are calculated.
	SpecularEnabled bool
}

// DefaultNormalLighting creates a NormalLighting with default settings.
func DefaultNormalLighting() *NormalLighting {
	return &NormalLighting{
		SunDirection:    [3]float64{-0.5, -0.7, 0.5}, // Angled from upper-left
		SunIntensity:    0.8,
		AmbientLight:    0.3,
		SunColor:        color.RGBA{R: 255, G: 248, B: 220, A: 255}, // Warm white
		SpecularEnabled: true,
	}
}

// Normalize normalizes a 3D vector in place and returns it.
func normalizeVec3(v *[3]float64) {
	length := math.Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2])
	if length > 0.0001 {
		v[0] /= length
		v[1] /= length
		v[2] /= length
	}
}

// SampleNormalMap samples the normal map at the given texture coordinates.
// Returns the surface normal in tangent space (x, y, z where z points outward).
func SampleNormalMap(tex *texture.Texture, texX, texY float64) [3]float64 {
	if tex == nil || !tex.HasNormalMap() {
		// Return flat normal (pointing outward)
		return [3]float64{0.0, 0.0, 1.0}
	}

	// Convert to pixel coordinates
	tx := int(texX*float64(tex.Width)) % tex.Width
	ty := int(texY*float64(tex.Height)) % tex.Height
	if tx < 0 {
		tx += tex.Width
	}
	if ty < 0 {
		ty += tex.Height
	}

	normal := tex.GetNormalAt(tx, ty)
	return [3]float64{normal.X, normal.Y, normal.Z}
}

// TransformNormalToWorld transforms a tangent-space normal to world space.
// side: 0 = X-facing wall (East-West), 1 = Y-facing wall (North-South)
func TransformNormalToWorld(tangentNormal [3]float64, side int) [3]float64 {
	// Tangent space to world space transformation
	// In tangent space: X = horizontal along wall, Y = vertical, Z = outward from wall
	// For X-facing walls (side 0): outward is along X axis
	// For Y-facing walls (side 1): outward is along Y axis
	worldNormal := [3]float64{}

	if side == 0 {
		// X-facing wall: tangent Z becomes world X
		worldNormal[0] = tangentNormal[2] // outward (Z) -> world X
		worldNormal[1] = tangentNormal[0] // horizontal (X) -> world Y
		worldNormal[2] = tangentNormal[1] // vertical (Y) -> world Z
	} else {
		// Y-facing wall: tangent Z becomes world Y
		worldNormal[0] = tangentNormal[0] // horizontal (X) -> world X
		worldNormal[1] = tangentNormal[2] // outward (Z) -> world Y
		worldNormal[2] = tangentNormal[1] // vertical (Y) -> world Z
	}

	normalizeVec3(&worldNormal)
	return worldNormal
}

// ComputeLightIntensity calculates the light intensity at a surface point.
// Returns a value from 0.0 (fully shadowed) to 1.0 (fully lit).
func (nl *NormalLighting) ComputeLightIntensity(worldNormal [3]float64) float64 {
	// Normalize sun direction
	sunDir := nl.SunDirection
	normalizeVec3(&sunDir)

	// Compute dot product (Lambertian diffuse lighting)
	// Light comes FROM sunDir, so we need -sunDir to get direction TO light
	dot := -sunDir[0]*worldNormal[0] - sunDir[1]*worldNormal[1] - sunDir[2]*worldNormal[2]

	// Clamp to [0, 1] - surfaces facing away from light get no direct light
	if dot < 0 {
		dot = 0
	}

	// Combine ambient and diffuse lighting
	intensity := nl.AmbientLight + dot*nl.SunIntensity

	// Clamp to [0, 1]
	if intensity > 1.0 {
		intensity = 1.0
	}

	return intensity
}

// ApplyNormalMapLighting applies normal map-based lighting to a wall pixel color.
// tex: the wall texture (may have normal map)
// baseColor: the sampled albedo color
// texX, texY: texture coordinates (0-1)
// side: wall orientation (0 = X-facing, 1 = Y-facing)
func (nl *NormalLighting) ApplyNormalMapLighting(
	tex *texture.Texture,
	baseColor color.RGBA,
	texX, texY float64,
	side int,
) color.RGBA {
	// Sample normal map
	tangentNormal := SampleNormalMap(tex, texX, texY)

	// Transform to world space
	worldNormal := TransformNormalToWorld(tangentNormal, side)

	// Compute light intensity
	intensity := nl.ComputeLightIntensity(worldNormal)

	// Apply intensity to color
	result := applyLightIntensity(baseColor, intensity)

	// Apply specular highlights if enabled (requires view direction - use default towards camera)
	// For simplified implementation, view direction is assumed to be (0, 0, 1) (looking at wall)
	if nl.SpecularEnabled {
		// Default material properties - could be extended to use MaterialRegistry
		reflectivity := 0.2 // Mild reflectivity
		roughness := 0.5    // Moderate roughness
		result = nl.ApplySpecularHighlight(result, worldNormal, reflectivity, roughness)
	}

	return result
}

// ApplySpecularHighlight adds specular highlight to a color.
// Uses simplified Blinn-Phong model with fixed view direction.
// reflectivity: 0.0-1.0, controls specular intensity
// roughness: 0.0-1.0, controls specular sharpness (lower = sharper)
func (nl *NormalLighting) ApplySpecularHighlight(
	baseColor color.RGBA,
	worldNormal [3]float64,
	reflectivity, roughness float64,
) color.RGBA {
	if reflectivity <= 0 {
		return baseColor
	}

	// View direction (camera looking into wall)
	viewDir := [3]float64{0, 0, 1}

	// Light direction (towards light source)
	lightDir := [3]float64{-nl.SunDirection[0], -nl.SunDirection[1], -nl.SunDirection[2]}
	normalizeVec3(&lightDir)

	// Calculate reflection vector: R = 2 * dot(N, L) * N - L
	dotNL := worldNormal[0]*lightDir[0] + worldNormal[1]*lightDir[1] + worldNormal[2]*lightDir[2]
	if dotNL < 0 {
		return baseColor // Surface facing away from light
	}

	reflectVec := [3]float64{
		2*dotNL*worldNormal[0] - lightDir[0],
		2*dotNL*worldNormal[1] - lightDir[1],
		2*dotNL*worldNormal[2] - lightDir[2],
	}
	normalizeVec3(&reflectVec)

	// Calculate specular intensity: spec = pow(max(dot(R, V), 0), shininess)
	dotRV := reflectVec[0]*viewDir[0] + reflectVec[1]*viewDir[1] + reflectVec[2]*viewDir[2]
	if dotRV < 0 {
		dotRV = 0
	}

	// shininess = (1.0 - Roughness) * 64.0
	shininess := (1.0 - roughness) * 64.0
	if shininess < 1 {
		shininess = 1
	}

	specIntensity := math.Pow(dotRV, shininess)

	// Add spec * Reflectivity * lightColor to final color
	specR := specIntensity * reflectivity * float64(nl.SunColor.R) / 255.0
	specG := specIntensity * reflectivity * float64(nl.SunColor.G) / 255.0
	specB := specIntensity * reflectivity * float64(nl.SunColor.B) / 255.0

	newR := float64(baseColor.R) + specR*255.0
	newG := float64(baseColor.G) + specG*255.0
	newB := float64(baseColor.B) + specB*255.0

	// Clamp to 255
	if newR > 255 {
		newR = 255
	}
	if newG > 255 {
		newG = 255
	}
	if newB > 255 {
		newB = 255
	}

	return color.RGBA{
		R: uint8(newR),
		G: uint8(newG),
		B: uint8(newB),
		A: baseColor.A,
	}
}

// applyLightIntensity multiplies a color by an intensity factor.
func applyLightIntensity(c color.RGBA, intensity float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(c.R) * intensity),
		G: uint8(float64(c.G) * intensity),
		B: uint8(float64(c.B) * intensity),
		A: c.A, // Preserve alpha
	}
}

// SetSunAngle sets the sun direction from horizontal and vertical angles.
// horizontal: 0 = East, π/2 = North, π = West, 3π/2 = South
// vertical: 0 = horizon, π/2 = directly overhead
func (nl *NormalLighting) SetSunAngle(horizontal, vertical float64) {
	nl.SunDirection[0] = math.Cos(horizontal) * math.Cos(vertical)
	nl.SunDirection[1] = math.Sin(horizontal) * math.Cos(vertical)
	nl.SunDirection[2] = math.Sin(vertical)
	normalizeVec3(&nl.SunDirection)
}

// SetTimeOfDay sets sun position based on time (0-24 hours).
// Assumes sun rises at 6:00 (angle 0), peaks at 12:00 (angle π/2), sets at 18:00 (angle π).
func (nl *NormalLighting) SetTimeOfDay(hour float64) {
	// Normalize hour to 0-24 range
	for hour < 0 {
		hour += 24
	}
	for hour >= 24 {
		hour -= 24
	}

	// Calculate sun angle (6 AM = 0°, 12 PM = 90°, 6 PM = 180°)
	// Night hours (18-6) keep sun below horizon
	if hour >= 6 && hour <= 18 {
		sunProgress := (hour - 6) / 12.0 // 0 at 6AM, 1 at 6PM
		vertical := math.Sin(sunProgress*math.Pi) * (math.Pi / 2.2)
		horizontal := sunProgress * math.Pi // East to West
		nl.SetSunAngle(horizontal, vertical)
		nl.SunIntensity = 0.5 + math.Sin(sunProgress*math.Pi)*0.5
	} else {
		// Night time - minimal lighting
		nl.SunDirection = [3]float64{0, 0, -1}
		nl.SunIntensity = 0.1
	}
}
