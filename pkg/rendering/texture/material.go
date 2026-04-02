// Package texture provides procedural texture generation.
package texture

import (
	"image/color"
)

// ============================================================
// Material System
// ============================================================

// MaterialID identifies a material type.
type MaterialID uint16

// Material IDs for common material types.
const (
	MaterialNone MaterialID = iota
	MaterialStone
	MaterialWood
	MaterialMetal
	MaterialGlass
	MaterialConcrete
	MaterialBrick
	MaterialDirt
	MaterialGrass
	MaterialWater
	MaterialPlastic
	MaterialFabric
	MaterialLeather
	MaterialBone
	MaterialIce
	MaterialLava
	MaterialRust
	MaterialChrome
	MaterialNeon
	MaterialOrganic
	MaterialCrystal
	// Add more as needed
	MaterialCustom MaterialID = 1000 // Start of custom materials
)

// PhysicalProperties defines the physical characteristics of a material.
type PhysicalProperties struct {
	// Hardness affects damage resistance and sound when hit (0.0-1.0)
	Hardness float64

	// Density affects weight calculations (kg/m³, normalized to 0.0-1.0)
	Density float64

	// Friction affects movement speed on surface (0.0=ice, 1.0=rubber)
	Friction float64

	// Elasticity affects bounce/restitution (0.0=absorbs, 1.0=bounces)
	Elasticity float64

	// Conductivity affects heat/electricity transfer (0.0-1.0)
	Conductivity float64

	// Flammability affects fire spread (0.0=fireproof, 1.0=highly flammable)
	Flammability float64

	// Brittleness affects how it breaks (0.0=bends, 1.0=shatters)
	Brittleness float64
}

// VisualProperties defines rendering characteristics of a material.
type VisualProperties struct {
	// Roughness affects specular reflection (0.0=mirror, 1.0=matte)
	Roughness float64

	// Metalness affects metallic appearance (0.0=dielectric, 1.0=metal)
	Metalness float64

	// Transparency affects alpha rendering (0.0=opaque, 1.0=fully transparent)
	Transparency float64

	// Emissive intensity for glowing materials (0.0=none, 1.0=full glow)
	Emissive float64

	// Reflectivity for environment mapping (0.0-1.0)
	Reflectivity float64

	// Refraction index for transparent materials (1.0=air, 1.5=glass, 2.4=diamond)
	Refraction float64

	// Subsurface scattering amount for translucent materials (0.0-1.0)
	Subsurface float64
}

// AcousticProperties defines sound characteristics of a material.
type AcousticProperties struct {
	// ImpactSound is the sound category when struck ("metal", "wood", "stone", etc.)
	ImpactSound string

	// FootstepSound is the sound category when walked on
	FootstepSound string

	// Resonance affects how long sound rings (0.0-1.0)
	Resonance float64

	// SoundAbsorption affects echo/reverb (0.0=echoes, 1.0=absorbs)
	SoundAbsorption float64
}

// Material defines a complete material type with all its properties.
type Material struct {
	ID       MaterialID
	Name     string
	Category string // "natural", "synthetic", "organic", "metal", "mineral"

	Physical PhysicalProperties
	Visual   VisualProperties
	Acoustic AcousticProperties

	// BaseColors are the primary colors for texture generation
	BaseColors []color.RGBA

	// GenreVariants maps genre to color palette overrides
	GenreVariants map[string][]color.RGBA
}

// MaterialRegistry stores all registered materials.
type MaterialRegistry struct {
	materials map[MaterialID]*Material
	byName    map[string]MaterialID
}

// NewMaterialRegistry creates a new registry with default materials.
func NewMaterialRegistry() *MaterialRegistry {
	r := &MaterialRegistry{
		materials: make(map[MaterialID]*Material),
		byName:    make(map[string]MaterialID),
	}
	r.registerDefaultMaterials()
	return r
}

// Register adds a material to the registry.
func (r *MaterialRegistry) Register(m *Material) {
	r.materials[m.ID] = m
	r.byName[m.Name] = m.ID
}

// Get returns a material by ID, or nil if not found.
func (r *MaterialRegistry) Get(id MaterialID) *Material {
	return r.materials[id]
}

// GetByName returns a material by name, or nil if not found.
func (r *MaterialRegistry) GetByName(name string) *Material {
	if id, ok := r.byName[name]; ok {
		return r.materials[id]
	}
	return nil
}

// GetID returns the MaterialID for a given name, or MaterialNone if not found.
func (r *MaterialRegistry) GetID(name string) MaterialID {
	if id, ok := r.byName[name]; ok {
		return id
	}
	return MaterialNone
}

// List returns all registered material IDs.
func (r *MaterialRegistry) List() []MaterialID {
	ids := make([]MaterialID, 0, len(r.materials))
	for id := range r.materials {
		ids = append(ids, id)
	}
	return ids
}

// Count returns the number of registered materials.
func (r *MaterialRegistry) Count() int {
	return len(r.materials)
}

// GetColorsForGenre returns the appropriate colors for a material and genre.
func (r *MaterialRegistry) GetColorsForGenre(id MaterialID, genre string) []color.RGBA {
	m := r.Get(id)
	if m == nil {
		return nil
	}
	if variants, ok := m.GenreVariants[genre]; ok && len(variants) > 0 {
		return variants
	}
	return m.BaseColors
}

