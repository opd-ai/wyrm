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

// Reputation represents an entity's standing with various factions.
type Reputation struct {
	// Standings maps faction ID to reputation value (-100 to 100).
	Standings map[string]float64
}

// Type returns the component type identifier for Reputation.
func (r *Reputation) Type() string { return "Reputation" }

// Crime represents an entity's criminal status.
type Crime struct {
	WantedLevel   int     // 0-5 stars
	BountyAmount  float64 // currency owed
	LastCrimeTime float64 // game time of last offense
}

// Type returns the component type identifier for Crime.
func (c *Crime) Type() string { return "Crime" }

// Witness is a tag component marking NPCs that can report crimes.
type Witness struct {
	CanReport bool
}

// Type returns the component type identifier for Witness.
func (w *Witness) Type() string { return "Witness" }

// EconomyNode represents a location with supply/demand pricing.
type EconomyNode struct {
	// PriceTable maps item type to current price.
	PriceTable map[string]float64
	// Supply maps item type to available quantity.
	Supply map[string]int
	// Demand maps item type to desired quantity.
	Demand map[string]int
}

// Type returns the component type identifier for EconomyNode.
func (e *EconomyNode) Type() string { return "EconomyNode" }

// Quest represents an active quest with branching state.
type Quest struct {
	ID           string
	CurrentStage int
	Flags        map[string]bool
	Completed    bool
}

// Type returns the component type identifier for Quest.
func (q *Quest) Type() string { return "Quest" }
