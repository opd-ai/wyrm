package systems

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

func TestNewCriminalFactionQuestSystem(t *testing.T) {
	frs := NewFactionRankSystem("fantasy")

	tests := []struct {
		name  string
		genre string
		seed  int64
	}{
		{"fantasy", "fantasy", 12345},
		{"cyberpunk", "cyberpunk", 67890},
		{"zero seed", "sci-fi", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cqs := NewCriminalFactionQuestSystem(frs, tt.genre, tt.seed)
			if cqs == nil {
				t.Fatal("NewCriminalFactionQuestSystem returned nil")
			}
			if cqs.factionRankSystem != frs {
				t.Error("factionRankSystem not set")
			}
			if cqs.genre != tt.genre {
				t.Errorf("genre = %q, want %q", cqs.genre, tt.genre)
			}
			if cqs.rngSeed != tt.seed {
				t.Errorf("rngSeed = %d, want %d", cqs.rngSeed, tt.seed)
			}
			if cqs.QuestsPerRankTier != 3 {
				t.Errorf("QuestsPerRankTier = %d, want 3", cqs.QuestsPerRankTier)
			}
		})
	}
}

func TestCriminalQuestType_String(t *testing.T) {
	tests := []struct {
		questType CriminalQuestType
		want      string
	}{
		{CriminalQuestTheft, "Theft"},
		{CriminalQuestSmuggling, "Smuggling"},
		{CriminalQuestHeist, "Heist"},
		{CriminalQuestAssassination, "Assassination"},
		{CriminalQuestExtortion, "Extortion"},
		{CriminalQuestSabotage, "Sabotage"},
		{CriminalQuestInfiltration, "Infiltration"},
		{CriminalQuestJailbreak, "Jailbreak"},
		{CriminalQuestTurf, "Turf War"},
		{CriminalQuestBoss, "Leadership Challenge"},
		{CriminalQuestType(99), "Unknown"},
	}

	for _, tt := range tests {
		got := tt.questType.String()
		if got != tt.want {
			t.Errorf("CriminalQuestType(%d).String() = %q, want %q", tt.questType, got, tt.want)
		}
	}
}

func TestCriminalQuestType_MinRank(t *testing.T) {
	tests := []struct {
		questType CriminalQuestType
		minRank   int
	}{
		{CriminalQuestTheft, 1},
		{CriminalQuestSmuggling, 2},
		{CriminalQuestExtortion, 3},
		{CriminalQuestSabotage, 4},
		{CriminalQuestHeist, 5},
		{CriminalQuestInfiltration, 6},
		{CriminalQuestAssassination, 7},
		{CriminalQuestJailbreak, 8},
		{CriminalQuestTurf, 9},
		{CriminalQuestBoss, 10},
	}

	for _, tt := range tests {
		got := tt.questType.MinRank()
		if got != tt.minRank {
			t.Errorf("CriminalQuestType(%d).MinRank() = %d, want %d", tt.questType, got, tt.minRank)
		}
	}
}

func TestCriminalQuestType_BaseXPReward(t *testing.T) {
	// Just verify all types return positive values
	allTypes := []CriminalQuestType{
		CriminalQuestTheft, CriminalQuestSmuggling, CriminalQuestExtortion,
		CriminalQuestSabotage, CriminalQuestHeist, CriminalQuestInfiltration,
		CriminalQuestAssassination, CriminalQuestJailbreak, CriminalQuestTurf,
		CriminalQuestBoss,
	}

	for _, qt := range allTypes {
		xp := qt.BaseXPReward()
		if xp <= 0 {
			t.Errorf("CriminalQuestType(%d).BaseXPReward() = %d, want > 0", qt, xp)
		}
	}
}

