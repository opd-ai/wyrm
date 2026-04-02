//go:build noebiten

package raycast

import (
	"math"
	"testing"
)

func TestNewInteractionRaycaster(t *testing.T) {
	r := NewRenderer(320, 200)
	ir := NewInteractionRaycaster(r, 5.0)

	if ir.renderer != r {
		t.Error("renderer reference should be set")
	}
	if ir.MaxRange != 5.0 {
		t.Errorf("expected max range 5.0, got %f", ir.MaxRange)
	}
	if len(ir.Targets) != 0 {
		t.Error("targets should be empty initially")
	}
}

func TestNewInteractionRaycasterDefaultRange(t *testing.T) {
	r := NewRenderer(320, 200)
	ir := NewInteractionRaycaster(r, 0)

	if ir.MaxRange != 5.0 {
		t.Errorf("expected default max range 5.0, got %f", ir.MaxRange)
	}
}

func TestInteractionRaycasterSetTargets(t *testing.T) {
	r := NewRenderer(320, 200)
	ir := NewInteractionRaycaster(r, 5.0)

	targets := []InteractionTarget{
		{EntityID: 1, WorldX: 5, WorldY: 5, Radius: 0.5, InteractionRange: 2.0},
		{EntityID: 2, WorldX: 6, WorldY: 6, Radius: 0.5, InteractionRange: 2.0},
	}
	ir.SetTargets(targets)

	if len(ir.Targets) != 2 {
		t.Errorf("expected 2 targets, got %d", len(ir.Targets))
	}
}

func TestInteractionRaycasterAddTarget(t *testing.T) {
	r := NewRenderer(320, 200)
	ir := NewInteractionRaycaster(r, 5.0)

	ir.AddTarget(InteractionTarget{EntityID: 1, WorldX: 5, WorldY: 5, Radius: 0.5})
	ir.AddTarget(InteractionTarget{EntityID: 2, WorldX: 6, WorldY: 6, Radius: 0.5})

	if len(ir.Targets) != 2 {
		t.Errorf("expected 2 targets, got %d", len(ir.Targets))
	}
}

func TestInteractionRaycasterClearTargets(t *testing.T) {
	r := NewRenderer(320, 200)
	ir := NewInteractionRaycaster(r, 5.0)

	ir.AddTarget(InteractionTarget{EntityID: 1})
	ir.AddTarget(InteractionTarget{EntityID: 2})
	ir.ClearTargets()

	if len(ir.Targets) != 0 {
		t.Errorf("expected 0 targets after clear, got %d", len(ir.Targets))
	}
}

func TestCastInteractionRayNoTargets(t *testing.T) {
	r := NewRenderer(320, 200)
	ir := NewInteractionRaycaster(r, 5.0)

	result := ir.CastInteractionRay()
	if result != nil {
		t.Error("expected nil when no targets")
	}
}

func TestCastInteractionRayNilRenderer(t *testing.T) {
	ir := &InteractionRaycaster{
		renderer: nil,
		MaxRange: 5.0,
		Targets:  []InteractionTarget{{EntityID: 1}},
	}

	result := ir.CastInteractionRay()
	if result != nil {
		t.Error("expected nil when renderer is nil")
	}
}

func TestCastInteractionRayTargetInFront(t *testing.T) {
	r := NewRenderer(320, 200)
	// Player at position without interior walls, facing right
	// The default map has walls at (4,4-6), (8,8-9), (9,8)
	// Use position (5,8) facing right - clear path until boundary wall at x=15
	r.PlayerX = 5.0
	r.PlayerY = 8.0
	r.PlayerA = 0 // Facing positive X

	ir := NewInteractionRaycaster(r, 10.0)

	// Place target directly in front of player (between player and boundary wall)
	ir.AddTarget(InteractionTarget{
		EntityID:         1,
		WorldX:           7.0, // 2 units in front
		WorldY:           8.0,
		Radius:           0.5,
		InteractionRange: 3.0,
	})

	result := ir.CastInteractionRay()
	if result == nil {
		// Debug: check why it failed
		rayDirX := math.Cos(r.PlayerA)
		rayDirY := math.Sin(r.PlayerA)
		wallDist := ir.castWallRay(rayDirX, rayDirY)
		targetDist := ir.distanceToTarget(&ir.Targets[0])
		intersects := ir.rayIntersectsTarget(rayDirX, rayDirY, &ir.Targets[0])
		t.Fatalf("expected to find target in front of player. wallDist=%f, targetDist=%f, intersects=%v", wallDist, targetDist, intersects)
	}
	if result.EntityID != 1 {
		t.Errorf("expected entity 1, got %d", result.EntityID)
	}
}

