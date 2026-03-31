package systems

import (
	"fmt"
	"sync"

	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// ============================================================================
// Mount System
// ============================================================================

// MountType represents the type of mount creature.
type MountType string

const (
	MountHorse    MountType = "horse"
	MountWolf     MountType = "wolf"
	MountBear     MountType = "bear"
	MountGriffin  MountType = "griffin"
	MountDragon   MountType = "dragon"
	MountMech     MountType = "mech"
	MountHover    MountType = "hover"
	MountSpider   MountType = "spider"
	MountUndead   MountType = "undead"
	MountMutant   MountType = "mutant"
	MountCyber    MountType = "cyber"
	MountRobot    MountType = "robot"
	MountRadbeast MountType = "radbeast"
	MountScorpion MountType = "scorpion"
)

// MountTrait represents special abilities or traits of a mount.
type MountTrait string

const (
	TraitSwift        MountTrait = "swift"        // Faster movement
	TraitSturdy       MountTrait = "sturdy"       // More health
	TraitFearless     MountTrait = "fearless"     // Ignores fear effects
	TraitNightVision  MountTrait = "night_vision" // Better vision at night
	TraitAquatic      MountTrait = "aquatic"      // Can swim
	TraitFlying       MountTrait = "flying"       // Can fly
	TraitArmored      MountTrait = "armored"      // Reduced damage
	TraitVenomous     MountTrait = "venomous"     // Deals poison damage
	TraitRegenerating MountTrait = "regenerating" // Health regeneration
	TraitStealthy     MountTrait = "stealthy"     // Reduces detection
	TraitCarrier      MountTrait = "carrier"      // Extra cargo capacity
	TraitWarMount     MountTrait = "war_mount"    // Combat bonuses
)

// MountStats represents the base stats of a mount.
type MountStats struct {
	Speed         float64
	Stamina       float64
	MaxStamina    float64
	StaminaRegen  float64
	Health        float64
	MaxHealth     float64
	HealthRegen   float64
	CargoCapacity int
	Passengers    int
}

// MountArchetype defines a mount template.
type MountArchetype struct {
	Type        MountType
	Name        string
	Genre       string
	Description string
	BaseStats   MountStats
	Traits      []MountTrait
	TameLevel   int // Minimum player level to tame
	Rarity      int // 1-5, higher = rarer
}

// MountState tracks the current state of a mount.
type MountState struct {
	EntityID    ecs.Entity
	OwnerID     uint64
	Archetype   *MountArchetype
	Name        string
	Stats       MountStats
	Traits      []MountTrait
	Mood        float64 // 0-100, affects performance
	Hunger      float64 // 0-100, 100 = starving
	Loyalty     float64 // 0-100, affects obedience
	Experience  float64
	Level       int
	IsMounted   bool
	RiderEntity ecs.Entity
	IsSprinting bool
	LastFed     float64
	LastRested  float64
	TamedTime   float64
	BondLevel   int               // 1-5, improves with time and care
	Equipment   map[string]string // Slot -> equipment ID
}

// MountSystem manages mounts and their behaviors.
type MountSystem struct {
	mu         sync.RWMutex
	Seed       int64
	Genre      string
	Mounts     map[ecs.Entity]*MountState
	Archetypes map[MountType]*MountArchetype
	GameTime   float64
	counter    uint64
}

// NewMountSystem creates a new mount system.
func NewMountSystem(seed int64, genre string) *MountSystem {
	sys := &MountSystem{
		Seed:       seed,
		Genre:      genre,
		Mounts:     make(map[ecs.Entity]*MountState),
		Archetypes: make(map[MountType]*MountArchetype),
	}
	sys.initArchetypes()
	return sys
}

// initArchetypes initializes mount archetypes for all genres.
func (s *MountSystem) initArchetypes() {
	// Fantasy mounts
	s.Archetypes[MountHorse] = &MountArchetype{
		Type: MountHorse, Name: "War Horse", Genre: "fantasy",
		Description: "A noble steed trained for battle",
		BaseStats: MountStats{
			Speed: 18, MaxStamina: 100, StaminaRegen: 5,
			MaxHealth: 80, HealthRegen: 0.5, CargoCapacity: 10, Passengers: 1,
		},
		Traits: []MountTrait{TraitSwift, TraitSturdy}, TameLevel: 1, Rarity: 1,
	}
	s.Archetypes[MountWolf] = &MountArchetype{
		Type: MountWolf, Name: "Dire Wolf", Genre: "fantasy",
		Description: "A massive wolf bred for riding",
		BaseStats: MountStats{
			Speed: 22, MaxStamina: 80, StaminaRegen: 6,
			MaxHealth: 60, HealthRegen: 1, CargoCapacity: 5, Passengers: 1,
		},
		Traits: []MountTrait{TraitSwift, TraitFearless, TraitNightVision}, TameLevel: 5, Rarity: 2,
	}
	s.Archetypes[MountGriffin] = &MountArchetype{
		Type: MountGriffin, Name: "Griffin", Genre: "fantasy",
		Description: "A majestic flying creature",
		BaseStats: MountStats{
			Speed: 30, MaxStamina: 120, StaminaRegen: 3,
			MaxHealth: 100, HealthRegen: 0.3, CargoCapacity: 15, Passengers: 1,
		},
		Traits: []MountTrait{TraitFlying, TraitFearless, TraitWarMount}, TameLevel: 15, Rarity: 4,
	}
	s.Archetypes[MountDragon] = &MountArchetype{
		Type: MountDragon, Name: "Young Dragon", Genre: "fantasy",
		Description: "A dragon that has bonded with you",
		BaseStats: MountStats{
			Speed: 40, MaxStamina: 150, StaminaRegen: 2,
			MaxHealth: 200, HealthRegen: 1, CargoCapacity: 25, Passengers: 2,
		},
		Traits: []MountTrait{TraitFlying, TraitFearless, TraitArmored, TraitWarMount}, TameLevel: 25, Rarity: 5,
	}

	// Sci-Fi mounts
	s.Archetypes[MountMech] = &MountArchetype{
		Type: MountMech, Name: "Mech Walker", Genre: "sci-fi",
		Description: "A bipedal mechanical walker",
		BaseStats: MountStats{
			Speed: 15, MaxStamina: 200, StaminaRegen: 10,
			MaxHealth: 150, HealthRegen: 0, CargoCapacity: 30, Passengers: 1,
		},
		Traits: []MountTrait{TraitArmored, TraitCarrier, TraitWarMount}, TameLevel: 5, Rarity: 2,
	}
	s.Archetypes[MountHover] = &MountArchetype{
		Type: MountHover, Name: "Hover Drone", Genre: "sci-fi",
		Description: "A personal hover transport",
		BaseStats: MountStats{
			Speed: 35, MaxStamina: 80, StaminaRegen: 15,
			MaxHealth: 50, HealthRegen: 0, CargoCapacity: 5, Passengers: 1,
		},
		Traits: []MountTrait{TraitSwift, TraitFlying, TraitStealthy}, TameLevel: 10, Rarity: 3,
	}
	s.Archetypes[MountRobot] = &MountArchetype{
		Type: MountRobot, Name: "Combat Drone", Genre: "sci-fi",
		Description: "An AI-controlled combat platform",
		BaseStats: MountStats{
			Speed: 25, MaxStamina: 150, StaminaRegen: 8,
			MaxHealth: 120, HealthRegen: 0, CargoCapacity: 20, Passengers: 1,
		},
		Traits: []MountTrait{TraitArmored, TraitWarMount, TraitFearless}, TameLevel: 15, Rarity: 4,
	}

	// Horror mounts
	s.Archetypes[MountUndead] = &MountArchetype{
		Type: MountUndead, Name: "Nightmare", Genre: "horror",
		Description: "An undead steed from beyond the grave",
		BaseStats: MountStats{
			Speed: 20, MaxStamina: 999, StaminaRegen: 0,
			MaxHealth: 60, HealthRegen: 2, CargoCapacity: 8, Passengers: 1,
		},
		Traits: []MountTrait{TraitFearless, TraitRegenerating, TraitNightVision}, TameLevel: 10, Rarity: 3,
	}
	s.Archetypes[MountSpider] = &MountArchetype{
		Type: MountSpider, Name: "Giant Spider", Genre: "horror",
		Description: "A massive arachnid that can climb walls",
		BaseStats: MountStats{
			Speed: 18, MaxStamina: 90, StaminaRegen: 4,
			MaxHealth: 70, HealthRegen: 0.5, CargoCapacity: 12, Passengers: 1,
		},
		Traits: []MountTrait{TraitStealthy, TraitVenomous, TraitNightVision}, TameLevel: 8, Rarity: 2,
	}

	// Cyberpunk mounts
	s.Archetypes[MountCyber] = &MountArchetype{
		Type: MountCyber, Name: "Cyber Hound", Genre: "cyberpunk",
		Description: "A cybernetically enhanced canine",
		BaseStats: MountStats{
			Speed: 28, MaxStamina: 100, StaminaRegen: 8,
			MaxHealth: 80, HealthRegen: 0, CargoCapacity: 10, Passengers: 1,
		},
		Traits: []MountTrait{TraitSwift, TraitNightVision, TraitWarMount}, TameLevel: 5, Rarity: 2,
	}

	// Post-Apocalyptic mounts
	s.Archetypes[MountMutant] = &MountArchetype{
		Type: MountMutant, Name: "Mutant Beast", Genre: "post-apocalyptic",
		Description: "A radiation-mutated creature",
		BaseStats: MountStats{
			Speed: 20, MaxStamina: 110, StaminaRegen: 6,
			MaxHealth: 100, HealthRegen: 1, CargoCapacity: 15, Passengers: 1,
		},
		Traits: []MountTrait{TraitSturdy, TraitRegenerating, TraitFearless}, TameLevel: 5, Rarity: 2,
	}
	s.Archetypes[MountRadbeast] = &MountArchetype{
		Type: MountRadbeast, Name: "Rad-Scorpion", Genre: "post-apocalyptic",
		Description: "A giant irradiated scorpion",
		BaseStats: MountStats{
			Speed: 16, MaxStamina: 130, StaminaRegen: 5,
			MaxHealth: 120, HealthRegen: 0.5, CargoCapacity: 20, Passengers: 2,
		},
		Traits: []MountTrait{TraitArmored, TraitVenomous, TraitCarrier}, TameLevel: 10, Rarity: 3,
	}
	s.Archetypes[MountScorpion] = s.Archetypes[MountRadbeast] // Alias

	// Copy stamina to stats
	for _, arch := range s.Archetypes {
		arch.BaseStats.Stamina = arch.BaseStats.MaxStamina
		arch.BaseStats.Health = arch.BaseStats.MaxHealth
	}
}

// GetArchetype returns a mount archetype by type.
func (s *MountSystem) GetArchetype(mountType MountType) *MountArchetype {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Archetypes[mountType]
}

// GetGenreArchetypes returns all archetypes available for the current genre.
func (s *MountSystem) GetGenreArchetypes() []*MountArchetype {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*MountArchetype, 0)
	for _, arch := range s.Archetypes {
		if arch.Genre == s.Genre {
			result = append(result, arch)
		}
	}
	return result
}