func TestCriminalQuestType_BaseCurrencyReward(t *testing.T) {
	allTypes := []CriminalQuestType{
		CriminalQuestTheft, CriminalQuestSmuggling, CriminalQuestExtortion,
		CriminalQuestSabotage, CriminalQuestHeist, CriminalQuestInfiltration,
		CriminalQuestAssassination, CriminalQuestJailbreak, CriminalQuestTurf,
		CriminalQuestBoss,
	}

	for _, qt := range allTypes {
		currency := qt.BaseCurrencyReward()
		if currency <= 0 {
			t.Errorf("CriminalQuestType(%d).BaseCurrencyReward() = %d, want > 0", qt, currency)
		}
	}
}

func TestCriminalFactionQuestSystem_GenerateQuestsForFaction(t *testing.T) {
	frs := NewFactionRankSystem("fantasy")
	cqs := NewCriminalFactionQuestSystem(frs, "fantasy", 12345)

	quests := cqs.GenerateQuestsForFaction("thieves_guild", 5)

	if len(quests) == 0 {
		t.Fatal("GenerateQuestsForFaction returned no quests")
	}

	for _, quest := range quests {
		if quest.ID == "" {
			t.Error("Quest has empty ID")
		}
		if quest.FactionID != "thieves_guild" {
			t.Errorf("Quest faction = %q, want thieves_guild", quest.FactionID)
		}
		if quest.State != CriminalQuestAvailable {
			t.Errorf("Quest state = %d, want Available", quest.State)
		}
		if quest.Title == "" {
			t.Error("Quest has empty title")
		}
		if len(quest.Objectives) == 0 {
			t.Error("Quest has no objectives")
		}
	}
}

func TestCriminalFactionQuestSystem_AcceptQuest(t *testing.T) {
	w := ecs.NewWorld()
	frs := NewFactionRankSystem("fantasy")
	cqs := NewCriminalFactionQuestSystem(frs, "fantasy", 12345)

	// Create player with faction membership
	player := w.CreateEntity()
	membership := &components.FactionMembership{
		Memberships: map[string]*components.FactionMemberInfo{
			"thieves_guild": {FactionID: "thieves_guild", Rank: 5},
		},
	}
	w.AddComponent(player, membership)

	// Generate quests
	quests := cqs.GenerateQuestsForFaction("thieves_guild", 5)
	if len(quests) == 0 {
		t.Fatal("No quests generated")
	}

	questID := quests[0].ID
	result := cqs.AcceptQuest(w, questID, uint64(player))
	if !result {
		t.Error("AcceptQuest returned false")
	}

	quest := cqs.GetQuest(questID)
	if quest.State != CriminalQuestActive {
		t.Errorf("Quest state = %d, want Active", quest.State)
	}
	if quest.AssignedTo != uint64(player) {
		t.Errorf("Quest assigned to %d, want %d", quest.AssignedTo, player)
	}

	// Verify active quest tracking
	activeQuest := cqs.GetActiveQuest(uint64(player))
	if activeQuest == nil || activeQuest.ID != questID {
		t.Error("GetActiveQuest doesn't return the accepted quest")
	}
}

func TestCriminalFactionQuestSystem_AcceptQuest_InvalidQuest(t *testing.T) {
	w := ecs.NewWorld()
	frs := NewFactionRankSystem("fantasy")
	cqs := NewCriminalFactionQuestSystem(frs, "fantasy", 12345)

	player := w.CreateEntity()

	result := cqs.AcceptQuest(w, "invalid-quest", uint64(player))
	if result {
		t.Error("AcceptQuest with invalid ID should return false")
	}
}

func TestCriminalFactionQuestSystem_AcceptQuest_AlreadyActive(t *testing.T) {
	w := ecs.NewWorld()
	frs := NewFactionRankSystem("fantasy")
	cqs := NewCriminalFactionQuestSystem(frs, "fantasy", 12345)

	player := w.CreateEntity()
	w.AddComponent(player, &components.FactionMembership{
		Memberships: map[string]*components.FactionMemberInfo{
			"thieves_guild": {FactionID: "thieves_guild", Rank: 5},
		},
	})

	quests := cqs.GenerateQuestsForFaction("thieves_guild", 5)
	cqs.AcceptQuest(w, quests[0].ID, uint64(player))

	// Try to accept another quest
	if len(quests) > 1 {
		result := cqs.AcceptQuest(w, quests[1].ID, uint64(player))
		if result {
			t.Error("Should not be able to accept another quest while one is active")
		}
	}
}

