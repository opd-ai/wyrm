// Package systems provides ECS systems for Wyrm.
// This file contains shared constants to reduce magic numbers throughout the system implementations.
package systems

// Physics constants for movement and collision.
const (
	// DefaultMoveSpeed is the base movement speed for entities (units per second).
	DefaultMoveSpeed = 5.0
	// DefaultTurnSpeed is the base rotation speed (radians per second).
	DefaultTurnSpeed = 3.0
	// GravityAcceleration is the standard gravity for physics (units per second squared).
	GravityAcceleration = 9.8
)

// Combat constants for damage and attack calculations.
const (
	// DefaultMeleeRangeUnits is the default melee attack range in world units.
	DefaultMeleeRangeUnits = 2.0
	// DefaultBaseDamage is the base unarmed damage.
	DefaultBaseDamage = 10.0
	// DefaultAttackCooldown is the time between attacks (seconds).
	DefaultAttackCooldown = 1.0
	// SkillDamageBonus is the damage multiplier per skill level (0.02 = 2% per level).
	SkillDamageBonus = 0.02
	// BackstabDamageMultiplier is the damage multiplier for attacking unaware targets.
	BackstabDamageMultiplier = 2.0
)

// Stealth constants for detection and sneaking.
const (
	// DefaultSneakSpeedReduction is the movement penalty when sneaking (0.5 = 50% slower).
	DefaultSneakSpeedReduction = 0.5
	// DefaultAlertDecayRate is how fast alert levels decay per second.
	DefaultAlertDecayRate = 0.1
	// AlertIncreasePerDetection is the alert level increase when detected.
	AlertIncreasePerDetection = 0.1
	// MaxAlertLevel is the maximum alert level (fully aware).
	MaxAlertLevel = 1.0
	// MinAlertLevel is the minimum alert level (completely unaware).
	MinAlertLevel = 0.0
	// AwarenessThreshold is the alert level below which a target is considered unaware.
	AwarenessThreshold = 0.5
	// PickpocketDifficultyMultiplier converts difficulty to skill requirement.
	PickpocketDifficultyMultiplier = 10.0
)

// Audio constants for spatial audio and ambient sound.
const (
	// DefaultCombatDetectionRange is the distance to detect combat for music intensity.
	DefaultCombatDetectionRange = 50.0
	// DefaultAmbientUpdateInterval is seconds between ambient sound checks.
	DefaultAmbientUpdateInterval = 5.0
	// MaxHostilesForIntensity is the number of hostiles that produces max combat intensity.
	MaxHostilesForIntensity = 10
	// MaxCombatIntensity is the maximum combat intensity value.
	MaxCombatIntensity = 1.0
	// HostileFactionThreshold is the reputation below which an entity is considered hostile.
	HostileFactionThreshold = -50
	// CityProximityRange is the distance within which a location is considered "in city".
	CityProximityRange = 100.0
	// LinearFalloffBase is the base attenuation at zero distance.
	LinearFalloffBase = 1.0
)

// Vehicle constants for movement and fuel consumption.
const (
	// DefaultFuelConsumptionRate is fuel used per unit of speed per second.
	DefaultFuelConsumptionRate = 0.01
	// MinFuelLevel is the minimum fuel level (empty tank).
	MinFuelLevel = 0.0
)

// Faction constants for reputation and politics.
const (
	// DefaultRelationDecayRate is how fast faction relations decay toward neutral per tick.
	DefaultRelationDecayRate = 0.1
	// DefaultReputationPerKill is the reputation change when a faction member is killed.
	DefaultReputationPerKill = -25.0
	// NeutralRelation is the baseline faction relation value.
	NeutralRelation = 0.0
)

// Crime constants for bounty and wanted levels.
const (
	// DefaultWitnessRange is the distance within which NPCs can witness crimes.
	DefaultWitnessRange = 50.0
	// MaxWantedLevel is the maximum wanted level (5 stars).
	MaxWantedLevel = 5
	// MinWantedLevel is the minimum wanted level (not wanted).
	MinWantedLevel = 0
)

// Economy constants for pricing and trading.
const (
	// BasePriceMultiplier is the neutral price multiplier.
	BasePriceMultiplier = 1.0
	// HighDemandPriceMultiplier is the price multiplier when demand exceeds supply.
	HighDemandPriceMultiplier = 2.0
	// MinPriceMultiplier is the minimum allowed price modifier.
	MinPriceMultiplier = 0.5
	// MaxPriceMultiplier is the maximum allowed price modifier.
	MaxPriceMultiplier = 2.0
)