// TameMount creates a new tamed mount for a player.
func (s *MountSystem) TameMount(entity ecs.Entity, ownerID uint64, mountType MountType, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.Mounts[entity]; exists {
		return fmt.Errorf("entity already registered as mount")
	}

	arch := s.Archetypes[mountType]
	if arch == nil {
		return fmt.Errorf("unknown mount type: %s", mountType)
	}

	state := &MountState{
		EntityID:  entity,
		OwnerID:   ownerID,
		Archetype: arch,
		Name:      name,
		Stats:     arch.BaseStats,
		Traits:    make([]MountTrait, len(arch.Traits)),
		Mood:      70,
		Hunger:    0,
		Loyalty:   50,
		Level:     1,
		BondLevel: 1,
		TamedTime: s.GameTime,
		Equipment: make(map[string]string),
	}
	copy(state.Traits, arch.Traits)

	s.Mounts[entity] = state
	return nil
}

// GetMount returns the mount state for an entity.
func (s *MountSystem) GetMount(entity ecs.Entity) *MountState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Mounts[entity]
}

// MountCreature mounts a rider on a mount.
func (s *MountSystem) MountCreature(mountEntity, riderEntity ecs.Entity) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	mount := s.Mounts[mountEntity]
	if mount == nil {
		return fmt.Errorf("not a registered mount")
	}

	if mount.IsMounted {
		return fmt.Errorf("mount already has a rider")
	}

	if mount.Stats.Stamina <= 0 {
		return fmt.Errorf("mount is too exhausted to ride")
	}

	if mount.Mood < 10 {
		return fmt.Errorf("mount refuses to be ridden (mood too low)")
	}

	mount.IsMounted = true
	mount.RiderEntity = riderEntity
	return nil
}

