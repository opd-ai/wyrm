package systems

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

func TestQuestSystemBasic(t *testing.T) {
	qs := NewQuestSystem()
	if qs.QuestStages == nil {
		t.Error("QuestStages should be initialized")
	}
}

func TestDefineQuest(t *testing.T) {
	qs := NewQuestSystem()

	stages := []QuestStageCondition{
		{RequiredFlag: "spoke_to_npc", FromStage: 0, NextStage: 1},
		{RequiredFlag: "item_collected", FromStage: 1, NextStage: 2, Completes: true},
	}

	qs.DefineQuest("test_quest", stages)

	if len(qs.QuestStages["test_quest"]) != 2 {
		t.Errorf("Expected 2 stages, got %d", len(qs.QuestStages["test_quest"]))
	}
}

func TestQuestStageAdvancement(t *testing.T) {
	qs := NewQuestSystem()

	stages := []QuestStageCondition{
		{RequiredFlag: "step1", FromStage: 0, NextStage: 1},
		{RequiredFlag: "step2", FromStage: 1, NextStage: 2, Completes: true},
	}
	qs.DefineQuest("test_quest", stages)

	world := ecs.NewWorld()
	entity := world.CreateEntity()
	quest := &components.Quest{
		ID:           "test_quest",
		CurrentStage: 0,
		Flags:        map[string]bool{"step1": true},
	}
	world.AddComponent(entity, quest)

	// Update should advance from stage 0 to stage 1
	qs.Update(world, 0.016)

	if quest.CurrentStage != 1 {
		t.Errorf("Expected stage 1, got %d", quest.CurrentStage)
	}

	// Set step2 flag and update again
	quest.Flags["step2"] = true
	qs.Update(world, 0.016)

	if !quest.Completed {
		t.Error("Quest should be completed")
	}
}

func TestQuestBranchLocking(t *testing.T) {
	qs := NewQuestSystem()

	stages := []QuestStageCondition{
		{RequiredFlag: "good_path", FromStage: 0, NextStage: 1, BranchID: "good", LocksBranch: "evil"},
		{RequiredFlag: "evil_path", FromStage: 0, NextStage: 2, BranchID: "evil", LocksBranch: "good"},
	}
	qs.DefineQuest("branching_quest", stages)

	world := ecs.NewWorld()
	entity := world.CreateEntity()
	quest := &components.Quest{
		ID:           "branching_quest",
		CurrentStage: 0,
		Flags:        map[string]bool{"good_path": true},
	}
	world.AddComponent(entity, quest)

	qs.Update(world, 0.016)

	if quest.CurrentStage != 1 {
		t.Errorf("Expected stage 1, got %d", quest.CurrentStage)
	}

	if !quest.IsBranchLocked("evil") {
		t.Error("Evil branch should be locked after taking good path")
	}
}

// Tests for Dynamic Quest Generation (FEATURES.md: Dynamic quest generation)

