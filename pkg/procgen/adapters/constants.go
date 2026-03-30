// Package adapters provides V-Series integration for Wyrm.
// This file contains shared constants to reduce magic numbers throughout the adapter implementations.
package adapters

// NPC generation constants.
const (
	// BaseNPCHealth is the minimum health for generated NPCs.
	BaseNPCHealth = 50.0
	// NPCHealthVariance is the random health range added to base.
	NPCHealthVariance = 100
	// NPCSeedMultiplier is used to derive unique seeds per NPC in batch spawns.
	NPCSeedMultiplier = 1000
	// DefaultNPCDepthDivisor divides NPC index to calculate depth.
	DefaultNPCDepthDivisor = 10
	// NPCGridColumns is the number of columns in the NPC spawn grid.
	NPCGridColumns = 5
	// NPCGridOffset centers the grid around the spawn point.
	NPCGridOffset = 2
	// NPCGridSpacingDivisor divides radius for grid cell spacing.
	NPCGridSpacingDivisor = 5
)

// Faction reputation constants.
const (
	// DefaultFactionReputation is the starting reputation with a faction.
	DefaultFactionReputation = 0
)

// Dialog word count limits.
const (
	// DefaultDialogMaxWords is the maximum words for a dialog line.
	DefaultDialogMaxWords = 50
	// DefaultDialogMinWords is the minimum words for a dialog line.
	DefaultDialogMinWords = 5
	// GreetingMaxWords is the maximum words for a greeting.
	GreetingMaxWords = 20
	// GreetingMinWords is the minimum words for a greeting.
	GreetingMinWords = 3
	// DialogLineSeedMultiplier is used to derive unique seeds per dialog line.
	DialogLineSeedMultiplier = 1000
)

// Room furniture counts for house generation.
const (
	// LivingRoomFurnitureCount is the default furniture count for living rooms.
	LivingRoomFurnitureCount = 5
	// BedroomFurnitureCount is the default furniture count for bedrooms.
	BedroomFurnitureCount = 4
	// KitchenFurnitureCount is the default furniture count for kitchens.
	KitchenFurnitureCount = 4
	// StorageFurnitureCount is the default furniture count for storage rooms.
	StorageFurnitureCount = 5
)

// Room seed offsets for deterministic generation.
const (
	// BedroomSeedOffset is added to base seed for bedroom generation.
	BedroomSeedOffset = 100
	// KitchenSeedOffset is added to base seed for kitchen generation.
	KitchenSeedOffset = 200
	// StorageSeedOffset is added to base seed for storage generation.
	StorageSeedOffset = 300
)

// Furniture value constants.
const (
	// BaseFurnitureValue is the base value for any furniture piece.
	BaseFurnitureValue = 10
	// WoodMaterialMultiplier is the value multiplier for wood furniture.
	WoodMaterialMultiplier = 1
	// MetalMaterialMultiplier is the value multiplier for metal furniture.
	MetalMaterialMultiplier = 2
	// StoneMaterialMultiplier is the value multiplier for stone furniture.
	StoneMaterialMultiplier = 2
	// CrystalMaterialMultiplier is the value multiplier for crystal furniture.
	CrystalMaterialMultiplier = 5
	// FabricMaterialMultiplier is the value multiplier for fabric furniture.
	FabricMaterialMultiplier = 1
)

// Furniture rarity multipliers.
const (
	// CommonRarityMultiplier is the value multiplier for common items.
	CommonRarityMultiplier = 1
	// UncommonRarityMultiplier is the value multiplier for uncommon items.
	UncommonRarityMultiplier = 2
	// RareRarityMultiplier is the value multiplier for rare items.
	RareRarityMultiplier = 5
	// EpicRarityMultiplier is the value multiplier for epic items.
	EpicRarityMultiplier = 10
	// LegendaryRarityMultiplier is the value multiplier for legendary items.
	LegendaryRarityMultiplier = 25
)

