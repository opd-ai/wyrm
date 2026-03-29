//go:build !noebiten

// This test file requires X11/display because the Venture faction generator
// imports ebiten, which requires GLFW/X11 initialization at import time.
// Run with: xvfb-run go test ./pkg/procgen/adapters/...

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

// =========================================================================
// Quest Adapter Tests
// =========================================================================

func TestQuestAdapter_GenerateQuests(t *testing.T) {
	adapter := NewQuestAdapter()

	tests := []struct {
		name       string
		seed       int64
		genre      string
		count      int
		difficulty float64
		wantErr    bool
	}{
		{"fantasy quests", 12345, "fantasy", 5, 0.5, false},
		{"sci-fi quests", 12345, "sci-fi", 3, 0.8, false},
		{"horror quests", 12345, "horror", 4, 0.3, false},
		{"cyberpunk quests", 12345, "cyberpunk", 6, 0.6, false},
		{"post-apocalyptic quests", 12345, "post-apocalyptic", 2, 1.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quests, err := adapter.GenerateQuests(tt.seed, tt.genre, tt.count, tt.difficulty)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateQuests() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if quests != nil {
				for _, q := range quests {
					if q.ID == "" {
						t.Error("Quest has empty ID")
					}
					if q.Name == "" {
						t.Error("Quest has empty name")
					}
				}
			}
		})
	}
}

func TestQuestAdapter_Deterministic(t *testing.T) {
	adapter := NewQuestAdapter()
	seed := int64(42)
	genre := "fantasy"

	quests1, err := adapter.GenerateQuests(seed, genre, 5, 0.5)
	if err != nil {
		t.Fatalf("First generation failed: %v", err)
	}

	quests2, err := adapter.GenerateQuests(seed, genre, 5, 0.5)
	if err != nil {
		t.Fatalf("Second generation failed: %v", err)
	}

	if len(quests1) != len(quests2) {
		t.Fatalf("Determinism failed: count differs (%d vs %d)", len(quests1), len(quests2))
	}

	for i := range quests1 {
		if quests1[i].Name != quests2[i].Name {
			t.Errorf("Determinism failed: quest %d names differ (%s vs %s)",
				i, quests1[i].Name, quests2[i].Name)
		}
	}
}

func TestSpawnQuestEntity(t *testing.T) {
	world := ecs.NewWorld()
	data := &QuestData{
		ID:   "test_quest_001",
		Name: "Test Quest",
		Type: "kill",
		Objectives: []ObjectiveData{
			{Type: "kill", Description: "Defeat enemy", Target: "enemy", Required: 1, Current: 0},
		},
	}

	entity, err := SpawnQuestEntity(world, data)
	if err != nil {
		t.Fatalf("SpawnQuestEntity failed: %v", err)
	}

	if entity == 0 {
		t.Error("SpawnQuestEntity returned zero entity ID")
	}

	quest, ok := world.GetComponent(entity, "Quest")
	if !ok {
		t.Error("Quest component not found")
	} else {
		t.Logf("Quest: %+v", quest)
	}
}

func TestRegisterQuestWithSystem(t *testing.T) {
	qs := systems.NewQuestSystem()
	data := &QuestData{
		ID:   "test_quest",
		Name: "Test Quest",
		Objectives: []ObjectiveData{
			{Type: "collect", Description: "Gather items", Target: "item", Required: 5},
			{Type: "deliver", Description: "Deliver to NPC", Target: "npc", Required: 1},
		},
	}

	RegisterQuestWithSystem(qs, data)
	// If no panic, registration succeeded
}

// =========================================================================
// Dialog Adapter Tests
// =========================================================================

func TestDialogAdapter_GenerateDialog(t *testing.T) {
	adapter := NewDialogAdapter()

	tests := []struct {
		name      string
		seed      int64
		genre     string
		npcType   string
		sentiment float64
		wantErr   bool
	}{
		{"fantasy merchant", 12345, "fantasy", "merchant", 0.5, false},
		{"sci-fi scientist", 12345, "sci-fi", "scientist", 0.8, false},
		{"horror cultist", 12345, "horror", "cultist", -0.5, false},
		{"cyberpunk hacker", 12345, "cyberpunk", "hacker", 0.0, false},
		{"post-apoc trader", 12345, "post-apocalyptic", "trader", 0.3, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialog, err := adapter.GenerateDialog(tt.seed, tt.genre, tt.npcType, tt.sentiment)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateDialog() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if dialog != nil {
				if len(dialog.Topics) == 0 {
					t.Error("Dialog has no topics")
				}
			}
		})
	}
}