func TestCastInteractionRayTargetBehind(t *testing.T) {
	r := NewRenderer(320, 200)
	r.PlayerX = 5.0
	r.PlayerY = 8.0
	r.PlayerA = 0 // Facing positive X

	ir := NewInteractionRaycaster(r, 10.0)

	// Place target behind player
	ir.AddTarget(InteractionTarget{
		EntityID:         1,
		WorldX:           3.0, // Behind player
		WorldY:           8.0,
		Radius:           0.5,
		InteractionRange: 3.0,
	})

	result := ir.CastInteractionRay()
	if result != nil {
		t.Error("should not find target behind player")
	}
}

func TestCastInteractionRayTargetTooFar(t *testing.T) {
	r := NewRenderer(320, 200)
	r.PlayerX = 5.0
	r.PlayerY = 8.0
	r.PlayerA = 0

	ir := NewInteractionRaycaster(r, 5.0)

	// Place target beyond max range (but before wall at x=15)
	ir.AddTarget(InteractionTarget{
		EntityID:         1,
		WorldX:           12.0, // 7 units away, beyond max range of 5
		WorldY:           8.0,
		Radius:           0.5,
		InteractionRange: 10.0,
	})

	result := ir.CastInteractionRay()
	if result != nil {
		t.Error("should not find target beyond max range")
	}
}

func TestCastInteractionRayTargetOutsideInteractionRange(t *testing.T) {
	r := NewRenderer(320, 200)
	r.PlayerX = 5.0
	r.PlayerY = 8.0
	r.PlayerA = 0

	ir := NewInteractionRaycaster(r, 10.0)

	// Place target within view but outside its interaction range
	ir.AddTarget(InteractionTarget{
		EntityID:         1,
		WorldX:           9.0, // 4 units away
		WorldY:           8.0,
		Radius:           0.5,
		InteractionRange: 2.0, // Only interactable within 2 units
	})

	result := ir.CastInteractionRay()
	if result != nil {
		t.Error("should not find target outside its interaction range")
	}
}

func TestCastInteractionRayTargetOffCenter(t *testing.T) {
	r := NewRenderer(320, 200)
	r.PlayerX = 5.0
	r.PlayerY = 8.0
	r.PlayerA = 0

	ir := NewInteractionRaycaster(r, 10.0)

	// Place target to the side (not in ray path)
	ir.AddTarget(InteractionTarget{
		EntityID:         1,
		WorldX:           7.0,
		WorldY:           10.0, // 2 units to the side
		Radius:           0.5,
		InteractionRange: 5.0,
	})

	result := ir.CastInteractionRay()
	if result != nil {
		t.Error("should not find target that is off-center")
	}
}

func TestCastInteractionRaySelectsClosest(t *testing.T) {
	r := NewRenderer(320, 200)
	r.PlayerX = 5.0
	r.PlayerY = 8.0
	r.PlayerA = 0

	ir := NewInteractionRaycaster(r, 10.0)

	// Place two targets in front, different distances
	ir.AddTarget(InteractionTarget{
		EntityID:         1,
		WorldX:           9.0, // 4 units away (farther)
		WorldY:           8.0,
		Radius:           0.5,
		InteractionRange: 5.0,
	})
	ir.AddTarget(InteractionTarget{
		EntityID:         2,
		WorldX:           6.5, // 1.5 units away (closer)
		WorldY:           8.0,
		Radius:           0.5,
		InteractionRange: 5.0,
	})

	result := ir.CastInteractionRay()
	if result == nil {
		t.Fatal("expected to find a target")
	}
	if result.EntityID != 2 {
		t.Errorf("expected closest target (entity 2), got entity %d", result.EntityID)
	}
}

