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
	FreqA1 = 55.0   // A1
	FreqE2 = 82.5   // E2
	FreqA2 = 110.0  // A2
	FreqE3 = 165.0  // E3
	FreqA3 = 220.0  // A3
	FreqE4 = 330.0  // E4
	FreqA4 = 440.0  // A4

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
