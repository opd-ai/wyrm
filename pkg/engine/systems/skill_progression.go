package systems

import (
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// SkillProgressionSystem manages skill experience gain and level-ups.
type SkillProgressionSystem struct {
	// XPPerLevel is the base XP required per level (scales with level).
	XPPerLevel float64
	// LevelCap is the maximum skill level.
	LevelCap int
}

// NewSkillProgressionSystem creates a new skill progression system.
func NewSkillProgressionSystem(xpPerLevel float64, levelCap int) *SkillProgressionSystem {
	if levelCap <= 0 {
		levelCap = DefaultLevelCap
	}
	if xpPerLevel <= 0 {
		xpPerLevel = DefaultXPPerLevel
	}
	return &SkillProgressionSystem{
		XPPerLevel: xpPerLevel,
		LevelCap:   levelCap,
	}
}

// Update processes skill experience and level-ups each tick.
func (s *SkillProgressionSystem) Update(w *ecs.World, dt float64) {
	for _, e := range w.Entities("Skills") {
		comp, ok := w.GetComponent(e, "Skills")
		if !ok {
			continue
		}
		skills := comp.(*components.Skills)
		s.processSkillProgression(skills)
	}
}

// processSkillProgression checks all skills for level-up conditions.
func (s *SkillProgressionSystem) processSkillProgression(skills *components.Skills) {
	if skills.Levels == nil || skills.Experience == nil {
		return
	}
	for skillID, xp := range skills.Experience {
		level := skills.Levels[skillID]
		if level >= s.LevelCap {
			continue
		}
		xpRequired := s.calculateXPRequired(level)
		if xp >= xpRequired {
			skills.Levels[skillID] = level + 1
			skills.Experience[skillID] = xp - xpRequired
		}
	}
}

// calculateXPRequired computes XP needed for the next level.
// Uses a simple scaling formula: base * (1 + level * LevelScalingFactor)
func (s *SkillProgressionSystem) calculateXPRequired(currentLevel int) float64 {
	return s.XPPerLevel * (BasePriceMultiplier + float64(currentLevel)*LevelScalingFactor)
}

// GrantSkillXP adds experience to a skill for an entity.
func (s *SkillProgressionSystem) GrantSkillXP(w *ecs.World, entity ecs.Entity, skillID string, xp float64) bool {
	comp, ok := w.GetComponent(entity, "Skills")
	if !ok {
		return false
	}
	skills := comp.(*components.Skills)
	if skills.Experience == nil {
		skills.Experience = make(map[string]float64)
	}
	if _, exists := skills.Levels[skillID]; !exists {
		return false
	}
	skills.Experience[skillID] += xp
	return true
}

// GetSkillLevel returns the current level of a skill for an entity.
func (s *SkillProgressionSystem) GetSkillLevel(w *ecs.World, entity ecs.Entity, skillID string) int {
	comp, ok := w.GetComponent(entity, "Skills")
	if !ok {
		return 0
	}
	skills := comp.(*components.Skills)
	if skills.Levels == nil {
		return 0
	}
	return skills.Levels[skillID]
}

// SkillDefinition defines a single skill's properties.
type SkillDefinition struct {
	ID          string
	Name        string // Genre-specific display name
	School      string // Which school this skill belongs to
	Description string // What the skill does
	MaxLevel    int    // Maximum level (usually 100)
}

// SkillSchoolDefinition defines a school of related skills.
type SkillSchoolDefinition struct {
	ID          string
	Name        string // Genre-specific school name
	Description string
	Skills      []string // Skill IDs in this school
}

// SkillRegistry holds all skill definitions organized by genre.
type SkillRegistry struct {
	Skills  map[string]*SkillDefinition       // skillID -> definition
	Schools map[string]*SkillSchoolDefinition // schoolID -> definition
	ByGenre map[string]map[string]string      // genre -> skillID -> display name
}

// NewSkillRegistry creates a skill registry with all 30+ skills.
func NewSkillRegistry() *SkillRegistry {
	r := &SkillRegistry{
		Skills:  make(map[string]*SkillDefinition),
		Schools: make(map[string]*SkillSchoolDefinition),
		ByGenre: make(map[string]map[string]string),
	}
	r.initializeSkills()
	return r
}

// initializeSkills populates all skill definitions.
func (r *SkillRegistry) initializeSkills() {
	// Define 6 schools with 5-6 skills each (32 total skills)
	r.Schools = map[string]*SkillSchoolDefinition{
		"combat": {
			ID:          "combat",
			Name:        "Combat",
			Description: "Physical fighting techniques",
			Skills:      []string{"melee", "ranged", "blocking", "critical", "armor", "dualwield"},
		},
		"stealth": {
			ID:          "stealth",
			Name:        "Stealth",
			Description: "Covert operations and subterfuge",
			Skills:      []string{"sneak", "lockpick", "pickpocket", "backstab", "disguise"},
		},
		"magic": {
			ID:          "magic",
			Name:        "Magic",
			Description: "Arcane and supernatural abilities",
			Skills:      []string{"destruction", "restoration", "conjuration", "illusion", "enchanting"},
		},
		"crafting": {
			ID:          "crafting",
			Name:        "Crafting",
			Description: "Creating and improving items",
			Skills:      []string{"smithing", "alchemy", "cooking", "tailoring", "engineering"},
		},
		"social": {
			ID:          "social",
			Name:        "Social",
			Description: "Interpersonal abilities",
			Skills:      []string{"speech", "intimidation", "barter", "leadership", "deception"},
		},
		"survival": {
			ID:          "survival",
			Name:        "Survival",
			Description: "Wilderness and exploration skills",
			Skills:      []string{"athletics", "perception", "tracking", "herbalism", "riding", "swimming"},
		},
	}

	// Define individual skills
	skillDefs := []SkillDefinition{
		// Combat skills
		{ID: "melee", Name: "Melee Combat", School: "combat", Description: "Proficiency with melee weapons", MaxLevel: 100},
		{ID: "ranged", Name: "Ranged Combat", School: "combat", Description: "Proficiency with ranged weapons", MaxLevel: 100},
		{ID: "blocking", Name: "Blocking", School: "combat", Description: "Ability to block attacks", MaxLevel: 100},
		{ID: "critical", Name: "Critical Strikes", School: "combat", Description: "Chance and damage of critical hits", MaxLevel: 100},
		{ID: "armor", Name: "Armor Proficiency", School: "combat", Description: "Effectiveness of worn armor", MaxLevel: 100},
		{ID: "dualwield", Name: "Dual Wielding", School: "combat", Description: "Fighting with two weapons", MaxLevel: 100},

		// Stealth skills
		{ID: "sneak", Name: "Sneak", School: "stealth", Description: "Moving undetected", MaxLevel: 100},
		{ID: "lockpick", Name: "Lockpicking", School: "stealth", Description: "Opening locked containers and doors", MaxLevel: 100},
		{ID: "pickpocket", Name: "Pickpocket", School: "stealth", Description: "Stealing from NPCs unnoticed", MaxLevel: 100},
		{ID: "backstab", Name: "Backstab", School: "stealth", Description: "Extra damage from stealth attacks", MaxLevel: 100},
		{ID: "disguise", Name: "Disguise", School: "stealth", Description: "Ability to blend into factions", MaxLevel: 100},

		// Magic skills
		{ID: "destruction", Name: "Destruction", School: "magic", Description: "Offensive magic damage", MaxLevel: 100},
		{ID: "restoration", Name: "Restoration", School: "magic", Description: "Healing and protective magic", MaxLevel: 100},
		{ID: "conjuration", Name: "Conjuration", School: "magic", Description: "Summoning creatures and objects", MaxLevel: 100},
		{ID: "illusion", Name: "Illusion", School: "magic", Description: "Deceptive and mind-affecting magic", MaxLevel: 100},
		{ID: "enchanting", Name: "Enchanting", School: "magic", Description: "Imbuing items with magical effects", MaxLevel: 100},

		// Crafting skills
		{ID: "smithing", Name: "Smithing", School: "crafting", Description: "Forging weapons and armor", MaxLevel: 100},
		{ID: "alchemy", Name: "Alchemy", School: "crafting", Description: "Creating potions and compounds", MaxLevel: 100},
		{ID: "cooking", Name: "Cooking", School: "crafting", Description: "Preparing food with beneficial effects", MaxLevel: 100},
		{ID: "tailoring", Name: "Tailoring", School: "crafting", Description: "Creating cloth and leather items", MaxLevel: 100},
		{ID: "engineering", Name: "Engineering", School: "crafting", Description: "Creating mechanical devices", MaxLevel: 100},

		// Social skills
		{ID: "speech", Name: "Speech", School: "social", Description: "Persuading others through conversation", MaxLevel: 100},
		{ID: "intimidation", Name: "Intimidation", School: "social", Description: "Coercing others through threats", MaxLevel: 100},
		{ID: "barter", Name: "Barter", School: "social", Description: "Getting better prices in trade", MaxLevel: 100},
		{ID: "leadership", Name: "Leadership", School: "social", Description: "Inspiring and commanding followers", MaxLevel: 100},
		{ID: "deception", Name: "Deception", School: "social", Description: "Lying convincingly", MaxLevel: 100},

		// Survival skills
		{ID: "athletics", Name: "Athletics", School: "survival", Description: "Running, jumping, and climbing", MaxLevel: 100},
		{ID: "perception", Name: "Perception", School: "survival", Description: "Noticing hidden things", MaxLevel: 100},
		{ID: "tracking", Name: "Tracking", School: "survival", Description: "Following trails and finding creatures", MaxLevel: 100},
		{ID: "herbalism", Name: "Herbalism", School: "survival", Description: "Identifying and gathering plants", MaxLevel: 100},
		{ID: "riding", Name: "Riding", School: "survival", Description: "Controlling mounts and vehicles", MaxLevel: 100},
		{ID: "swimming", Name: "Swimming", School: "survival", Description: "Moving through water", MaxLevel: 100},
	}

	for _, sd := range skillDefs {
		def := sd // Create copy for pointer
		r.Skills[def.ID] = &def
	}

	// Define genre-specific skill names
	r.initializeGenreNames()
}

// initializeGenreNames sets genre-appropriate display names for skills.
func (r *SkillRegistry) initializeGenreNames() {
	r.ByGenre = map[string]map[string]string{
		"fantasy": {
			"melee": "Swordsmanship", "ranged": "Archery", "blocking": "Shield Work",
			"destruction": "Destruction Magic", "restoration": "Healing Magic",
			"conjuration": "Summoning", "illusion": "Mind Magic", "enchanting": "Enchanting",
			"smithing": "Blacksmithing", "alchemy": "Potion Craft", "cooking": "Culinary Arts",
			"sneak": "Stealth", "lockpick": "Lockpicking", "pickpocket": "Pickpocketing",
			"speech": "Speechcraft", "barter": "Mercantile", "riding": "Horsemanship",
		},
		"sci-fi": {
			"melee": "Melee Weapons", "ranged": "Firearms", "blocking": "Deflection",
			"destruction": "Energy Weapons", "restoration": "Medical", "engineering": "Tech Repair",
			"conjuration": "Drone Control", "illusion": "Electronic Warfare", "enchanting": "Modification",
			"smithing": "Fabrication", "alchemy": "Chemistry", "cooking": "Ration Prep",
			"sneak": "Infiltration", "lockpick": "Hacking", "pickpocket": "Sleight of Hand",
			"speech": "Negotiation", "barter": "Trading", "riding": "Piloting",
		},
		"horror": {
			"melee": "Desperate Fighting", "ranged": "Firearms", "blocking": "Defensive Stance",
			"destruction": "Dark Arts", "restoration": "First Aid", "conjuration": "Ritual Summoning",
			"illusion": "Mind Shield", "enchanting": "Occult Binding",
			"smithing": "Improvised Weapons", "alchemy": "Medicine", "cooking": "Preserved Foods",
			"sneak": "Hiding", "lockpick": "Forced Entry", "pickpocket": "Quick Hands",
			"speech": "Pleading", "intimidation": "Survival Instinct", "barter": "Scavenging",
		},
		"cyberpunk": {
			"melee": "Street Fighting", "ranged": "Firearms", "blocking": "Reflex Block",
			"destruction": "Cyberweapons", "restoration": "Biotechnics", "engineering": "Cybertech",
			"conjuration": "Combat Drones", "illusion": "Neural Hacking", "enchanting": "Chrome Upgrade",
			"smithing": "Tech Crafting", "alchemy": "Drug Synthesis", "cooking": "Street Food",
			"sneak": "Stealth Tech", "lockpick": "Security Bypass", "pickpocket": "Quickfingers",
			"speech": "Street Cred", "barter": "Black Market", "riding": "Vehicle Control",
		},
		"post-apocalyptic": {
			"melee": "Brawling", "ranged": "Guns", "blocking": "Scrap Armor",
			"destruction": "Radiation Burns", "restoration": "Field Medicine", "engineering": "Jury-Rigging",
			"conjuration": "Taming Mutants", "illusion": "Camouflage", "enchanting": "Modification",
			"smithing": "Weapon Smithing", "alchemy": "Chem Craft", "cooking": "Wasteland Cooking",
			"sneak": "Scavenging", "lockpick": "Lock Breaking", "pickpocket": "Thievery",
			"speech": "Bartering", "barter": "Trading", "riding": "Vehicle Operation",
		},
	}
}

// GetSkillName returns the genre-appropriate name for a skill.
func (r *SkillRegistry) GetSkillName(skillID, genre string) string {
	if genreNames, ok := r.ByGenre[genre]; ok {
		if name, ok := genreNames[skillID]; ok {
			return name
		}
	}
	if skill, ok := r.Skills[skillID]; ok {
		return skill.Name
	}
	return skillID
}

// GetSkillsForSchool returns all skill IDs in a school.
func (r *SkillRegistry) GetSkillsForSchool(schoolID string) []string {
	if school, ok := r.Schools[schoolID]; ok {
		return school.Skills
	}
	return nil
}

// GetAllSkillIDs returns all registered skill IDs.
func (r *SkillRegistry) GetAllSkillIDs() []string {
	ids := make([]string, 0, len(r.Skills))
	for id := range r.Skills {
		ids = append(ids, id)
	}
	return ids
}

// SkillCount returns the total number of registered skills.
func (r *SkillRegistry) SkillCount() int {
	return len(r.Skills)
}

// SchoolCount returns the total number of skill schools.
func (r *SkillRegistry) SchoolCount() int {
	return len(r.Schools)
}

// NPCTrainer represents an NPC that can train players in skills.
type NPCTrainer struct {
	NPCID         uint64
	Name          string
	TrainedSkills []string // Skill IDs this trainer can teach
	MaxTrainLevel int      // Maximum level they can train to
	CostPerLevel  int      // Base gold cost per level trained
	TrainXPBonus  float64  // XP bonus when training (1.0 = 100% of normal)
}

// TrainingResult represents the outcome of a training session.
type TrainingResult struct {
	Success      bool
	SkillID      string
	LevelsGained int
	XPGained     float64
	GoldSpent    int
	ErrorMessage string
}

// NPCTrainingSystem handles skill training from NPCs.
type NPCTrainingSystem struct {
	Trainers       map[uint64]*NPCTrainer // NPC entity ID -> trainer info
	skillRegistry  *SkillRegistry
	progressionSys *SkillProgressionSystem
}

// NewNPCTrainingSystem creates a new NPC training system.
func NewNPCTrainingSystem(skillRegistry *SkillRegistry, progressionSys *SkillProgressionSystem) *NPCTrainingSystem {
	return &NPCTrainingSystem{
		Trainers:       make(map[uint64]*NPCTrainer),
		skillRegistry:  skillRegistry,
		progressionSys: progressionSys,
	}
}

// RegisterTrainer adds an NPC as a skill trainer.
func (s *NPCTrainingSystem) RegisterTrainer(trainer *NPCTrainer) {
	s.Trainers[trainer.NPCID] = trainer
}

// GetTrainer retrieves trainer info for an NPC.
func (s *NPCTrainingSystem) GetTrainer(npcID uint64) *NPCTrainer {
	return s.Trainers[npcID]
}

// IsTrainer checks if an NPC can train players.
func (s *NPCTrainingSystem) IsTrainer(npcID uint64) bool {
	_, ok := s.Trainers[npcID]
	return ok
}

// GetTrainableSkills returns skills an NPC can teach that a player can benefit from.
func (s *NPCTrainingSystem) GetTrainableSkills(w *ecs.World, npcID uint64, playerEntity ecs.Entity) []string {
	trainer := s.Trainers[npcID]
	if trainer == nil {
		return nil
	}

	comp, ok := w.GetComponent(playerEntity, "Skills")
	if !ok {
		return nil
	}
	skills := comp.(*components.Skills)

	var trainable []string
	for _, skillID := range trainer.TrainedSkills {
		playerLevel := 0
		if skills.Levels != nil {
			playerLevel = skills.Levels[skillID]
		}
		if playerLevel < trainer.MaxTrainLevel {
			trainable = append(trainable, skillID)
		}
	}
	return trainable
}

// CalculateTrainingCost computes gold cost to train a skill.
func (s *NPCTrainingSystem) CalculateTrainingCost(npcID uint64, currentLevel, targetLevel int) int {
	trainer := s.Trainers[npcID]
	if trainer == nil {
		return 0
	}

	totalCost := 0
	for level := currentLevel; level < targetLevel; level++ {
		// Cost increases with level
		levelCost := trainer.CostPerLevel * (1 + level/10)
		totalCost += levelCost
	}
	return totalCost
}

// TrainSkill attempts to train a player in a skill.
func (s *NPCTrainingSystem) TrainSkill(
	w *ecs.World,
	npcID uint64,
	playerEntity ecs.Entity,
	skillID string,
	levelsToTrain int,
	playerGold *int, // pointer to player's gold (will be modified)
) *TrainingResult {
	result := &TrainingResult{SkillID: skillID}

	trainer := s.Trainers[npcID]
	if trainer == nil {
		result.ErrorMessage = "NPC is not a trainer"
		return result
	}

	if err := s.validateTrainerSkill(trainer, skillID); err != "" {
		result.ErrorMessage = err
		return result
	}

	skills, err := s.getPlayerSkills(w, playerEntity)
	if err != "" {
		result.ErrorMessage = err
		return result
	}

	currentLevel := s.getSkillLevel(skills, skillID)
	targetLevel, levelsToTrain := s.clampTrainingLevels(currentLevel, levelsToTrain, trainer.MaxTrainLevel)
	if levelsToTrain <= 0 {
		result.ErrorMessage = "Already at max trainable level"
		return result
	}

	cost := s.CalculateTrainingCost(npcID, currentLevel, targetLevel)
	if playerGold != nil && *playerGold < cost {
		result.ErrorMessage = "Insufficient gold"
		return result
	}

	xpGained := s.applyTraining(skills, skillID, targetLevel, levelsToTrain, trainer)
	s.deductGold(playerGold, cost)

	result.Success = true
	result.LevelsGained = levelsToTrain
	result.XPGained = xpGained
	result.GoldSpent = cost
	return result
}

// validateTrainerSkill checks if the trainer teaches the requested skill.
func (s *NPCTrainingSystem) validateTrainerSkill(trainer *NPCTrainer, skillID string) string {
	for _, sid := range trainer.TrainedSkills {
		if sid == skillID {
			return ""
		}
	}
	return "Trainer doesn't teach this skill"
}

// getPlayerSkills retrieves the player's Skills component.
func (s *NPCTrainingSystem) getPlayerSkills(w *ecs.World, playerEntity ecs.Entity) (*components.Skills, string) {
	comp, ok := w.GetComponent(playerEntity, "Skills")
	if !ok {
		return nil, "Player has no skills component"
	}
	return comp.(*components.Skills), ""
}

// getSkillLevel returns the current level of a skill, defaulting to 0.
func (s *NPCTrainingSystem) getSkillLevel(skills *components.Skills, skillID string) int {
	if skills.Levels == nil {
		return 0
	}
	return skills.Levels[skillID]
}

// clampTrainingLevels adjusts target level and levels to train based on max.
func (s *NPCTrainingSystem) clampTrainingLevels(currentLevel, levelsToTrain, maxLevel int) (targetLevel, adjustedLevels int) {
	targetLevel = currentLevel + levelsToTrain
	if targetLevel > maxLevel {
		targetLevel = maxLevel
		levelsToTrain = targetLevel - currentLevel
	}
	return targetLevel, levelsToTrain
}

// applyTraining applies the skill level increase and grants XP.
func (s *NPCTrainingSystem) applyTraining(skills *components.Skills, skillID string, targetLevel, levelsToTrain int, trainer *NPCTrainer) float64 {
	if skills.Levels == nil {
		skills.Levels = make(map[string]int)
	}
	skills.Levels[skillID] = targetLevel

	xpGained := float64(levelsToTrain) * s.progressionSys.XPPerLevel * trainer.TrainXPBonus * TrainingXPFraction
	if skills.Experience == nil {
		skills.Experience = make(map[string]float64)
	}
	skills.Experience[skillID] += xpGained
	return xpGained
}

// deductGold subtracts the cost from player's gold if provided.
func (s *NPCTrainingSystem) deductGold(playerGold *int, cost int) {
	if playerGold != nil {
		*playerGold -= cost
	}
}

// Update processes NPC training system state (currently no per-tick logic needed).
func (s *NPCTrainingSystem) Update(w *ecs.World, dt float64) {
	// Training is event-driven, no per-tick updates needed
}

// TrainerCount returns the number of registered trainers.
func (s *NPCTrainingSystem) TrainerCount() int {
	return len(s.Trainers)
}

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
	comp, ok := w.GetComponent(entity, "Skills")
	if !ok {
		return
	}
	skills := comp.(*components.Skills)
	if skills.Levels == nil {
		return
	}

	entityID := uint64(entity)
	if s.PlayerUnlocks[entityID] == nil {
		s.PlayerUnlocks[entityID] = make(map[string]bool)
	}

	// Loop until no new unlocks are made (handles prerequisite chains)
	for {
		newUnlocks := false
		for actionID, unlock := range s.Unlocks {
			if s.PlayerUnlocks[entityID][actionID] {
				continue // Already unlocked
			}
			if s.canUnlock(skills, unlock, entityID) {
				s.PlayerUnlocks[entityID][actionID] = true
				newUnlocks = true
			}
		}
		if !newUnlocks {
			break // No more unlocks possible this frame
		}
	}
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
		// Skip already unlocked
		if s.PlayerUnlocks[entityID] != nil && s.PlayerUnlocks[entityID][unlock.ID] {
			continue
		}
		// Check prerequisites are met
		prereqsMet := true
		for _, prereqID := range unlock.Prerequisites {
			if s.PlayerUnlocks[entityID] == nil || !s.PlayerUnlocks[entityID][prereqID] {
				prereqsMet = false
				break
			}
		}
		if !prereqsMet {
			continue
		}
		// Player has the skill but not high enough level
		if skills.Levels != nil {
			if _, hasSkill := skills.Levels[unlock.SkillID]; hasSkill {
				available = append(available, unlock)
			}
		}
	}
	return available
}

