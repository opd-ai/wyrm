package systems

import (
	"math"
	"sync"
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

func TestNewFlyingVehicleSystem(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}
	for _, genre := range genres {
		sys := NewFlyingVehicleSystem(genre)
		if sys == nil {
			t.Errorf("NewFlyingVehicleSystem(%s) returned nil", genre)
			continue
		}
		if sys.Genre != genre {
			t.Errorf("expected genre %s, got %s", genre, sys.Genre)
		}
		if len(sys.Archetypes) == 0 {
			t.Errorf("no archetypes for genre %s", genre)
		}
	}
}

func TestFlyingArchetypes(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	archetypes := sys.GetArchetypes()

	tests := []struct {
		aircraftType FlyingVehicleType
		name         string
		canHover     bool
	}{
		{FlyingGriffin, "Griffin", true},
		{FlyingDragon, "Dragon", true},
		{FlyingAirship, "Airship", true},
		{FlyingMagicCarpet, "Magic Carpet", true},
	}

	for _, tc := range tests {
		arch, ok := archetypes[tc.aircraftType]
		if !ok {
			t.Errorf("archetype %d not found", tc.aircraftType)
			continue
		}
		if arch.Name != tc.name {
			t.Errorf("expected name %s, got %s", tc.name, arch.Name)
		}
		if arch.CanHover != tc.canHover {
			t.Errorf("expected canHover=%v for %s", tc.canHover, tc.name)
		}
	}
}

func TestSpawnAircraft(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	entity := ecs.Entity(1)

	err := sys.SpawnAircraft(entity, FlyingDragon)
	if err != nil {
		t.Fatalf("SpawnAircraft failed: %v", err)
	}

	aircraft := sys.GetAircraft(entity)
	if aircraft == nil {
		t.Fatal("GetAircraft returned nil")
	}
	if aircraft.Archetype.Type != FlyingDragon {
		t.Errorf("wrong archetype type")
	}
	if aircraft.CurrentHealth != aircraft.Archetype.MaxHealth {
		t.Errorf("health not at max")
	}
}

func TestSpawnUnknownAircraft(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	entity := ecs.Entity(1)

	err := sys.SpawnAircraft(entity, FlyingHelicopter)
	if err == nil {
		t.Error("expected error for unknown aircraft type")
	}
}

func TestTakeOffAndLand(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingDragon)

	pilot := ecs.Entity(100)
	sys.BoardPilot(entity, pilot)

	err := sys.TakeOff(entity)
	if err != nil {
		t.Fatalf("TakeOff failed: %v", err)
	}

	aircraft := sys.GetAircraft(entity)
	if !aircraft.IsTakingOff {
		t.Error("should be taking off")
	}

	for i := 0; i < 10; i++ {
		sys.Update(nil, 1.0)
	}

	if !aircraft.IsFlying {
		t.Error("should be flying after takeoff")
	}

	err = sys.Land(entity)
	if err != nil {
		t.Fatalf("Land failed: %v", err)
	}

	for i := 0; i < 100; i++ {
		sys.Update(nil, 1.0)
	}

	if aircraft.IsFlying {
		t.Error("should have landed")
	}
}

func TestTakeOffNoPilot(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingGriffin)

	err := sys.TakeOff(entity)
	if err == nil {
		t.Error("should require pilot")
	}
}

func TestTakeOffNoFuel(t *testing.T) {
	sys := NewFlyingVehicleSystem("sci-fi")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingHelicopter)

	pilot := ecs.Entity(100)
	sys.BoardPilot(entity, pilot)

	aircraft := sys.GetAircraft(entity)
	aircraft.CurrentFuel = 0

	err := sys.TakeOff(entity)
	if err == nil {
		t.Error("should require fuel")
	}
}

