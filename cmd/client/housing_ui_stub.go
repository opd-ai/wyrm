//go:build noebiten

// Stub for housing UI when building without Ebiten.
package main

import "github.com/opd-ai/wyrm/pkg/world/housing"

// HousingUI stub for testing.
type HousingUI struct{}

// HousingUIMode stub.
type HousingUIMode int

const (
	HousingModeNone HousingUIMode = iota
	HousingModeListings
	HousingModeMyHouses
	HousingModeFurniture
	HousingModeGuild
)

// NewHousingUI stub.
func NewHousingUI() *HousingUI { return &HousingUI{} }

// Initialize stub.
func (ui *HousingUI) Initialize() {}

// SetPlayerState stub.
func (ui *HousingUI) SetPlayerState(entity uint64, gold, day int) {}

// IsActive stub.
func (ui *HousingUI) IsActive() bool { return false }

// Open stub.
func (ui *HousingUI) Open() {}

// OpenFurnitureMode stub.
func (ui *HousingUI) OpenFurnitureMode(houseID string) {}

// Close stub.
func (ui *HousingUI) Close() {}

// Update stub.
func (ui *HousingUI) Update() {}

// UpdateFurniturePreview stub.
func (ui *HousingUI) UpdateFurniturePreview(playerX, playerZ, viewAngle float64) {}

// GetHouseManager stub.
func (ui *HousingUI) GetHouseManager() *housing.HouseManager {
	return housing.NewHouseManager()
}

// GetGuildManager stub.
func (ui *HousingUI) GetGuildManager() *housing.GuildManager {
	return housing.NewGuildManager()
}

// IsInPlayerHouse stub.
func (ui *HousingUI) IsInPlayerHouse(playerEntity uint64, x, z float64) bool { return false }

// GetNearbyHouseForPurchase stub.
func (ui *HousingUI) GetNearbyHouseForPurchase(x, z float64) *housing.PropertyListing { return nil }

// getPlayerGold stub.
func getPlayerGold(g *Game) int { return 0 }
