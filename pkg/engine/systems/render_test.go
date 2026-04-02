package systems

import (
	"math"
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

func TestNewRenderSystem(t *testing.T) {
	sys := NewRenderSystem()
	if sys.MaxTargetDistance <= 0 {
		t.Error("MaxTargetDistance should be positive")
	}
	if sys.TargetingAngleTolerance <= 0 {
		t.Error("TargetingAngleTolerance should be positive")
	}
}

func TestRenderSystem_TargetingDirectlyAhead(t *testing.T) {
	world := ecs.NewWorld()
	sys := NewRenderSystem()

	// Create player facing East (angle 0)
	player := world.CreateEntity()
	playerPos := &components.Position{X: 0.0, Y: 0.0, Z: 0.0, Angle: 0.0}
	world.AddComponent(player, playerPos)
	sys.SetPlayerEntity(player)

	// Create target directly ahead of player
	target := world.CreateEntity()
	targetPos := &components.Position{X: 3.0, Y: 0.0, Z: 0.0}
	targetEnv := &components.EnvironmentObject{
		Category:    components.ObjectCategoryInteractive,
		ObjectType:  "test",
		DisplayName: "Test Object",
	}
	world.AddComponent(target, targetPos)
	world.AddComponent(target, targetEnv)

	sys.Update(world, 0.016)

	if !sys.TargetValid {
		t.Error("Target directly ahead should be valid")
	}
	if sys.TargetEntity != target {
		t.Errorf("Expected target entity %d, got %d", target, sys.TargetEntity)
	}
}

func TestRenderSystem_TargetingBehind(t *testing.T) {
	world := ecs.NewWorld()
	sys := NewRenderSystem()

	// Create player facing East (angle 0)
	player := world.CreateEntity()
	playerPos := &components.Position{X: 0.0, Y: 0.0, Z: 0.0, Angle: 0.0}
	world.AddComponent(player, playerPos)
	sys.SetPlayerEntity(player)

	// Create target behind player
	target := world.CreateEntity()
	targetPos := &components.Position{X: -3.0, Y: 0.0, Z: 0.0}
	targetEnv := &components.EnvironmentObject{
		Category:    components.ObjectCategoryInteractive,
		ObjectType:  "test",
		DisplayName: "Test Object",
	}
	world.AddComponent(target, targetPos)
	world.AddComponent(target, targetEnv)

	sys.Update(world, 0.016)

	if sys.TargetValid {
		t.Error("Target behind player should not be valid")
	}
}

func TestRenderSystem_TargetingTooFar(t *testing.T) {
	world := ecs.NewWorld()
	sys := NewRenderSystem()
	sys.MaxTargetDistance = 5.0

	// Create player facing East (angle 0)
	player := world.CreateEntity()
	playerPos := &components.Position{X: 0.0, Y: 0.0, Z: 0.0, Angle: 0.0}
	world.AddComponent(player, playerPos)
	sys.SetPlayerEntity(player)

	// Create target beyond max distance
	target := world.CreateEntity()
	targetPos := &components.Position{X: 10.0, Y: 0.0, Z: 0.0}
	targetEnv := &components.EnvironmentObject{
		Category:    components.ObjectCategoryInteractive,
		ObjectType:  "test",
		DisplayName: "Test Object",
	}
	world.AddComponent(target, targetPos)
	world.AddComponent(target, targetEnv)

	sys.Update(world, 0.016)

	if sys.TargetValid {
		t.Error("Target beyond max distance should not be valid")
	}
}

func TestRenderSystem_TargetingSide(t *testing.T) {
	world := ecs.NewWorld()
	sys := NewRenderSystem()
	sys.TargetingAngleTolerance = 0.1 // ~5.7 degrees

	// Create player facing East (angle 0)
	player := world.CreateEntity()
	playerPos := &components.Position{X: 0.0, Y: 0.0, Z: 0.0, Angle: 0.0}
	world.AddComponent(player, playerPos)
	sys.SetPlayerEntity(player)

	// Create target 90 degrees to the side (North)
	target := world.CreateEntity()
	targetPos := &components.Position{X: 0.0, Y: 3.0, Z: 0.0}
	targetEnv := &components.EnvironmentObject{
		Category:    components.ObjectCategoryInteractive,
		ObjectType:  "test",
		DisplayName: "Test Object",
	}
	world.AddComponent(target, targetPos)
	world.AddComponent(target, targetEnv)

	sys.Update(world, 0.016)

	if sys.TargetValid {
		t.Error("Target to the side should not be valid with tight angle tolerance")
	}
}

func TestRenderSystem_CloserTargetPriority(t *testing.T) {
	world := ecs.NewWorld()
	sys := NewRenderSystem()

	// Create player facing East (angle 0)
	player := world.CreateEntity()
	playerPos := &components.Position{X: 0.0, Y: 0.0, Z: 0.0, Angle: 0.0}
	world.AddComponent(player, playerPos)
	sys.SetPlayerEntity(player)

	// Create far target
	farTarget := world.CreateEntity()
	farPos := &components.Position{X: 5.0, Y: 0.0, Z: 0.0}
	farEnv := &components.EnvironmentObject{
		Category:    components.ObjectCategoryInteractive,
		ObjectType:  "test",
		DisplayName: "Far Object",
	}
	world.AddComponent(farTarget, farPos)
	world.AddComponent(farTarget, farEnv)

	// Create close target
	closeTarget := world.CreateEntity()
	closePos := &components.Position{X: 2.0, Y: 0.0, Z: 0.0}
	closeEnv := &components.EnvironmentObject{
		Category:    components.ObjectCategoryInteractive,
		ObjectType:  "test",
		DisplayName: "Close Object",
	}
	world.AddComponent(closeTarget, closePos)
	world.AddComponent(closeTarget, closeEnv)

	sys.Update(world, 0.016)

	if !sys.TargetValid {
		t.Error("Should have a valid target")
	}
	if sys.TargetEntity != closeTarget {
		t.Errorf("Closer target should be prioritized, got entity %d instead of %d", sys.TargetEntity, closeTarget)
	}
}

func TestRenderSystem_GetTarget(t *testing.T) {
	sys := NewRenderSystem()
	sys.TargetEntity = 42
	sys.TargetValid = true

	entity, valid := sys.GetTarget()
	if !valid {
		t.Error("GetTarget should return valid=true when target is set")
	}
	if entity != 42 {
		t.Errorf("GetTarget returned wrong entity: got %d, want 42", entity)
	}
}

func TestRenderSystem_NoPlayer(t *testing.T) {
	world := ecs.NewWorld()
	sys := NewRenderSystem()
	// Don't set player entity

	// Should not panic
	sys.Update(world, 0.016)

	if sys.TargetValid {
		t.Error("Should not have valid target without player")
	}
}

func TestRenderSystem_TargetingDiagonal(t *testing.T) {
	world := ecs.NewWorld()
	sys := NewRenderSystem()
	sys.TargetingAngleTolerance = 0.2 // ~11 degrees

	// Create player facing NorthEast (angle π/4)
	player := world.CreateEntity()
	playerPos := &components.Position{X: 0.0, Y: 0.0, Z: 0.0, Angle: math.Pi / 4}
	world.AddComponent(player, playerPos)
	sys.SetPlayerEntity(player)

	// Create target in NorthEast direction
	target := world.CreateEntity()
	targetPos := &components.Position{X: 3.0, Y: 3.0, Z: 0.0}
	targetEnv := &components.EnvironmentObject{
		Category:    components.ObjectCategoryInteractive,
		ObjectType:  "test",
		DisplayName: "Test Object",
	}
	world.AddComponent(target, targetPos)
	world.AddComponent(target, targetEnv)

	sys.Update(world, 0.016)

	if !sys.TargetValid {
		t.Error("Target in diagonal look direction should be valid")
	}
}