func TestHover(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingDragon)

	pilot := ecs.Entity(100)
	sys.BoardPilot(entity, pilot)
	sys.TakeOff(entity)

	aircraft := sys.GetAircraft(entity)
	aircraft.IsFlying = true
	aircraft.IsTakingOff = false
	aircraft.Speed = 50

	err := sys.Hover(entity, true)
	if err != nil {
		t.Fatalf("Hover failed: %v", err)
	}

	if !aircraft.IsHovering {
		t.Error("should be hovering")
	}
	if aircraft.Speed != 0 {
		t.Error("speed should be 0 while hovering")
	}

	sys.Hover(entity, false)
	if aircraft.IsHovering {
		t.Error("should stop hovering")
	}
}

func TestHoverNotFlying(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingDragon)

	err := sys.Hover(entity, true)
	if err == nil {
		t.Error("should not hover when not flying")
	}
}

func TestHoverCannotHover(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingWingedHorse)

	pilot := ecs.Entity(100)
	sys.BoardPilot(entity, pilot)

	aircraft := sys.GetAircraft(entity)
	aircraft.IsFlying = true

	err := sys.Hover(entity, true)
	if err == nil {
		t.Error("winged horse cannot hover")
	}
}

func TestThrottle(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingDragon)

	aircraft := sys.GetAircraft(entity)
	aircraft.IsFlying = true

	err := sys.SetThrottle(entity, 0.5)
	if err != nil {
		t.Fatalf("SetThrottle failed: %v", err)
	}

	expectedSpeed := aircraft.Archetype.MaxSpeed * 0.5
	if aircraft.Speed != expectedSpeed {
		t.Errorf("expected speed %f, got %f", expectedSpeed, aircraft.Speed)
	}
}

func TestThrottleWhileHovering(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingDragon)

	aircraft := sys.GetAircraft(entity)
	aircraft.IsFlying = true
	aircraft.IsHovering = true

	err := sys.SetThrottle(entity, 0.5)
	if err == nil {
		t.Error("should not throttle while hovering")
	}
}

func TestClimbAndDive(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingDragon)

	aircraft := sys.GetAircraft(entity)
	aircraft.IsFlying = true
	aircraft.Altitude = 100

	err := sys.Climb(entity, 1.0)
	if err != nil {
		t.Fatalf("Climb failed: %v", err)
	}
	if aircraft.Altitude <= 100 {
		t.Error("altitude should increase")
	}

	initialAlt := aircraft.Altitude
	err = sys.Dive(entity, 1.0)
	if err != nil {
		t.Fatalf("Dive failed: %v", err)
	}
	if aircraft.Altitude >= initialAlt {
		t.Error("altitude should decrease")
	}
}

func TestClimbNoFuel(t *testing.T) {
	sys := NewFlyingVehicleSystem("sci-fi")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingHelicopter)

	aircraft := sys.GetAircraft(entity)
	aircraft.IsFlying = true
	aircraft.CurrentFuel = 0

	err := sys.Climb(entity, 1.0)
	if err == nil {
		t.Error("should not climb without fuel")
	}
}

func TestAltitudeLimits(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingDragon)

	aircraft := sys.GetAircraft(entity)
	aircraft.IsFlying = true
	aircraft.Altitude = aircraft.Archetype.MaxAltitude - 10

	for i := 0; i < 10; i++ {
		sys.Climb(entity, 1.0)
	}

	if aircraft.Altitude > aircraft.Archetype.MaxAltitude {
		t.Error("should not exceed max altitude")
	}

	aircraft.Altitude = aircraft.Archetype.MinAltitude + 10
	for i := 0; i < 10; i++ {
		sys.Dive(entity, 1.0)
	}

	if aircraft.Altitude < aircraft.Archetype.MinAltitude {
		t.Error("should not go below min altitude")
	}
}

func TestBoardDisembarkPilot(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingGriffin)

	pilot := ecs.Entity(100)
	err := sys.BoardPilot(entity, pilot)
	if err != nil {
		t.Fatalf("BoardPilot failed: %v", err)
	}

	aircraft := sys.GetAircraft(entity)
	if aircraft.Pilot != pilot {
		t.Error("pilot not set")
	}

	err = sys.BoardPilot(entity, ecs.Entity(101))
	if err == nil {
		t.Error("should not allow second pilot")
	}

	err = sys.DisembarkPilot(entity)
	if err != nil {
		t.Fatalf("DisembarkPilot failed: %v", err)
	}
	if aircraft.Pilot != 0 {
		t.Error("pilot should be cleared")
	}
}

