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
		&FactionTerritory{Vertices: []Point2D{}, KillTracker: make(map[uint64]int)},
		&Schedule{TimeSlots: make(map[int]string)},
		&Inventory{},
		&Vehicle{},
		&Reputation{Standings: make(map[string]float64)},
		&Crime{},
		&Witness{},
		&EconomyNode{},
		&Quest{Flags: make(map[string]bool)},
		&WorldClock{},
		&Skills{Levels: make(map[string]int), Experience: make(map[string]float64)},
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

func TestSkillsType(t *testing.T) {
	s := &Skills{
		Levels:     map[string]int{"fire_magic": 10},
		Experience: map[string]float64{"fire_magic": 50.0},
	}
	if s.Type() != "Skills" {
		t.Errorf("expected Skills, got %s", s.Type())
	}
}

func TestGetSkillSchools(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			schools := GetSkillSchools(genre)
			if len(schools) != 6 {
				t.Errorf("expected 6 schools for %s, got %d", genre, len(schools))
			}
			for _, school := range schools {
				if school.ID == "" {
					t.Error("school has empty ID")
				}
				if school.Name == "" {
					t.Error("school has empty Name")
				}
				if len(school.Skills) != 5 {
					t.Errorf("expected 5 skills per school, got %d", len(school.Skills))
				}
			}
		})
	}
}

func TestGetSkillSchoolsDefault(t *testing.T) {
	// Unknown genre should fall back to fantasy
	schools := GetSkillSchools("unknown_genre")
	if len(schools) != 6 {
		t.Errorf("expected 6 schools for unknown genre (fallback), got %d", len(schools))
	}
}

func TestGetAllSkillIDs(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			skills := GetAllSkillIDs(genre)
			// 6 schools * 5 skills = 30 skills
			if len(skills) != 30 {
				t.Errorf("expected 30 skills for %s, got %d", genre, len(skills))
			}
		})
	}
}

func TestGenreSkillNamesUnique(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	// Collect all school names per genre
	schoolNames := make(map[string]map[string]bool)
	for _, genre := range genres {
		schoolNames[genre] = make(map[string]bool)
		for _, school := range GetSkillSchools(genre) {
			schoolNames[genre][school.Name] = true
		}
	}

	// At least some names should differ between genres
	fantasyNames := schoolNames["fantasy"]
	for _, otherGenre := range []string{"sci-fi", "cyberpunk", "post-apocalyptic"} {
		differentCount := 0
		for name := range fantasyNames {
			if !schoolNames[otherGenre][name] {
				differentCount++
			}
		}
		if differentCount < 3 {
			t.Errorf("expected at least 3 different school names between fantasy and %s", otherGenre)
		}
	}
}

func TestNewSkills(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			skills := NewSkills(genre)
			if skills == nil {
				t.Fatal("NewSkills returned nil")
			}
			if len(skills.Levels) != 30 {
				t.Errorf("expected 30 skills, got %d", len(skills.Levels))
			}
			// All skills should start at level 1
			for skillID, level := range skills.Levels {
				if level != 1 {
					t.Errorf("skill %s should start at level 1, got %d", skillID, level)
				}
			}
			// All skills should have 0 XP
			for skillID, xp := range skills.Experience {
				if xp != 0 {
					t.Errorf("skill %s should start with 0 XP, got %f", skillID, xp)
				}
			}
		})
	}
}

func TestGetVehicleArchetypes(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			archetypes := GetVehicleArchetypes(genre)
			if len(archetypes) != 3 {
				t.Errorf("expected 3 vehicle archetypes for %s, got %d", genre, len(archetypes))
			}
			for _, arch := range archetypes {
				if arch.ID == "" {
					t.Error("archetype has empty ID")
				}
				if arch.Name == "" {
					t.Error("archetype has empty Name")
				}
				if arch.BaseSpeed <= 0 {
					t.Errorf("archetype %s has invalid BaseSpeed", arch.ID)
				}
				if arch.MaxFuel <= 0 {
					t.Errorf("archetype %s has invalid MaxFuel", arch.ID)
				}
			}
		})
	}
}

func TestVehicleArchetypesUnique(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	// Collect all vehicle IDs across genres
	vehicleIDs := make(map[string]string) // id -> genre
	for _, genre := range genres {
		for _, arch := range GetVehicleArchetypes(genre) {
			// Vehicle types should be unique per genre
			if existingGenre, exists := vehicleIDs[arch.ID]; exists && existingGenre != genre {
				// Same ID in different genres is OK (each genre has its own)
				continue
			}
			vehicleIDs[arch.ID] = genre
		}
	}

	// Verify each genre has distinct vehicles
	for _, genre := range []string{"fantasy", "sci-fi"} {
		archetypes := GetVehicleArchetypes(genre)
		for _, arch := range archetypes {
			otherGenre := "cyberpunk"
			otherArchetypes := GetVehicleArchetypes(otherGenre)
			found := false
			for _, other := range otherArchetypes {
				if arch.ID == other.ID {
					found = true
					break
				}
			}
			if found {
				t.Errorf("vehicle %s found in both %s and %s", arch.ID, genre, otherGenre)
			}
		}
	}
}

func TestNewVehicleFromArchetype(t *testing.T) {
	arch := VehicleArchetype{
		ID:        "test_vehicle",
		Name:      "Test Vehicle",
		BaseSpeed: 20,
		MaxFuel:   150,
		FuelRate:  0.01,
	}
	vehicle := NewVehicleFromArchetype(arch)
	if vehicle.VehicleType != "test_vehicle" {
		t.Errorf("expected vehicle type test_vehicle, got %s", vehicle.VehicleType)
	}
	if vehicle.Speed != 20 {
		t.Errorf("expected speed 20, got %f", vehicle.Speed)
	}
	if vehicle.Fuel != 150 {
		t.Errorf("expected fuel 150, got %f", vehicle.Fuel)
	}
}