// DismountCreature removes the rider from a mount.
func (s *MountSystem) DismountCreature(mountEntity ecs.Entity) ecs.Entity {
	s.mu.Lock()
	defer s.mu.Unlock()

	mount := s.Mounts[mountEntity]
	if mount == nil || !mount.IsMounted {
		return 0
	}

	rider := mount.RiderEntity
	mount.IsMounted = false
	mount.RiderEntity = 0
	mount.IsSprinting = false

	return rider
}

// SetSprinting toggles sprint mode.
func (s *MountSystem) SetSprinting(mountEntity ecs.Entity, sprinting bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	mount := s.Mounts[mountEntity]
	if mount == nil {
		return fmt.Errorf("not a registered mount")
	}

	if sprinting && mount.Stats.Stamina < 10 {
		return fmt.Errorf("not enough stamina to sprint")
	}

	mount.IsSprinting = sprinting
	return nil
}

// FeedMount feeds the mount to reduce hunger.
func (s *MountSystem) FeedMount(mountEntity ecs.Entity, foodValue float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	mount := s.Mounts[mountEntity]
	if mount == nil {
		return fmt.Errorf("not a registered mount")
	}

	mount.Hunger -= foodValue
	if mount.Hunger < 0 {
		mount.Hunger = 0
	}

	// Feeding improves mood and loyalty
	mount.Mood += foodValue * 0.5
	if mount.Mood > 100 {
		mount.Mood = 100
	}

	mount.Loyalty += foodValue * 0.1
	if mount.Loyalty > 100 {
		mount.Loyalty = 100
	}

	mount.LastFed = s.GameTime
	return nil
}

