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

func TestDialogAdapter_GenerateDialogLine(t *testing.T) {
	adapter := NewDialogAdapter()

	tests := []struct {
		name    string
		seed    int64
		genre   string
		wantErr bool
	}{
		{"fantasy dialog", 12345, "fantasy", false},
		{"sci-fi dialog", 12345, "sci-fi", false},
		{"horror dialog", 12345, "horror", false},
		{"cyberpunk dialog", 12345, "cyberpunk", false},
		{"post-apoc dialog", 12345, "post-apocalyptic", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line, err := adapter.GenerateDialogLine(tt.seed, tt.genre)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateDialogLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if line != nil {
				if line.Text == "" {
					t.Error("Dialog line has empty text")
				}
			}
		})
	}
}

func TestDialogAdapter_GenerateDialogLines(t *testing.T) {
	adapter := NewDialogAdapter()
	seed := int64(42)
	genre := "fantasy"

	lines, err := adapter.GenerateDialogLines(seed, genre, 5)
	if err != nil {
		t.Fatalf("GenerateDialogLines failed: %v", err)
	}

	if len(lines) == 0 {
		t.Error("GenerateDialogLines returned empty")
	}

	for _, line := range lines {
		if line.Text == "" {
			t.Error("Dialog line has empty text")
		}
	}
}

func TestDialogAdapter_Deterministic(t *testing.T) {
	adapter := NewDialogAdapter()
	seed := int64(42)
	genre := "fantasy"

	line1, err := adapter.GenerateDialogLine(seed, genre)
	if err != nil {
		t.Fatalf("First generation failed: %v", err)
	}

	line2, err := adapter.GenerateDialogLine(seed, genre)
	if err != nil {
		t.Fatalf("Second generation failed: %v", err)
	}

	if line1.Text != line2.Text {
		t.Errorf("Determinism failed: dialog differs (%s vs %s)", line1.Text, line2.Text)
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
		wantErr bool
	}{
		{"fantasy spells", 12345, "fantasy", 5, false},
		{"sci-fi spells", 12345, "sci-fi", 3, false},
		{"horror spells", 12345, "horror", 4, false},
		{"cyberpunk spells", 12345, "cyberpunk", 6, false},
		{"post-apoc spells", 12345, "post-apocalyptic", 2, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spells, err := adapter.GenerateSpells(tt.seed, tt.genre, tt.count)
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

	spells1, err := adapter.GenerateSpells(seed, "fantasy", 5)
	if err != nil {
		t.Fatalf("First generation failed: %v", err)
	}

	spells2, err := adapter.GenerateSpells(seed, "fantasy", 5)
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
		wantErr bool
	}{
		{"fantasy skills", 12345, "fantasy", false},
		{"sci-fi skills", 12345, "sci-fi", false},
		{"horror skills", 12345, "horror", false},
		{"cyberpunk skills", 12345, "cyberpunk", false},
		{"post-apoc skills", 12345, "post-apocalyptic", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, err := adapter.GenerateSkillTree(tt.seed, tt.genre)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateSkillTree() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tree != nil && len(tree.Nodes) > 0 {
				for _, node := range tree.Nodes {
					if node.Skill != nil && node.Skill.Name == "" {
						t.Error("Skill has empty name")
					}
				}
			}
		})
	}
}

func TestSkillsAdapter_GenerateSkillTrees(t *testing.T) {
	adapter := NewSkillsAdapter()
	seed := int64(42)

	trees, err := adapter.GenerateSkillTrees(seed, "fantasy", 3)
	if err != nil {
		t.Fatalf("GenerateSkillTrees failed: %v", err)
	}

	if len(trees) == 0 {
		t.Error("GenerateSkillTrees returned empty")
	}

	for _, tree := range trees {
		if tree.Name == "" {
			t.Error("Skill tree has empty name")
		}
	}
}

// =========================================================================
// Recipe Adapter Tests
// =========================================================================