func TestCriminalFactionQuestSystem_AcceptQuest_RankTooLow(t *testing.T) {
	w := ecs.NewWorld()
	frs := NewFactionRankSystem("fantasy")
	cqs := NewCriminalFactionQuestSystem(frs, "fantasy", 12345)

	player := w.CreateEntity()
	w.AddComponent(player, &components.FactionMembership{
		Memberships: map[string]*components.FactionMemberInfo{
			"thieves_guild": {FactionID: "thieves_guild", Rank: 1}, // Low rank
		},
	})

	// Generate high-rank quests
	quests := cqs.GenerateQuestsForFaction("thieves_guild", 10)

	// Find a quest requiring high rank
	for _, quest := range quests {
		if quest.Type.MinRank() > 1 {
			result := cqs.AcceptQuest(w, quest.ID, uint64(player))
			if result {
				t.Errorf("Should not accept quest requiring rank %d with rank 1", quest.Type.MinRank())
			}
			break
		}
	}
}

func TestCriminalFactionQuestSystem_CompleteObjective(t *testing.T) {
	w := ecs.NewWorld()
	frs := NewFactionRankSystem("fantasy")
	cqs := NewCriminalFactionQuestSystem(frs, "fantasy", 12345)

	player := w.CreateEntity()
	w.AddComponent(player, &components.FactionMembership{
		Memberships: map[string]*components.FactionMemberInfo{
			"thieves_guild": {FactionID: "thieves_guild", Rank: 5},
		},
	})

	quests := cqs.GenerateQuestsForFaction("thieves_guild", 5)
	questID := quests[0].ID
	cqs.AcceptQuest(w, questID, uint64(player))

	quest := cqs.GetQuest(questID)
	if len(quest.Objectives) == 0 {
		t.Fatal("Quest has no objectives")
	}

	objID := quest.Objectives[0].ID
	result := cqs.CompleteObjective(questID, objID)
	if !result {
		t.Error("CompleteObjective returned false")
	}

	// Check progress
	if quest.Objectives[0].Progress < 1 {
		t.Error("Objective progress not incremented")
	}
}

func TestCriminalFactionQuestSystem_CompleteObjective_InvalidQuest(t *testing.T) {
	frs := NewFactionRankSystem("fantasy")
	cqs := NewCriminalFactionQuestSystem(frs, "fantasy", 12345)

	result := cqs.CompleteObjective("invalid-quest", "obj1")
	if result {
		t.Error("CompleteObjective with invalid quest should return false")
	}
}

func TestCriminalFactionQuestSystem_CheckQuestComplete(t *testing.T) {
	w := ecs.NewWorld()
	frs := NewFactionRankSystem("fantasy")
	cqs := NewCriminalFactionQuestSystem(frs, "fantasy", 12345)

	player := w.CreateEntity()
	w.AddComponent(player, &components.FactionMembership{
		Memberships: map[string]*components.FactionMemberInfo{
			"thieves_guild": {FactionID: "thieves_guild", Rank: 5},
		},
	})

	quests := cqs.GenerateQuestsForFaction("thieves_guild", 5)
	questID := quests[0].ID
	cqs.AcceptQuest(w, questID, uint64(player))

	// Not complete yet
	if cqs.CheckQuestComplete(questID) {
		t.Error("Quest should not be complete initially")
	}

	// Complete all required objectives
	quest := cqs.GetQuest(questID)
	for i := range quest.Objectives {
		if !quest.Objectives[i].IsOptional {
			quest.Objectives[i].IsCompleted = true
		}
	}

	if !cqs.CheckQuestComplete(questID) {
		t.Error("Quest should be complete when all required objectives done")
	}
}

