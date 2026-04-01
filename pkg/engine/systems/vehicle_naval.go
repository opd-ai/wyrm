package systems

import (
	"fmt"
	"math"
	"sync"

	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/seedutil"
)

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
	delta = seedutil.ClampDelta(delta, maxTurn)
	vessel.Heading = seedutil.NormalizeAngle(vessel.Heading + delta)
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
	crewList, err := boardPassengerToSlice(vessel.Crew, vessel.Archetype.CrewCapacity, crew)
	if err != nil {
		return err
	}
	vessel.Crew = crewList
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
	crewList, err := disembarkPassengerFromSlice(vessel.Crew, crew)
	if err != nil {
		err = fmt.Errorf("crew member not found")
	}
	if err == nil {
		vessel.Crew = crewList
	}
	return err
}

// LoadCargo adds cargo to the vessel.
func (s *NavalVehicleSystem) LoadCargo(entity ecs.Entity, item string, amount float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	vessel := s.Vessels[entity]
	if vessel == nil {
		return fmt.Errorf("vessel not found")
	}
	return loadCargoToContainer(vessel.Cargo, vessel.Archetype.CargoCapacity, item, amount)
}

// UnloadCargo removes cargo from the vessel.
func (s *NavalVehicleSystem) UnloadCargo(entity ecs.Entity, item string, amount float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	vessel := s.Vessels[entity]
	if vessel == nil {
		return fmt.Errorf("vessel not found")
	}
	return unloadCargoFromContainer(vessel.Cargo, item, amount)
}

func (s *NavalVehicleSystem) calculateCurrentCargo(vessel *NavalVehicleState) float64 {
	return calculateCargoTotal(vessel.Cargo)
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
	return refuelVehicle(&vessel.CurrentFuel, vessel.Archetype.FuelCapacity, amount)
}

// DamageHull damages the vessel's hull.
func (s *NavalVehicleSystem) DamageHull(entity ecs.Entity, damage int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	vessel := s.Vessels[entity]
	if vessel == nil {
		return fmt.Errorf("vessel not found")
	}
	damageVehicleHealth(&vessel.CurrentHull, damage)
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
	repairVehicleHealth(&vessel.CurrentHull, vessel.Archetype.MaxHull, repair)
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
