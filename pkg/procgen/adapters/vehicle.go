//go:build !noebiten

// Package adapters provides V-Series integration for Wyrm.
package adapters

import (
	"fmt"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/vehicle"
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// VehicleAdapter wraps Venture's vehicle generator for Wyrm integration.
type VehicleAdapter struct {
	generator *vehicle.VehicleGenerator
}

// NewVehicleAdapter creates a new vehicle adapter.
func NewVehicleAdapter() *VehicleAdapter {
	return &VehicleAdapter{
		generator: vehicle.NewVehicleGenerator(),
	}
}

// VehicleData holds generated vehicle data for Wyrm integration.
type VehicleData struct {
	Name           string
	Description    string
	VehicleType    string
	Rarity         string
	MaxSpeed       float64
	Acceleration   float64
	Handling       float64
	MaxDurability  float64
	FuelCapacity   float64
	Capacity       int
	FuelType       string
	GenreID        string
	Color          uint32
	HasCombat      bool
	HasWeapon      bool
	WeaponType     string
	CargoSlots     int
	CargoWeight    float64
	UpgradeSlots   int
	SpecialAbility string
	Decorations    []string
	DamageState    float64
	SecondaryColor uint32
	DecalPattern   string
}

// GenerateVehicles creates vehicles using Venture's generator.
func (a *VehicleAdapter) GenerateVehicles(seed int64, genre string, count int) ([]*VehicleData, error) {
	params := procgen.GenerationParams{
		GenreID:    mapGenreID(genre),
		Difficulty: DefaultGenerationDifficulty,
		Custom: map[string]interface{}{
			"count": count,
		},
	}

	result, err := a.generator.Generate(seed, params)
	if err != nil {
		return nil, fmt.Errorf("vehicle generation failed: %w", err)
	}

	vehicles, ok := result.([]*vehicle.Vehicle)
	if !ok {
		return nil, fmt.Errorf("invalid vehicle result type: expected []*vehicle.Vehicle, got %T", result)
	}

	vehicleData := make([]*VehicleData, len(vehicles))
	for i, v := range vehicles {
		vehicleData[i] = convertVehicle(v)
	}

	return vehicleData, nil
}

// GenerateVehicle creates a single vehicle.
func (a *VehicleAdapter) GenerateVehicle(seed int64, genre string) (*VehicleData, error) {
	vehicles, err := a.GenerateVehicles(seed, genre, 1)
	if err != nil {
		return nil, err
	}
	if len(vehicles) == 0 {
		return nil, fmt.Errorf("no vehicle generated")
	}
	return vehicles[0], nil
}

// convertVehicle transforms a Venture vehicle to Wyrm format.
func convertVehicle(v *vehicle.Vehicle) *VehicleData {
	decorations := make([]string, len(v.Decorations))
	copy(decorations, v.Decorations)

	return &VehicleData{
		Name:           v.Name,
		Description:    v.Description,
		VehicleType:    v.VehicleType.String(),
		Rarity:         v.Rarity.String(),
		MaxSpeed:       v.MaxSpeed,
		Acceleration:   v.Acceleration,
		Handling:       v.Handling,
		MaxDurability:  v.MaxDurability,
		FuelCapacity:   v.FuelCapacity,
		Capacity:       v.Capacity,
		FuelType:       v.FuelType,
		GenreID:        v.GenreID,
		Color:          v.Color,
		HasCombat:      v.HasCombat,
		HasWeapon:      v.HasWeapon,
		WeaponType:     v.WeaponType,
		CargoSlots:     v.CargoSlots,
		CargoWeight:    v.CargoWeight,
		UpgradeSlots:   v.UpgradeSlots,
		SpecialAbility: v.SpecialAbility,
		Decorations:    decorations,
		DamageState:    v.DamageState,
		SecondaryColor: v.SecondaryColor,
		DecalPattern:   v.DecalPattern,
	}
}

// SpawnVehicleEntity creates a vehicle entity in the ECS world.
func SpawnVehicleEntity(world *ecs.World, v *VehicleData, x, y, z float64) ecs.Entity {
	entity := world.CreateEntity()

	_ = world.AddComponent(entity, &components.Position{X: x, Y: y, Z: z})
	_ = world.AddComponent(entity, &components.Vehicle{
		VehicleType: v.VehicleType,
		Speed:       0,
		Fuel:        v.FuelCapacity,
		Direction:   0,
	})

	return entity
}

// VehicleRarityMultiplier returns the stat multiplier for a given rarity.
func VehicleRarityMultiplier(rarity string) float64 {
	switch rarity {
	case "Common":
		return VehicleCommonStatMultiplier
	case "Uncommon":
		return VehicleUncommonStatMultiplier
	case "Rare":
		return VehicleRareStatMultiplier
	case "Epic":
		return VehicleEpicStatMultiplier
	case "Legendary":
		return VehicleLegendaryStatMultiplier
	default:
		return VehicleCommonStatMultiplier
	}
}