func TestDialogAdapter_Deterministic(t *testing.T) {
	adapter := NewDialogAdapter()
	seed := int64(42)
	genre := "fantasy"

	dialog1, err := adapter.GenerateDialog(seed, genre, "merchant", 0.5)
	if err != nil {
		t.Fatalf("First generation failed: %v", err)
	}

	dialog2, err := adapter.GenerateDialog(seed, genre, "merchant", 0.5)
	if err != nil {
		t.Fatalf("Second generation failed: %v", err)
	}

	if len(dialog1.Topics) != len(dialog2.Topics) {
		t.Fatalf("Determinism failed: topic count differs (%d vs %d)",
			len(dialog1.Topics), len(dialog2.Topics))
	}
}

// =========================================================================
// Magic Adapter Tests
// =========================================================================

func TestMagicAdapter_GenerateSpells(t *testing.T) {
	adapter := NewMagicAdapter()

	tests := []struct {
		name    string
		seed    int64
		genre   string
		count   int
		school  string
		wantErr bool
	}{
		{"fantasy destruction", 12345, "fantasy", 5, "destruction", false},
		{"sci-fi tech", 12345, "sci-fi", 3, "tech", false},
		{"horror dark", 12345, "horror", 4, "dark", false},
		{"cyberpunk netrunner", 12345, "cyberpunk", 6, "hacking", false},
		{"post-apoc mutation", 12345, "post-apocalyptic", 2, "mutation", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spells, err := adapter.GenerateSpells(tt.seed, tt.genre, tt.count, tt.school)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateSpells() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if spells != nil {
				for _, spell := range spells {
					if spell.Name == "" {
						t.Error("Spell has empty name")
					}
					if spell.ManaCost < 0 {
						t.Error("Spell has negative mana cost")
					}
				}
			}
		})
	}
}

func TestMagicAdapter_Deterministic(t *testing.T) {
	adapter := NewMagicAdapter()
	seed := int64(42)

	spells1, err := adapter.GenerateSpells(seed, "fantasy", 5, "destruction")
	if err != nil {
		t.Fatalf("First generation failed: %v", err)
	}

	spells2, err := adapter.GenerateSpells(seed, "fantasy", 5, "destruction")
	if err != nil {
		t.Fatalf("Second generation failed: %v", err)
	}

	if len(spells1) != len(spells2) {
		t.Fatalf("Determinism failed: count differs (%d vs %d)", len(spells1), len(spells2))
	}

	for i := range spells1 {
		if spells1[i].Name != spells2[i].Name {
			t.Errorf("Determinism failed: spell %d names differ (%s vs %s)",
				i, spells1[i].Name, spells2[i].Name)
		}
	}
}

// =========================================================================
// Skills Adapter Tests
// =========================================================================

func TestSkillsAdapter_GenerateSkillTree(t *testing.T) {
	adapter := NewSkillsAdapter()

	tests := []struct {
		name    string
		seed    int64
		genre   string
		school  string
		wantErr bool
	}{
		{"fantasy combat", 12345, "fantasy", "combat", false},
		{"sci-fi engineering", 12345, "sci-fi", "engineering", false},
		{"horror survival", 12345, "horror", "survival", false},
		{"cyberpunk hacking", 12345, "cyberpunk", "hacking", false},
		{"post-apoc scavenge", 12345, "post-apocalyptic", "scavenging", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, err := adapter.GenerateSkillTree(tt.seed, tt.genre, tt.school)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateSkillTree() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tree != nil && len(tree.Skills) > 0 {
				for _, skill := range tree.Skills {
					if skill.Name == "" {
						t.Error("Skill has empty name")
					}
				}
			}
		})
	}
}

// =========================================================================
// Recipe Adapter Tests
// =========================================================================

