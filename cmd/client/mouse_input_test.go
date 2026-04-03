//go:build noebiten

// Package main contains tests for the Wyrm client mouse input system.
// These tests use the noebiten build tag to run without Ebiten initialization.
package main

import (
	"math"
	"testing"

	"github.com/opd-ai/wyrm/config"
	"github.com/opd-ai/wyrm/pkg/rendering/raycast"
)

// ============================================================
// Sensitivity Tests
// ============================================================

func TestApplyMouseSensitivity(t *testing.T) {
	tests := []struct {
		name        string
		deltaX      float64
		deltaY      float64
		sensitivity float64
		wantX       float64
		wantY       float64
	}{
		{"zero sensitivity", 10.0, 10.0, 0.0, 0.0, 0.0},
		{"default sensitivity", 10.0, 10.0, 1.0, 0.05, 0.05},
		{"high sensitivity", 10.0, 10.0, 2.0, 0.10, 0.10},
		{"low sensitivity", 10.0, 10.0, 0.5, 0.025, 0.025},
		{"negative delta", -10.0, -10.0, 1.0, -0.05, -0.05},
		{"zero delta", 0.0, 0.0, 1.0, 0.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotX, gotY := applyMouseSensitivity(tt.deltaX, tt.deltaY, tt.sensitivity)
			if math.Abs(gotX-tt.wantX) > 0.0001 {
				t.Errorf("X: got %f, want %f", gotX, tt.wantX)
			}
			if math.Abs(gotY-tt.wantY) > 0.0001 {
				t.Errorf("Y: got %f, want %f", gotY, tt.wantY)
			}
		})
	}
}

// ============================================================
// Acceleration Tests
// ============================================================

func TestApplyMouseAcceleration(t *testing.T) {
	tests := []struct {
		name         string
		deltaX       float64
		deltaY       float64
		acceleration float64
		wantIncrease bool
	}{
		{"zero acceleration", 10.0, 10.0, 0.0, false},
		{"positive acceleration increases", 10.0, 10.0, 1.0, true},
		{"higher acceleration increases more", 10.0, 10.0, 2.0, true},
		{"zero delta unchanged", 0.0, 0.0, 1.0, false},
		{"small movement gets smaller boost", 1.0, 1.0, 1.0, true},
		{"large movement gets larger boost", 50.0, 50.0, 1.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotX, gotY := applyMouseAcceleration(tt.deltaX, tt.deltaY, tt.acceleration)

			originalMag := math.Sqrt(tt.deltaX*tt.deltaX + tt.deltaY*tt.deltaY)
			resultMag := math.Sqrt(gotX*gotX + gotY*gotY)

			if tt.wantIncrease && resultMag <= originalMag && originalMag > 0 {
				t.Errorf("expected increase, but got magnitude %f <= %f", resultMag, originalMag)
			}
			if !tt.wantIncrease && math.Abs(resultMag-originalMag) > 0.0001 {
				t.Errorf("expected no change, but got magnitude %f != %f", resultMag, originalMag)
			}
		})
	}
}

func TestMouseAccelerationScalesWithSpeed(t *testing.T) {
	// Faster mouse movement should result in proportionally more acceleration
	acceleration := 1.0

	slowX, slowY := 5.0, 5.0
	fastX, fastY := 50.0, 50.0

	slowResultX, slowResultY := applyMouseAcceleration(slowX, slowY, acceleration)
	fastResultX, fastResultY := applyMouseAcceleration(fastX, fastY, acceleration)

	slowFactor := math.Sqrt(slowResultX*slowResultX+slowResultY*slowResultY) / math.Sqrt(slowX*slowX+slowY*slowY)
	fastFactor := math.Sqrt(fastResultX*fastResultX+fastResultY*fastResultY) / math.Sqrt(fastX*fastX+fastY*fastY)

	if fastFactor <= slowFactor {
		t.Errorf("faster movement should get higher acceleration factor: slow=%f, fast=%f", slowFactor, fastFactor)
	}
}

// ============================================================
// Smoothing Tests
// ============================================================