func TestCriminalFactionQuestSystem_CompleteQuest(t *testing.T) {
	w := ecs.NewWorld()
	frs := NewFactionRankSystem("fantasy")
	cqs := NewCriminalFactionQuestSystem(frs, "fantasy", 12345)

	player := w.CreateEntity()
	w.AddComponent(player, &components.FactionMembership{
		Memberships: map[string]*components.FactionMemberInfo{
			"thieves_guild": {FactionID: "thieves_guild", Rank: 5, XP: 0, XPToNext: 1000},
		},
	})

	quests := cqs.GenerateQuestsForFaction("thieves_guild", 5)
	questID := quests[0].ID
	cqs.AcceptQuest(w, questID, uint64(player))

	// Complete all objectives manually
	quest := cqs.GetQuest(questID)
	for i := range quest.Objectives {
		quest.Objectives[i].IsCompleted = true
	}

	result := cqs.CompleteQuest(w, questID)
	if !result {
		t.Error("CompleteQuest returned false")
	}

	if quest.State != CriminalQuestCompleted {
		t.Errorf("Quest state = %d, want Completed", quest.State)
	}

	// Should no longer have active quest
	if cqs.GetActiveQuest(uint64(player)) != nil {
		t.Error("Player should not have active quest after completion")
	}
}

func TestCriminalFactionQuestSystem_CompleteQuest_NotAllDone(t *testing.T) {
	w := ecs.NewWorld()
	frs := NewFactionRankSystem("fantasy")
	cqs := NewCriminalFactionQuestSystem(frs, "fantasy", 12345)

	player := w.CreateEntity()
	w.AddComponent(player, &components.FactionMembership{
		Memberships: map[string]*components.FactionMemberInfo{
			"thieves_guild": {FactionID: "thieves_guild", Rank: 5},
		},
	})

	quests := cqs.GenerateQuestsForFaction("thieves_guild", 5)
	questID := quests[0].ID
	cqs.AcceptQuest(w, questID, uint64(player))

	// Don't complete objectives
	result := cqs.CompleteQuest(w, questID)
	if result {
		t.Error("CompleteQuest should fail when objectives not done")
	}
}

func TestCriminalFactionQuestSystem_AbandonQuest(t *testing.T) {
	w := ecs.NewWorld()
	frs := NewFactionRankSystem("fantasy")
	cqs := NewCriminalFactionQuestSystem(frs, "fantasy", 12345)

	player := w.CreateEntity()
	w.AddComponent(player, &components.FactionMembership{
		Memberships: map[string]*components.FactionMemberInfo{
			"thieves_guild": {FactionID: "thieves_guild", Rank: 5, Reputation: 50},
		},
	})

	quests := cqs.GenerateQuestsForFaction("thieves_guild", 5)
	questID := quests[0].ID
	cqs.AcceptQuest(w, questID, uint64(player))

	result := cqs.AbandonQuest(w, questID)
	if !result {
		t.Error("AbandonQuest returned false")
	}

	quest := cqs.GetQuest(questID)
	if quest.State != CriminalQuestFailed {
		t.Errorf("Quest state = %d, want Failed", quest.State)
	}

	// Check reputation loss
	comp, _ := w.GetComponent(player, "FactionMembership")
	membership := comp.(*components.FactionMembership)
	info := membership.GetMembership("thieves_guild")
	if info.Reputation >= 50 {
		t.Error("Reputation should decrease on quest failure")
	}
}

func TestCriminalFactionQuestSystem_Update_TimeExpired(t *testing.T) {
	w := ecs.NewWorld()
	frs := NewFactionRankSystem("fantasy")
	cqs := NewCriminalFactionQuestSystem(frs, "fantasy", 12345)

	player := w.CreateEntity()
	w.AddComponent(player, &components.FactionMembership{
		Memberships: map[string]*components.FactionMemberInfo{
			"thieves_guild": {FactionID: "thieves_guild", Rank: 5},
		},
	})

	quests := cqs.GenerateQuestsForFaction("thieves_guild", 5)
	// Find a quest with time limit
	var timedQuest *CriminalQuest
	for _, q := range quests {
		if q.Type == CriminalQuestTheft { // Theft has 30 min limit
			timedQuest = q
			break
		}
	}
	if timedQuest == nil {
		t.Skip("No timed quest generated")
	}

	cqs.AcceptQuest(w, timedQuest.ID, uint64(player))

	// Simulate time passing beyond limit
	cqs.Update(w, timedQuest.TimeLimit+1)

	if timedQuest.State != CriminalQuestFailed {
		t.Errorf("Quest should fail after time expires, state = %d", timedQuest.State)
	}
}

