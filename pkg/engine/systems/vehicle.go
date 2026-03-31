package systems

import (
	"fmt"
	"math"
	"sync"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// VehicleSystem manages vehicle movement and physics.
type VehicleSystem struct{}

// Update processes vehicle physics each tick.
func (s *VehicleSystem) Update(w *ecs.World, dt float64) {
	for _, e := range w.Entities("Vehicle", "Position") {
		s.updateVehicle(w, e, dt)
	}
}

// updateVehicle updates a single vehicle's position based on its physics.
func (s *VehicleSystem) updateVehicle(w *ecs.World, e ecs.Entity, dt float64) {
	vehicle, pos := s.getVehicleComponents(w, e)
	if vehicle == nil || pos == nil {
		return
	}
	if vehicle.Fuel > 0 && vehicle.Speed > 0 {
		s.applyVehicleMovement(vehicle, pos, dt)
	}
}

// getVehicleComponents retrieves vehicle and position components for an entity.
func (s *VehicleSystem) getVehicleComponents(w *ecs.World, e ecs.Entity) (*components.Vehicle, *components.Position) {
	vComp, ok := w.GetComponent(e, "Vehicle")
	if !ok {
		return nil, nil
	}
	pComp, ok := w.GetComponent(e, "Position")
	if !ok {
		return nil, nil
	}
	return vComp.(*components.Vehicle), pComp.(*components.Position)
}

// applyVehicleMovement updates position and consumes fuel.
func (s *VehicleSystem) applyVehicleMovement(vehicle *components.Vehicle, pos *components.Position, dt float64) {
	pos.X += math.Cos(vehicle.Direction) * vehicle.Speed * dt
	pos.Y += math.Sin(vehicle.Direction) * vehicle.Speed * dt
	vehicle.Fuel -= vehicle.Speed * dt * DefaultFuelConsumptionRate
	if vehicle.Fuel < MinFuelLevel {
		vehicle.Fuel = MinFuelLevel
	}
}

// VehicleWeaponType represents a type of vehicle weapon.
type VehicleWeaponType int

const (
	WeaponMachineGun VehicleWeaponType = iota
	WeaponCannon
	WeaponMissile
	WeaponLaser
	WeaponRam // For ramming attacks
)

// VehicleWeapon represents a weapon mounted on a vehicle.
type VehicleWeapon struct {
	Type         VehicleWeaponType
	Damage       float64
	Range        float64
	RateOfFire   float64 // Shots per second
	AmmoCapacity int
	AmmoCount    int
	LastFired    float64 // Time since last shot
	MountAngle   float64 // Angle offset from vehicle facing
	Enabled      bool
}

// VehicleCombatState represents the combat state of a vehicle.
type VehicleCombatState struct {
	EntityID    ecs.Entity
	Weapons     []*VehicleWeapon
	Armor       float64 // Damage reduction percentage
	ShieldPower float64 // Energy shield (0-100)
	ShieldRegen float64 // Shield regen per second
	Health      float64
	MaxHealth   float64
	InCombat    bool
	LastDamaged float64
	RamDamage   float64 // Damage dealt when ramming
	RamCooldown float64
}

// VehicleCombatSystem manages vehicle-to-vehicle combat.
type VehicleCombatSystem struct {
	vehicles map[ecs.Entity]*VehicleCombatState
}

// NewVehicleCombatSystem creates a new vehicle combat system.
func NewVehicleCombatSystem() *VehicleCombatSystem {
	return &VehicleCombatSystem{
		vehicles: make(map[ecs.Entity]*VehicleCombatState),
	}
}

// RegisterVehicle registers a vehicle for combat.
func (s *VehicleCombatSystem) RegisterVehicle(entity ecs.Entity, health, armor float64) *VehicleCombatState {
	state := &VehicleCombatState{
		EntityID:    entity,
		Weapons:     make([]*VehicleWeapon, 0),
		Armor:       armor,
		Health:      health,
		MaxHealth:   health,
		ShieldPower: 0,
		ShieldRegen: 5.0,
		RamDamage:   20.0,
		RamCooldown: 0,
	}
	s.vehicles[entity] = state
	return state
}

// AddWeapon adds a weapon to a vehicle.
func (s *VehicleCombatSystem) AddWeapon(entity ecs.Entity, weaponType VehicleWeaponType) bool {
	state, ok := s.vehicles[entity]
	if !ok {
		return false
	}
	weapon := s.createWeapon(weaponType)
	state.Weapons = append(state.Weapons, weapon)
	return true
}

// createWeapon creates a weapon with default stats based on type.
func (s *VehicleCombatSystem) createWeapon(weaponType VehicleWeaponType) *VehicleWeapon {
	switch weaponType {
	case WeaponMachineGun:
		return &VehicleWeapon{
			Type: WeaponMachineGun, Damage: 5, Range: 150,
			RateOfFire: 10, AmmoCapacity: 200, AmmoCount: 200, Enabled: true,
		}
	case WeaponCannon:
		return &VehicleWeapon{
			Type: WeaponCannon, Damage: 40, Range: 300,
			RateOfFire: 0.5, AmmoCapacity: 20, AmmoCount: 20, Enabled: true,
		}
	case WeaponMissile:
		return &VehicleWeapon{
			Type: WeaponMissile, Damage: 80, Range: 500,
			RateOfFire: 0.2, AmmoCapacity: 4, AmmoCount: 4, Enabled: true,
		}
	case WeaponLaser:
		return &VehicleWeapon{
			Type: WeaponLaser, Damage: 15, Range: 250,
			RateOfFire: 5, AmmoCapacity: 100, AmmoCount: 100, Enabled: true,
		}
	default:
		return &VehicleWeapon{
			Type: WeaponRam, Damage: 20, Range: 10,
			RateOfFire: 0.3, AmmoCapacity: -1, AmmoCount: -1, Enabled: true,
		}
	}
}

// Fire attempts to fire a weapon at a target.
func (s *VehicleCombatSystem) Fire(attacker ecs.Entity, weaponIdx int, target ecs.Entity, dist float64) float64 {
	state, ok := s.vehicles[attacker]
	if !ok || weaponIdx >= len(state.Weapons) {
		return 0
	}
	weapon := state.Weapons[weaponIdx]
	if !weapon.Enabled || weapon.AmmoCount == 0 || dist > weapon.Range {
		return 0
	}
	cooldown := 1.0 / weapon.RateOfFire
	if weapon.LastFired < cooldown {
		return 0
	}
	weapon.LastFired = 0
	if weapon.AmmoCount > 0 {
		weapon.AmmoCount--
	}
	return s.ApplyDamage(target, weapon.Damage)
}

// ApplyDamage applies damage to a vehicle, respecting armor and shields.
func (s *VehicleCombatSystem) ApplyDamage(entity ecs.Entity, damage float64) float64 {
	state, ok := s.vehicles[entity]
	if !ok {
		return 0
	}
	state.InCombat = true
	state.LastDamaged = 0
	// Shields absorb damage first
	if state.ShieldPower > 0 {
		if state.ShieldPower >= damage {
			state.ShieldPower -= damage
			return 0
		}
		damage -= state.ShieldPower
		state.ShieldPower = 0
	}
	// Apply armor reduction
	effectiveDamage := damage * (1.0 - state.Armor/100.0)
	state.Health -= effectiveDamage
	if state.Health < 0 {
		state.Health = 0
	}
	return effectiveDamage
}

// ProcessRam handles a ramming attack between vehicles.
func (s *VehicleCombatSystem) ProcessRam(attacker, target ecs.Entity, attackerSpeed float64) (float64, float64) {
	attackerState, ok1 := s.vehicles[attacker]
	_, ok2 := s.vehicles[target]
	if !ok1 || !ok2 {
		return 0, 0
	}
	if attackerState.RamCooldown > 0 {
		return 0, 0
	}
	attackerState.RamCooldown = 3.0 // 3 second cooldown
	// Damage scales with speed
	speedMultiplier := attackerSpeed / 50.0 // 50 units/sec = 1x damage
	attackerDamage := attackerState.RamDamage * speedMultiplier
	// Both vehicles take damage, attacker takes less
	targetDamage := s.ApplyDamage(target, attackerDamage)
	selfDamage := s.ApplyDamage(attacker, attackerDamage*0.3)
	return targetDamage, selfDamage
}

// Update processes combat state each tick.
func (s *VehicleCombatSystem) Update(w *ecs.World, dt float64) {
	for _, state := range s.vehicles {
		s.updateVehicleState(state, dt)
	}
}

// updateVehicleState updates cooldowns, shields, and combat status for one vehicle.
func (s *VehicleCombatSystem) updateVehicleState(state *VehicleCombatState, dt float64) {
	s.updateWeaponCooldowns(state, dt)
	s.updateRamCooldown(state, dt)
	s.regenerateShields(state, dt)
	s.updateCombatStatus(state, dt)
}

// updateWeaponCooldowns advances weapon cooldown timers.
func (s *VehicleCombatSystem) updateWeaponCooldowns(state *VehicleCombatState, dt float64) {
	for _, weapon := range state.Weapons {
		weapon.LastFired += dt
	}
}

// updateRamCooldown reduces ram cooldown timer.
func (s *VehicleCombatSystem) updateRamCooldown(state *VehicleCombatState, dt float64) {
	if state.RamCooldown > 0 {
		state.RamCooldown -= dt
	}
}

// regenerateShields restores shield power when out of combat.
func (s *VehicleCombatSystem) regenerateShields(state *VehicleCombatState, dt float64) {
	if state.ShieldPower < 100 && !state.InCombat {
		state.ShieldPower += state.ShieldRegen * dt
		if state.ShieldPower > 100 {
			state.ShieldPower = 100
		}
	}
}

// updateCombatStatus exits combat after no damage for 5 seconds.
func (s *VehicleCombatSystem) updateCombatStatus(state *VehicleCombatState, dt float64) {
	state.LastDamaged += dt
	if state.LastDamaged > 5.0 {
		state.InCombat = false
	}
}

// IsDestroyed returns whether a vehicle is destroyed.
func (s *VehicleCombatSystem) IsDestroyed(entity ecs.Entity) bool {
	state, ok := s.vehicles[entity]
	return ok && state.Health <= 0
}

// GetHealth returns the current health of a vehicle.
func (s *VehicleCombatSystem) GetHealth(entity ecs.Entity) float64 {
	state, ok := s.vehicles[entity]
	if !ok {
		return 0
	}
	return state.Health
}

// ============================================================================
// Vehicle Customization System
// ============================================================================

// CustomizationCategory represents a type of customization.
type CustomizationCategory string

const (
	CategoryEngine      CustomizationCategory = "engine"
	CategoryArmor       CustomizationCategory = "armor"
	CategoryWeapons     CustomizationCategory = "weapons"
	CategoryStorage     CustomizationCategory = "storage"
	CategoryAppearance  CustomizationCategory = "appearance"
	CategoryPerformance CustomizationCategory = "performance"
)

// CustomizationSlot represents where a customization can be applied.
type CustomizationSlot string

const (
	SlotEnginePrimary    CustomizationSlot = "engine_primary"
	SlotEngineSecondary  CustomizationSlot = "engine_secondary"
	SlotArmorFront       CustomizationSlot = "armor_front"
	SlotArmorRear        CustomizationSlot = "armor_rear"
	SlotArmorSides       CustomizationSlot = "armor_sides"
	SlotWeaponPrimary    CustomizationSlot = "weapon_primary"
	SlotWeaponSecondary  CustomizationSlot = "weapon_secondary"
	SlotWeaponTurret     CustomizationSlot = "weapon_turret"
	SlotStorageCargo     CustomizationSlot = "storage_cargo"
	SlotStoragePassenger CustomizationSlot = "storage_passenger"
	SlotPaintPrimary     CustomizationSlot = "paint_primary"
	SlotPaintSecondary   CustomizationSlot = "paint_secondary"
	SlotDecals           CustomizationSlot = "decals"
	SlotTires            CustomizationSlot = "tires"
	SlotSuspension       CustomizationSlot = "suspension"
	SlotExhaust          CustomizationSlot = "exhaust"
	SlotNitro            CustomizationSlot = "nitro"
)

// VehicleCustomization represents a single customization item.
type VehicleCustomization struct {
	ID          string
	Name        string
	Category    CustomizationCategory
	Slot        CustomizationSlot
	Cost        float64
	Weight      float64 // Affects handling
	Description string

	// Stat modifiers (multiplicative, 1.0 = no change)
	SpeedMod        float64
	AccelerationMod float64
	HandlingMod     float64 // Affects steering responsiveness
	ArmorMod        float64
	FuelEfficiency  float64 // Lower = more efficient
	CargoCapacity   int     // Additional cargo slots
	PassengerSeats  int     // Additional passenger slots
	DamageMod       float64 // Weapon damage multiplier
	RangeMod        float64 // Weapon range multiplier

	// Visual properties
	PrimaryColor   uint32 // RGB color
	SecondaryColor uint32
	DecalID        string

	// Requirements
	MinLevel          int
	RequiredMaterials map[string]int
	Incompatible      []string // IDs of incompatible customizations
}

// VehicleCustomizationState tracks customizations for a vehicle.
type VehicleCustomizationState struct {
	EntityID       ecs.Entity
	VehicleType    string
	InstalledMods  map[CustomizationSlot]*VehicleCustomization
	PrimaryColor   uint32
	SecondaryColor uint32
	DecalIDs       []string
	CustomName     string
	TotalWeight    float64
	MaxWeight      float64
	Level          int // Vehicle customization level (affects available mods)
	Experience     float64
	NextLevelXP    float64
}

// VehicleCustomizationSystem manages vehicle customizations.
type VehicleCustomizationSystem struct {
	mu            sync.RWMutex
	Seed          int64
	Genre         string
	Vehicles      map[ecs.Entity]*VehicleCustomizationState
	Catalog       map[string]*VehicleCustomization // All available customizations
	GenreCatalogs map[string][]string              // Genre -> customization IDs
	counter       uint64
}

// NewVehicleCustomizationSystem creates a new vehicle customization system.
func NewVehicleCustomizationSystem(seed int64, genre string) *VehicleCustomizationSystem {
	sys := &VehicleCustomizationSystem{
		Seed:          seed,
		Genre:         genre,
		Vehicles:      make(map[ecs.Entity]*VehicleCustomizationState),
		Catalog:       make(map[string]*VehicleCustomization),
		GenreCatalogs: make(map[string][]string),
	}
	sys.initCatalog()
	return sys
}

// initCatalog initializes the customization catalog with genre-specific items.
func (s *VehicleCustomizationSystem) initCatalog() {
	s.initEngineUpgrades()
	s.initArmorUpgrades()
	s.initPerformanceUpgrades()
	s.initStorageUpgrades()
	s.initAppearanceUpgrades()
	s.initWeaponUpgrades()
}

func (s *VehicleCustomizationSystem) initEngineUpgrades() {
	engines := []struct {
		id       string
		name     string
		genre    string
		speed    float64
		accel    float64
		fuel     float64
		cost     float64
		weight   float64
		minLevel int
	}{
		// Fantasy
		{"engine_enchanted", "Enchanted Harness", "fantasy", 1.15, 1.1, 0.9, 500, 5, 1},
		{"engine_spirit", "Spirit Steed Binding", "fantasy", 1.3, 1.2, 0.85, 1500, 0, 5},
		{"engine_elemental", "Elemental Core", "fantasy", 1.5, 1.4, 0.8, 5000, -10, 10},
		// Sci-Fi
		{"engine_ion", "Ion Thruster", "sci-fi", 1.2, 1.15, 0.9, 800, 10, 1},
		{"engine_fusion", "Fusion Core", "sci-fi", 1.4, 1.3, 0.85, 2500, 15, 5},
		{"engine_quantum", "Quantum Drive", "sci-fi", 1.6, 1.5, 0.75, 8000, 20, 10},
		// Horror
		{"engine_cursed", "Cursed Engine", "horror", 1.1, 1.2, 1.1, 400, 10, 1},
		{"engine_spectral", "Spectral Propulsion", "horror", 1.25, 1.3, 0.95, 1200, -5, 5},
		{"engine_void", "Void Heart", "horror", 1.4, 1.4, 0.9, 4000, 0, 10},
		// Cyberpunk
		{"engine_turbo", "Turbo Charger", "cyberpunk", 1.25, 1.2, 1.05, 600, 8, 1},
		{"engine_hybrid", "Hybrid Reactor", "cyberpunk", 1.35, 1.25, 0.8, 2000, 12, 5},
		{"engine_neural", "Neural-Linked Engine", "cyberpunk", 1.55, 1.5, 0.7, 7000, 5, 10},
		// Post-Apocalyptic
		{"engine_salvaged", "Salvaged Supercharger", "post-apocalyptic", 1.15, 1.15, 1.1, 300, 15, 1},
		{"engine_nitro", "Jury-Rigged Nitro", "post-apocalyptic", 1.4, 1.35, 1.2, 1000, 10, 5},
		{"engine_nuclear", "Mini-Reactor", "post-apocalyptic", 1.5, 1.4, 0.5, 6000, 30, 10},
	}

	for _, e := range engines {
		custom := &VehicleCustomization{
			ID:              e.id,
			Name:            e.name,
			Category:        CategoryEngine,
			Slot:            SlotEnginePrimary,
			Cost:            e.cost,
			Weight:          e.weight,
			SpeedMod:        e.speed,
			AccelerationMod: e.accel,
			FuelEfficiency:  e.fuel,
			HandlingMod:     1.0,
			ArmorMod:        1.0,
			DamageMod:       1.0,
			RangeMod:        1.0,
			MinLevel:        e.minLevel,
		}
		s.Catalog[e.id] = custom
		s.GenreCatalogs[e.genre] = append(s.GenreCatalogs[e.genre], e.id)
	}
}

func (s *VehicleCustomizationSystem) initArmorUpgrades() {
	armors := []struct {
		id       string
		name     string
		genre    string
		armor    float64
		speed    float64
		weight   float64
		cost     float64
		minLevel int
	}{
		// Fantasy
		{"armor_iron", "Iron Plating", "fantasy", 1.2, 0.95, 20, 400, 1},
		{"armor_mithril", "Mithril Coating", "fantasy", 1.4, 0.98, 10, 2000, 5},
		{"armor_dragon", "Dragonscale Armor", "fantasy", 1.6, 1.0, 5, 8000, 10},
		// Sci-Fi
		{"armor_composite", "Composite Plating", "sci-fi", 1.25, 0.95, 15, 600, 1},
		{"armor_reactive", "Reactive Armor", "sci-fi", 1.45, 0.92, 25, 2500, 5},
		{"armor_energy", "Energy Shield Generator", "sci-fi", 1.7, 1.0, 5, 10000, 10},
		// Horror
		{"armor_bone", "Bone Reinforcement", "horror", 1.15, 0.97, 15, 350, 1},
		{"armor_shadow", "Shadowweave Shell", "horror", 1.35, 1.02, 0, 1800, 5},
		{"armor_eldritch", "Eldritch Carapace", "horror", 1.55, 0.95, 20, 7000, 10},
		// Cyberpunk
		{"armor_ceramic", "Ceramic Plating", "cyberpunk", 1.2, 0.96, 12, 500, 1},
		{"armor_nano", "Nanoweave Coating", "cyberpunk", 1.4, 0.98, 8, 2200, 5},
		{"armor_smart", "Smart Armor System", "cyberpunk", 1.65, 1.0, 10, 9000, 10},
		// Post-Apocalyptic
		{"armor_scrap", "Scrap Metal Plating", "post-apocalyptic", 1.15, 0.9, 30, 200, 1},
		{"armor_welded", "Welded Cage", "post-apocalyptic", 1.35, 0.88, 40, 800, 5},
		{"armor_bunker", "Bunker Panels", "post-apocalyptic", 1.6, 0.85, 50, 4000, 10},
	}

	for _, a := range armors {
		custom := &VehicleCustomization{
			ID:              a.id,
			Name:            a.name,
			Category:        CategoryArmor,
			Slot:            SlotArmorFront,
			Cost:            a.cost,
			Weight:          a.weight,
			ArmorMod:        a.armor,
			SpeedMod:        a.speed,
			AccelerationMod: 1.0,
			FuelEfficiency:  1.0,
			HandlingMod:     1.0,
			DamageMod:       1.0,
			RangeMod:        1.0,
			MinLevel:        a.minLevel,
		}
		s.Catalog[a.id] = custom
		s.GenreCatalogs[a.genre] = append(s.GenreCatalogs[a.genre], a.id)
	}
}

func (s *VehicleCustomizationSystem) initPerformanceUpgrades() {
	perf := []struct {
		id       string
		name     string
		genre    string
		handling float64
		speed    float64
		cost     float64
		weight   float64
		minLevel int
	}{
		// Fantasy
		{"perf_feather", "Featherlight Enchantment", "fantasy", 1.2, 1.05, 600, -10, 2},
		{"perf_wind", "Wind Walker Blessing", "fantasy", 1.35, 1.1, 1800, -15, 6},
		// Sci-Fi
		{"perf_gyro", "Gyro Stabilizers", "sci-fi", 1.25, 1.0, 700, 5, 2},
		{"perf_hover", "Hover Assist", "sci-fi", 1.4, 1.05, 2000, 8, 6},
		// Horror
		{"perf_shadow", "Shadow Drift", "horror", 1.15, 1.08, 500, 0, 2},
		{"perf_ghost", "Ghostly Glide", "horror", 1.3, 1.12, 1500, -5, 6},
		// Cyberpunk
		{"perf_chip", "Performance Chip", "cyberpunk", 1.2, 1.08, 650, 2, 2},
		{"perf_auto", "Auto-Balance System", "cyberpunk", 1.38, 1.06, 1900, 5, 6},
		// Post-Apocalyptic
		{"perf_springs", "Salvaged Springs", "post-apocalyptic", 1.18, 0.98, 250, 5, 2},
		{"perf_shocks", "Heavy-Duty Shocks", "post-apocalyptic", 1.32, 1.02, 900, 10, 6},
	}

	for _, p := range perf {
		custom := &VehicleCustomization{
			ID:              p.id,
			Name:            p.name,
			Category:        CategoryPerformance,
			Slot:            SlotSuspension,
			Cost:            p.cost,
			Weight:          p.weight,
			HandlingMod:     p.handling,
			SpeedMod:        p.speed,
			AccelerationMod: 1.0,
			FuelEfficiency:  1.0,
			ArmorMod:        1.0,
			DamageMod:       1.0,
			RangeMod:        1.0,
			MinLevel:        p.minLevel,
		}
		s.Catalog[p.id] = custom
		s.GenreCatalogs[p.genre] = append(s.GenreCatalogs[p.genre], p.id)
	}
}

func (s *VehicleCustomizationSystem) initStorageUpgrades() {
	storage := []struct {
		id        string
		name      string
		genre     string
		cargo     int
		passenger int
		speed     float64
		cost      float64
		weight    float64
		minLevel  int
	}{
		// Fantasy
		{"storage_saddlebags", "Enchanted Saddlebags", "fantasy", 5, 0, 0.98, 300, 5, 1},
		{"storage_cart_expand", "Expanded Cart Bed", "fantasy", 15, 2, 0.92, 1200, 30, 4},
		// Sci-Fi
		{"storage_module", "Cargo Module", "sci-fi", 8, 0, 0.97, 500, 10, 1},
		{"storage_bay", "Expanded Bay", "sci-fi", 20, 4, 0.9, 2000, 50, 4},
		// Horror
		{"storage_coffin", "Coffin Storage", "horror", 4, 0, 0.99, 250, 8, 1},
		{"storage_crypt", "Crypt Compartment", "horror", 12, 1, 0.94, 1000, 25, 4},
		// Cyberpunk
		{"storage_stealth", "Stealth Compartment", "cyberpunk", 3, 0, 1.0, 400, 5, 1},
		{"storage_modular", "Modular Storage System", "cyberpunk", 18, 2, 0.92, 1800, 40, 4},
		// Post-Apocalyptic
		{"storage_cage", "Cage Extension", "post-apocalyptic", 6, 1, 0.95, 200, 15, 1},
		{"storage_trailer", "Makeshift Trailer", "post-apocalyptic", 25, 3, 0.85, 800, 60, 4},
	}

	for _, st := range storage {
		custom := &VehicleCustomization{
			ID:              st.id,
			Name:            st.name,
			Category:        CategoryStorage,
			Slot:            SlotStorageCargo,
			Cost:            st.cost,
			Weight:          st.weight,
			CargoCapacity:   st.cargo,
			PassengerSeats:  st.passenger,
			SpeedMod:        st.speed,
			AccelerationMod: 1.0,
			FuelEfficiency:  1.0,
			HandlingMod:     1.0,
			ArmorMod:        1.0,
			DamageMod:       1.0,
			RangeMod:        1.0,
			MinLevel:        st.minLevel,
		}
		s.Catalog[st.id] = custom
		s.GenreCatalogs[st.genre] = append(s.GenreCatalogs[st.genre], st.id)
	}
}

func (s *VehicleCustomizationSystem) initAppearanceUpgrades() {
	paints := []struct {
		id    string
		name  string
		genre string
		color uint32
		cost  float64
	}{
		// Fantasy
		{"paint_royal_blue", "Royal Blue", "fantasy", 0x1E3A8A, 100},
		{"paint_forest_green", "Forest Green", "fantasy", 0x166534, 100},
		{"paint_golden", "Golden Shimmer", "fantasy", 0xFBBF24, 500},
		// Sci-Fi
		{"paint_chrome", "Chrome Silver", "sci-fi", 0xC0C0C0, 150},
		{"paint_neon_blue", "Neon Blue", "sci-fi", 0x00BFFF, 200},
		{"paint_void_black", "Void Black", "sci-fi", 0x0A0A0A, 300},
		// Horror
		{"paint_blood_red", "Blood Red", "horror", 0x8B0000, 100},
		{"paint_corpse_grey", "Corpse Grey", "horror", 0x4A4A4A, 80},
		{"paint_midnight", "Midnight", "horror", 0x191970, 150},
		// Cyberpunk
		{"paint_hot_pink", "Hot Pink", "cyberpunk", 0xFF1493, 120},
		{"paint_cyber_yellow", "Cyber Yellow", "cyberpunk", 0xFFFF00, 120},
		{"paint_matrix_green", "Matrix Green", "cyberpunk", 0x00FF00, 200},
		// Post-Apocalyptic
		{"paint_rust", "Rust Orange", "post-apocalyptic", 0xB7410E, 50},
		{"paint_dust", "Desert Dust", "post-apocalyptic", 0xC4A35A, 50},
		{"paint_camo", "Wasteland Camo", "post-apocalyptic", 0x78866B, 150},
	}

	for _, p := range paints {
		custom := &VehicleCustomization{
			ID:              p.id,
			Name:            p.name,
			Category:        CategoryAppearance,
			Slot:            SlotPaintPrimary,
			Cost:            p.cost,
			PrimaryColor:    p.color,
			SpeedMod:        1.0,
			AccelerationMod: 1.0,
			FuelEfficiency:  1.0,
			HandlingMod:     1.0,
			ArmorMod:        1.0,
			DamageMod:       1.0,
			RangeMod:        1.0,
			MinLevel:        1,
		}
		s.Catalog[p.id] = custom
		s.GenreCatalogs[p.genre] = append(s.GenreCatalogs[p.genre], p.id)
	}
}

func (s *VehicleCustomizationSystem) initWeaponUpgrades() {
	weapons := []struct {
		id       string
		name     string
		genre    string
		damage   float64
		wepRange float64
		cost     float64
		weight   float64
		minLevel int
	}{
		// Fantasy
		{"weapon_crossbow", "Mounted Crossbow", "fantasy", 1.2, 1.1, 800, 10, 3},
		{"weapon_ballista", "Mini Ballista", "fantasy", 1.5, 1.3, 3000, 30, 7},
		// Sci-Fi
		{"weapon_laser", "Laser Turret", "sci-fi", 1.25, 1.2, 1200, 15, 3},
		{"weapon_plasma", "Plasma Cannon", "sci-fi", 1.6, 1.4, 5000, 25, 7},
		// Horror
		{"weapon_spike", "Impaling Spikes", "horror", 1.3, 0.8, 600, 15, 3},
		{"weapon_curse", "Curse Projector", "horror", 1.45, 1.2, 2500, 10, 7},
		// Cyberpunk
		{"weapon_smg", "Mounted SMG", "cyberpunk", 1.15, 1.0, 700, 8, 3},
		{"weapon_railgun", "Railgun", "cyberpunk", 1.7, 1.5, 6000, 30, 7},
		// Post-Apocalyptic
		{"weapon_spear", "Harpoon Launcher", "post-apocalyptic", 1.2, 1.1, 400, 12, 3},
		{"weapon_flamer", "Salvaged Flamer", "post-apocalyptic", 1.4, 0.7, 1500, 20, 7},
	}

	for _, w := range weapons {
		custom := &VehicleCustomization{
			ID:              w.id,
			Name:            w.name,
			Category:        CategoryWeapons,
			Slot:            SlotWeaponPrimary,
			Cost:            w.cost,
			Weight:          w.weight,
			DamageMod:       w.damage,
			RangeMod:        w.wepRange,
			SpeedMod:        1.0,
			AccelerationMod: 1.0,
			FuelEfficiency:  1.0,
			HandlingMod:     1.0,
			ArmorMod:        1.0,
			MinLevel:        w.minLevel,
		}
		s.Catalog[w.id] = custom
		s.GenreCatalogs[w.genre] = append(s.GenreCatalogs[w.genre], w.id)
	}
}

// RegisterVehicle registers a vehicle for customization tracking.
func (s *VehicleCustomizationSystem) RegisterVehicle(entity ecs.Entity, vehicleType string, maxWeight float64) *VehicleCustomizationState {
	s.mu.Lock()
	defer s.mu.Unlock()

	state := &VehicleCustomizationState{
		EntityID:       entity,
		VehicleType:    vehicleType,
		InstalledMods:  make(map[CustomizationSlot]*VehicleCustomization),
		PrimaryColor:   0xFFFFFF,
		SecondaryColor: 0x808080,
		DecalIDs:       make([]string, 0),
		MaxWeight:      maxWeight,
		Level:          1,
		NextLevelXP:    100,
	}

	s.Vehicles[entity] = state
	return state
}

// GetVehicleState returns the customization state for a vehicle.
func (s *VehicleCustomizationSystem) GetVehicleState(entity ecs.Entity) *VehicleCustomizationState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Vehicles[entity]
}