// Vehicle rarity stat multipliers.
const (
	// VehicleCommonStatMultiplier is the stat multiplier for common vehicles.
	VehicleCommonStatMultiplier = 1.0
	// VehicleUncommonStatMultiplier is the stat multiplier for uncommon vehicles.
	VehicleUncommonStatMultiplier = 1.2
	// VehicleRareStatMultiplier is the stat multiplier for rare vehicles.
	VehicleRareStatMultiplier = 1.5
	// VehicleEpicStatMultiplier is the stat multiplier for epic vehicles.
	VehicleEpicStatMultiplier = 2.0
	// VehicleLegendaryStatMultiplier is the stat multiplier for legendary vehicles.
	VehicleLegendaryStatMultiplier = 3.0
)

// Generation difficulty and depth constants.
const (
	// DefaultGenerationDifficulty is the default difficulty for Venture generators.
	DefaultGenerationDifficulty = 0.5
	// DefaultGenerationDepth is the default depth for nested generation.
	DefaultGenerationDepth = 1
)

// Terrain generation constants.
const (
	// DefaultChunkSize is the default size for terrain chunks.
	DefaultChunkSize = 512
	// DefaultOctaves is the number of noise octaves for terrain.
	DefaultOctaves = 6
	// DefaultPersistence controls noise detail falloff.
	DefaultPersistence = 0.5
	// DefaultLacunarity controls noise frequency scaling.
	DefaultLacunarity = 2.0
	// DefaultScale controls overall terrain feature size.
	DefaultScale = 100.0
)

// Biome temperature and humidity thresholds.
const (
	// TundraTemperatureMax is the maximum temperature for tundra biome.
	TundraTemperatureMax = 0.2
	// BorealTemperatureMax is the maximum temperature for boreal biome.
	BorealTemperatureMax = 0.35
	// TemperateTemperatureMin is the minimum temperature for temperate biome.
	TemperateTemperatureMin = 0.35
	// TemperateTemperatureMax is the maximum temperature for temperate biome.
	TemperateTemperatureMax = 0.65
	// SubtropicalTemperatureMin is the minimum temperature for subtropical biome.
	SubtropicalTemperatureMin = 0.65
	// SubtropicalTemperatureMax is the maximum temperature for subtropical biome.
	SubtropicalTemperatureMax = 0.8
	// TropicalTemperatureMin is the minimum temperature for tropical biome.
	TropicalTemperatureMin = 0.8

	// DesertHumidityMax is the maximum humidity for desert biome.
	DesertHumidityMax = 0.2
	// SavannaHumidityMax is the maximum humidity for savanna biome.
	SavannaHumidityMax = 0.4
	// GrasslandHumidityMax is the maximum humidity for grassland biome.
	GrasslandHumidityMax = 0.6
	// ForestHumidityMin is the minimum humidity for forest biome.
	ForestHumidityMin = 0.6
	// SwampHumidityMin is the minimum humidity for swamp biome.
	SwampHumidityMin = 0.85
)

// Environment detail generation constants.
const (
	// TreeSpawnChance is the probability of spawning a tree in appropriate biomes.
	TreeSpawnChance = 0.3
	// RockSpawnChance is the probability of spawning a rock in appropriate biomes.
	RockSpawnChance = 0.15
	// FlowerSpawnChance is the probability of spawning flowers.
	FlowerSpawnChance = 0.2
	// MushroomSpawnChance is the probability of spawning mushrooms.
	MushroomSpawnChance = 0.1
	// CrystalSpawnChance is the probability of spawning crystals (caves).
	CrystalSpawnChance = 0.05
)

// Vehicle archetype stat multipliers.
const (
	// HorseBaseSpeed is the base speed for horse vehicles.
	HorseBaseSpeed = 15.0
	// HorseBaseFuel is the base stamina for horse vehicles.
	HorseBaseFuel = 100.0
	// CartBaseSpeed is the base speed for cart vehicles.
	CartBaseSpeed = 10.0
	// CartBaseFuel is the base fuel for cart vehicles.
	CartBaseFuel = 50.0
	// BoatBaseSpeed is the base speed for boat vehicles.
	BoatBaseSpeed = 8.0
	// BoatBaseFuel is the base fuel/wind for boat vehicles.
	BoatBaseFuel = 75.0
)