func TestCriminalFactionQuestSystem_GetQuestProgress(t *testing.T) {
	w := ecs.NewWorld()
	frs := NewFactionRankSystem("fantasy")
	cqs := NewCriminalFactionQuestSystem(frs, "fantasy", 12345)

	player := w.CreateEntity()
	w.AddComponent(player, &components.FactionMembership{
		Memberships: map[string]*components.FactionMemberInfo{
			"thieves_guild": {FactionID: "thieves_guild", Rank: 5},
		},
	})

	quests := cqs.GenerateQuestsForFaction("thieves_guild", 5)
	questID := quests[0].ID
	cqs.AcceptQuest(w, questID, uint64(player))

	completed, total := cqs.GetQuestProgress(questID)
	if completed != 0 {
		t.Errorf("Initial completed = %d, want 0", completed)
	}
	if total == 0 {
		t.Error("Total should be > 0")
	}

	// Complete one objective
	quest := cqs.GetQuest(questID)
	for i := range quest.Objectives {
		if !quest.Objectives[i].IsOptional {
			quest.Objectives[i].IsCompleted = true
			break
		}
	}

	completed, _ = cqs.GetQuestProgress(questID)
	if completed != 1 {
		t.Errorf("After completing one, completed = %d, want 1", completed)
	}
}

func TestCriminalFactionQuestSystem_GetRemainingTime(t *testing.T) {
	w := ecs.NewWorld()
	frs := NewFactionRankSystem("fantasy")
	cqs := NewCriminalFactionQuestSystem(frs, "fantasy", 12345)

	player := w.CreateEntity()
	w.AddComponent(player, &components.FactionMembership{
		Memberships: map[string]*components.FactionMemberInfo{
			"thieves_guild": {FactionID: "thieves_guild", Rank: 5},
		},
	})

	quests := cqs.GenerateQuestsForFaction("thieves_guild", 5)
	// Find timed quest
	var timedQuest *CriminalQuest
	for _, q := range quests {
		if q.TimeLimit > 0 {
			timedQuest = q
			break
		}
	}
	if timedQuest == nil {
		t.Skip("No timed quest generated")
	}

	cqs.AcceptQuest(w, timedQuest.ID, uint64(player))

	remaining := cqs.GetRemainingTime(timedQuest.ID)
	if remaining <= 0 {
		t.Errorf("Remaining time = %v, want > 0", remaining)
	}

	// Simulate time passing
	cqs.Update(w, 100)
	remaining2 := cqs.GetRemainingTime(timedQuest.ID)
	if remaining2 >= remaining {
		t.Error("Remaining time should decrease")
	}
}

func TestCriminalFactionQuestSystem_GetAvailableQuests(t *testing.T) {
	frs := NewFactionRankSystem("fantasy")
	cqs := NewCriminalFactionQuestSystem(frs, "fantasy", 12345)

	cqs.GenerateQuestsForFaction("thieves_guild", 5)
	cqs.GenerateQuestsForFaction("assassins", 5)

	thiefQuests := cqs.GetAvailableQuests("thieves_guild")
	if len(thiefQuests) == 0 {
		t.Error("Should have available quests for thieves_guild")
	}

	for _, q := range thiefQuests {
		if q.FactionID != "thieves_guild" {
			t.Errorf("Quest has wrong faction: %s", q.FactionID)
		}
		if q.State != CriminalQuestAvailable {
			t.Error("Non-available quest in GetAvailableQuests result")
		}
	}
}

