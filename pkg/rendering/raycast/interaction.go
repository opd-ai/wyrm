// Package raycast provides the first-person raycasting renderer.
package raycast

import (
	"math"
)

// InteractionTarget represents an object that can be interacted with.
type InteractionTarget struct {
	// EntityID is the ECS entity identifier.
	EntityID uint64
	// WorldX is the object's world X position.
	WorldX float64
	// WorldY is the object's world Y position.
	WorldY float64
	// Radius is the object's interaction radius.
	Radius float64
	// Distance is the distance from the player to the object.
	Distance float64
	// InteractionRange is the maximum distance for interaction.
	InteractionRange float64
}

// InteractionRaycaster handles interaction targeting via raycasting.
type InteractionRaycaster struct {
	// Renderer reference for accessing map data and player position.
	renderer *Renderer
	// MaxRange is the maximum distance to check for interactable objects.
	MaxRange float64
	// Targets is the list of potential interaction targets.
	Targets []InteractionTarget
}

// NewInteractionRaycaster creates an interaction raycaster linked to a renderer.
func NewInteractionRaycaster(r *Renderer, maxRange float64) *InteractionRaycaster {
	if maxRange <= 0 {
		maxRange = 5.0 // Default max interaction range
	}
	return &InteractionRaycaster{
		renderer: r,
		MaxRange: maxRange,
		Targets:  make([]InteractionTarget, 0, 32),
	}
}

// SetTargets updates the list of interactable objects to check against.
func (ir *InteractionRaycaster) SetTargets(targets []InteractionTarget) {
	ir.Targets = targets
}

// AddTarget adds a potential interaction target.
func (ir *InteractionRaycaster) AddTarget(target InteractionTarget) {
	ir.Targets = append(ir.Targets, target)
}

// ClearTargets removes all targets.
func (ir *InteractionRaycaster) ClearTargets() {
	ir.Targets = ir.Targets[:0]
}

// CastInteractionRay casts a ray from the screen center and returns the closest
// interactable object within range, or nil if none found.
func (ir *InteractionRaycaster) CastInteractionRay() *InteractionTarget {
	if ir.renderer == nil || len(ir.Targets) == 0 {
		return nil
	}

	// Ray direction is the player's look direction (center of screen)
	rayDirX := math.Cos(ir.renderer.PlayerA)
	rayDirY := math.Sin(ir.renderer.PlayerA)

	// First, cast a ray to find the wall distance (objects behind walls are not targetable)
	wallDist := ir.castWallRay(rayDirX, rayDirY)

	// Check all targets for intersection with the center ray
	var closest *InteractionTarget
	closestDist := ir.MaxRange + 1

	for i := range ir.Targets {
		target := &ir.Targets[i]

		// Skip objects beyond wall
		targetDist := ir.distanceToTarget(target)
		if targetDist >= wallDist {
			continue
		}

		// Skip objects beyond max range
		if targetDist > ir.MaxRange {
			continue
		}

		// Check if ray passes within object's radius
		if ir.rayIntersectsTarget(rayDirX, rayDirY, target) {
			// Check if within interaction range
			if targetDist <= target.InteractionRange && targetDist < closestDist {
				closest = target
				closestDist = targetDist
			}
		}
	}

	return closest
}

// castWallRay casts a ray using DDA and returns the wall distance.
func (ir *InteractionRaycaster) castWallRay(rayDirX, rayDirY float64) float64 {
	mapX := int(ir.renderer.PlayerX)
	mapY := int(ir.renderer.PlayerY)

	deltaDistX, deltaDistY := calculateDeltaDist(rayDirX, rayDirY)
	sideDistX, sideDistY, stepX, stepY := calculateSideDist(
		ir.renderer.PlayerX, ir.renderer.PlayerY, mapX, mapY,
		rayDirX, rayDirY, deltaDistX, deltaDistY,
	)

	hit, side, sideDistX, sideDistY, _, _ := ir.renderer.performDDA(
		sideDistX, sideDistY, deltaDistX, deltaDistY,
		stepX, stepY, mapX, mapY,
	)

	if !hit {
		return MaxRayDistance
	}

	// Calculate perpendicular wall distance
	var perpWallDist float64
	if side == 0 {
		perpWallDist = sideDistX - deltaDistX
	} else {
		perpWallDist = sideDistY - deltaDistY
	}

	return perpWallDist
}

// distanceToTarget calculates the distance from the player to a target.
func (ir *InteractionRaycaster) distanceToTarget(target *InteractionTarget) float64 {
	dx := target.WorldX - ir.renderer.PlayerX
	dy := target.WorldY - ir.renderer.PlayerY
	return math.Sqrt(dx*dx + dy*dy)
}

// rayIntersectsTarget checks if the center ray passes within the target's radius.
func (ir *InteractionRaycaster) rayIntersectsTarget(rayDirX, rayDirY float64, target *InteractionTarget) bool {
	// Vector from player to target center
	dx := target.WorldX - ir.renderer.PlayerX
	dy := target.WorldY - ir.renderer.PlayerY

	// Project target position onto ray direction
	// dot = dx*rayDirX + dy*rayDirY (how far along the ray the target is)
	dot := dx*rayDirX + dy*rayDirY

	// Target is behind the player
	if dot < 0 {
		return false
	}

	// Calculate perpendicular distance from target to ray
	// perpX, perpY = position on ray closest to target
	perpX := dot * rayDirX
	perpY := dot * rayDirY

	// Distance from target to closest point on ray
	perpDist := math.Sqrt(
		(dx-perpX)*(dx-perpX) + (dy-perpY)*(dy-perpY),
	)

	// Ray intersects if perpendicular distance is within radius
	return perpDist <= target.Radius
}

// GetTargetedObject returns the currently targeted object from a list of targets.
// This is a convenience method that combines SetTargets + CastInteractionRay.
func (ir *InteractionRaycaster) GetTargetedObject(targets []InteractionTarget) *InteractionTarget {
	ir.SetTargets(targets)
	return ir.CastInteractionRay()
}

// UpdateTargetDistances recalculates distances for all targets.
func (ir *InteractionRaycaster) UpdateTargetDistances() {
	for i := range ir.Targets {
		ir.Targets[i].Distance = ir.distanceToTarget(&ir.Targets[i])
	}
}
