// Package components defines all ECS component data types.
package components

// Position represents a 3D location in the world.
type Position struct {
	X, Y, Z float64
}

// Type returns the component type identifier for Position.
func (p *Position) Type() string { return "Position" }

// Health represents an entity's health state.
type Health struct {
	Current, Max float64
}

// Type returns the component type identifier for Health.
func (h *Health) Type() string { return "Health" }

// Faction represents an entity's faction allegiance and reputation.
type Faction struct {
	ID         string
	Reputation float64
}

// Type returns the component type identifier for Faction.
func (f *Faction) Type() string { return "Faction" }

// Schedule represents an NPC's daily activity schedule.
type Schedule struct {
	CurrentActivity string
	TimeSlots       map[int]string // hour -> activity
}

// Type returns the component type identifier for Schedule.
func (s *Schedule) Type() string { return "Schedule" }

// Inventory represents an entity's carried items.
type Inventory struct {
	Items    []string
	Capacity int
}

// Type returns the component type identifier for Inventory.
func (i *Inventory) Type() string { return "Inventory" }

// Vehicle represents a vehicle component.
type Vehicle struct {
	VehicleType string
	Speed       float64
	Fuel        float64
}

// Type returns the component type identifier for Vehicle.
func (v *Vehicle) Type() string { return "Vehicle" }
