package systems

import (
	"math"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// VehicleSystem manages vehicle movement and physics.
type VehicleSystem struct{}

// Update processes vehicle physics each tick.
func (s *VehicleSystem) Update(w *ecs.World, dt float64) {
	for _, e := range w.Entities("Vehicle", "Position") {
		s.updateVehicle(w, e, dt)
	}
}

// updateVehicle updates a single vehicle's position based on its physics.
func (s *VehicleSystem) updateVehicle(w *ecs.World, e ecs.Entity, dt float64) {
	vehicle, pos := s.getVehicleComponents(w, e)
	if vehicle == nil || pos == nil {
		return
	}
	if vehicle.Fuel > 0 && vehicle.Speed > 0 {
		s.applyVehicleMovement(vehicle, pos, dt)
	}
}

// getVehicleComponents retrieves vehicle and position components for an entity.
func (s *VehicleSystem) getVehicleComponents(w *ecs.World, e ecs.Entity) (*components.Vehicle, *components.Position) {
	vComp, ok := w.GetComponent(e, "Vehicle")
	if !ok {
		return nil, nil
	}
	pComp, ok := w.GetComponent(e, "Position")
	if !ok {
		return nil, nil
	}
	return vComp.(*components.Vehicle), pComp.(*components.Position)
}

// applyVehicleMovement updates position and consumes fuel.
func (s *VehicleSystem) applyVehicleMovement(vehicle *components.Vehicle, pos *components.Position, dt float64) {
	pos.X += math.Cos(vehicle.Direction) * vehicle.Speed * dt
	pos.Y += math.Sin(vehicle.Direction) * vehicle.Speed * dt
	vehicle.Fuel -= vehicle.Speed * dt * 0.01
	if vehicle.Fuel < 0 {
		vehicle.Fuel = 0
	}
}
