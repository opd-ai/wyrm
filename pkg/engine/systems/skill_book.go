package systems

import (
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// ============================================================================
// Skill Book System
// ============================================================================

// SkillBook represents a learnable tome that grants skill experience or levels.
type SkillBook struct {
	// ID uniquely identifies this skill book.
	ID string
	// Name is the display name of the book.
	Name string
	// Description explains what knowledge the book imparts.
	Description string
	// SkillID is the skill this book trains.
	SkillID string
	// XPGrant is the flat XP amount granted when read (0 if using level grant).
	XPGrant float64
	// LevelGrant is the number of levels granted when read (0 if using XP grant).
	LevelGrant int
	// RequiredLevel is the minimum skill level needed to comprehend the book.
	RequiredLevel int
	// MaxLevel is the maximum skill level at which the book is useful.
	MaxLevel int
	// Rarity determines how often the book spawns (common=100, rare=10, legendary=1).
	Rarity int
	// Genre specifies which genre variant this book belongs to ("" for universal).
	Genre string
	// IsOneTimeUse determines if the book is consumed on use.
	IsOneTimeUse bool
	// ReadTime is the in-game time (seconds) required to read the book.
	ReadTime float64
}

// SkillBookSystem manages skill book inventory and usage.
type SkillBookSystem struct {
	// Books is the registry of all available skill books.
	Books map[string]*SkillBook
	// PlayerBooks tracks which books each player has read (entityID -> bookID -> true).
	PlayerBooks map[uint64]map[string]bool
	// ActiveReaders tracks players currently reading books (entityID -> bookID, timeRemaining).
	ActiveReaders map[uint64]*ActiveReading
	// SkillRegistry provides skill metadata.
	SkillRegistry *SkillRegistry
	// ProgressionSystem handles XP grants.
	ProgressionSystem *SkillProgressionSystem
}

// ActiveReading tracks a book being actively read.
type ActiveReading struct {
	BookID        string
	TimeRemaining float64
}

// NewSkillBookSystem creates a new skill book system with default books.
func NewSkillBookSystem(registry *SkillRegistry, progression *SkillProgressionSystem) *SkillBookSystem {
	s := &SkillBookSystem{
		Books:             make(map[string]*SkillBook),
		PlayerBooks:       make(map[uint64]map[string]bool),
		ActiveReaders:     make(map[uint64]*ActiveReading),
		SkillRegistry:     registry,
		ProgressionSystem: progression,
	}
	s.registerDefaultBooks()
	return s
}

// registerDefaultBooks populates the skill book registry with starter books.
func (s *SkillBookSystem) registerDefaultBooks() {
	// Combat skill books
	s.RegisterBook(&SkillBook{
		ID: "tome_melee_basics", Name: "Warrior's Primer", Description: "Basic melee techniques",
		SkillID: "melee", XPGrant: 100, RequiredLevel: 0, MaxLevel: 20, Rarity: 100, IsOneTimeUse: true, ReadTime: 10,
	})
	s.RegisterBook(&SkillBook{
		ID: "tome_melee_advanced", Name: "Art of the Blade", Description: "Advanced swordplay",
		SkillID: "melee", XPGrant: 300, RequiredLevel: 15, MaxLevel: 50, Rarity: 30, IsOneTimeUse: true, ReadTime: 30,
	})
	s.RegisterBook(&SkillBook{
		ID: "tome_melee_master", Name: "Sword Saint's Treatise", Description: "Legendary combat wisdom",
		SkillID: "melee", LevelGrant: 2, RequiredLevel: 50, MaxLevel: 90, Rarity: 5, IsOneTimeUse: true, ReadTime: 60,
	})
	s.RegisterBook(&SkillBook{
		ID: "tome_ranged_basics", Name: "Archer's Handbook", Description: "Fundamentals of ranged combat",
		SkillID: "ranged", XPGrant: 100, RequiredLevel: 0, MaxLevel: 20, Rarity: 100, IsOneTimeUse: true, ReadTime: 10,
	})
	s.RegisterBook(&SkillBook{
		ID: "tome_ranged_advanced", Name: "Marksman's Manual", Description: "Precision shooting techniques",
		SkillID: "ranged", XPGrant: 300, RequiredLevel: 15, MaxLevel: 50, Rarity: 30, IsOneTimeUse: true, ReadTime: 30,
	})
	s.RegisterBook(&SkillBook{
		ID: "tome_defense_basics", Name: "Shield Bearer's Guide", Description: "Basic defensive stances",
		SkillID: "defense", XPGrant: 100, RequiredLevel: 0, MaxLevel: 20, Rarity: 100, IsOneTimeUse: true, ReadTime: 10,
	})

	// Stealth skill books
	s.RegisterBook(&SkillBook{
		ID: "tome_sneak_basics", Name: "Shadow Walker's Primer", Description: "Basic sneaking techniques",
		SkillID: "sneak", XPGrant: 100, RequiredLevel: 0, MaxLevel: 20, Rarity: 80, IsOneTimeUse: true, ReadTime: 10,
	})
	s.RegisterBook(&SkillBook{
		ID: "tome_sneak_advanced", Name: "Art of Invisibility", Description: "Advanced concealment",
		SkillID: "sneak", XPGrant: 300, RequiredLevel: 15, MaxLevel: 50, Rarity: 25, IsOneTimeUse: true, ReadTime: 30,
	})
	s.RegisterBook(&SkillBook{
		ID: "tome_lockpicking", Name: "Locksmith's Secrets", Description: "Mechanical lock manipulation",
		SkillID: "lockpicking", XPGrant: 150, RequiredLevel: 0, MaxLevel: 30, Rarity: 50, IsOneTimeUse: true, ReadTime: 15,
	})
	s.RegisterBook(&SkillBook{
		ID: "tome_pickpocket", Name: "Light Fingers", Description: "The art of pilfering",
		SkillID: "pickpocket", XPGrant: 150, RequiredLevel: 0, MaxLevel: 30, Rarity: 40, IsOneTimeUse: true, ReadTime: 15,
	})

	// Magic skill books
	s.RegisterBook(&SkillBook{
		ID: "tome_destruction_basics", Name: "Elemental Primer", Description: "Basic destructive magic",
		SkillID: "destruction", XPGrant: 100, RequiredLevel: 0, MaxLevel: 20, Rarity: 70, IsOneTimeUse: true, ReadTime: 15,
	})
	s.RegisterBook(&SkillBook{
		ID: "tome_destruction_advanced", Name: "Pyromancer's Codex", Description: "Advanced fire magic",
		SkillID: "destruction", XPGrant: 350, RequiredLevel: 20, MaxLevel: 60, Rarity: 20, IsOneTimeUse: true, ReadTime: 45,
	})
	s.RegisterBook(&SkillBook{
		ID: "tome_restoration", Name: "Healer's Compendium", Description: "Restorative magic techniques",
		SkillID: "restoration", XPGrant: 150, RequiredLevel: 0, MaxLevel: 35, Rarity: 60, IsOneTimeUse: true, ReadTime: 20,
	})
	s.RegisterBook(&SkillBook{
		ID: "tome_conjuration", Name: "Summoner's Grimoire", Description: "Conjuration fundamentals",
		SkillID: "conjuration", XPGrant: 200, RequiredLevel: 5, MaxLevel: 40, Rarity: 40, IsOneTimeUse: true, ReadTime: 25,
	})
	s.RegisterBook(&SkillBook{
		ID: "tome_alteration", Name: "Reality Shaper's Guide", Description: "Altering physical laws",
		SkillID: "alteration", XPGrant: 150, RequiredLevel: 0, MaxLevel: 35, Rarity: 50, IsOneTimeUse: true, ReadTime: 20,
	})
	s.RegisterBook(&SkillBook{
		ID: "tome_enchanting", Name: "Enchanter's Handbook", Description: "Imbuing items with magic",
		SkillID: "enchanting", XPGrant: 200, RequiredLevel: 5, MaxLevel: 40, Rarity: 35, IsOneTimeUse: true, ReadTime: 30,
	})

	// Crafting skill books
	s.RegisterBook(&SkillBook{
		ID: "tome_smithing_basics", Name: "Blacksmith's Manual", Description: "Basic metalworking",
		SkillID: "smithing", XPGrant: 100, RequiredLevel: 0, MaxLevel: 20, Rarity: 90, IsOneTimeUse: true, ReadTime: 10,
	})
	s.RegisterBook(&SkillBook{
		ID: "tome_smithing_advanced", Name: "Master Smith's Secrets", Description: "Advanced forging techniques",
		SkillID: "smithing", XPGrant: 400, RequiredLevel: 30, MaxLevel: 70, Rarity: 15, IsOneTimeUse: true, ReadTime: 45,
	})
	s.RegisterBook(&SkillBook{
		ID: "tome_alchemy", Name: "Alchemist's Reference", Description: "Potion brewing basics",
		SkillID: "alchemy", XPGrant: 150, RequiredLevel: 0, MaxLevel: 30, Rarity: 70, IsOneTimeUse: true, ReadTime: 20,
	})
	s.RegisterBook(&SkillBook{
		ID: "tome_alchemy_advanced", Name: "Grand Alchemist's Formulae", Description: "Rare potion recipes",
		SkillID: "alchemy", LevelGrant: 3, RequiredLevel: 40, MaxLevel: 80, Rarity: 10, IsOneTimeUse: true, ReadTime: 60,
	})

	// Social skill books
	s.RegisterBook(&SkillBook{
		ID: "tome_speech", Name: "Orator's Guide", Description: "Persuasion techniques",
		SkillID: "speech", XPGrant: 100, RequiredLevel: 0, MaxLevel: 25, Rarity: 80, IsOneTimeUse: true, ReadTime: 15,
	})
	s.RegisterBook(&SkillBook{
		ID: "tome_leadership", Name: "Commander's Handbook", Description: "Leading others effectively",
		SkillID: "leadership", XPGrant: 200, RequiredLevel: 10, MaxLevel: 40, Rarity: 40, IsOneTimeUse: true, ReadTime: 25,
	})
	s.RegisterBook(&SkillBook{
		ID: "tome_barter", Name: "Merchant's Ledger", Description: "Trading strategies",
		SkillID: "barter", XPGrant: 150, RequiredLevel: 0, MaxLevel: 30, Rarity: 70, IsOneTimeUse: true, ReadTime: 15,
	})

	// Survival skill books
	s.RegisterBook(&SkillBook{
		ID: "tome_athletics", Name: "Athlete's Training Guide", Description: "Physical conditioning",
		SkillID: "athletics", XPGrant: 100, RequiredLevel: 0, MaxLevel: 20, Rarity: 90, IsOneTimeUse: true, ReadTime: 10,
	})
	s.RegisterBook(&SkillBook{
		ID: "tome_survival", Name: "Wilderness Survival Manual", Description: "Outdoor survival skills",
		SkillID: "survival", XPGrant: 150, RequiredLevel: 0, MaxLevel: 30, Rarity: 60, IsOneTimeUse: true, ReadTime: 20,
	})
	s.RegisterBook(&SkillBook{
		ID: "tome_herbalism", Name: "Herbalist's Codex", Description: "Identifying useful plants",
		SkillID: "herbalism", XPGrant: 150, RequiredLevel: 0, MaxLevel: 30, Rarity: 50, IsOneTimeUse: true, ReadTime: 20,
	})
}

// RegisterBook adds a book to the registry.
func (s *SkillBookSystem) RegisterBook(book *SkillBook) {
	if book == nil || book.ID == "" {
		return
	}
	s.Books[book.ID] = book
}

// GetBook retrieves a book by ID.
func (s *SkillBookSystem) GetBook(bookID string) *SkillBook {
	return s.Books[bookID]
}

// GetBooksForSkill returns all books that train a specific skill.
func (s *SkillBookSystem) GetBooksForSkill(skillID string) []*SkillBook {
	books := make([]*SkillBook, 0)
	for _, book := range s.Books {
		if book.SkillID == skillID {
			books = append(books, book)
		}
	}
	return books
}

// GetBooksByRarity returns books at or above a rarity threshold.
func (s *SkillBookSystem) GetBooksByRarity(minRarity, maxRarity int) []*SkillBook {
	books := make([]*SkillBook, 0)
	for _, book := range s.Books {
		if book.Rarity >= minRarity && book.Rarity <= maxRarity {
			books = append(books, book)
		}
	}
	return books
}

// Update processes active reading progress.
func (s *SkillBookSystem) Update(w *ecs.World, dt float64) {
	for _, e := range w.Entities("Skills") {
		entityID := uint64(e)
		reading, ok := s.ActiveReaders[entityID]
		if !ok {
			continue
		}

		reading.TimeRemaining -= dt
		if reading.TimeRemaining <= 0 {
			// Reading complete - apply book effects
			s.completeReading(w, e)
		}
	}
}

// StartReading begins reading a skill book.
func (s *SkillBookSystem) StartReading(w *ecs.World, entity ecs.Entity, bookID string) bool {
	entityID := uint64(entity)
	if s.isAlreadyReading(entityID) {
		return false
	}

	book := s.Books[bookID]
	if book == nil {
		return false
	}

	if !s.canReadBook(w, entity, entityID, book, bookID) {
		return false
	}

	s.beginReading(entityID, bookID, book)
	return true
}

// isAlreadyReading checks if entity is currently reading a book.
func (s *SkillBookSystem) isAlreadyReading(entityID uint64) bool {
	_, ok := s.ActiveReaders[entityID]
	return ok
}

// canReadBook validates all conditions for reading a book.
func (s *SkillBookSystem) canReadBook(w *ecs.World, entity ecs.Entity, entityID uint64, book *SkillBook, bookID string) bool {
	if book.IsOneTimeUse && s.hasAlreadyReadBook(entityID, bookID) {
		return false
	}
	return s.meetsSkillRequirements(w, entity, book)
}

// hasAlreadyReadBook checks if entity has already read a one-time-use book.
func (s *SkillBookSystem) hasAlreadyReadBook(entityID uint64, bookID string) bool {
	return s.PlayerBooks[entityID] != nil && s.PlayerBooks[entityID][bookID]
}

// meetsSkillRequirements checks if entity meets skill level requirements for a book.
func (s *SkillBookSystem) meetsSkillRequirements(w *ecs.World, entity ecs.Entity, book *SkillBook) bool {
	comp, ok := w.GetComponent(entity, "Skills")
	if !ok {
		return false
	}
	skills := comp.(*components.Skills)
	if skills.Levels == nil {
		return false
	}

	currentLevel := skills.Levels[book.SkillID]
	return currentLevel >= book.RequiredLevel && currentLevel < book.MaxLevel
}

// beginReading starts the reading session.
func (s *SkillBookSystem) beginReading(entityID uint64, bookID string, book *SkillBook) {
	s.ActiveReaders[entityID] = &ActiveReading{
		BookID:        bookID,
		TimeRemaining: book.ReadTime,
	}
}

// completeReading finishes reading and applies the book's benefits.
func (s *SkillBookSystem) completeReading(w *ecs.World, entity ecs.Entity) {
	entityID := uint64(entity)
	reading, ok := s.ActiveReaders[entityID]
	if !ok {
		return
	}
	defer delete(s.ActiveReaders, entityID)

	book := s.Books[reading.BookID]
	if book == nil {
		return
	}

	comp, ok := w.GetComponent(entity, "Skills")
	if !ok {
		return
	}
	skills := comp.(*components.Skills)

	s.applyBookBonus(skills, book)
	s.markBookAsRead(entityID, reading.BookID, book)
}

// applyBookBonus applies XP or level grant from a skill book.
func (s *SkillBookSystem) applyBookBonus(skills *components.Skills, book *SkillBook) {
	if book.LevelGrant > 0 {
		s.applyLevelGrant(skills, book)
	} else if book.XPGrant > 0 {
		s.applyXPGrant(skills, book)
	}
}

// applyLevelGrant directly increases skill levels.
func (s *SkillBookSystem) applyLevelGrant(skills *components.Skills, book *SkillBook) {
	if skills.Levels == nil {
		skills.Levels = make(map[string]int)
	}
	skills.Levels[book.SkillID] += book.LevelGrant
	if skills.Levels[book.SkillID] > s.ProgressionSystem.LevelCap {
		skills.Levels[book.SkillID] = s.ProgressionSystem.LevelCap
	}
}

// applyXPGrant adds experience points to a skill.
func (s *SkillBookSystem) applyXPGrant(skills *components.Skills, book *SkillBook) {
	if skills.Experience == nil {
		skills.Experience = make(map[string]float64)
	}
	skills.Experience[book.SkillID] += book.XPGrant
}

// markBookAsRead records that a player has read a one-time-use book.
func (s *SkillBookSystem) markBookAsRead(entityID uint64, bookID string, book *SkillBook) {
	if !book.IsOneTimeUse {
		return
	}
	if s.PlayerBooks[entityID] == nil {
		s.PlayerBooks[entityID] = make(map[string]bool)
	}
	s.PlayerBooks[entityID][bookID] = true
}

// CancelReading stops reading a book without gaining benefits.
func (s *SkillBookSystem) CancelReading(entity ecs.Entity) {
	entityID := uint64(entity)
	delete(s.ActiveReaders, entityID)
}

// IsReading checks if an entity is currently reading a book.
func (s *SkillBookSystem) IsReading(entity ecs.Entity) bool {
	entityID := uint64(entity)
	_, ok := s.ActiveReaders[entityID]
	return ok
}

// GetReadingProgress returns the current reading progress (0.0 to 1.0).
func (s *SkillBookSystem) GetReadingProgress(entity ecs.Entity) float64 {
	entityID := uint64(entity)
	reading, ok := s.ActiveReaders[entityID]
	if !ok {
		return 0
	}
	book := s.Books[reading.BookID]
	if book == nil || book.ReadTime <= 0 {
		return 0
	}
	elapsed := book.ReadTime - reading.TimeRemaining
	return elapsed / book.ReadTime
}

// HasRead checks if an entity has read a specific book.
func (s *SkillBookSystem) HasRead(entity ecs.Entity, bookID string) bool {
	entityID := uint64(entity)
	if s.PlayerBooks[entityID] == nil {
		return false
	}
	return s.PlayerBooks[entityID][bookID]
}

// GetReadBooks returns all books an entity has read.
func (s *SkillBookSystem) GetReadBooks(entity ecs.Entity) []*SkillBook {
	entityID := uint64(entity)
	if s.PlayerBooks[entityID] == nil {
		return nil
	}

	books := make([]*SkillBook, 0)
	for bookID := range s.PlayerBooks[entityID] {
		if book := s.Books[bookID]; book != nil {
			books = append(books, book)
		}
	}
	return books
}

// GetUnreadBooks returns books an entity can still benefit from.
func (s *SkillBookSystem) GetUnreadBooks(w *ecs.World, entity ecs.Entity) []*SkillBook {
	entityID := uint64(entity)

	comp, ok := w.GetComponent(entity, "Skills")
	if !ok {
		return nil
	}
	skills := comp.(*components.Skills)

	books := make([]*SkillBook, 0)
	for bookID, book := range s.Books {
		// Skip if already read (one-time use)
		if book.IsOneTimeUse && s.PlayerBooks[entityID] != nil && s.PlayerBooks[entityID][bookID] {
			continue
		}

		// Check level requirements
		currentLevel := 0
		if skills.Levels != nil {
			currentLevel = skills.Levels[book.SkillID]
		}
		if currentLevel < book.RequiredLevel {
			continue // Can't comprehend yet
		}
		if currentLevel >= book.MaxLevel {
			continue // Already beyond this book
		}

		books = append(books, book)
	}
	return books
}

// BookCount returns the total number of registered books.
func (s *SkillBookSystem) BookCount() int {
	return len(s.Books)
}

// ============================================================================
// Skill Synergy System
// ============================================================================

// SkillSynergy defines a bonus from training related skills together.
type SkillSynergy struct {
	// ID uniquely identifies this synergy.
	ID string
	// Name is the display name.
	Name string
	// Description explains the synergy bonus.
	Description string
	// PrimarySkill is the skill that receives the bonus.
	PrimarySkill string
	// SecondarySkills are supporting skills (must have at least one).
	SecondarySkills []string
	// MinLevel is the minimum level in secondary skills to activate.
	MinLevel int
	// BonusType determines what bonus is applied ("xp_mult", "level_bonus", "ability_boost").
	BonusType string
	// BonusValue is the magnitude of the bonus.
	BonusValue float64
	// MaxBonus caps the total bonus (for scaling synergies).
	MaxBonus float64
	// Genre restricts this synergy to a specific genre ("" for all).
	Genre string
}

// SkillSynergySystem tracks and applies skill synergy bonuses.
type SkillSynergySystem struct {
	// Synergies is the registry of all skill synergies.
	Synergies map[string]*SkillSynergy
	// ActiveSynergies tracks active synergies per entity (entityID -> synergyID -> true).
	ActiveSynergies map[uint64]map[string]bool
	// SynergyBonuses caches computed bonuses (entityID -> skillID -> bonus multiplier).
	SynergyBonuses map[uint64]map[string]float64
	// SkillRegistry provides skill metadata.
	SkillRegistry *SkillRegistry
}

// NewSkillSynergySystem creates a new synergy system with default synergies.
func NewSkillSynergySystem(registry *SkillRegistry) *SkillSynergySystem {
	s := &SkillSynergySystem{
		Synergies:       make(map[string]*SkillSynergy),
		ActiveSynergies: make(map[uint64]map[string]bool),
		SynergyBonuses:  make(map[uint64]map[string]float64),
		SkillRegistry:   registry,
	}
	s.registerDefaultSynergies()
	return s
}

// registerDefaultSynergies adds the standard skill synergies.
func (s *SkillSynergySystem) registerDefaultSynergies() {
	// Combat synergies
	s.RegisterSynergy(&SkillSynergy{
		ID: "warrior_prowess", Name: "Warrior's Prowess",
		Description:  "Training defense improves melee combat",
		PrimarySkill: "melee", SecondarySkills: []string{"defense"},
		MinLevel: 15, BonusType: "xp_mult", BonusValue: 0.1, MaxBonus: 0.25,
	})
	s.RegisterSynergy(&SkillSynergy{
		ID: "weapon_master", Name: "Weapon Master",
		Description:  "Melee skill enhances ranged accuracy",
		PrimarySkill: "ranged", SecondarySkills: []string{"melee"},
		MinLevel: 20, BonusType: "xp_mult", BonusValue: 0.08, MaxBonus: 0.20,
	})
	s.RegisterSynergy(&SkillSynergy{
		ID: "combat_athletics", Name: "Combat Athlete",
		Description:  "Athletics training improves defense",
		PrimarySkill: "defense", SecondarySkills: []string{"athletics"},
		MinLevel: 10, BonusType: "xp_mult", BonusValue: 0.12, MaxBonus: 0.30,
	})

	// Stealth synergies
	s.RegisterSynergy(&SkillSynergy{
		ID: "shadow_master", Name: "Shadow Master",
		Description:  "Lockpicking skill improves stealth",
		PrimarySkill: "sneak", SecondarySkills: []string{"lockpicking"},
		MinLevel: 15, BonusType: "xp_mult", BonusValue: 0.1, MaxBonus: 0.25,
	})
	s.RegisterSynergy(&SkillSynergy{
		ID: "light_fingers", Name: "Light Fingers",
		Description:  "Sneaking improves pickpocketing",
		PrimarySkill: "pickpocket", SecondarySkills: []string{"sneak"},
		MinLevel: 10, BonusType: "xp_mult", BonusValue: 0.15, MaxBonus: 0.35,
	})
	s.RegisterSynergy(&SkillSynergy{
		ID: "keen_observer", Name: "Keen Observer",
		Description:  "Perception aids in detecting locks",
		PrimarySkill: "lockpicking", SecondarySkills: []string{"perception"},
		MinLevel: 10, BonusType: "xp_mult", BonusValue: 0.1, MaxBonus: 0.25,
	})

	// Magic synergies
	s.RegisterSynergy(&SkillSynergy{
		ID: "battle_mage", Name: "Battle Mage",
		Description:  "Melee combat enhances destruction magic",
		PrimarySkill: "destruction", SecondarySkills: []string{"melee"},
		MinLevel: 20, BonusType: "ability_boost", BonusValue: 0.05, MaxBonus: 0.15,
	})
	s.RegisterSynergy(&SkillSynergy{
		ID: "healer_wisdom", Name: "Healer's Wisdom",
		Description:  "Alchemy knowledge improves restoration",
		PrimarySkill: "restoration", SecondarySkills: []string{"alchemy"},
		MinLevel: 15, BonusType: "xp_mult", BonusValue: 0.12, MaxBonus: 0.30,
	})
	s.RegisterSynergy(&SkillSynergy{
		ID: "summoner_charisma", Name: "Summoner's Charisma",
		Description:  "Leadership aids conjuration",
		PrimarySkill: "conjuration", SecondarySkills: []string{"leadership"},
		MinLevel: 20, BonusType: "xp_mult", BonusValue: 0.1, MaxBonus: 0.25,
	})
	s.RegisterSynergy(&SkillSynergy{
		ID: "enchanter_insight", Name: "Enchanter's Insight",
		Description:  "Destruction knowledge improves enchanting",
		PrimarySkill: "enchanting", SecondarySkills: []string{"destruction"},
		MinLevel: 15, BonusType: "xp_mult", BonusValue: 0.1, MaxBonus: 0.25,
	})
	s.RegisterSynergy(&SkillSynergy{
		ID: "reality_bender", Name: "Reality Bender",
		Description:  "Conjuration and alteration reinforce each other",
		PrimarySkill: "alteration", SecondarySkills: []string{"conjuration"},
		MinLevel: 25, BonusType: "xp_mult", BonusValue: 0.15, MaxBonus: 0.35,
	})

	// Crafting synergies
	s.RegisterSynergy(&SkillSynergy{
		ID: "master_smith", Name: "Master Smith",
		Description:  "Melee combat knowledge improves smithing",
		PrimarySkill: "smithing", SecondarySkills: []string{"melee"},
		MinLevel: 20, BonusType: "xp_mult", BonusValue: 0.1, MaxBonus: 0.25,
	})
	s.RegisterSynergy(&SkillSynergy{
		ID: "herb_alchemy", Name: "Herbalist Alchemist",
		Description:  "Herbalism directly improves alchemy",
		PrimarySkill: "alchemy", SecondarySkills: []string{"herbalism"},
		MinLevel: 10, BonusType: "xp_mult", BonusValue: 0.2, MaxBonus: 0.50,
	})
	s.RegisterSynergy(&SkillSynergy{
		ID: "arcane_smith", Name: "Arcane Smith",
		Description:  "Enchanting knowledge improves smithing quality",
		PrimarySkill: "smithing", SecondarySkills: []string{"enchanting"},
		MinLevel: 25, BonusType: "ability_boost", BonusValue: 0.08, MaxBonus: 0.20,
	})

	// Social synergies
	s.RegisterSynergy(&SkillSynergy{
		ID: "silver_tongue", Name: "Silver Tongue",
		Description:  "Intimidation enhances persuasion",
		PrimarySkill: "speech", SecondarySkills: []string{"intimidation"},
		MinLevel: 15, BonusType: "xp_mult", BonusValue: 0.1, MaxBonus: 0.25,
	})
	s.RegisterSynergy(&SkillSynergy{
		ID: "merchant_prince", Name: "Merchant Prince",
		Description:  "Speech skill improves bartering",
		PrimarySkill: "barter", SecondarySkills: []string{"speech"},
		MinLevel: 10, BonusType: "xp_mult", BonusValue: 0.15, MaxBonus: 0.35,
	})
	s.RegisterSynergy(&SkillSynergy{
		ID: "commanding_presence", Name: "Commanding Presence",
		Description:  "Combat prowess enhances leadership",
		PrimarySkill: "leadership", SecondarySkills: []string{"melee", "defense"},
		MinLevel: 20, BonusType: "xp_mult", BonusValue: 0.08, MaxBonus: 0.20,
	})
	s.RegisterSynergy(&SkillSynergy{
		ID: "fearsome_warrior", Name: "Fearsome Warrior",
		Description:  "Combat skill enhances intimidation",
		PrimarySkill: "intimidation", SecondarySkills: []string{"melee"},
		MinLevel: 15, BonusType: "xp_mult", BonusValue: 0.12, MaxBonus: 0.30,
	})

	// Survival synergies
	s.RegisterSynergy(&SkillSynergy{
		ID: "outdoor_expert", Name: "Outdoor Expert",
		Description:  "Herbalism improves survival",
		PrimarySkill: "survival", SecondarySkills: []string{"herbalism"},
		MinLevel: 10, BonusType: "xp_mult", BonusValue: 0.15, MaxBonus: 0.35,
	})
	s.RegisterSynergy(&SkillSynergy{
		ID: "trackers_eye", Name: "Tracker's Eye",
		Description:  "Survival skill improves perception",
		PrimarySkill: "perception", SecondarySkills: []string{"survival"},
		MinLevel: 15, BonusType: "xp_mult", BonusValue: 0.1, MaxBonus: 0.25,
	})
	s.RegisterSynergy(&SkillSynergy{
		ID: "endurance_training", Name: "Endurance Training",
		Description:  "Survival knowledge enhances athletics",
		PrimarySkill: "athletics", SecondarySkills: []string{"survival"},
		MinLevel: 10, BonusType: "xp_mult", BonusValue: 0.1, MaxBonus: 0.25,
	})

	// Cross-category synergies
	s.RegisterSynergy(&SkillSynergy{
		ID: "spellsword", Name: "Spellsword",
		Description:  "Magic and combat enhance each other",
		PrimarySkill: "melee", SecondarySkills: []string{"destruction", "alteration"},
		MinLevel: 25, BonusType: "ability_boost", BonusValue: 0.1, MaxBonus: 0.25,
	})
	s.RegisterSynergy(&SkillSynergy{
		ID: "nightblade", Name: "Nightblade",
		Description:  "Stealth and illusion work together",
		PrimarySkill: "sneak", SecondarySkills: []string{"alteration"},
		MinLevel: 20, BonusType: "ability_boost", BonusValue: 0.08, MaxBonus: 0.20,
	})
	s.RegisterSynergy(&SkillSynergy{
		ID: "ranger", Name: "Ranger",
		Description:  "Archery and survival enhance each other",
		PrimarySkill: "ranged", SecondarySkills: []string{"survival", "perception"},
		MinLevel: 15, BonusType: "xp_mult", BonusValue: 0.1, MaxBonus: 0.25,
	})
}

// RegisterSynergy adds a synergy to the registry.
func (s *SkillSynergySystem) RegisterSynergy(synergy *SkillSynergy) {
	if synergy == nil || synergy.ID == "" {
		return
	}
	s.Synergies[synergy.ID] = synergy
}

// GetSynergy retrieves a synergy by ID.
func (s *SkillSynergySystem) GetSynergy(synergyID string) *SkillSynergy {
	return s.Synergies[synergyID]
}

// GetSynergiesForSkill returns all synergies that benefit a specific skill.
func (s *SkillSynergySystem) GetSynergiesForSkill(skillID string) []*SkillSynergy {
	synergies := make([]*SkillSynergy, 0)
	for _, synergy := range s.Synergies {
		if synergy.PrimarySkill == skillID {
			synergies = append(synergies, synergy)
		}
	}
	return synergies
}

// GetSynergiesRequiringSkill returns synergies where the skill is a secondary.
func (s *SkillSynergySystem) GetSynergiesRequiringSkill(skillID string) []*SkillSynergy {
	synergies := make([]*SkillSynergy, 0)
	for _, synergy := range s.Synergies {
		for _, secondary := range synergy.SecondarySkills {
			if secondary == skillID {
				synergies = append(synergies, synergy)
				break
			}
		}
	}
	return synergies
}

// Update evaluates synergies for all entities with skills.
func (s *SkillSynergySystem) Update(w *ecs.World, dt float64) {
	for _, e := range w.Entities("Skills") {
		s.evaluateSynergies(w, e)
	}
}

// evaluateSynergies checks and updates active synergies for an entity.
func (s *SkillSynergySystem) evaluateSynergies(w *ecs.World, entity ecs.Entity) {
	comp, ok := w.GetComponent(entity, "Skills")
	if !ok {
		return
	}
	skills := comp.(*components.Skills)
	if skills.Levels == nil {
		return
	}

	entityID := uint64(entity)
	s.ensureSynergyMaps(entityID)
	s.clearBonuses(entityID)
	s.updateAllSynergies(entityID, skills)
}

// ensureSynergyMaps initializes synergy tracking maps for an entity if needed.
func (s *SkillSynergySystem) ensureSynergyMaps(entityID uint64) {
	if s.ActiveSynergies[entityID] == nil {
		s.ActiveSynergies[entityID] = make(map[string]bool)
	}
	if s.SynergyBonuses[entityID] == nil {
		s.SynergyBonuses[entityID] = make(map[string]float64)
	}
}

// clearBonuses resets all synergy bonuses for an entity.
func (s *SkillSynergySystem) clearBonuses(entityID uint64) {
	for skillID := range s.SynergyBonuses[entityID] {
		s.SynergyBonuses[entityID][skillID] = 0
	}
}

// updateAllSynergies evaluates each synergy and updates bonuses.
func (s *SkillSynergySystem) updateAllSynergies(entityID uint64, skills *components.Skills) {
	for synergyID, synergy := range s.Synergies {
		isActive := s.checkSynergyActive(skills, synergy)
		s.ActiveSynergies[entityID][synergyID] = isActive

		if isActive {
			bonus := s.calculateSynergyBonus(skills, synergy)
			s.SynergyBonuses[entityID][synergy.PrimarySkill] += bonus
		}
	}
}

// checkSynergyActive determines if a synergy's requirements are met.
func (s *SkillSynergySystem) checkSynergyActive(skills *components.Skills, synergy *SkillSynergy) bool {
	// Check all secondary skills meet minimum level
	for _, secondary := range synergy.SecondarySkills {
		if skills.Levels[secondary] < synergy.MinLevel {
			return false
		}
	}
	return true
}

// calculateSynergyBonus computes the bonus value for an active synergy.
func (s *SkillSynergySystem) calculateSynergyBonus(skills *components.Skills, synergy *SkillSynergy) float64 {
	// Calculate average level above minimum across secondary skills
	totalBonus := 0.0
	for _, secondary := range synergy.SecondarySkills {
		level := skills.Levels[secondary]
		if level > synergy.MinLevel {
			// Scale bonus based on how far above minimum
			levelAboveMin := float64(level - synergy.MinLevel)
			contribution := synergy.BonusValue * (levelAboveMin / 50.0) // Scale over 50 levels
			totalBonus += contribution
		}
	}

	// Average across secondary skills
	avgBonus := totalBonus / float64(len(synergy.SecondarySkills))

	// Add base bonus
	finalBonus := synergy.BonusValue + avgBonus

	// Cap at maximum
	if finalBonus > synergy.MaxBonus {
		finalBonus = synergy.MaxBonus
	}

	return finalBonus
}

// IsSynergyActive checks if a specific synergy is active for an entity.
func (s *SkillSynergySystem) IsSynergyActive(entity ecs.Entity, synergyID string) bool {
	entityID := uint64(entity)
	if s.ActiveSynergies[entityID] == nil {
		return false
	}
	return s.ActiveSynergies[entityID][synergyID]
}

// GetActiveSynergies returns all active synergies for an entity.
func (s *SkillSynergySystem) GetActiveSynergies(entity ecs.Entity) []*SkillSynergy {
	entityID := uint64(entity)
	if s.ActiveSynergies[entityID] == nil {
		return nil
	}

	synergies := make([]*SkillSynergy, 0)
	for synergyID, active := range s.ActiveSynergies[entityID] {
		if active {
			if synergy := s.Synergies[synergyID]; synergy != nil {
				synergies = append(synergies, synergy)
			}
		}
	}
	return synergies
}

// GetSkillBonus returns the total synergy bonus for a skill.
func (s *SkillSynergySystem) GetSkillBonus(entity ecs.Entity, skillID string) float64 {
	entityID := uint64(entity)
	if s.SynergyBonuses[entityID] == nil {
		return 0
	}
	return s.SynergyBonuses[entityID][skillID]
}

// GetXPMultiplier returns the XP gain multiplier for a skill (1.0 + bonuses).
func (s *SkillSynergySystem) GetXPMultiplier(entity ecs.Entity, skillID string) float64 {
	entityID := uint64(entity)
	if s.SynergyBonuses[entityID] == nil {
		return 1.0
	}

	multiplier := 1.0
	for _, synergy := range s.GetActiveSynergies(entity) {
		if synergy.PrimarySkill == skillID && synergy.BonusType == "xp_mult" {
			multiplier += s.SynergyBonuses[entityID][skillID]
			break
		}
	}
	return multiplier
}

// GetAbilityBoost returns the ability power boost for a skill.
func (s *SkillSynergySystem) GetAbilityBoost(entity ecs.Entity, skillID string) float64 {
	entityID := uint64(entity)
	if s.SynergyBonuses[entityID] == nil {
		return 0
	}

	boost := 0.0
	for _, synergy := range s.GetActiveSynergies(entity) {
		if synergy.PrimarySkill == skillID && synergy.BonusType == "ability_boost" {
			boost += s.SynergyBonuses[entityID][skillID]
		}
	}
	return boost
}

// GetPotentialSynergies returns synergies an entity could unlock with more training.
func (s *SkillSynergySystem) GetPotentialSynergies(w *ecs.World, entity ecs.Entity) []*SkillSynergy {
	comp, ok := w.GetComponent(entity, "Skills")
	if !ok {
		return nil
	}
	skills := comp.(*components.Skills)
	entityID := uint64(entity)

	potentials := make([]*SkillSynergy, 0)
	for synergyID, synergy := range s.Synergies {
		if s.isSynergyCandidate(entityID, synergyID, synergy, skills) {
			potentials = append(potentials, synergy)
		}
	}
	return potentials
}

// isSynergyCandidate checks if a synergy is a potential unlock for the entity.
func (s *SkillSynergySystem) isSynergyCandidate(entityID uint64, synergyID string, synergy *SkillSynergy, skills *components.Skills) bool {
	if s.isSynergyActive(entityID, synergyID) {
		return false
	}
	return s.hasAnySecondarySkillTrained(synergy, skills)
}

// isSynergyActive checks if a synergy is already active for the entity.
func (s *SkillSynergySystem) isSynergyActive(entityID uint64, synergyID string) bool {
	return s.ActiveSynergies[entityID] != nil && s.ActiveSynergies[entityID][synergyID]
}

// hasAnySecondarySkillTrained checks if any secondary skill has been trained.
func (s *SkillSynergySystem) hasAnySecondarySkillTrained(synergy *SkillSynergy, skills *components.Skills) bool {
	for _, secondary := range synergy.SecondarySkills {
		if skills.Levels[secondary] > 0 {
			return true
		}
	}
	return false
}

// SynergyCount returns the total number of registered synergies.
func (s *SkillSynergySystem) SynergyCount() int {
	return len(s.Synergies)
}

// GetSynergiesByBonusType returns synergies that provide a specific bonus type.
func (s *SkillSynergySystem) GetSynergiesByBonusType(bonusType string) []*SkillSynergy {
	synergies := make([]*SkillSynergy, 0)
	for _, synergy := range s.Synergies {
		if synergy.BonusType == bonusType {
			synergies = append(synergies, synergy)
		}
	}
	return synergies
}
