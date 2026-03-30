//go:build noebiten

// Package adapters provides V-Series integration for Wyrm.
// This file provides stub implementations for headless testing.
package adapters

import (
	"image/color"
	"math/rand"
)

// FurnitureAdapter wraps Venture's furniture generator.
// Stub implementation for headless testing.
type FurnitureAdapter struct{}

// NewFurnitureAdapter creates a new furniture adapter.
func NewFurnitureAdapter() *FurnitureAdapter { return &FurnitureAdapter{} }

// FurnitureData holds generated furniture information.
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

// GenerateFurniture creates a furniture piece.
func (a *FurnitureAdapter) GenerateFurniture(seed int64, genre, subType string) (*FurnitureData, error) {
	rng := rand.New(rand.NewSource(seed))
	return &FurnitureData{
		ID:         "furniture_" + subType,
		Name:       subType + " furniture",
		Type:       "furniture",
		SubType:    subType,
		Material:   "wood",
		Rarity:     "common",
		Genre:      genre,
		Width:      1 + rng.Float64(),
		Height:     1 + rng.Float64(),
		Depth:      1 + rng.Float64(),
		Functional: rng.Float64() > 0.5,
		Capacity:   rng.Intn(10),
	}, nil
}

// GenerateRoomFurniture creates furniture for a specific room type.
func (a *FurnitureAdapter) GenerateRoomFurniture(seed int64, genre, roomType string, count int) ([]*FurnitureData, error) {
	furniture := make([]*FurnitureData, count)
	for i := 0; i < count; i++ {
		furniture[i], _ = a.GenerateFurniture(seed+int64(i), genre, roomType)
	}
	return furniture, nil
}

// GenerateHouseFurniture creates furniture for an entire house.
func (a *FurnitureAdapter) GenerateHouseFurniture(seed int64, genre string) (map[string][]*FurnitureData, error) {
	result := make(map[string][]*FurnitureData)

	result["living"], _ = a.GenerateRoomFurniture(seed, genre, "living", LivingRoomFurnitureCount)
	result["bedroom"], _ = a.GenerateRoomFurniture(seed+BedroomSeedOffset, genre, "bedroom", BedroomFurnitureCount)
	result["kitchen"], _ = a.GenerateRoomFurniture(seed+KitchenSeedOffset, genre, "kitchen", KitchenFurnitureCount)
	result["storage"], _ = a.GenerateRoomFurniture(seed+StorageSeedOffset, genre, "storage", StorageFurnitureCount)

	return result, nil
}

// GenerateCraftingStations creates workbench furniture.
func (a *FurnitureAdapter) GenerateCraftingStations(seed int64, genre string) ([]*FurnitureData, error) {
	return a.GenerateRoomFurniture(seed, genre, "crafting", 4)
}

// IsFurnitureFunctional checks if furniture has a function.
func IsFurnitureFunctional(data *FurnitureData) bool { return data.Functional }

// GetFurnitureValue returns furniture's gold value.
func GetFurnitureValue(data *FurnitureData) int { return data.Capacity * BaseFurnitureValue }