func TestApplyMouseSmoothing(t *testing.T) {
	// Test that smoothing interpolates between old and new values
	state := &MouseInputState{SmoothedDeltaX: 10.0, SmoothedDeltaY: 10.0}

	// With factor=0.5, result should be average of old and new
	newX, newY := 0.0, 0.0
	resultX, resultY := applyMouseSmoothing(state, newX, newY, 0.5)

	expectedX := 5.0 // (10 * 0.5) + (0 * 0.5)
	expectedY := 5.0

	if math.Abs(resultX-expectedX) > 0.0001 {
		t.Errorf("smoothed X: got %f, want %f", resultX, expectedX)
	}
	if math.Abs(resultY-expectedY) > 0.0001 {
		t.Errorf("smoothed Y: got %f, want %f", resultY, expectedY)
	}
}

func TestMouseSmoothingStateUpdates(t *testing.T) {
	state := &MouseInputState{SmoothedDeltaX: 0, SmoothedDeltaY: 0}

	// Apply smoothing multiple times and verify state converges
	for i := 0; i < 10; i++ {
		applyMouseSmoothing(state, 10.0, 10.0, 0.5)
	}

	// After many iterations, should approach the input value
	if math.Abs(state.SmoothedDeltaX-10.0) > 0.1 {
		t.Errorf("smoothed state should converge to input: got %f", state.SmoothedDeltaX)
	}
}

func TestMouseSmoothingFactorExtremes(t *testing.T) {
	// Factor=1.0 means no smoothing (instant response)
	state1 := &MouseInputState{SmoothedDeltaX: 0, SmoothedDeltaY: 0}
	resultX, _ := applyMouseSmoothing(state1, 10.0, 10.0, 1.0)
	if math.Abs(resultX-10.0) > 0.0001 {
		t.Errorf("factor=1.0 should give immediate response: got %f, want 10.0", resultX)
	}

	// Factor=0.0 means infinite smoothing (never changes)
	state2 := &MouseInputState{SmoothedDeltaX: 5.0, SmoothedDeltaY: 5.0}
	resultX2, _ := applyMouseSmoothing(state2, 10.0, 10.0, 0.0)
	if math.Abs(resultX2-5.0) > 0.0001 {
		t.Errorf("factor=0.0 should never change: got %f, want 5.0", resultX2)
	}
}

// ============================================================
// Pitch Clamping Tests
// ============================================================

func TestClampPitch(t *testing.T) {
	maxPitch := raycast.MaxPitchAngle

	tests := []struct {
		name  string
		pitch float64
		want  float64
	}{
		{"at max", maxPitch, maxPitch},
		{"above max", maxPitch + 0.5, maxPitch},
		{"way above max", maxPitch + 5.0, maxPitch},
		{"at min", -maxPitch, -maxPitch},
		{"below min", -maxPitch - 0.5, -maxPitch},
		{"way below min", -maxPitch - 5.0, -maxPitch},
		{"zero", 0.0, 0.0},
		{"mid positive", maxPitch / 2, maxPitch / 2},
		{"mid negative", -maxPitch / 2, -maxPitch / 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clampPitch(tt.pitch, maxPitch)
			if math.Abs(got-tt.want) > 0.0001 {
				t.Errorf("got %f, want %f", got, tt.want)
			}
		})
	}
}

func TestUpdateRendererPitchTestable(t *testing.T) {
	maxPitch := raycast.MaxPitchAngle

	tests := []struct {
		name         string
		currentPitch float64
		deltaY       float64
		want         float64
	}{
		{"zero delta", 0.0, 0.0, 0.0},
		{"look up from zero", 0.0, -0.5, 0.5},
		{"look down from zero", 0.0, 0.5, -0.5},
		{"clamp at max", maxPitch - 0.1, -0.5, maxPitch},
		{"clamp at min", -maxPitch + 0.1, 0.5, -maxPitch},
		{"already at max", maxPitch, -0.5, maxPitch},
		{"already at min", -maxPitch, 0.5, -maxPitch},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := updateRendererPitchTestable(tt.currentPitch, tt.deltaY)
			if math.Abs(got-tt.want) > 0.0001 {
				t.Errorf("got %f, want %f", got, tt.want)
			}
		})
	}
}

