//go:build noebiten

// Package adapters provides V-Series integration for Wyrm.
// This file provides stub implementations for headless testing.
package adapters

import (
	"math/rand"
)

// BuildingAdapter wraps Venture's building generator for Wyrm's city structures.
// Stub implementation for headless testing.
type BuildingAdapter struct{}

// NewBuildingAdapter creates a new building adapter.
func NewBuildingAdapter() *BuildingAdapter { return &BuildingAdapter{} }

// BuildingData holds generated building data for Wyrm integration.
type BuildingData struct {
	Type    string
	Style   string
	GenreID string
	Width   int
	Height  int
	Floors  int
	Rooms   []BuildingRoomData
	Doors   []BuildingDoorData
	Windows []BuildingWindowData
}

// BuildingRoomData holds room information within a building.
type BuildingRoomData struct {
	Name   string
	X, Y   int
	Width  int
	Height int
}

// BuildingDoorData holds door placement information.
type BuildingDoorData struct{ X, Y int }

// BuildingWindowData holds window placement information.
type BuildingWindowData struct{ X, Y int }

// GenerateBuilding creates a building of the specified type.
func (a *BuildingAdapter) GenerateBuilding(seed int64, genre string, buildingType, floors int) (*BuildingData, error) {
	rng := rand.New(rand.NewSource(seed))
	return &BuildingData{
		Type:   "house",
		Style:  genre,
		Width:  10 + rng.Intn(10),
		Height: 10 + rng.Intn(10),
		Floors: floors,
		Rooms:  []BuildingRoomData{{Name: "main", X: 0, Y: 0, Width: 8, Height: 8}},
	}, nil
}

// GenerateHouse creates a residential building.
func (a *BuildingAdapter) GenerateHouse(seed int64, genre string) (*BuildingData, error) {
	return a.GenerateBuilding(seed, genre, 0, 1)
}

// GenerateWorkshop creates a crafting workshop building.
func (a *BuildingAdapter) GenerateWorkshop(seed int64, genre string) (*BuildingData, error) {
	return a.GenerateBuilding(seed, genre, 1, 1)
}

// GenerateTower creates a tall tower structure.
func (a *BuildingAdapter) GenerateTower(seed int64, genre string) (*BuildingData, error) {
	return a.GenerateBuilding(seed, genre, 2, 3)
}

// GenerateManor creates a large residential building.
func (a *BuildingAdapter) GenerateManor(seed int64, genre string) (*BuildingData, error) {
	return a.GenerateBuilding(seed, genre, 3, 2)
}

// GenerateGuildHall creates a guild headquarters building.
func (a *BuildingAdapter) GenerateGuildHall(seed int64, genre string) (*BuildingData, error) {
	return a.GenerateBuilding(seed, genre, 4, 2)
}

// GenerateCityBuildings creates multiple buildings for a city.
func (a *BuildingAdapter) GenerateCityBuildings(seed int64, genre string, count int) ([]*BuildingData, error) {
	buildings := make([]*BuildingData, count)
	for i := 0; i < count; i++ {
		b, _ := a.GenerateBuilding(seed+int64(i), genre, i%5, 1+i%3)
		buildings[i] = b
	}
	return buildings, nil
}
