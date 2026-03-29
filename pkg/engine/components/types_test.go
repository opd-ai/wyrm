package components

import "testing"

func TestPositionType(t *testing.T) {
	p := &Position{X: 1, Y: 2, Z: 3}
	if p.Type() != "Position" {
		t.Errorf("expected Position, got %s", p.Type())
	}
}

func TestHealthType(t *testing.T) {
	h := &Health{Current: 100, Max: 100}
	if h.Type() != "Health" {
		t.Errorf("expected Health, got %s", h.Type())
	}
}

func TestFactionType(t *testing.T) {
	f := &Faction{ID: "guild", Reputation: 50}
	if f.Type() != "Faction" {
		t.Errorf("expected Faction, got %s", f.Type())
	}
}

func TestScheduleType(t *testing.T) {
	s := &Schedule{
		CurrentActivity: "work",
		TimeSlots:       map[int]string{8: "work", 12: "eat"},
	}
	if s.Type() != "Schedule" {
		t.Errorf("expected Schedule, got %s", s.Type())
	}
}

func TestInventoryType(t *testing.T) {
	i := &Inventory{Items: []string{"sword"}, Capacity: 10}
	if i.Type() != "Inventory" {
		t.Errorf("expected Inventory, got %s", i.Type())
	}
}

func TestVehicleType(t *testing.T) {
	v := &Vehicle{VehicleType: "horse", Speed: 10, Fuel: 100}
	if v.Type() != "Vehicle" {
		t.Errorf("expected Vehicle, got %s", v.Type())
	}
}

func TestComponentImplementsInterface(t *testing.T) {
	// Verify all components implement the Component interface via Type()
	components := []interface{ Type() string }{
		&Position{},
		&Health{},
		&Faction{},
		&Schedule{TimeSlots: make(map[int]string)},
		&Inventory{},
		&Vehicle{},
		&Reputation{Standings: make(map[string]float64)},
		&Crime{},
		&Witness{},
		&EconomyNode{},
		&Quest{Flags: make(map[string]bool)},
		&WorldClock{},
	}

	for _, c := range components {
		if c.Type() == "" {
			t.Error("component Type() returned empty string")
		}
	}
}

func TestReputationType(t *testing.T) {
	r := &Reputation{Standings: map[string]float64{"guild": 50.0}}
	if r.Type() != "Reputation" {
		t.Errorf("expected Reputation, got %s", r.Type())
	}
}

func TestCrimeType(t *testing.T) {
	c := &Crime{WantedLevel: 2, BountyAmount: 500.0}
	if c.Type() != "Crime" {
		t.Errorf("expected Crime, got %s", c.Type())
	}
}

func TestWitnessType(t *testing.T) {
	w := &Witness{CanReport: true}
	if w.Type() != "Witness" {
		t.Errorf("expected Witness, got %s", w.Type())
	}
}

func TestEconomyNodeType(t *testing.T) {
	e := &EconomyNode{PriceTable: map[string]float64{"sword": 100.0}}
	if e.Type() != "EconomyNode" {
		t.Errorf("expected EconomyNode, got %s", e.Type())
	}
}

func TestQuestType(t *testing.T) {
	q := &Quest{ID: "main", CurrentStage: 1, Flags: map[string]bool{"start": true}}
	if q.Type() != "Quest" {
		t.Errorf("expected Quest, got %s", q.Type())
	}
}

func TestWorldClockType(t *testing.T) {
	wc := &WorldClock{Hour: 12, Day: 1, HourLength: 60.0}
	if wc.Type() != "WorldClock" {
		t.Errorf("expected WorldClock, got %s", wc.Type())
	}
}
