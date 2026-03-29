//go:build ebitentest

// Package adapters provides V-Series integration for Wyrm.
package adapters

import (
	"fmt"
	"image/color"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/furniture"
)

// FurnitureAdapter wraps Venture's furniture generator for Wyrm's housing system.
type FurnitureAdapter struct {
	generator *furniture.Generator
}

// NewFurnitureAdapter creates a new furniture adapter.
func NewFurnitureAdapter() *FurnitureAdapter {
	return &FurnitureAdapter{
		generator: furniture.NewGenerator(),
	}
}

// FurnitureData holds furniture information adapted for Wyrm's housing.
type FurnitureData struct {
	ID             string
	Type           string
	SubType        string
	Material       string
	Rarity         string
	Name           string
	Description    string
	Genre          string
	Width          float64
	Height         float64
	Depth          float64
	PrimaryColor   color.RGBA
	SecondaryColor color.RGBA
	DetailLevel    float64
	Direction      string
	Walkable       bool
	CollisionWidth float64
	CollisionDepth float64
	Functional     bool
	Capacity       int
	CraftingType   string
	LightIntensity float64
}

// GenerateFurniture generates a single furniture piece.
func (a *FurnitureAdapter) GenerateFurniture(seed int64, genre, subType string) (*FurnitureData, error) {
	params := procgen.GenerationParams{
		GenreID:    mapGenreID(genre),
		Difficulty: 0.5,
		Custom:     map[string]interface{}{},
	}

	if subType != "" {
		params.Custom["SubType"] = subType
	}

	result, err := a.generator.Generate(seed, params)
	if err != nil {
		return nil, fmt.Errorf("furniture generation failed: %w", err)
	}

	f, ok := result.(*furniture.Furniture)
	if !ok {
		return nil, fmt.Errorf("invalid furniture result type: expected *furniture.Furniture, got %T", result)
	}

	return convertFurniture(f), nil
}

// GenerateRoomFurniture generates appropriate furniture for a room type.
func (a *FurnitureAdapter) GenerateRoomFurniture(seed int64, genre, roomType string, count int) ([]*FurnitureData, error) {
	subTypes := getRoomFurnitureTypes(roomType)
	if len(subTypes) == 0 {
		return nil, fmt.Errorf("unknown room type: %s", roomType)
	}

	var items []*FurnitureData
	for i := 0; i < count; i++ {
		subType := subTypes[i%len(subTypes)]
		item, err := a.GenerateFurniture(seed+int64(i), genre, subType)
		if err != nil {
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

// GenerateHouseFurniture generates a full set of furniture for a house.
func (a *FurnitureAdapter) GenerateHouseFurniture(seed int64, genre string) (map[string][]*FurnitureData, error) {
	rooms := map[string][]*FurnitureData{}

	// Living room
	livingRoom, _ := a.GenerateRoomFurniture(seed, genre, "living", 5)
	rooms["living"] = livingRoom

	// Bedroom
	bedroom, _ := a.GenerateRoomFurniture(seed+100, genre, "bedroom", 4)
	rooms["bedroom"] = bedroom

	// Kitchen
	kitchen, _ := a.GenerateRoomFurniture(seed+200, genre, "kitchen", 4)
	rooms["kitchen"] = kitchen

	// Storage
	storage, _ := a.GenerateRoomFurniture(seed+300, genre, "storage", 5)
	rooms["storage"] = storage

	return rooms, nil
}

// GenerateCraftingStations generates crafting furniture for a workshop.
func (a *FurnitureAdapter) GenerateCraftingStations(seed int64, genre string) ([]*FurnitureData, error) {
	craftingTypes := []string{"Workbench", "Forge", "Alchemy Table", "Enchanting Table", "Anvil"}
	var stations []*FurnitureData

	for i, subType := range craftingTypes {
		station, err := a.GenerateFurniture(seed+int64(i), genre, subType)
		if err != nil {
			continue
		}
		stations = append(stations, station)
	}
	return stations, nil
}

// getRoomFurnitureTypes returns furniture subtypes appropriate for a room.
func getRoomFurnitureTypes(roomType string) []string {
	switch roomType {
	case "living":
		return []string{"Chair", "Couch", "Table", "Shelf", "Painting"}
	case "bedroom":
		return []string{"Bed", "Wardrobe", "Chest", "Mirror"}
	case "kitchen":
		return []string{"Table", "Chair", "Cabinet", "Fireplace"}
	case "storage":
		return []string{"Chest", "Shelf", "Barrel", "Crate", "Cabinet"}
	case "workshop":
		return []string{"Workbench", "Forge", "Anvil", "Shelf"}
	case "throne_room":
		return []string{"Throne", "Tapestry", "Chandelier", "Statue"}
	case "tavern":
		return []string{"Table", "Chair", "Bench", "Barrel", "Lantern"}
	default:
		return []string{"Chair", "Table", "Chest"}
	}
}

// convertFurniture converts Venture furniture to Wyrm format.
func convertFurniture(f *furniture.Furniture) *FurnitureData {
	return &FurnitureData{
		ID:             f.ID,
		Type:           f.Type.String(),
		SubType:        f.SubType,
		Material:       f.Material.String(),
		Rarity:         f.Rarity.String(),
		Name:           f.Name,
		Description:    f.Description,
		Genre:          f.GenreID,
		Width:          f.Width,
		Height:         f.Height,
		Depth:          f.Depth,
		PrimaryColor:   f.PrimaryColor,
		SecondaryColor: f.SecondaryColor,
		DetailLevel:    f.DetailLevel,
		Direction:      f.Direction.String(),
		Walkable:       f.Walkable,
		CollisionWidth: f.CollisionWidth,
		CollisionDepth: f.CollisionDepth,
		Functional:     f.Functional,
		Capacity:       f.Capacity,
		CraftingType:   f.CraftingType,
		LightIntensity: f.LightIntensity,
	}
}

// IsFurnitureFunctional returns true if the furniture provides functionality.
func IsFurnitureFunctional(data *FurnitureData) bool {
	return data.Functional
}

// GetFurnitureValue calculates the value of a furniture piece.
func GetFurnitureValue(data *FurnitureData) int {
	baseValue := 10
	// Material multiplier
	switch data.Material {
	case "Wood":
		baseValue *= 1
	case "Metal":
		baseValue *= 2
	case "Stone":
		baseValue *= 2
	case "Crystal":
		baseValue *= 5
	case "Fabric":
		baseValue *= 1
	}
	// Rarity multiplier
	switch data.Rarity {
	case "Common":
		baseValue *= 1
	case "Uncommon":
		baseValue *= 2
	case "Rare":
		baseValue *= 5
	case "Epic":
		baseValue *= 10
	case "Legendary":
		baseValue *= 25
	}
	return baseValue
}