func TestDisembarkPilotWhileFlying(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingGriffin)

	pilot := ecs.Entity(100)
	sys.BoardPilot(entity, pilot)

	aircraft := sys.GetAircraft(entity)
	aircraft.IsFlying = true

	err := sys.DisembarkPilot(entity)
	if err == nil {
		t.Error("should not disembark while flying")
	}
}

func TestPassengerManagement(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingDragon)

	p1 := ecs.Entity(101)
	p2 := ecs.Entity(102)
	p3 := ecs.Entity(103)

	sys.BoardPassenger(entity, p1)
	sys.BoardPassenger(entity, p2)

	aircraft := sys.GetAircraft(entity)
	if len(aircraft.Passengers) != 2 {
		t.Errorf("expected 2 passengers, got %d", len(aircraft.Passengers))
	}

	err := sys.BoardPassenger(entity, p3)
	if err == nil {
		t.Error("should exceed capacity")
	}

	err = sys.BoardPassenger(entity, p1)
	if err == nil {
		t.Error("should not allow duplicate")
	}

	sys.DisembarkPassenger(entity, p1)
	if len(aircraft.Passengers) != 1 {
		t.Error("passenger should be removed")
	}
}

func TestDisembarkPassengerWhileFlying(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingDragon)

	p1 := ecs.Entity(101)
	sys.BoardPassenger(entity, p1)

	aircraft := sys.GetAircraft(entity)
	aircraft.IsFlying = true

	err := sys.DisembarkPassenger(entity, p1)
	if err == nil {
		t.Error("should not disembark while flying")
	}
}

func TestCargoManagement(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingAirship)

	err := sys.LoadCargo(entity, "gold", 500)
	if err != nil {
		t.Fatalf("LoadCargo failed: %v", err)
	}

	current := sys.GetCurrentCargo(entity)
	if current != 500 {
		t.Errorf("expected 500 cargo, got %f", current)
	}

	err = sys.LoadCargo(entity, "silk", 2000)
	if err == nil {
		t.Error("should exceed capacity")
	}

	err = sys.UnloadCargo(entity, "gold", 200)
	if err != nil {
		t.Fatalf("UnloadCargo failed: %v", err)
	}

	current = sys.GetCurrentCargo(entity)
	if current != 300 {
		t.Errorf("expected 300 cargo, got %f", current)
	}
}

func TestFuelManagement(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingAirship)

	aircraft := sys.GetAircraft(entity)
	aircraft.CurrentFuel = 50

	err := sys.Refuel(entity, 30)
	if err != nil {
		t.Fatalf("Refuel failed: %v", err)
	}
	if aircraft.CurrentFuel != 80 {
		t.Errorf("expected 80 fuel, got %f", aircraft.CurrentFuel)
	}

	sys.Refuel(entity, 1000)
	if aircraft.CurrentFuel != aircraft.Archetype.FuelCapacity {
		t.Error("fuel should cap at capacity")
	}
}

func TestRefuelWhileFlying(t *testing.T) {
	sys := NewFlyingVehicleSystem("sci-fi")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingHelicopter)

	aircraft := sys.GetAircraft(entity)
	aircraft.IsFlying = true

	err := sys.Refuel(entity, 100)
	if err == nil {
		t.Error("should not refuel while flying")
	}
}

func TestRefuelNoFuelAircraft(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingDragon)

	err := sys.Refuel(entity, 100)
	if err == nil {
		t.Error("dragon does not use fuel")
	}
}