// GetNextUnlockForSkill returns the next unlock a player will get for a skill.
func (s *ActionUnlockSystem) GetNextUnlockForSkill(w *ecs.World, entity ecs.Entity, skillID string) *ActionUnlock {
	comp, ok := w.GetComponent(entity, "Skills")
	if !ok {
		return nil
	}
	skills := comp.(*components.Skills)
	playerLevel := 0
	if skills.Levels != nil {
		playerLevel = skills.Levels[skillID]
	}

	entityID := uint64(entity)
	var nextUnlock *ActionUnlock
	lowestLevel := 999

	for _, unlock := range s.GetUnlocksForSkill(skillID) {
		// Skip already unlocked
		if s.PlayerUnlocks[entityID] != nil && s.PlayerUnlocks[entityID][unlock.ID] {
			continue
		}
		// Must be above current level and lower than current best
		if unlock.RequiredLevel > playerLevel && unlock.RequiredLevel < lowestLevel {
			// Check prerequisites
			prereqsMet := true
			for _, prereqID := range unlock.Prerequisites {
				if s.PlayerUnlocks[entityID] == nil || !s.PlayerUnlocks[entityID][prereqID] {
					prereqsMet = false
					break
				}
			}
			if prereqsMet {
				nextUnlock = unlock
				lowestLevel = unlock.RequiredLevel
			}
		}
	}
	return nextUnlock
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

	// Check if already reading
	if _, ok := s.ActiveReaders[entityID]; ok {
		return false
	}

	book := s.Books[bookID]
	if book == nil {
		return false
	}

	// Check if already read (one-time use)
	if book.IsOneTimeUse {
		if s.PlayerBooks[entityID] != nil && s.PlayerBooks[entityID][bookID] {
			return false
		}
	}

	// Check skill level requirements
	comp, ok := w.GetComponent(entity, "Skills")
	if !ok {
		return false
	}
	skills := comp.(*components.Skills)
	if skills.Levels == nil {
		return false
	}

	currentLevel := skills.Levels[book.SkillID]
	if currentLevel < book.RequiredLevel {
		return false // Not skilled enough to comprehend
	}
	if currentLevel >= book.MaxLevel {
		return false // Already beyond what this book teaches
	}

	// Start reading
	s.ActiveReaders[entityID] = &ActiveReading{
		BookID:        bookID,
		TimeRemaining: book.ReadTime,
	}
	return true
}