// registerDefaultMaterials adds all built-in material definitions.
func (r *MaterialRegistry) registerDefaultMaterials() {
	// Stone - common building material
	r.Register(&Material{
		ID:       MaterialStone,
		Name:     "stone",
		Category: "mineral",
		Physical: PhysicalProperties{
			Hardness:     0.8,
			Density:      0.65,
			Friction:     0.7,
			Elasticity:   0.05,
			Conductivity: 0.1,
			Flammability: 0.0,
			Brittleness:  0.6,
		},
		Visual: VisualProperties{
			Roughness:    0.9,
			Metalness:    0.0,
			Transparency: 0.0,
			Emissive:     0.0,
			Reflectivity: 0.05,
			Refraction:   1.0,
			Subsurface:   0.0,
		},
		Acoustic: AcousticProperties{
			ImpactSound:     "stone",
			FootstepSound:   "stone",
			Resonance:       0.2,
			SoundAbsorption: 0.3,
		},
		BaseColors: []color.RGBA{
			{R: 0x80, G: 0x80, B: 0x80, A: 255},
			{R: 0x70, G: 0x70, B: 0x70, A: 255},
			{R: 0x90, G: 0x90, B: 0x90, A: 255},
			{R: 0x75, G: 0x75, B: 0x75, A: 255},
		},
		GenreVariants: map[string][]color.RGBA{
			"fantasy": {
				{R: 0x8B, G: 0x83, B: 0x78, A: 255},
				{R: 0x7B, G: 0x73, B: 0x68, A: 255},
			},
			"horror": {
				{R: 0x50, G: 0x50, B: 0x50, A: 255},
				{R: 0x40, G: 0x42, B: 0x44, A: 255},
			},
		},
	})

	// Wood - organic building material
	r.Register(&Material{
		ID:       MaterialWood,
		Name:     "wood",
		Category: "organic",
		Physical: PhysicalProperties{
			Hardness:     0.5,
			Density:      0.4,
			Friction:     0.6,
			Elasticity:   0.15,
			Conductivity: 0.05,
			Flammability: 0.8,
			Brittleness:  0.3,
		},
		Visual: VisualProperties{
			Roughness:    0.7,
			Metalness:    0.0,
			Transparency: 0.0,
			Emissive:     0.0,
			Reflectivity: 0.1,
			Refraction:   1.0,
			Subsurface:   0.1,
		},
		Acoustic: AcousticProperties{
			ImpactSound:     "wood",
			FootstepSound:   "wood",
			Resonance:       0.5,
			SoundAbsorption: 0.5,
		},
		BaseColors: []color.RGBA{
			{R: 0x8B, G: 0x5A, B: 0x2B, A: 255},
			{R: 0x7B, G: 0x4A, B: 0x1B, A: 255},
			{R: 0x9B, G: 0x6A, B: 0x3B, A: 255},
			{R: 0x6B, G: 0x4A, B: 0x2B, A: 255},
		},
		GenreVariants: map[string][]color.RGBA{
			"cyberpunk": {
				{R: 0x4A, G: 0x3A, B: 0x2A, A: 255}, // Darker, treated
			},
			"post-apocalyptic": {
				{R: 0x5A, G: 0x4A, B: 0x3A, A: 255}, // Weathered
				{R: 0x50, G: 0x40, B: 0x30, A: 255},
			},
		},
	})

	// Metal - generic metal
	r.Register(&Material{
		ID:       MaterialMetal,
		Name:     "metal",
		Category: "metal",
		Physical: PhysicalProperties{
			Hardness:     0.9,
			Density:      0.8,
			Friction:     0.4,
			Elasticity:   0.3,
			Conductivity: 0.9,
			Flammability: 0.0,
			Brittleness:  0.4,
		},
		Visual: VisualProperties{
			Roughness:    0.3,
			Metalness:    1.0,
			Transparency: 0.0,
			Emissive:     0.0,
			Reflectivity: 0.7,
			Refraction:   1.0,
			Subsurface:   0.0,
		},
		Acoustic: AcousticProperties{
			ImpactSound:     "metal",
			FootstepSound:   "metal",
			Resonance:       0.8,
			SoundAbsorption: 0.1,
		},
		BaseColors: []color.RGBA{
			{R: 0xA0, G: 0xA0, B: 0xA0, A: 255},
			{R: 0x90, G: 0x90, B: 0x95, A: 255},
			{R: 0xB0, G: 0xB0, B: 0xB0, A: 255},
		},
		GenreVariants: map[string][]color.RGBA{
			"sci-fi": {
				{R: 0xC0, G: 0xC0, B: 0xC8, A: 255},
				{R: 0xD0, G: 0xD0, B: 0xD8, A: 255},
			},
		},
	})

	// Glass - transparent material
	r.Register(&Material{
		ID:       MaterialGlass,
		Name:     "glass",
		Category: "mineral",
		Physical: PhysicalProperties{
			Hardness:     0.6,
			Density:      0.5,
			Friction:     0.3,
			Elasticity:   0.0,
			Conductivity: 0.3,
			Flammability: 0.0,
			Brittleness:  1.0,
		},
		Visual: VisualProperties{
			Roughness:    0.0,
			Metalness:    0.0,
			Transparency: 0.9,
			Emissive:     0.0,
			Reflectivity: 0.5,
			Refraction:   1.5,
			Subsurface:   0.0,
		},
		Acoustic: AcousticProperties{
			ImpactSound:     "glass",
			FootstepSound:   "glass",
			Resonance:       0.3,
			SoundAbsorption: 0.05,
		},
		BaseColors: []color.RGBA{
			{R: 0xE0, G: 0xE8, B: 0xF0, A: 128},
			{R: 0xD0, G: 0xD8, B: 0xE8, A: 128},
		},
	})

	// Concrete - modern building material
	r.Register(&Material{
		ID:       MaterialConcrete,
		Name:     "concrete",
		Category: "mineral",
		Physical: PhysicalProperties{
			Hardness:     0.85,
			Density:      0.7,
			Friction:     0.75,
			Elasticity:   0.02,
			Conductivity: 0.2,
			Flammability: 0.0,
			Brittleness:  0.7,
		},
		Visual: VisualProperties{
			Roughness:    0.95,
			Metalness:    0.0,
			Transparency: 0.0,
			Emissive:     0.0,
			Reflectivity: 0.02,
			Refraction:   1.0,
			Subsurface:   0.0,
		},
		Acoustic: AcousticProperties{
			ImpactSound:     "stone",
			FootstepSound:   "stone",
			Resonance:       0.15,
			SoundAbsorption: 0.4,
		},
		BaseColors: []color.RGBA{
			{R: 0x90, G: 0x90, B: 0x90, A: 255},
			{R: 0x85, G: 0x85, B: 0x85, A: 255},
			{R: 0x9A, G: 0x9A, B: 0x9A, A: 255},
		},
		GenreVariants: map[string][]color.RGBA{
			"cyberpunk": {
				{R: 0x70, G: 0x70, B: 0x75, A: 255},
				{R: 0x60, G: 0x60, B: 0x68, A: 255},
			},
		},
	})

	// Brick - construction material
	r.Register(&Material{
		ID:       MaterialBrick,
		Name:     "brick",
		Category: "mineral",
		Physical: PhysicalProperties{
			Hardness:     0.75,
			Density:      0.55,
			Friction:     0.8,
			Elasticity:   0.05,
			Conductivity: 0.15,
			Flammability: 0.0,
			Brittleness:  0.65,
		},
		Visual: VisualProperties{
			Roughness:    0.85,
			Metalness:    0.0,
			Transparency: 0.0,
			Emissive:     0.0,
			Reflectivity: 0.05,
			Refraction:   1.0,
			Subsurface:   0.0,
		},
		Acoustic: AcousticProperties{
			ImpactSound:     "stone",
			FootstepSound:   "stone",
			Resonance:       0.25,
			SoundAbsorption: 0.35,
		},
		BaseColors: []color.RGBA{
			{R: 0xB2, G: 0x4C, B: 0x3C, A: 255},
			{R: 0xA2, G: 0x3C, B: 0x2C, A: 255},
			{R: 0xC2, G: 0x5C, B: 0x4C, A: 255},
		},
	})

	// Dirt - natural ground material
	r.Register(&Material{
		ID:       MaterialDirt,
		Name:     "dirt",
		Category: "natural",
		Physical: PhysicalProperties{
			Hardness:     0.2,
			Density:      0.35,
			Friction:     0.6,
			Elasticity:   0.1,
			Conductivity: 0.05,
			Flammability: 0.0,
			Brittleness:  0.1,
		},
		Visual: VisualProperties{
			Roughness:    1.0,
			Metalness:    0.0,
			Transparency: 0.0,
			Emissive:     0.0,
			Reflectivity: 0.01,
			Refraction:   1.0,
			Subsurface:   0.0,
		},
		Acoustic: AcousticProperties{
			ImpactSound:     "dirt",
			FootstepSound:   "dirt",
			Resonance:       0.0,
			SoundAbsorption: 0.9,
		},
		BaseColors: []color.RGBA{
			{R: 0x6B, G: 0x4E, B: 0x31, A: 255},
			{R: 0x5B, G: 0x3E, B: 0x21, A: 255},
			{R: 0x7B, G: 0x5E, B: 0x41, A: 255},
		},
	})

	// Grass - natural ground cover
	r.Register(&Material{
		ID:       MaterialGrass,
		Name:     "grass",
		Category: "organic",
		Physical: PhysicalProperties{
			Hardness:     0.1,
			Density:      0.15,
			Friction:     0.5,
			Elasticity:   0.2,
			Conductivity: 0.02,
			Flammability: 0.6,
			Brittleness:  0.0,
		},
		Visual: VisualProperties{
			Roughness:    0.8,
			Metalness:    0.0,
			Transparency: 0.0,
			Emissive:     0.0,
			Reflectivity: 0.05,
			Refraction:   1.0,
			Subsurface:   0.2,
		},
		Acoustic: AcousticProperties{
			ImpactSound:     "grass",
			FootstepSound:   "grass",
			Resonance:       0.0,
			SoundAbsorption: 0.95,
		},
		BaseColors: []color.RGBA{
			{R: 0x4A, G: 0x7C, B: 0x23, A: 255},
			{R: 0x3A, G: 0x6C, B: 0x13, A: 255},
			{R: 0x5A, G: 0x8C, B: 0x33, A: 255},
		},
		GenreVariants: map[string][]color.RGBA{
			"horror": {
				{R: 0x3A, G: 0x4C, B: 0x23, A: 255},
				{R: 0x2A, G: 0x3C, B: 0x13, A: 255},
			},
			"post-apocalyptic": {
				{R: 0x5A, G: 0x5C, B: 0x23, A: 255},
				{R: 0x4A, G: 0x4C, B: 0x13, A: 255},
			},
		},
	})

	// Rust - corroded metal
	r.Register(&Material{
		ID:       MaterialRust,
		Name:     "rust",
		Category: "metal",
		Physical: PhysicalProperties{
			Hardness:     0.4,
			Density:      0.6,
			Friction:     0.8,
			Elasticity:   0.05,
			Conductivity: 0.3,
			Flammability: 0.0,
			Brittleness:  0.7,
		},
		Visual: VisualProperties{
			Roughness:    0.95,
			Metalness:    0.6,
			Transparency: 0.0,
			Emissive:     0.0,
			Reflectivity: 0.1,
			Refraction:   1.0,
			Subsurface:   0.0,
		},
		Acoustic: AcousticProperties{
			ImpactSound:     "metal",
			FootstepSound:   "metal",
			Resonance:       0.3,
			SoundAbsorption: 0.4,
		},
		BaseColors: []color.RGBA{
			{R: 0xB7, G: 0x41, B: 0x0E, A: 255},
			{R: 0xA7, G: 0x31, B: 0x0E, A: 255},
			{R: 0x87, G: 0x51, B: 0x1E, A: 255},
		},
	})

	// Chrome - polished metal
	r.Register(&Material{
		ID:       MaterialChrome,
		Name:     "chrome",
		Category: "metal",
		Physical: PhysicalProperties{
			Hardness:     0.95,
			Density:      0.85,
			Friction:     0.2,
			Elasticity:   0.35,
			Conductivity: 0.95,
			Flammability: 0.0,
			Brittleness:  0.3,
		},
		Visual: VisualProperties{
			Roughness:    0.05,
			Metalness:    1.0,
			Transparency: 0.0,
			Emissive:     0.0,
			Reflectivity: 0.95,
			Refraction:   1.0,
			Subsurface:   0.0,
		},
		Acoustic: AcousticProperties{
			ImpactSound:     "metal",
			FootstepSound:   "metal",
			Resonance:       0.9,
			SoundAbsorption: 0.05,
		},
		BaseColors: []color.RGBA{
			{R: 0xE0, G: 0xE0, B: 0xE8, A: 255},
			{R: 0xD0, G: 0xD0, B: 0xD8, A: 255},
			{R: 0xF0, G: 0xF0, B: 0xF8, A: 255},
		},
	})

	// Neon - glowing material
	r.Register(&Material{
		ID:       MaterialNeon,
		Name:     "neon",
		Category: "synthetic",
		Physical: PhysicalProperties{
			Hardness:     0.6,
			Density:      0.3,
			Friction:     0.3,
			Elasticity:   0.0,
			Conductivity: 0.7,
			Flammability: 0.0,
			Brittleness:  0.9,
		},
		Visual: VisualProperties{
			Roughness:    0.1,
			Metalness:    0.0,
			Transparency: 0.3,
			Emissive:     0.9,
			Reflectivity: 0.2,
			Refraction:   1.3,
			Subsurface:   0.5,
		},
		Acoustic: AcousticProperties{
			ImpactSound:     "glass",
			FootstepSound:   "glass",
			Resonance:       0.4,
			SoundAbsorption: 0.1,
		},
		BaseColors: []color.RGBA{
			{R: 0xFF, G: 0x00, B: 0xFF, A: 255},
			{R: 0x00, G: 0xFF, B: 0xFF, A: 255},
			{R: 0xFF, G: 0xFF, B: 0x00, A: 255},
		},
	})

	// Ice - frozen water
	r.Register(&Material{
		ID:       MaterialIce,
		Name:     "ice",
		Category: "natural",
		Physical: PhysicalProperties{
			Hardness:     0.4,
			Density:      0.35,
			Friction:     0.05,
			Elasticity:   0.0,
			Conductivity: 0.5,
			Flammability: 0.0,
			Brittleness:  0.85,
		},
		Visual: VisualProperties{
			Roughness:    0.2,
			Metalness:    0.0,
			Transparency: 0.7,
			Emissive:     0.0,
			Reflectivity: 0.3,
			Refraction:   1.31,
			Subsurface:   0.4,
		},
		Acoustic: AcousticProperties{
			ImpactSound:     "ice",
			FootstepSound:   "ice",
			Resonance:       0.2,
			SoundAbsorption: 0.15,
		},
		BaseColors: []color.RGBA{
			{R: 0xD0, G: 0xE8, B: 0xF0, A: 200},
			{R: 0xC0, G: 0xD8, B: 0xE8, A: 200},
			{R: 0xE0, G: 0xF0, B: 0xF8, A: 200},
		},
	})

	// Organic - flesh/plant material
	r.Register(&Material{
		ID:       MaterialOrganic,
		Name:     "organic",
		Category: "organic",
		Physical: PhysicalProperties{
			Hardness:     0.3,
			Density:      0.4,
			Friction:     0.6,
			Elasticity:   0.4,
			Conductivity: 0.3,
			Flammability: 0.7,
			Brittleness:  0.2,
		},
		Visual: VisualProperties{
			Roughness:    0.6,
			Metalness:    0.0,
			Transparency: 0.0,
			Emissive:     0.0,
			Reflectivity: 0.1,
			Refraction:   1.0,
			Subsurface:   0.7,
		},
		Acoustic: AcousticProperties{
			ImpactSound:     "flesh",
			FootstepSound:   "wet",
			Resonance:       0.1,
			SoundAbsorption: 0.8,
		},
		BaseColors: []color.RGBA{
			{R: 0x8B, G: 0x6B, B: 0x5B, A: 255},
			{R: 0x7B, G: 0x5B, B: 0x4B, A: 255},
		},
		GenreVariants: map[string][]color.RGBA{
			"horror": {
				{R: 0x6B, G: 0x4B, B: 0x4B, A: 255},
				{R: 0x8B, G: 0x3B, B: 0x3B, A: 255},
			},
		},
	})

	// Crystal - magical/precious material
	r.Register(&Material{
		ID:       MaterialCrystal,
		Name:     "crystal",
		Category: "mineral",
		Physical: PhysicalProperties{
			Hardness:     0.85,
			Density:      0.55,
			Friction:     0.2,
			Elasticity:   0.0,
			Conductivity: 0.6,
			Flammability: 0.0,
			Brittleness:  0.95,
		},
		Visual: VisualProperties{
			Roughness:    0.05,
			Metalness:    0.0,
			Transparency: 0.85,
			Emissive:     0.2,
			Reflectivity: 0.8,
			Refraction:   2.0,
			Subsurface:   0.3,
		},
		Acoustic: AcousticProperties{
			ImpactSound:     "crystal",
			FootstepSound:   "glass",
			Resonance:       0.95,
			SoundAbsorption: 0.02,
		},
		BaseColors: []color.RGBA{
			{R: 0xE0, G: 0xD0, B: 0xF0, A: 180},
			{R: 0xD0, G: 0xE0, B: 0xF0, A: 180},
			{R: 0xF0, G: 0xD0, B: 0xE0, A: 180},
		},
		GenreVariants: map[string][]color.RGBA{
			"fantasy": {
				{R: 0xD0, G: 0xA0, B: 0xF0, A: 180},
				{R: 0xA0, G: 0xD0, B: 0xF0, A: 180},
			},
			"sci-fi": {
				{R: 0xA0, G: 0xF0, B: 0xD0, A: 180},
				{R: 0xF0, G: 0xA0, B: 0xA0, A: 180},
			},
		},
	})
}