func TestDynamicQuestTypeString(t *testing.T) {
	tests := []struct {
		qtype DynamicQuestType
		want  string
	}{
		{QuestTypeFetch, "fetch"},
		{QuestTypeKill, "kill"},
		{QuestTypeEscort, "escort"},
		{QuestTypeInvestigate, "investigate"},
		{QuestTypeDeliver, "deliver"},
		{QuestTypeRescue, "rescue"},
		{QuestTypeSabotage, "sabotage"},
		{QuestTypeNegotiate, "negotiate"},
		{QuestTypeExplore, "explore"},
		{QuestTypeDefend, "defend"},
		{DynamicQuestType(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.qtype.String(); got != tt.want {
			t.Errorf("%d.String() = %s, want %s", tt.qtype, got, tt.want)
		}
	}
}

func TestDynamicQuestGenerator(t *testing.T) {
	gen := NewDynamicQuestGenerator(12345)

	if gen == nil {
		t.Fatal("Generator should not be nil")
	}

	if gen.config == nil {
		t.Error("Generator config should be initialized")
	}
}

func TestGenerateFromFamineState(t *testing.T) {
	gen := NewDynamicQuestGenerator(12345)

	worldState := &WorldState{
		FamineLevel: 0.7, // Above 0.5 threshold
	}

	quests := gen.GenerateFromWorldState(worldState, "fantasy")

	if len(quests) == 0 {
		t.Error("Should generate at least one quest for famine")
	}

	// Verify famine quest properties
	found := false
	for _, q := range quests {
		if q.WorldTrigger == "famine" {
			found = true
			if q.Type != QuestTypeFetch {
				t.Errorf("Famine quest should be fetch type, got %s", q.Type)
			}
			if q.Reward <= 0 {
				t.Error("Quest should have positive reward")
			}
		}
	}

	if !found {
		t.Error("Should have a famine-triggered quest")
	}
}

func TestGenerateFromWarState(t *testing.T) {
	gen := NewDynamicQuestGenerator(12345)

	worldState := &WorldState{
		WarIntensity: 0.7,
	}

	quests := gen.GenerateFromWorldState(worldState, "sci-fi")

	found := false
	for _, q := range quests {
		if q.WorldTrigger == "war" {
			found = true
			if q.Type != QuestTypeKill {
				t.Errorf("War quest should be kill type, got %s", q.Type)
			}
		}
	}

	if !found {
		t.Error("Should have a war-triggered quest")
	}
}

func TestGenerateFromMultipleWorldConditions(t *testing.T) {
	gen := NewDynamicQuestGenerator(12345)

	worldState := &WorldState{
		FamineLevel:     0.8,
		WarIntensity:    0.6,
		BanditActivity:  0.7,
		MonsterThreat:   0.9,
		PoliticalUnrest: 0.6,
	}

	quests := gen.GenerateFromWorldState(worldState, "fantasy")

	if len(quests) < 5 {
		t.Errorf("Should generate quests for all 5 conditions, got %d", len(quests))
	}

	// Verify all triggers are represented
	triggers := make(map[string]bool)
	for _, q := range quests {
		triggers[q.WorldTrigger] = true
	}

	expected := []string{"famine", "war", "bandits", "monsters", "politics"}
	for _, e := range expected {
		if !triggers[e] {
			t.Errorf("Missing quest trigger: %s", e)
		}
	}
}

func TestGenerateFromLowWorldState(t *testing.T) {
	gen := NewDynamicQuestGenerator(12345)

	worldState := &WorldState{
		FamineLevel:    0.1, // Below threshold
		WarIntensity:   0.2,
		BanditActivity: 0.3,
	}

	quests := gen.GenerateFromWorldState(worldState, "fantasy")

	if len(quests) != 0 {
		t.Errorf("Should not generate quests when conditions are below threshold, got %d", len(quests))
	}
}

func TestAllGenresHaveDynamicQuests(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		gen := NewDynamicQuestGenerator(12345)
		worldState := &WorldState{
			FamineLevel:  0.8,
			WarIntensity: 0.8,
		}

		quests := gen.GenerateFromWorldState(worldState, genre)

		if len(quests) < 2 {
			t.Errorf("Genre %s should generate at least 2 quests, got %d", genre, len(quests))
		}

		// Verify quests have genre-appropriate content
		for _, q := range quests {
			if q.Title == "" || q.Description == "" {
				t.Errorf("Genre %s quest missing title or description", genre)
			}
		}
	}
}

// Tests for Radiant Quest System (FEATURES.md: Radiant quest system)

func TestRadiantQuestBoard(t *testing.T) {
	board := NewRadiantQuestBoard("town_square", "fantasy")

	if board == nil {
		t.Fatal("Board should not be nil")
	}

	if board.LocationID != "town_square" {
		t.Errorf("LocationID = %s, want town_square", board.LocationID)
	}

	if board.Genre != "fantasy" {
		t.Errorf("Genre = %s, want fantasy", board.Genre)
	}
}

func TestRadiantQuestBoardRefresh(t *testing.T) {
	board := NewRadiantQuestBoard("inn_notice", "fantasy")

	board.RefreshQuests(12345)

	quests := board.GetActiveQuests()

	if len(quests) == 0 {
		t.Error("Should have generated some quests")
	}

	if board.QuestCount() != len(quests) {
		t.Error("QuestCount should match GetActiveQuests length")
	}

	// Verify quest properties
	for _, q := range quests {
		if q.ID == "" {
			t.Error("Quest should have ID")
		}
		if q.Title == "" {
			t.Error("Quest should have title")
		}
		if q.Description == "" {
			t.Error("Quest should have description")
		}
		if q.Reward <= 0 {
			t.Error("Quest should have positive reward")
		}
		if q.Difficulty < 1 || q.Difficulty > 5 {
			t.Errorf("Difficulty %d should be 1-5", q.Difficulty)
		}
	}
}

func TestRadiantQuestBoardDeterminism(t *testing.T) {
	board1 := NewRadiantQuestBoard("test_loc", "fantasy")
	board2 := NewRadiantQuestBoard("test_loc", "fantasy")

	board1.RefreshQuests(42)
	board2.RefreshQuests(42)

	quests1 := board1.GetActiveQuests()
	quests2 := board2.GetActiveQuests()

	if len(quests1) != len(quests2) {
		t.Fatal("Same seed should produce same number of quests")
	}

	for i := range quests1 {
		if quests1[i].Title != quests2[i].Title {
			t.Errorf("Quest %d titles differ: %s vs %s", i, quests1[i].Title, quests2[i].Title)
		}
	}
}

func TestRadiantQuestBoardAllGenres(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		board := NewRadiantQuestBoard("test", genre)
		board.RefreshQuests(12345)

		quests := board.GetActiveQuests()
		if len(quests) == 0 {
			t.Errorf("Genre %s should generate quests", genre)
		}
	}
}

