package systems

import (
	"math"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// Vehicle physics constants.
const (
	// DefaultMaxSpeed is the default max speed for vehicles.
	DefaultMaxSpeed = 20.0
	// DefaultAcceleration is the default acceleration rate.
	DefaultAcceleration = 10.0
	// DefaultDeceleration is the default braking deceleration.
	DefaultDeceleration = 20.0
	// DefaultFrictionDecel is the default friction slowdown.
	DefaultFrictionDecel = 5.0
	// DefaultMaxSteeringAngle is the max steering angle (45 degrees).
	DefaultMaxSteeringAngle = math.Pi / 4
	// DefaultSteeringSpeed is how fast steering responds.
	DefaultSteeringSpeed = 2.0
	// DefaultTurningRadius is the default minimum turning radius.
	DefaultTurningRadius = 5.0
	// DefaultMass is the default vehicle mass.
	DefaultMass = 1000.0
	// ReverseSpeedMultiplier reduces max speed when in reverse.
	ReverseSpeedMultiplier = 0.3
	// FuelConsumptionRate is fuel consumed per unit of speed per second.
	FuelConsumptionRate = 0.01
	// MinSpeedThreshold is the minimum speed to be considered moving.
	MinSpeedThreshold = 0.1
)

// VehiclePhysicsSystem handles vehicle movement, steering, and physics.
type VehiclePhysicsSystem struct {
	// Genre affects vehicle behavior and sounds.
	Genre string
}

// NewVehiclePhysicsSystem creates a new vehicle physics system.
func NewVehiclePhysicsSystem(genre string) *VehiclePhysicsSystem {
	return &VehiclePhysicsSystem{Genre: genre}
}

// Update processes vehicle physics for all vehicles with physics components.
func (s *VehiclePhysicsSystem) Update(w *ecs.World, dt float64) {
	for _, e := range w.Entities("Vehicle", "VehiclePhysics", "Position") {
		s.updateVehicle(w, e, dt)
	}
}

// updateVehicle processes physics for a single vehicle.
func (s *VehiclePhysicsSystem) updateVehicle(w *ecs.World, entity ecs.Entity, dt float64) {
	vehicleComp, _ := w.GetComponent(entity, "Vehicle")
	physicsComp, _ := w.GetComponent(entity, "VehiclePhysics")
	posComp, _ := w.GetComponent(entity, "Position")

	vehicle := vehicleComp.(*components.Vehicle)
	physics := physicsComp.(*components.VehiclePhysics)
	pos := posComp.(*components.Position)

	// Check if vehicle is operational
	stateComp, hasState := w.GetComponent(entity, "VehicleState")
	if hasState {
		state := stateComp.(*components.VehicleState)
		if !state.EngineRunning || state.DamagePercent >= 100 {
			// Vehicle not running - apply friction only
			s.applyFriction(physics, dt)
			return
		}
	}

	// Check fuel
	if vehicle.Fuel <= 0 {
		s.applyFriction(physics, dt)
		return
	}

	// Update steering angle based on input
	s.updateSteering(physics, dt)

	// Update speed based on throttle/brake
	s.updateSpeed(physics, vehicle, dt)

	// Calculate movement
	s.applyMovement(physics, vehicle, pos, dt)

	// Consume fuel
	s.consumeFuel(physics, vehicle, dt)
}

// updateSteering adjusts steering angle toward input.
func (s *VehiclePhysicsSystem) updateSteering(physics *components.VehiclePhysics, dt float64) {
	targetAngle := physics.Steering * physics.MaxSteeringAngle
	diff := targetAngle - physics.SteeringAngle

	maxChange := physics.SteeringSpeed * dt
	if math.Abs(diff) < maxChange {
		physics.SteeringAngle = targetAngle
	} else if diff > 0 {
		physics.SteeringAngle += maxChange
	} else {
		physics.SteeringAngle -= maxChange
	}
}

// updateSpeed adjusts current speed based on input.
func (s *VehiclePhysicsSystem) updateSpeed(physics *components.VehiclePhysics, vehicle *components.Vehicle, dt float64) {
	if physics.IsBraking {
		// Apply brakes
		if physics.CurrentSpeed > 0 {
			physics.CurrentSpeed -= physics.Deceleration * dt
			if physics.CurrentSpeed < 0 {
				physics.CurrentSpeed = 0
			}
		} else if physics.CurrentSpeed < 0 {
			physics.CurrentSpeed += physics.Deceleration * dt
			if physics.CurrentSpeed > 0 {
				physics.CurrentSpeed = 0
			}
		}
		return
	}

	maxSpeed := physics.MaxSpeed
	if physics.InReverse {
		maxSpeed *= ReverseSpeedMultiplier
	}

	if physics.Throttle > 0 && !physics.InReverse {
		// Accelerate forward
		physics.CurrentSpeed += physics.Acceleration * physics.Throttle * dt
		if physics.CurrentSpeed > maxSpeed {
			physics.CurrentSpeed = maxSpeed
		}
	} else if physics.Throttle < 0 || physics.InReverse {
		// Accelerate backward (reverse)
		throttle := physics.Throttle
		if physics.InReverse && physics.Throttle > 0 {
			throttle = -physics.Throttle
		}
		physics.CurrentSpeed += physics.Acceleration * throttle * dt
		if physics.CurrentSpeed < -maxSpeed {
			physics.CurrentSpeed = -maxSpeed
		}
	} else {
		// No input - apply friction
		s.applyFriction(physics, dt)
	}

	// Update vehicle Speed field for compatibility
	vehicle.Speed = math.Abs(physics.CurrentSpeed)
}

// applyFriction slows the vehicle passively.
func (s *VehiclePhysicsSystem) applyFriction(physics *components.VehiclePhysics, dt float64) {
	if math.Abs(physics.CurrentSpeed) < MinSpeedThreshold {
		physics.CurrentSpeed = 0
		return
	}

	friction := physics.FrictionDecel * dt
	if physics.CurrentSpeed > 0 {
		physics.CurrentSpeed -= friction
		if physics.CurrentSpeed < 0 {
			physics.CurrentSpeed = 0
		}
	} else {
		physics.CurrentSpeed += friction
		if physics.CurrentSpeed > 0 {
			physics.CurrentSpeed = 0
		}
	}
}

// applyMovement updates position and direction based on physics.
func (s *VehiclePhysicsSystem) applyMovement(physics *components.VehiclePhysics, vehicle *components.Vehicle, pos *components.Position, dt float64) {
	if math.Abs(physics.CurrentSpeed) < MinSpeedThreshold {
		return
	}

	// Calculate angular velocity based on steering and speed
	// Using bicycle model: angular_velocity = speed * tan(steering_angle) / wheelbase
	// Simplified: angular_velocity = speed * steering_angle / turning_radius
	angularVelocity := 0.0
	if physics.TurningRadius > 0 && math.Abs(physics.SteeringAngle) > 0.01 {
		angularVelocity = physics.CurrentSpeed * math.Tan(physics.SteeringAngle) / physics.TurningRadius
	}

	// Update direction
	vehicle.Direction += angularVelocity * dt

	// Normalize direction to [0, 2*PI)
	for vehicle.Direction < 0 {
		vehicle.Direction += 2 * math.Pi
	}
	for vehicle.Direction >= 2*math.Pi {
		vehicle.Direction -= 2 * math.Pi
	}

	// Calculate velocity components
	vx := physics.CurrentSpeed * math.Cos(vehicle.Direction)
	vy := physics.CurrentSpeed * math.Sin(vehicle.Direction)

	// Update position
	pos.X += vx * dt
	pos.Y += vy * dt
}

// consumeFuel reduces fuel based on speed and time.
func (s *VehiclePhysicsSystem) consumeFuel(physics *components.VehiclePhysics, vehicle *components.Vehicle, dt float64) {
	consumption := math.Abs(physics.CurrentSpeed) * FuelConsumptionRate * dt
	vehicle.Fuel -= consumption
	if vehicle.Fuel < 0 {
		vehicle.Fuel = 0
	}
}

// SetThrottle sets the throttle input (-1 to 1).
func (s *VehiclePhysicsSystem) SetThrottle(w *ecs.World, vehicle ecs.Entity, throttle float64) {
	physicsComp, ok := w.GetComponent(vehicle, "VehiclePhysics")
	if !ok {
		return
	}
	physics := physicsComp.(*components.VehiclePhysics)

	if throttle < -1 {
		throttle = -1
	} else if throttle > 1 {
		throttle = 1
	}
	physics.Throttle = throttle
}

// SetSteering sets the steering input (-1 = full left, 1 = full right).
func (s *VehiclePhysicsSystem) SetSteering(w *ecs.World, vehicle ecs.Entity, steering float64) {
	physicsComp, ok := w.GetComponent(vehicle, "VehiclePhysics")
	if !ok {
		return
	}
	physics := physicsComp.(*components.VehiclePhysics)

	if steering < -1 {
		steering = -1
	} else if steering > 1 {
		steering = 1
	}
	physics.Steering = steering
}

// SetBraking sets the brake state.
func (s *VehiclePhysicsSystem) SetBraking(w *ecs.World, vehicle ecs.Entity, braking bool) {
	physicsComp, ok := w.GetComponent(vehicle, "VehiclePhysics")
	if !ok {
		return
	}
	physics := physicsComp.(*components.VehiclePhysics)
	physics.IsBraking = braking
}

// SetReverse toggles reverse mode.
func (s *VehiclePhysicsSystem) SetReverse(w *ecs.World, vehicle ecs.Entity, reverse bool) {
	physicsComp, ok := w.GetComponent(vehicle, "VehiclePhysics")
	if !ok {
		return
	}
	physics := physicsComp.(*components.VehiclePhysics)
	physics.InReverse = reverse
}

// EnterVehicle places a driver in a vehicle.
func (s *VehiclePhysicsSystem) EnterVehicle(w *ecs.World, vehicle, driver ecs.Entity) bool {
	stateComp, ok := w.GetComponent(vehicle, "VehicleState")
	if !ok {
		// Create state if not exists
		state := &components.VehicleState{
			IsOccupied:        false,
			MaxPassengers:     1,
			InCockpitView:     false,
			EngineRunning:     false,
			DamagePercent:     0,
			PassengerEntities: []uint64{},
		}
		w.AddComponent(vehicle, state)
		stateComp, _ = w.GetComponent(vehicle, "VehicleState")
	}
	state := stateComp.(*components.VehicleState)

	if state.IsOccupied {
		return false
	}

	state.IsOccupied = true
	state.DriverEntity = uint64(driver)
	state.EngineRunning = true
	state.InCockpitView = true

	return true
}

// ExitVehicle removes the driver from a vehicle.
func (s *VehiclePhysicsSystem) ExitVehicle(w *ecs.World, vehicle ecs.Entity) ecs.Entity {
	stateComp, ok := w.GetComponent(vehicle, "VehicleState")
	if !ok {
		return 0
	}
	state := stateComp.(*components.VehicleState)

	if !state.IsOccupied {
		return 0
	}

	driver := ecs.Entity(state.DriverEntity)
	state.IsOccupied = false
	state.DriverEntity = 0
	state.EngineRunning = false
	state.InCockpitView = false

	// Stop the vehicle when exiting
	physicsComp, _ := w.GetComponent(vehicle, "VehiclePhysics")
	if physicsComp != nil {
		physics := physicsComp.(*components.VehiclePhysics)
		physics.Throttle = 0
		physics.Steering = 0
	}

	return driver
}

// IsVehicleOccupied checks if a vehicle has a driver.
func (s *VehiclePhysicsSystem) IsVehicleOccupied(w *ecs.World, vehicle ecs.Entity) bool {
	stateComp, ok := w.GetComponent(vehicle, "VehicleState")
	if !ok {
		return false
	}
	state := stateComp.(*components.VehicleState)
	return state.IsOccupied
}

// GetVehicleSpeed returns the current speed of a vehicle.
func (s *VehiclePhysicsSystem) GetVehicleSpeed(w *ecs.World, vehicle ecs.Entity) float64 {
	physicsComp, ok := w.GetComponent(vehicle, "VehiclePhysics")
	if !ok {
		return 0
	}
	physics := physicsComp.(*components.VehiclePhysics)
	return math.Abs(physics.CurrentSpeed)
}

// CreateVehicle creates a new vehicle entity with all necessary components.
func (s *VehiclePhysicsSystem) CreateVehicle(w *ecs.World, vehicleType string, x, y, z, direction float64) ecs.Entity {
	entity := w.CreateEntity()

	// Add position
	w.AddComponent(entity, &components.Position{X: x, Y: y, Z: z})

	// Get archetype for this vehicle type
	archetype := s.getArchetype(vehicleType)

	// Add vehicle component
	w.AddComponent(entity, &components.Vehicle{
		VehicleType: vehicleType,
		Speed:       0,
		Fuel:        archetype.MaxFuel,
		Direction:   direction,
	})

	// Add physics component
	w.AddComponent(entity, &components.VehiclePhysics{
		CurrentSpeed:     0,
		MaxSpeed:         archetype.BaseSpeed,
		Acceleration:     DefaultAcceleration,
		Deceleration:     DefaultDeceleration,
		FrictionDecel:    DefaultFrictionDecel,
		SteeringAngle:    0,
		MaxSteeringAngle: DefaultMaxSteeringAngle,
		SteeringSpeed:    DefaultSteeringSpeed,
		TurningRadius:    DefaultTurningRadius,
		Mass:             DefaultMass,
		Throttle:         0,
		Steering:         0,
		IsBraking:        false,
		InReverse:        false,
	})

	// Add vehicle state
	w.AddComponent(entity, &components.VehicleState{
		IsOccupied:        false,
		DriverEntity:      0,
		PassengerEntities: []uint64{},
		MaxPassengers:     s.getMaxPassengers(vehicleType),
		InCockpitView:     false,
		EngineRunning:     false,
		DamagePercent:     0,
	})

	return entity
}

// getArchetype returns the archetype for a vehicle type.
func (s *VehiclePhysicsSystem) getArchetype(vehicleType string) components.VehicleArchetype {
	archetypes, exists := components.GenreVehicleArchetypes[s.Genre]
	if !exists {
		archetypes = components.GenreVehicleArchetypes["fantasy"]
	}

	for _, arch := range archetypes {
		if arch.ID == vehicleType {
			return arch
		}
	}

	// Return default
	return components.VehicleArchetype{
		ID:        vehicleType,
		Name:      vehicleType,
		BaseSpeed: DefaultMaxSpeed,
		MaxFuel:   200,
		FuelRate:  0.01,
	}
}

// getMaxPassengers returns max passengers for a vehicle type.
func (s *VehiclePhysicsSystem) getMaxPassengers(vehicleType string) int {
	switch vehicleType {
	case "horse", "hoverbike", "motorbike":
		return 1
	case "cart", "buggy", "hearse":
		return 3
	case "ship", "shuttle", "apc", "truck":
		return 6
	case "mech":
		return 1
	default:
		return 2
	}
}

// RefuelVehicle adds fuel to a vehicle.
func (s *VehiclePhysicsSystem) RefuelVehicle(w *ecs.World, vehicle ecs.Entity, amount float64) float64 {
	vehicleComp, ok := w.GetComponent(vehicle, "Vehicle")
	if !ok {
		return 0
	}
	veh := vehicleComp.(*components.Vehicle)

	archetype := s.getArchetype(veh.VehicleType)
	maxFuel := archetype.MaxFuel

	oldFuel := veh.Fuel
	veh.Fuel += amount
	if veh.Fuel > maxFuel {
		veh.Fuel = maxFuel
	}

	return veh.Fuel - oldFuel
}

// DamageVehicle applies damage to a vehicle.
func (s *VehiclePhysicsSystem) DamageVehicle(w *ecs.World, vehicle ecs.Entity, damage float64) {
	stateComp, ok := w.GetComponent(vehicle, "VehicleState")
	if !ok {
		return
	}
	state := stateComp.(*components.VehicleState)

	state.DamagePercent += damage
	if state.DamagePercent > 100 {
		state.DamagePercent = 100
		state.EngineRunning = false
	}
}

// RepairVehicle reduces damage on a vehicle.
func (s *VehiclePhysicsSystem) RepairVehicle(w *ecs.World, vehicle ecs.Entity, repair float64) {
	stateComp, ok := w.GetComponent(vehicle, "VehicleState")
	if !ok {
		return
	}
	state := stateComp.(*components.VehicleState)

	state.DamagePercent -= repair
	if state.DamagePercent < 0 {
		state.DamagePercent = 0
	}
}
