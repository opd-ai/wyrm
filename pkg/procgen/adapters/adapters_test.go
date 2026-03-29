//go:build ebitentest

// This test file requires the "ebitentest" build tag because the Venture
// faction generator imports ebiten, which requires GLFW/X11 initialization.
// Run with: go test -tags=ebitentest ./pkg/procgen/adapters/...
// Or use xvfb-run: xvfb-run go test -tags=ebitentest ./pkg/procgen/adapters/...

package adapters

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/engine/systems"
)

func TestEntityAdapter_GenerateEntity(t *testing.T) {
	adapter := NewEntityAdapter()

	tests := []struct {
		name    string
		seed    int64
		genre   string
		depth   int
		wantErr bool
	}{
		{"fantasy entity", 12345, "fantasy", 5, false},
		{"sci-fi entity", 12345, "sci-fi", 10, false},
		{"horror entity", 12345, "horror", 15, false},
		{"cyberpunk entity", 12345, "cyberpunk", 20, false},
		{"post-apocalyptic entity", 12345, "post-apocalyptic", 25, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := adapter.GenerateEntity(tt.seed, tt.genre, tt.depth)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateEntity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if data == nil && !tt.wantErr {
				t.Error("GenerateEntity() returned nil data")
				return
			}
			if data != nil {
				if data.Name == "" {
					t.Error("Generated entity has empty name")
				}
				if data.Health <= 0 {
					t.Error("Generated entity has non-positive health")
				}
			}
		})
	}
}

func TestEntityAdapter_Deterministic(t *testing.T) {
	adapter := NewEntityAdapter()
	seed := int64(42)
	genre := "fantasy"
	depth := 5

	data1, err := adapter.GenerateEntity(seed, genre, depth)
	if err != nil {
		t.Fatalf("First generation failed: %v", err)
	}

	data2, err := adapter.GenerateEntity(seed, genre, depth)
	if err != nil {
		t.Fatalf("Second generation failed: %v", err)
	}

	if data1.Name != data2.Name {
		t.Errorf("Determinism failed: names differ (%s vs %s)", data1.Name, data2.Name)
	}
	if data1.Health != data2.Health {
		t.Errorf("Determinism failed: health differs (%v vs %v)", data1.Health, data2.Health)
	}
}

func TestSpawnNPC(t *testing.T) {
	world := ecs.NewWorld()
	data := &NPCData{
		Name:   "TestNPC",
		Health: 100,
		Tags:   []string{"test"},
	}

	entity, err := SpawnNPC(world, data, 10.0, 20.0, "test_faction")
	if err != nil {
		t.Fatalf("SpawnNPC failed: %v", err)
	}

	if entity == 0 {
		t.Error("SpawnNPC returned zero entity ID")
	}

	// Verify components were added
	pos, ok := world.GetComponent(entity, "Position")
	if !ok {
		t.Error("Position component not found")
	} else {
		t.Logf("Position: %+v", pos)
	}

	health, ok := world.GetComponent(entity, "Health")
	if !ok {
		t.Error("Health component not found")
	} else {
		t.Logf("Health: %+v", health)
	}

	faction, ok := world.GetComponent(entity, "Faction")
	if !ok {
		t.Error("Faction component not found")
	} else {
		t.Logf("Faction: %+v", faction)
	}

	schedule, ok := world.GetComponent(entity, "Schedule")
	if !ok {
		t.Error("Schedule component not found")
	} else {
		t.Logf("Schedule: %+v", schedule)
	}
}

func TestFactionAdapter_GenerateFactions(t *testing.T) {
	adapter := NewFactionAdapter()

	tests := []struct {
		name    string
		seed    int64
		genre   string
		depth   int
		wantErr bool
	}{
		{"fantasy factions", 12345, "fantasy", 10, false},
		{"sci-fi factions", 12345, "sci-fi", 20, false},
		{"horror factions", 12345, "horror", 30, false},
		{"cyberpunk factions", 12345, "cyberpunk", 40, false},
		{"post-apocalyptic factions", 12345, "post-apocalyptic", 50, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factions, err := adapter.GenerateFactions(tt.seed, tt.genre, tt.depth)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateFactions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(factions) == 0 && !tt.wantErr {
				t.Error("GenerateFactions() returned no factions")
				return
			}
			for _, f := range factions {
				if f.Name == "" {
					t.Error("Faction has empty name")
				}
				if f.ID == "" {
					t.Error("Faction has empty ID")
				}
			}
		})
	}
}

func TestFactionAdapter_Deterministic(t *testing.T) {
	adapter := NewFactionAdapter()
	seed := int64(42)
	genre := "fantasy"
	depth := 20

	factions1, err := adapter.GenerateFactions(seed, genre, depth)
	if err != nil {
		t.Fatalf("First generation failed: %v", err)
	}

	factions2, err := adapter.GenerateFactions(seed, genre, depth)
	if err != nil {
		t.Fatalf("Second generation failed: %v", err)
	}

	if len(factions1) != len(factions2) {
		t.Fatalf("Determinism failed: count differs (%d vs %d)", len(factions1), len(factions2))
	}

	for i := range factions1 {
		if factions1[i].Name != factions2[i].Name {
			t.Errorf("Determinism failed: faction %d names differ (%s vs %s)",
				i, factions1[i].Name, factions2[i].Name)
		}
	}
}

func TestRegisterFactionsWithPoliticsSystem(t *testing.T) {
	fps := systems.NewFactionPoliticsSystem(0.1)

	factions := []*FactionData{
		{
			ID:   "faction_0",
			Name: "TestFaction1",
			Relationships: map[string]int{
				"faction_1": -60, // Hostile
			},
		},
		{
			ID:   "faction_1",
			Name: "TestFaction2",
			Relationships: map[string]int{
				"faction_0": -60, // Hostile
				"faction_2": 70,  // Ally
			},
		},
		{
			ID:   "faction_2",
			Name: "TestFaction3",
			Relationships: map[string]int{
				"faction_1": 70, // Ally
			},
		},
	}

	RegisterFactionsWithPoliticsSystem(fps, factions)

	// Verify relationships were registered
	rel01 := fps.GetRelation("faction_0", "faction_1")
	if rel01 != systems.RelationHostile {
		t.Errorf("Expected hostile relation between faction_0 and faction_1, got %v", rel01)
	}

	rel12 := fps.GetRelation("faction_1", "faction_2")
	if rel12 != systems.RelationAlly {
		t.Errorf("Expected ally relation between faction_1 and faction_2, got %v", rel12)
	}
}