// ============================================================
// Yaw Normalization Tests
// ============================================================

func TestNormalizePlayerYaw(t *testing.T) {
	tests := []struct {
		name  string
		angle float64
	}{
		{"zero", 0.0},
		{"positive", math.Pi / 4},
		{"negative", -math.Pi / 4},
		{"at pi", math.Pi},
		{"at negative pi", -math.Pi},
		{"above pi", math.Pi + 0.5},
		{"below negative pi", -math.Pi - 0.5},
		{"multiple rotations positive", 5 * math.Pi},
		{"multiple rotations negative", -5 * math.Pi},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizePlayerYaw(tt.angle)

			// Result should be in [-π, π]
			if result > math.Pi || result < -math.Pi {
				t.Errorf("result %f is outside [-π, π] for input %f", result, tt.angle)
			}
		})
	}
}

func TestUpdatePlayerYawTestable(t *testing.T) {
	tests := []struct {
		name    string
		current float64
		deltaX  float64
	}{
		{"zero delta", 0.0, 0.0},
		{"turn right", 0.0, 0.5},
		{"turn left", 0.0, -0.5},
		{"wrap around positive", math.Pi - 0.1, 0.5},
		{"wrap around negative", -math.Pi + 0.1, -0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := updatePlayerYawTestable(tt.current, tt.deltaX)

			// Result should always be in [-π, π]
			if result > math.Pi || result < -math.Pi {
				t.Errorf("result %f is outside [-π, π]", result)
			}
		})
	}
}

// ============================================================
// Invert Y Tests
// ============================================================

func TestApplyInvertY(t *testing.T) {
	tests := []struct {
		name   string
		deltaY float64
		invert bool
		want   float64
	}{
		{"no invert positive", 5.0, false, 5.0},
		{"no invert negative", -5.0, false, -5.0},
		{"invert positive", 5.0, true, -5.0},
		{"invert negative", -5.0, true, 5.0},
		{"invert zero", 0.0, true, 0.0},
		{"no invert zero", 0.0, false, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applyInvertY(tt.deltaY, tt.invert)
			if got != tt.want {
				t.Errorf("got %f, want %f", got, tt.want)
			}
		})
	}
}

// ============================================================
// Delta Calculation Tests
// ============================================================

func TestCalculateMouseDeltaTestable(t *testing.T) {
	t.Run("first call initializes state", func(t *testing.T) {
		state := &MouseInputState{}
		_, _, ok := calculateMouseDeltaTestable(state, 100, 100)

		if ok {
			t.Error("first call should return ok=false")
		}
		if !state.Initialized {
			t.Error("state should be initialized")
		}
		if state.LastX != 100 || state.LastY != 100 {
			t.Errorf("state should have last position: got (%d, %d)", state.LastX, state.LastY)
		}
	})

	t.Run("second call returns delta", func(t *testing.T) {
		state := &MouseInputState{Initialized: true, LastX: 100, LastY: 100}
		dx, dy, ok := calculateMouseDeltaTestable(state, 110, 120)

		if !ok {
			t.Error("should return ok=true for movement")
		}
		if dx != 10.0 {
			t.Errorf("deltaX: got %f, want 10.0", dx)
		}
		if dy != 20.0 {
			t.Errorf("deltaY: got %f, want 20.0", dy)
		}
	})

	t.Run("no movement returns ok=false", func(t *testing.T) {
		state := &MouseInputState{Initialized: true, LastX: 100, LastY: 100}
		_, _, ok := calculateMouseDeltaTestable(state, 100, 100)

		if ok {
			t.Error("should return ok=false for no movement")
		}
	})

	t.Run("negative deltas work", func(t *testing.T) {
		state := &MouseInputState{Initialized: true, LastX: 100, LastY: 100}
		dx, dy, _ := calculateMouseDeltaTestable(state, 90, 80)

		if dx != -10.0 {
			t.Errorf("deltaX: got %f, want -10.0", dx)
		}
		if dy != -20.0 {
			t.Errorf("deltaY: got %f, want -20.0", dy)
		}
	})
}

// ============================================================
// Full Pipeline Integration Tests
// ============================================================

