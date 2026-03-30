//go:build noebiten

// Package adapters provides V-Series integration for Wyrm.
// This file provides stub implementations for headless testing.
package adapters

import (
	"math/rand"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// VehicleAdapter wraps Venture's vehicle generator.
// Stub implementation for headless testing.
type VehicleAdapter struct{}

// NewVehicleAdapter creates a new vehicle adapter.
func NewVehicleAdapter() *VehicleAdapter { return &VehicleAdapter{} }

// VehicleData holds generated vehicle information.
type VehicleData struct {
	Name          string
	Description   string
	VehicleType   string
	Rarity        string
	MaxSpeed      float64
	Acceleration  float64
	Handling      float64
	MaxDurability float64
	FuelCapacity  float64
	Capacity      int
	FuelType      string
	GenreID       string
	Color         uint32
	HasCombat     bool
	HasWeapon     bool
}

// GenerateVehicles creates multiple vehicles.
func (a *VehicleAdapter) GenerateVehicles(seed int64, genre string, count int) ([]*VehicleData, error) {
	vehicles := make([]*VehicleData, count)
	for i := 0; i < count; i++ {
		vehicles[i], _ = a.GenerateVehicle(seed+int64(i), genre)
	}
	return vehicles, nil
}

// GenerateVehicle creates a single vehicle.
func (a *VehicleAdapter) GenerateVehicle(seed int64, genre string) (*VehicleData, error) {
	rng := rand.New(rand.NewSource(seed))
	types := map[string][]string{
		"fantasy":          {"horse", "cart", "ship"},
		"sci-fi":           {"hover-bike", "shuttle", "mech"},
		"horror":           {"bone-cart", "hearse", "barge"},
		"cyberpunk":        {"motorbike", "APC", "drone"},
		"post-apocalyptic": {"buggy", "bus", "gyrocopter"},
	}
	genreTypes := types[genre]
	if genreTypes == nil {
		genreTypes = types["fantasy"]
	}
	vehicleType := genreTypes[rng.Intn(len(genreTypes))]
	rarities := []string{"common", "uncommon", "rare", "legendary"}

	return &VehicleData{
		Name:          genre + " " + vehicleType,
		Description:   "A " + vehicleType + " vehicle",
		VehicleType:   vehicleType,
		Rarity:        rarities[rng.Intn(len(rarities))],
		MaxSpeed:      10 + rng.Float64()*40,
		Acceleration:  5 + rng.Float64()*20,
		Handling:      0.5 + rng.Float64()*0.5,
		MaxDurability: 100 + rng.Float64()*200,
		FuelCapacity:  50 + rng.Float64()*150,
		Capacity:      1 + rng.Intn(4),
		FuelType:      "standard",
		GenreID:       genre,
		Color:         uint32(rng.Uint32()),
		HasCombat:     rng.Float64() > 0.5,
		HasWeapon:     rng.Float64() > 0.7,
	}, nil
}

// SpawnVehicleEntity creates a vehicle entity in the ECS world.
func SpawnVehicleEntity(world *ecs.World, v *VehicleData, x, y, z float64) ecs.Entity {
	e := world.CreateEntity()
	world.AddComponent(e, &components.Position{X: x, Y: y, Z: z})
	world.AddComponent(e, &components.Vehicle{
		VehicleType: v.VehicleType,
		Speed:       v.MaxSpeed,
		Fuel:        v.FuelCapacity,
		Direction:   0,
	})
	return e
}

// mapVehicleType converts vehicle type string to enum.
func mapVehicleType(vehicleType string) int {
	switch vehicleType {
	case "Mount":
		return 0
	case "Cart":
		return 1
	case "Boat":
		return 2
	case "Glider":
		return 3
	case "Mech":
		return 4
	default:
		return 0
	}
}

// VehicleRarityMultiplier returns rarity-based stat multiplier.
func VehicleRarityMultiplier(rarity string) float64 {
	switch rarity {
	case "legendary":
		return 2.0
	case "rare":
		return 1.5
	case "uncommon":
		return 1.25
	default:
		return 1.0
	}
}