// DefaultMaterialRegistry is the global material registry instance.
var DefaultMaterialRegistry = NewMaterialRegistry()

// ============================================================
// Genre-Specific Material Palettes
// ============================================================

// GenrePalette defines a complete color scheme for a genre.
type GenreColorScheme struct {
	// Primary colors for the genre
	Primary []color.RGBA
	// Accent colors (highlights, details)
	Accent []color.RGBA
	// Ambient tint applied to all colors
	AmbientTint color.RGBA
	// Saturation modifier (1.0 = normal, <1 = desaturated, >1 = vivid)
	Saturation float64
	// Brightness modifier (1.0 = normal)
	Brightness float64
	// Contrast modifier (1.0 = normal)
	Contrast float64
}

// GenrePalettes contains predefined palettes for each genre.
var GenreColorSchemes = map[string]*GenreColorScheme{
	"fantasy": {
		Primary: []color.RGBA{
			{R: 0x8B, G: 0x73, B: 0x55, A: 255}, // Earthy brown
			{R: 0x6B, G: 0x8E, B: 0x4E, A: 255}, // Forest green
			{R: 0xC4, G: 0xA3, B: 0x5A, A: 255}, // Golden
			{R: 0x7B, G: 0x68, B: 0x8F, A: 255}, // Mystical purple
		},
		Accent: []color.RGBA{
			{R: 0xFF, G: 0xD7, B: 0x00, A: 255}, // Gold
			{R: 0x88, G: 0xCC, B: 0xFF, A: 255}, // Arcane blue
			{R: 0xFF, G: 0x66, B: 0x99, A: 255}, // Enchanted pink
		},
		AmbientTint: color.RGBA{R: 0x10, G: 0x08, B: 0x00, A: 255},
		Saturation:  1.1,
		Brightness:  1.05,
		Contrast:    1.0,
	},
	"sci-fi": {
		Primary: []color.RGBA{
			{R: 0x2A, G: 0x3A, B: 0x4D, A: 255}, // Deep blue
			{R: 0x45, G: 0x45, B: 0x50, A: 255}, // Tech gray
			{R: 0x88, G: 0xCC, B: 0xFF, A: 255}, // Holographic blue
			{R: 0xE0, G: 0xE0, B: 0xE8, A: 255}, // Clean white
		},
		Accent: []color.RGBA{
			{R: 0x00, G: 0xFF, B: 0xFF, A: 255}, // Cyan
			{R: 0xFF, G: 0x88, B: 0x00, A: 255}, // Warning orange
			{R: 0x00, G: 0xFF, B: 0x88, A: 255}, // Tech green
		},
		AmbientTint: color.RGBA{R: 0x00, G: 0x05, B: 0x10, A: 255},
		Saturation:  0.9,
		Brightness:  1.1,
		Contrast:    1.15,
	},
	"horror": {
		Primary: []color.RGBA{
			{R: 0x2D, G: 0x2D, B: 0x2D, A: 255}, // Dark gray
			{R: 0x3D, G: 0x30, B: 0x30, A: 255}, // Bloody shadow
			{R: 0x50, G: 0x45, B: 0x35, A: 255}, // Decay brown
			{R: 0x28, G: 0x28, B: 0x20, A: 255}, // Mold green
		},
		Accent: []color.RGBA{
			{R: 0x8B, G: 0x00, B: 0x00, A: 255}, // Deep red
			{R: 0x55, G: 0xAA, B: 0x55, A: 255}, // Sickly green
			{R: 0xFF, G: 0xFF, B: 0x88, A: 255}, // Pale light
		},
		AmbientTint: color.RGBA{R: 0x08, G: 0x00, B: 0x00, A: 255},
		Saturation:  0.7,
		Brightness:  0.8,
		Contrast:    1.3,
	},
	"cyberpunk": {
		Primary: []color.RGBA{
			{R: 0x15, G: 0x15, B: 0x25, A: 255}, // Night black
			{R: 0x25, G: 0x25, B: 0x35, A: 255}, // Urban gray
			{R: 0x40, G: 0x30, B: 0x50, A: 255}, // Purple-gray
			{R: 0x2A, G: 0x40, B: 0x4D, A: 255}, // Teal shadow
		},
		Accent: []color.RGBA{
			{R: 0xFF, G: 0x00, B: 0x80, A: 255}, // Neon pink
			{R: 0x00, G: 0xFF, B: 0xCC, A: 255}, // Neon cyan
			{R: 0xFF, G: 0xFF, B: 0x00, A: 255}, // Electric yellow
			{R: 0x80, G: 0x00, B: 0xFF, A: 255}, // Neon purple
		},
		AmbientTint: color.RGBA{R: 0x10, G: 0x00, B: 0x15, A: 255},
		Saturation:  1.3,
		Brightness:  0.95,
		Contrast:    1.4,
	},
	"post-apocalyptic": {
		Primary: []color.RGBA{
			{R: 0x6D, G: 0x5E, B: 0x4E, A: 255}, // Dust brown
			{R: 0x8B, G: 0x6F, B: 0x4E, A: 255}, // Rust orange
			{R: 0x55, G: 0x55, B: 0x4D, A: 255}, // Ash gray
			{R: 0x70, G: 0x58, B: 0x42, A: 255}, // Decay tan
		},
		Accent: []color.RGBA{
			{R: 0xA0, G: 0x50, B: 0x00, A: 255}, // Deep rust
			{R: 0x80, G: 0x90, B: 0x60, A: 255}, // Overgrowth green
			{R: 0xD0, G: 0xB0, B: 0x70, A: 255}, // Faded yellow
		},
		AmbientTint: color.RGBA{R: 0x10, G: 0x08, B: 0x00, A: 255},
		Saturation:  0.75,
		Brightness:  0.9,
		Contrast:    1.1,
	},
}

