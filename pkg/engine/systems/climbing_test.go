package systems

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// testPlayerMarker is a simple component to mark player entities in tests.
type testPlayerMarker struct{}

func (t *testPlayerMarker) Type() string { return "Player" }

func TestClimbableSystem_NewClimbableSystem(t *testing.T) {
	s := NewClimbableSystem()

	if s == nil {
		t.Fatal("expected non-nil system")
	}
	if s.PlayerStepHeight != DefaultPlayerStepHeight {
		t.Errorf("expected step height %f, got %f", DefaultPlayerStepHeight, s.PlayerStepHeight)
	}
	if s.ClimbDuration != DefaultClimbDuration {
		t.Errorf("expected climb duration %f, got %f", DefaultClimbDuration, s.ClimbDuration)
	}
	if s.ActiveClimbs == nil {
		t.Error("expected non-nil ActiveClimbs map")
	}
}

func TestClimbableSystem_DetectsClimbableBarrier(t *testing.T) {
	w := ecs.NewWorld()
	s := NewClimbableSystem()

	// Create a player entity
	player := w.CreateEntity()
	w.AddComponent(player, &components.Position{X: 1.0, Y: 1.0, Z: 0.5})
	w.AddComponent(player, &testPlayerMarker{})

	// Create a climbable barrier near the player
	barrier := w.CreateEntity()
	w.AddComponent(barrier, &components.Position{X: 1.3, Y: 1.0, Z: 0.0})
	w.AddComponent(barrier, &components.Barrier{
		Shape: components.BarrierShape{
			ShapeType:   "box",
			Width:       0.5,
			Depth:       0.5,
			Height:      0.4,
			ClimbHeight: 0.5, // Climbable
		},
	})

	// Run update
	s.Update(w, 0.016)

	// Check if climb started
	if !s.IsClimbing(player) {
		t.Error("expected player to start climbing")
	}
}

func TestClimbableSystem_IgnoresNonClimbableBarrier(t *testing.T) {
	w := ecs.NewWorld()
	s := NewClimbableSystem()

	// Create a player entity
	player := w.CreateEntity()
	w.AddComponent(player, &components.Position{X: 1.0, Y: 1.0, Z: 0.5})
	w.AddComponent(player, &testPlayerMarker{})

	// Create a non-climbable barrier near the player
	barrier := w.CreateEntity()
	w.AddComponent(barrier, &components.Position{X: 1.3, Y: 1.0, Z: 0.0})
	w.AddComponent(barrier, &components.Barrier{
		Shape: components.BarrierShape{
			ShapeType:   "box",
			Width:       0.5,
			Depth:       0.5,
			Height:      2.0, // Tall barrier
			ClimbHeight: 0.0, // Not climbable
		},
	})

	// Run update
	s.Update(w, 0.016)

	// Check that climb did not start
	if s.IsClimbing(player) {
		t.Error("expected player not to climb non-climbable barrier")
	}
}

func TestClimbableSystem_IgnoresTooTallBarrier(t *testing.T) {
	w := ecs.NewWorld()
	s := NewClimbableSystem()

	// Create a player entity
	player := w.CreateEntity()
	w.AddComponent(player, &components.Position{X: 1.0, Y: 1.0, Z: 0.5})
	w.AddComponent(player, &testPlayerMarker{})

	// Create a barrier that exceeds its climb height
	barrier := w.CreateEntity()
	w.AddComponent(barrier, &components.Position{X: 1.3, Y: 1.0, Z: 0.0})
	w.AddComponent(barrier, &components.Barrier{
		Shape: components.BarrierShape{
			ShapeType:   "box",
			Width:       0.5,
			Depth:       0.5,
			Height:      1.0, // Taller than climb height
			ClimbHeight: 0.5, // Can only climb 0.5
		},
	})

	// Run update
	s.Update(w, 0.016)

	// Check that climb did not start
	if s.IsClimbing(player) {
		t.Error("expected player not to climb barrier taller than climb height")
	}
}