func TestDamageAndRepair(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingDragon)

	aircraft := sys.GetAircraft(entity)
	maxHealth := aircraft.Archetype.MaxHealth

	sys.DamageAircraft(entity, 100)
	if aircraft.CurrentHealth != maxHealth-100 {
		t.Errorf("expected health %d, got %d", maxHealth-100, aircraft.CurrentHealth)
	}

	sys.RepairAircraft(entity, 50)
	if aircraft.CurrentHealth != maxHealth-50 {
		t.Errorf("expected health %d, got %d", maxHealth-50, aircraft.CurrentHealth)
	}

	sys.RepairAircraft(entity, 100)
	if aircraft.CurrentHealth != maxHealth {
		t.Error("repair should cap at max health")
	}
}

func TestDestruction(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingGriffin)

	if sys.IsDestroyed(entity) {
		t.Error("should not be destroyed initially")
	}

	sys.DamageAircraft(entity, 1000)
	if !sys.IsDestroyed(entity) {
		t.Error("should be destroyed after massive damage")
	}
}

func TestFireWeapon(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingDragon)

	err := sys.FireWeapon(entity)
	if err != nil {
		t.Fatalf("FireWeapon failed: %v", err)
	}

	aircraft := sys.GetAircraft(entity)
	if aircraft.WeaponCooldown != 2.0 {
		t.Errorf("expected cooldown 2.0, got %f", aircraft.WeaponCooldown)
	}

	err = sys.FireWeapon(entity)
	if err == nil {
		t.Error("weapon should be on cooldown")
	}
}

func TestFireWeaponNoWeapons(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingMagicCarpet)

	err := sys.FireWeapon(entity)
	if err == nil {
		t.Error("magic carpet has no weapons")
	}
}

func TestDestroyAircraft(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingGriffin)

	sys.DestroyAircraft(entity)
	if sys.GetAircraft(entity) != nil {
		t.Error("aircraft should be nil after destruction")
	}
}

func TestUpdateFuelConsumption(t *testing.T) {
	sys := NewFlyingVehicleSystem("sci-fi")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingHelicopter)

	aircraft := sys.GetAircraft(entity)
	aircraft.IsFlying = true
	initialFuel := aircraft.CurrentFuel

	sys.Update(nil, 1.0)

	if aircraft.CurrentFuel >= initialFuel {
		t.Error("fuel should decrease while flying")
	}
}

func TestUpdateHoverFuelConsumption(t *testing.T) {
	sys := NewFlyingVehicleSystem("sci-fi")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingHelicopter)

	aircraft := sys.GetAircraft(entity)
	aircraft.IsFlying = true
	aircraft.CurrentFuel = 100
	aircraft.IsHovering = true

	sys.Update(nil, 1.0)
	fuelAfterHover := aircraft.CurrentFuel

	aircraft.CurrentFuel = 100
	aircraft.IsHovering = false
	sys.Update(nil, 1.0)
	fuelAfterNormal := aircraft.CurrentFuel

	if fuelAfterHover >= fuelAfterNormal {
		t.Error("hovering should consume more fuel")
	}
}

func TestUpdateOutOfFuel(t *testing.T) {
	sys := NewFlyingVehicleSystem("sci-fi")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingHelicopter)

	aircraft := sys.GetAircraft(entity)
	aircraft.IsFlying = true
	aircraft.CurrentFuel = 0
	aircraft.Altitude = 500

	sys.Update(nil, 1.0)

	if aircraft.Altitude >= 500 {
		t.Error("should lose altitude when out of fuel")
	}
}

func TestUpdateWeaponCooldown(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingDragon)

	aircraft := sys.GetAircraft(entity)
	aircraft.IsFlying = true
	sys.FireWeapon(entity)
	initialCooldown := aircraft.WeaponCooldown

	sys.Update(nil, 1.0)

	if aircraft.WeaponCooldown >= initialCooldown {
		t.Error("cooldown should decrease")
	}
}

func TestTurn(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingDragon)

	aircraft := sys.GetAircraft(entity)
	aircraft.IsFlying = true
	aircraft.Heading = 0

	err := sys.Turn(entity, 0.1, 1.0)
	if err != nil {
		t.Fatalf("Turn failed: %v", err)
	}

	if aircraft.Heading <= 0 {
		t.Error("heading should increase")
	}
}