// GetAvailableCustomizations returns customizations available for the current genre.
func (s *VehicleCustomizationSystem) GetAvailableCustomizations(entity ecs.Entity) []*VehicleCustomization {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state := s.Vehicles[entity]
	if state == nil {
		return nil
	}

	ids := s.GenreCatalogs[s.Genre]
	result := make([]*VehicleCustomization, 0, len(ids))

	for _, id := range ids {
		if custom := s.Catalog[id]; custom != nil {
			if custom.MinLevel <= state.Level {
				result = append(result, custom)
			}
		}
	}

	return result
}

// InstallCustomization installs a customization on a vehicle.
func (s *VehicleCustomizationSystem) InstallCustomization(entity ecs.Entity, customizationID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state := s.Vehicles[entity]
	if state == nil {
		return fmt.Errorf("vehicle not registered")
	}

	custom := s.Catalog[customizationID]
	if custom == nil {
		return fmt.Errorf("customization not found: %s", customizationID)
	}

	// Validate installation requirements
	if err := s.validateInstallation(state, custom); err != nil {
		return err
	}

	// Perform installation
	s.applyCustomization(state, custom)

	return nil
}

// validateInstallation checks if the customization can be installed.
func (s *VehicleCustomizationSystem) validateInstallation(state *VehicleCustomizationState, custom *VehicleCustomization) error {
	// Check level requirement
	if custom.MinLevel > state.Level {
		return fmt.Errorf("requires level %d, current level %d", custom.MinLevel, state.Level)
	}

	// Check weight capacity
	newWeight := state.TotalWeight + custom.Weight
	if newWeight > state.MaxWeight {
		return fmt.Errorf("exceeds weight capacity: %.1f/%.1f", newWeight, state.MaxWeight)
	}

	// Check for incompatibilities
	return s.checkIncompatibilities(state, custom)
}