func TestCriminalFactionQuestSystem_GetPlayerQuestHistory(t *testing.T) {
	w := ecs.NewWorld()
	frs := NewFactionRankSystem("fantasy")
	cqs := NewCriminalFactionQuestSystem(frs, "fantasy", 12345)

	player := w.CreateEntity()
	w.AddComponent(player, &components.FactionMembership{
		Memberships: map[string]*components.FactionMemberInfo{
			"thieves_guild": {FactionID: "thieves_guild", Rank: 5},
		},
	})

	quests := cqs.GenerateQuestsForFaction("thieves_guild", 5)
	cqs.AcceptQuest(w, quests[0].ID, uint64(player))

	history := cqs.GetPlayerQuestHistory(uint64(player))
	if len(history) != 1 {
		t.Errorf("History has %d quests, want 1", len(history))
	}
}

func TestCriminalFactionQuestSystem_GenreTitles(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}
	frs := NewFactionRankSystem("fantasy")

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			cqs := NewCriminalFactionQuestSystem(frs, genre, 12345)
			quests := cqs.GenerateQuestsForFaction("test_faction", 10)

			for _, quest := range quests {
				if quest.Title == "" {
					t.Errorf("Quest %s has empty title for genre %s", quest.ID, genre)
				}
			}
		})
	}
}

func TestCriminalFactionQuestSystem_isStealthyQuest(t *testing.T) {
	frs := NewFactionRankSystem("fantasy")
	cqs := NewCriminalFactionQuestSystem(frs, "fantasy", 12345)

	stealthyTypes := []CriminalQuestType{
		CriminalQuestTheft, CriminalQuestSmuggling, CriminalQuestInfiltration, CriminalQuestHeist,
	}
	nonStealthyTypes := []CriminalQuestType{
		CriminalQuestExtortion, CriminalQuestAssassination, CriminalQuestSabotage,
		CriminalQuestJailbreak, CriminalQuestTurf, CriminalQuestBoss,
	}

	for _, qt := range stealthyTypes {
		if !cqs.isStealthyQuest(qt) {
			t.Errorf("Quest type %s should be stealthy", qt.String())
		}
	}

	for _, qt := range nonStealthyTypes {
		if cqs.isStealthyQuest(qt) {
			t.Errorf("Quest type %s should not be stealthy", qt.String())
		}
	}
}

func TestCriminalFactionQuestSystem_getAvailableQuestTypes(t *testing.T) {
	frs := NewFactionRankSystem("fantasy")
	cqs := NewCriminalFactionQuestSystem(frs, "fantasy", 12345)

	// Rank 1 should only get theft
	types := cqs.getAvailableQuestTypes(1)
	found := false
	for _, qt := range types {
		if qt == CriminalQuestTheft {
			found = true
		}
		if qt.MinRank() > 1 {
			t.Errorf("Rank 1 got quest requiring rank %d", qt.MinRank())
		}
	}
	if !found {
		t.Error("Rank 1 should have theft available")
	}

	// Rank 10 should have all types
	types = cqs.getAvailableQuestTypes(10)
	if len(types) < 10 {
		t.Errorf("Rank 10 should have all 10 quest types, got %d", len(types))
	}
}

func TestFormatQuestID(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "CQ-0"},
		{1, "CQ-1"},
		{123, "CQ-123"},
		{99999, "CQ-99999"},
	}

	for _, tt := range tests {
		got := formatQuestID(tt.n)
		if got != tt.want {
			t.Errorf("formatQuestID(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

func TestCriminalFactionQuestSystem_pseudoRandom(t *testing.T) {
	frs := NewFactionRankSystem("fantasy")
	cqs1 := NewCriminalFactionQuestSystem(frs, "fantasy", 12345)
	cqs2 := NewCriminalFactionQuestSystem(frs, "fantasy", 12345)

	// Test determinism
	for i := 0; i < 10; i++ {
		v1 := cqs1.pseudoRandom()
		v2 := cqs2.pseudoRandom()
		if v1 != v2 {
			t.Errorf("pseudoRandom not deterministic at %d: %v vs %v", i, v1, v2)
		}
		if v1 < 0 || v1 > 1 {
			t.Errorf("pseudoRandom out of range: %v", v1)
		}
	}
}
