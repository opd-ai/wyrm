package systems

import (
	"math"
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

func TestVehiclePhysicsSystem_CreateVehicle(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewVehiclePhysicsSystem("fantasy")

	vehicle := sys.CreateVehicle(w, "horse", 100, 100, 0, 0)

	// Check all required components
	_, hasPos := w.GetComponent(vehicle, "Position")
	_, hasVehicle := w.GetComponent(vehicle, "Vehicle")
	_, hasPhysics := w.GetComponent(vehicle, "VehiclePhysics")
	_, hasState := w.GetComponent(vehicle, "VehicleState")

	if !hasPos || !hasVehicle || !hasPhysics || !hasState {
		t.Error("Vehicle should have Position, Vehicle, VehiclePhysics, and VehicleState components")
	}

	// Check vehicle properties
	vehicleComp, _ := w.GetComponent(vehicle, "Vehicle")
	veh := vehicleComp.(*components.Vehicle)

	if veh.VehicleType != "horse" {
		t.Errorf("Expected vehicle type 'horse', got %s", veh.VehicleType)
	}
	if veh.Fuel <= 0 {
		t.Error("Vehicle should have fuel")
	}
}

func TestVehiclePhysicsSystem_Acceleration(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewVehiclePhysicsSystem("fantasy")

	vehicle := sys.CreateVehicle(w, "horse", 0, 0, 0, 0)

	// Enter vehicle to enable engine
	driver := w.CreateEntity()
	sys.EnterVehicle(w, vehicle, driver)

	// Apply throttle
	sys.SetThrottle(w, vehicle, 1.0)

	// Update physics
	for i := 0; i < 10; i++ {
		sys.Update(w, 0.1)
	}

	speed := sys.GetVehicleSpeed(w, vehicle)
	if speed <= 0 {
		t.Error("Vehicle should have accelerated")
	}
}

func TestVehiclePhysicsSystem_Braking(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewVehiclePhysicsSystem("fantasy")

	vehicle := sys.CreateVehicle(w, "horse", 0, 0, 0, 0)
	driver := w.CreateEntity()
	sys.EnterVehicle(w, vehicle, driver)

	// Accelerate
	sys.SetThrottle(w, vehicle, 1.0)
	for i := 0; i < 20; i++ {
		sys.Update(w, 0.1)
	}

	speedBefore := sys.GetVehicleSpeed(w, vehicle)

	// Apply brakes
	sys.SetBraking(w, vehicle, true)
	sys.SetThrottle(w, vehicle, 0)
	for i := 0; i < 10; i++ {
		sys.Update(w, 0.1)
	}

	speedAfter := sys.GetVehicleSpeed(w, vehicle)
	if speedAfter >= speedBefore {
		t.Errorf("Speed should decrease with braking, before: %f, after: %f", speedBefore, speedAfter)
	}
}

func TestVehiclePhysicsSystem_Steering(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewVehiclePhysicsSystem("fantasy")

	vehicle := sys.CreateVehicle(w, "horse", 0, 0, 0, 0)
	driver := w.CreateEntity()
	sys.EnterVehicle(w, vehicle, driver)

	vehicleComp, _ := w.GetComponent(vehicle, "Vehicle")
	veh := vehicleComp.(*components.Vehicle)

	initialDirection := veh.Direction

	// Accelerate and steer right
	sys.SetThrottle(w, vehicle, 1.0)
	sys.SetSteering(w, vehicle, 1.0)

	for i := 0; i < 50; i++ {
		sys.Update(w, 0.1)
	}

	if veh.Direction == initialDirection {
		t.Error("Vehicle direction should change when steering")
	}
}

func TestVehiclePhysicsSystem_FuelConsumption(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewVehiclePhysicsSystem("fantasy")

	vehicle := sys.CreateVehicle(w, "horse", 0, 0, 0, 0)
	driver := w.CreateEntity()
	sys.EnterVehicle(w, vehicle, driver)

	vehicleComp, _ := w.GetComponent(vehicle, "Vehicle")
	veh := vehicleComp.(*components.Vehicle)

	initialFuel := veh.Fuel

	// Drive for a while
	sys.SetThrottle(w, vehicle, 1.0)
	for i := 0; i < 100; i++ {
		sys.Update(w, 0.1)
	}

	if veh.Fuel >= initialFuel {
		t.Error("Fuel should be consumed while driving")
	}
}

func TestVehiclePhysicsSystem_EnterExitVehicle(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewVehiclePhysicsSystem("fantasy")

	vehicle := sys.CreateVehicle(w, "horse", 0, 0, 0, 0)
	driver := w.CreateEntity()

	// Initially not occupied
	if sys.IsVehicleOccupied(w, vehicle) {
		t.Error("Vehicle should not be occupied initially")
	}

	// Enter vehicle
	success := sys.EnterVehicle(w, vehicle, driver)
	if !success {
		t.Error("Entering vehicle should succeed")
	}

	if !sys.IsVehicleOccupied(w, vehicle) {
		t.Error("Vehicle should be occupied after entering")
	}

	// Try to enter again (should fail)
	driver2 := w.CreateEntity()
	success2 := sys.EnterVehicle(w, vehicle, driver2)
	if success2 {
		t.Error("Second driver should not be able to enter")
	}

	// Exit vehicle
	exitedDriver := sys.ExitVehicle(w, vehicle)
	if exitedDriver != driver {
		t.Error("ExitVehicle should return the driver")
	}

	if sys.IsVehicleOccupied(w, vehicle) {
		t.Error("Vehicle should not be occupied after exiting")
	}
}

func TestVehiclePhysicsSystem_Reverse(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewVehiclePhysicsSystem("fantasy")

	vehicle := sys.CreateVehicle(w, "horse", 0, 0, 0, 0)
	driver := w.CreateEntity()
	sys.EnterVehicle(w, vehicle, driver)

	posComp, _ := w.GetComponent(vehicle, "Position")
	pos := posComp.(*components.Position)
	initialX := pos.X

	// Set reverse and throttle
	sys.SetReverse(w, vehicle, true)
	sys.SetThrottle(w, vehicle, 1.0)

	for i := 0; i < 30; i++ {
		sys.Update(w, 0.1)
	}

	physicsComp, _ := w.GetComponent(vehicle, "VehiclePhysics")
	physics := physicsComp.(*components.VehiclePhysics)

	if physics.CurrentSpeed >= 0 {
		t.Error("Speed should be negative in reverse")
	}

	// Position should move backward (negative X for direction 0)
	if pos.X >= initialX {
		t.Error("Vehicle should move backward in reverse")
	}
}

func TestVehiclePhysicsSystem_NoFuelStops(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewVehiclePhysicsSystem("fantasy")

	vehicle := sys.CreateVehicle(w, "horse", 0, 0, 0, 0)
	driver := w.CreateEntity()
	sys.EnterVehicle(w, vehicle, driver)

	// Drain all fuel
	vehicleComp, _ := w.GetComponent(vehicle, "Vehicle")
	veh := vehicleComp.(*components.Vehicle)
	veh.Fuel = 0

	// Try to accelerate
	sys.SetThrottle(w, vehicle, 1.0)
	for i := 0; i < 10; i++ {
		sys.Update(w, 0.1)
	}

	speed := sys.GetVehicleSpeed(w, vehicle)
	if speed > MinSpeedThreshold {
		t.Error("Vehicle should not accelerate without fuel")
	}
}

func TestVehiclePhysicsSystem_DamageStops(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewVehiclePhysicsSystem("fantasy")

	vehicle := sys.CreateVehicle(w, "horse", 0, 0, 0, 0)
	driver := w.CreateEntity()
	sys.EnterVehicle(w, vehicle, driver)

	// Destroy vehicle
	sys.DamageVehicle(w, vehicle, 100)

	// Try to accelerate
	sys.SetThrottle(w, vehicle, 1.0)
	for i := 0; i < 10; i++ {
		sys.Update(w, 0.1)
	}

	speed := sys.GetVehicleSpeed(w, vehicle)
	if speed > MinSpeedThreshold {
		t.Error("Destroyed vehicle should not move")
	}
}

func TestVehiclePhysicsSystem_RefuelVehicle(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewVehiclePhysicsSystem("fantasy")

	vehicle := sys.CreateVehicle(w, "horse", 0, 0, 0, 0)

	vehicleComp, _ := w.GetComponent(vehicle, "Vehicle")
	veh := vehicleComp.(*components.Vehicle)

	// Use some fuel
	veh.Fuel = 50

	// Refuel
	added := sys.RefuelVehicle(w, vehicle, 100)

	if added <= 0 {
		t.Error("Should have added fuel")
	}
	if veh.Fuel < 100 {
		t.Errorf("Fuel should be at least 100, got %f", veh.Fuel)
	}
}

func TestVehiclePhysicsSystem_RepairVehicle(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewVehiclePhysicsSystem("fantasy")

	vehicle := sys.CreateVehicle(w, "horse", 0, 0, 0, 0)

	// Damage vehicle
	sys.DamageVehicle(w, vehicle, 50)

	stateComp, _ := w.GetComponent(vehicle, "VehicleState")
	state := stateComp.(*components.VehicleState)

	if state.DamagePercent != 50 {
		t.Errorf("Expected 50%% damage, got %f%%", state.DamagePercent)
	}

	// Repair
	sys.RepairVehicle(w, vehicle, 30)

	if state.DamagePercent != 20 {
		t.Errorf("Expected 20%% damage after repair, got %f%%", state.DamagePercent)
	}
}

func TestVehiclePhysicsSystem_FrictionSlowdown(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewVehiclePhysicsSystem("fantasy")

	vehicle := sys.CreateVehicle(w, "horse", 0, 0, 0, 0)
	driver := w.CreateEntity()
	sys.EnterVehicle(w, vehicle, driver)

	// Accelerate
	sys.SetThrottle(w, vehicle, 1.0)
	for i := 0; i < 20; i++ {
		sys.Update(w, 0.1)
	}

	speedBefore := sys.GetVehicleSpeed(w, vehicle)

	// Release throttle
	sys.SetThrottle(w, vehicle, 0)
	for i := 0; i < 50; i++ {
		sys.Update(w, 0.1)
	}

	speedAfter := sys.GetVehicleSpeed(w, vehicle)
	if speedAfter >= speedBefore {
		t.Error("Vehicle should slow down due to friction")
	}
}

func TestVehiclePhysicsSystem_PositionUpdate(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewVehiclePhysicsSystem("fantasy")

	vehicle := sys.CreateVehicle(w, "horse", 0, 0, 0, 0) // Direction 0 = East
	driver := w.CreateEntity()
	sys.EnterVehicle(w, vehicle, driver)

	posComp, _ := w.GetComponent(vehicle, "Position")
	pos := posComp.(*components.Position)

	initialX := pos.X

	// Accelerate
	sys.SetThrottle(w, vehicle, 1.0)
	for i := 0; i < 20; i++ {
		sys.Update(w, 0.1)
	}

	// Position should have moved east (positive X)
	if pos.X <= initialX {
		t.Error("Vehicle should have moved in the X direction")
	}
}

func TestVehiclePhysicsSystem_CockpitView(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewVehiclePhysicsSystem("fantasy")

	vehicle := sys.CreateVehicle(w, "horse", 0, 0, 0, 0)
	driver := w.CreateEntity()

	// Enter vehicle
	sys.EnterVehicle(w, vehicle, driver)

	stateComp, _ := w.GetComponent(vehicle, "VehicleState")
	state := stateComp.(*components.VehicleState)

	if !state.InCockpitView {
		t.Error("Should be in cockpit view when entering vehicle")
	}

	// Exit vehicle
	sys.ExitVehicle(w, vehicle)

	if state.InCockpitView {
		t.Error("Should not be in cockpit view after exiting vehicle")
	}
}

func TestVehiclePhysicsSystem_MaxSpeedLimit(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewVehiclePhysicsSystem("fantasy")

	vehicle := sys.CreateVehicle(w, "horse", 0, 0, 0, 0)
	driver := w.CreateEntity()
	sys.EnterVehicle(w, vehicle, driver)

	physicsComp, _ := w.GetComponent(vehicle, "VehiclePhysics")
	physics := physicsComp.(*components.VehiclePhysics)

	// Accelerate for a long time
	sys.SetThrottle(w, vehicle, 1.0)
	for i := 0; i < 100; i++ {
		sys.Update(w, 0.1)
	}

	if physics.CurrentSpeed > physics.MaxSpeed {
		t.Errorf("Speed %f should not exceed max speed %f", physics.CurrentSpeed, physics.MaxSpeed)
	}
}

func TestVehiclePhysicsSystem_GenreVehicles(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}
	vehicleTypes := map[string]string{
		"fantasy":          "horse",
		"sci-fi":           "hoverbike",
		"horror":           "hearse",
		"cyberpunk":        "motorbike",
		"post-apocalyptic": "buggy",
	}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			w := ecs.NewWorld()
			sys := NewVehiclePhysicsSystem(genre)

			vehicleType := vehicleTypes[genre]
			vehicle := sys.CreateVehicle(w, vehicleType, 0, 0, 0, 0)

			vehicleComp, ok := w.GetComponent(vehicle, "Vehicle")
			if !ok {
				t.Fatal("Vehicle should have Vehicle component")
			}
			veh := vehicleComp.(*components.Vehicle)

			if veh.VehicleType != vehicleType {
				t.Errorf("Expected vehicle type %s, got %s", vehicleType, veh.VehicleType)
			}

			physicsComp, ok := w.GetComponent(vehicle, "VehiclePhysics")
			if !ok {
				t.Fatal("Vehicle should have VehiclePhysics component")
			}
			physics := physicsComp.(*components.VehiclePhysics)

			if physics.MaxSpeed <= 0 {
				t.Error("Vehicle should have positive max speed")
			}
		})
	}
}

func TestVehiclePhysicsSystem_TurningRadius(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewVehiclePhysicsSystem("fantasy")

	vehicle := sys.CreateVehicle(w, "horse", 0, 0, 0, 0)
	driver := w.CreateEntity()
	sys.EnterVehicle(w, vehicle, driver)

	vehicleComp, _ := w.GetComponent(vehicle, "Vehicle")
	veh := vehicleComp.(*components.Vehicle)

	// Accelerate and turn
	sys.SetThrottle(w, vehicle, 1.0)
	sys.SetSteering(w, vehicle, 1.0) // Full right

	// Track angle change
	directions := make([]float64, 0)
	for i := 0; i < 50; i++ {
		sys.Update(w, 0.1)
		directions = append(directions, veh.Direction)
	}

	// Direction should have changed consistently
	totalChange := directions[len(directions)-1] - directions[0]
	if math.Abs(totalChange) < 0.1 {
		t.Error("Vehicle should turn when steering is applied")
	}
}
