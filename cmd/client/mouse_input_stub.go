//go:build noebiten

// Package main contains stub implementations for the Wyrm client testing.
// This file provides testable mouse input functions without Ebiten dependency.
package main

import (
	"math"

	"github.com/opd-ai/wyrm/config"
	"github.com/opd-ai/wyrm/pkg/rendering/raycast"
)

// MouseInputState holds the state for mouse input processing.
type MouseInputState struct {
	LastX          int
	LastY          int
	Initialized    bool
	SmoothedDeltaX float64
	SmoothedDeltaY float64
}

// applyMouseAcceleration applies acceleration to mouse deltas.
// Returns the accelerated delta values.
func applyMouseAcceleration(deltaX, deltaY, acceleration float64) (float64, float64) {
	magnitude := math.Sqrt(deltaX*deltaX + deltaY*deltaY)
	accelerationFactor := 1.0 + (magnitude * acceleration * 0.01)
	return deltaX * accelerationFactor, deltaY * accelerationFactor
}

// applyMouseSensitivity applies sensitivity scaling to mouse deltas.
func applyMouseSensitivity(deltaX, deltaY, sensitivity float64) (float64, float64) {
	scale := sensitivity * 0.005
	return deltaX * scale, deltaY * scale
}

// applyMouseSmoothing applies exponential smoothing to mouse deltas.
// state is modified in-place with the new smoothed values.
func applyMouseSmoothing(state *MouseInputState, deltaX, deltaY, factor float64) (float64, float64) {
	state.SmoothedDeltaX = state.SmoothedDeltaX*(1-factor) + deltaX*factor
	state.SmoothedDeltaY = state.SmoothedDeltaY*(1-factor) + deltaY*factor
	return state.SmoothedDeltaX, state.SmoothedDeltaY
}

// applyInvertY inverts the Y axis if enabled.
func applyInvertY(deltaY float64, invert bool) float64 {
	if invert {
		return -deltaY
	}
	return deltaY
}

// applyMouseModifiersTestable applies all mouse modifiers in sequence.
// This is a testable version of the Game.applyMouseModifiers method.
func applyMouseModifiersTestable(state *MouseInputState, cfg config.MouseConfig, deltaX, deltaY float64) (float64, float64) {
	if cfg.AccelerationOn {
		deltaX, deltaY = applyMouseAcceleration(deltaX, deltaY, cfg.Acceleration)
	}

	deltaX, deltaY = applyMouseSensitivity(deltaX, deltaY, cfg.Sensitivity)

	if cfg.SmoothingOn {
		deltaX, deltaY = applyMouseSmoothing(state, deltaX, deltaY, cfg.SmoothingFactor)
	}

	deltaY = applyInvertY(deltaY, cfg.InvertY)

	return deltaX, deltaY
}

// calculateMouseDeltaTestable computes the raw mouse movement delta.
// Returns (deltaX, deltaY, ok) where ok is false if not initialized or no movement.
func calculateMouseDeltaTestable(state *MouseInputState, cursorX, cursorY int) (float64, float64, bool) {
	if !state.Initialized {
		state.LastX = cursorX
		state.LastY = cursorY
		state.Initialized = true
		return 0, 0, false
	}

	deltaX := float64(cursorX - state.LastX)
	deltaY := float64(cursorY - state.LastY)
	state.LastX = cursorX
	state.LastY = cursorY

	if deltaX == 0 && deltaY == 0 {
		return 0, 0, false
	}
	return deltaX, deltaY, true
}

// normalizePlayerYaw normalizes the player angle to [-π, π] range.
func normalizePlayerYaw(angle float64) float64 {
	for angle > math.Pi {
		angle -= 2 * math.Pi
	}
	for angle < -math.Pi {
		angle += 2 * math.Pi
	}
	return angle
}

// clampPitch clamps the pitch angle to the maximum allowed range.
func clampPitch(pitch, maxPitch float64) float64 {
	if pitch > maxPitch {
		return maxPitch
	}
	if pitch < -maxPitch {
		return -maxPitch
	}
	return pitch
}

// updatePlayerYawTestable updates the player's horizontal angle and normalizes.
func updatePlayerYawTestable(currentAngle, deltaX float64) float64 {
	return normalizePlayerYaw(currentAngle + deltaX)
}

// updateRendererPitchTestable updates vertical look angle with clamping.
func updateRendererPitchTestable(currentPitch, deltaY float64) float64 {
	newPitch := currentPitch - deltaY
	return clampPitch(newPitch, raycast.MaxPitchAngle)
}
