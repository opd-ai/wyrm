package systems

import (
	"math"

	"github.com/opd-ai/wyrm/pkg/engine/components"
)

// Collision helper functions shared between BarrierCollisionSystem and PhysicsSystem.
// These functions provide unified collision detection and resolution for various shapes.

// checkCylinderCollision tests circle-circle collision.
// Returns true if a circle at (x,y) with given radius overlaps with a cylinder barrier.
func checkCylinderCollision(x, y, radius float64, barrier *components.Barrier, barrierPos *components.Position) bool {
	dx := x - barrierPos.X
	dy := y - barrierPos.Y
	distSq := dx*dx + dy*dy
	combinedRadius := radius + barrier.Shape.Radius
	return distSq < combinedRadius*combinedRadius
}

// checkBoxCollisionAABB tests circle-AABB collision.
// Returns true if a circle at (x,y) with given radius overlaps with a box barrier.
func checkBoxCollisionAABB(x, y, radius float64, barrier *components.Barrier, barrierPos *components.Position) bool {
	halfW := barrier.Shape.Width / 2
	halfD := barrier.Shape.Depth / 2

	// Find closest point on box to circle center
	closestX := clampFloat64(x, barrierPos.X-halfW, barrierPos.X+halfW)
	closestY := clampFloat64(y, barrierPos.Y-halfD, barrierPos.Y+halfD)

	// Check if closest point is within circle radius
	dx := x - closestX
	dy := y - closestY
	distSq := dx*dx + dy*dy

	return distSq < radius*radius
}

// resolveCylinderCollisionPush resolves circle-circle collision by pushing out.
// Returns resolved position that is outside the barrier's combined radius.
func resolveCylinderCollisionPush(oldX, oldY, newX, newY, radius float64, barrier *components.Barrier, barrierPos *components.Position) (float64, float64) {
	dx := newX - barrierPos.X
	dy := newY - barrierPos.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist == 0 {
		// Entity at center of barrier, push to old position
		return oldX, oldY
	}

	combinedRadius := radius + barrier.Shape.Radius + 0.001
	resolvedX := barrierPos.X + (dx/dist)*combinedRadius
	resolvedY := barrierPos.Y + (dy/dist)*combinedRadius
	return resolvedX, resolvedY
}

// resolveBoxCollisionPush resolves circle-AABB collision by pushing out.
// Returns resolved position that is outside the box barrier.
func resolveBoxCollisionPush(oldX, oldY, newX, newY, radius float64, barrier *components.Barrier, barrierPos *components.Position) (float64, float64) {
	halfW := barrier.Shape.Width / 2
	halfD := barrier.Shape.Depth / 2

	// Find closest point on box to circle center
	closestX := clampFloat64(newX, barrierPos.X-halfW, barrierPos.X+halfW)
	closestY := clampFloat64(newY, barrierPos.Y-halfD, barrierPos.Y+halfD)

	// Direction from closest point to circle center
	dx := newX - closestX
	dy := newY - closestY
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist == 0 {
		// Entity inside box, push to old position
		return oldX, oldY
	}

	// Push entity away from closest point if overlapping
	pushDist := radius + 0.001 - dist
	if pushDist > 0 {
		return newX + (dx/dist)*pushDist, newY + (dy/dist)*pushDist
	}

	return newX, newY
}