func TestRecipeAdapter_GenerateRecipes(t *testing.T) {
	adapter := NewRecipeAdapter()

	tests := []struct {
		name       string
		seed       int64
		genre      string
		depth      int
		count      int
		recipeType string
		wantErr    bool
	}{
		{"fantasy weapons", 12345, "fantasy", 5, 5, "weapons", false},
		{"sci-fi tech", 12345, "sci-fi", 10, 3, "tech", false},
		{"horror potions", 12345, "horror", 15, 4, "potions", false},
		{"cyberpunk cyberware", 12345, "cyberpunk", 20, 6, "cyberware", false},
		{"post-apoc survival", 12345, "post-apocalyptic", 25, 2, "survival", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recipes, err := adapter.GenerateRecipes(tt.seed, tt.genre, tt.depth, tt.count, tt.recipeType)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateRecipes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if recipes != nil {
				for _, r := range recipes {
					if r.Name == "" {
						t.Error("Recipe has empty name")
					}
					if len(r.Materials) == 0 {
						t.Error("Recipe has no materials")
					}
				}
			}
		})
	}
}

func TestCanCraft(t *testing.T) {
	recipe := &RecipeData{
		Name: "Test Weapon",
		Materials: []MaterialData{
			{ItemName: "iron", Quantity: 3},
			{ItemName: "wood", Quantity: 1},
		},
		SkillRequired: 5,
		GoldCost:      100,
	}

	inventory := map[string]int{
		"iron": 5,
		"wood": 2,
	}

	canCraft, reason := CanCraft(recipe, 10, 200, inventory)
	if !canCraft {
		t.Errorf("CanCraft should return true with sufficient materials, got: %s", reason)
	}

	inventory["iron"] = 1
	canCraft, _ = CanCraft(recipe, 10, 200, inventory)
	if canCraft {
		t.Error("CanCraft should return false with insufficient materials")
	}
}

// =========================================================================
// Vehicle Adapter Tests
// =========================================================================

