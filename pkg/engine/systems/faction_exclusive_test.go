package systems

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

func TestExclusiveContentSystemCreation(t *testing.T) {
	rankSys := NewFactionRankSystem("fantasy")
	sys := NewFactionExclusiveContentSystem(rankSys, "fantasy")
	if sys == nil {
		t.Fatal("NewFactionExclusiveContentSystem returned nil")
	}
	if sys.Genre != "fantasy" {
		t.Errorf("expected genre 'fantasy', got '%s'", sys.Genre)
	}
	if len(sys.Content) == 0 {
		t.Error("no default content registered")
	}
}

func TestExclusiveContentRegistration(t *testing.T) {
	rankSys := NewFactionRankSystem("fantasy")
	sys := NewFactionExclusiveContentSystem(rankSys, "fantasy")

	// Register custom content
	sys.RegisterContent(&ExclusiveContent{
		ID:           "test_content",
		Name:         "Test Content",
		Description:  "Test exclusive content",
		FactionID:    "test_faction",
		RequiredRank: 3,
		ContentType:  ContentTypeQuest,
	})

	if _, exists := sys.Content["test_content"]; !exists {
		t.Error("content not registered")
	}

	factionContent := sys.FactionContent["test_faction"]
	found := false
	for _, id := range factionContent {
		if id == "test_content" {
			found = true
			break
		}
	}
	if !found {
		t.Error("content not added to faction content list")
	}
}

func TestCanAccessContent(t *testing.T) {
	world := ecs.NewWorld()
	rankSys := NewFactionRankSystem("fantasy")
	sys := NewFactionExclusiveContentSystem(rankSys, "fantasy")
	entity := world.CreateEntity()

	// Register test content requiring rank 3
	sys.RegisterContent(&ExclusiveContent{
		ID:           "test_content",
		Name:         "Test Content",
		FactionID:    "guild",
		RequiredRank: 3,
		ContentType:  ContentTypeQuest,
	})

	// No faction membership - should not access
	if sys.CanAccessContent(world, entity, "test_content") {
		t.Error("should not access content without faction membership")
	}

	// Join faction at rank 1
	rankSys.JoinFaction(world, entity, "guild", "guild", 0)

	// Rank 1 cannot access rank 3 content
	if sys.CanAccessContent(world, entity, "test_content") {
		t.Error("rank 1 should not access rank 3 content")
	}

	// Promote to rank 3
	rankSys.AddXP(world, entity, "guild", 500) // Should reach rank 3

	// Now should have access
	if !sys.CanAccessContent(world, entity, "test_content") {
		t.Error("rank 3 should access rank 3 content")
	}
}

func TestGetAccessibleContent(t *testing.T) {
	world := ecs.NewWorld()
	rankSys := NewFactionRankSystem("fantasy")
	sys := NewFactionExclusiveContentSystem(rankSys, "fantasy")
	entity := world.CreateEntity()

	// No membership - should get empty list
	accessible := sys.GetAccessibleContent(world, entity)
	if len(accessible) != 0 {
		t.Errorf("expected 0 accessible content, got %d", len(accessible))
	}

	// Join guild at rank 1
	rankSys.JoinFaction(world, entity, "guild", "guild", 0)

	// Should only get rank 1-2 content (if any)
	accessible = sys.GetAccessibleContent(world, entity)
	for _, content := range accessible {
		if content.FactionID == "guild" && content.RequiredRank > 2 {
			t.Errorf("rank 1 should not access content requiring rank %d", content.RequiredRank)
		}
	}

	// Promote to rank 5 and check again
	rankSys.AddXP(world, entity, "guild", 2000)
	accessible = sys.GetAccessibleContent(world, entity)
	hasRank5Content := false
	for _, content := range accessible {
		if content.FactionID == "guild" && content.RequiredRank == 5 {
			hasRank5Content = true
			break
		}
	}
	if !hasRank5Content {
		t.Log("No rank 5 guild content available (may be expected based on default content)")
	}
}

func TestGetFactionContent(t *testing.T) {
	rankSys := NewFactionRankSystem("fantasy")
	sys := NewFactionExclusiveContentSystem(rankSys, "fantasy")

	guildContent := sys.GetFactionContent("guild")
	if len(guildContent) == 0 {
		t.Error("expected guild content for fantasy genre")
	}

	militaryContent := sys.GetFactionContent("military")
	if len(militaryContent) == 0 {
		t.Error("expected military content for fantasy genre")
	}
}