// completeReading finishes reading and applies the book's benefits.
func (s *SkillBookSystem) completeReading(w *ecs.World, entity ecs.Entity) {
	entityID := uint64(entity)
	reading, ok := s.ActiveReaders[entityID]
	if !ok {
		return
	}

	book := s.Books[reading.BookID]
	if book == nil {
		delete(s.ActiveReaders, entityID)
		return
	}

	comp, ok := w.GetComponent(entity, "Skills")
	if !ok {
		delete(s.ActiveReaders, entityID)
		return
	}
	skills := comp.(*components.Skills)

	// Apply XP or level grant
	if book.LevelGrant > 0 {
		// Direct level grant
		if skills.Levels == nil {
			skills.Levels = make(map[string]int)
		}
		skills.Levels[book.SkillID] += book.LevelGrant
		if skills.Levels[book.SkillID] > s.ProgressionSystem.LevelCap {
			skills.Levels[book.SkillID] = s.ProgressionSystem.LevelCap
		}
	} else if book.XPGrant > 0 {
		// XP grant
		if skills.Experience == nil {
			skills.Experience = make(map[string]float64)
		}
		skills.Experience[book.SkillID] += book.XPGrant
	}

	// Mark book as read
	if book.IsOneTimeUse {
		if s.PlayerBooks[entityID] == nil {
			s.PlayerBooks[entityID] = make(map[string]bool)
		}
		s.PlayerBooks[entityID][reading.BookID] = true
	}

	// Clean up
	delete(s.ActiveReaders, entityID)
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
	if s.ActiveSynergies[entityID] == nil {
		s.ActiveSynergies[entityID] = make(map[string]bool)
	}
	if s.SynergyBonuses[entityID] == nil {
		s.SynergyBonuses[entityID] = make(map[string]float64)
	}

	// Clear old bonuses
	for skillID := range s.SynergyBonuses[entityID] {
		s.SynergyBonuses[entityID][skillID] = 0
	}

	for synergyID, synergy := range s.Synergies {
		wasActive := s.ActiveSynergies[entityID][synergyID]
		isActive := s.checkSynergyActive(skills, synergy)

		s.ActiveSynergies[entityID][synergyID] = isActive

		if isActive {
			bonus := s.calculateSynergyBonus(skills, synergy)
			s.SynergyBonuses[entityID][synergy.PrimarySkill] += bonus
		}

		// Log activation/deactivation (for future event system)
		_ = wasActive // Could emit events here
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
		// Skip if already active
		if s.ActiveSynergies[entityID] != nil && s.ActiveSynergies[entityID][synergyID] {
			continue
		}

		// Check if any secondary skill is trained (but not all at min level)
		anyTrained := false
		for _, secondary := range synergy.SecondarySkills {
			if skills.Levels[secondary] > 0 {
				anyTrained = true
				break
			}
		}

		if anyTrained {
			potentials = append(potentials, synergy)
		}
	}
	return potentials
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
