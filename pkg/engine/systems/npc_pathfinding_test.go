package systems

import (
	"math"
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

func TestNewNPCPathfindingSystem(t *testing.T) {
	system := NewNPCPathfindingSystem()
	if system == nil {
		t.Fatal("NewNPCPathfindingSystem returned nil")
	}
	if system.DefaultMoveSpeed <= 0 {
		t.Error("DefaultMoveSpeed should be positive")
	}
	if system.DefaultArrivalThreshold <= 0 {
		t.Error("DefaultArrivalThreshold should be positive")
	}
}

func TestNPCPathfindingComponent(t *testing.T) {
	path := &components.NPCPathfinding{
		TargetX:   100,
		TargetY:   200,
		HasTarget: true,
		MoveSpeed: 5.0,
	}

	if path.Type() != "NPCPathfinding" {
		t.Errorf("Type() = %v, want NPCPathfinding", path.Type())
	}
}

func TestSetActivityLocation(t *testing.T) {
	path := &components.NPCPathfinding{}

	SetActivityLocation(path, "working", 100, 200, "shop1")

	if path.ActivityLocations == nil {
		t.Fatal("ActivityLocations should be initialized")
	}

	loc, exists := path.ActivityLocations["working"]
	if !exists {
		t.Fatal("working location should exist")
	}
	if loc.X != 100 || loc.Y != 200 {
		t.Errorf("Location = (%v, %v), want (100, 200)", loc.X, loc.Y)
	}
	if loc.LocationID != "shop1" {
		t.Errorf("LocationID = %v, want shop1", loc.LocationID)
	}
}

func TestSetDirectTarget(t *testing.T) {
	path := &components.NPCPathfinding{}

	SetDirectTarget(path, 50, 75)

	if !path.HasTarget {
		t.Error("HasTarget should be true")
	}
	if !path.IsMoving {
		t.Error("IsMoving should be true")
	}
	if path.TargetX != 50 || path.TargetY != 75 {
		t.Errorf("Target = (%v, %v), want (50, 75)", path.TargetX, path.TargetY)
	}
	if len(path.CurrentPath) != 1 {
		t.Errorf("Path length = %d, want 1", len(path.CurrentPath))
	}
}

func TestClearPath(t *testing.T) {
	path := &components.NPCPathfinding{
		HasTarget:            true,
		IsMoving:             true,
		CurrentPath:          []components.Waypoint{{X: 1, Y: 2}},
		CurrentWaypointIndex: 0,
		StuckTime:            2.0,
	}

	ClearPath(path)

	if path.HasTarget {
		t.Error("HasTarget should be false")
	}
	if path.IsMoving {
		t.Error("IsMoving should be false")
	}
	if path.CurrentPath != nil {
		t.Error("CurrentPath should be nil")
	}
	if path.StuckTime != 0 {
		t.Error("StuckTime should be 0")
	}
}

func TestGetDistanceToTarget(t *testing.T) {
	pos := &components.Position{X: 0, Y: 0}
	path := &components.NPCPathfinding{
		TargetX:   3,
		TargetY:   4,
		HasTarget: true,
	}

	dist := GetDistanceToTarget(pos, path)
	expected := 5.0 // 3-4-5 triangle

	if math.Abs(dist-expected) > 0.001 {
		t.Errorf("Distance = %v, want %v", dist, expected)
	}

	// Test with no target
	path.HasTarget = false
	dist = GetDistanceToTarget(pos, path)
	if dist != 0 {
		t.Errorf("Distance with no target = %v, want 0", dist)
	}
}

func TestIsAtDestination(t *testing.T) {
	pos := &components.Position{X: 10, Y: 10}
	path := &components.NPCPathfinding{
		TargetX:   10.5,
		TargetY:   10,
		HasTarget: true,
	}

	// Within threshold
	if !IsAtDestination(pos, path, 1.0) {
		t.Error("Should be at destination with threshold 1.0")
	}

	// Outside threshold
	if IsAtDestination(pos, path, 0.1) {
		t.Error("Should not be at destination with threshold 0.1")
	}

	// No target means always at destination
	path.HasTarget = false
	if !IsAtDestination(pos, path, 0.1) {
		t.Error("Should be at destination with no target")
	}
}

func TestGenerateScheduleLocations(t *testing.T) {
	path := &components.NPCPathfinding{}

	GenerateScheduleLocations(path, 0, 0, 100, 100)

	if path.ActivityLocations == nil {
		t.Fatal("ActivityLocations should be initialized")
	}

	activities := []string{"sleeping", "resting", "eating", "working", "crafting", "trading"}
	for _, act := range activities {
		if _, exists := path.ActivityLocations[act]; !exists {
			t.Errorf("Activity %s should have a location", act)
		}
	}

	// Home activities should be at home location
	if path.ActivityLocations["sleeping"].X != 0 || path.ActivityLocations["sleeping"].Y != 0 {
		t.Error("sleeping should be at home location")
	}

	// Work activities should be at work location
	if path.ActivityLocations["working"].X != 100 || path.ActivityLocations["working"].Y != 100 {
		t.Error("working should be at work location")
	}
}

func TestNPCPathfindingSystemUpdate(t *testing.T) {
	system := NewNPCPathfindingSystem()
	world := ecs.NewWorld()

	// Create an NPC with schedule and pathfinding
	npc := world.CreateEntity()
	world.AddComponent(npc, &components.Position{X: 0, Y: 0})
	world.AddComponent(npc, &components.Schedule{
		CurrentActivity: "working",
		TimeSlots:       map[int]string{9: "working"},
	})

	path := &components.NPCPathfinding{
		MoveSpeed:        5.0,
		ArrivalThreshold: 1.0,
	}
	SetActivityLocation(path, "working", 10, 0, "shop")
	world.AddComponent(npc, path)

	// Run update to set target
	system.Update(world, 0)

	// Check that target was set
	pathComp, _ := world.GetComponent(npc, "NPCPathfinding")
	pathResult := pathComp.(*components.NPCPathfinding)
	if !pathResult.HasTarget {
		t.Error("Target should be set")
	}
	if pathResult.TargetX != 10 || pathResult.TargetY != 0 {
		t.Errorf("Target = (%v, %v), want (10, 0)", pathResult.TargetX, pathResult.TargetY)
	}

	// Run update to move NPC
	system.Update(world, 1.0) // 1 second, should move 5 units

	posComp, _ := world.GetComponent(npc, "Position")
	pos := posComp.(*components.Position)
	if pos.X < 4 || pos.X > 6 {
		t.Errorf("NPC X position = %v, expected around 5", pos.X)
	}
}

func TestNPCMovesToDestination(t *testing.T) {
	system := NewNPCPathfindingSystem()
	world := ecs.NewWorld()

	// Create an NPC
	npc := world.CreateEntity()
	world.AddComponent(npc, &components.Position{X: 0, Y: 0})

	path := &components.NPCPathfinding{
		MoveSpeed:        10.0,
		ArrivalThreshold: 0.5,
	}
	SetDirectTarget(path, 10, 0)
	world.AddComponent(npc, path)

	// Run multiple updates until NPC arrives
	for i := 0; i < 20; i++ {
		system.Update(world, 0.1)

		pathComp, _ := world.GetComponent(npc, "NPCPathfinding")
		pathResult := pathComp.(*components.NPCPathfinding)
		if !pathResult.IsMoving {
			break
		}
	}

	// Check NPC arrived
	posComp, _ := world.GetComponent(npc, "Position")
	pos := posComp.(*components.Position)
	dist := math.Sqrt(math.Pow(pos.X-10, 2) + math.Pow(pos.Y-0, 2))
	if dist > 1.0 {
		t.Errorf("NPC should be near destination, distance = %v", dist)
	}

	pathComp, _ := world.GetComponent(npc, "NPCPathfinding")
	pathResult := pathComp.(*components.NPCPathfinding)
	if pathResult.IsMoving {
		t.Error("NPC should stop moving when arrived")
	}
}

func TestNPCUpdatesFacing(t *testing.T) {
	system := NewNPCPathfindingSystem()
	world := ecs.NewWorld()

	// Create an NPC facing right
	npc := world.CreateEntity()
	world.AddComponent(npc, &components.Position{X: 0, Y: 0, Angle: 0})

	path := &components.NPCPathfinding{
		MoveSpeed:        5.0,
		ArrivalThreshold: 0.5,
	}
	// Target is above NPC (positive Y)
	SetDirectTarget(path, 0, 10)
	world.AddComponent(npc, path)

	system.Update(world, 0.1)

	posComp, _ := world.GetComponent(npc, "Position")
	pos := posComp.(*components.Position)

	// Angle should be close to PI/2 (facing north/up)
	expectedAngle := math.Pi / 2
	angleDiff := math.Abs(pos.Angle - expectedAngle)
	if angleDiff > 0.1 {
		t.Errorf("Angle = %v, expected around %v", pos.Angle, expectedAngle)
	}
}

func TestActivityTargetChange(t *testing.T) {
	system := NewNPCPathfindingSystem()
	world := ecs.NewWorld()

	npc := world.CreateEntity()
	world.AddComponent(npc, &components.Position{X: 0, Y: 0})

	sched := &components.Schedule{
		CurrentActivity: "working",
		TimeSlots:       map[int]string{9: "working", 18: "sleeping"},
	}
	world.AddComponent(npc, sched)

	path := &components.NPCPathfinding{
		MoveSpeed:        5.0,
		ArrivalThreshold: 1.0,
	}
	SetActivityLocation(path, "working", 100, 0, "shop")
	SetActivityLocation(path, "sleeping", 0, 0, "home")
	world.AddComponent(npc, path)

	// Initial target should be work
	system.Update(world, 0)
	pathComp, _ := world.GetComponent(npc, "NPCPathfinding")
	pathResult := pathComp.(*components.NPCPathfinding)
	if pathResult.TargetX != 100 {
		t.Errorf("Initial target should be work (100), got %v", pathResult.TargetX)
	}

	// Change activity to sleeping
	sched.CurrentActivity = "sleeping"

	system.Update(world, 0)
	pathComp, _ = world.GetComponent(npc, "NPCPathfinding")
	pathResult = pathComp.(*components.NPCPathfinding)
	if pathResult.TargetX != 0 {
		t.Errorf("Target should change to home (0), got %v", pathResult.TargetX)
	}
}

func TestStuckDetection(t *testing.T) {
	path := &components.NPCPathfinding{
		HasTarget:    true,
		IsMoving:     true,
		MaxStuckTime: 1.0,
		CurrentPath:  []components.Waypoint{{X: 100, Y: 100}},
	}

	pos := &components.Position{X: 0, Y: 0}
	system := NewNPCPathfindingSystem()

	// Simulate being stuck (not moving)
	// We can't directly test this without more infrastructure,
	// but we can verify the StuckTime accumulation logic exists
	if path.StuckTime != 0 {
		t.Error("Initial StuckTime should be 0")
	}

	// Move the NPC normally - stuck time should stay 0
	system.moveTowardTarget(pos, path, 0.1)
	if path.StuckTime != 0 {
		t.Error("StuckTime should be 0 when moving")
	}
}