// GetGenreColorScheme returns the palette for a genre, or a default neutral palette.
func GetGenreColorScheme(genre string) *GenreColorScheme {
	if p, ok := GenreColorSchemes[genre]; ok {
		return p
	}
	// Default neutral palette
	return &GenreColorScheme{
		Primary: []color.RGBA{
			{R: 0x80, G: 0x80, B: 0x80, A: 255},
			{R: 0x70, G: 0x70, B: 0x70, A: 255},
			{R: 0x90, G: 0x90, B: 0x90, A: 255},
		},
		Accent: []color.RGBA{
			{R: 0xFF, G: 0xFF, B: 0xFF, A: 255},
		},
		AmbientTint: color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 255},
		Saturation:  1.0,
		Brightness:  1.0,
		Contrast:    1.0,
	}
}

// MaterialGenrePalette combines material properties with genre styling.
type MaterialGenrePalette struct {
	// Material being styled
	Material MaterialID
	// Genre to apply
	Genre string
	// Condition affects wear/aging (0.0 = pristine, 1.0 = heavily worn)
	Condition float64
}

// GetMaterialPaletteForGenre returns specialized colors for a material in a genre context.
// This provides more nuanced palettes than simple material base colors.
func GetMaterialPaletteForGenre(materialID MaterialID, genre string, condition float64) []color.RGBA {
	// Get base material colors
	mat := DefaultMaterialRegistry.Get(materialID)
	if mat == nil {
		return nil
	}

	// Check for explicit genre override first
	if colors := DefaultMaterialRegistry.GetColorsForGenre(materialID, genre); colors != nil && len(colors) > 0 {
		// Apply condition-based modifications
		return applyConditionToColors(colors, condition, genre)
	}

	// Generate genre-appropriate colors based on material properties
	palette := GetGenreColorScheme(genre)
	baseColors := mat.BaseColors

	// Modify colors based on genre palette
	result := make([]color.RGBA, len(baseColors))
	for i, c := range baseColors {
		result[i] = applyGenreStyling(c, palette, mat.Category)
	}

	return applyConditionToColors(result, condition, genre)
}

