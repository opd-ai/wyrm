//go:build noebiten

// Package adapters provides V-Series integration for Wyrm.
// This file provides stub implementations for headless testing.
package adapters

import "math/rand"

// ItemAdapter wraps Venture's item generator.
// Stub implementation for headless testing.
type ItemAdapter struct{}

// NewItemAdapter creates a new item adapter.
func NewItemAdapter() *ItemAdapter { return &ItemAdapter{} }

// ItemData holds generated item information.
type ItemData struct {
	ID          string
	Name        string
	Type        string
	SubType     string
	Rarity      string
	Description string
	Tags        []string
	Stats       ItemStatsData
	Seed        int64
	Value       int
	Weight      float64
	Equippable  bool
	Consumable  bool
}

// ItemStatsData holds item stat modifiers.
type ItemStatsData struct {
	Damage             int
	Defense            int
	AttackSpeed        float64
	Value              int
	Weight             float64
	RequiredLevel      int
	DurabilityMax      int
	Durability         int
	IsProjectile       bool
	ProjectileSpeed    float64
	ProjectileLifetime float64
}

// GenerateItems creates multiple items.
func (a *ItemAdapter) GenerateItems(seed int64, genre string, depth, count int, itemType string) ([]*ItemData, error) {
	items := make([]*ItemData, count)
	for i := 0; i < count; i++ {
		items[i], _ = a.GenerateSingleItem(seed+int64(i), genre, depth, itemType)
	}
	return items, nil
}

// GenerateSingleItem creates a single item.
func (a *ItemAdapter) GenerateSingleItem(seed int64, genre string, depth int, itemType string) (*ItemData, error) {
	rng := rand.New(rand.NewSource(seed))
	rarities := []string{"common", "uncommon", "rare", "epic", "legendary"}
	return &ItemData{
		ID:         "item_" + string(rune('A'+rng.Intn(26))),
		Name:       itemType + " item",
		Type:       itemType,
		SubType:    "general",
		Rarity:     rarities[rng.Intn(len(rarities))],
		Value:      10 + rng.Intn(100)*depth,
		Weight:     0.5 + rng.Float64()*5,
		Stats:      ItemStatsData{Damage: rng.Intn(20), Defense: rng.Intn(10), Value: 10 + rng.Intn(100), Weight: 1.0 + rng.Float64()*5},
		Equippable: itemType == "weapon" || itemType == "armor",
		Consumable: itemType == "potion" || itemType == "food",
		Seed:       seed,
	}, nil
}

// GenerateLootTable creates a loot table for a location.
func (a *ItemAdapter) GenerateLootTable(seed int64, genre string, depth int) ([]*ItemData, error) {
	return a.GenerateItems(seed, genre, depth, 5+depth/2, "loot")
}

// GenerateShopInventory creates shop inventory based on wealth.
func (a *ItemAdapter) GenerateShopInventory(seed int64, genre string, cityWealth int) ([]*ItemData, error) {
	return a.GenerateItems(seed, genre, cityWealth/10, 10+cityWealth/5, "shop")
}

// GetItemValue returns adjusted item value.
func GetItemValue(data *ItemData, conditionPercent float64) int {
	baseValue := data.Stats.Value
	return int(float64(baseValue) * conditionPercent)
}

// IsItemEquippable checks if item can be equipped.
func IsItemEquippable(data *ItemData) bool {
	return data.Type == "weapon" || data.Type == "armor" || data.Type == "accessory"
}

// IsItemConsumable checks if item can be consumed.
func IsItemConsumable(data *ItemData) bool {
	return data.Type == "consumable"
}
