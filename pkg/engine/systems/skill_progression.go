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