// checkIncompatibilities verifies no incompatible mods are installed.
func (s *VehicleCustomizationSystem) checkIncompatibilities(state *VehicleCustomizationState, custom *VehicleCustomization) error {
	for _, incompatID := range custom.Incompatible {
		for _, installed := range state.InstalledMods {
			if installed.ID == incompatID {
				return fmt.Errorf("incompatible with installed: %s", installed.Name)
			}
		}
	}
	return nil
}

// applyCustomization installs the customization to the vehicle.
func (s *VehicleCustomizationSystem) applyCustomization(state *VehicleCustomizationState, custom *VehicleCustomization) {
	// Uninstall existing mod in the same slot
	if existing := state.InstalledMods[custom.Slot]; existing != nil {
		state.TotalWeight -= existing.Weight
	}

	// Install the new customization
	state.InstalledMods[custom.Slot] = custom
	state.TotalWeight += custom.Weight

	// Update colors if appearance item
	if custom.Category == CategoryAppearance && custom.PrimaryColor != 0 {
		state.PrimaryColor = custom.PrimaryColor
	}
}

// UninstallCustomization removes a customization from a vehicle.
func (s *VehicleCustomizationSystem) UninstallCustomization(entity ecs.Entity, slot CustomizationSlot) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state := s.Vehicles[entity]
	if state == nil {
		return fmt.Errorf("vehicle not registered")
	}

	existing := state.InstalledMods[slot]
	if existing == nil {
		return fmt.Errorf("no customization in slot: %s", slot)
	}

	state.TotalWeight -= existing.Weight
	delete(state.InstalledMods, slot)

	return nil
}