func TestTurnRateLimit(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingAirship)

	aircraft := sys.GetAircraft(entity)
	aircraft.IsFlying = true
	maxTurn := aircraft.Archetype.TurnRate * 1.0

	sys.SetHeading(entity, 0)
	sys.Turn(entity, 10.0, 1.0)

	if aircraft.Heading > maxTurn+0.001 {
		t.Errorf("turn exceeded turn rate limit")
	}
}

func TestTurnNotFlying(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingDragon)

	err := sys.Turn(entity, 0.1, 1.0)
	if err == nil {
		t.Error("should not turn when not flying")
	}
}

func TestConcurrentFlyingAccess(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingDragon)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sys.GetAircraft(entity)
			sys.GetAltitude(entity)
			sys.IsFlying(entity)
		}()
	}
	wg.Wait()
}

func TestSciFiArchetypes(t *testing.T) {
	sys := NewFlyingVehicleSystem("sci-fi")
	archetypes := sys.GetArchetypes()

	expectedTypes := []FlyingVehicleType{
		FlyingHelicopter,
		FlyingJetpack,
		FlyingSpaceship,
		FlyingDrone,
	}

	for _, ft := range expectedTypes {
		if _, ok := archetypes[ft]; !ok {
			t.Errorf("sci-fi missing archetype %d", ft)
		}
	}
}

func TestHorrorArchetypes(t *testing.T) {
	sys := NewFlyingVehicleSystem("horror")
	archetypes := sys.GetArchetypes()

	if _, ok := archetypes[FlyingBroomstick]; !ok {
		t.Error("horror missing broomstick")
	}
	if _, ok := archetypes[FlyingGlider]; !ok {
		t.Error("horror missing glider")
	}
}

func TestCyberpunkArchetypes(t *testing.T) {
	sys := NewFlyingVehicleSystem("cyberpunk")
	archetypes := sys.GetArchetypes()

	if _, ok := archetypes[FlyingHoverboard]; !ok {
		t.Error("cyberpunk missing hoverboard")
	}
	if _, ok := archetypes[FlyingHelicopter]; !ok {
		t.Error("cyberpunk missing helicopter")
	}
}

func TestPostApocArchetypes(t *testing.T) {
	sys := NewFlyingVehicleSystem("post-apocalyptic")
	archetypes := sys.GetArchetypes()

	if _, ok := archetypes[FlyingGlider]; !ok {
		t.Error("post-apoc missing glider")
	}
	if _, ok := archetypes[FlyingHotAirBalloon]; !ok {
		t.Error("post-apoc missing hot air balloon")
	}
}

func TestAircraftNotFound(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	entity := ecs.Entity(999)

	if sys.GetAircraft(entity) != nil {
		t.Error("should return nil for non-existent aircraft")
	}
	if err := sys.TakeOff(entity); err == nil {
		t.Error("TakeOff should fail")
	}
	if err := sys.Land(entity); err == nil {
		t.Error("Land should fail")
	}
	if err := sys.Hover(entity, true); err == nil {
		t.Error("Hover should fail")
	}
	if err := sys.SetThrottle(entity, 0.5); err == nil {
		t.Error("SetThrottle should fail")
	}
	if err := sys.SetHeading(entity, 0); err == nil {
		t.Error("SetHeading should fail")
	}
	if err := sys.Climb(entity, 1.0); err == nil {
		t.Error("Climb should fail")
	}
	if err := sys.Dive(entity, 1.0); err == nil {
		t.Error("Dive should fail")
	}
	if err := sys.BoardPilot(entity, ecs.Entity(1)); err == nil {
		t.Error("BoardPilot should fail")
	}
	if err := sys.DisembarkPilot(entity); err == nil {
		t.Error("DisembarkPilot should fail")
	}
	if err := sys.BoardPassenger(entity, ecs.Entity(1)); err == nil {
		t.Error("BoardPassenger should fail")
	}
	if err := sys.DisembarkPassenger(entity, ecs.Entity(1)); err == nil {
		t.Error("DisembarkPassenger should fail")
	}
	if err := sys.LoadCargo(entity, "gold", 100); err == nil {
		t.Error("LoadCargo should fail")
	}
	if err := sys.UnloadCargo(entity, "gold", 100); err == nil {
		t.Error("UnloadCargo should fail")
	}
	if err := sys.Refuel(entity, 100); err == nil {
		t.Error("Refuel should fail")
	}
	if err := sys.DamageAircraft(entity, 100); err == nil {
		t.Error("DamageAircraft should fail")
	}
	if err := sys.RepairAircraft(entity, 100); err == nil {
		t.Error("RepairAircraft should fail")
	}
	if err := sys.FireWeapon(entity); err == nil {
		t.Error("FireWeapon should fail")
	}
	if err := sys.Turn(entity, 0.1, 1.0); err == nil {
		t.Error("Turn should fail")
	}
}