// applyGenreStyling modifies a color based on genre palette and material category.
func applyGenreStyling(c color.RGBA, palette *GenreColorScheme, category string) color.RGBA {
	// Convert to float for manipulation
	r := float64(c.R)
	g := float64(c.G)
	b := float64(c.B)

	// Apply ambient tint
	r += float64(palette.AmbientTint.R)
	g += float64(palette.AmbientTint.G)
	b += float64(palette.AmbientTint.B)

	// Apply saturation adjustment
	gray := 0.299*r + 0.587*g + 0.114*b
	r = gray + (r-gray)*palette.Saturation
	g = gray + (g-gray)*palette.Saturation
	b = gray + (b-gray)*palette.Saturation

	// Apply brightness
	r *= palette.Brightness
	g *= palette.Brightness
	b *= palette.Brightness

	// Apply contrast
	r = (r-128)*palette.Contrast + 128
	g = (g-128)*palette.Contrast + 128
	b = (b-128)*palette.Contrast + 128

	// Clamp
	return color.RGBA{
		R: clampByte(r),
		G: clampByte(g),
		B: clampByte(b),
		A: c.A,
	}
}

// applyConditionToColors modifies colors based on wear condition.
func applyConditionToColors(colors []color.RGBA, condition float64, genre string) []color.RGBA {
	if condition <= 0 {
		return colors
	}
	if condition > 1 {
		condition = 1
	}

	result := make([]color.RGBA, len(colors))
	for i, c := range colors {
		// Simulate weathering: reduce saturation and add dust/grime
		r := float64(c.R)
		g := float64(c.G)
		b := float64(c.B)

		// Calculate gray for desaturation
		gray := 0.299*r + 0.587*g + 0.114*b

		// Desaturate based on condition
		saturationLoss := condition * 0.4
		r = r + (gray-r)*saturationLoss
		g = g + (gray-g)*saturationLoss
		b = b + (gray-b)*saturationLoss

		// Add genre-appropriate grime
		switch genre {
		case "horror":
			// Add greenish-brown grime
			r *= 1.0 - condition*0.15
			g *= 1.0 - condition*0.1
			b *= 1.0 - condition*0.2
		case "post-apocalyptic":
			// Add dusty orange tint
			r += condition * 15
			g += condition * 5
			b -= condition * 10
		case "cyberpunk":
			// Add gritty urban dirt
			r *= 1.0 - condition*0.1
			g *= 1.0 - condition*0.1
			b *= 1.0 - condition*0.05
		default:
			// Generic wear
			r *= 1.0 - condition*0.1
			g *= 1.0 - condition*0.1
			b *= 1.0 - condition*0.1
		}

		result[i] = color.RGBA{
			R: clampByte(r),
			G: clampByte(g),
			B: clampByte(b),
			A: c.A,
		}
	}
	return result
}

