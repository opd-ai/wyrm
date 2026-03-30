//go:build !noebiten

// Package adapters provides V-Series integration for Wyrm.
package adapters

import (
	"fmt"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/item"
)

// ItemAdapter wraps Venture's item generator for Wyrm's inventory system.
type ItemAdapter struct {
	generator *item.ItemGenerator
}

// NewItemAdapter creates a new item adapter.
func NewItemAdapter() *ItemAdapter {
	return &ItemAdapter{
		generator: item.NewItemGenerator(),
	}
}

// ItemData holds item information adapted for Wyrm's inventory.
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
}

// ItemStatsData holds item statistics adapted for Wyrm.
type ItemStatsData struct {
	Damage        int
	Defense       int
	AttackSpeed   float64
	Value         int
	Weight        float64
	RequiredLevel int
	DurabilityMax int
	Durability    int
	// Projectile properties
	IsProjectile       bool
	ProjectileSpeed    float64
	ProjectileLifetime float64
	ProjectileType     string
	Pierce             int
	Bounce             int
	Explosive          bool
	ExplosionRadius    float64
}

// GenerateItems generates a batch of items for loot drops or shops.
func (a *ItemAdapter) GenerateItems(seed int64, genre string, depth, count int, itemType string) ([]*ItemData, error) {
	params := procgen.GenerationParams{
		GenreID:    mapGenreID(genre),
		Difficulty: float64(depth) / DepthToDifficultyDivisor,
		Depth:      depth,
		Custom: map[string]interface{}{
			"count": count,
			"type":  itemType,
		},
	}

	result, err := a.generator.Generate(seed, params)
	if err != nil {
		return nil, fmt.Errorf("item generation failed: %w", err)
	}

	items, ok := result.([]*item.Item)
	if !ok {
		return nil, fmt.Errorf("invalid item result type: expected []*item.Item, got %T", result)
	}

	return convertItems(items), nil
}

// GenerateSingleItem generates a single item for a specific purpose.
func (a *ItemAdapter) GenerateSingleItem(seed int64, genre string, depth int, itemType string) (*ItemData, error) {
	items, err := a.GenerateItems(seed, genre, depth, 1, itemType)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("no item generated")
	}
	return items[0], nil
}

// GenerateLootTable generates a loot table for a chest or enemy drop.
func (a *ItemAdapter) GenerateLootTable(seed int64, genre string, depth int) ([]*ItemData, error) {
	// Generate a mix of item types for a typical loot drop
	var allItems []*ItemData

	// Weapons
	weapons, err := a.GenerateItems(seed, genre, depth, LootTableWeaponCount, "weapon")
	if err == nil {
		allItems = append(allItems, weapons...)
	}

	// Armor
	armor, err := a.GenerateItems(seed+1, genre, depth, LootTableArmorCount, "armor")
	if err == nil {
		allItems = append(allItems, armor...)
	}

	// Consumables
	consumables, err := a.GenerateItems(seed+2, genre, depth, LootTableConsumableCount, "consumable")
	if err == nil {
		allItems = append(allItems, consumables...)
	}

	return allItems, nil
}

// GenerateShopInventory generates items for a vendor's shop.
func (a *ItemAdapter) GenerateShopInventory(seed int64, genre string, cityWealth int) ([]*ItemData, error) {
	// Depth scales with city wealth
	depth := cityWealth / ShopWealthDivisor
	if depth < MinShopDepth {
		depth = MinShopDepth
	}
	if depth > MaxShopDepth {
		depth = MaxShopDepth
	}

	// Generate more items for shops
	return a.GenerateItems(seed, genre, depth, ShopInventorySize, "")
}

// convertItems converts Venture items to Wyrm format.
func convertItems(items []*item.Item) []*ItemData {
	result := make([]*ItemData, len(items))
	for i, it := range items {
		result[i] = convertItem(it)
	}
	return result
}

// convertItem converts a single Venture item to Wyrm format.
func convertItem(it *item.Item) *ItemData {
	return &ItemData{
		ID:          it.ID,
		Name:        it.Name,
		Type:        it.Type.String(),
		SubType:     getSubType(it),
		Rarity:      it.Rarity.String(),
		Description: it.Description,
		Tags:        it.Tags,
		Stats: ItemStatsData{
			Damage:             it.Stats.Damage,
			Defense:            it.Stats.Defense,
			AttackSpeed:        it.Stats.AttackSpeed,
			Value:              it.Stats.Value,
			Weight:             it.Stats.Weight,
			RequiredLevel:      it.Stats.RequiredLevel,
			DurabilityMax:      it.Stats.DurabilityMax,
			Durability:         it.Stats.Durability,
			IsProjectile:       it.Stats.IsProjectile,
			ProjectileSpeed:    it.Stats.ProjectileSpeed,
			ProjectileLifetime: it.Stats.ProjectileLifetime,
			ProjectileType:     it.Stats.ProjectileType,
			Pierce:             it.Stats.Pierce,
			Bounce:             it.Stats.Bounce,
			Explosive:          it.Stats.Explosive,
			ExplosionRadius:    it.Stats.ExplosionRadius,
		},
		Seed: it.Seed,
	}
}

// getSubType returns the specific subtype for an item.
func getSubType(it *item.Item) string {
	switch it.Type {
	case item.TypeWeapon:
		return it.WeaponType.String()
	case item.TypeArmor:
		return it.ArmorType.String()
	case item.TypeConsumable:
		return it.ConsumableType.String()
	default:
		return ""
	}
}

// GetItemValue calculates the value of an item considering condition.
func GetItemValue(data *ItemData, conditionPercent float64) int {
	baseValue := data.Stats.Value
	return int(float64(baseValue) * conditionPercent)
}

// IsItemEquippable returns true if the item can be equipped.
func IsItemEquippable(data *ItemData) bool {
	return data.Type == "weapon" || data.Type == "armor" || data.Type == "accessory"
}

// IsItemConsumable returns true if the item is a consumable.
func IsItemConsumable(data *ItemData) bool {
	return data.Type == "consumable"
}