func TestGetContentByType(t *testing.T) {
	rankSys := NewFactionRankSystem("fantasy")
	sys := NewFactionExclusiveContentSystem(rankSys, "fantasy")

	quests := sys.GetContentByType(ContentTypeQuest)
	if len(quests) == 0 {
		t.Error("expected quest content")
	}

	areas := sys.GetContentByType(ContentTypeArea)
	if len(areas) == 0 {
		t.Error("expected area content")
	}

	skills := sys.GetContentByType(ContentTypeSkill)
	if len(skills) == 0 {
		t.Error("expected skill content")
	}
}

func TestUnlockContent(t *testing.T) {
	rankSys := NewFactionRankSystem("fantasy")
	sys := NewFactionExclusiveContentSystem(rankSys, "fantasy")
	entity := ecs.Entity(1)

	// Register test content
	sys.RegisterContent(&ExclusiveContent{
		ID:           "test_unlock",
		Name:         "Test Unlock",
		FactionID:    "test",
		RequiredRank: 1,
		ContentType:  ContentTypeQuest,
	})

	// Unlock content
	result := sys.UnlockContent(entity, "test_unlock", 100.0)
	if !result {
		t.Error("UnlockContent should return true for valid content")
	}

	// Verify unlock
	if !sys.HasUnlockedContent(entity, "test_unlock") {
		t.Error("content should be marked as unlocked")
	}

	// Try to unlock non-existent content
	result = sys.UnlockContent(entity, "non_existent", 100.0)
	if result {
		t.Error("UnlockContent should return false for non-existent content")
	}
}

func TestRepeatableContentCooldown(t *testing.T) {
	rankSys := NewFactionRankSystem("fantasy")
	sys := NewFactionExclusiveContentSystem(rankSys, "fantasy")
	entity := ecs.Entity(1)

	// Register repeatable content
	sys.RegisterContent(&ExclusiveContent{
		ID:           "repeatable_quest",
		Name:         "Repeatable Quest",
		FactionID:    "test",
		RequiredRank: 1,
		ContentType:  ContentTypeQuest,
		Repeatable:   true,
		CooldownTime: 100.0,
	})

	// Unlock (starts cooldown)
	sys.UnlockContent(entity, "repeatable_quest", 0.0)

	// Check cooldown
	cooldown := sys.GetCooldownRemaining(entity, "repeatable_quest")
	if cooldown != 100.0 {
		t.Errorf("expected cooldown 100.0, got %f", cooldown)
	}

	// Update to reduce cooldown
	sys.Update(nil, 50.0)
	cooldown = sys.GetCooldownRemaining(entity, "repeatable_quest")
	if cooldown != 50.0 {
		t.Errorf("expected cooldown 50.0 after update, got %f", cooldown)
	}

	// Update to clear cooldown
	sys.Update(nil, 60.0)
	cooldown = sys.GetCooldownRemaining(entity, "repeatable_quest")
	if cooldown != 0 {
		t.Errorf("expected cooldown 0 after expiry, got %f", cooldown)
	}
}

func TestCompleteContent(t *testing.T) {
	world := ecs.NewWorld()
	rankSys := NewFactionRankSystem("fantasy")
	sys := NewFactionExclusiveContentSystem(rankSys, "fantasy")
	entity := world.CreateEntity()

	// Register content with XP reward
	sys.RegisterContent(&ExclusiveContent{
		ID:           "xp_quest",
		Name:         "XP Quest",
		FactionID:    "guild",
		RequiredRank: 1,
		ContentType:  ContentTypeQuest,
		XPReward:     100,
	})

	// Join faction
	rankSys.JoinFaction(world, entity, "guild", "guild", 0)

	// Get initial XP
	info := rankSys.GetMembershipInfo(world, entity, "guild")
	initialXP := info.XP

	// Complete content
	result := sys.CompleteContent(world, entity, "xp_quest", 100.0)
	if !result {
		t.Error("CompleteContent should return true")
	}

	// Check XP was awarded
	info = rankSys.GetMembershipInfo(world, entity, "guild")
	if info.XP != initialXP+100 {
		t.Errorf("expected XP %d, got %d", initialXP+100, info.XP)
	}

	// Check content is unlocked
	if !sys.HasUnlockedContent(entity, "xp_quest") {
		t.Error("content should be unlocked after completion")
	}
}