// clampByte clamps a float to 0-255 range and returns as uint8.
func clampByte(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

// GetRustyMetalPalette returns rusty metal colors appropriate for the genre.
func GetRustyMetalPalette(genre string, severity float64) []color.RGBA {
	// Base rusty colors
	base := []color.RGBA{
		{R: 0x8B, G: 0x45, B: 0x13, A: 255}, // Saddle brown rust
		{R: 0xA0, G: 0x52, B: 0x2D, A: 255}, // Sienna rust
		{R: 0x6B, G: 0x3A, B: 0x1A, A: 255}, // Dark rust
		{R: 0xB8, G: 0x73, B: 0x33, A: 255}, // Light rust
	}

	// Modify based on genre
	switch genre {
	case "post-apocalyptic":
		// More orange, sun-bleached rust
		for i := range base {
			base[i].R = clampByte(float64(base[i].R) + 20*severity)
			base[i].G = clampByte(float64(base[i].G) - 10*severity)
		}
	case "horror":
		// Darker, blood-like rust
		for i := range base {
			base[i].R = clampByte(float64(base[i].R) + 15*severity)
			base[i].G = clampByte(float64(base[i].G) * (1 - 0.3*severity))
			base[i].B = clampByte(float64(base[i].B) * (1 - 0.2*severity))
		}
	case "sci-fi":
		// Minimal rust, more corrosion patches
		for i := range base {
			base[i].R = clampByte(float64(base[i].R) * (1 - 0.2*severity))
			base[i].G = clampByte(float64(base[i].G) + 10*severity)
		}
	}

	return base
}

// GetPolishedChromePalette returns polished chrome colors appropriate for the genre.
func GetPolishedChromePalette(genre string, reflectivity float64) []color.RGBA {
	// Base chrome colors
	base := []color.RGBA{
		{R: 0xD4, G: 0xD4, B: 0xD8, A: 255}, // Silver
		{R: 0xE8, G: 0xE8, B: 0xEC, A: 255}, // Bright chrome
		{R: 0xB8, G: 0xB8, B: 0xC0, A: 255}, // Muted chrome
		{R: 0xF0, G: 0xF0, B: 0xF4, A: 255}, // Highlight
	}

	// Modify based on genre
	switch genre {
	case "cyberpunk":
		// Neon reflections in chrome
		for i := range base {
			base[i].R = clampByte(float64(base[i].R) + 10*reflectivity)
			base[i].B = clampByte(float64(base[i].B) + 15*reflectivity)
		}
	case "sci-fi":
		// Blue-tinted clean chrome
		for i := range base {
			base[i].B = clampByte(float64(base[i].B) + 20*reflectivity)
		}
	case "post-apocalyptic":
		// Scratched, tarnished chrome
		for i := range base {
			base[i].R = clampByte(float64(base[i].R) - 20*(1-reflectivity))
			base[i].G = clampByte(float64(base[i].G) - 25*(1-reflectivity))
			base[i].B = clampByte(float64(base[i].B) - 30*(1-reflectivity))
		}
	case "horror":
		// Dirty, ominous chrome
		for i := range base {
			gray := (float64(base[i].R) + float64(base[i].G) + float64(base[i].B)) / 3
			base[i].R = clampByte(gray - 10)
			base[i].G = clampByte(gray - 15)
			base[i].B = clampByte(gray - 5)
		}
	}

	return base
}

// GetWeatheredStonePalette returns weathered stone colors appropriate for the genre.
func GetWeatheredStonePalette(genre string, weathering float64) []color.RGBA {
	// Base weathered stone colors
	base := []color.RGBA{
		{R: 0x78, G: 0x78, B: 0x70, A: 255}, // Gray-green stone
		{R: 0x68, G: 0x68, B: 0x60, A: 255}, // Dark weathered
		{R: 0x88, G: 0x85, B: 0x7D, A: 255}, // Light weathered
		{R: 0x70, G: 0x6B, B: 0x65, A: 255}, // Mid weathered
	}

	// Modify based on genre
	switch genre {
	case "fantasy":
		// Mossy, ancient stone
		for i := range base {
			base[i].G = clampByte(float64(base[i].G) + 15*weathering)
			base[i].R = clampByte(float64(base[i].R) - 10*weathering)
		}
	case "horror":
		// Dark, stained stone
		for i := range base {
			base[i].R = clampByte(float64(base[i].R) - 20*weathering)
			base[i].G = clampByte(float64(base[i].G) - 25*weathering)
			base[i].B = clampByte(float64(base[i].B) - 20*weathering)
		}
	case "post-apocalyptic":
		// Crumbling, dusty stone
		for i := range base {
			base[i].R = clampByte(float64(base[i].R) + 15*weathering)
			base[i].G = clampByte(float64(base[i].G) + 10*weathering)
			base[i].B = clampByte(float64(base[i].B) - 5*weathering)
		}
	case "sci-fi":
		// Clean but eroded
		for i := range base {
			gray := (float64(base[i].R) + float64(base[i].G) + float64(base[i].B)) / 3
			base[i].R = clampByte(gray + 5)
			base[i].G = clampByte(gray + 5)
			base[i].B = clampByte(gray + 10)
		}
	}

	return base
}

// ============================================================
// Surface Wear/Aging System
// ============================================================

// WearType describes the kind of wear effect to apply.
type WearType int

const (
	WearNone      WearType = iota
	WearScratches          // Surface scratches and abrasions
	WearRust               // Oxidation/corrosion
	WearDirt               // Accumulated dirt/grime
	WearFade               // Color fading from UV exposure
	WearChip               // Chipped edges and damage
	WearMoss               // Organic growth (moss, lichen)
	WearStain              // Stains and discoloration
)

// WearConfig describes wear parameters for texture aging.
type WearConfig struct {
	// Age represents the surface age in arbitrary units (0.0 = new, 1.0+ = very old)
	Age float64

	// WearResistance determines how well the material resists wear (0.0-1.0)
	// Higher values mean slower aging
	WearResistance float64

	// ExposureType indicates the environment (indoor, outdoor, underwater, etc.)
	ExposureType string

	// PrimaryWear is the dominant wear type for this material
	PrimaryWear WearType

	// SecondaryWear is a secondary wear effect (can be WearNone)
	SecondaryWear WearType

	// WearSeed for deterministic wear pattern generation
	WearSeed int64
}

// GetWearConfigForMaterial returns appropriate wear settings for a material.
func GetWearConfigForMaterial(materialID MaterialID, age float64, seed int64) WearConfig {
	material := DefaultMaterialRegistry.Get(materialID)
	if material == nil {
		return WearConfig{Age: age, WearSeed: seed, PrimaryWear: WearDirt}
	}

	config := WearConfig{
		Age:            age,
		WearSeed:       seed,
		WearResistance: 0.5,
		ExposureType:   "outdoor",
	}

	// Assign wear types based on material category
	switch material.Category {
	case "metal":
		config.PrimaryWear = WearRust
		config.SecondaryWear = WearScratches
		config.WearResistance = 0.6
	case "organic":
		config.PrimaryWear = WearFade
		config.SecondaryWear = WearMoss
		config.WearResistance = 0.3
	case "mineral":
		config.PrimaryWear = WearChip
		config.SecondaryWear = WearMoss
		config.WearResistance = 0.8
	case "natural":
		config.PrimaryWear = WearDirt
		config.SecondaryWear = WearMoss
		config.WearResistance = 0.2
	case "synthetic":
		config.PrimaryWear = WearFade
		config.SecondaryWear = WearStain
		config.WearResistance = 0.7
	}

	return config
}

// ApplyWear applies wear effects to a texture based on age and material properties.
// Returns a new slice of pixels with wear applied.
func ApplyWear(pixels []color.RGBA, width, height int, config WearConfig) []color.RGBA {
	if len(pixels) == 0 || width <= 0 || height <= 0 {
		return pixels
	}

	// Calculate effective wear amount
	effectiveAge := config.Age * (1.0 - config.WearResistance*0.7)
	if effectiveAge <= 0 {
		return pixels
	}
	if effectiveAge > 2.0 {
		effectiveAge = 2.0 // Cap at maximum wear
	}

	result := make([]color.RGBA, len(pixels))
	copy(result, pixels)

	// Apply primary wear
	applyWearEffect(result, width, height, config.PrimaryWear, effectiveAge, config.WearSeed)

	// Apply secondary wear at reduced intensity
	if config.SecondaryWear != WearNone {
		applyWearEffect(result, width, height, config.SecondaryWear, effectiveAge*0.4, config.WearSeed+1000)
	}

	// Apply edge wear (erosion at top/bottom rows of texture)
	applyEdgeWear(result, width, height, effectiveAge, config.WearSeed+2000)

	return result
}

// applyWearEffect applies a specific wear type to pixels.
func applyWearEffect(pixels []color.RGBA, width, height int, wearType WearType, intensity float64, seed int64) {
	switch wearType {
	case WearScratches:
		applyScratchWear(pixels, width, height, intensity, seed)
	case WearRust:
		applyRustWear(pixels, width, height, intensity, seed)
	case WearDirt:
		applyDirtWear(pixels, width, height, intensity, seed)
	case WearFade:
		applyFadeWear(pixels, width, height, intensity, seed)
	case WearChip:
		applyChipWear(pixels, width, height, intensity, seed)
	case WearMoss:
		applyMossWear(pixels, width, height, intensity, seed)
	case WearStain:
		applyStainWear(pixels, width, height, intensity, seed)
	}
}

// applyScratchWear adds linear scratch marks to the surface.
func applyScratchWear(pixels []color.RGBA, width, height int, intensity float64, seed int64) {
	numScratches := int(intensity * 20)
	rng := seedRng(seed)

	for s := 0; s < numScratches; s++ {
		// Random scratch line
		x1 := rng.Intn(width)
		y1 := rng.Intn(height)
		length := rng.Intn(width/4) + 5
		angle := rng.Float64() * 3.14159

		for i := 0; i < length; i++ {
			x := x1 + int(float64(i)*cosApprox(angle))
			y := y1 + int(float64(i)*sinApprox(angle))
			if x >= 0 && x < width && y >= 0 && y < height {
				idx := y*width + x
				// Lighten (expose metal beneath)
				pixels[idx] = blendColor(pixels[idx], color.RGBA{R: 180, G: 180, B: 180, A: 255}, intensity*0.3)
			}
		}
	}
}

// applyRustWear adds rust/oxidation patches.
func applyRustWear(pixels []color.RGBA, width, height int, intensity float64, seed int64) {
	rustColor := color.RGBA{R: 139, G: 69, B: 19, A: 255}
	darkRust := color.RGBA{R: 101, G: 67, B: 33, A: 255}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Noise-based rust pattern
			n := noiseAt(float64(x)*0.1, float64(y)*0.1, seed)
			threshold := 1.0 - intensity*0.5
			if n > threshold {
				idx := y*width + x
				rustAmount := (n - threshold) / (1.0 - threshold) * intensity
				// Choose between light and dark rust
				if noiseAt(float64(x)*0.3, float64(y)*0.3, seed+100) > 0.5 {
					pixels[idx] = blendColor(pixels[idx], rustColor, rustAmount)
				} else {
					pixels[idx] = blendColor(pixels[idx], darkRust, rustAmount)
				}
			}
		}
	}
}