func TestRayIntersectsTarget(t *testing.T) {
	r := NewRenderer(320, 200)
	r.PlayerX = 0
	r.PlayerY = 0

	ir := NewInteractionRaycaster(r, 10.0)

	// Ray pointing right (1, 0)
	rayDirX := 1.0
	rayDirY := 0.0

	// Target directly on ray
	targetOnRay := &InteractionTarget{
		WorldX: 5.0,
		WorldY: 0.0,
		Radius: 0.5,
	}
	if !ir.rayIntersectsTarget(rayDirX, rayDirY, targetOnRay) {
		t.Error("should intersect target on ray")
	}

	// Target slightly off ray but within radius
	targetNearRay := &InteractionTarget{
		WorldX: 5.0,
		WorldY: 0.3, // Within 0.5 radius
		Radius: 0.5,
	}
	if !ir.rayIntersectsTarget(rayDirX, rayDirY, targetNearRay) {
		t.Error("should intersect target near ray within radius")
	}

	// Target too far from ray
	targetFarFromRay := &InteractionTarget{
		WorldX: 5.0,
		WorldY: 2.0, // Beyond 0.5 radius
		Radius: 0.5,
	}
	if ir.rayIntersectsTarget(rayDirX, rayDirY, targetFarFromRay) {
		t.Error("should not intersect target far from ray")
	}

	// Target behind player
	targetBehind := &InteractionTarget{
		WorldX: -5.0,
		WorldY: 0.0,
		Radius: 0.5,
	}
	if ir.rayIntersectsTarget(rayDirX, rayDirY, targetBehind) {
		t.Error("should not intersect target behind player")
	}
}

func TestDistanceToTarget(t *testing.T) {
	r := NewRenderer(320, 200)
	r.PlayerX = 0
	r.PlayerY = 0

	ir := NewInteractionRaycaster(r, 10.0)

	target := &InteractionTarget{
		WorldX: 3.0,
		WorldY: 4.0,
	}

	dist := ir.distanceToTarget(target)
	expected := 5.0 // 3-4-5 triangle
	if math.Abs(dist-expected) > 0.001 {
		t.Errorf("expected distance %f, got %f", expected, dist)
	}
}

func TestUpdateTargetDistances(t *testing.T) {
	r := NewRenderer(320, 200)
	r.PlayerX = 0
	r.PlayerY = 0

	ir := NewInteractionRaycaster(r, 10.0)

	ir.AddTarget(InteractionTarget{WorldX: 3.0, WorldY: 4.0})
	ir.AddTarget(InteractionTarget{WorldX: 6.0, WorldY: 8.0})

	ir.UpdateTargetDistances()

	if math.Abs(ir.Targets[0].Distance-5.0) > 0.001 {
		t.Errorf("expected distance 5.0 for target 0, got %f", ir.Targets[0].Distance)
	}
	if math.Abs(ir.Targets[1].Distance-10.0) > 0.001 {
		t.Errorf("expected distance 10.0 for target 1, got %f", ir.Targets[1].Distance)
	}
}

func TestGetTargetedObject(t *testing.T) {
	r := NewRenderer(320, 200)
	r.PlayerX = 5.0
	r.PlayerY = 8.0
	r.PlayerA = 0

	ir := NewInteractionRaycaster(r, 10.0)

	targets := []InteractionTarget{
		{EntityID: 1, WorldX: 7.0, WorldY: 8.0, Radius: 0.5, InteractionRange: 5.0},
	}

	result := ir.GetTargetedObject(targets)
	if result == nil {
		t.Fatal("expected to find target")
	}
	if result.EntityID != 1 {
		t.Errorf("expected entity 1, got %d", result.EntityID)
	}
}

func BenchmarkCastInteractionRay(b *testing.B) {
	r := NewRenderer(320, 200)
	r.PlayerX = 5.0
	r.PlayerY = 8.0
	r.PlayerA = 0

	ir := NewInteractionRaycaster(r, 10.0)

	// Add several targets in front of player (x > 5)
	for i := 0; i < 20; i++ {
		ir.AddTarget(InteractionTarget{
			EntityID:         uint64(i),
			WorldX:           6.0 + float64(i)*0.3,
			WorldY:           8.0,
			Radius:           0.5,
			InteractionRange: 5.0,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ir.CastInteractionRay()
	}
}