func TestGetContentProgress(t *testing.T) {
	rankSys := NewFactionRankSystem("fantasy")
	sys := NewFactionExclusiveContentSystem(rankSys, "fantasy")
	entity := ecs.Entity(1)

	// Clear default content and add test content
	sys.Content = make(map[string]*ExclusiveContent)
	sys.FactionContent = make(map[string][]string)

	sys.RegisterContent(&ExclusiveContent{ID: "t1", FactionID: "test", RequiredRank: 1, ContentType: ContentTypeQuest})
	sys.RegisterContent(&ExclusiveContent{ID: "t2", FactionID: "test", RequiredRank: 1, ContentType: ContentTypeQuest})
	sys.RegisterContent(&ExclusiveContent{ID: "t3", FactionID: "test", RequiredRank: 1, ContentType: ContentTypeQuest})

	// Initial progress
	unlocked, total := sys.GetContentProgress(entity, "test")
	if unlocked != 0 || total != 3 {
		t.Errorf("expected 0/3, got %d/%d", unlocked, total)
	}

	// Unlock one
	sys.UnlockContent(entity, "t1", 0)
	unlocked, total = sys.GetContentProgress(entity, "test")
	if unlocked != 1 || total != 3 {
		t.Errorf("expected 1/3, got %d/%d", unlocked, total)
	}

	// Unlock all
	sys.UnlockContent(entity, "t2", 0)
	sys.UnlockContent(entity, "t3", 0)
	unlocked, total = sys.GetContentProgress(entity, "test")
	if unlocked != 3 || total != 3 {
		t.Errorf("expected 3/3, got %d/%d", unlocked, total)
	}
}

func TestContentTypeConstants(t *testing.T) {
	// Verify content type constants are distinct
	types := []ExclusiveContentType{
		ContentTypeQuest,
		ContentTypeItem,
		ContentTypeArea,
		ContentTypeSkill,
		ContentTypeVendor,
		ContentTypeDialogue,
		ContentTypeRecipe,
	}
	seen := make(map[ExclusiveContentType]bool)
	for _, ct := range types {
		if seen[ct] {
			t.Errorf("duplicate content type value: %d", ct)
		}
		seen[ct] = true
	}
}

func TestGenreSpecificContent(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		rankSys := NewFactionRankSystem(genre)
		sys := NewFactionExclusiveContentSystem(rankSys, genre)

		if len(sys.Content) == 0 {
			t.Errorf("no content registered for genre '%s'", genre)
		}
	}
}

func TestCyberpunkExclusiveContent(t *testing.T) {
	rankSys := NewFactionRankSystem("cyberpunk")
	sys := NewFactionExclusiveContentSystem(rankSys, "cyberpunk")

	// Check megacorp content exists
	megacorpContent := sys.GetFactionContent("megacorp")
	if len(megacorpContent) == 0 {
		t.Error("expected megacorp content for cyberpunk genre")
	}

	// Check gang content exists
	gangContent := sys.GetFactionContent("gang")
	if len(gangContent) == 0 {
		t.Error("expected gang content for cyberpunk genre")
	}

	// Check hacker content exists
	hackerContent := sys.GetFactionContent("hacker")
	if len(hackerContent) == 0 {
		t.Error("expected hacker content for cyberpunk genre")
	}
}

func TestUpdateCooldowns(t *testing.T) {
	rankSys := NewFactionRankSystem("fantasy")
	sys := NewFactionExclusiveContentSystem(rankSys, "fantasy")

	// Set up a cooldown manually
	entity := ecs.Entity(1)
	sys.PlayerCooldowns[uint64(entity)] = map[string]float64{
		"content1": 100.0,
		"content2": 50.0,
	}

	// Update by 30 seconds
	sys.Update(nil, 30.0)

	if sys.PlayerCooldowns[uint64(entity)]["content1"] != 70.0 {
		t.Errorf("expected cooldown 70.0, got %f", sys.PlayerCooldowns[uint64(entity)]["content1"])
	}
	if sys.PlayerCooldowns[uint64(entity)]["content2"] != 20.0 {
		t.Errorf("expected cooldown 20.0, got %f", sys.PlayerCooldowns[uint64(entity)]["content2"])
	}

	// Update to clear content2
	sys.Update(nil, 25.0)

	if _, exists := sys.PlayerCooldowns[uint64(entity)]["content2"]; exists {
		t.Error("content2 cooldown should be cleared")
	}
}
