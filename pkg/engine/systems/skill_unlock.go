package systems

import (
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// ============================================================================
// Skill-Based Action Unlocks
// ============================================================================

// ActionUnlock represents an ability or action unlocked at a skill level.
type ActionUnlock struct {
	ID              string   // Unique action identifier
	Name            string   // Display name
	Description     string   // What the action does
	SkillID         string   // Required skill
	RequiredLevel   int      // Level needed to unlock
	Genre           string   // Genre-specific (empty = all genres)
	ActionType      string   // Type: "ability", "perk", "passive", "recipe", "interaction"
	EffectType      string   // Effect: "damage", "heal", "buff", "utility", "crafting"
	BasePower       float64  // Base power value for scaling
	Cooldown        float64  // Cooldown in seconds (0 = no cooldown)
	ManaCost        float64  // Mana/resource cost
	Prerequisites   []string // Other action IDs required first
	MutualExclusive []string // Actions that can't be used with this
}

// UnlockTier defines standard unlock thresholds.
const (
	UnlockTierNovice      = 1   // Basic actions
	UnlockTierApprentice  = 15  // Early unlocks
	UnlockTierJourneyman  = 30  // Mid-tier abilities
	UnlockTierExpert      = 50  // Advanced techniques
	UnlockTierMaster      = 75  // Powerful abilities
	UnlockTierGrandmaster = 100 // Ultimate abilities
)

// ActionUnlockSystem manages skill-based action unlocks.
type ActionUnlockSystem struct {
	Unlocks        map[string]*ActionUnlock   // Action ID -> unlock definition
	BySkill        map[string][]string        // Skill ID -> action IDs
	PlayerUnlocks  map[uint64]map[string]bool // Player entity -> unlocked action IDs
	skillRegistry  *SkillRegistry
	progressionSys *SkillProgressionSystem
}

// NewActionUnlockSystem creates a new action unlock system.
func NewActionUnlockSystem(skillRegistry *SkillRegistry, progressionSys *SkillProgressionSystem) *ActionUnlockSystem {
	s := &ActionUnlockSystem{
		Unlocks:        make(map[string]*ActionUnlock),
		BySkill:        make(map[string][]string),
		PlayerUnlocks:  make(map[uint64]map[string]bool),
		skillRegistry:  skillRegistry,
		progressionSys: progressionSys,
	}
	s.registerDefaultUnlocks()
	return s
}

// registerDefaultUnlocks adds all default skill-based unlocks.
func (s *ActionUnlockSystem) registerDefaultUnlocks() {
	// Combat unlocks
	s.RegisterUnlock(&ActionUnlock{
		ID: "power_attack", Name: "Power Attack", Description: "A devastating strike dealing 50% more damage",
		SkillID: "melee", RequiredLevel: UnlockTierApprentice, ActionType: "ability",
		EffectType: "damage", BasePower: 1.5, Cooldown: 3.0,
	})
	s.RegisterUnlock(&ActionUnlock{
		ID: "cleave", Name: "Cleave", Description: "Strike multiple enemies in an arc",
		SkillID: "melee", RequiredLevel: UnlockTierJourneyman, ActionType: "ability",
		EffectType: "damage", BasePower: 0.8, Cooldown: 5.0, Prerequisites: []string{"power_attack"},
	})
	s.RegisterUnlock(&ActionUnlock{
		ID: "execute", Name: "Execute", Description: "Instantly kill low-health enemies",
		SkillID: "melee", RequiredLevel: UnlockTierMaster, ActionType: "ability",
		EffectType: "damage", BasePower: 10.0, Cooldown: 30.0, Prerequisites: []string{"cleave"},
	})
	s.RegisterUnlock(&ActionUnlock{
		ID: "precision_shot", Name: "Precision Shot", Description: "Carefully aimed shot with +25% critical chance",
		SkillID: "ranged", RequiredLevel: UnlockTierApprentice, ActionType: "ability",
		EffectType: "damage", BasePower: 1.25, Cooldown: 4.0,
	})
	s.RegisterUnlock(&ActionUnlock{
		ID: "multishot", Name: "Multishot", Description: "Fire three projectiles at once",
		SkillID: "ranged", RequiredLevel: UnlockTierJourneyman, ActionType: "ability",
		EffectType: "damage", BasePower: 0.6, Cooldown: 6.0, Prerequisites: []string{"precision_shot"},
	})
	s.RegisterUnlock(&ActionUnlock{
		ID: "shield_bash", Name: "Shield Bash", Description: "Stun an enemy with your shield",
		SkillID: "blocking", RequiredLevel: UnlockTierApprentice, ActionType: "ability",
		EffectType: "utility", BasePower: 2.0, Cooldown: 5.0,
	})
	s.RegisterUnlock(&ActionUnlock{
		ID: "perfect_block", Name: "Perfect Block", Description: "Timed blocks reflect damage back",
		SkillID: "blocking", RequiredLevel: UnlockTierExpert, ActionType: "passive",
		EffectType: "buff", BasePower: 0.5, Prerequisites: []string{"shield_bash"},
	})

	// Stealth unlocks
	s.RegisterUnlock(&ActionUnlock{
		ID: "silent_move", Name: "Silent Movement", Description: "Move without making sound",
		SkillID: "sneak", RequiredLevel: UnlockTierApprentice, ActionType: "passive",
		EffectType: "utility",
	})
	s.RegisterUnlock(&ActionUnlock{
		ID: "shadow_step", Name: "Shadow Step", Description: "Short-range teleport while sneaking",
		SkillID: "sneak", RequiredLevel: UnlockTierExpert, ActionType: "ability",
		EffectType: "utility", Cooldown: 15.0, Prerequisites: []string{"silent_move"},
	})
	s.RegisterUnlock(&ActionUnlock{
		ID: "vanish", Name: "Vanish", Description: "Instantly enter stealth mid-combat",
		SkillID: "sneak", RequiredLevel: UnlockTierMaster, ActionType: "ability",
		EffectType: "utility", Cooldown: 60.0, Prerequisites: []string{"shadow_step"},
	})
	s.RegisterUnlock(&ActionUnlock{
		ID: "pick_advanced", Name: "Advanced Lockpicking", Description: "Pick expert-level locks",
		SkillID: "lockpick", RequiredLevel: UnlockTierJourneyman, ActionType: "interaction",
		EffectType: "utility",
	})
	s.RegisterUnlock(&ActionUnlock{
		ID: "pick_master", Name: "Master Lockpicking", Description: "Pick master-level locks",
		SkillID: "lockpick", RequiredLevel: UnlockTierMaster, ActionType: "interaction",
		EffectType: "utility", Prerequisites: []string{"pick_advanced"},
	})
	s.RegisterUnlock(&ActionUnlock{
		ID: "lethal_strike", Name: "Lethal Strike", Description: "Backstabs deal 3x damage",
		SkillID: "backstab", RequiredLevel: UnlockTierJourneyman, ActionType: "passive",
		EffectType: "damage", BasePower: 3.0,
	})
	s.RegisterUnlock(&ActionUnlock{
		ID: "assassinate", Name: "Assassinate", Description: "Instant kill on unaware humanoids",
		SkillID: "backstab", RequiredLevel: UnlockTierGrandmaster, ActionType: "ability",
		EffectType: "damage", BasePower: 100.0, Cooldown: 120.0, Prerequisites: []string{"lethal_strike"},
	})

	// Magic unlocks
	s.RegisterUnlock(&ActionUnlock{
		ID: "fireball", Name: "Fireball", Description: "Hurl an explosive ball of fire",
		SkillID: "destruction", RequiredLevel: UnlockTierApprentice, ActionType: "ability",
		EffectType: "damage", BasePower: 30.0, ManaCost: 25.0, Cooldown: 2.0,
	})
	s.RegisterUnlock(&ActionUnlock{
		ID: "flame_wave", Name: "Flame Wave", Description: "Cone of fire in front of you",
		SkillID: "destruction", RequiredLevel: UnlockTierJourneyman, ActionType: "ability",
		EffectType: "damage", BasePower: 20.0, ManaCost: 40.0, Cooldown: 4.0, Prerequisites: []string{"fireball"},
	})
	s.RegisterUnlock(&ActionUnlock{
		ID: "meteor_strike", Name: "Meteor Strike", Description: "Call down a devastating meteor",
		SkillID: "destruction", RequiredLevel: UnlockTierMaster, ActionType: "ability",
		EffectType: "damage", BasePower: 100.0, ManaCost: 100.0, Cooldown: 60.0, Prerequisites: []string{"flame_wave"},
	})
	s.RegisterUnlock(&ActionUnlock{
		ID: "heal", Name: "Healing Touch", Description: "Restore health to yourself or ally",
		SkillID: "restoration", RequiredLevel: UnlockTierNovice, ActionType: "ability",
		EffectType: "heal", BasePower: 20.0, ManaCost: 15.0, Cooldown: 1.0,
	})
	s.RegisterUnlock(&ActionUnlock{
		ID: "regeneration", Name: "Regeneration", Description: "Heal over time effect",
		SkillID: "restoration", RequiredLevel: UnlockTierJourneyman, ActionType: "ability",
		EffectType: "heal", BasePower: 5.0, ManaCost: 30.0, Cooldown: 10.0, Prerequisites: []string{"heal"},
	})
	s.RegisterUnlock(&ActionUnlock{
		ID: "revive", Name: "Revive", Description: "Bring a fallen ally back to life",
		SkillID: "restoration", RequiredLevel: UnlockTierGrandmaster, ActionType: "ability",
		EffectType: "heal", BasePower: 50.0, ManaCost: 200.0, Cooldown: 300.0, Prerequisites: []string{"regeneration"},
	})

	// Crafting unlocks
	s.RegisterUnlock(&ActionUnlock{
		ID: "craft_steel", Name: "Steel Forging", Description: "Craft steel-tier equipment",
		SkillID: "smithing", RequiredLevel: UnlockTierJourneyman, ActionType: "recipe",
		EffectType: "crafting",
	})
	s.RegisterUnlock(&ActionUnlock{
		ID: "craft_mithril", Name: "Mithril Forging", Description: "Craft mithril-tier equipment",
		SkillID: "smithing", RequiredLevel: UnlockTierMaster, ActionType: "recipe",
		EffectType: "crafting", Prerequisites: []string{"craft_steel"},
	})
	s.RegisterUnlock(&ActionUnlock{
		ID: "craft_potion_adv", Name: "Advanced Alchemy", Description: "Brew advanced potions",
		SkillID: "alchemy", RequiredLevel: UnlockTierJourneyman, ActionType: "recipe",
		EffectType: "crafting",
	})
	s.RegisterUnlock(&ActionUnlock{
		ID: "craft_potion_master", Name: "Master Alchemy", Description: "Brew legendary potions",
		SkillID: "alchemy", RequiredLevel: UnlockTierMaster, ActionType: "recipe",
		EffectType: "crafting", Prerequisites: []string{"craft_potion_adv"},
	})

	// Social unlocks
	s.RegisterUnlock(&ActionUnlock{
		ID: "persuade", Name: "Persuade", Description: "Attempt to convince NPCs",
		SkillID: "speech", RequiredLevel: UnlockTierApprentice, ActionType: "interaction",
		EffectType: "utility",
	})
	s.RegisterUnlock(&ActionUnlock{
		ID: "bribe", Name: "Bribe", Description: "Pay off NPCs to look the other way",
		SkillID: "speech", RequiredLevel: UnlockTierJourneyman, ActionType: "interaction",
		EffectType: "utility", Prerequisites: []string{"persuade"},
	})
	s.RegisterUnlock(&ActionUnlock{
		ID: "command", Name: "Command", Description: "Order NPCs to perform actions",
		SkillID: "leadership", RequiredLevel: UnlockTierExpert, ActionType: "ability",
		EffectType: "utility", Cooldown: 30.0,
	})
	s.RegisterUnlock(&ActionUnlock{
		ID: "rally", Name: "Rally", Description: "Buff all allies in range",
		SkillID: "leadership", RequiredLevel: UnlockTierMaster, ActionType: "ability",
		EffectType: "buff", BasePower: 1.2, Cooldown: 60.0, Prerequisites: []string{"command"},
	})
	s.RegisterUnlock(&ActionUnlock{
		ID: "threaten", Name: "Threaten", Description: "Intimidate NPCs into compliance",
		SkillID: "intimidation", RequiredLevel: UnlockTierApprentice, ActionType: "interaction",
		EffectType: "utility",
	})
	s.RegisterUnlock(&ActionUnlock{
		ID: "terrify", Name: "Terrify", Description: "Cause enemies to flee in fear",
		SkillID: "intimidation", RequiredLevel: UnlockTierExpert, ActionType: "ability",
		EffectType: "utility", Cooldown: 45.0, Prerequisites: []string{"threaten"},
	})

	// Survival unlocks
	s.RegisterUnlock(&ActionUnlock{
		ID: "sprint", Name: "Sprint", Description: "Run faster for a short time",
		SkillID: "athletics", RequiredLevel: UnlockTierNovice, ActionType: "ability",
		EffectType: "buff", BasePower: 1.5, Cooldown: 10.0,
	})
	s.RegisterUnlock(&ActionUnlock{
		ID: "dodge_roll", Name: "Dodge Roll", Description: "Quick evasive roll",
		SkillID: "athletics", RequiredLevel: UnlockTierJourneyman, ActionType: "ability",
		EffectType: "utility", Cooldown: 3.0, Prerequisites: []string{"sprint"},
	})
	s.RegisterUnlock(&ActionUnlock{
		ID: "detect_hidden", Name: "Detect Hidden", Description: "Reveal hidden objects and traps",
		SkillID: "perception", RequiredLevel: UnlockTierApprentice, ActionType: "ability",
		EffectType: "utility", Cooldown: 15.0,
	})
	s.RegisterUnlock(&ActionUnlock{
		ID: "eagle_eye", Name: "Eagle Eye", Description: "See enemies through walls briefly",
		SkillID: "perception", RequiredLevel: UnlockTierMaster, ActionType: "ability",
		EffectType: "utility", Cooldown: 60.0, Prerequisites: []string{"detect_hidden"},
	})
	s.RegisterUnlock(&ActionUnlock{
		ID: "track_prey", Name: "Track Prey", Description: "Follow creature trails",
		SkillID: "tracking", RequiredLevel: UnlockTierApprentice, ActionType: "ability",
		EffectType: "utility", Cooldown: 30.0,
	})
	s.RegisterUnlock(&ActionUnlock{
		ID: "hunters_mark", Name: "Hunter's Mark", Description: "Mark a target for bonus damage",
		SkillID: "tracking", RequiredLevel: UnlockTierExpert, ActionType: "ability",
		EffectType: "buff", BasePower: 1.25, Cooldown: 45.0, Prerequisites: []string{"track_prey"},
	})
}

// RegisterUnlock adds an action unlock to the system.
func (s *ActionUnlockSystem) RegisterUnlock(unlock *ActionUnlock) {
	s.Unlocks[unlock.ID] = unlock
	s.BySkill[unlock.SkillID] = append(s.BySkill[unlock.SkillID], unlock.ID)
}

// GetUnlock retrieves an unlock definition by ID.
func (s *ActionUnlockSystem) GetUnlock(actionID string) *ActionUnlock {
	return s.Unlocks[actionID]
}

// GetUnlocksForSkill returns all unlocks for a skill, sorted by level.
func (s *ActionUnlockSystem) GetUnlocksForSkill(skillID string) []*ActionUnlock {
	actionIDs := s.BySkill[skillID]
	unlocks := make([]*ActionUnlock, 0, len(actionIDs))
	for _, id := range actionIDs {
		if unlock, ok := s.Unlocks[id]; ok {
			unlocks = append(unlocks, unlock)
		}
	}
	return unlocks
}

// Update checks for new unlocks based on player skill levels.
func (s *ActionUnlockSystem) Update(w *ecs.World, dt float64) {
	for _, e := range w.Entities("Skills") {
		s.checkPlayerUnlocks(w, e)
	}
}

// checkPlayerUnlocks evaluates unlock conditions for a player.
func (s *ActionUnlockSystem) checkPlayerUnlocks(w *ecs.World, entity ecs.Entity) {
	skills := s.getValidSkills(w, entity)
	if skills == nil {
		return
	}

	entityID := uint64(entity)
	s.ensurePlayerUnlocksMap(entityID)
	s.processUnlockChain(skills, entityID)
}

// getValidSkills returns the skills component if valid, nil otherwise.
func (s *ActionUnlockSystem) getValidSkills(w *ecs.World, entity ecs.Entity) *components.Skills {
	comp, ok := w.GetComponent(entity, "Skills")
	if !ok {
		return nil
	}
	skills := comp.(*components.Skills)
	if skills.Levels == nil {
		return nil
	}
	return skills
}

// ensurePlayerUnlocksMap initializes the player's unlock map if needed.
func (s *ActionUnlockSystem) ensurePlayerUnlocksMap(entityID uint64) {
	if s.PlayerUnlocks[entityID] == nil {
		s.PlayerUnlocks[entityID] = make(map[string]bool)
	}
}

// processUnlockChain processes all unlocks until no new ones are granted.
func (s *ActionUnlockSystem) processUnlockChain(skills *components.Skills, entityID uint64) {
	for {
		if !s.grantNewUnlocks(skills, entityID) {
			break
		}
	}
}

// grantNewUnlocks grants all eligible unlocks, returns true if any were granted.
func (s *ActionUnlockSystem) grantNewUnlocks(skills *components.Skills, entityID uint64) bool {
	newUnlocks := false
	for actionID, unlock := range s.Unlocks {
		if s.PlayerUnlocks[entityID][actionID] {
			continue
		}
		if s.canUnlock(skills, unlock, entityID) {
			s.PlayerUnlocks[entityID][actionID] = true
			newUnlocks = true
		}
	}
	return newUnlocks
}

// canUnlock checks if a player meets unlock requirements.
func (s *ActionUnlockSystem) canUnlock(skills *components.Skills, unlock *ActionUnlock, entityID uint64) bool {
	// Check skill level
	playerLevel := skills.Levels[unlock.SkillID]
	if playerLevel < unlock.RequiredLevel {
		return false
	}

	// Check prerequisites
	for _, prereqID := range unlock.Prerequisites {
		if !s.PlayerUnlocks[entityID][prereqID] {
			return false
		}
	}

	return true
}

// IsUnlocked checks if a player has unlocked an action.
func (s *ActionUnlockSystem) IsUnlocked(entity ecs.Entity, actionID string) bool {
	entityID := uint64(entity)
	if s.PlayerUnlocks[entityID] == nil {
		return false
	}
	return s.PlayerUnlocks[entityID][actionID]
}

// GetUnlockedActions returns all actions unlocked by a player.
func (s *ActionUnlockSystem) GetUnlockedActions(entity ecs.Entity) []*ActionUnlock {
	entityID := uint64(entity)
	if s.PlayerUnlocks[entityID] == nil {
		return nil
	}

	unlocks := make([]*ActionUnlock, 0)
	for actionID, unlocked := range s.PlayerUnlocks[entityID] {
		if unlocked {
			if unlock, ok := s.Unlocks[actionID]; ok {
				unlocks = append(unlocks, unlock)
			}
		}
	}
	return unlocks
}

// GetAvailableUnlocks returns actions a player could unlock with more levels.
func (s *ActionUnlockSystem) GetAvailableUnlocks(w *ecs.World, entity ecs.Entity) []*ActionUnlock {
	comp, ok := w.GetComponent(entity, "Skills")
	if !ok {
		return nil
	}
	skills := comp.(*components.Skills)
	entityID := uint64(entity)
	available := make([]*ActionUnlock, 0)

	for _, unlock := range s.Unlocks {
		if s.isAvailableUnlock(entityID, unlock, skills) {
			available = append(available, unlock)
		}
	}
	return available
}

// isAvailableUnlock checks if an unlock is available (not unlocked, prereqs met, skill exists).
func (s *ActionUnlockSystem) isAvailableUnlock(entityID uint64, unlock *ActionUnlock, skills *components.Skills) bool {
	if s.isAlreadyUnlocked(entityID, unlock.ID) {
		return false
	}
	if !s.arePrerequisitesMet(entityID, unlock.Prerequisites) {
		return false
	}
	return s.hasSkill(skills, unlock.SkillID)
}

// hasSkill checks if the skills component has a specific skill.
func (s *ActionUnlockSystem) hasSkill(skills *components.Skills, skillID string) bool {
	if skills.Levels == nil {
		return false
	}
	_, hasSkill := skills.Levels[skillID]
	return hasSkill
}

// GetNextUnlockForSkill returns the next unlock a player will get for a skill.
func (s *ActionUnlockSystem) GetNextUnlockForSkill(w *ecs.World, entity ecs.Entity, skillID string) *ActionUnlock {
	comp, ok := w.GetComponent(entity, "Skills")
	if !ok {
		return nil
	}
	skills := comp.(*components.Skills)
	playerLevel := s.getPlayerSkillLevel(skills, skillID)
	entityID := uint64(entity)

	var nextUnlock *ActionUnlock
	lowestLevel := 999

	for _, unlock := range s.GetUnlocksForSkill(skillID) {
		if s.isAlreadyUnlocked(entityID, unlock.ID) {
			continue
		}
		if s.isNextUnlockCandidate(unlock, playerLevel, lowestLevel, entityID) {
			nextUnlock = unlock
			lowestLevel = unlock.RequiredLevel
		}
	}
	return nextUnlock
}

// getPlayerSkillLevel returns the player's level in a skill, defaulting to 0.
func (s *ActionUnlockSystem) getPlayerSkillLevel(skills *components.Skills, skillID string) int {
	if skills.Levels == nil {
		return 0
	}
	return skills.Levels[skillID]
}

// isAlreadyUnlocked checks if an entity has already unlocked an action.
func (s *ActionUnlockSystem) isAlreadyUnlocked(entityID uint64, actionID string) bool {
	return s.PlayerUnlocks[entityID] != nil && s.PlayerUnlocks[entityID][actionID]
}

// isNextUnlockCandidate checks if an unlock is a valid next-unlock candidate.
func (s *ActionUnlockSystem) isNextUnlockCandidate(unlock *ActionUnlock, playerLevel, lowestLevel int, entityID uint64) bool {
	if unlock.RequiredLevel <= playerLevel || unlock.RequiredLevel >= lowestLevel {
		return false
	}
	return s.arePrerequisitesMet(entityID, unlock.Prerequisites)
}

// arePrerequisitesMet checks if all prerequisites are unlocked for an entity.
func (s *ActionUnlockSystem) arePrerequisitesMet(entityID uint64, prerequisites []string) bool {
	for _, prereqID := range prerequisites {
		if s.PlayerUnlocks[entityID] == nil || !s.PlayerUnlocks[entityID][prereqID] {
			return false
		}
	}
	return true
}

// UnlockCount returns the total number of registered unlocks.
func (s *ActionUnlockSystem) UnlockCount() int {
	return len(s.Unlocks)
}

// GetUnlocksByType returns unlocks of a specific action type.
func (s *ActionUnlockSystem) GetUnlocksByType(actionType string) []*ActionUnlock {
	unlocks := make([]*ActionUnlock, 0)
	for _, unlock := range s.Unlocks {
		if unlock.ActionType == actionType {
			unlocks = append(unlocks, unlock)
		}
	}
	return unlocks
}