func TestApplyMouseModifiersTestable(t *testing.T) {
	t.Run("all modifiers off", func(t *testing.T) {
		state := &MouseInputState{}
		cfg := config.MouseConfig{
			Sensitivity:    1.0,
			AccelerationOn: false,
			SmoothingOn:    false,
			InvertY:        false,
		}

		outX, outY := applyMouseModifiersTestable(state, cfg, 10.0, 10.0)

		// With sensitivity=1.0, result should be input * 0.005
		expectedX := 10.0 * 0.005
		if math.Abs(outX-expectedX) > 0.0001 {
			t.Errorf("X: got %f, want %f", outX, expectedX)
		}
		if math.Abs(outY-expectedX) > 0.0001 {
			t.Errorf("Y: got %f, want %f", outY, expectedX)
		}
	})

	t.Run("with acceleration", func(t *testing.T) {
		state := &MouseInputState{}
		cfgNoAccel := config.MouseConfig{Sensitivity: 1.0, AccelerationOn: false}
		cfgAccel := config.MouseConfig{Sensitivity: 1.0, AccelerationOn: true, Acceleration: 1.0}

		noAccelX, _ := applyMouseModifiersTestable(state, cfgNoAccel, 10.0, 10.0)
		accelX, _ := applyMouseModifiersTestable(state, cfgAccel, 10.0, 10.0)

		if accelX <= noAccelX {
			t.Errorf("acceleration should increase output: noAccel=%f, accel=%f", noAccelX, accelX)
		}
	})

	t.Run("with invert Y", func(t *testing.T) {
		state := &MouseInputState{}
		cfg := config.MouseConfig{Sensitivity: 1.0, InvertY: true}

		_, outY := applyMouseModifiersTestable(state, cfg, 0.0, 10.0)

		// Output Y should be negative (inverted)
		if outY >= 0 {
			t.Errorf("inverted Y should be negative: got %f", outY)
		}
	})

	t.Run("with smoothing", func(t *testing.T) {
		state := &MouseInputState{SmoothedDeltaX: 0, SmoothedDeltaY: 0}
		cfg := config.MouseConfig{Sensitivity: 1.0, SmoothingOn: true, SmoothingFactor: 0.5}

		// First call with movement
		out1X, _ := applyMouseModifiersTestable(state, cfg, 10.0, 10.0)

		// Second call with same movement should be different due to smoothing
		out2X, _ := applyMouseModifiersTestable(state, cfg, 10.0, 10.0)

		// Second call should be larger due to smoothing accumulation
		if out2X <= out1X {
			t.Errorf("smoothing should accumulate: first=%f, second=%f", out1X, out2X)
		}
	})
}

// ============================================================
// Benchmark Tests
// ============================================================

func BenchmarkApplyMouseSensitivity(b *testing.B) {
	for i := 0; i < b.N; i++ {
		applyMouseSensitivity(10.0, 10.0, 1.0)
	}
}

func BenchmarkApplyMouseAcceleration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		applyMouseAcceleration(10.0, 10.0, 1.0)
	}
}

func BenchmarkApplyMouseSmoothing(b *testing.B) {
	state := &MouseInputState{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		applyMouseSmoothing(state, 10.0, 10.0, 0.5)
	}
}

func BenchmarkNormalizePlayerYaw(b *testing.B) {
	for i := 0; i < b.N; i++ {
		normalizePlayerYaw(float64(i) * 0.1)
	}
}

func BenchmarkClampPitch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		clampPitch(float64(i)*0.01, raycast.MaxPitchAngle)
	}
}

func BenchmarkFullMousePipeline(b *testing.B) {
	state := &MouseInputState{Initialized: true, LastX: 0, LastY: 0}
	cfg := config.MouseConfig{
		Sensitivity:     1.0,
		AccelerationOn:  true,
		Acceleration:    1.0,
		SmoothingOn:     true,
		SmoothingFactor: 0.5,
		InvertY:         false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dx, dy, ok := calculateMouseDeltaTestable(state, i%1000, i%1000)
		if ok {
			applyMouseModifiersTestable(state, cfg, dx, dy)
		}
	}
}