// Weather constants for duration and transitions.
const (
	// DefaultWeatherDuration is the base duration for weather states (seconds).
	DefaultWeatherDuration = 300.0 // 5 minutes
)

// World clock constants for time progression.
const (
	// HoursPerDay is the number of hours in a game day.
	HoursPerDay = 24
	// SecondsPerHour is the real-time seconds per game hour (at 1x speed).
	SecondsPerHour = 60.0
)

// Music constants for adaptive music system.
const (
	// DefaultSampleRate is the audio sample rate in Hz.
	DefaultSampleRate = 44100
	// DefaultCrossfadeDuration is the music transition time (seconds).
	DefaultCrossfadeDuration = 2.0
	// CombatExitDelay is seconds after last enemy death before exiting combat music.
	CombatExitDelay = 5.0
	// CombatMusicReduction is how much to reduce exploration volume during combat.
	CombatMusicReduction = 0.3
	// VolumeThreshold is the minimum volume before considering a layer inactive.
	VolumeThreshold = 0.001

	// Musical frequency constants (Hz) for common notes.
	FreqA1 = 55.0  // A1
	FreqE2 = 82.5  // E2
	FreqA2 = 110.0 // A2
	FreqE3 = 165.0 // E3
	FreqA3 = 220.0 // A3
	FreqE4 = 330.0 // E4
	FreqA4 = 440.0 // A4

	// ADSR envelope constants (seconds).
	DefaultAttackTime  = 0.02
	DefaultReleaseTime = 0.05
)

// Skill progression constants.
const (
	// DefaultXPPerLevel is the base XP required to level up.
	DefaultXPPerLevel = 100.0
	// LevelScalingFactor is the XP increase per level (0.1 = 10% per level).
	LevelScalingFactor = 0.1
)

// Crafting constants for quality, tool efficiency, and gathering.
const (
	// CancelCraftMaterialReturn is the fraction of materials returned when canceling (50%).
	CancelCraftMaterialReturn = 0.5
	// MaterialReturnDivisor is the divisor for calculating material returns.
	MaterialReturnDivisor = 2
	// GatheringSkillQualityBonus is quality bonus per gathering skill level (+0.5% per level).
	GatheringSkillQualityBonus = 0.005
	// CraftingSkillQualityBonus is quality bonus per crafting skill level (+1% per level).
	CraftingSkillQualityBonus = 0.01
	// QualityVarianceFactor is the quality variance multiplier (+/- 10%).
	QualityVarianceFactor = 0.2
	// QualityVarianceCenter is the center of the variance range.
	QualityVarianceCenter = 0.5
	// MinimumQuality is the minimum allowed quality value.
	MinimumQuality = 0.1
	// MaximumQuality is the maximum allowed quality value.
	MaximumQuality = 1.0
	// ZeroQuality is zero quality for calculations.
	ZeroQuality = 0.0
	// BaseGatherAmount is the minimum amount gathered per action.
	BaseGatherAmount = 1
	// DefaultToolDurabilityLoss is durability lost per tool use.
	DefaultToolDurabilityLoss = 1.0
)

// Quality tier thresholds for item rarity.
const (
	// LegendaryQualityThreshold is the minimum quality for Legendary items.
	LegendaryQualityThreshold = 0.95
	// EpicQualityThreshold is the minimum quality for Epic items.
	EpicQualityThreshold = 0.85
	// RareQualityThreshold is the minimum quality for Rare items.
	RareQualityThreshold = 0.70
	// UncommonQualityThreshold is the minimum quality for Uncommon items.
	UncommonQualityThreshold = 0.50
)

// Tool efficiency multipliers based on tier difference.
const (
	// ToolMuchBetterEfficiency is the multiplier when tool is 2+ tiers above resource.
	ToolMuchBetterEfficiency = 2.0
	// ToolSlightlyBetterEfficiency is the multiplier when tool is 1 tier above resource.
	ToolSlightlyBetterEfficiency = 1.5
	// ToolMatchedEfficiency is the multiplier when tool matches resource tier.
	ToolMatchedEfficiency = 1.0
	// ToolSlightlyWorseEfficiency is the multiplier when tool is 1 tier below resource.
	ToolSlightlyWorseEfficiency = 0.5
	// ToolMuchWorseEfficiency is the multiplier when tool is 2+ tiers below resource.
	ToolMuchWorseEfficiency = 0.25
)