func TestRecipeAdapter_GenerateRecipes(t *testing.T) {
	adapter := NewRecipeAdapter()

	tests := []struct {
		name     string
		seed     int64
		genre    string
		category string
		count    int
		wantErr  bool
	}{
		{"fantasy weapons", 12345, "fantasy", "weapons", 5, false},
		{"sci-fi tech", 12345, "sci-fi", "tech", 3, false},
		{"horror potions", 12345, "horror", "potions", 4, false},
		{"cyberpunk cyberware", 12345, "cyberpunk", "cyberware", 6, false},
		{"post-apoc survival", 12345, "post-apocalyptic", "survival", 2, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recipes, err := adapter.GenerateRecipes(tt.seed, tt.genre, tt.category, tt.count)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateRecipes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if recipes != nil {
				for _, r := range recipes {
					if r.Name == "" {
						t.Error("Recipe has empty name")
					}
					if len(r.Ingredients) == 0 {
						t.Error("Recipe has no ingredients")
					}
				}
			}
		})
	}
}

func TestCanCraft(t *testing.T) {
	recipe := &RecipeData{
		Name: "Test Weapon",
		Ingredients: []IngredientData{
			{Name: "iron", Required: 3},
			{Name: "wood", Required: 1},
		},
	}

	inventory := map[string]int{
		"iron": 5,
		"wood": 2,
	}

	if !CanCraft(recipe, inventory) {
		t.Error("CanCraft should return true with sufficient materials")
	}

	inventory["iron"] = 1
	if CanCraft(recipe, inventory) {
		t.Error("CanCraft should return false with insufficient materials")
	}
}

// =========================================================================
// Vehicle Adapter Tests
// =========================================================================

func TestVehicleAdapter_GenerateVehicles(t *testing.T) {
	adapter := NewVehicleAdapter()

	tests := []struct {
		name        string
		seed        int64
		genre       string
		vehicleType string
		count       int
		wantErr     bool
	}{
		{"fantasy horses", 12345, "fantasy", "mount", 3, false},
		{"sci-fi hovercrafts", 12345, "sci-fi", "hovercraft", 2, false},
		{"horror carriages", 12345, "horror", "carriage", 2, false},
		{"cyberpunk bikes", 12345, "cyberpunk", "motorcycle", 4, false},
		{"post-apoc buggies", 12345, "post-apocalyptic", "buggy", 3, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vehicles, err := adapter.GenerateVehicles(tt.seed, tt.genre, tt.vehicleType, tt.count)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateVehicles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if vehicles != nil {
				for _, v := range vehicles {
					if v.Name == "" {
						t.Error("Vehicle has empty name")
					}
					if v.MaxSpeed <= 0 {
						t.Error("Vehicle has non-positive max speed")
					}
				}
			}
		})
	}
}

func TestSpawnVehicleEntity(t *testing.T) {
	world := ecs.NewWorld()
	data := &VehicleData{
		Name:         "Test Car",
		Type:         "buggy",
		MaxSpeed:     100.0,
		Acceleration: 20.0,
		FuelCapacity: 50.0,
		Health:       200,
	}

	entity, err := SpawnVehicleEntity(world, data, 10.0, 20.0)
	if err != nil {
		t.Fatalf("SpawnVehicleEntity failed: %v", err)
	}

	if entity == 0 {
		t.Error("SpawnVehicleEntity returned zero entity ID")
	}

	vehicle, ok := world.GetComponent(entity, "Vehicle")
	if !ok {
		t.Error("Vehicle component not found")
	} else {
		t.Logf("Vehicle: %+v", vehicle)
	}

	pos, ok := world.GetComponent(entity, "Position")
	if !ok {
		t.Error("Position component not found")
	} else {
		t.Logf("Position: %+v", pos)
	}
}

// =========================================================================
// Building Adapter Tests
// =========================================================================

func TestBuildingAdapter_GenerateBuildings(t *testing.T) {
	adapter := NewBuildingAdapter()

	tests := []struct {
		name         string
		seed         int64
		genre        string
		buildingType string
		count        int
		wantErr      bool
	}{
		{"fantasy taverns", 12345, "fantasy", "tavern", 3, false},
		{"sci-fi labs", 12345, "sci-fi", "lab", 2, false},
		{"horror mansions", 12345, "horror", "mansion", 2, false},
		{"cyberpunk clubs", 12345, "cyberpunk", "club", 4, false},
		{"post-apoc shelters", 12345, "post-apocalyptic", "shelter", 3, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buildings, err := adapter.GenerateBuildings(tt.seed, tt.genre, tt.buildingType, tt.count)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateBuildings() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if buildings != nil {
				for _, b := range buildings {
					if b.Name == "" {
						t.Error("Building has empty name")
					}
				}
			}
		})
	}
}