// GetModifiedStats returns the cumulative stat modifiers for a vehicle.
func (s *VehicleCustomizationSystem) GetModifiedStats(entity ecs.Entity) (speed, accel, handling, armor, fuel, damage, weaponRange float64, cargo, passengers int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Default values (no modification)
	speed, accel, handling, armor, fuel, damage, weaponRange = 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0
	cargo, passengers = 0, 0

	state := s.Vehicles[entity]
	if state == nil {
		return speed, accel, handling, armor, fuel, damage, weaponRange, cargo, passengers
	}

	for _, custom := range state.InstalledMods {
		speed *= custom.SpeedMod
		accel *= custom.AccelerationMod
		handling *= custom.HandlingMod
		armor *= custom.ArmorMod
		fuel *= custom.FuelEfficiency
		damage *= custom.DamageMod
		weaponRange *= custom.RangeMod
		cargo += custom.CargoCapacity
		passengers += custom.PassengerSeats
	}

	return speed, accel, handling, armor, fuel, damage, weaponRange, cargo, passengers
}

// SetVehicleName sets a custom name for a vehicle.
func (s *VehicleCustomizationSystem) SetVehicleName(entity ecs.Entity, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state := s.Vehicles[entity]
	if state == nil {
		return fmt.Errorf("vehicle not registered")
	}

	state.CustomName = name
	return nil
}

// SetColors sets the primary and secondary colors for a vehicle.
func (s *VehicleCustomizationSystem) SetColors(entity ecs.Entity, primary, secondary uint32) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state := s.Vehicles[entity]
	if state == nil {
		return fmt.Errorf("vehicle not registered")
	}

	state.PrimaryColor = primary
	state.SecondaryColor = secondary
	return nil
}

// AddDecal adds a decal to a vehicle.
func (s *VehicleCustomizationSystem) AddDecal(entity ecs.Entity, decalID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state := s.Vehicles[entity]
	if state == nil {
		return fmt.Errorf("vehicle not registered")
	}

	// Max 5 decals
	if len(state.DecalIDs) >= 5 {
		return fmt.Errorf("maximum decals reached")
	}

	state.DecalIDs = append(state.DecalIDs, decalID)
	return nil
}

// RemoveDecal removes a decal from a vehicle.
func (s *VehicleCustomizationSystem) RemoveDecal(entity ecs.Entity, decalID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state := s.Vehicles[entity]
	if state == nil {
		return fmt.Errorf("vehicle not registered")
	}

	for i, id := range state.DecalIDs {
		if id == decalID {
			state.DecalIDs = append(state.DecalIDs[:i], state.DecalIDs[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("decal not found: %s", decalID)
}

// AddExperience adds XP to a vehicle and handles leveling.
func (s *VehicleCustomizationSystem) AddExperience(entity ecs.Entity, xp float64) (leveledUp bool, newLevel int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state := s.Vehicles[entity]
	if state == nil {
		return false, 0
	}

	state.Experience += xp

	// Check for level up (max level 15)
	for state.Experience >= state.NextLevelXP && state.Level < 15 {
		state.Experience -= state.NextLevelXP
		state.Level++
		state.NextLevelXP = float64(state.Level) * 100 * 1.5 // Increasing XP requirements
		leveledUp = true
	}

	return leveledUp, state.Level
}

// GetInstalledCount returns the number of installed customizations.
func (s *VehicleCustomizationSystem) GetInstalledCount(entity ecs.Entity) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state := s.Vehicles[entity]
	if state == nil {
		return 0
	}
	return len(state.InstalledMods)
}

// GetVehicleLevel returns the customization level of a vehicle.
func (s *VehicleCustomizationSystem) GetVehicleLevel(entity ecs.Entity) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state := s.Vehicles[entity]
	if state == nil {
		return 0
	}
	return state.Level
}

// VehicleCount returns the number of registered vehicles.
func (s *VehicleCustomizationSystem) VehicleCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.Vehicles)
}

// CatalogSize returns the number of customizations in the catalog.
func (s *VehicleCustomizationSystem) CatalogSize() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.Catalog)
}

// GetCustomization returns a customization by ID.
func (s *VehicleCustomizationSystem) GetCustomization(id string) *VehicleCustomization {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Catalog[id]
}

// UnregisterVehicle removes a vehicle from the customization system.
func (s *VehicleCustomizationSystem) UnregisterVehicle(entity ecs.Entity) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.Vehicles, entity)
}

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

	// Sprinting bonus
	if mount.IsSprinting {
		speed *= 1.5
	}

	// Mood modifier
	moodMod := 0.5 + (mount.Mood / 200) // 0.5-1.0 based on mood
	speed *= moodMod

	// Hunger penalty
	if mount.Hunger > 80 {
		speed *= 0.7
	} else if mount.Hunger > 50 {
		speed *= 0.85
	}

	// Stamina penalty
	staminaPercent := mount.Stats.Stamina / mount.Stats.MaxStamina
	if staminaPercent < 0.2 {
		speed *= 0.6
	} else if staminaPercent < 0.5 {
		speed *= 0.8
	}

	// Level bonus
	speed *= 1.0 + float64(mount.Level-1)*0.02 // 2% per level

	// Bond level bonus
	speed *= 1.0 + float64(mount.BondLevel-1)*0.05 // 5% per bond level

	return speed
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

// NavalVehicleType represents a type of naval vehicle.
type NavalVehicleType int

const (
	NavalRowboat NavalVehicleType = iota
	NavalSailboat
	NavalGalleon
	NavalFrigate
	NavalSubmarine
	NavalSpeedboat
	NavalHovercraft
	NavalAircraftCarrier
	NavalFishingBoat
	NavalYacht
	NavalWarship
	NavalRaft
)

// NavalVehicleArchetype defines properties of a naval vehicle type.
type NavalVehicleArchetype struct {
	Type          NavalVehicleType
	Name          string
	Genre         string
	MaxSpeed      float64
	Acceleration  float64
	TurnRate      float64
	Hull          int
	MaxHull       int
	CrewCapacity  int
	CargoCapacity float64
	FuelCapacity  float64
	FuelRate      float64
	CanSubmerge   bool
	HasSails      bool
	HasEngine     bool
	Cannons       int
	TorpedoTubes  int
	Anchored      bool
}

// NavalVehicleState represents the current state of a naval vehicle.
type NavalVehicleState struct {
	Entity           ecs.Entity
	Archetype        *NavalVehicleArchetype
	CurrentHull      int
	CurrentFuel      float64
	Speed            float64
	Heading          float64
	Depth            float64
	IsSubmerged      bool
	IsAnchored       bool
	WindDirection    float64
	WindStrength     float64
	Crew             []ecs.Entity
	Cargo            map[string]float64
	CannonCooldowns  []float64
	TorpedoCooldowns []float64
}

// NavalVehicleSystem manages naval vehicles (boats, ships, submarines).
type NavalVehicleSystem struct {
	mu         sync.RWMutex
	Archetypes map[NavalVehicleType]*NavalVehicleArchetype
	Vessels    map[ecs.Entity]*NavalVehicleState
	Genre      string
}

// NewNavalVehicleSystem creates a new naval vehicle system.
func NewNavalVehicleSystem(genre string) *NavalVehicleSystem {
	s := &NavalVehicleSystem{
		Archetypes: make(map[NavalVehicleType]*NavalVehicleArchetype),
		Vessels:    make(map[ecs.Entity]*NavalVehicleState),
		Genre:      genre,
	}
	s.initializeArchetypes()
	return s
}

// initializeArchetypes sets up naval vehicle archetypes based on genre.
func (s *NavalVehicleSystem) initializeArchetypes() {
	switch s.Genre {
	case "fantasy":
		s.initializeFantasyVessels()
	case "sci-fi":
		s.initializeSciFiVessels()
	case "horror":
		s.initializeHorrorVessels()
	case "cyberpunk":
		s.initializeCyberpunkVessels()
	case "post-apocalyptic":
		s.initializePostApocVessels()
	default:
		s.initializeFantasyVessels()
	}
}

func (s *NavalVehicleSystem) initializeFantasyVessels() {
	s.Archetypes[NavalRowboat] = &NavalVehicleArchetype{
		Type: NavalRowboat, Name: "Rowboat", Genre: "fantasy",
		MaxSpeed: 5, Acceleration: 2, TurnRate: 1.5, MaxHull: 50, Hull: 50,
		CrewCapacity: 4, CargoCapacity: 100, FuelCapacity: 0, FuelRate: 0,
		HasSails: false, HasEngine: false,
	}
	s.Archetypes[NavalSailboat] = &NavalVehicleArchetype{
		Type: NavalSailboat, Name: "Sailboat", Genre: "fantasy",
		MaxSpeed: 15, Acceleration: 3, TurnRate: 0.8, MaxHull: 150, Hull: 150,
		CrewCapacity: 8, CargoCapacity: 500, FuelCapacity: 0, FuelRate: 0,
		HasSails: true, HasEngine: false,
	}
	s.Archetypes[NavalGalleon] = &NavalVehicleArchetype{
		Type: NavalGalleon, Name: "Galleon", Genre: "fantasy",
		MaxSpeed: 20, Acceleration: 2, TurnRate: 0.4, MaxHull: 500, Hull: 500,
		CrewCapacity: 100, CargoCapacity: 5000, FuelCapacity: 0, FuelRate: 0,
		HasSails: true, HasEngine: false, Cannons: 24,
	}
	s.Archetypes[NavalFrigate] = &NavalVehicleArchetype{
		Type: NavalFrigate, Name: "War Frigate", Genre: "fantasy",
		MaxSpeed: 25, Acceleration: 3, TurnRate: 0.5, MaxHull: 400, Hull: 400,
		CrewCapacity: 80, CargoCapacity: 2000, FuelCapacity: 0, FuelRate: 0,
		HasSails: true, HasEngine: false, Cannons: 36,
	}
}

func (s *NavalVehicleSystem) initializeSciFiVessels() {
	s.Archetypes[NavalSubmarine] = &NavalVehicleArchetype{
		Type: NavalSubmarine, Name: "Attack Submarine", Genre: "sci-fi",
		MaxSpeed: 40, Acceleration: 8, TurnRate: 0.6, MaxHull: 600, Hull: 600,
		CrewCapacity: 50, CargoCapacity: 1000, FuelCapacity: 1000, FuelRate: 0.5,
		CanSubmerge: true, HasEngine: true, TorpedoTubes: 6,
	}
	s.Archetypes[NavalHovercraft] = &NavalVehicleArchetype{
		Type: NavalHovercraft, Name: "Hovercraft", Genre: "sci-fi",
		MaxSpeed: 60, Acceleration: 15, TurnRate: 1.2, MaxHull: 200, Hull: 200,
		CrewCapacity: 20, CargoCapacity: 800, FuelCapacity: 500, FuelRate: 1.0,
		HasEngine: true,
	}
	s.Archetypes[NavalAircraftCarrier] = &NavalVehicleArchetype{
		Type: NavalAircraftCarrier, Name: "Carrier", Genre: "sci-fi",
		MaxSpeed: 35, Acceleration: 2, TurnRate: 0.2, MaxHull: 2000, Hull: 2000,
		CrewCapacity: 500, CargoCapacity: 20000, FuelCapacity: 5000, FuelRate: 2.0,
		HasEngine: true, Cannons: 12,
	}
	s.Archetypes[NavalSpeedboat] = &NavalVehicleArchetype{
		Type: NavalSpeedboat, Name: "Patrol Boat", Genre: "sci-fi",
		MaxSpeed: 80, Acceleration: 20, TurnRate: 1.5, MaxHull: 100, Hull: 100,
		CrewCapacity: 6, CargoCapacity: 200, FuelCapacity: 300, FuelRate: 1.5,
		HasEngine: true,
	}
}

