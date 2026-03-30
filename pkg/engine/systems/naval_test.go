package systems

import (
	"math"
	"sync"
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

func TestNewNavalVehicleSystem(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}
	for _, genre := range genres {
		sys := NewNavalVehicleSystem(genre)
		if sys == nil {
			t.Errorf("NewNavalVehicleSystem(%s) returned nil", genre)
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

func TestNavalArchetypes(t *testing.T) {
	sys := NewNavalVehicleSystem("fantasy")
	archetypes := sys.GetArchetypes()

	tests := []struct {
		vesselType NavalVehicleType
		name       string
		hasSails   bool
	}{
		{NavalRowboat, "Rowboat", false},
		{NavalSailboat, "Sailboat", true},
		{NavalGalleon, "Galleon", true},
		{NavalFrigate, "War Frigate", true},
	}

	for _, tc := range tests {
		arch, ok := archetypes[tc.vesselType]
		if !ok {
			t.Errorf("archetype %d not found", tc.vesselType)
			continue
		}
		if arch.Name != tc.name {
			t.Errorf("expected name %s, got %s", tc.name, arch.Name)
		}
		if arch.HasSails != tc.hasSails {
			t.Errorf("expected hasSails=%v for %s", tc.hasSails, tc.name)
		}
	}
}

func TestSpawnVessel(t *testing.T) {
	sys := NewNavalVehicleSystem("fantasy")
	entity := ecs.Entity(1)

	err := sys.SpawnVessel(entity, NavalGalleon)
	if err != nil {
		t.Fatalf("SpawnVessel failed: %v", err)
	}

	vessel := sys.GetVessel(entity)
	if vessel == nil {
		t.Fatal("GetVessel returned nil")
	}
	if vessel.Archetype.Type != NavalGalleon {
		t.Errorf("wrong archetype type")
	}
	if vessel.CurrentHull != vessel.Archetype.MaxHull {
		t.Errorf("hull not at max")
	}
}

func TestSpawnUnknownVessel(t *testing.T) {
	sys := NewNavalVehicleSystem("fantasy")
	entity := ecs.Entity(1)

	err := sys.SpawnVessel(entity, NavalSubmarine)
	if err == nil {
		t.Error("expected error for unknown vessel type")
	}
}

func TestVesselThrottle(t *testing.T) {
	sys := NewNavalVehicleSystem("sci-fi")
	entity := ecs.Entity(1)
	sys.SpawnVessel(entity, NavalSpeedboat)

	err := sys.SetThrottle(entity, 0.5)
	if err != nil {
		t.Fatalf("SetThrottle failed: %v", err)
	}

	vessel := sys.GetVessel(entity)
	expectedSpeed := vessel.Archetype.MaxSpeed * 0.5
	if vessel.Speed != expectedSpeed {
		t.Errorf("expected speed %f, got %f", expectedSpeed, vessel.Speed)
	}
}

func TestVesselThrottleClamp(t *testing.T) {
	sys := NewNavalVehicleSystem("sci-fi")
	entity := ecs.Entity(1)
	sys.SpawnVessel(entity, NavalSpeedboat)

	sys.SetThrottle(entity, 2.0)
	vessel := sys.GetVessel(entity)
	if vessel.Speed != vessel.Archetype.MaxSpeed {
		t.Errorf("throttle should clamp to 1.0")
	}

	sys.SetThrottle(entity, -1.0)
	if vessel.Speed != 0 {
		t.Errorf("throttle should clamp to 0.0")
	}
}

func TestVesselHeading(t *testing.T) {
	sys := NewNavalVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnVessel(entity, NavalSailboat)

	heading := math.Pi / 2
	err := sys.SetHeading(entity, heading)
	if err != nil {
		t.Fatalf("SetHeading failed: %v", err)
	}

	vessel := sys.GetVessel(entity)
	if vessel.Heading != heading {
		t.Errorf("expected heading %f, got %f", heading, vessel.Heading)
	}
}

func TestVesselTurn(t *testing.T) {
	sys := NewNavalVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnVessel(entity, NavalSailboat)

	sys.SetHeading(entity, 0)
	err := sys.Turn(entity, 0.1, 1.0)
	if err != nil {
		t.Fatalf("Turn failed: %v", err)
	}

	vessel := sys.GetVessel(entity)
	if vessel.Heading <= 0 {
		t.Error("heading should have increased")
	}
}

func TestVesselTurnRateLimit(t *testing.T) {
	sys := NewNavalVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnVessel(entity, NavalGalleon)

	vessel := sys.GetVessel(entity)
	maxTurn := vessel.Archetype.TurnRate * 1.0

	sys.SetHeading(entity, 0)
	sys.Turn(entity, 10.0, 1.0)

	if vessel.Heading > maxTurn+0.001 {
		t.Errorf("turn exceeded turn rate limit: %f > %f", vessel.Heading, maxTurn)
	}
}

func TestAnchor(t *testing.T) {
	sys := NewNavalVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnVessel(entity, NavalSailboat)

	sys.SetThrottle(entity, 1.0)
	sys.DropAnchor(entity)

	vessel := sys.GetVessel(entity)
	if !vessel.IsAnchored {
		t.Error("vessel should be anchored")
	}
	if vessel.Speed != 0 {
		t.Error("speed should be 0 when anchored")
	}

	err := sys.SetThrottle(entity, 0.5)
	if err == nil {
		t.Error("should not be able to throttle while anchored")
	}

	sys.RaiseAnchor(entity)
	if vessel.IsAnchored {
		t.Error("vessel should not be anchored after raising")
	}
}

func TestSubmerge(t *testing.T) {
	sys := NewNavalVehicleSystem("sci-fi")
	entity := ecs.Entity(1)
	sys.SpawnVessel(entity, NavalSubmarine)

	err := sys.Submerge(entity, 50)
	if err != nil {
		t.Fatalf("Submerge failed: %v", err)
	}

	vessel := sys.GetVessel(entity)
	if !vessel.IsSubmerged {
		t.Error("vessel should be submerged")
	}
	if vessel.Depth != 50 {
		t.Errorf("expected depth 50, got %f", vessel.Depth)
	}

	sys.Surface(entity)
	if vessel.IsSubmerged || vessel.Depth != 0 {
		t.Error("vessel should be surfaced")
	}
}

func TestSubmergeNonSubmarine(t *testing.T) {
	sys := NewNavalVehicleSystem("sci-fi")
	entity := ecs.Entity(1)
	sys.SpawnVessel(entity, NavalSpeedboat)

	err := sys.Submerge(entity, 50)
	if err == nil {
		t.Error("speedboat should not be able to submerge")
	}
}

func TestSubmergeWhileAnchored(t *testing.T) {
	sys := NewNavalVehicleSystem("sci-fi")
	entity := ecs.Entity(1)
	sys.SpawnVessel(entity, NavalSubmarine)

	sys.DropAnchor(entity)
	err := sys.Submerge(entity, 50)
	if err == nil {
		t.Error("should not submerge while anchored")
	}
}

func TestCrewManagement(t *testing.T) {
	sys := NewNavalVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnVessel(entity, NavalRowboat)

	crew1 := ecs.Entity(100)
	crew2 := ecs.Entity(101)
	crew3 := ecs.Entity(102)
	crew4 := ecs.Entity(103)
	crew5 := ecs.Entity(104)

	sys.BoardCrew(entity, crew1)
	sys.BoardCrew(entity, crew2)
	sys.BoardCrew(entity, crew3)
	sys.BoardCrew(entity, crew4)

	vessel := sys.GetVessel(entity)
	if len(vessel.Crew) != 4 {
		t.Errorf("expected 4 crew, got %d", len(vessel.Crew))
	}

	err := sys.BoardCrew(entity, crew5)
	if err == nil {
		t.Error("should not exceed crew capacity")
	}

	err = sys.BoardCrew(entity, crew1)
	if err == nil {
		t.Error("should not allow duplicate crew")
	}

	sys.DisembarkCrew(entity, crew2)
	if len(vessel.Crew) != 3 {
		t.Errorf("expected 3 crew after disembark, got %d", len(vessel.Crew))
	}
}

func TestNavalCargoManagement(t *testing.T) {
	sys := NewNavalVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnVessel(entity, NavalSailboat)

	err := sys.LoadCargo(entity, "gold", 200)
	if err != nil {
		t.Fatalf("LoadCargo failed: %v", err)
	}

	current := sys.GetCurrentCargo(entity)
	if current != 200 {
		t.Errorf("expected 200 cargo, got %f", current)
	}

	err = sys.LoadCargo(entity, "silk", 400)
	if err == nil {
		t.Error("should not exceed cargo capacity")
	}

	err = sys.UnloadCargo(entity, "gold", 50)
	if err != nil {
		t.Fatalf("UnloadCargo failed: %v", err)
	}

	current = sys.GetCurrentCargo(entity)
	if current != 150 {
		t.Errorf("expected 150 cargo after unload, got %f", current)
	}

	err = sys.UnloadCargo(entity, "gold", 200)
	if err == nil {
		t.Error("should not unload more than available")
	}
}

func TestNavalFuelManagement(t *testing.T) {
	sys := NewNavalVehicleSystem("sci-fi")
	entity := ecs.Entity(1)
	sys.SpawnVessel(entity, NavalSpeedboat)

	vessel := sys.GetVessel(entity)
	initialFuel := vessel.CurrentFuel
	vessel.CurrentFuel = 100

	err := sys.Refuel(entity, 50)
	if err != nil {
		t.Fatalf("Refuel failed: %v", err)
	}
	if vessel.CurrentFuel != 150 {
		t.Errorf("expected 150 fuel, got %f", vessel.CurrentFuel)
	}

	sys.Refuel(entity, 1000)
	if vessel.CurrentFuel != initialFuel {
		t.Errorf("fuel should cap at max capacity")
	}
}

func TestFuelNotUsed(t *testing.T) {
	sys := NewNavalVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnVessel(entity, NavalSailboat)

	err := sys.Refuel(entity, 50)
	if err == nil {
		t.Error("sailboat should not use fuel")
	}
}

func TestHullDamageAndRepair(t *testing.T) {
	sys := NewNavalVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnVessel(entity, NavalGalleon)

	vessel := sys.GetVessel(entity)
	maxHull := vessel.Archetype.MaxHull

	sys.DamageHull(entity, 100)
	if vessel.CurrentHull != maxHull-100 {
		t.Errorf("expected hull %d, got %d", maxHull-100, vessel.CurrentHull)
	}

	sys.RepairHull(entity, 50)
	if vessel.CurrentHull != maxHull-50 {
		t.Errorf("expected hull %d, got %d", maxHull-50, vessel.CurrentHull)
	}

	sys.RepairHull(entity, 100)
	if vessel.CurrentHull != maxHull {
		t.Errorf("repair should cap at max hull")
	}
}

func TestHullDestruction(t *testing.T) {
	sys := NewNavalVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnVessel(entity, NavalRowboat)

	if sys.IsDestroyed(entity) {
		t.Error("vessel should not be destroyed initially")
	}

	sys.DamageHull(entity, 1000)
	if !sys.IsDestroyed(entity) {
		t.Error("vessel should be destroyed after massive damage")
	}
}

func TestFireCannon(t *testing.T) {
	sys := NewNavalVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnVessel(entity, NavalGalleon)

	err := sys.FireCannon(entity, 0)
	if err != nil {
		t.Fatalf("FireCannon failed: %v", err)
	}

	vessel := sys.GetVessel(entity)
	if vessel.CannonCooldowns[0] != 5.0 {
		t.Errorf("expected cooldown 5.0, got %f", vessel.CannonCooldowns[0])
	}

	err = sys.FireCannon(entity, 0)
	if err == nil {
		t.Error("should not fire cannon on cooldown")
	}

	err = sys.FireCannon(entity, 100)
	if err == nil {
		t.Error("should not fire invalid cannon index")
	}
}

func TestFireTorpedo(t *testing.T) {
	sys := NewNavalVehicleSystem("sci-fi")
	entity := ecs.Entity(1)
	sys.SpawnVessel(entity, NavalSubmarine)

	err := sys.FireTorpedo(entity, 0)
	if err != nil {
		t.Fatalf("FireTorpedo failed: %v", err)
	}

	vessel := sys.GetVessel(entity)
	if vessel.TorpedoCooldowns[0] != 10.0 {
		t.Errorf("expected cooldown 10.0, got %f", vessel.TorpedoCooldowns[0])
	}

	err = sys.FireTorpedo(entity, 0)
	if err == nil {
		t.Error("should not fire torpedo on cooldown")
	}
}

func TestFireTorpedoNonSubmarine(t *testing.T) {
	sys := NewNavalVehicleSystem("sci-fi")
	entity := ecs.Entity(1)
	sys.SpawnVessel(entity, NavalSpeedboat)

	err := sys.FireTorpedo(entity, 0)
	if err == nil {
		t.Error("speedboat should not have torpedo tubes")
	}
}

func TestWindEffect(t *testing.T) {
	sys := NewNavalVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnVessel(entity, NavalSailboat)

	sys.SetWind(entity, 0, 100)
	sys.SetHeading(entity, 0)

	vessel := sys.GetVessel(entity)
	sys.Update(nil, 1.0)

	if vessel.Speed <= 0 {
		t.Error("sailboat should have speed with favorable wind")
	}
}

func TestDestroyVessel(t *testing.T) {
	sys := NewNavalVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnVessel(entity, NavalSailboat)

	sys.DestroyVessel(entity)
	vessel := sys.GetVessel(entity)
	if vessel != nil {
		t.Error("vessel should be nil after destruction")
	}
}

func TestNavalUpdateCooldowns(t *testing.T) {
	sys := NewNavalVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnVessel(entity, NavalGalleon)

	sys.FireCannon(entity, 0)
	vessel := sys.GetVessel(entity)
	initialCooldown := vessel.CannonCooldowns[0]

	sys.Update(nil, 1.0)

	if vessel.CannonCooldowns[0] >= initialCooldown {
		t.Error("cooldown should decrease over time")
	}
}

func TestNavalUpdateFuelConsumption(t *testing.T) {
	sys := NewNavalVehicleSystem("sci-fi")
	entity := ecs.Entity(1)
	sys.SpawnVessel(entity, NavalSpeedboat)

	vessel := sys.GetVessel(entity)
	initialFuel := vessel.CurrentFuel
	vessel.Speed = vessel.Archetype.MaxSpeed

	sys.Update(nil, 1.0)

	if vessel.CurrentFuel >= initialFuel {
		t.Error("fuel should decrease while moving")
	}
}

func TestAnchoredNoUpdate(t *testing.T) {
	sys := NewNavalVehicleSystem("sci-fi")
	entity := ecs.Entity(1)
	sys.SpawnVessel(entity, NavalSpeedboat)

	vessel := sys.GetVessel(entity)
	vessel.Speed = 50
	sys.DropAnchor(entity)

	sys.Update(nil, 1.0)

	if vessel.Speed != 0 {
		t.Error("anchored vessel should have 0 speed")
	}
}

func TestConcurrentNavalAccess(t *testing.T) {
	sys := NewNavalVehicleSystem("sci-fi")
	entity := ecs.Entity(1)
	sys.SpawnVessel(entity, NavalSpeedboat)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sys.GetVessel(entity)
			sys.SetThrottle(entity, 0.5)
			sys.Turn(entity, 0.1, 0.1)
		}()
	}
	wg.Wait()
}

func TestNavalSciFiArchetypes(t *testing.T) {
	sys := NewNavalVehicleSystem("sci-fi")
	archetypes := sys.GetArchetypes()

	expectedTypes := []NavalVehicleType{
		NavalSubmarine,
		NavalHovercraft,
		NavalAircraftCarrier,
		NavalSpeedboat,
	}

	for _, vt := range expectedTypes {
		if _, ok := archetypes[vt]; !ok {
			t.Errorf("sci-fi missing archetype %d", vt)
		}
	}
}

func TestNavalHorrorArchetypes(t *testing.T) {
	sys := NewNavalVehicleSystem("horror")
	archetypes := sys.GetArchetypes()

	if _, ok := archetypes[NavalRaft]; !ok {
		t.Error("horror missing raft")
	}
	if _, ok := archetypes[NavalFishingBoat]; !ok {
		t.Error("horror missing fishing boat")
	}
}

func TestNavalCyberpunkArchetypes(t *testing.T) {
	sys := NewNavalVehicleSystem("cyberpunk")
	archetypes := sys.GetArchetypes()

	if _, ok := archetypes[NavalSpeedboat]; !ok {
		t.Error("cyberpunk missing speedboat")
	}
	if _, ok := archetypes[NavalYacht]; !ok {
		t.Error("cyberpunk missing yacht")
	}
}

func TestNavalPostApocArchetypes(t *testing.T) {
	sys := NewNavalVehicleSystem("post-apocalyptic")
	archetypes := sys.GetArchetypes()

	if _, ok := archetypes[NavalRaft]; !ok {
		t.Error("post-apoc missing raft")
	}
	if _, ok := archetypes[NavalWarship]; !ok {
		t.Error("post-apoc missing warship")
	}
}

func TestVesselNotFound(t *testing.T) {
	sys := NewNavalVehicleSystem("fantasy")
	entity := ecs.Entity(999)

	if sys.GetVessel(entity) != nil {
		t.Error("should return nil for non-existent vessel")
	}
	if err := sys.SetThrottle(entity, 0.5); err == nil {
		t.Error("SetThrottle should fail for non-existent vessel")
	}
	if err := sys.SetHeading(entity, 0); err == nil {
		t.Error("SetHeading should fail for non-existent vessel")
	}
	if err := sys.Turn(entity, 0.1, 0.1); err == nil {
		t.Error("Turn should fail for non-existent vessel")
	}
	if err := sys.DropAnchor(entity); err == nil {
		t.Error("DropAnchor should fail for non-existent vessel")
	}
	if err := sys.RaiseAnchor(entity); err == nil {
		t.Error("RaiseAnchor should fail for non-existent vessel")
	}
	if err := sys.Submerge(entity, 50); err == nil {
		t.Error("Submerge should fail for non-existent vessel")
	}
	if err := sys.Surface(entity); err == nil {
		t.Error("Surface should fail for non-existent vessel")
	}
	if err := sys.BoardCrew(entity, ecs.Entity(1)); err == nil {
		t.Error("BoardCrew should fail for non-existent vessel")
	}
	if err := sys.DisembarkCrew(entity, ecs.Entity(1)); err == nil {
		t.Error("DisembarkCrew should fail for non-existent vessel")
	}
	if err := sys.LoadCargo(entity, "gold", 100); err == nil {
		t.Error("LoadCargo should fail for non-existent vessel")
	}
	if err := sys.UnloadCargo(entity, "gold", 100); err == nil {
		t.Error("UnloadCargo should fail for non-existent vessel")
	}
	if err := sys.Refuel(entity, 100); err == nil {
		t.Error("Refuel should fail for non-existent vessel")
	}
	if err := sys.DamageHull(entity, 100); err == nil {
		t.Error("DamageHull should fail for non-existent vessel")
	}
	if err := sys.RepairHull(entity, 100); err == nil {
		t.Error("RepairHull should fail for non-existent vessel")
	}
	if err := sys.FireCannon(entity, 0); err == nil {
		t.Error("FireCannon should fail for non-existent vessel")
	}
	if err := sys.FireTorpedo(entity, 0); err == nil {
		t.Error("FireTorpedo should fail for non-existent vessel")
	}
	if err := sys.SetWind(entity, 0, 50); err == nil {
		t.Error("SetWind should fail for non-existent vessel")
	}
}

func TestNavalGetCurrentCargoEmpty(t *testing.T) {
	sys := NewNavalVehicleSystem("fantasy")
	entity := ecs.Entity(999)

	if cargo := sys.GetCurrentCargo(entity); cargo != 0 {
		t.Error("GetCurrentCargo should return 0 for non-existent vessel")
	}
}

func TestNavalHeadingWrap(t *testing.T) {
	sys := NewNavalVehicleSystem("sci-fi")
	entity := ecs.Entity(1)
	sys.SpawnVessel(entity, NavalHovercraft)

	vessel := sys.GetVessel(entity)
	vessel.Heading = 0
	sys.Turn(entity, -0.1, 1.0)

	if vessel.Heading < 0 || vessel.Heading > 2*math.Pi {
		t.Errorf("heading should wrap: %f", vessel.Heading)
	}
}

func TestEngineOutOfFuel(t *testing.T) {
	sys := NewNavalVehicleSystem("sci-fi")
	entity := ecs.Entity(1)
	sys.SpawnVessel(entity, NavalSpeedboat)

	vessel := sys.GetVessel(entity)
	vessel.CurrentFuel = 0
	vessel.Speed = 50

	sys.Update(nil, 1.0)

	if vessel.Speed >= 50 {
		t.Error("speed should decrease when out of fuel")
	}
}

func TestAnchorWhileSubmerged(t *testing.T) {
	sys := NewNavalVehicleSystem("sci-fi")
	entity := ecs.Entity(1)
	sys.SpawnVessel(entity, NavalSubmarine)

	sys.Submerge(entity, 50)
	err := sys.DropAnchor(entity)
	if err == nil {
		t.Error("should not be able to anchor while submerged")
	}
}

func TestDisembarkCrewNotFound(t *testing.T) {
	sys := NewNavalVehicleSystem("fantasy")
	entity := ecs.Entity(1)
	sys.SpawnVessel(entity, NavalSailboat)

	err := sys.DisembarkCrew(entity, ecs.Entity(999))
	if err == nil {
		t.Error("should fail to disembark crew not on vessel")
	}
}