func TestGetCurrentCargoEmpty(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	entity := ecs.Entity(999)

	if cargo := sys.GetCurrentCargo(entity); cargo != 0 {
		t.Error("GetCurrentCargo should return 0 for non-existent aircraft")
	}
}

func TestGetAltitudeEmpty(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	entity := ecs.Entity(999)

	if alt := sys.GetAltitude(entity); alt != 0 {
		t.Error("GetAltitude should return 0 for non-existent aircraft")
	}
}

func TestIsFlyingEmpty(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	entity := ecs.Entity(999)

	if sys.IsFlying(entity) {
		t.Error("IsFlying should return false for non-existent aircraft")
	}
}

func TestHeadingWrap(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingDragon)

	aircraft := sys.GetAircraft(entity)
	aircraft.IsFlying = true
	aircraft.Heading = 0

	sys.Turn(entity, -0.1, 1.0)

	if aircraft.Heading < 0 || aircraft.Heading > 2*math.Pi {
		t.Errorf("heading should wrap: %f", aircraft.Heading)
	}
}

func TestDisembarkPassengerNotFound(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingDragon)

	err := sys.DisembarkPassenger(entity, ecs.Entity(999))
	if err == nil {
		t.Error("should fail for passenger not on aircraft")
	}
}

func TestThrottleClamp(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingDragon)

	aircraft := sys.GetAircraft(entity)
	aircraft.IsFlying = true

	sys.SetThrottle(entity, 2.0)
	if aircraft.Speed != aircraft.Archetype.MaxSpeed {
		t.Error("throttle should clamp to 1.0")
	}

	sys.SetThrottle(entity, -1.0)
	if aircraft.Speed != 0 {
		t.Error("throttle should clamp to 0.0")
	}
}

func TestLandNotFlying(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingDragon)

	err := sys.Land(entity)
	if err == nil {
		t.Error("should not land when not flying")
	}
}

func TestTakeOffAlreadyFlying(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingDragon)

	pilot := ecs.Entity(100)
	sys.BoardPilot(entity, pilot)

	aircraft := sys.GetAircraft(entity)
	aircraft.IsFlying = true

	err := sys.TakeOff(entity)
	if err == nil {
		t.Error("should not take off when already flying")
	}
}

func TestDiveAndClimbNotFlying(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingDragon)

	if err := sys.Climb(entity, 1.0); err == nil {
		t.Error("should not climb when not flying")
	}
	if err := sys.Dive(entity, 1.0); err == nil {
		t.Error("should not dive when not flying")
	}
}

func TestThrottleNotFlying(t *testing.T) {
	sys := NewFlyingVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingDragon)

	err := sys.SetThrottle(entity, 0.5)
	if err == nil {
		t.Error("should not throttle when not flying")
	}
}

func TestDroneNoPilot(t *testing.T) {
	sys := NewFlyingVehicleSystem("sci-fi")
	entity := ecs.Entity(1)
	sys.SpawnAircraft(entity, FlyingDrone)

	aircraft := sys.GetAircraft(entity)
	aircraft.CurrentFuel = 100

	err := sys.TakeOff(entity)
	if err != nil {
		t.Errorf("drone should not require pilot: %v", err)
	}
}