// Dialog and NPC interaction constants.
const (
	// DefaultSentimentNeutral is the neutral sentiment value.
	DefaultSentimentNeutral = 0.5
	// MinimumSentiment is the minimum sentiment value.
	MinimumSentiment = 0.0
	// MaximumSentiment is the maximum sentiment value.
	MaximumSentiment = 1.0
	// SentimentChangeRate is how quickly sentiment changes per interaction.
	SentimentChangeRate = 0.1
)

// Quest generation constants.
const (
	// DefaultQuestRewardGold is the base gold reward for quests.
	DefaultQuestRewardGold = 100
	// QuestRewardDepthMultiplier scales rewards by quest depth.
	QuestRewardDepthMultiplier = 50
	// QuestDifficultyXPMultiplier scales XP by difficulty.
	QuestDifficultyXPMultiplier = 25.0
)

// Magic spell generation constants.
const (
	// BaseSpellManaCost is the minimum mana cost for spells.
	BaseSpellManaCost = 10.0
	// SpellManaCostPerLevel scales mana cost with spell level.
	SpellManaCostPerLevel = 5.0
	// BaseSpellDamage is the minimum damage for damage spells.
	BaseSpellDamage = 15.0
	// SpellDamagePerLevel scales damage with spell level.
	SpellDamagePerLevel = 10.0
	// BaseSpellCooldown is the minimum cooldown for spells.
	BaseSpellCooldown = 2.0
)

// Item generation constants.
const (
	// BaseWeaponDamage is the minimum damage for weapons.
	BaseWeaponDamage = 5.0
	// WeaponDamagePerTier scales damage by item tier.
	WeaponDamagePerTier = 10.0
	// BaseArmorDefense is the minimum defense for armor.
	BaseArmorDefense = 2.0
	// ArmorDefensePerTier scales defense by item tier.
	ArmorDefensePerTier = 5.0
	// ConsumableHealAmount is the base healing for consumables.
	ConsumableHealAmount = 25.0
)

// Recipe crafting constants.
const (
	// BaseCraftingTime is the minimum time to craft an item (seconds).
	BaseCraftingTime = 5.0
	// CraftingTimePerComplexity scales crafting time with recipe complexity.
	CraftingTimePerComplexity = 2.0
	// MinimumMaterialCount is the minimum materials for any recipe.
	MinimumMaterialCount = 1
	// MaximumMaterialCount is the maximum materials for complex recipes.
	MaximumMaterialCount = 8
)

// Puzzle generation constants.
const (
	// MinPuzzleSteps is the minimum steps in a puzzle.
	MinPuzzleSteps = 3
	// MaxPuzzleSteps is the maximum steps in a puzzle.
	MaxPuzzleSteps = 10
	// PuzzleStepsPerDifficulty scales puzzle length with difficulty.
	PuzzleStepsPerDifficulty = 2
	// BasePuzzleReward is the base XP reward for solving puzzles.
	BasePuzzleReward = 50
)

// Building generation constants.
const (
	// MinBuildingWidth is the minimum width for generated buildings.
	MinBuildingWidth = 5
	// MaxBuildingWidth is the maximum width for generated buildings.
	MaxBuildingWidth = 20
	// MinBuildingHeight is the minimum height for generated buildings.
	MinBuildingHeight = 3
	// MaxBuildingHeight is the maximum height for generated buildings.
	MaxBuildingHeight = 5
	// MinBuildingDepth is the minimum depth for generated buildings.
	MinBuildingDepth = 5
	// MaxBuildingDepth is the maximum depth for generated buildings.
	MaxBuildingDepth = 15
)

// Skill generation constants.
const (
	// BaseSkillXPRequired is the XP needed for level 1.
	BaseSkillXPRequired = 100
	// SkillXPScaling is the XP increase per level.
	SkillXPScaling = 1.1
	// MaxSkillLevel is the maximum skill level.
	MaxSkillLevel = 100
	// SkillsPerSchool is the number of skills in each school.
	SkillsPerSchool = 5
)

// Narrative arc constants.
const (
	// MinArcStages is the minimum stages in a narrative arc.
	MinArcStages = 3
	// MaxArcStages is the maximum stages in a narrative arc.
	MaxArcStages = 7
	// ArcStageDurationBase is the base duration per stage (game hours).
	ArcStageDurationBase = 2
)
