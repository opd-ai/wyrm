//go:build ebitentest

// Package adapters provides V-Series integration for Wyrm.
package adapters

import (
	"fmt"

	"github.com/opd-ai/venture/pkg/procgen/environment"
)

// EnvironmentAdapter wraps Venture's environment generator for Wyrm's chunk decoration.
type EnvironmentAdapter struct {
	generator *environment.Generator
	placer    *environment.Placer
}

// NewEnvironmentAdapter creates a new environment adapter.
func NewEnvironmentAdapter() *EnvironmentAdapter {
	return &EnvironmentAdapter{
		generator: environment.NewGenerator(),
		placer:    environment.NewPlacer(),
	}
}

// EnvironmentObjectData holds environment object information adapted for Wyrm.
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

// PlacedObjectData holds placed object information for room decoration.
type PlacedObjectData struct {
	Object    *EnvironmentObjectData
	X         float64
	Y         float64
	Rotation  float64
	Scale     float64
	Placement string
}

// GenerateEnvironmentObject generates a single environment object.
func (a *EnvironmentAdapter) GenerateEnvironmentObject(seed int64, genre, subType string) (*EnvironmentObjectData, error) {
	config := environment.Config{
		GenreID: mapGenreID(genre),
		Seed:    seed,
		Width:   64,
		Height:  64,
	}

	if subType != "" {
		config.SubType = mapEnvironmentSubType(subType)
	}

	obj, err := a.generator.GenerateFromConfig(config)
	if err != nil {
		return nil, fmt.Errorf("environment object generation failed: %w", err)
	}

	return convertEnvironmentObject(obj), nil
}

// GenerateChunkDecorations generates decorations for a world chunk.
func (a *EnvironmentAdapter) GenerateChunkDecorations(seed int64, genre, biomeType string) ([]*PlacedObjectData, error) {
	config := environment.PlacementConfig{
		RoomWidth:  64, // Chunk size in units
		RoomHeight: 64,
		Density:    getDensityForBiome(biomeType),
		GenreID:    mapGenreID(genre),
		Seed:       seed,
	}

	placements, err := a.placer.PlaceDecorations(config)
	if err != nil {
		return nil, fmt.Errorf("decoration placement failed: %w", err)
	}

	result := make([]*PlacedObjectData, len(placements))
	for i, p := range placements {
		result[i] = &PlacedObjectData{
			Object:    convertEnvironmentObject(p.Object),
			X:         float64(p.X),
			Y:         float64(p.Y),
			Rotation:  p.Variation.Rotation,
			Scale:     p.Variation.Scale,
			Placement: p.PlacementType.String(),
		}
	}
	return result, nil
}

// GenerateRoomDecorations generates decorations for a specific room.
func (a *EnvironmentAdapter) GenerateRoomDecorations(seed int64, genre string, roomWidth, roomHeight int, density float64) ([]*PlacedObjectData, error) {
	config := environment.PlacementConfig{
		RoomWidth:  roomWidth,
		RoomHeight: roomHeight,
		Density:    density,
		GenreID:    mapGenreID(genre),
		Seed:       seed,
	}

	placements, err := a.placer.PlaceDecorations(config)
	if err != nil {
		return nil, fmt.Errorf("room decoration placement failed: %w", err)
	}

	result := make([]*PlacedObjectData, len(placements))
	for i, p := range placements {
		result[i] = &PlacedObjectData{
			Object:    convertEnvironmentObject(p.Object),
			X:         float64(p.X),
			Y:         float64(p.Y),
			Rotation:  p.Variation.Rotation,
			Scale:     p.Variation.Scale,
			Placement: p.PlacementType.String(),
		}
	}
	return result, nil
}

// GenerateBiomeObjects generates biome-appropriate objects for wilderness areas.
func (a *EnvironmentAdapter) GenerateBiomeObjects(seed int64, genre, biomeType string, count int) ([]*EnvironmentObjectData, error) {
	subTypes := getBiomeObjectTypes(biomeType)
	if len(subTypes) == 0 {
		return nil, fmt.Errorf("unknown biome type: %s", biomeType)
	}

	objects := make([]*EnvironmentObjectData, 0, count)
	for i := 0; i < count; i++ {
		subType := subTypes[i%len(subTypes)]
		obj, err := a.GenerateEnvironmentObject(seed+int64(i), genre, subType)
		if err != nil {
			continue
		}
		objects = append(objects, obj)
	}
	return objects, nil
}

// getDensityForBiome returns decoration density for a biome type.
func getDensityForBiome(biomeType string) float64 {
	switch biomeType {
	case "forest":
		return 0.8
	case "desert":
		return 0.2
	case "swamp":
		return 0.7
	case "mountain":
		return 0.3
	case "plains":
		return 0.4
	case "tundra":
		return 0.2
	case "urban":
		return 0.6
	default:
		return 0.5
	}
}

// getBiomeObjectTypes returns object types appropriate for a biome.
func getBiomeObjectTypes(biomeType string) []string {
	switch biomeType {
	case "forest":
		return []string{"plant", "mushroom", "moss", "grass"}
	case "desert":
		return []string{"rubble", "skull", "crate"}
	case "swamp":
		return []string{"plant", "mushroom", "moss", "spider_web"}
	case "mountain":
		return []string{"rubble", "pillar", "moss"}
	case "plains":
		return []string{"grass", "plant", "barrel"}
	case "urban":
		return []string{"barrel", "crate", "graffiti", "chains"}
	case "dungeon":
		return []string{"skull", "chains", "spider_web", "bloodstain", "sconce"}
	default:
		return []string{"plant", "barrel", "crate"}
	}
}

// mapEnvironmentSubType maps string subtype to Venture's SubType.
func mapEnvironmentSubType(subType string) environment.SubType {
	switch subType {
	case "table":
		return environment.SubTypeTable
	case "plant":
		return environment.SubTypePlant
	case "barrel":
		return environment.SubTypeBarrel
	case "spikes":
		return environment.SubTypeSpikes
	default:
		return environment.SubTypePlant
	}
}

// convertEnvironmentObject converts Venture environment object to Wyrm format.
func convertEnvironmentObject(obj *environment.EnvironmentalObject) *EnvironmentObjectData {
	return &EnvironmentObjectData{
		Type:         obj.Type.String(),
		SubType:      obj.SubType.String(),
		Width:        obj.Width,
		Height:       obj.Height,
		Collidable:   obj.Collidable,
		Interactable: obj.Interactable,
		Harmful:      obj.Harmful,
		Damage:       obj.Damage,
		Genre:        obj.GenreID,
		Name:         obj.Name,
		Seed:         obj.Seed,
	}
}

// IsEnvironmentObjectHazard returns true if the object is a hazard.
func IsEnvironmentObjectHazard(obj *EnvironmentObjectData) bool {
	return obj.Harmful
}

// GetEnvironmentObjectDamage returns the damage value for a hazardous object.
func GetEnvironmentObjectDamage(obj *EnvironmentObjectData) int {
	if !obj.Harmful {
		return 0
	}
	return obj.Damage
}
