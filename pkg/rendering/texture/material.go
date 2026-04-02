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