func TestDefaultRadiantConfig(t *testing.T) {
	config := DefaultRadiantConfig()

	if config == nil {
		t.Fatal("Config should not be nil")
	}

	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		templates := config.Templates[genre]
		if len(templates) == 0 {
			t.Errorf("Genre %s should have templates", genre)
		}

		for _, tmpl := range templates {
			if len(tmpl.Targets) == 0 {
				t.Errorf("Genre %s template should have targets", genre)
			}
			if len(tmpl.Givers) == 0 {
				t.Errorf("Genre %s template should have givers", genre)
			}
			if tmpl.BaseReward <= 0 {
				t.Errorf("Genre %s template should have positive reward", genre)
			}
		}
	}
}

func TestFormatPattern(t *testing.T) {
	tests := []struct {
		pattern string
		target  string
		giver   string
		want    string
	}{
		{"Fetch %s", "herbs", "healer", "Fetch herbs"},
		{"Bring %s to %s", "package", "merchant", "Bring package to merchant"},
		{"No placeholders", "target", "giver", "No placeholders"},
		{"%s needs your help", "merchant", "nobody", "merchant needs your help"},
	}

	for _, tt := range tests {
		got := formatPattern(tt.pattern, tt.target, tt.giver)
		if got != tt.want {
			t.Errorf("formatPattern(%q, %q, %q) = %q, want %q",
				tt.pattern, tt.target, tt.giver, got, tt.want)
		}
	}
}

func TestUniqueQuestIDs(t *testing.T) {
	gen := NewDynamicQuestGenerator(12345)

	worldState := &WorldState{
		FamineLevel:     0.8,
		WarIntensity:    0.8,
		BanditActivity:  0.8,
		MonsterThreat:   0.8,
		PoliticalUnrest: 0.8,
	}

	quests := gen.GenerateFromWorldState(worldState, "fantasy")

	ids := make(map[string]bool)
	for _, q := range quests {
		if ids[q.ID] {
			t.Errorf("Duplicate quest ID: %s", q.ID)
		}
		ids[q.ID] = true
	}
}

func BenchmarkDynamicQuestGeneration(b *testing.B) {
	gen := NewDynamicQuestGenerator(12345)
	worldState := &WorldState{
		FamineLevel:    0.8,
		WarIntensity:   0.8,
		BanditActivity: 0.8,
	}

	for i := 0; i < b.N; i++ {
		gen.GenerateFromWorldState(worldState, "fantasy")
	}
}

func BenchmarkRadiantQuestRefresh(b *testing.B) {
	board := NewRadiantQuestBoard("test", "fantasy")

	for i := 0; i < b.N; i++ {
		board.RefreshQuests(int64(i))
	}
}

// ============================================================================
// Faction Arc Tests
// ============================================================================

func TestNewFactionArcManager(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		m := NewFactionArcManager(genre)
		if m == nil {
			t.Fatalf("Manager for genre %s should not be nil", genre)
		}
		if len(m.Arcs) == 0 {
			t.Errorf("Genre %s should have default arcs", genre)
		}
	}
}

func TestFactionArcManagerRegisterArc(t *testing.T) {
	m := NewFactionArcManager("fantasy")

	arc := &FactionQuestArc{
		ID:          "custom_arc",
		FactionID:   "custom_faction",
		Name:        "Custom Arc",
		Description: "A custom test arc",
		Quests: []FactionArcQuest{
			{ID: "q1", Title: "Quest 1", NextQuestID: "q2"},
			{ID: "q2", Title: "Quest 2"},
		},
	}

	m.RegisterArc(arc)

	if _, ok := m.Arcs["custom_arc"]; !ok {
		t.Error("Arc should be registered")
	}

	factionArcs := m.GetFactionArcs("custom_faction")
	if len(factionArcs) == 0 {
		t.Error("Should find arc by faction ID")
	}
}