func TestClimbableSystem_ClimbAnimation(t *testing.T) {
	w := ecs.NewWorld()
	s := NewClimbableSystem()
	s.ClimbDuration = 0.1 // Fast climbing for test

	// Create a player entity
	player := w.CreateEntity()
	w.AddComponent(player, &components.Position{X: 1.0, Y: 1.0, Z: 0.5})
	w.AddComponent(player, &testPlayerMarker{})

	// Create a climbable barrier
	barrier := w.CreateEntity()
	w.AddComponent(barrier, &components.Position{X: 1.3, Y: 1.0, Z: 0.0})
	w.AddComponent(barrier, &components.Barrier{
		Shape: components.BarrierShape{
			ShapeType:   "box",
			Width:       0.5,
			Depth:       0.5,
			Height:      0.4,
			ClimbHeight: 0.5,
		},
	})

	// Start climb
	s.Update(w, 0.016)

	if !s.IsClimbing(player) {
		t.Fatal("expected climb to start")
	}

	initialProgress := s.GetClimbProgress(player)
	if initialProgress < 0 || initialProgress > 0.5 {
		t.Errorf("expected initial progress in ascending phase, got %f", initialProgress)
	}

	// Advance through ascending phase (duration 0.1s)
	for i := 0; i < 10; i++ {
		s.Update(w, 0.015)
	}

	// Get position
	posComp, _ := w.GetComponent(player, "Position")
	pos := posComp.(*components.Position)

	// Z should have increased (player climbing)
	if pos.Z <= 0.5 {
		t.Errorf("expected Z to increase during climb, got %f", pos.Z)
	}

	// Continue until climb completes (ascending + descending phases = 2 * 0.1s)
	// Total time needed: 0.2s, we've spent about 0.166s (1 * 0.016 + 10 * 0.015)
	// Need at least 5 more updates of 0.015 to complete
	for i := 0; i < 30; i++ {
		s.Update(w, 0.015)
	}

	// Climb should be complete
	if s.IsClimbing(player) {
		t.Error("expected climb to complete")
	}
}

func TestClimbableSystem_GetClimbProgress(t *testing.T) {
	s := NewClimbableSystem()

	// Non-climbing entity should return -1
	progress := s.GetClimbProgress(12345)
	if progress != -1 {
		t.Errorf("expected -1 for non-climbing entity, got %f", progress)
	}
}

func TestClimbableSystem_getClimbRange(t *testing.T) {
	s := NewClimbableSystem()

	tests := []struct {
		name     string
		barrier  *components.Barrier
		expected float64
	}{
		{
			name: "cylinder",
			barrier: &components.Barrier{
				Shape: components.BarrierShape{
					ShapeType: "cylinder",
					Radius:    0.5,
				},
			},
			expected: 0.8, // 0.5 + 0.3
		},
		{
			name: "box",
			barrier: &components.Barrier{
				Shape: components.BarrierShape{
					ShapeType: "box",
					Width:     1.0,
					Depth:     0.5,
				},
			},
			expected: 0.8, // max(1.0, 0.5)/2 + 0.3 = 0.5 + 0.3
		},
		{
			name: "default",
			barrier: &components.Barrier{
				Shape: components.BarrierShape{
					ShapeType: "billboard",
				},
			},
			expected: 0.5,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := s.getClimbRange(tc.barrier)
			if result != tc.expected {
				t.Errorf("expected %f, got %f", tc.expected, result)
			}
		})
	}
}

func TestLerp(t *testing.T) {
	tests := []struct {
		a, b, t  float64
		expected float64
	}{
		{0, 1, 0, 0},
		{0, 1, 1, 1},
		{0, 1, 0.5, 0.5},
		{10, 20, 0.25, 12.5},
	}

	for _, tc := range tests {
		result := lerp(tc.a, tc.b, tc.t)
		if result != tc.expected {
			t.Errorf("lerp(%f, %f, %f) = %f, expected %f", tc.a, tc.b, tc.t, result, tc.expected)
		}
	}
}

func TestSmoothStep(t *testing.T) {
	// Test edge cases
	if smoothStep(0) != 0 {
		t.Errorf("smoothStep(0) should be 0, got %f", smoothStep(0))
	}
	if smoothStep(1) != 1 {
		t.Errorf("smoothStep(1) should be 1, got %f", smoothStep(1))
	}

	// Test middle value (should be 0.5 for t=0.5 with smooth step)
	mid := smoothStep(0.5)
	if mid != 0.5 {
		t.Errorf("smoothStep(0.5) should be 0.5, got %f", mid)
	}

	// Test that it's monotonically increasing
	prev := 0.0
	for i := 0.0; i <= 1.0; i += 0.1 {
		val := smoothStep(i)
		if val < prev {
			t.Errorf("smoothStep should be monotonic, but %f < %f", val, prev)
		}
		prev = val
	}
}