func (s *NavalVehicleSystem) initializeHorrorVessels() {
	s.Archetypes[NavalRaft] = &NavalVehicleArchetype{
		Type: NavalRaft, Name: "Cursed Raft", Genre: "horror",
		MaxSpeed: 3, Acceleration: 1, TurnRate: 0.5, MaxHull: 30, Hull: 30,
		CrewCapacity: 2, CargoCapacity: 50, FuelCapacity: 0, FuelRate: 0,
	}
	s.Archetypes[NavalFishingBoat] = &NavalVehicleArchetype{
		Type: NavalFishingBoat, Name: "Ghost Ship", Genre: "horror",
		MaxSpeed: 12, Acceleration: 2, TurnRate: 0.6, MaxHull: 120, Hull: 120,
		CrewCapacity: 10, CargoCapacity: 400, FuelCapacity: 0, FuelRate: 0,
		HasSails: true,
	}
	s.Archetypes[NavalGalleon] = &NavalVehicleArchetype{
		Type: NavalGalleon, Name: "Phantom Galleon", Genre: "horror",
		MaxSpeed: 18, Acceleration: 2, TurnRate: 0.3, MaxHull: 450, Hull: 450,
		CrewCapacity: 80, CargoCapacity: 4000, FuelCapacity: 0, FuelRate: 0,
		HasSails: true, Cannons: 20,
	}
}

func (s *NavalVehicleSystem) initializeCyberpunkVessels() {
	s.Archetypes[NavalSpeedboat] = &NavalVehicleArchetype{
		Type: NavalSpeedboat, Name: "Hydrojet", Genre: "cyberpunk",
		MaxSpeed: 100, Acceleration: 25, TurnRate: 1.8, MaxHull: 80, Hull: 80,
		CrewCapacity: 4, CargoCapacity: 150, FuelCapacity: 200, FuelRate: 2.0,
		HasEngine: true,
	}
	s.Archetypes[NavalYacht] = &NavalVehicleArchetype{
		Type: NavalYacht, Name: "Luxury Yacht", Genre: "cyberpunk",
		MaxSpeed: 45, Acceleration: 10, TurnRate: 0.7, MaxHull: 300, Hull: 300,
		CrewCapacity: 30, CargoCapacity: 2000, FuelCapacity: 800, FuelRate: 1.0,
		HasEngine: true,
	}
	s.Archetypes[NavalSubmarine] = &NavalVehicleArchetype{
		Type: NavalSubmarine, Name: "Stealth Sub", Genre: "cyberpunk",
		MaxSpeed: 50, Acceleration: 12, TurnRate: 0.8, MaxHull: 400, Hull: 400,
		CrewCapacity: 20, CargoCapacity: 600, FuelCapacity: 600, FuelRate: 0.8,
		CanSubmerge: true, HasEngine: true, TorpedoTubes: 4,
	}
}

func (s *NavalVehicleSystem) initializePostApocVessels() {
	s.Archetypes[NavalRaft] = &NavalVehicleArchetype{
		Type: NavalRaft, Name: "Scrap Raft", Genre: "post-apocalyptic",
		MaxSpeed: 4, Acceleration: 1.5, TurnRate: 0.8, MaxHull: 40, Hull: 40,
		CrewCapacity: 3, CargoCapacity: 80, FuelCapacity: 0, FuelRate: 0,
	}
	s.Archetypes[NavalFishingBoat] = &NavalVehicleArchetype{
		Type: NavalFishingBoat, Name: "Salvage Trawler", Genre: "post-apocalyptic",
		MaxSpeed: 10, Acceleration: 2, TurnRate: 0.5, MaxHull: 100, Hull: 100,
		CrewCapacity: 8, CargoCapacity: 600, FuelCapacity: 200, FuelRate: 0.3,
		HasEngine: true,
	}
	s.Archetypes[NavalWarship] = &NavalVehicleArchetype{
		Type: NavalWarship, Name: "Raider Warship", Genre: "post-apocalyptic",
		MaxSpeed: 25, Acceleration: 4, TurnRate: 0.4, MaxHull: 350, Hull: 350,
		CrewCapacity: 50, CargoCapacity: 1500, FuelCapacity: 400, FuelRate: 0.6,
		HasEngine: true, Cannons: 8,
	}
}

// Update processes naval vehicles each tick.
func (s *NavalVehicleSystem) Update(w *ecs.World, dt float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, vessel := range s.Vessels {
		s.updateVessel(vessel, dt)
	}
}

func (s *NavalVehicleSystem) updateVessel(vessel *NavalVehicleState, dt float64) {
	if vessel.IsAnchored {
		vessel.Speed = 0
		return
	}

	s.updateFuel(vessel, dt)
	s.updateSpeed(vessel, dt)
	s.updateCooldowns(vessel, dt)
}

func (s *NavalVehicleSystem) updateFuel(vessel *NavalVehicleState, dt float64) {
	if !vessel.Archetype.HasEngine {
		return
	}
	if vessel.Speed > 0 && vessel.CurrentFuel > 0 {
		fuelUsed := vessel.Archetype.FuelRate * (vessel.Speed / vessel.Archetype.MaxSpeed) * dt
		vessel.CurrentFuel -= fuelUsed
		if vessel.CurrentFuel < 0 {
			vessel.CurrentFuel = 0
		}
	}
}

func (s *NavalVehicleSystem) updateSpeed(vessel *NavalVehicleState, dt float64) {
	if vessel.Archetype.HasEngine && vessel.CurrentFuel <= 0 {
		vessel.Speed *= 0.95
		if vessel.Speed < 0.1 {
			vessel.Speed = 0
		}
		return
	}

	if vessel.Archetype.HasSails {
		windEffect := s.calculateWindEffect(vessel)
		vessel.Speed = vessel.Archetype.MaxSpeed * windEffect
	}
}

func (s *NavalVehicleSystem) calculateWindEffect(vessel *NavalVehicleState) float64 {
	angleDiff := math.Abs(vessel.Heading - vessel.WindDirection)
	if angleDiff > math.Pi {
		angleDiff = 2*math.Pi - angleDiff
	}
	effect := 0.2 + 0.8*math.Cos(angleDiff/2)
	return effect * (vessel.WindStrength / 100.0)
}

func (s *NavalVehicleSystem) updateCooldowns(vessel *NavalVehicleState, dt float64) {
	for i := range vessel.CannonCooldowns {
		if vessel.CannonCooldowns[i] > 0 {
			vessel.CannonCooldowns[i] -= dt
		}
	}
	for i := range vessel.TorpedoCooldowns {
		if vessel.TorpedoCooldowns[i] > 0 {
			vessel.TorpedoCooldowns[i] -= dt
		}
	}
}

// SpawnVessel creates a new naval vehicle of the given type.
func (s *NavalVehicleSystem) SpawnVessel(entity ecs.Entity, vesselType NavalVehicleType) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	archetype, ok := s.Archetypes[vesselType]
	if !ok {
		return fmt.Errorf("unknown vessel type: %d", vesselType)
	}

	vessel := &NavalVehicleState{
		Entity:           entity,
		Archetype:        archetype,
		CurrentHull:      archetype.MaxHull,
		CurrentFuel:      archetype.FuelCapacity,
		Speed:            0,
		Heading:          0,
		Depth:            0,
		WindDirection:    0,
		WindStrength:     50,
		Crew:             make([]ecs.Entity, 0, archetype.CrewCapacity),
		Cargo:            make(map[string]float64),
		CannonCooldowns:  make([]float64, archetype.Cannons),
		TorpedoCooldowns: make([]float64, archetype.TorpedoTubes),
	}
	s.Vessels[entity] = vessel
	return nil
}

// GetVessel returns the state of a naval vehicle.
func (s *NavalVehicleSystem) GetVessel(entity ecs.Entity) *NavalVehicleState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Vessels[entity]
}

// SetThrottle sets the speed as a fraction of max speed (0.0 to 1.0).
func (s *NavalVehicleSystem) SetThrottle(entity ecs.Entity, throttle float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	vessel := s.Vessels[entity]
	if vessel == nil {
		return fmt.Errorf("vessel not found")
	}
	if vessel.IsAnchored {
		return fmt.Errorf("vessel is anchored")
	}
	if throttle < 0 {
		throttle = 0
	} else if throttle > 1 {
		throttle = 1
	}
	vessel.Speed = vessel.Archetype.MaxSpeed * throttle
	return nil
}

// SetHeading sets the vessel's heading in radians.
func (s *NavalVehicleSystem) SetHeading(entity ecs.Entity, heading float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	vessel := s.Vessels[entity]
	if vessel == nil {
		return fmt.Errorf("vessel not found")
	}
	vessel.Heading = heading
	return nil
}

// Turn changes the heading by delta radians, limited by turn rate.
func (s *NavalVehicleSystem) Turn(entity ecs.Entity, delta, dt float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	vessel := s.Vessels[entity]
	if vessel == nil {
		return fmt.Errorf("vessel not found")
	}
	maxTurn := vessel.Archetype.TurnRate * dt
	if delta > maxTurn {
		delta = maxTurn
	} else if delta < -maxTurn {
		delta = -maxTurn
	}
	vessel.Heading += delta
	for vessel.Heading > 2*math.Pi {
		vessel.Heading -= 2 * math.Pi
	}
	for vessel.Heading < 0 {
		vessel.Heading += 2 * math.Pi
	}
	return nil
}

// DropAnchor anchors the vessel in place.
func (s *NavalVehicleSystem) DropAnchor(entity ecs.Entity) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	vessel := s.Vessels[entity]
	if vessel == nil {
		return fmt.Errorf("vessel not found")
	}
	if vessel.IsSubmerged {
		return fmt.Errorf("cannot anchor while submerged")
	}
	vessel.IsAnchored = true
	vessel.Speed = 0
	return nil
}

// RaiseAnchor releases the anchor.
func (s *NavalVehicleSystem) RaiseAnchor(entity ecs.Entity) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	vessel := s.Vessels[entity]
	if vessel == nil {
		return fmt.Errorf("vessel not found")
	}
	vessel.IsAnchored = false
	return nil
}

// Submerge submerges the vessel if capable.
func (s *NavalVehicleSystem) Submerge(entity ecs.Entity, depth float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	vessel := s.Vessels[entity]
	if vessel == nil {
		return fmt.Errorf("vessel not found")
	}
	if !vessel.Archetype.CanSubmerge {
		return fmt.Errorf("vessel cannot submerge")
	}
	if vessel.IsAnchored {
		return fmt.Errorf("cannot submerge while anchored")
	}
	if depth < 0 {
		depth = 0
	}
	vessel.Depth = depth
	vessel.IsSubmerged = depth > 0
	return nil
}

// Surface brings the vessel to the surface.
func (s *NavalVehicleSystem) Surface(entity ecs.Entity) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	vessel := s.Vessels[entity]
	if vessel == nil {
		return fmt.Errorf("vessel not found")
	}
	vessel.Depth = 0
	vessel.IsSubmerged = false
	return nil
}

// BoardCrew adds a crew member to the vessel.
func (s *NavalVehicleSystem) BoardCrew(entity, crew ecs.Entity) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	vessel := s.Vessels[entity]
	if vessel == nil {
		return fmt.Errorf("vessel not found")
	}
	if len(vessel.Crew) >= vessel.Archetype.CrewCapacity {
		return fmt.Errorf("crew capacity full")
	}
	for _, c := range vessel.Crew {
		if c == crew {
			return fmt.Errorf("already aboard")
		}
	}
	vessel.Crew = append(vessel.Crew, crew)
	return nil
}

