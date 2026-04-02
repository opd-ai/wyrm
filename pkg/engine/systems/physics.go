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
}

// NewPhysicsSystem creates a new physics system with default settings.
func NewPhysicsSystem() *PhysicsSystem {
	return &PhysicsSystem{
		Gravity:              9.8,
		MaxVelocity:          50.0,
		MinVelocityThreshold: 0.01,
		SwingGravityFactor:   2.0,
	}
}

// Update processes physics for all entities with PhysicsBody components.
func (s *PhysicsSystem) Update(w *ecs.World, dt float64) {
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
		} else if !phys.IsKinematic {
			s.updateLinear(phys, pos, dt)
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