func TestFactionArcStartAndProgress(t *testing.T) {
	m := NewFactionArcManager("fantasy")

	// Create a simple test arc
	arc := &FactionQuestArc{
		ID:           "test_arc",
		FactionID:    "test_faction",
		Name:         "Test Arc",
		Description:  "A test arc",
		RequiredRank: 0,
		Quests: []FactionArcQuest{
			{
				ID:          "test_q1",
				Title:       "Test Quest 1",
				Objectives:  []ArcQuestGoal{{Type: "kill", Target: "enemy", Count: 5}},
				NextQuestID: "test_q2",
			},
			{
				ID:         "test_q2",
				Title:      "Test Quest 2",
				Objectives: []ArcQuestGoal{{Type: "fetch", Target: "item", Count: 3}},
			},
		},
	}
	m.RegisterArc(arc)

	playerEntity := uint64(1)

	// Start the arc
	if !m.StartArc(playerEntity, "test_arc") {
		t.Fatal("Should be able to start arc")
	}

	// Verify current quest
	quest := m.GetCurrentQuest(playerEntity, "test_arc")
	if quest == nil {
		t.Fatal("Should have current quest")
	}
	if quest.ID != "test_q1" {
		t.Errorf("Current quest should be test_q1, got %s", quest.ID)
	}

	// Complete objective
	if !m.CompleteObjective(playerEntity, "test_arc", 0) {
		t.Error("Should complete objective")
	}

	// Quest should advance
	quest = m.GetCurrentQuest(playerEntity, "test_arc")
	if quest.ID != "test_q2" {
		t.Errorf("Should advance to test_q2, got %s", quest.ID)
	}

	// Complete final quest
	if !m.CompleteObjective(playerEntity, "test_arc", 0) {
		t.Error("Should complete final objective")
	}

	// Arc should be complete
	if !m.IsArcComplete(playerEntity, "test_arc") {
		t.Error("Arc should be complete")
	}
}

func TestFactionArcMutualExclusivity(t *testing.T) {
	m := NewFactionArcManager("fantasy")

	// Create mutually exclusive arcs
	arc1 := &FactionQuestArc{
		ID:               "arc_good",
		FactionID:        "test_faction",
		Name:             "Good Path",
		MutuallyExcludes: []string{"arc_evil"},
		IsExclusive:      true,
		Quests: []FactionArcQuest{
			{ID: "good_q1", Title: "Good Quest", Objectives: []ArcQuestGoal{{Type: "talk", Target: "sage"}}},
		},
	}
	arc2 := &FactionQuestArc{
		ID:               "arc_evil",
		FactionID:        "test_faction",
		Name:             "Evil Path",
		MutuallyExcludes: []string{"arc_good"},
		IsExclusive:      true,
		Quests: []FactionArcQuest{
			{ID: "evil_q1", Title: "Evil Quest", Objectives: []ArcQuestGoal{{Type: "kill", Target: "innocent"}}},
		},
	}

	m.RegisterArc(arc1)
	m.RegisterArc(arc2)

	playerEntity := uint64(2)

	// Start good path
	if !m.StartArc(playerEntity, "arc_good") {
		t.Fatal("Should start good arc")
	}

	// Evil path should now be locked
	available := m.GetAvailableArcs(playerEntity, "test_faction")
	for _, arc := range available {
		if arc.ID == "arc_evil" {
			t.Error("Evil arc should be locked after choosing good")
		}
	}
}

func TestFactionArcRankRequirement(t *testing.T) {
	m := NewFactionArcManager("fantasy")

	arc := &FactionQuestArc{
		ID:           "high_rank_arc",
		FactionID:    "test_faction",
		Name:         "Elite Quest",
		RequiredRank: 5,
		Quests: []FactionArcQuest{
			{ID: "elite_q1", Title: "Elite Quest", Objectives: []ArcQuestGoal{{Type: "kill", Target: "boss"}}},
		},
	}
	m.RegisterArc(arc)

	playerEntity := uint64(3)

	// Should not be available without rank
	available := m.GetAvailableArcs(playerEntity, "test_faction")
	for _, a := range available {
		if a.ID == "high_rank_arc" {
			t.Error("High rank arc should not be available without sufficient rank")
		}
	}

	// Should not be able to start
	if m.StartArc(playerEntity, "high_rank_arc") {
		t.Error("Should not start arc without sufficient rank")
	}

	// Manually set rank
	progress := m.getOrCreateProgress(playerEntity)
	progress.FactionRanks["test_faction"] = 5

	// Now should be available
	available = m.GetAvailableArcs(playerEntity, "test_faction")
	found := false
	for _, a := range available {
		if a.ID == "high_rank_arc" {
			found = true
			break
		}
	}
	if !found {
		t.Error("High rank arc should be available with sufficient rank")
	}
}