// =========================================================================
// Item Adapter Tests
// =========================================================================

func TestItemAdapter_GenerateItems(t *testing.T) {
	adapter := NewItemAdapter()

	tests := []struct {
		name     string
		seed     int64
		genre    string
		itemType string
		quality  float64
		count    int
		wantErr  bool
	}{
		{"fantasy weapons", 12345, "fantasy", "weapon", 0.5, 5, false},
		{"sci-fi gadgets", 12345, "sci-fi", "gadget", 0.8, 3, false},
		{"horror artifacts", 12345, "horror", "artifact", 0.3, 4, false},
		{"cyberpunk implants", 12345, "cyberpunk", "implant", 0.6, 6, false},
		{"post-apoc scrap", 12345, "post-apocalyptic", "scrap", 1.0, 2, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, err := adapter.GenerateItems(tt.seed, tt.genre, tt.itemType, tt.quality, tt.count)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateItems() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if items != nil {
				for _, item := range items {
					if item.Name == "" {
						t.Error("Item has empty name")
					}
				}
			}
		})
	}
}

func TestItemAdapter_Deterministic(t *testing.T) {
	adapter := NewItemAdapter()
	seed := int64(42)

	items1, err := adapter.GenerateItems(seed, "fantasy", "weapon", 0.5, 5)
	if err != nil {
		t.Fatalf("First generation failed: %v", err)
	}

	items2, err := adapter.GenerateItems(seed, "fantasy", "weapon", 0.5, 5)
	if err != nil {
		t.Fatalf("Second generation failed: %v", err)
	}

	if len(items1) != len(items2) {
		t.Fatalf("Determinism failed: count differs (%d vs %d)", len(items1), len(items2))
	}

	for i := range items1 {
		if items1[i].Name != items2[i].Name {
			t.Errorf("Determinism failed: item %d names differ (%s vs %s)",
				i, items1[i].Name, items2[i].Name)
		}
	}
}

// =========================================================================
// Puzzle Adapter Tests
// =========================================================================

func TestPuzzleAdapter_GeneratePuzzles(t *testing.T) {
	adapter := NewPuzzleAdapter()

	tests := []struct {
		name       string
		seed       int64
		genre      string
		difficulty float64
		count      int
		wantErr    bool
	}{
		{"fantasy puzzles", 12345, "fantasy", 0.5, 3, false},
		{"sci-fi puzzles", 12345, "sci-fi", 0.8, 2, false},
		{"horror puzzles", 12345, "horror", 0.3, 4, false},
		{"cyberpunk puzzles", 12345, "cyberpunk", 0.6, 3, false},
		{"post-apoc puzzles", 12345, "post-apocalyptic", 1.0, 2, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			puzzles, err := adapter.GeneratePuzzles(tt.seed, tt.genre, tt.difficulty, tt.count)
			if (err != nil) != tt.wantErr {
				t.Errorf("GeneratePuzzles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if puzzles != nil {
				for _, p := range puzzles {
					if p.Type == "" {
						t.Error("Puzzle has empty type")
					}
				}
			}
		})
	}
}

// =========================================================================
// Environment Adapter Tests
// =========================================================================

func TestEnvironmentAdapter_GenerateEnvironment(t *testing.T) {
	adapter := NewEnvironmentAdapter()

	tests := []struct {
		name    string
		seed    int64
		genre   string
		biome   string
		wantErr bool
	}{
		{"fantasy forest", 12345, "fantasy", "forest", false},
		{"sci-fi crater", 12345, "sci-fi", "crater", false},
		{"horror swamp", 12345, "horror", "swamp", false},
		{"cyberpunk urban", 12345, "cyberpunk", "urban", false},
		{"post-apoc wasteland", 12345, "post-apocalyptic", "wasteland", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env, err := adapter.GenerateEnvironment(tt.seed, tt.genre, tt.biome, 512)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateEnvironment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if env != nil {
				if len(env.Details) == 0 {
					t.Log("Environment has no details (may be expected)")
				}
			}
		})
	}
}

