// Package systems implements all ECS game systems.
package systems

import (
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// ExclusiveContentType defines the type of exclusive content.
type ExclusiveContentType int

const (
	// ContentTypeQuest is a faction-exclusive quest.
	ContentTypeQuest ExclusiveContentType = iota
	// ContentTypeItem is a faction-exclusive item or equipment.
	ContentTypeItem
	// ContentTypeArea is a faction-exclusive area or location.
	ContentTypeArea
	// ContentTypeSkill is a faction-exclusive skill or ability.
	ContentTypeSkill
	// ContentTypeVendor is a faction-exclusive vendor or shop.
	ContentTypeVendor
	// ContentTypeDialogue is faction-exclusive dialogue options.
	ContentTypeDialogue
	// ContentTypeRecipe is a faction-exclusive crafting recipe.
	ContentTypeRecipe
)

// ExclusiveContent represents content that requires faction membership.
type ExclusiveContent struct {
	ID           string
	Name         string
	Description  string
	FactionID    string
	RequiredRank int
	ContentType  ExclusiveContentType
	// Rewards given when content is completed/acquired
	XPReward    int
	ItemRewards []string
	// For time-limited content
	AvailableFrom  float64 // 0 = always available
	AvailableUntil float64 // 0 = no end time
	// For repeatable content
	Repeatable   bool
	CooldownTime float64 // Time before can repeat (seconds)
}

// FactionExclusiveContentSystem manages faction-exclusive content access.
type FactionExclusiveContentSystem struct {
	// Content maps content ID to definition
	Content map[string]*ExclusiveContent
	// FactionContent maps faction ID to list of content IDs
	FactionContent map[string][]string
	// PlayerUnlocks maps player entity -> content ID -> unlock time
	PlayerUnlocks map[uint64]map[string]float64
	// PlayerCooldowns maps player entity -> content ID -> cooldown end time
	PlayerCooldowns map[uint64]map[string]float64
	// RankSystem reference for rank checking
	RankSystem *FactionRankSystem
	// Genre for content theming
	Genre string
}

// NewFactionExclusiveContentSystem creates a new exclusive content system.
func NewFactionExclusiveContentSystem(rankSystem *FactionRankSystem, genre string) *FactionExclusiveContentSystem {
	sys := &FactionExclusiveContentSystem{
		Content:         make(map[string]*ExclusiveContent),
		FactionContent:  make(map[string][]string),
		PlayerUnlocks:   make(map[uint64]map[string]float64),
		PlayerCooldowns: make(map[uint64]map[string]float64),
		RankSystem:      rankSystem,
		Genre:           genre,
	}
	sys.registerDefaultContent()
	return sys
}

// registerDefaultContent adds genre-appropriate exclusive content.
func (s *FactionExclusiveContentSystem) registerDefaultContent() {
	switch s.Genre {
	case "fantasy":
		s.registerFantasyContent()
	case "sci-fi":
		s.registerSciFiContent()
	case "horror":
		s.registerHorrorContent()
	case "cyberpunk":
		s.registerCyberpunkContent()
	case "post-apocalyptic":
		s.registerPostApocContent()
	default:
		s.registerFantasyContent()
	}
}

func (s *FactionExclusiveContentSystem) registerFantasyContent() {
	// Guild content
	s.RegisterContent(&ExclusiveContent{
		ID:           "guild_armory",
		Name:         "Guild Armory Access",
		Description:  "Access to the guild's exclusive armory with enchanted weapons.",
		FactionID:    "guild",
		RequiredRank: 3,
		ContentType:  ContentTypeVendor,
	})
	s.RegisterContent(&ExclusiveContent{
		ID:           "guild_quest_artifact",
		Name:         "Artifact Recovery",
		Description:  "A quest to recover a lost guild artifact from ancient ruins.",
		FactionID:    "guild",
		RequiredRank: 5,
		ContentType:  ContentTypeQuest,
		XPReward:     500,
	})
	s.RegisterContent(&ExclusiveContent{
		ID:           "guild_master_training",
		Name:         "Master's Training",
		Description:  "Learn advanced techniques from the guild masters.",
		FactionID:    "guild",
		RequiredRank: 7,
		ContentType:  ContentTypeSkill,
		XPReward:     750,
	})

	// Military content
	s.RegisterContent(&ExclusiveContent{
		ID:           "military_war_room",
		Name:         "War Room Briefings",
		Description:  "Access to strategic information and military intelligence.",
		FactionID:    "military",
		RequiredRank: 4,
		ContentType:  ContentTypeArea,
	})
	s.RegisterContent(&ExclusiveContent{
		ID:           "military_siege_quest",
		Name:         "Siege Commander",
		Description:  "Lead a siege against an enemy fortress.",
		FactionID:    "military",
		RequiredRank: 6,
		ContentType:  ContentTypeQuest,
		XPReward:     600,
	})

	// Religious content
	s.RegisterContent(&ExclusiveContent{
		ID:           "religious_sanctuary",
		Name:         "Inner Sanctuary",
		Description:  "Access to the sacred inner sanctuary and its blessings.",
		FactionID:    "religious",
		RequiredRank: 4,
		ContentType:  ContentTypeArea,
	})
	s.RegisterContent(&ExclusiveContent{
		ID:           "religious_miracle",
		Name:         "Perform Miracles",
		Description:  "Learn to perform healing miracles.",
		FactionID:    "religious",
		RequiredRank: 8,
		ContentType:  ContentTypeSkill,
		XPReward:     1000,
	})
}

func (s *FactionExclusiveContentSystem) registerSciFiContent() {
	// Corporation content
	s.RegisterContent(&ExclusiveContent{
		ID:           "corp_executive_lounge",
		Name:         "Executive Lounge",
		Description:  "Access to the exclusive executive lounge with premium amenities.",
		FactionID:    "corporation",
		RequiredRank: 5,
		ContentType:  ContentTypeArea,
	})
	s.RegisterContent(&ExclusiveContent{
		ID:           "corp_prototype_access",
		Name:         "Prototype Testing",
		Description:  "Test experimental corporate prototypes before release.",
		FactionID:    "corporation",
		RequiredRank: 6,
		ContentType:  ContentTypeItem,
		XPReward:     400,
	})

	// Military content
	s.RegisterContent(&ExclusiveContent{
		ID:           "military_black_ops",
		Name:         "Black Operations",
		Description:  "Classified missions behind enemy lines.",
		FactionID:    "military",
		RequiredRank: 7,
		ContentType:  ContentTypeQuest,
		XPReward:     800,
	})
}

func (s *FactionExclusiveContentSystem) registerHorrorContent() {
	// Cult content
	s.RegisterContent(&ExclusiveContent{
		ID:           "cult_forbidden_texts",
		Name:         "Forbidden Knowledge",
		Description:  "Study the forbidden texts in the cult's secret library.",
		FactionID:    "cult",
		RequiredRank: 4,
		ContentType:  ContentTypeSkill,
		XPReward:     300,
	})
	s.RegisterContent(&ExclusiveContent{
		ID:           "cult_ritual_chamber",
		Name:         "Ritual Chamber Access",
		Description:  "Participate in dark rituals within the inner chamber.",
		FactionID:    "cult",
		RequiredRank: 6,
		ContentType:  ContentTypeArea,
	})

	// Survivor content
	s.RegisterContent(&ExclusiveContent{
		ID:           "survivor_bunker",
		Name:         "Secret Bunker",
		Description:  "Access to the survivors' hidden emergency bunker.",
		FactionID:    "survivor",
		RequiredRank: 5,
		ContentType:  ContentTypeArea,
	})
}

func (s *FactionExclusiveContentSystem) registerCyberpunkContent() {
	// Megacorp content
	s.RegisterContent(&ExclusiveContent{
		ID:           "megacorp_augments",
		Name:         "Premium Augmentations",
		Description:  "Access to cutting-edge corporate cybernetic augmentations.",
		FactionID:    "megacorp",
		RequiredRank: 4,
		ContentType:  ContentTypeItem,
		XPReward:     350,
	})
	s.RegisterContent(&ExclusiveContent{
		ID:           "megacorp_penthouse",
		Name:         "Corporate Penthouse",
		Description:  "Luxurious living quarters in the corporate tower.",
		FactionID:    "megacorp",
		RequiredRank: 8,
		ContentType:  ContentTypeArea,
	})

	// Gang content
	s.RegisterContent(&ExclusiveContent{
		ID:           "gang_turf",
		Name:         "Gang Territory",
		Description:  "Free passage and protection in gang-controlled territory.",
		FactionID:    "gang",
		RequiredRank: 3,
		ContentType:  ContentTypeArea,
	})
	s.RegisterContent(&ExclusiveContent{
		ID:           "gang_weapons",
		Name:         "Street Hardware",
		Description:  "Access to illegal military-grade weapons.",
		FactionID:    "gang",
		RequiredRank: 5,
		ContentType:  ContentTypeVendor,
	})

	// Hacker content
	s.RegisterContent(&ExclusiveContent{
		ID:           "hacker_darknet",
		Name:         "Darknet Access",
		Description:  "Access to hidden darknet nodes and resources.",
		FactionID:    "hacker",
		RequiredRank: 4,
		ContentType:  ContentTypeArea,
	})
	s.RegisterContent(&ExclusiveContent{
		ID:           "hacker_exploits",
		Name:         "Zero-Day Exploits",
		Description:  "Learn advanced hacking techniques and exploits.",
		FactionID:    "hacker",
		RequiredRank: 7,
		ContentType:  ContentTypeSkill,
		XPReward:     600,
	})
}

func (s *FactionExclusiveContentSystem) registerPostApocContent() {
	// Tribe content
	s.RegisterContent(&ExclusiveContent{
		ID:           "tribe_sacred_grounds",
		Name:         "Sacred Hunting Grounds",
		Description:  "Access to the tribe's sacred hunting grounds.",
		FactionID:    "tribe",
		RequiredRank: 4,
		ContentType:  ContentTypeArea,
	})
	s.RegisterContent(&ExclusiveContent{
		ID:           "tribe_rituals",
		Name:         "Warrior Rituals",
		Description:  "Participate in ancient warrior rituals for strength.",
		FactionID:    "tribe",
		RequiredRank: 6,
		ContentType:  ContentTypeSkill,
		XPReward:     500,
	})

	// Raider content
	s.RegisterContent(&ExclusiveContent{
		ID:           "raider_war_camp",
		Name:         "War Camp",
		Description:  "Access to the raider war camp and its resources.",
		FactionID:    "raider",
		RequiredRank: 3,
		ContentType:  ContentTypeArea,
	})
	s.RegisterContent(&ExclusiveContent{
		ID:           "raider_convoy_raid",
		Name:         "Convoy Raid",
		Description:  "Lead a raid on a merchant convoy.",
		FactionID:    "raider",
		RequiredRank: 5,
		ContentType:  ContentTypeQuest,
		XPReward:     450,
		Repeatable:   true,
		CooldownTime: 3600, // 1 hour
	})

	// Trader content
	s.RegisterContent(&ExclusiveContent{
		ID:           "trader_caravan_routes",
		Name:         "Caravan Routes",
		Description:  "Access to safe caravan trade routes.",
		FactionID:    "trader",
		RequiredRank: 4,
		ContentType:  ContentTypeArea,
	})
	s.RegisterContent(&ExclusiveContent{
		ID:           "trader_black_market",
		Name:         "Black Market",
		Description:  "Access to rare items in the trader's black market.",
		FactionID:    "trader",
		RequiredRank: 6,
		ContentType:  ContentTypeVendor,
	})
}

// RegisterContent adds exclusive content to the system.
func (s *FactionExclusiveContentSystem) RegisterContent(content *ExclusiveContent) {
	s.Content[content.ID] = content
	s.FactionContent[content.FactionID] = append(s.FactionContent[content.FactionID], content.ID)
}

// Update processes exclusive content system each tick.
func (s *FactionExclusiveContentSystem) Update(w *ecs.World, dt float64) {
	// Process cooldown timers
	for _, cooldowns := range s.PlayerCooldowns {
		for contentID, endTime := range cooldowns {
			if endTime > 0 {
				cooldowns[contentID] = endTime - dt
				if cooldowns[contentID] <= 0 {
					delete(cooldowns, contentID)
				}
			}
		}
	}
}

// CanAccessContent checks if a player can access exclusive content.
func (s *FactionExclusiveContentSystem) CanAccessContent(w *ecs.World, entity ecs.Entity, contentID string) bool {
	content, exists := s.Content[contentID]
	if !exists {
		return false
	}
	if !s.hasRequiredFactionRank(w, entity, content) {
		return false
	}
	return !s.isOnCooldown(entity, content, contentID)
}

// hasRequiredFactionRank checks if entity has required faction membership and rank.
func (s *FactionExclusiveContentSystem) hasRequiredFactionRank(w *ecs.World, entity ecs.Entity, content *ExclusiveContent) bool {
	if s.RankSystem == nil {
		return false
	}
	return s.RankSystem.CanAccessRankContent(w, entity, content.FactionID, content.RequiredRank)
}

// isOnCooldown checks if repeatable content is on cooldown for the entity.
func (s *FactionExclusiveContentSystem) isOnCooldown(entity ecs.Entity, content *ExclusiveContent, contentID string) bool {
	if !content.Repeatable {
		return false
	}
	cooldowns := s.PlayerCooldowns[uint64(entity)]
	if cooldowns == nil {
		return false
	}
	cooldownEnd, onCooldown := cooldowns[contentID]
	return onCooldown && cooldownEnd > 0
}

// GetAccessibleContent returns all content a player can currently access.
func (s *FactionExclusiveContentSystem) GetAccessibleContent(w *ecs.World, entity ecs.Entity) []*ExclusiveContent {
	accessible := make([]*ExclusiveContent, 0)
	for contentID := range s.Content {
		if s.CanAccessContent(w, entity, contentID) {
			accessible = append(accessible, s.Content[contentID])
		}
	}
	return accessible
}

// GetFactionContent returns all exclusive content for a faction.
func (s *FactionExclusiveContentSystem) GetFactionContent(factionID string) []*ExclusiveContent {
	contentIDs := s.FactionContent[factionID]
	content := make([]*ExclusiveContent, 0, len(contentIDs))
	for _, id := range contentIDs {
		if c, exists := s.Content[id]; exists {
			content = append(content, c)
		}
	}
	return content
}

// GetContentByType returns exclusive content of a specific type.
func (s *FactionExclusiveContentSystem) GetContentByType(contentType ExclusiveContentType) []*ExclusiveContent {
	content := make([]*ExclusiveContent, 0)
	for _, c := range s.Content {
		if c.ContentType == contentType {
			content = append(content, c)
		}
	}
	return content
}

// UnlockContent marks content as unlocked for a player.
func (s *FactionExclusiveContentSystem) UnlockContent(entity ecs.Entity, contentID string, gameTime float64) bool {
	content, exists := s.Content[contentID]
	if !exists {
		return false
	}

	entityID := uint64(entity)
	if s.PlayerUnlocks[entityID] == nil {
		s.PlayerUnlocks[entityID] = make(map[string]float64)
	}
	s.PlayerUnlocks[entityID][contentID] = gameTime

	// Start cooldown for repeatable content
	if content.Repeatable && content.CooldownTime > 0 {
		if s.PlayerCooldowns[entityID] == nil {
			s.PlayerCooldowns[entityID] = make(map[string]float64)
		}
		s.PlayerCooldowns[entityID][contentID] = content.CooldownTime
	}

	return true
}

// HasUnlockedContent checks if a player has ever unlocked content.
func (s *FactionExclusiveContentSystem) HasUnlockedContent(entity ecs.Entity, contentID string) bool {
	unlocks := s.PlayerUnlocks[uint64(entity)]
	if unlocks == nil {
		return false
	}
	_, unlocked := unlocks[contentID]
	return unlocked
}

// GetContentProgress returns content completion stats for a player.
func (s *FactionExclusiveContentSystem) GetContentProgress(entity ecs.Entity, factionID string) (unlocked, total int) {
	factionContent := s.FactionContent[factionID]
	total = len(factionContent)
	unlocks := s.PlayerUnlocks[uint64(entity)]
	if unlocks == nil {
		return 0, total
	}
	for _, contentID := range factionContent {
		if _, ok := unlocks[contentID]; ok {
			unlocked++
		}
	}
	return unlocked, total
}

// CompleteContent handles completing exclusive content and awards rewards.
func (s *FactionExclusiveContentSystem) CompleteContent(w *ecs.World, entity ecs.Entity, contentID string, gameTime float64) bool {
	content, exists := s.Content[contentID]
	if !exists {
		return false
	}

	// Check access
	if !s.CanAccessContent(w, entity, contentID) {
		return false
	}

	// Award XP reward
	if content.XPReward > 0 && s.RankSystem != nil {
		s.RankSystem.AddXP(w, entity, content.FactionID, content.XPReward)
	}

	// Mark as unlocked
	s.UnlockContent(entity, contentID, gameTime)

	// Track in membership
	comp, ok := w.GetComponent(entity, "FactionMembership")
	if ok {
		membership := comp.(*components.FactionMembership)
		info := membership.GetMembership(content.FactionID)
		if info != nil && info.UnlockedContent == nil {
			info.UnlockedContent = make(map[string]bool)
		}
		if info != nil {
			info.UnlockedContent[contentID] = true
		}
	}

	return true
}

// GetCooldownRemaining returns remaining cooldown time for repeatable content.
func (s *FactionExclusiveContentSystem) GetCooldownRemaining(entity ecs.Entity, contentID string) float64 {
	cooldowns := s.PlayerCooldowns[uint64(entity)]
	if cooldowns == nil {
		return 0
	}
	return cooldowns[contentID]
}