// applyDirtWear adds accumulated dirt/grime.
func applyDirtWear(pixels []color.RGBA, width, height int, intensity float64, seed int64) {
	dirtColor := color.RGBA{R: 80, G: 70, B: 55, A: 255}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Dirt accumulates in crevices (low-frequency areas)
			n := noiseAt(float64(x)*0.05, float64(y)*0.05, seed)
			detail := noiseAt(float64(x)*0.2, float64(y)*0.2, seed+50)

			// Combine noises for natural-looking dirt
			dirtAmount := n*0.7 + detail*0.3
			dirtAmount *= intensity * 0.6

			if dirtAmount > 0.1 {
				idx := y*width + x
				pixels[idx] = blendColor(pixels[idx], dirtColor, dirtAmount)
			}
		}
	}
}

// applyFadeWear reduces color saturation and brightness.
func applyFadeWear(pixels []color.RGBA, width, height int, intensity float64, seed int64) {
	fadeFactor := 1.0 - intensity*0.4

	for i := range pixels {
		p := pixels[i]
		// Reduce saturation
		gray := uint8(0.299*float64(p.R) + 0.587*float64(p.G) + 0.114*float64(p.B))
		pixels[i] = color.RGBA{
			R: uint8(float64(p.R)*fadeFactor + float64(gray)*(1-fadeFactor)),
			G: uint8(float64(p.G)*fadeFactor + float64(gray)*(1-fadeFactor)),
			B: uint8(float64(p.B)*fadeFactor + float64(gray)*(1-fadeFactor)),
			A: p.A,
		}
	}
}