// =========================================================================
// Furniture Adapter Tests
// =========================================================================

func TestFurnitureAdapter_GenerateFurniture(t *testing.T) {
	adapter := NewFurnitureAdapter()

	tests := []struct {
		name     string
		seed     int64
		genre    string
		roomType string
		count    int
		wantErr  bool
	}{
		{"fantasy tavern furniture", 12345, "fantasy", "tavern", 10, false},
		{"sci-fi lab furniture", 12345, "sci-fi", "lab", 8, false},
		{"horror bedroom furniture", 12345, "horror", "bedroom", 6, false},
		{"cyberpunk office furniture", 12345, "cyberpunk", "office", 12, false},
		{"post-apoc shelter furniture", 12345, "post-apocalyptic", "shelter", 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			furniture, err := adapter.GenerateFurniture(tt.seed, tt.genre, tt.roomType, tt.count)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateFurniture() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if furniture != nil {
				for _, f := range furniture {
					if f.Name == "" {
						t.Error("Furniture has empty name")
					}
				}
			}
		})
	}
}

// =========================================================================
// Narrative Adapter Tests
// =========================================================================

func TestNarrativeAdapter_GenerateStoryArc(t *testing.T) {
	adapter := NewNarrativeAdapter()

	tests := []struct {
		name    string
		seed    int64
		genre   string
		arcType string
		wantErr bool
	}{
		{"fantasy hero arc", 12345, "fantasy", "hero", false},
		{"sci-fi mystery arc", 12345, "sci-fi", "mystery", false},
		{"horror survival arc", 12345, "horror", "survival", false},
		{"cyberpunk heist arc", 12345, "cyberpunk", "heist", false},
		{"post-apoc redemption arc", 12345, "post-apocalyptic", "redemption", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arc, err := adapter.GenerateStoryArc(tt.seed, tt.genre, tt.arcType)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateStoryArc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if arc != nil {
				if arc.Name == "" {
					t.Error("Story arc has empty name")
				}
			}
		})
	}
}

// =========================================================================
// Integration Tests
// =========================================================================

func TestEntityAdapter_GenerateAndSpawnNPCs(t *testing.T) {
	adapter := NewEntityAdapter()
	world := ecs.NewWorld()

	entities, err := adapter.GenerateAndSpawnNPCs(world, 12345, "fantasy", "faction_test", 5, 100.0, 100.0, 50.0)
	if err != nil {
		t.Fatalf("GenerateAndSpawnNPCs failed: %v", err)
	}

	if len(entities) == 0 {
		t.Error("No entities spawned")
	}

	for _, e := range entities {
		// Verify each entity has required components
		_, hasPos := world.GetComponent(e, "Position")
		_, hasHealth := world.GetComponent(e, "Health")
		_, hasFaction := world.GetComponent(e, "Faction")

		if !hasPos {
			t.Errorf("Entity %d missing Position component", e)
		}
		if !hasHealth {
			t.Errorf("Entity %d missing Health component", e)
		}
		if !hasFaction {
			t.Errorf("Entity %d missing Faction component", e)
		}
	}
}

func TestQuestAdapter_GenerateAndSpawnQuests(t *testing.T) {
	adapter := NewQuestAdapter()
	world := ecs.NewWorld()
	qs := systems.NewQuestSystem()

	entities, err := adapter.GenerateAndSpawnQuests(world, qs, 12345, "fantasy", 3)
	if err != nil {
		t.Fatalf("GenerateAndSpawnQuests failed: %v", err)
	}

	if len(entities) == 0 {
		t.Error("No quest entities spawned")
	}

	for _, e := range entities {
		_, hasQuest := world.GetComponent(e, "Quest")
		if !hasQuest {
			t.Errorf("Entity %d missing Quest component", e)
		}
	}
}

// =========================================================================
// Genre Mapping Tests
// =========================================================================

func TestMapGenreID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"fantasy", "fantasy"},
		{"sci-fi", "sci-fi"},
		{"horror", "horror"},
		{"cyberpunk", "cyberpunk"},
		{"post-apocalyptic", "post-apocalyptic"},
		{"unknown", "fantasy"}, // Should default to fantasy
		{"", "fantasy"},        // Empty should default to fantasy
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := mapGenreID(tt.input)
			if result != tt.expected {
				t.Errorf("mapGenreID(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}
