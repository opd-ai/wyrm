package systems

import (
	"fmt"
	"sync"

	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

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

// Update processes vehicle customization events each tick.
func (s *VehicleCustomizationSystem) Update(w *ecs.World, dt float64) {
	// Customization is event-driven; no per-tick processing required
}