// applyChipWear adds chipped edges and damage.
func applyChipWear(pixels []color.RGBA, width, height int, intensity float64, seed int64) {
	chipColor := color.RGBA{R: 100, G: 100, B: 100, A: 255}
	numChips := int(intensity * 15)
	rng := seedRng(seed)

	for c := 0; c < numChips; c++ {
		cx := rng.Intn(width)
		cy := rng.Intn(height)
		size := rng.Intn(5) + 2

		// Irregular chip shape
		for dy := -size; dy <= size; dy++ {
			for dx := -size; dx <= size; dx++ {
				x, y := cx+dx, cy+dy
				if x >= 0 && x < width && y >= 0 && y < height {
					dist := float64(dx*dx + dy*dy)
					if dist < float64(size*size) && rng.Float64() > 0.3 {
						idx := y*width + x
						pixels[idx] = blendColor(pixels[idx], chipColor, intensity*0.5)
					}
				}
			}
		}
	}
}

// applyMossWear adds organic growth (moss, lichen).
func applyMossWear(pixels []color.RGBA, width, height int, intensity float64, seed int64) {
	mossColor := color.RGBA{R: 50, G: 80, B: 40, A: 255}
	lichenColor := color.RGBA{R: 120, G: 120, B: 80, A: 255}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			n := noiseAt(float64(x)*0.08, float64(y)*0.08, seed)
			if n > 0.7-intensity*0.3 {
				idx := y*width + x
				mossAmount := (n - 0.5) * intensity
				// Choose between moss and lichen
				if noiseAt(float64(x)*0.2, float64(y)*0.2, seed+200) > 0.5 {
					pixels[idx] = blendColor(pixels[idx], mossColor, mossAmount)
				} else {
					pixels[idx] = blendColor(pixels[idx], lichenColor, mossAmount*0.7)
				}
			}
		}
	}
}

// applyStainWear adds water stains and discoloration.
func applyStainWear(pixels []color.RGBA, width, height int, intensity float64, seed int64) {
	stainColor := color.RGBA{R: 90, G: 85, B: 70, A: 255}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Stains flow downward
			n := noiseAt(float64(x)*0.1, float64(y)*0.03, seed) // Elongated vertically
			if n > 0.6 {
				idx := y*width + x
				stainAmount := (n - 0.6) * 2.5 * intensity
				pixels[idx] = blendColor(pixels[idx], stainColor, stainAmount*0.4)
			}
		}
	}
}

// blendColor blends two colors with the given factor (0.0=a, 1.0=b).
func blendColor(a, b color.RGBA, factor float64) color.RGBA {
	if factor <= 0 {
		return a
	}
	if factor >= 1 {
		return b
	}
	invFactor := 1.0 - factor
	return color.RGBA{
		R: uint8(float64(a.R)*invFactor + float64(b.R)*factor),
		G: uint8(float64(a.G)*invFactor + float64(b.G)*factor),
		B: uint8(float64(a.B)*invFactor + float64(b.B)*factor),
		A: a.A,
	}
}

// applyEdgeWear applies increased wear at texture edges to simulate erosion.
// The top and bottom rows of wall textures receive enhanced weathering.
func applyEdgeWear(pixels []color.RGBA, width, height int, intensity float64, seed int64) {
	if height < 4 || intensity <= 0 {
		return
	}

	// Edge zone is the top and bottom 10% of texture height (min 2 rows)
	edgeRows := height / 10
	if edgeRows < 2 {
		edgeRows = 2
	}

	// Edge erosion color (darkened, weathered)
	erosionColor := color.RGBA{R: 60, G: 55, B: 50, A: 255}

	for x := 0; x < width; x++ {
		// Top edge wear
		for y := 0; y < edgeRows; y++ {
			// Wear intensity increases toward the very top
			edgeFactor := float64(edgeRows-y) / float64(edgeRows)
			n := noiseAt(float64(x)*0.15, float64(y)*0.15, seed)
			wearAmount := intensity * edgeFactor * n * 0.4
			if wearAmount > 0.01 {
				idx := y*width + x
				pixels[idx] = blendColor(pixels[idx], erosionColor, wearAmount)
			}
		}

		// Bottom edge wear
		for y := height - edgeRows; y < height; y++ {
			// Wear intensity increases toward the very bottom
			edgeFactor := float64(y-(height-edgeRows)) / float64(edgeRows)
			n := noiseAt(float64(x)*0.15, float64(y)*0.15, seed+500)
			wearAmount := intensity * edgeFactor * n * 0.4
			if wearAmount > 0.01 {
				idx := y*width + x
				pixels[idx] = blendColor(pixels[idx], erosionColor, wearAmount)
			}
		}
	}
}

// noiseAt is a simplified noise lookup for wear patterns.
func noiseAt(x, y float64, seed int64) float64 {
	// Simple hash-based noise
	ix := int(x * 100)
	iy := int(y * 100)
	h := uint64(seed) ^ uint64(ix*374761393) ^ uint64(iy*668265263)
	h ^= h >> 13
	h *= 1274126177
	h ^= h >> 16
	return float64(h&0xFFFFFF) / float64(0xFFFFFF)
}

// seedRng creates a simple deterministic random generator.
type simpleRng struct {
	state uint64
}

func seedRng(seed int64) *simpleRng {
	return &simpleRng{state: uint64(seed)}
}

func (r *simpleRng) Intn(n int) int {
	r.state = r.state*6364136223846793005 + 1442695040888963407
	return int((r.state >> 33) % uint64(n))
}

func (r *simpleRng) Float64() float64 {
	r.state = r.state*6364136223846793005 + 1442695040888963407
	return float64(r.state>>11) / float64(1<<53)
}

// sinApprox and cosApprox are simple trig approximations for wear patterns.
func sinApprox(x float64) float64 {
	// Normalize to [0, 2π]
	pi2 := 6.28318530718
	for x < 0 {
		x += pi2
	}
	for x >= pi2 {
		x -= pi2
	}
	// Taylor series (good enough for wear patterns)
	if x > 3.14159 {
		x -= 3.14159
		return -sinApproxInner(x)
	}
	return sinApproxInner(x)
}

func sinApproxInner(x float64) float64 {
	x2 := x * x
	return x * (1 - x2/6 + x2*x2/120)
}

func cosApprox(x float64) float64 {
	return sinApprox(x + 1.5708)
}