func TestGetArcProgress(t *testing.T) {
	m := NewFactionArcManager("fantasy")

	arc := &FactionQuestArc{
		ID:        "progress_arc",
		FactionID: "test_faction",
		Name:      "Progress Test",
		Quests: []FactionArcQuest{
			{
				ID:    "prog_q1",
				Title: "Progress Quest",
				Objectives: []ArcQuestGoal{
					{Type: "kill", Target: "enemy1", Count: 1},
					{Type: "kill", Target: "enemy2", Count: 1},
					{Type: "kill", Target: "enemy3", Count: 1},
				},
			},
		},
	}
	m.RegisterArc(arc)

	playerEntity := uint64(4)
	m.StartArc(playerEntity, "progress_arc")

	// Initial progress
	questID, done, total := m.GetArcProgress(playerEntity, "progress_arc")
	if questID != "prog_q1" {
		t.Errorf("Quest ID should be prog_q1, got %s", questID)
	}
	if done != 0 {
		t.Errorf("Done should be 0, got %d", done)
	}
	if total != 3 {
		t.Errorf("Total should be 3, got %d", total)
	}

	// Complete one objective
	m.CompleteObjective(playerEntity, "progress_arc", 0)

	_, done, _ = m.GetArcProgress(playerEntity, "progress_arc")
	if done != 1 {
		t.Errorf("Done should be 1, got %d", done)
	}
}

func TestGetCompletedArcs(t *testing.T) {
	m := NewFactionArcManager("fantasy")

	arc := &FactionQuestArc{
		ID:        "complete_arc",
		FactionID: "test_faction",
		Name:      "Completable Arc",
		Quests: []FactionArcQuest{
			{
				ID:         "complete_q1",
				Title:      "Only Quest",
				Objectives: []ArcQuestGoal{{Type: "talk", Target: "npc"}},
			},
		},
	}
	m.RegisterArc(arc)

	playerEntity := uint64(5)
	m.StartArc(playerEntity, "complete_arc")
	m.CompleteObjective(playerEntity, "complete_arc", 0)

	completed := m.GetCompletedArcs(playerEntity)
	if len(completed) != 1 {
		t.Errorf("Should have 1 completed arc, got %d", len(completed))
	}
	if completed[0] != "complete_arc" {
		t.Errorf("Completed arc should be complete_arc, got %s", completed[0])
	}
}

func TestGenreFactionArcs(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		arcs := getGenreFactionArcs(genre)
		if len(arcs) == 0 {
			t.Errorf("Genre %s should have faction arcs", genre)
		}

		for _, arc := range arcs {
			if arc.ID == "" {
				t.Errorf("Arc in genre %s has empty ID", genre)
			}
			if arc.FactionID == "" {
				t.Errorf("Arc %s in genre %s has empty faction ID", arc.ID, genre)
			}
			if len(arc.Quests) == 0 {
				t.Errorf("Arc %s in genre %s has no quests", arc.ID, genre)
			}

			for _, quest := range arc.Quests {
				if quest.ID == "" {
					t.Errorf("Quest in arc %s has empty ID", arc.ID)
				}
				if len(quest.Objectives) == 0 {
					t.Errorf("Quest %s in arc %s has no objectives", quest.ID, arc.ID)
				}
			}
		}
	}
}

func BenchmarkFactionArcStartAndComplete(b *testing.B) {
	m := NewFactionArcManager("fantasy")

	arc := &FactionQuestArc{
		ID:        "bench_arc",
		FactionID: "bench_faction",
		Name:      "Benchmark Arc",
		Quests: []FactionArcQuest{
			{
				ID:         "bench_q1",
				Title:      "Bench Quest",
				Objectives: []ArcQuestGoal{{Type: "kill", Target: "enemy"}},
			},
		},
	}
	m.RegisterArc(arc)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		playerEntity := uint64(i)
		m.StartArc(playerEntity, "bench_arc")
		m.CompleteObjective(playerEntity, "bench_arc", 0)
	}
}
