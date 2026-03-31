package systems

import (
	"fmt"
	"sync"

	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/util"
)

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
	delta = util.ClampDelta(delta, maxTurn)
	aircraft.Heading = util.NormalizeAngle(aircraft.Heading + delta)
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
