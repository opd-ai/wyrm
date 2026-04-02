package systems

import (
	"math"
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// TestNewPhysicsSystem verifies creation of a new physics system.
func TestNewPhysicsSystem(t *testing.T) {
	sys := NewPhysicsSystem()

	if sys == nil {
		t.Fatal("NewPhysicsSystem returned nil")
	}

	if sys.Gravity <= 0 {
		t.Errorf("Gravity = %f, expected positive value", sys.Gravity)
	}

	if sys.MaxVelocity <= 0 {
		t.Errorf("MaxVelocity = %f, expected positive value", sys.MaxVelocity)
	}
}

// TestPhysicsSystem_Update verifies the Update method runs without error.
func TestPhysicsSystem_Update(t *testing.T) {
	world := ecs.NewWorld()
	sys := NewPhysicsSystem()

	// Create an entity with physics
	e := world.CreateEntity()
	pos := &components.Position{X: 0, Y: 0, Z: 0}
	phys := components.NewPushableBody(1.0, 0.5)
	world.AddComponent(e, pos)
	world.AddComponent(e, phys)

	// Update should not panic
	sys.Update(world, 0.016)
}

// TestPhysicsSystem_LinearMovement verifies linear velocity affects position.
func TestPhysicsSystem_LinearMovement(t *testing.T) {
	world := ecs.NewWorld()
	sys := NewPhysicsSystem()

	e := world.CreateEntity()
	pos := &components.Position{X: 0, Y: 0, Z: 0}
	phys := &components.PhysicsBody{
		VelocityX: 10.0, // 10 units/second in X
		VelocityY: 5.0,  // 5 units/second in Y
		Mass:      1.0,
		Friction:  0.0,
		Grounded:  true,
	}
	world.AddComponent(e, pos)
	world.AddComponent(e, phys)

	// Simulate 1 second
	dt := 0.016
	for i := 0; i < 62; i++ { // ~1 second
		sys.Update(world, dt)
	}

	// Position should have moved approximately 10 units in X
	if pos.X < 9.0 || pos.X > 11.0 {
		t.Errorf("pos.X = %f, expected ~10.0", pos.X)
	}
	if pos.Y < 4.0 || pos.Y > 6.0 {
		t.Errorf("pos.Y = %f, expected ~5.0", pos.Y)
	}
}

// TestPhysicsSystem_Friction verifies friction slows down objects.
func TestPhysicsSystem_Friction(t *testing.T) {
	world := ecs.NewWorld()
	sys := NewPhysicsSystem()

	e := world.CreateEntity()
	pos := &components.Position{X: 0, Y: 0, Z: 0}
	phys := &components.PhysicsBody{
		VelocityX: 10.0,
		VelocityY: 0,
		Mass:      1.0,
		Friction:  0.8, // High friction
		Grounded:  true,
	}
	world.AddComponent(e, pos)
	world.AddComponent(e, phys)

	initialVelocity := phys.VelocityX

	// Simulate some time
	for i := 0; i < 30; i++ {
		sys.Update(world, 0.016)
	}

	// Velocity should have decreased due to friction
	if phys.VelocityX >= initialVelocity {
		t.Errorf("Velocity should decrease with friction, got %f (initial: %f)", phys.VelocityX, initialVelocity)
	}
}

// TestPhysicsSystem_Gravity verifies gravity affects falling objects.
func TestPhysicsSystem_Gravity(t *testing.T) {
	world := ecs.NewWorld()
	sys := NewPhysicsSystem()

	e := world.CreateEntity()
	pos := &components.Position{X: 0, Y: 0, Z: 10.0} // Start above ground
	phys := &components.PhysicsBody{
		Mass:     1.0,
		Friction: 0.5,
		Grounded: false, // Not on ground, should fall
	}
	world.AddComponent(e, pos)
	world.AddComponent(e, phys)

	// Simulate some time
	for i := 0; i < 60; i++ {
		sys.Update(world, 0.016)
	}

	// Should have fallen (Z decreased) and gained downward velocity
	if phys.VelocityZ >= 0 {
		t.Errorf("VelocityZ = %f, expected negative (falling)", phys.VelocityZ)
	}
	if pos.Z >= 10.0 {
		t.Errorf("pos.Z = %f, expected to have fallen below 10.0", pos.Z)
	}
}

// TestPhysicsSystem_GroundCollision verifies objects stop at ground level.
func TestPhysicsSystem_GroundCollision(t *testing.T) {
	world := ecs.NewWorld()
	sys := NewPhysicsSystem()

	e := world.CreateEntity()
	pos := &components.Position{X: 0, Y: 0, Z: 5.0}
	phys := &components.PhysicsBody{
		VelocityZ:  -20.0, // Falling fast
		Mass:       1.0,
		Bounciness: 0.0, // No bounce
		Grounded:   false,
	}
	world.AddComponent(e, pos)
	world.AddComponent(e, phys)

	// Simulate until it hits ground
	for i := 0; i < 100; i++ {
		sys.Update(world, 0.016)
	}

	// Should be at ground level
	if pos.Z < 0 {
		t.Errorf("pos.Z = %f, expected >= 0 (at ground)", pos.Z)
	}
	if !phys.Grounded {
		t.Error("Object should be grounded after landing")
	}
}

// TestPhysicsSystem_Bounce verifies bouncing objects.
func TestPhysicsSystem_Bounce(t *testing.T) {
	world := ecs.NewWorld()
	sys := NewPhysicsSystem()

	e := world.CreateEntity()
	pos := &components.Position{X: 0, Y: 0, Z: 1.0}
	phys := &components.PhysicsBody{
		VelocityZ:  -10.0, // Falling
		Mass:       1.0,
		Bounciness: 0.8, // Should bounce
		Grounded:   false,
	}
	world.AddComponent(e, pos)
	world.AddComponent(e, phys)

	// Simulate until it bounces
	for i := 0; i < 20; i++ {
		sys.Update(world, 0.016)
	}

	// After hitting ground, velocity should reverse (positive Z = going up)
	if phys.VelocityZ < 0 && pos.Z <= 0.1 {
		// If at ground and still negative velocity, bounce didn't happen
		t.Errorf("Expected bounce, but VelocityZ = %f at Z = %f", phys.VelocityZ, pos.Z)
	}
}

// TestPhysicsSystem_Swinging verifies swinging motion.
func TestPhysicsSystem_Swinging(t *testing.T) {
	world := ecs.NewWorld()
	sys := NewPhysicsSystem()

	e := world.CreateEntity()
	pos := &components.Position{X: 0, Y: 0, Z: 0, Angle: 0}
	phys := components.NewSwingingBody(math.Pi/4, 0.1) // Max 45 degrees
	phys.SwingAngle = 0.3                              // Start offset
	world.AddComponent(e, pos)
	world.AddComponent(e, phys)

	initialAngle := phys.SwingAngle

	// Simulate for some time
	for i := 0; i < 30; i++ {
		sys.Update(world, 0.016)
	}

	// Angle should have changed (swinging back toward center)
	if math.Abs(phys.SwingAngle-initialAngle) < 0.01 {
		t.Error("Swing angle should change over time")
	}

	// Position angle should match swing angle
	if math.Abs(pos.Angle-phys.SwingAngle) > 0.001 {
		t.Errorf("pos.Angle (%f) should match SwingAngle (%f)", pos.Angle, phys.SwingAngle)
	}
}

// TestPhysicsSystem_SwingDamping verifies swing damping reduces oscillation.
func TestPhysicsSystem_SwingDamping(t *testing.T) {
	world := ecs.NewWorld()
	sys := NewPhysicsSystem()

	e := world.CreateEntity()
	pos := &components.Position{X: 0, Y: 0, Z: 0}
	phys := components.NewSwingingBody(math.Pi/2, 0.5) // High damping
	phys.SwingVelocity = 2.0                           // Initial swing velocity
	world.AddComponent(e, pos)
	world.AddComponent(e, phys)

	initialVelocity := math.Abs(phys.SwingVelocity)

	// Simulate for some time
	for i := 0; i < 60; i++ {
		sys.Update(world, 0.016)
	}

	// Swing velocity should have decreased due to damping
	if math.Abs(phys.SwingVelocity) >= initialVelocity {
		t.Errorf("Swing velocity should decrease with damping, got %f (initial: %f)",
			math.Abs(phys.SwingVelocity), initialVelocity)
	}
}

// TestPhysicsSystem_MaxSwingAngle verifies swing angle is clamped.
func TestPhysicsSystem_MaxSwingAngle(t *testing.T) {
	world := ecs.NewWorld()
	sys := NewPhysicsSystem()

	maxAngle := math.Pi / 6 // 30 degrees

	e := world.CreateEntity()
	pos := &components.Position{X: 0, Y: 0, Z: 0}
	phys := components.NewSwingingBody(maxAngle, 0.01) // Low damping
	phys.SwingVelocity = 10.0                          // Very high initial velocity
	world.AddComponent(e, pos)
	world.AddComponent(e, phys)

	// Simulate
	for i := 0; i < 100; i++ {
		sys.Update(world, 0.016)
		// Check angle never exceeds max
		if math.Abs(phys.SwingAngle) > maxAngle+0.01 {
			t.Errorf("SwingAngle %f exceeded MaxSwingAngle %f", phys.SwingAngle, maxAngle)
			break
		}
	}
}

// TestPhysicsSystem_PushEntity verifies the PushEntity helper.
func TestPhysicsSystem_PushEntity(t *testing.T) {
	world := ecs.NewWorld()
	sys := NewPhysicsSystem()

	e := world.CreateEntity()
	pos := &components.Position{X: 0, Y: 0, Z: 0}
	phys := components.NewPushableBody(1.0, 0.5)
	world.AddComponent(e, pos)
	world.AddComponent(e, phys)

	// Push the entity
	pushed := sys.PushEntity(world, e, 5.0, 0.0)
	if !pushed {
		t.Error("PushEntity should return true for pushable entity")
	}

	if phys.VelocityX != 5.0 {
		t.Errorf("VelocityX = %f, expected 5.0", phys.VelocityX)
	}
}

// TestPhysicsSystem_PushEntity_NotPushable verifies non-pushable entities.
func TestPhysicsSystem_PushEntity_NotPushable(t *testing.T) {
	world := ecs.NewWorld()
	sys := NewPhysicsSystem()

	e := world.CreateEntity()
	pos := &components.Position{X: 0, Y: 0, Z: 0}
	phys := &components.PhysicsBody{
		Mass:       1.0,
		IsPushable: false, // Not pushable
	}
	world.AddComponent(e, pos)
	world.AddComponent(e, phys)

	pushed := sys.PushEntity(world, e, 5.0, 0.0)
	if pushed {
		t.Error("PushEntity should return false for non-pushable entity")
	}

	if phys.VelocityX != 0 {
		t.Error("Non-pushable entity should not have velocity changed")
	}
}

// TestPhysicsSystem_SwingEntity verifies the SwingEntity helper.
func TestPhysicsSystem_SwingEntity(t *testing.T) {
	world := ecs.NewWorld()
	sys := NewPhysicsSystem()

	e := world.CreateEntity()
	pos := &components.Position{X: 0, Y: 0, Z: 0}
	phys := components.NewSwingingBody(math.Pi/2, 0.1)
	world.AddComponent(e, pos)
	world.AddComponent(e, phys)

	swung := sys.SwingEntity(world, e, 1.5)
	if !swung {
		t.Error("SwingEntity should return true for swinging entity")
	}

	if phys.SwingVelocity != 1.5 {
		t.Errorf("SwingVelocity = %f, expected 1.5", phys.SwingVelocity)
	}
}

// TestPhysicsSystem_IsEntityMoving verifies movement detection.
func TestPhysicsSystem_IsEntityMoving(t *testing.T) {
	world := ecs.NewWorld()
	sys := NewPhysicsSystem()

	// Stationary entity
	e1 := world.CreateEntity()
	pos1 := &components.Position{X: 0, Y: 0, Z: 0}
	phys1 := components.NewPushableBody(1.0, 0.5)
	world.AddComponent(e1, pos1)
	world.AddComponent(e1, phys1)

	if sys.IsEntityMoving(world, e1) {
		t.Error("Stationary entity should not be moving")
	}

	// Moving entity
	e2 := world.CreateEntity()
	pos2 := &components.Position{X: 0, Y: 0, Z: 0}
	phys2 := &components.PhysicsBody{
		VelocityX: 5.0,
		Mass:      1.0,
	}
	world.AddComponent(e2, pos2)
	world.AddComponent(e2, phys2)

	if !sys.IsEntityMoving(world, e2) {
		t.Error("Entity with velocity should be moving")
	}
}

// TestPhysicsSystem_StopEntity verifies stopping an entity.
func TestPhysicsSystem_StopEntity(t *testing.T) {
	world := ecs.NewWorld()
	sys := NewPhysicsSystem()

	e := world.CreateEntity()
	pos := &components.Position{X: 0, Y: 0, Z: 0}
	phys := &components.PhysicsBody{
		VelocityX:     5.0,
		VelocityY:     3.0,
		VelocityZ:     2.0,
		SwingVelocity: 1.0,
		Mass:          1.0,
	}
	world.AddComponent(e, pos)
	world.AddComponent(e, phys)

	sys.StopEntity(world, e)

	if phys.VelocityX != 0 || phys.VelocityY != 0 || phys.VelocityZ != 0 || phys.SwingVelocity != 0 {
		t.Error("StopEntity should zero all velocities")
	}
}

// TestPhysicsSystem_Kinematic verifies kinematic objects ignore physics.
func TestPhysicsSystem_Kinematic(t *testing.T) {
	world := ecs.NewWorld()
	sys := NewPhysicsSystem()

	e := world.CreateEntity()
	pos := &components.Position{X: 0, Y: 0, Z: 5.0}
	phys := &components.PhysicsBody{
		Mass:        1.0,
		IsKinematic: true,
		Grounded:    false,
	}
	world.AddComponent(e, pos)
	world.AddComponent(e, phys)

	initialZ := pos.Z

	// Simulate
	for i := 0; i < 30; i++ {
		sys.Update(world, 0.016)
	}

	// Position should not change (kinematic ignores physics)
	if pos.Z != initialZ {
		t.Errorf("Kinematic object Z changed from %f to %f", initialZ, pos.Z)
	}
}

// TestPhysicsBody_ApplyForce verifies force application.
func TestPhysicsBody_ApplyForce(t *testing.T) {
	phys := &components.PhysicsBody{
		Mass: 2.0,
	}

	phys.ApplyForce(10.0, 0, 0) // Force of 10 on mass of 2 = acceleration of 5

	if phys.VelocityX != 5.0 {
		t.Errorf("VelocityX = %f, expected 5.0 (force/mass)", phys.VelocityX)
	}
}

// TestPhysicsBody_ApplyImpulse verifies impulse application.
func TestPhysicsBody_ApplyImpulse(t *testing.T) {
	phys := &components.PhysicsBody{
		Mass: 2.0,
	}

	phys.ApplyImpulse(5.0, 3.0, 1.0)

	if phys.VelocityX != 5.0 {
		t.Errorf("VelocityX = %f, expected 5.0", phys.VelocityX)
	}
	if phys.VelocityY != 3.0 {
		t.Errorf("VelocityY = %f, expected 3.0", phys.VelocityY)
	}
	if phys.VelocityZ != 1.0 {
		t.Errorf("VelocityZ = %f, expected 1.0", phys.VelocityZ)
	}
}
