//go:build noebiten

// Package adapters provides V-Series integration for Wyrm.
// This file provides stub implementations for headless testing.
package adapters

import "math/rand"

// EnvironmentAdapter wraps Venture's environment generator.
// Stub implementation for headless testing.
type EnvironmentAdapter struct{}

// NewEnvironmentAdapter creates a new environment adapter.
func NewEnvironmentAdapter() *EnvironmentAdapter { return &EnvironmentAdapter{} }

// EnvironmentObjectData holds data for environmental objects.
type EnvironmentObjectData struct {
	Type         string
	SubType      string
	Width        int
	Height       int
	Collidable   bool
	Interactable bool
	Harmful      bool
	Damage       int
	Genre        string
	Name         string
	Seed         int64
}

// PlacedObjectData holds an environmental object with position.
type PlacedObjectData struct {
	Object *EnvironmentObjectData
	X, Y   int
}

// GenerateEnvironmentObject creates an environmental object.
func (a *EnvironmentAdapter) GenerateEnvironmentObject(seed int64, genre, subType string) (*EnvironmentObjectData, error) {
	return &EnvironmentObjectData{
		Type:       "decoration",
		SubType:    subType,
		Harmful:    false,
		Width:      1,
		Height:     1,
		Collidable: false,
		Genre:      genre,
		Name:       subType + " object",
		Seed:       seed,
	}, nil
}

// GenerateChunkDecorations creates decorations for a terrain chunk.
func (a *EnvironmentAdapter) GenerateChunkDecorations(seed int64, genre, biomeType string) ([]*PlacedObjectData, error) {
	rng := rand.New(rand.NewSource(seed))
	count := 5 + rng.Intn(10)
	objects := make([]*PlacedObjectData, count)
	for i := 0; i < count; i++ {
		obj, _ := a.GenerateEnvironmentObject(seed+int64(i), genre, biomeType)
		objects[i] = &PlacedObjectData{Object: obj, X: rng.Intn(64), Y: rng.Intn(64)}
	}
	return objects, nil
}

// GenerateRoomDecorations creates decorations for a room.
func (a *EnvironmentAdapter) GenerateRoomDecorations(seed int64, genre string, roomWidth, roomHeight int, density float64) ([]*PlacedObjectData, error) {
	rng := rand.New(rand.NewSource(seed))
	count := int(float64(roomWidth*roomHeight) * density * 0.1)
	objects := make([]*PlacedObjectData, count)
	for i := 0; i < count; i++ {
		obj, _ := a.GenerateEnvironmentObject(seed+int64(i), genre, "room")
		objects[i] = &PlacedObjectData{Object: obj, X: rng.Intn(roomWidth), Y: rng.Intn(roomHeight)}
	}
	return objects, nil
}

// GenerateBiomeObjects creates objects for a specific biome.
func (a *EnvironmentAdapter) GenerateBiomeObjects(seed int64, genre, biomeType string, count int) ([]*EnvironmentObjectData, error) {
	objects := make([]*EnvironmentObjectData, count)
	for i := 0; i < count; i++ {
		objects[i], _ = a.GenerateEnvironmentObject(seed+int64(i), genre, biomeType)
	}
	return objects, nil
}

// IsEnvironmentObjectHazard checks if an object is hazardous.
func IsEnvironmentObjectHazard(obj *EnvironmentObjectData) bool { return obj.Harmful }

// GetEnvironmentObjectDamage returns hazard damage value.
func GetEnvironmentObjectDamage(obj *EnvironmentObjectData) int { return obj.Damage }

// getDensityForBiome returns decoration density for biome type.
func getDensityForBiome(biomeType string) float64 {
	densities := map[string]float64{
		"forest": 0.8, "mountain": 0.4, "swamp": 0.6,
		"wasteland": 0.2, "urban": 0.5, "ruins": 0.3,
	}
	if d, ok := densities[biomeType]; ok {
		return d
	}
	return 0.5
}

// getBiomeObjectTypes returns valid object types for a biome.
func getBiomeObjectTypes(biomeType string) []string {
	types := map[string][]string{
		"forest":    {"tree", "bush", "rock", "mushroom"},
		"mountain":  {"boulder", "cliff", "snow"},
		"swamp":     {"log", "mud", "reeds"},
		"wasteland": {"debris", "crater", "bones"},
		"urban":     {"sign", "bench", "trash"},
		"ruins":     {"rubble", "pillar", "wall"},
	}
	if t, ok := types[biomeType]; ok {
		return t
	}
	return []string{"rock", "bush"}
}