// DisembarkCrew removes a crew member from the vessel.
func (s *NavalVehicleSystem) DisembarkCrew(entity, crew ecs.Entity) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	vessel := s.Vessels[entity]
	if vessel == nil {
		return fmt.Errorf("vessel not found")
	}
	for i, c := range vessel.Crew {
		if c == crew {
			vessel.Crew = append(vessel.Crew[:i], vessel.Crew[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("crew member not found")
}

// LoadCargo adds cargo to the vessel.
func (s *NavalVehicleSystem) LoadCargo(entity ecs.Entity, item string, amount float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	vessel := s.Vessels[entity]
	if vessel == nil {
		return fmt.Errorf("vessel not found")
	}
	currentCargo := s.calculateCurrentCargo(vessel)
	if currentCargo+amount > vessel.Archetype.CargoCapacity {
		return fmt.Errorf("cargo capacity exceeded")
	}
	vessel.Cargo[item] += amount
	return nil
}

// UnloadCargo removes cargo from the vessel.
func (s *NavalVehicleSystem) UnloadCargo(entity ecs.Entity, item string, amount float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	vessel := s.Vessels[entity]
	if vessel == nil {
		return fmt.Errorf("vessel not found")
	}
	if vessel.Cargo[item] < amount {
		return fmt.Errorf("not enough cargo")
	}
	vessel.Cargo[item] -= amount
	if vessel.Cargo[item] <= 0 {
		delete(vessel.Cargo, item)
	}
	return nil
}

func (s *NavalVehicleSystem) calculateCurrentCargo(vessel *NavalVehicleState) float64 {
	total := 0.0
	for _, amount := range vessel.Cargo {
		total += amount
	}
	return total
}

// GetCurrentCargo returns the total cargo weight.
func (s *NavalVehicleSystem) GetCurrentCargo(entity ecs.Entity) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	vessel := s.Vessels[entity]
	if vessel == nil {
		return 0
	}
	return s.calculateCurrentCargo(vessel)
}

// Refuel adds fuel to the vessel.
func (s *NavalVehicleSystem) Refuel(entity ecs.Entity, amount float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	vessel := s.Vessels[entity]
	if vessel == nil {
		return fmt.Errorf("vessel not found")
	}
	if vessel.Archetype.FuelCapacity == 0 {
		return fmt.Errorf("vessel does not use fuel")
	}
	vessel.CurrentFuel += amount
	if vessel.CurrentFuel > vessel.Archetype.FuelCapacity {
		vessel.CurrentFuel = vessel.Archetype.FuelCapacity
	}
	return nil
}

// DamageHull damages the vessel's hull.
func (s *NavalVehicleSystem) DamageHull(entity ecs.Entity, damage int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	vessel := s.Vessels[entity]
	if vessel == nil {
		return fmt.Errorf("vessel not found")
	}
	vessel.CurrentHull -= damage
	if vessel.CurrentHull < 0 {
		vessel.CurrentHull = 0
	}
	return nil
}

// RepairHull repairs the vessel's hull.
func (s *NavalVehicleSystem) RepairHull(entity ecs.Entity, repair int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	vessel := s.Vessels[entity]
	if vessel == nil {
		return fmt.Errorf("vessel not found")
	}
	vessel.CurrentHull += repair
	if vessel.CurrentHull > vessel.Archetype.MaxHull {
		vessel.CurrentHull = vessel.Archetype.MaxHull
	}
	return nil
}

// FireCannon fires a cannon if available and ready.
func (s *NavalVehicleSystem) FireCannon(entity ecs.Entity, cannonIndex int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	vessel := s.Vessels[entity]
	if vessel == nil {
		return fmt.Errorf("vessel not found")
	}
	if cannonIndex < 0 || cannonIndex >= len(vessel.CannonCooldowns) {
		return fmt.Errorf("invalid cannon index")
	}
	if vessel.CannonCooldowns[cannonIndex] > 0 {
		return fmt.Errorf("cannon on cooldown")
	}
	vessel.CannonCooldowns[cannonIndex] = 5.0
	return nil
}

// FireTorpedo fires a torpedo if available and ready.
func (s *NavalVehicleSystem) FireTorpedo(entity ecs.Entity, tubeIndex int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	vessel := s.Vessels[entity]
	if vessel == nil {
		return fmt.Errorf("vessel not found")
	}
	if !vessel.Archetype.CanSubmerge {
		return fmt.Errorf("vessel has no torpedo tubes")
	}
	if tubeIndex < 0 || tubeIndex >= len(vessel.TorpedoCooldowns) {
		return fmt.Errorf("invalid tube index")
	}
	if vessel.TorpedoCooldowns[tubeIndex] > 0 {
		return fmt.Errorf("torpedo tube on cooldown")
	}
	vessel.TorpedoCooldowns[tubeIndex] = 10.0
	return nil
}

// SetWind sets the wind conditions for a vessel.
func (s *NavalVehicleSystem) SetWind(entity ecs.Entity, direction, strength float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	vessel := s.Vessels[entity]
	if vessel == nil {
		return fmt.Errorf("vessel not found")
	}
	vessel.WindDirection = direction
	vessel.WindStrength = strength
	return nil
}

// DestroyVessel removes a vessel from the system.
func (s *NavalVehicleSystem) DestroyVessel(entity ecs.Entity) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.Vessels, entity)
}

// IsDestroyed checks if a vessel's hull is at zero.
func (s *NavalVehicleSystem) IsDestroyed(entity ecs.Entity) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	vessel := s.Vessels[entity]
	return vessel != nil && vessel.CurrentHull <= 0
}

// GetArchetypes returns all available archetypes for testing.
func (s *NavalVehicleSystem) GetArchetypes() map[NavalVehicleType]*NavalVehicleArchetype {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Archetypes
}

// FlyingVehicleType represents a type of flying vehicle.
type FlyingVehicleType int

const (
	FlyingGriffin FlyingVehicleType = iota
	FlyingDragon
	FlyingAirship
	FlyingHotAirBalloon
	FlyingMagicCarpet
	FlyingHelicopter
	FlyingJetpack
	FlyingHoverboard
	FlyingSpaceship
	FlyingDrone
	FlyingGlider
	FlyingBroomstick
	FlyingWingedHorse
)

// FlyingVehicleArchetype defines properties of a flying vehicle type.
type FlyingVehicleArchetype struct {
	Type          FlyingVehicleType
	Name          string
	Genre         string
	MaxSpeed      float64
	Acceleration  float64
	ClimbRate     float64
	DiveRate      float64
	TurnRate      float64
	MaxAltitude   float64
	MinAltitude   float64
	Health        int
	MaxHealth     int
	FuelCapacity  float64
	FuelRate      float64
	CargoCapacity float64
	Passengers    int
	HasWeapons    bool
	CanHover      bool
	RequiresPilot bool
}

// FlyingVehicleState represents the current state of a flying vehicle.
type FlyingVehicleState struct {
	Entity         ecs.Entity
	Archetype      *FlyingVehicleArchetype
	CurrentHealth  int
	CurrentFuel    float64
	Speed          float64
	Altitude       float64
	Heading        float64
	Pitch          float64
	IsFlying       bool
	IsHovering     bool
	IsLanding      bool
	IsTakingOff    bool
	Pilot          ecs.Entity
	Passengers     []ecs.Entity
	Cargo          map[string]float64
	WeaponCooldown float64
}

// FlyingVehicleSystem manages flying vehicles (aircraft, airships, flying mounts).
type FlyingVehicleSystem struct {
	mu         sync.RWMutex
	Archetypes map[FlyingVehicleType]*FlyingVehicleArchetype
	Aircraft   map[ecs.Entity]*FlyingVehicleState
	Genre      string
}

// NewFlyingVehicleSystem creates a new flying vehicle system.
func NewFlyingVehicleSystem(genre string) *FlyingVehicleSystem {
	s := &FlyingVehicleSystem{
		Archetypes: make(map[FlyingVehicleType]*FlyingVehicleArchetype),
		Aircraft:   make(map[ecs.Entity]*FlyingVehicleState),
		Genre:      genre,
	}
	s.initializeArchetypes()
	return s
}

// initializeArchetypes sets up flying vehicle archetypes based on genre.
func (s *FlyingVehicleSystem) initializeArchetypes() {
	switch s.Genre {
	case "fantasy":
		s.initializeFantasyAircraft()
	case "sci-fi":
		s.initializeSciFiAircraft()
	case "horror":
		s.initializeHorrorAircraft()
	case "cyberpunk":
		s.initializeCyberpunkAircraft()
	case "post-apocalyptic":
		s.initializePostApocAircraft()
	default:
		s.initializeFantasyAircraft()
	}
}

func (s *FlyingVehicleSystem) initializeFantasyAircraft() {
	s.Archetypes[FlyingGriffin] = &FlyingVehicleArchetype{
		Type: FlyingGriffin, Name: "Griffin", Genre: "fantasy",
		MaxSpeed: 60, Acceleration: 15, ClimbRate: 20, DiveRate: 40, TurnRate: 1.5,
		MaxAltitude: 500, MinAltitude: 0, MaxHealth: 200, Health: 200,
		FuelCapacity: 0, FuelRate: 0, CargoCapacity: 100, Passengers: 1,
		CanHover: true, RequiresPilot: true,
	}
	s.Archetypes[FlyingDragon] = &FlyingVehicleArchetype{
		Type: FlyingDragon, Name: "Dragon", Genre: "fantasy",
		MaxSpeed: 100, Acceleration: 20, ClimbRate: 30, DiveRate: 60, TurnRate: 1.0,
		MaxAltitude: 1000, MinAltitude: 0, MaxHealth: 500, Health: 500,
		FuelCapacity: 0, FuelRate: 0, CargoCapacity: 300, Passengers: 2,
		HasWeapons: true, CanHover: true, RequiresPilot: true,
	}
	s.Archetypes[FlyingAirship] = &FlyingVehicleArchetype{
		Type: FlyingAirship, Name: "Airship", Genre: "fantasy",
		MaxSpeed: 30, Acceleration: 5, ClimbRate: 5, DiveRate: 10, TurnRate: 0.3,
		MaxAltitude: 300, MinAltitude: 50, MaxHealth: 400, Health: 400,
		FuelCapacity: 100, FuelRate: 0.2, CargoCapacity: 2000, Passengers: 20,
		CanHover: true, RequiresPilot: true,
	}
	s.Archetypes[FlyingMagicCarpet] = &FlyingVehicleArchetype{
		Type: FlyingMagicCarpet, Name: "Magic Carpet", Genre: "fantasy",
		MaxSpeed: 40, Acceleration: 25, ClimbRate: 15, DiveRate: 20, TurnRate: 2.0,
		MaxAltitude: 200, MinAltitude: 0, MaxHealth: 50, Health: 50,
		FuelCapacity: 0, FuelRate: 0, CargoCapacity: 50, Passengers: 2,
		CanHover: true, RequiresPilot: true,
	}
	s.Archetypes[FlyingBroomstick] = &FlyingVehicleArchetype{
		Type: FlyingBroomstick, Name: "Broomstick", Genre: "fantasy",
		MaxSpeed: 50, Acceleration: 30, ClimbRate: 25, DiveRate: 35, TurnRate: 2.5,
		MaxAltitude: 150, MinAltitude: 0, MaxHealth: 30, Health: 30,
		FuelCapacity: 0, FuelRate: 0, CargoCapacity: 10, Passengers: 1,
		CanHover: true, RequiresPilot: true,
	}
	s.Archetypes[FlyingWingedHorse] = &FlyingVehicleArchetype{
		Type: FlyingWingedHorse, Name: "Pegasus", Genre: "fantasy",
		MaxSpeed: 70, Acceleration: 18, ClimbRate: 22, DiveRate: 45, TurnRate: 1.8,
		MaxAltitude: 600, MinAltitude: 0, MaxHealth: 180, Health: 180,
		FuelCapacity: 0, FuelRate: 0, CargoCapacity: 80, Passengers: 1,
		CanHover: false, RequiresPilot: true,
	}
}