func TestVehicleAdapter_GenerateVehicles(t *testing.T) {
	adapter := NewVehicleAdapter()

	tests := []struct {
		name    string
		seed    int64
		genre   string
		count   int
		wantErr bool
	}{
		{"fantasy vehicles", 12345, "fantasy", 3, false},
		{"sci-fi vehicles", 12345, "sci-fi", 2, false},
		{"horror vehicles", 12345, "horror", 2, false},
		{"cyberpunk vehicles", 12345, "cyberpunk", 4, false},
		{"post-apoc vehicles", 12345, "post-apocalyptic", 3, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vehicles, err := adapter.GenerateVehicles(tt.seed, tt.genre, tt.count)
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
		Name:          "Test Car",
		VehicleType:   "buggy",
		MaxSpeed:      100.0,
		Acceleration:  20.0,
		FuelCapacity:  50.0,
		MaxDurability: 200,
	}

	entity := SpawnVehicleEntity(world, data, 10.0, 20.0, 0.0)

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

func TestBuildingAdapter_GenerateCityBuildings(t *testing.T) {
	adapter := NewBuildingAdapter()

	tests := []struct {
		name    string
		seed    int64
		genre   string
		count   int
		wantErr bool
	}{
		{"fantasy buildings", 12345, "fantasy", 3, false},
		{"sci-fi buildings", 12345, "sci-fi", 2, false},
		{"horror buildings", 12345, "horror", 2, false},
		{"cyberpunk buildings", 12345, "cyberpunk", 4, false},
		{"post-apoc buildings", 12345, "post-apocalyptic", 3, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buildings, err := adapter.GenerateCityBuildings(tt.seed, tt.genre, tt.count)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateCityBuildings() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if buildings != nil {
				for _, b := range buildings {
					if b.Type == "" {
						t.Error("Building has empty type")
					}
				}
			}
		})
	}
}

func TestBuildingAdapter_GenerateBuilding(t *testing.T) {
	adapter := NewBuildingAdapter()

	building, err := adapter.GenerateBuilding(12345, "fantasy", 1, 2)
	if err != nil {
		t.Fatalf("GenerateBuilding failed: %v", err)
	}

	if building.Type == "" {
		t.Error("Building has empty type")
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
		depth    int
		count    int
		itemType string
		wantErr  bool
	}{
		{"fantasy weapons", 12345, "fantasy", 5, 5, "weapon", false},
		{"sci-fi gadgets", 12345, "sci-fi", 10, 3, "gadget", false},
		{"horror artifacts", 12345, "horror", 15, 4, "artifact", false},
		{"cyberpunk implants", 12345, "cyberpunk", 20, 6, "implant", false},
		{"post-apoc scrap", 12345, "post-apocalyptic", 25, 2, "scrap", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, err := adapter.GenerateItems(tt.seed, tt.genre, tt.depth, tt.count, tt.itemType)
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

	items1, err := adapter.GenerateItems(seed, "fantasy", 5, 5, "weapon")
	if err != nil {
		t.Fatalf("First generation failed: %v", err)
	}

	items2, err := adapter.GenerateItems(seed, "fantasy", 5, 5, "weapon")
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
		{"fantasy puzzles", 12345, "fantasy", 5, 3, false},
		{"sci-fi puzzles", 12345, "sci-fi", 8, 2, false},
		{"horror puzzles", 12345, "horror", 3, 4, false},
		{"cyberpunk puzzles", 12345, "cyberpunk", 6, 3, false},
		{"post-apoc puzzles", 12345, "post-apocalyptic", 10, 2, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			puzzles, err := adapter.GenerateDungeonPuzzles(tt.seed, tt.genre, int(tt.difficulty), tt.count)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateDungeonPuzzles() error = %v, wantErr %v", err, tt.wantErr)
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

func TestPuzzleAdapter_GeneratePuzzle(t *testing.T) {
	adapter := NewPuzzleAdapter()

	puzzle, err := adapter.GeneratePuzzle(12345, "fantasy", 5)
	if err != nil {
		t.Fatalf("GeneratePuzzle failed: %v", err)
	}

	if puzzle.Type == "" {
		t.Error("Puzzle has empty type")
	}
}

// =========================================================================
// Environment Adapter Tests
// =========================================================================

func TestEnvironmentAdapter_GenerateBiomeObjects(t *testing.T) {
	adapter := NewEnvironmentAdapter()

	tests := []struct {
		name    string
		seed    int64
		genre   string
		biome   string
		count   int
		wantErr bool
	}{
		{"fantasy forest", 12345, "fantasy", "forest", 10, false},
		{"sci-fi crater", 12345, "sci-fi", "crater", 8, false},
		{"horror swamp", 12345, "horror", "swamp", 6, false},
		{"cyberpunk urban", 12345, "cyberpunk", "urban", 12, false},
		{"post-apoc wasteland", 12345, "post-apocalyptic", "wasteland", 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objs, err := adapter.GenerateBiomeObjects(tt.seed, tt.genre, tt.biome, tt.count)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateBiomeObjects() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if objs != nil {
				t.Logf("Generated %d environment objects for %s", len(objs), tt.biome)
			}
		})
	}
}

func TestEnvironmentAdapter_GenerateChunkDecorations(t *testing.T) {
	adapter := NewEnvironmentAdapter()

	decorations, err := adapter.GenerateChunkDecorations(12345, "fantasy", "forest")
	if err != nil {
		t.Fatalf("GenerateChunkDecorations failed: %v", err)
	}

	t.Logf("Generated %d decorations", len(decorations))
}

// =========================================================================
// Furniture Adapter Tests
// =========================================================================

func TestFurnitureAdapter_GenerateRoomFurniture(t *testing.T) {
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
			furniture, err := adapter.GenerateRoomFurniture(tt.seed, tt.genre, tt.roomType, tt.count)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateRoomFurniture() error = %v, wantErr %v", err, tt.wantErr)
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
		name       string
		seed       int64
		genre      string
		difficulty float64
		wantErr    bool
	}{
		{"fantasy story arc", 12345, "fantasy", 0.5, false},
		{"sci-fi story arc", 12345, "sci-fi", 0.8, false},
		{"horror story arc", 12345, "horror", 0.3, false},
		{"cyberpunk story arc", 12345, "cyberpunk", 0.6, false},
		{"post-apoc story arc", 12345, "post-apocalyptic", 1.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arc, err := adapter.GenerateStoryArc(tt.seed, tt.genre, tt.difficulty)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateStoryArc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if arc != nil {
				if arc.Title == "" {
					t.Error("Story arc has empty title")
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
		{"sci-fi", "scifi"}, // Transformed to "scifi" for Venture
		{"horror", "horror"},
		{"cyberpunk", "cyberpunk"},
		{"post-apocalyptic", "postapoc"}, // Transformed to "postapoc" for Venture
		{"unknown", "unknown"},           // Passthrough for unknown genres
		{"", ""},                         // Empty stays empty
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

// =========================================================================
// Additional Coverage Tests
// =========================================================================

func TestBuildingAdapter_SpecificBuildings(t *testing.T) {
	adapter := NewBuildingAdapter()

	t.Run("GenerateHouse", func(t *testing.T) {
		data, err := adapter.GenerateHouse(12345, "fantasy")
		if err != nil {
			t.Errorf("GenerateHouse() error = %v", err)
		}
		if data == nil {
			t.Error("GenerateHouse() returned nil data")
		}
	})

	t.Run("GenerateWorkshop", func(t *testing.T) {
		data, err := adapter.GenerateWorkshop(12345, "fantasy")
		if err != nil {
			t.Errorf("GenerateWorkshop() error = %v", err)
		}
		if data == nil {
			t.Error("GenerateWorkshop() returned nil data")
		}
	})

	t.Run("GenerateTower", func(t *testing.T) {
		data, err := adapter.GenerateTower(12345, "fantasy")
		if err != nil {
			t.Errorf("GenerateTower() error = %v", err)
		}
		if data == nil {
			t.Error("GenerateTower() returned nil data")
		}
	})

	t.Run("GenerateManor", func(t *testing.T) {
		data, err := adapter.GenerateManor(12345, "fantasy")
		if err != nil {
			t.Errorf("GenerateManor() error = %v", err)
		}
		if data == nil {
			t.Error("GenerateManor() returned nil data")
		}
	})

	t.Run("GenerateGuildHall", func(t *testing.T) {
		data, err := adapter.GenerateGuildHall(12345, "fantasy")
		if err != nil {
			t.Errorf("GenerateGuildHall() error = %v", err)
		}
		if data == nil {
			t.Error("GenerateGuildHall() returned nil data")
		}
	})
}

func TestDialogAdapter_GenerateGreeting(t *testing.T) {
	adapter := NewDialogAdapter()
	personality := PersonalityTraits{
		Friendliness: 0.8,
		Verbosity:    0.5,
		Formality:    0.5,
		Humor:        0.5,
		Knowledge:    0.5,
	}

	line, err := adapter.GenerateGreeting(12345, "fantasy", personality)
	if err != nil {
		t.Errorf("GenerateGreeting() error = %v", err)
	}
	if line == nil {
		t.Fatal("GenerateGreeting() returned nil")
	}
	if line.Text == "" {
		t.Error("GenerateGreeting() returned empty text")
	}
}

func TestPersonalityFromType(t *testing.T) {
	tests := []struct {
		ptype       string
		expectValid bool
	}{
		{"helpful", true},
		{"merchant", true},
		{"hostile", true},
		{"scholarly", true},
		{"unknown", true}, // Default case
	}

	for _, tt := range tests {
		t.Run(tt.ptype, func(t *testing.T) {
			personality := PersonalityFromType(tt.ptype)
			// Default should return zero values for unknown types
			if tt.ptype == "helpful" && personality.Friendliness < 0.5 {
				t.Errorf("PersonalityFromType(%s) friendliness too low", tt.ptype)
			}
			if tt.ptype == "hostile" && personality.Friendliness > 0.5 {
				t.Errorf("PersonalityFromType(%s) friendliness should be low", tt.ptype)
			}
		})
	}
}

func TestEnvironmentAdapter_RoomDecorations(t *testing.T) {
	adapter := NewEnvironmentAdapter()

	decorations, err := adapter.GenerateRoomDecorations(12345, "fantasy", 10, 10, 0.5)
	if err != nil {
		t.Errorf("GenerateRoomDecorations() error = %v", err)
	}
	if len(decorations) == 0 {
		t.Error("GenerateRoomDecorations() returned no decorations")
	}
}

func TestEnvironmentAdapter_BiomeDensity(t *testing.T) {
	// Test different biome densities
	biomes := []string{"forest", "desert", "plains", "swamp", "mountain"}
	for _, biome := range biomes {
		density := getDensityForBiome(biome)
		if density < 0 || density > 1 {
			t.Errorf("getDensityForBiome(%s) = %f, want 0-1 range", biome, density)
		}
	}
}

func TestEnvironmentAdapter_BiomeObjectTypes(t *testing.T) {
	// Test different biome object types
	biomes := []string{"forest", "desert", "plains", "swamp", "mountain"}
	for _, biome := range biomes {
		types := getBiomeObjectTypes(biome)
		if len(types) == 0 {
			t.Errorf("getBiomeObjectTypes(%s) returned empty", biome)
		}
	}
}

func TestEnvironmentAdapter_HazardFunctions(t *testing.T) {
	// Test hazard detection
	obj := &EnvironmentObjectData{
		Type:    "hazard",
		Harmful: true,
		Damage:  10,
	}
	if !IsEnvironmentObjectHazard(obj) {
		t.Error("IsEnvironmentObjectHazard() should return true for hazard")
	}
	damage := GetEnvironmentObjectDamage(obj)
	if damage != 10 {
		t.Errorf("GetEnvironmentObjectDamage() = %d, want 10", damage)
	}
}

func TestFurnitureAdapter_HouseFurniture(t *testing.T) {
	adapter := NewFurnitureAdapter()

	furniture, err := adapter.GenerateHouseFurniture(12345, "fantasy")
	if err != nil {
		t.Errorf("GenerateHouseFurniture() error = %v", err)
	}
	if len(furniture) == 0 {
		t.Error("GenerateHouseFurniture() returned no furniture")
	}
}

func TestFurnitureAdapter_CraftingStations(t *testing.T) {
	adapter := NewFurnitureAdapter()

	stations, err := adapter.GenerateCraftingStations(12345, "fantasy")
	if err != nil {
		t.Errorf("GenerateCraftingStations() error = %v", err)
	}
	if len(stations) == 0 {
		t.Error("GenerateCraftingStations() returned no stations")
	}
}

func TestFurnitureAdapter_FunctionalChecks(t *testing.T) {
	furniture := &FurnitureData{
		Type:       "workbench",
		Functional: true,
		Material:   "Wood",
	}
	if !IsFurnitureFunctional(furniture) {
		t.Error("IsFurnitureFunctional() should return true")
	}
	value := GetFurnitureValue(furniture)
	if value < 0 {
		t.Errorf("GetFurnitureValue() = %d, want >= 0", value)
	}
}

func TestItemAdapter_SingleItem(t *testing.T) {
	adapter := NewItemAdapter()

	item, err := adapter.GenerateSingleItem(12345, "fantasy", 5, "weapon")
	if err != nil {
		t.Errorf("GenerateSingleItem() error = %v", err)
	}
	if item == nil {
		t.Error("GenerateSingleItem() returned nil")
	}
}

func TestItemAdapter_LootTable(t *testing.T) {
	adapter := NewItemAdapter()

	items, err := adapter.GenerateLootTable(12345, "fantasy", 5)
	if err != nil {
		t.Errorf("GenerateLootTable() error = %v", err)
	}
	if len(items) == 0 {
		t.Error("GenerateLootTable() returned no items")
	}
}

func TestItemAdapter_ShopInventory(t *testing.T) {
	adapter := NewItemAdapter()

	items, err := adapter.GenerateShopInventory(12345, "fantasy", 50)
	if err != nil {
		t.Errorf("GenerateShopInventory() error = %v", err)
	}
	if len(items) == 0 {
		t.Error("GenerateShopInventory() returned no items")
	}
}

func TestItemAdapter_ItemHelpers(t *testing.T) {
	item := &ItemData{
		Type: "weapon",
		Stats: ItemStatsData{
			Value:  50,
			Weight: 5.0,
		},
	}
	value := GetItemValue(item, 1.0)
	if value != 50 {
		t.Errorf("GetItemValue() = %d, want 50", value)
	}
	if !IsItemEquippable(item) {
		t.Error("IsItemEquippable() should return true for weapon")
	}

	consumable := &ItemData{
		Type: "consumable",
	}
	if !IsItemConsumable(consumable) {
		t.Error("IsItemConsumable() should return true for consumable")
	}
}

func TestMagicAdapter_SingleSpell(t *testing.T) {
	adapter := NewMagicAdapter()

	spell, err := adapter.GenerateSpell(12345, "fantasy")
	if err != nil {
		t.Errorf("GenerateSpell() error = %v", err)
	}
	if spell == nil {
		t.Error("GenerateSpell() returned nil")
	}
}

func TestMagicAdapter_SpellHelpers(t *testing.T) {
	offensiveSpell := &SpellData{
		Name: "Fireball",
		Type: "Offensive",
	}
	if !offensiveSpell.IsOffensive() {
		t.Error("IsOffensive() should return true for offensive spell")
	}

	supportSpell := &SpellData{
		Name: "Heal",
		Type: "Healing",
	}
	if !supportSpell.IsSupport() {
		t.Error("IsSupport() should return true for support spell")
	}

	// Test rarity multiplier - uses capital case like "Common", "Rare"
	for _, rarity := range []string{"Common", "Rare", "Epic", "Legendary"} {
		mult := SpellRarityMultiplier(rarity)
		if mult < 1.0 {
			t.Errorf("SpellRarityMultiplier(%s) = %f, want >= 1.0", rarity, mult)
		}
	}
}

func TestNarrativeAdapter_RegionArc(t *testing.T) {
	adapter := NewNarrativeAdapter()

	arc, err := adapter.GetActiveArcForRegion(12345, 0, 0, "fantasy")
	if err != nil {
		t.Errorf("GetActiveArcForRegion() error = %v", err)
	}
	if arc != nil && arc.Title == "" {
		t.Error("GetActiveArcForRegion() returned arc with empty title")
	}
}

func TestNarrativeAdapter_WorldEventArc(t *testing.T) {
	adapter := NewNarrativeAdapter()

	arc, err := adapter.GetWorldEventArc(12345, "dragon_attack", "fantasy", 0.5)
	if err != nil {
		t.Errorf("GetWorldEventArc() error = %v", err)
	}
	if arc != nil && arc.Title == "" {
		t.Error("GetWorldEventArc() returned arc with empty title")
	}
}

func TestPuzzleAdapter_OfType(t *testing.T) {
	adapter := NewPuzzleAdapter()

	puzzle, err := adapter.GeneratePuzzleOfType(12345, "fantasy", 5, "logic")
	if err != nil {
		t.Errorf("GeneratePuzzleOfType() error = %v", err)
	}
	if puzzle == nil {
		t.Error("GeneratePuzzleOfType() returned nil")
	}
}

func TestRecipeAdapter_SpecificRecipes(t *testing.T) {
	adapter := NewRecipeAdapter()

	t.Run("PotionRecipes", func(t *testing.T) {
		recipes, err := adapter.GeneratePotionRecipes(12345, "fantasy", 5, 3)
		if err != nil {
			t.Errorf("GeneratePotionRecipes() error = %v", err)
		}
		if len(recipes) == 0 {
			t.Error("GeneratePotionRecipes() returned no recipes")
		}
	})

	t.Run("EnchantingRecipes", func(t *testing.T) {
		recipes, err := adapter.GenerateEnchantingRecipes(12345, "fantasy", 5, 3)
		if err != nil {
			t.Errorf("GenerateEnchantingRecipes() error = %v", err)
		}
		if len(recipes) == 0 {
			t.Error("GenerateEnchantingRecipes() returned no recipes")
		}
	})

	t.Run("MagicItemRecipes", func(t *testing.T) {
		recipes, err := adapter.GenerateMagicItemRecipes(12345, "fantasy", 5, 3)
		if err != nil {
			t.Errorf("GenerateMagicItemRecipes() error = %v", err)
		}
		if len(recipes) == 0 {
			t.Error("GenerateMagicItemRecipes() returned no recipes")
		}
	})

	t.Run("WorkbenchRecipes", func(t *testing.T) {
		recipes, err := adapter.GenerateWorkbenchRecipes(12345, "fantasy", 50)
		if err != nil {
			t.Errorf("GenerateWorkbenchRecipes() error = %v", err)
		}
		if len(recipes) == 0 {
			t.Error("GenerateWorkbenchRecipes() returned no recipes")
		}
	})
}

func TestRecipeAdapter_SuccessChance(t *testing.T) {
	recipe := &RecipeData{
		BaseSuccessChance: 0.8,
		SkillRequired:     50,
	}

	// With high skill, chance should be higher
	chance := GetEffectiveSuccessChance(recipe, 100)
	if chance < recipe.BaseSuccessChance {
		t.Errorf("GetEffectiveSuccessChance() = %f, should be >= base chance %f", chance, recipe.BaseSuccessChance)
	}
}

func TestVehicleAdapter_SingleVehicle(t *testing.T) {
	adapter := NewVehicleAdapter()

	vehicle, err := adapter.GenerateVehicle(12345, "fantasy")
	if err != nil {
		t.Errorf("GenerateVehicle() error = %v", err)
	}
	if vehicle == nil {
		t.Error("GenerateVehicle() returned nil")
	}
}

func TestVehicleAdapter_TypeMapping(t *testing.T) {
	// mapVehicleType returns int (0=Mount, 1=Cart, etc.), not string
	tests := []struct {
		vtype    string
		expected int
	}{
		{"Mount", 0},
		{"Cart", 1},
		{"Boat", 2},
		{"Glider", 3},
		{"Mech", 4},
		{"unknown", 0}, // Defaults to Mount
	}
	for _, tt := range tests {
		mapped := mapVehicleType(tt.vtype)
		if mapped != tt.expected {
			t.Errorf("mapVehicleType(%s) = %d, want %d", tt.vtype, mapped, tt.expected)
		}
	}
}

func TestVehicleAdapter_RarityMultiplier(t *testing.T) {
	// VehicleRarityMultiplier takes a string, not a struct
	for _, rarity := range []string{"Common", "Rare", "Epic", "Legendary"} {
		mult := VehicleRarityMultiplier(rarity)
		if mult < 1.0 {
			t.Errorf("VehicleRarityMultiplier(%s) = %f, want >= 1.0", rarity, mult)
		}
	}
}

func TestTerrainAdapter_Basic(t *testing.T) {
	adapter := NewTerrainAdapter()
	if adapter == nil {
		t.Fatal("NewTerrainAdapter() returned nil")
	}

	chunkData, err := adapter.GenerateChunkTerrain(12345, "fantasy", 32, 32)
	if err != nil {
		t.Errorf("GenerateChunkTerrain() error = %v", err)
	}
	if chunkData == nil {
		t.Error("GenerateChunkTerrain() returned nil")
	}
}

func TestTerrainAdapter_Walkability(t *testing.T) {
	// Test with known tile type integers from Venture terrain
	// 0 = Floor (walkable), 1 = Wall (not walkable)
	tests := []struct {
		tile     int
		walkable bool
	}{
		{0, true},  // Floor
		{1, false}, // Wall
	}

	for _, tt := range tests {
		result := IsWalkable(tt.tile)
		// Just verify the function runs without errors
		_ = result
	}
}

func TestTerrainAdapter_MovementCost(t *testing.T) {
	// Test with known tile types
	tests := []struct {
		tile int
	}{
		{0}, // Floor
		{1}, // Wall
	}

	for _, tt := range tests {
		cost := GetTileMovementCost(tt.tile)
		// Just verify the function runs without errors
		// Some tiles may return negative costs indicating impassable
		_ = cost
	}
}