// RestMount lets the mount rest to recover stamina.
func (s *MountSystem) RestMount(mountEntity ecs.Entity, duration float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	mount := s.Mounts[mountEntity]
	if mount == nil {
		return fmt.Errorf("not a registered mount")
	}

	if mount.IsMounted {
		return fmt.Errorf("cannot rest while mounted")
	}

	// Resting recovers stamina at 3x normal rate
	recovery := mount.Stats.StaminaRegen * 3 * duration
	mount.Stats.Stamina += recovery
	if mount.Stats.Stamina > mount.Stats.MaxStamina {
		mount.Stats.Stamina = mount.Stats.MaxStamina
	}

	// Resting improves mood
	mount.Mood += duration * 2
	if mount.Mood > 100 {
		mount.Mood = 100
	}

	mount.LastRested = s.GameTime
	return nil
}

// GetSpeed returns the current speed of a mount.
func (s *MountSystem) GetSpeed(mountEntity ecs.Entity) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	mount := s.Mounts[mountEntity]
	if mount == nil {
		return 0
	}

	speed := mount.Stats.Speed
	speed *= s.calculateSprintModifier(mount)
	speed *= s.calculateMoodModifier(mount)
	speed *= s.calculateHungerModifier(mount)
	speed *= s.calculateStaminaModifier(mount)
	speed *= s.calculateLevelModifier(mount)
	speed *= s.calculateBondModifier(mount)

	return speed
}

// calculateSprintModifier returns the speed multiplier for sprinting.
func (s *MountSystem) calculateSprintModifier(mount *MountState) float64 {
	if mount.IsSprinting {
		return 1.5
	}
	return 1.0
}

// calculateMoodModifier returns the speed multiplier based on mood.
func (s *MountSystem) calculateMoodModifier(mount *MountState) float64 {
	return 0.5 + (mount.Mood / 200) // 0.5-1.0 based on mood
}

// calculateHungerModifier returns the speed multiplier based on hunger.
func (s *MountSystem) calculateHungerModifier(mount *MountState) float64 {
	if mount.Hunger > 80 {
		return 0.7
	} else if mount.Hunger > 50 {
		return 0.85
	}
	return 1.0
}

// calculateStaminaModifier returns the speed multiplier based on stamina.
func (s *MountSystem) calculateStaminaModifier(mount *MountState) float64 {
	staminaPercent := mount.Stats.Stamina / mount.Stats.MaxStamina
	if staminaPercent < 0.2 {
		return 0.6
	} else if staminaPercent < 0.5 {
		return 0.8
	}
	return 1.0
}

// calculateLevelModifier returns the speed bonus from mount level.
func (s *MountSystem) calculateLevelModifier(mount *MountState) float64 {
	return 1.0 + float64(mount.Level-1)*0.02 // 2% per level
}