func (s *FlyingVehicleSystem) initializeSciFiAircraft() {
	s.Archetypes[FlyingHelicopter] = &FlyingVehicleArchetype{
		Type: FlyingHelicopter, Name: "VTOL Craft", Genre: "sci-fi",
		MaxSpeed: 80, Acceleration: 12, ClimbRate: 15, DiveRate: 25, TurnRate: 1.2,
		MaxAltitude: 800, MinAltitude: 0, MaxHealth: 300, Health: 300,
		FuelCapacity: 500, FuelRate: 1.0, CargoCapacity: 500, Passengers: 6,
		HasWeapons: true, CanHover: true, RequiresPilot: true,
	}
	s.Archetypes[FlyingJetpack] = &FlyingVehicleArchetype{
		Type: FlyingJetpack, Name: "Jetpack", Genre: "sci-fi",
		MaxSpeed: 60, Acceleration: 40, ClimbRate: 30, DiveRate: 40, TurnRate: 3.0,
		MaxAltitude: 200, MinAltitude: 0, MaxHealth: 50, Health: 50,
		FuelCapacity: 100, FuelRate: 2.0, CargoCapacity: 20, Passengers: 0,
		CanHover: true, RequiresPilot: true,
	}
	s.Archetypes[FlyingSpaceship] = &FlyingVehicleArchetype{
		Type: FlyingSpaceship, Name: "Shuttle", Genre: "sci-fi",
		MaxSpeed: 200, Acceleration: 25, ClimbRate: 50, DiveRate: 50, TurnRate: 0.8,
		MaxAltitude: 10000, MinAltitude: 0, MaxHealth: 800, Health: 800,
		FuelCapacity: 2000, FuelRate: 3.0, CargoCapacity: 5000, Passengers: 20,
		HasWeapons: true, CanHover: true, RequiresPilot: true,
	}
	s.Archetypes[FlyingDrone] = &FlyingVehicleArchetype{
		Type: FlyingDrone, Name: "Combat Drone", Genre: "sci-fi",
		MaxSpeed: 100, Acceleration: 35, ClimbRate: 40, DiveRate: 50, TurnRate: 2.5,
		MaxAltitude: 500, MinAltitude: 5, MaxHealth: 100, Health: 100,
		FuelCapacity: 200, FuelRate: 0.5, CargoCapacity: 50, Passengers: 0,
		HasWeapons: true, CanHover: true, RequiresPilot: false,
	}
}

func (s *FlyingVehicleSystem) initializeHorrorAircraft() {
	s.Archetypes[FlyingBroomstick] = &FlyingVehicleArchetype{
		Type: FlyingBroomstick, Name: "Witch's Broom", Genre: "horror",
		MaxSpeed: 45, Acceleration: 25, ClimbRate: 20, DiveRate: 30, TurnRate: 2.0,
		MaxAltitude: 100, MinAltitude: 0, MaxHealth: 25, Health: 25,
		FuelCapacity: 0, FuelRate: 0, CargoCapacity: 10, Passengers: 1,
		CanHover: true, RequiresPilot: true,
	}
	s.Archetypes[FlyingGlider] = &FlyingVehicleArchetype{
		Type: FlyingGlider, Name: "Phantom Glider", Genre: "horror",
		MaxSpeed: 35, Acceleration: 5, ClimbRate: 0, DiveRate: 10, TurnRate: 0.8,
		MaxAltitude: 200, MinAltitude: 20, MaxHealth: 40, Health: 40,
		FuelCapacity: 0, FuelRate: 0, CargoCapacity: 20, Passengers: 1,
		CanHover: false, RequiresPilot: true,
	}
}

func (s *FlyingVehicleSystem) initializeCyberpunkAircraft() {
	s.Archetypes[FlyingHoverboard] = &FlyingVehicleArchetype{
		Type: FlyingHoverboard, Name: "Hoverboard", Genre: "cyberpunk",
		MaxSpeed: 50, Acceleration: 30, ClimbRate: 10, DiveRate: 15, TurnRate: 2.5,
		MaxAltitude: 50, MinAltitude: 1, MaxHealth: 40, Health: 40,
		FuelCapacity: 50, FuelRate: 0.3, CargoCapacity: 10, Passengers: 0,
		CanHover: true, RequiresPilot: true,
	}
	s.Archetypes[FlyingHelicopter] = &FlyingVehicleArchetype{
		Type: FlyingHelicopter, Name: "Corp Chopper", Genre: "cyberpunk",
		MaxSpeed: 90, Acceleration: 15, ClimbRate: 18, DiveRate: 30, TurnRate: 1.5,
		MaxAltitude: 600, MinAltitude: 0, MaxHealth: 350, Health: 350,
		FuelCapacity: 400, FuelRate: 1.2, CargoCapacity: 400, Passengers: 8,
		HasWeapons: true, CanHover: true, RequiresPilot: true,
	}
	s.Archetypes[FlyingDrone] = &FlyingVehicleArchetype{
		Type: FlyingDrone, Name: "Surveillance Drone", Genre: "cyberpunk",
		MaxSpeed: 80, Acceleration: 40, ClimbRate: 35, DiveRate: 45, TurnRate: 3.0,
		MaxAltitude: 300, MinAltitude: 5, MaxHealth: 60, Health: 60,
		FuelCapacity: 100, FuelRate: 0.3, CargoCapacity: 20, Passengers: 0,
		CanHover: true, RequiresPilot: false,
	}
}

func (s *FlyingVehicleSystem) initializePostApocAircraft() {
	s.Archetypes[FlyingGlider] = &FlyingVehicleArchetype{
		Type: FlyingGlider, Name: "Scrap Glider", Genre: "post-apocalyptic",
		MaxSpeed: 30, Acceleration: 3, ClimbRate: 0, DiveRate: 8, TurnRate: 0.6,
		MaxAltitude: 150, MinAltitude: 10, MaxHealth: 30, Health: 30,
		FuelCapacity: 0, FuelRate: 0, CargoCapacity: 30, Passengers: 1,
		CanHover: false, RequiresPilot: true,
	}
	s.Archetypes[FlyingHotAirBalloon] = &FlyingVehicleArchetype{
		Type: FlyingHotAirBalloon, Name: "Salvage Balloon", Genre: "post-apocalyptic",
		MaxSpeed: 15, Acceleration: 2, ClimbRate: 3, DiveRate: 5, TurnRate: 0.2,
		MaxAltitude: 200, MinAltitude: 50, MaxHealth: 80, Health: 80,
		FuelCapacity: 50, FuelRate: 0.5, CargoCapacity: 400, Passengers: 4,
		CanHover: true, RequiresPilot: true,
	}
	s.Archetypes[FlyingHelicopter] = &FlyingVehicleArchetype{
		Type: FlyingHelicopter, Name: "Junker Copter", Genre: "post-apocalyptic",
		MaxSpeed: 60, Acceleration: 8, ClimbRate: 10, DiveRate: 18, TurnRate: 1.0,
		MaxAltitude: 400, MinAltitude: 0, MaxHealth: 200, Health: 200,
		FuelCapacity: 200, FuelRate: 1.5, CargoCapacity: 300, Passengers: 4,
		HasWeapons: true, CanHover: true, RequiresPilot: true,
	}
}

// Update processes flying vehicles each tick.
func (s *FlyingVehicleSystem) Update(w *ecs.World, dt float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, aircraft := range s.Aircraft {
		s.updateAircraft(aircraft, dt)
	}
}

func (s *FlyingVehicleSystem) updateAircraft(aircraft *FlyingVehicleState, dt float64) {
	if !aircraft.IsFlying {
		s.handleGroundState(aircraft, dt)
		return
	}
	s.updateFuel(aircraft, dt)
	s.updateAltitude(aircraft, dt)
	s.updateWeaponCooldown(aircraft, dt)
}

func (s *FlyingVehicleSystem) handleGroundState(aircraft *FlyingVehicleState, dt float64) {
	if aircraft.IsTakingOff {
		aircraft.Altitude += aircraft.Archetype.ClimbRate * dt * 0.5
		if aircraft.Altitude >= aircraft.Archetype.MinAltitude+10 {
			aircraft.IsFlying = true
			aircraft.IsTakingOff = false
		}
	}
}

func (s *FlyingVehicleSystem) updateFuel(aircraft *FlyingVehicleState, dt float64) {
	if aircraft.Archetype.FuelCapacity == 0 {
		return
	}
	if aircraft.CurrentFuel > 0 {
		fuelUsed := aircraft.Archetype.FuelRate * dt
		if aircraft.IsHovering {
			fuelUsed *= 1.5
		}
		aircraft.CurrentFuel -= fuelUsed
		if aircraft.CurrentFuel < 0 {
			aircraft.CurrentFuel = 0
		}
	}
}

func (s *FlyingVehicleSystem) updateAltitude(aircraft *FlyingVehicleState, dt float64) {
	if aircraft.Archetype.FuelCapacity > 0 && aircraft.CurrentFuel <= 0 {
		aircraft.Altitude -= aircraft.Archetype.DiveRate * dt * 0.5
		if aircraft.Altitude < 0 {
			aircraft.Altitude = 0
			aircraft.IsFlying = false
		}
		return
	}

	if aircraft.IsLanding {
		aircraft.Altitude -= aircraft.Archetype.DiveRate * dt * 0.3
		if aircraft.Altitude <= 0 {
			aircraft.Altitude = 0
			aircraft.IsFlying = false
			aircraft.IsLanding = false
		}
	}
}

func (s *FlyingVehicleSystem) updateWeaponCooldown(aircraft *FlyingVehicleState, dt float64) {
	if aircraft.WeaponCooldown > 0 {
		aircraft.WeaponCooldown -= dt
		if aircraft.WeaponCooldown < 0 {
			aircraft.WeaponCooldown = 0
		}
	}
}

// SpawnAircraft creates a new flying vehicle of the given type.
func (s *FlyingVehicleSystem) SpawnAircraft(entity ecs.Entity, aircraftType FlyingVehicleType) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	archetype, ok := s.Archetypes[aircraftType]
	if !ok {
		return fmt.Errorf("unknown aircraft type: %d", aircraftType)
	}

	aircraft := &FlyingVehicleState{
		Entity:        entity,
		Archetype:     archetype,
		CurrentHealth: archetype.MaxHealth,
		CurrentFuel:   archetype.FuelCapacity,
		Speed:         0,
		Altitude:      0,
		Heading:       0,
		Pitch:         0,
		IsFlying:      false,
		Passengers:    make([]ecs.Entity, 0, archetype.Passengers),
		Cargo:         make(map[string]float64),
	}
	s.Aircraft[entity] = aircraft
	return nil
}

// GetAircraft returns the state of a flying vehicle.
func (s *FlyingVehicleSystem) GetAircraft(entity ecs.Entity) *FlyingVehicleState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Aircraft[entity]
}

