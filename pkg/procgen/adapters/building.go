// Package adapters provides V-Series integration for Wyrm.
package adapters

import (
	"fmt"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/building"
)

// BuildingAdapter wraps Venture's building generator for Wyrm's city structures.
type BuildingAdapter struct {
	generator *building.Generator
}

// NewBuildingAdapter creates a new building adapter.
func NewBuildingAdapter() *BuildingAdapter {
	return &BuildingAdapter{
		generator: building.NewGenerator(),
	}
}

// BuildingData holds generated building data for Wyrm integration.
type BuildingData struct {
	Type     string
	Style    string
	GenreID  string
	Width    int
	Height   int
	Floors   int
	Rooms    []BuildingRoomData
	Doors    []BuildingDoorData
	Windows  []BuildingWindowData
	RoofType string
}

// BuildingRoomData holds room information from generated buildings.
type BuildingRoomData struct {
	X, Y          int
	Width, Height int
	Type          string
}

// BuildingDoorData holds door information from generated buildings.
type BuildingDoorData struct {
	X, Y int
	Type string
}

// BuildingWindowData holds window information from generated buildings.
type BuildingWindowData struct {
	X, Y int
	Type string
}

// GenerateBuilding creates a building using Venture's generator.
func (a *BuildingAdapter) GenerateBuilding(seed int64, genre string, buildingType, floors int) (*BuildingData, error) {
	params := procgen.GenerationParams{
		GenreID:    mapGenreID(genre),
		Difficulty: 0.5,
		Custom: map[string]interface{}{
			"buildingType": building.BuildingType(buildingType),
			"floors":       floors,
		},
	}

	result, err := a.generator.Generate(seed, params)
	if err != nil {
		return nil, fmt.Errorf("building generation failed: %w", err)
	}

	ventureBuilding, ok := result.(*building.Building)
	if !ok {
		return nil, fmt.Errorf("invalid building result type: expected *building.Building, got %T", result)
	}

	return convertBuilding(ventureBuilding), nil
}

// convertBuilding transforms a Venture building to Wyrm format.
func convertBuilding(b *building.Building) *BuildingData {
	rooms := make([]BuildingRoomData, len(b.Rooms))
	for i, room := range b.Rooms {
		rooms[i] = BuildingRoomData{
			X:      room.X,
			Y:      room.Y,
			Width:  room.Width,
			Height: room.Height,
			Type:   room.Type.String(),
		}
	}

	doors := make([]BuildingDoorData, len(b.Doors))
	for i, door := range b.Doors {
		doors[i] = BuildingDoorData{
			X:    door.X,
			Y:    door.Y,
			Type: door.Type.String(),
		}
	}

	windows := make([]BuildingWindowData, len(b.Windows))
	for i, window := range b.Windows {
		windows[i] = BuildingWindowData{
			X:    window.X,
			Y:    window.Y,
			Type: window.Type.String(),
		}
	}

	return &BuildingData{
		Type:     b.Type.String(),
		Style:    b.Style.String(),
		GenreID:  b.GenreID,
		Width:    b.Width,
		Height:   b.Height,
		Floors:   b.Floors,
		Rooms:    rooms,
		Doors:    doors,
		Windows:  windows,
		RoofType: b.RoofType.String(),
	}
}

// GenerateHouse creates a house building.
func (a *BuildingAdapter) GenerateHouse(seed int64, genre string) (*BuildingData, error) {
	return a.GenerateBuilding(seed, genre, int(building.TypeHouse), 1)
}

// GenerateWorkshop creates a workshop building.
func (a *BuildingAdapter) GenerateWorkshop(seed int64, genre string) (*BuildingData, error) {
	return a.GenerateBuilding(seed, genre, int(building.TypeWorkshop), 1)
}

// GenerateTower creates a tower building with multiple floors.
func (a *BuildingAdapter) GenerateTower(seed int64, genre string) (*BuildingData, error) {
	return a.GenerateBuilding(seed, genre, int(building.TypeTower), 3)
}

// GenerateManor creates a manor building with multiple floors.
func (a *BuildingAdapter) GenerateManor(seed int64, genre string) (*BuildingData, error) {
	return a.GenerateBuilding(seed, genre, int(building.TypeManor), 2)
}

// GenerateGuildHall creates a guild hall building.
func (a *BuildingAdapter) GenerateGuildHall(seed int64, genre string) (*BuildingData, error) {
	return a.GenerateBuilding(seed, genre, int(building.TypeGuildHall), 2)
}

// GenerateCityBuildings generates multiple buildings for a city.
func (a *BuildingAdapter) GenerateCityBuildings(seed int64, genre string, count int) ([]*BuildingData, error) {
	buildings := make([]*BuildingData, 0, count)

	for i := 0; i < count; i++ {
		buildingSeed := seed + int64(i)*1000
		// Cycle through building types
		buildingType := i % 6
		floors := 1
		if buildingType == int(building.TypeTower) {
			floors = 3
		} else if buildingType == int(building.TypeManor) || buildingType == int(building.TypeGuildHall) {
			floors = 2
		}

		b, err := a.GenerateBuilding(buildingSeed, genre, buildingType, floors)
		if err != nil {
			continue
		}
		buildings = append(buildings, b)
	}

	return buildings, nil
}

// BuildingTypeID constants for convenience.
const (
	BuildingTypeHouse     = int(building.TypeHouse)
	BuildingTypeWorkshop  = int(building.TypeWorkshop)
	BuildingTypeStorage   = int(building.TypeStorage)
	BuildingTypeTower     = int(building.TypeTower)
	BuildingTypeManor     = int(building.TypeManor)
	BuildingTypeGuildHall = int(building.TypeGuildHall)
)