// calculateBondModifier returns the speed bonus from bond level.
func (s *MountSystem) calculateBondModifier(mount *MountState) float64 {
	return 1.0 + float64(mount.BondLevel-1)*0.05 // 5% per bond level
}

// HasTrait checks if a mount has a specific trait.
func (s *MountSystem) HasTrait(mountEntity ecs.Entity, trait MountTrait) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	mount := s.Mounts[mountEntity]
	if mount == nil {
		return false
	}

	for _, t := range mount.Traits {
		if t == trait {
			return true
		}
	}
	return false
}

// AddExperience adds XP to a mount.
func (s *MountSystem) AddExperience(mountEntity ecs.Entity, xp float64) (leveledUp bool, newLevel int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	mount := s.Mounts[mountEntity]
	if mount == nil {
		return false, 0
	}

	mount.Experience += xp
	requiredXP := float64(mount.Level) * 100

	for mount.Experience >= requiredXP && mount.Level < 20 {
		mount.Experience -= requiredXP
		mount.Level++
		leveledUp = true

		// Stat increases on level up
		mount.Stats.MaxHealth += 5
		mount.Stats.MaxStamina += 3
		mount.Stats.Speed += 0.5

		requiredXP = float64(mount.Level) * 100
	}

	return leveledUp, mount.Level
}

// ImproveBond improves the bond level with the mount.
func (s *MountSystem) ImproveBond(mountEntity ecs.Entity) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	mount := s.Mounts[mountEntity]
	if mount == nil || mount.BondLevel >= 5 {
		return false
	}

	// Bond improves based on loyalty
	if mount.Loyalty >= float64(mount.BondLevel)*20 {
		mount.BondLevel++
		return true
	}
	return false
}

// Update processes mount states (hunger, stamina, mood).
func (s *MountSystem) Update(dt float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.GameTime += dt

	for _, mount := range s.Mounts {
		s.updateMount(mount, dt)
	}
}

func (s *MountSystem) updateMount(mount *MountState, dt float64) {
	s.updateMountHunger(mount, dt)
	s.updateMountMood(mount, dt)
	s.updateMountStamina(mount, dt)
	s.updateMountHealth(mount, dt)
}

// updateMountHunger increases hunger over time and affects mood.
func (s *MountSystem) updateMountHunger(mount *MountState, dt float64) {
	mount.Hunger += dt * 0.01 // ~1% per 100 seconds
	if mount.Hunger > 100 {
		mount.Hunger = 100
	}
	// Mood decreases if hungry
	if mount.Hunger > 70 {
		mount.Mood -= dt * 0.05
	}
}

// updateMountMood clamps mood to valid range.
func (s *MountSystem) updateMountMood(mount *MountState, dt float64) {
	if mount.Mood < 0 {
		mount.Mood = 0
	}
}

// updateMountStamina manages stamina drain and regeneration.
func (s *MountSystem) updateMountStamina(mount *MountState, dt float64) {
	if mount.IsMounted {
		drainRate := 1.0 // Normal riding
		if mount.IsSprinting {
			drainRate = 5.0 // Sprinting drains faster
		}
		mount.Stats.Stamina -= dt * drainRate
	} else {
		mount.Stats.Stamina += mount.Stats.StaminaRegen * dt
	}
	// Clamp stamina
	if mount.Stats.Stamina < 0 {
		mount.Stats.Stamina = 0
		mount.IsSprinting = false
	}
	if mount.Stats.Stamina > mount.Stats.MaxStamina {
		mount.Stats.Stamina = mount.Stats.MaxStamina
	}
}

// updateMountHealth handles health regeneration.
func (s *MountSystem) updateMountHealth(mount *MountState, dt float64) {
	hasRegenTrait := s.mountHasTrait(mount, TraitRegenerating)
	if !hasRegenTrait && mount.IsMounted {
		return // No regen while mounted without trait
	}
	regenRate := mount.Stats.HealthRegen
	if hasRegenTrait {
		regenRate *= 2
	}
	mount.Stats.Health += regenRate * dt
	if mount.Stats.Health > mount.Stats.MaxHealth {
		mount.Stats.Health = mount.Stats.MaxHealth
	}
}

// mountHasTrait checks if a mount has a specific trait.
func (s *MountSystem) mountHasTrait(mount *MountState, trait MountTrait) bool {
	for _, t := range mount.Traits {
		if t == trait {
			return true
		}
	}
	return false
}

