package systems

import (
	"math"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// PhysicsSystem handles physics simulation for entities with PhysicsBody components.
// It processes linear velocity, friction, gravity, and swinging motion.
type PhysicsSystem struct {
	// Gravity is the downward acceleration in units/second² (positive = down).
	Gravity float64
	// MaxVelocity clamps velocity to prevent runaway speeds.
	MaxVelocity float64
	// MinVelocityThreshold velocities below this are set to zero (prevents drift).
	MinVelocityThreshold float64
	// SwingGravityFactor affects how strongly gravity influences swinging objects.
	SwingGravityFactor float64
	// MaxPushSpeed limits how fast objects can be pushed to prevent wall phasing.
	MaxPushSpeed float64
	// EnableBarrierCollision enables collision checking against barriers.
	EnableBarrierCollision bool
}

// NewPhysicsSystem creates a new physics system with default settings.
func NewPhysicsSystem() *PhysicsSystem {
	return &PhysicsSystem{
		Gravity:                9.8,
		MaxVelocity:            50.0,
		MinVelocityThreshold:   0.01,
		SwingGravityFactor:     2.0,
		MaxPushSpeed:           5.0,  // Limits push velocity to prevent wall phasing
		EnableBarrierCollision: true, // Enable collision checking against barriers
	}
}

// Update processes physics for all entities with PhysicsBody components.
func (s *PhysicsSystem) Update(w *ecs.World, dt float64) {
	// Pre-fetch barriers once for all physics entities
	var barriers []ecs.Entity
	if s.EnableBarrierCollision {
		barriers = w.Entities("Barrier", "Position")
	}

	for _, e := range w.Entities("PhysicsBody", "Position") {
		physComp, ok := w.GetComponent(e, "PhysicsBody")
		if !ok {
			continue
		}
		phys := physComp.(*components.PhysicsBody)

		posComp, ok := w.GetComponent(e, "Position")
		if !ok {
			continue
		}
		pos := posComp.(*components.Position)

		if phys.IsSwinging {
			s.updateSwinging(phys, pos, dt)
			// Update door collision polygon for swinging entities
			s.updateDoorCollision(w, e, phys, pos)
		} else if !phys.IsKinematic {
			s.updateLinearWithCollision(w, e, phys, pos, barriers, dt)
		}
	}
}

// updateLinear handles linear physics (velocity, friction, gravity).
func (s *PhysicsSystem) updateLinear(phys *components.PhysicsBody, pos *components.Position, dt float64) {
	// Apply gravity if not grounded
	if !phys.Grounded && phys.Mass > 0 {
		phys.VelocityZ -= s.Gravity * dt
	}

	// Apply friction to horizontal movement
	if phys.Grounded && phys.Friction > 0 {
		frictionFactor := 1.0 - phys.Friction*dt*10 // Scale friction for game feel
		if frictionFactor < 0 {
			frictionFactor = 0
		}
		phys.VelocityX *= frictionFactor
		phys.VelocityY *= frictionFactor
	}

	// Apply air resistance to all velocities when not grounded
	if !phys.Grounded {
		airResistance := 0.98
		phys.VelocityX *= airResistance
		phys.VelocityY *= airResistance
	}

	// Clamp velocities
	phys.VelocityX = clampVelocity(phys.VelocityX, s.MaxVelocity)
	phys.VelocityY = clampVelocity(phys.VelocityY, s.MaxVelocity)
	phys.VelocityZ = clampVelocity(phys.VelocityZ, s.MaxVelocity)

	// Zero out small velocities to prevent drift
	if math.Abs(phys.VelocityX) < s.MinVelocityThreshold {
		phys.VelocityX = 0
	}
	if math.Abs(phys.VelocityY) < s.MinVelocityThreshold {
		phys.VelocityY = 0
	}
	if math.Abs(phys.VelocityZ) < s.MinVelocityThreshold {
		phys.VelocityZ = 0
	}

	// Update position
	pos.X += phys.VelocityX * dt
	pos.Y += phys.VelocityY * dt
	pos.Z += phys.VelocityZ * dt

	// Ground collision (simple floor at Z=0)
	if pos.Z < 0 {
		pos.Z = 0
		// Bounce or stop
		if phys.Bounciness > 0 && math.Abs(phys.VelocityZ) > 0.5 {
			phys.VelocityZ = -phys.VelocityZ * phys.Bounciness
		} else {
			phys.VelocityZ = 0
			phys.Grounded = true
		}
	}
}

// updateLinearWithCollision handles linear physics with barrier collision checking.
// This is the enhanced version that checks for wall/barrier collisions for pushable objects.
func (s *PhysicsSystem) updateLinearWithCollision(w *ecs.World, entity ecs.Entity, phys *components.PhysicsBody, pos *components.Position, barriers []ecs.Entity, dt float64) {
	// Apply gravity if not grounded
	if !phys.Grounded && phys.Mass > 0 {
		phys.VelocityZ -= s.Gravity * dt
	}

	// Apply friction to horizontal movement
	if phys.Grounded && phys.Friction > 0 {
		frictionFactor := 1.0 - phys.Friction*dt*10
		if frictionFactor < 0 {
			frictionFactor = 0
		}
		phys.VelocityX *= frictionFactor
		phys.VelocityY *= frictionFactor
	}

	// Apply air resistance to all velocities when not grounded
	if !phys.Grounded {
		airResistance := 0.98
		phys.VelocityX *= airResistance
		phys.VelocityY *= airResistance
	}

	// Limit push speed to prevent wall phasing
	if phys.IsPushable && s.MaxPushSpeed > 0 {
		phys.VelocityX = clampVelocity(phys.VelocityX, s.MaxPushSpeed)
		phys.VelocityY = clampVelocity(phys.VelocityY, s.MaxPushSpeed)
	}

	// Clamp velocities to max
	phys.VelocityX = clampVelocity(phys.VelocityX, s.MaxVelocity)
	phys.VelocityY = clampVelocity(phys.VelocityY, s.MaxVelocity)
	phys.VelocityZ = clampVelocity(phys.VelocityZ, s.MaxVelocity)

	// Zero out small velocities to prevent drift
	if math.Abs(phys.VelocityX) < s.MinVelocityThreshold {
		phys.VelocityX = 0
	}
	if math.Abs(phys.VelocityY) < s.MinVelocityThreshold {
		phys.VelocityY = 0
	}
	if math.Abs(phys.VelocityZ) < s.MinVelocityThreshold {
		phys.VelocityZ = 0
	}

	// Calculate new position
	oldX, oldY := pos.X, pos.Y
	newX := pos.X + phys.VelocityX*dt
	newY := pos.Y + phys.VelocityY*dt
	newZ := pos.Z + phys.VelocityZ*dt

	// Check barrier collisions for pushable objects
	if phys.IsPushable && len(barriers) > 0 {
		newX, newY = s.resolveBarrierCollisions(w, entity, oldX, oldY, newX, newY, phys.CollisionRadius, barriers)
	}

	// Update position
	pos.X = newX
	pos.Y = newY
	pos.Z = newZ

	// Ground collision (simple floor at Z=0)
	if pos.Z < 0 {
		pos.Z = 0
		if phys.Bounciness > 0 && math.Abs(phys.VelocityZ) > 0.5 {
			phys.VelocityZ = -phys.VelocityZ * phys.Bounciness
		} else {
			phys.VelocityZ = 0
			phys.Grounded = true
		}
	}
}

// resolveBarrierCollisions checks for collisions with barriers and resolves them.
// Returns the resolved position after collision resolution.
func (s *PhysicsSystem) resolveBarrierCollisions(w *ecs.World, entity ecs.Entity, oldX, oldY, newX, newY, radius float64, barriers []ecs.Entity) (float64, float64) {
	if radius <= 0 {
		radius = 0.5 // Default collision radius
	}

	resolvedX, resolvedY := newX, newY

	for _, be := range barriers {
		if be == entity {
			continue // Don't collide with self
		}

		barrierComp, bOK := w.GetComponent(be, "Barrier")
		barrierPosComp, bpOK := w.GetComponent(be, "Position")
		if !bOK || !bpOK {
			continue
		}

		barrier := barrierComp.(*components.Barrier)
		barrierPos := barrierPosComp.(*components.Position)

		if barrier.IsDestroyed() {
			continue
		}

		// Check collision and resolve
		if checkBarrierCollision(resolvedX, resolvedY, radius, barrier, barrierPos) {
			resolvedX, resolvedY = resolveBarrierCollision(oldX, oldY, resolvedX, resolvedY, radius, barrier, barrierPos)
		}
	}

	return resolvedX, resolvedY
}

// checkBarrierCollision tests if a point with radius collides with a barrier.
func checkBarrierCollision(x, y, radius float64, barrier *components.Barrier, barrierPos *components.Position) bool {
	switch barrier.Shape.ShapeType {
	case "cylinder":
		dx := x - barrierPos.X
		dy := y - barrierPos.Y
		distSq := dx*dx + dy*dy
		combinedRadius := radius + barrier.Shape.Radius
		return distSq < combinedRadius*combinedRadius

	case "box":
		halfW := barrier.Shape.Width / 2
		halfD := barrier.Shape.Depth / 2
		closestX := physicsClamp(x, barrierPos.X-halfW, barrierPos.X+halfW)
		closestY := physicsClamp(y, barrierPos.Y-halfD, barrierPos.Y+halfD)
		dx := x - closestX
		dy := y - closestY
		distSq := dx*dx + dy*dy
		return distSq < radius*radius

	default:
		// Default to cylinder
		dx := x - barrierPos.X
		dy := y - barrierPos.Y
		distSq := dx*dx + dy*dy
		combinedRadius := radius + barrier.Shape.Radius
		return distSq < combinedRadius*combinedRadius
	}
}

// resolveBarrierCollision resolves collision by pushing the entity out of the barrier.
func resolveBarrierCollision(oldX, oldY, newX, newY, radius float64, barrier *components.Barrier, barrierPos *components.Position) (float64, float64) {
	switch barrier.Shape.ShapeType {
	case "cylinder":
		dx := newX - barrierPos.X
		dy := newY - barrierPos.Y
		dist := math.Sqrt(dx*dx + dy*dy)

		if dist == 0 {
			return oldX, oldY
		}

		combinedRadius := radius + barrier.Shape.Radius + 0.001
		resolvedX := barrierPos.X + (dx/dist)*combinedRadius
		resolvedY := barrierPos.Y + (dy/dist)*combinedRadius
		return resolvedX, resolvedY

	case "box":
		halfW := barrier.Shape.Width / 2
		halfD := barrier.Shape.Depth / 2
		closestX := physicsClamp(newX, barrierPos.X-halfW, barrierPos.X+halfW)
		closestY := physicsClamp(newY, barrierPos.Y-halfD, barrierPos.Y+halfD)

		dx := newX - closestX
		dy := newY - closestY
		dist := math.Sqrt(dx*dx + dy*dy)

		if dist == 0 {
			return oldX, oldY
		}

		pushDist := radius + 0.001 - dist
		if pushDist > 0 {
			return newX + (dx/dist)*pushDist, newY + (dy/dist)*pushDist
		}
		return newX, newY

	default:
		return resolveBarrierCollision(oldX, oldY, newX, newY, radius, &components.Barrier{
			Shape: components.BarrierShape{ShapeType: "cylinder", Radius: barrier.Shape.Radius},
		}, barrierPos)
	}
}

// physicsClamp limits a value to [min, max] range.
func physicsClamp(v, minV, maxV float64) float64 {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}

// updateSwinging handles pendulum-like swinging motion.
func (s *PhysicsSystem) updateSwinging(phys *components.PhysicsBody, pos *components.Position, dt float64) {
	// Apply gravity torque (proportional to sin of angle)
	gravityTorque := -s.SwingGravityFactor * math.Sin(phys.SwingAngle)
	phys.SwingVelocity += gravityTorque * dt

	// Apply damping
	phys.SwingVelocity *= 1.0 - phys.SwingDamping*dt

	// Update angle
	phys.SwingAngle += phys.SwingVelocity * dt

	// Clamp to max swing angle
	if phys.MaxSwingAngle > 0 {
		if phys.SwingAngle > phys.MaxSwingAngle {
			phys.SwingAngle = phys.MaxSwingAngle
			phys.SwingVelocity = -phys.SwingVelocity * 0.5 // Bounce back
		} else if phys.SwingAngle < -phys.MaxSwingAngle {
			phys.SwingAngle = -phys.MaxSwingAngle
			phys.SwingVelocity = -phys.SwingVelocity * 0.5 // Bounce back
		}
	}

	// Zero out small swing velocities
	if math.Abs(phys.SwingVelocity) < 0.01 && math.Abs(phys.SwingAngle) < 0.05 {
		phys.SwingVelocity = 0
		phys.SwingAngle = 0
	}

	// Update the entity's angle to match swing
	pos.Angle = phys.SwingAngle
}

// updateDoorCollision updates the barrier collision shape for a swinging door.
// For doors with a Barrier component, the collision box rotates around the pivot.
func (s *PhysicsSystem) updateDoorCollision(w *ecs.World, entity ecs.Entity, phys *components.PhysicsBody, pos *components.Position) {
	barrierComp, hasBarrier := w.GetComponent(entity, "Barrier")
	if !hasBarrier {
		return
	}
	barrier := barrierComp.(*components.Barrier)

	// Only update polygon barriers (doors should use polygon collision)
	if barrier.Shape.ShapeType != "box" && barrier.Shape.ShapeType != "polygon" {
		return
	}

	// For box shapes, we convert to a rotated polygon representing the door
	// Door dimensions: Width is the door width, Depth is the door thickness
	halfW := barrier.Shape.Width / 2
	halfD := barrier.Shape.Depth / 2
	if halfW == 0 {
		halfW = 0.5 // Default door half-width
	}
	if halfD == 0 {
		halfD = 0.05 // Default door half-thickness
	}

	// Calculate the 4 corners of the door relative to pivot
	// The pivot is at one edge of the door (hinge side)
	// Door extends from pivot along the positive X axis in local space
	pivotX := phys.PivotOffsetX
	pivotY := phys.PivotOffsetY

	// Door corners in local space (before rotation)
	// Assuming hinge is at the left edge, door extends rightward
	doorLength := halfW * 2
	corners := [][2]float64{
		{pivotX, pivotY - halfD},              // Near left (hinge side)
		{pivotX + doorLength, pivotY - halfD}, // Near right
		{pivotX + doorLength, pivotY + halfD}, // Far right
		{pivotX, pivotY + halfD},              // Far left (hinge side)
	}

	// Rotate corners by swing angle around the pivot point
	sinA := math.Sin(phys.SwingAngle)
	cosA := math.Cos(phys.SwingAngle)

	rotatedVerts := make([]float64, 8)
	for i, corner := range corners {
		// Translate to pivot, rotate, then translate back
		localX := corner[0] - pivotX
		localY := corner[1] - pivotY
		rotatedX := localX*cosA - localY*sinA + pivotX
		rotatedY := localX*sinA + localY*cosA + pivotY
		rotatedVerts[i*2] = rotatedX
		rotatedVerts[i*2+1] = rotatedY
	}

	// Update the barrier's collision polygon
	barrier.Shape.ShapeType = "polygon"
	barrier.Shape.Vertices = rotatedVerts
}

// clampVelocity limits velocity magnitude.
func clampVelocity(v, max float64) float64 {
	if v > max {
		return max
	}
	if v < -max {
		return -max
	}
	return v
}

// PushEntity applies a push force to an entity if it has a pushable physics body.
// Returns true if the push was applied.
func (s *PhysicsSystem) PushEntity(w *ecs.World, entity ecs.Entity, forceX, forceY float64) bool {
	physComp, ok := w.GetComponent(entity, "PhysicsBody")
	if !ok {
		return false
	}
	phys := physComp.(*components.PhysicsBody)

	if !phys.IsPushable {
		return false
	}

	phys.ApplyImpulse(forceX, forceY, 0)
	return true
}

// SwingEntity applies an angular impulse to a swinging entity.
// Returns true if the swing was applied.
func (s *PhysicsSystem) SwingEntity(w *ecs.World, entity ecs.Entity, angularImpulse float64) bool {
	physComp, ok := w.GetComponent(entity, "PhysicsBody")
	if !ok {
		return false
	}
	phys := physComp.(*components.PhysicsBody)

	if !phys.IsSwinging {
		return false
	}

	phys.ApplySwingImpulse(angularImpulse)
	return true
}

// IsEntityMoving returns true if the entity's physics body has significant velocity.
func (s *PhysicsSystem) IsEntityMoving(w *ecs.World, entity ecs.Entity) bool {
	physComp, ok := w.GetComponent(entity, "PhysicsBody")
	if !ok {
		return false
	}
	phys := physComp.(*components.PhysicsBody)

	threshold := s.MinVelocityThreshold * 10
	return math.Abs(phys.VelocityX) > threshold ||
		math.Abs(phys.VelocityY) > threshold ||
		math.Abs(phys.VelocityZ) > threshold ||
		math.Abs(phys.SwingVelocity) > threshold
}

// GetEntityVelocity returns the current velocity of an entity.
func (s *PhysicsSystem) GetEntityVelocity(w *ecs.World, entity ecs.Entity) (vx, vy, vz float64, ok bool) {
	physComp, hasPhys := w.GetComponent(entity, "PhysicsBody")
	if !hasPhys {
		return 0, 0, 0, false
	}
	phys := physComp.(*components.PhysicsBody)
	return phys.VelocityX, phys.VelocityY, phys.VelocityZ, true
}

// SetEntityVelocity directly sets an entity's velocity.
func (s *PhysicsSystem) SetEntityVelocity(w *ecs.World, entity ecs.Entity, vx, vy, vz float64) bool {
	physComp, ok := w.GetComponent(entity, "PhysicsBody")
	if !ok {
		return false
	}
	phys := physComp.(*components.PhysicsBody)

	if phys.IsKinematic {
		return false
	}

	phys.VelocityX = vx
	phys.VelocityY = vy
	phys.VelocityZ = vz
	return true
}

// StopEntity immediately stops all movement on an entity.
func (s *PhysicsSystem) StopEntity(w *ecs.World, entity ecs.Entity) {
	physComp, ok := w.GetComponent(entity, "PhysicsBody")
	if !ok {
		return
	}
	phys := physComp.(*components.PhysicsBody)

	phys.VelocityX = 0
	phys.VelocityY = 0
	phys.VelocityZ = 0
	phys.SwingVelocity = 0
}