// Minigame score thresholds.
const (
	// MinigameExcellentThreshold is the score for excellent performance (90%+).
	MinigameExcellentThreshold = 0.9
	// MinigameGoodThreshold is the score for good performance (70%+).
	MinigameGoodThreshold = 0.7
	// MinigamePoorThreshold is the score below which penalties apply (30%-).
	MinigamePoorThreshold = 0.3
	// MinigamePassThreshold is the minimum score to succeed (50%+).
	MinigamePassThreshold = 0.5
	// MinigameGoodTimingMultiplier is the score multiplier for good (not perfect) timing.
	MinigameGoodTimingMultiplier = 0.5
	// MinigamePrecisionGoodScore is the score for good precision (0.7 of perfect).
	MinigamePrecisionGoodScore = 0.7
	// MinigameSequenceWrongPenalty is the score penalty for wrong sequence input.
	MinigameSequenceWrongPenalty = 0.1
)

// Minigame timing constants.
const (
	// MinigameDefaultMaxAttempts is the default number of minigame steps.
	MinigameDefaultMaxAttempts = 3
	// MinigameSequenceLength is the length of sequence minigames.
	MinigameSequenceLength = 4
	// MinigameTimingWindowDouble is 2x the timing window for "good" hits.
	MinigameTimingWindowDouble = 2
	// MinigamePrecisionWindowHalf is 0.5x the timing window for "perfect" precision.
	MinigamePrecisionWindowHalf = 0.5
)

// Timing minigame target ranges.
const (
	// TimingMinTargetDelay is the minimum delay before next timing target (seconds).
	TimingMinTargetDelay = 1.0
	// TimingMaxTargetVariance is the random variance added to timing target (seconds).
	TimingMaxTargetVariance = 2.0
	// PrecisionMinTargetDelay is the minimum delay for precision targets (seconds).
	PrecisionMinTargetDelay = 0.5
	// PrecisionMaxTargetVariance is the random variance for precision targets (seconds).
	PrecisionMaxTargetVariance = 0.5
)

// Quest generation thresholds for world state conditions.
const (
	// WorldStateHighThreshold is the level at which world state triggers quest generation.
	WorldStateHighThreshold = 0.5
)

// Default skill and progression caps.
const (
	// DefaultLevelCap is the default maximum skill level.
	DefaultLevelCap = 100
	// DefaultTrainingCostBase is the base cost for training sessions.
	DefaultTrainingCostBase = 50.0
	// TrainingCostMultiplier is the multiplier per skill level for training cost.
	TrainingCostMultiplier = 10.0
	// TrainingXPGain is the base XP granted per training session.
	TrainingXPGain = 25.0
	// MaxTrainerLevel is the maximum level a trainer can teach to.
	MaxTrainerLevel = 75
	// TrainingXPFraction is the fraction of XP per level granted during training.
	TrainingXPFraction = 0.1
)

// City event constants for event timing and probabilities.
const (
	// CityEventMinInterval is the minimum hours between automatic events.
	CityEventMinInterval = 24.0
	// CityEventMaxActive is the maximum number of simultaneous events.
	CityEventMaxActive = 3
	// CityEventBaseProbability is the base chance per check for a new event.
	CityEventBaseProbability = 0.15

	// Event duration constants (in game hours).
	CityEventDurationFestival     = 48.0
	CityEventDurationMarket       = 12.0
	CityEventDurationPlague       = 168.0 // 7 days
	CityEventDurationRiot         = 8.0
	CityEventDurationSiege        = 72.0 // 3 days
	CityEventDurationCelebration  = 24.0
	CityEventDurationMartialLaw   = 72.0
	CityEventDurationCaravan      = 24.0
	CityEventDurationRite         = 6.0
	CityEventDurationTournament   = 8.0
	CityEventDurationBlackout     = 4.0
	CityEventDurationHacking      = 12.0
	CityEventDurationRadiation    = 6.0
	CityEventDurationMutantAttack = 4.0
	CityEventDurationCultRitual   = 3.0
	CityEventDurationHaunting     = 12.0
	CityEventDurationDefault      = 12.0
)

// NPC pathfinding constants.
const (
	// NPCDefaultMoveSpeed is the default NPC movement speed (units per second).
	NPCDefaultMoveSpeed = 3.0
	// NPCDefaultArrivalThreshold is the distance at which NPCs stop moving.
	NPCDefaultArrivalThreshold = 1.0
	// NPCMinMovementThreshold is the minimum movement to not be considered stuck.
	NPCMinMovementThreshold = 0.01
	// NPCDefaultMaxStuckTime is the time before giving up on a path.
	NPCDefaultMaxStuckTime = 5.0
)