// DamageMount applies damage to a mount.
func (s *MountSystem) DamageMount(mountEntity ecs.Entity, damage float64) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	mount := s.Mounts[mountEntity]
	if mount == nil {
		return 0
	}

	// Armor trait reduces damage
	for _, t := range mount.Traits {
		if t == TraitArmored {
			damage *= 0.7
			break
		}
	}

	mount.Stats.Health -= damage
	if mount.Stats.Health < 0 {
		mount.Stats.Health = 0
	}

	// Damage reduces mood and loyalty
	mount.Mood -= damage * 0.5
	mount.Loyalty -= damage * 0.1

	return damage
}

// HealMount heals a mount.
func (s *MountSystem) HealMount(mountEntity ecs.Entity, amount float64) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	mount := s.Mounts[mountEntity]
	if mount == nil {
		return 0
	}

	oldHealth := mount.Stats.Health
	mount.Stats.Health += amount
	if mount.Stats.Health > mount.Stats.MaxHealth {
		mount.Stats.Health = mount.Stats.MaxHealth
	}

	return mount.Stats.Health - oldHealth
}

// IsAlive checks if a mount is alive.
func (s *MountSystem) IsAlive(mountEntity ecs.Entity) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	mount := s.Mounts[mountEntity]
	return mount != nil && mount.Stats.Health > 0
}

// CanFly checks if a mount can fly.
func (s *MountSystem) CanFly(mountEntity ecs.Entity) bool {
	return s.HasTrait(mountEntity, TraitFlying)
}

// CanSwim checks if a mount can swim.
func (s *MountSystem) CanSwim(mountEntity ecs.Entity) bool {
	return s.HasTrait(mountEntity, TraitAquatic)
}

// GetCargoCapacity returns the cargo capacity of a mount.
func (s *MountSystem) GetCargoCapacity(mountEntity ecs.Entity) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	mount := s.Mounts[mountEntity]
	if mount == nil {
		return 0
	}

	capacity := mount.Stats.CargoCapacity

	// Carrier trait bonus
	for _, t := range mount.Traits {
		if t == TraitCarrier {
			capacity += 10
			break
		}
	}

	return capacity
}

// GetMountStats returns the current stats of a mount.
func (s *MountSystem) GetMountStats(mountEntity ecs.Entity) *MountStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	mount := s.Mounts[mountEntity]
	if mount == nil {
		return nil
	}

	// Return a copy
	stats := mount.Stats
	return &stats
}

// MountCount returns the number of registered mounts.
func (s *MountSystem) MountCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.Mounts)
}

// GetMountLevel returns the level of a mount.
func (s *MountSystem) GetMountLevel(mountEntity ecs.Entity) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	mount := s.Mounts[mountEntity]
	if mount == nil {
		return 0
	}
	return mount.Level
}

// GetBondLevel returns the bond level of a mount.
func (s *MountSystem) GetBondLevel(mountEntity ecs.Entity) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	mount := s.Mounts[mountEntity]
	if mount == nil {
		return 0
	}
	return mount.BondLevel
}

// ReleaseMount releases a mount (removes it from the system).
func (s *MountSystem) ReleaseMount(mountEntity ecs.Entity) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	mount := s.Mounts[mountEntity]
	if mount == nil {
		return fmt.Errorf("not a registered mount")
	}

	if mount.IsMounted {
		return fmt.Errorf("cannot release while mounted")
	}

	delete(s.Mounts, mountEntity)
	return nil
}

// SetMountEquipment sets equipment in a slot.
func (s *MountSystem) SetMountEquipment(mountEntity ecs.Entity, slot, equipmentID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	mount := s.Mounts[mountEntity]
	if mount == nil {
		return fmt.Errorf("not a registered mount")
	}

	mount.Equipment[slot] = equipmentID
	return nil
}

// GetMountEquipment returns the equipment in a slot.
func (s *MountSystem) GetMountEquipment(mountEntity ecs.Entity, slot string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	mount := s.Mounts[mountEntity]
	if mount == nil {
		return ""
	}
	return mount.Equipment[slot]
}

// IsMounted checks if a mount has a rider.
func (s *MountSystem) IsMounted(mountEntity ecs.Entity) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	mount := s.Mounts[mountEntity]
	return mount != nil && mount.IsMounted
}