// TakeOff initiates takeoff sequence.
func (s *FlyingVehicleSystem) TakeOff(entity ecs.Entity) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	aircraft := s.Aircraft[entity]
	if aircraft == nil {
		return fmt.Errorf("aircraft not found")
	}
	if aircraft.IsFlying {
		return fmt.Errorf("already flying")
	}
	if aircraft.Archetype.RequiresPilot && aircraft.Pilot == 0 {
		return fmt.Errorf("no pilot")
	}
	if aircraft.Archetype.FuelCapacity > 0 && aircraft.CurrentFuel <= 0 {
		return fmt.Errorf("no fuel")
	}
	aircraft.IsTakingOff = true
	return nil
}

// Land initiates landing sequence.
func (s *FlyingVehicleSystem) Land(entity ecs.Entity) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	aircraft := s.Aircraft[entity]
	if aircraft == nil {
		return fmt.Errorf("aircraft not found")
	}
	if !aircraft.IsFlying {
		return fmt.Errorf("not flying")
	}
	aircraft.IsLanding = true
	aircraft.IsHovering = false
	return nil
}

// Hover toggles hover mode if the aircraft can hover.
func (s *FlyingVehicleSystem) Hover(entity ecs.Entity, hover bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	aircraft := s.Aircraft[entity]
	if aircraft == nil {
		return fmt.Errorf("aircraft not found")
	}
	if !aircraft.Archetype.CanHover {
		return fmt.Errorf("aircraft cannot hover")
	}
	if !aircraft.IsFlying {
		return fmt.Errorf("not flying")
	}
	aircraft.IsHovering = hover
	if hover {
		aircraft.Speed = 0
	}
	return nil
}

// SetThrottle sets the speed as a fraction of max speed (0.0 to 1.0).
func (s *FlyingVehicleSystem) SetThrottle(entity ecs.Entity, throttle float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	aircraft := s.Aircraft[entity]
	if aircraft == nil {
		return fmt.Errorf("aircraft not found")
	}
	if !aircraft.IsFlying {
		return fmt.Errorf("not flying")
	}
	if aircraft.IsHovering {
		return fmt.Errorf("hovering")
	}
	if throttle < 0 {
		throttle = 0
	} else if throttle > 1 {
		throttle = 1
	}
	aircraft.Speed = aircraft.Archetype.MaxSpeed * throttle
	return nil
}

// SetHeading sets the aircraft's heading in radians.
func (s *FlyingVehicleSystem) SetHeading(entity ecs.Entity, heading float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	aircraft := s.Aircraft[entity]
	if aircraft == nil {
		return fmt.Errorf("aircraft not found")
	}
	aircraft.Heading = heading
	return nil
}

// Climb increases altitude up to max altitude.
func (s *FlyingVehicleSystem) Climb(entity ecs.Entity, dt float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	aircraft := s.Aircraft[entity]
	if aircraft == nil {
		return fmt.Errorf("aircraft not found")
	}
	if !aircraft.IsFlying {
		return fmt.Errorf("not flying")
	}
	if aircraft.Archetype.FuelCapacity > 0 && aircraft.CurrentFuel <= 0 {
		return fmt.Errorf("no fuel")
	}
	aircraft.Altitude += aircraft.Archetype.ClimbRate * dt
	if aircraft.Altitude > aircraft.Archetype.MaxAltitude {
		aircraft.Altitude = aircraft.Archetype.MaxAltitude
	}
	return nil
}

// Dive decreases altitude down to min altitude.
func (s *FlyingVehicleSystem) Dive(entity ecs.Entity, dt float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	aircraft := s.Aircraft[entity]
	if aircraft == nil {
		return fmt.Errorf("aircraft not found")
	}
	if !aircraft.IsFlying {
		return fmt.Errorf("not flying")
	}
	aircraft.Altitude -= aircraft.Archetype.DiveRate * dt
	if aircraft.Altitude < aircraft.Archetype.MinAltitude {
		aircraft.Altitude = aircraft.Archetype.MinAltitude
	}
	return nil
}

// BoardPilot sets the pilot.
func (s *FlyingVehicleSystem) BoardPilot(entity, pilot ecs.Entity) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	aircraft := s.Aircraft[entity]
	if aircraft == nil {
		return fmt.Errorf("aircraft not found")
	}
	if aircraft.Pilot != 0 {
		return fmt.Errorf("pilot seat occupied")
	}
	aircraft.Pilot = pilot
	return nil
}

// DisembarkPilot removes the pilot.
func (s *FlyingVehicleSystem) DisembarkPilot(entity ecs.Entity) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	aircraft := s.Aircraft[entity]
	if aircraft == nil {
		return fmt.Errorf("aircraft not found")
	}
	if aircraft.IsFlying {
		return fmt.Errorf("cannot disembark while flying")
	}
	aircraft.Pilot = 0
	return nil
}

// BoardPassenger adds a passenger.
func (s *FlyingVehicleSystem) BoardPassenger(entity, passenger ecs.Entity) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	aircraft := s.Aircraft[entity]
	if aircraft == nil {
		return fmt.Errorf("aircraft not found")
	}
	if len(aircraft.Passengers) >= aircraft.Archetype.Passengers {
		return fmt.Errorf("passenger capacity full")
	}
	for _, p := range aircraft.Passengers {
		if p == passenger {
			return fmt.Errorf("already aboard")
		}
	}
	aircraft.Passengers = append(aircraft.Passengers, passenger)
	return nil
}

// DisembarkPassenger removes a passenger.
func (s *FlyingVehicleSystem) DisembarkPassenger(entity, passenger ecs.Entity) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	aircraft := s.Aircraft[entity]
	if aircraft == nil {
		return fmt.Errorf("aircraft not found")
	}
	if aircraft.IsFlying {
		return fmt.Errorf("cannot disembark while flying")
	}
	for i, p := range aircraft.Passengers {
		if p == passenger {
			aircraft.Passengers = append(aircraft.Passengers[:i], aircraft.Passengers[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("passenger not found")
}

// LoadCargo adds cargo to the aircraft.
func (s *FlyingVehicleSystem) LoadCargo(entity ecs.Entity, item string, amount float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	aircraft := s.Aircraft[entity]
	if aircraft == nil {
		return fmt.Errorf("aircraft not found")
	}
	currentCargo := s.calculateCurrentCargo(aircraft)
	if currentCargo+amount > aircraft.Archetype.CargoCapacity {
		return fmt.Errorf("cargo capacity exceeded")
	}
	aircraft.Cargo[item] += amount
	return nil
}

// UnloadCargo removes cargo from the aircraft.
func (s *FlyingVehicleSystem) UnloadCargo(entity ecs.Entity, item string, amount float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	aircraft := s.Aircraft[entity]
	if aircraft == nil {
		return fmt.Errorf("aircraft not found")
	}
	if aircraft.Cargo[item] < amount {
		return fmt.Errorf("not enough cargo")
	}
	aircraft.Cargo[item] -= amount
	if aircraft.Cargo[item] <= 0 {
		delete(aircraft.Cargo, item)
	}
	return nil
}

func (s *FlyingVehicleSystem) calculateCurrentCargo(aircraft *FlyingVehicleState) float64 {
	total := 0.0
	for _, amount := range aircraft.Cargo {
		total += amount
	}
	return total
}

// GetCurrentCargo returns the total cargo weight.
func (s *FlyingVehicleSystem) GetCurrentCargo(entity ecs.Entity) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	aircraft := s.Aircraft[entity]
	if aircraft == nil {
		return 0
	}
	return s.calculateCurrentCargo(aircraft)
}

// Refuel adds fuel to the aircraft.
func (s *FlyingVehicleSystem) Refuel(entity ecs.Entity, amount float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	aircraft := s.Aircraft[entity]
	if aircraft == nil {
		return fmt.Errorf("aircraft not found")
	}
	if aircraft.Archetype.FuelCapacity == 0 {
		return fmt.Errorf("aircraft does not use fuel")
	}
	if aircraft.IsFlying {
		return fmt.Errorf("cannot refuel while flying")
	}
	aircraft.CurrentFuel += amount
	if aircraft.CurrentFuel > aircraft.Archetype.FuelCapacity {
		aircraft.CurrentFuel = aircraft.Archetype.FuelCapacity
	}
	return nil
}

// DamageAircraft damages the aircraft's health.
func (s *FlyingVehicleSystem) DamageAircraft(entity ecs.Entity, damage int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	aircraft := s.Aircraft[entity]
	if aircraft == nil {
		return fmt.Errorf("aircraft not found")
	}
	aircraft.CurrentHealth -= damage
	if aircraft.CurrentHealth < 0 {
		aircraft.CurrentHealth = 0
	}
	return nil
}

// RepairAircraft repairs the aircraft's health.
func (s *FlyingVehicleSystem) RepairAircraft(entity ecs.Entity, repair int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	aircraft := s.Aircraft[entity]
	if aircraft == nil {
		return fmt.Errorf("aircraft not found")
	}
	aircraft.CurrentHealth += repair
	if aircraft.CurrentHealth > aircraft.Archetype.MaxHealth {
		aircraft.CurrentHealth = aircraft.Archetype.MaxHealth
	}
	return nil
}

// FireWeapon fires the aircraft's weapon if available and ready.
func (s *FlyingVehicleSystem) FireWeapon(entity ecs.Entity) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	aircraft := s.Aircraft[entity]
	if aircraft == nil {
		return fmt.Errorf("aircraft not found")
	}
	if !aircraft.Archetype.HasWeapons {
		return fmt.Errorf("aircraft has no weapons")
	}
	if aircraft.WeaponCooldown > 0 {
		return fmt.Errorf("weapon on cooldown")
	}
	aircraft.WeaponCooldown = 2.0
	return nil
}

// DestroyAircraft removes an aircraft from the system.
func (s *FlyingVehicleSystem) DestroyAircraft(entity ecs.Entity) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.Aircraft, entity)
}

// IsDestroyed checks if an aircraft's health is at zero.
func (s *FlyingVehicleSystem) IsDestroyed(entity ecs.Entity) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	aircraft := s.Aircraft[entity]
	return aircraft != nil && aircraft.CurrentHealth <= 0
}

// GetArchetypes returns all available archetypes for testing.
func (s *FlyingVehicleSystem) GetArchetypes() map[FlyingVehicleType]*FlyingVehicleArchetype {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Archetypes
}

// Turn changes the heading by delta radians, limited by turn rate.
func (s *FlyingVehicleSystem) Turn(entity ecs.Entity, delta, dt float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	aircraft := s.Aircraft[entity]
	if aircraft == nil {
		return fmt.Errorf("aircraft not found")
	}
	if !aircraft.IsFlying {
		return fmt.Errorf("not flying")
	}
	maxTurn := aircraft.Archetype.TurnRate * dt
	if delta > maxTurn {
		delta = maxTurn
	} else if delta < -maxTurn {
		delta = -maxTurn
	}
	aircraft.Heading += delta
	for aircraft.Heading > 2*math.Pi {
		aircraft.Heading -= 2 * math.Pi
	}
	for aircraft.Heading < 0 {
		aircraft.Heading += 2 * math.Pi
	}
	return nil
}

// GetAltitude returns the aircraft's current altitude.
func (s *FlyingVehicleSystem) GetAltitude(entity ecs.Entity) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	aircraft := s.Aircraft[entity]
	if aircraft == nil {
		return 0
	}
	return aircraft.Altitude
}

// IsFlying checks if the aircraft is in flight.
func (s *FlyingVehicleSystem) IsFlying(entity ecs.Entity) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	aircraft := s.Aircraft[entity]
	return aircraft != nil && aircraft.IsFlying
}
