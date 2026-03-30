package systems

import (
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// QuestSystem manages quest state, branching, and consequence flags.
type QuestSystem struct {
	// QuestStages maps quest ID to list of stage conditions.
	QuestStages map[string][]QuestStageCondition
}

// QuestStageCondition defines what must be true to advance a quest stage.
type QuestStageCondition struct {
	RequiredFlag string // Flag that must be true to advance
	FromStage    int    // Stage this condition applies from
	NextStage    int    // Stage to advance to
	Completes    bool   // If true, this transition completes the quest
	LocksBranch  string // Branch ID to lock when this transition is taken
	BranchID     string // Branch ID this condition belongs to (blocked if locked)
}

// NewQuestSystem creates a new quest system.
func NewQuestSystem() *QuestSystem {
	return &QuestSystem{
		QuestStages: make(map[string][]QuestStageCondition),
	}
}

// DefineQuest adds stage conditions for a quest.
func (s *QuestSystem) DefineQuest(questID string, stages []QuestStageCondition) {
	if s.QuestStages == nil {
		s.QuestStages = make(map[string][]QuestStageCondition)
	}
	s.QuestStages[questID] = stages
}

// Update processes quest state transitions each tick.
func (s *QuestSystem) Update(w *ecs.World, dt float64) {
	if s.QuestStages == nil {
		s.QuestStages = make(map[string][]QuestStageCondition)
	}
	for _, e := range w.Entities("Quest") {
		comp, ok := w.GetComponent(e, "Quest")
		if !ok {
			continue
		}
		quest := comp.(*components.Quest)
		s.processQuestStage(quest)
	}
}

// processQuestStage checks and advances a single quest's stage.
func (s *QuestSystem) processQuestStage(quest *components.Quest) {
	if quest.Completed {
		return
	}
	if quest.Flags == nil {
		quest.Flags = make(map[string]bool)
	}
	stages, ok := s.QuestStages[quest.ID]
	if !ok {
		return
	}
	s.checkStageConditions(quest, stages)
}

// checkStageConditions evaluates stage conditions and advances the quest.
func (s *QuestSystem) checkStageConditions(quest *components.Quest, stages []QuestStageCondition) {
	for _, cond := range stages {
		if cond.FromStage != quest.CurrentStage {
			continue
		}
		// Skip if this branch is locked
		if cond.BranchID != "" && quest.IsBranchLocked(cond.BranchID) {
			continue
		}
		if quest.Flags[cond.RequiredFlag] {
			s.advanceQuest(quest, cond)
			break
		}
	}
}

// advanceQuest moves the quest to the next stage or completes it.
func (s *QuestSystem) advanceQuest(quest *components.Quest, cond QuestStageCondition) {
	// Lock the competing branch if specified
	if cond.LocksBranch != "" {
		quest.LockBranch(cond.LocksBranch)
	}
	if cond.Completes {
		quest.Completed = true
	} else {
		quest.CurrentStage = cond.NextStage
	}
}

// WorldState represents current world conditions for dynamic quest generation.
type WorldState struct {
	FamineLevel      float64 // 0-1, how severe the famine is
	WarIntensity     float64 // 0-1, how intense the ongoing conflicts
	BanditActivity   float64 // 0-1, bandit threat level
	PlagueSeverity   float64 // 0-1, disease outbreak severity
	MonsterThreat    float64 // 0-1, monster activity level
	TreasureRumors   float64 // 0-1, rumors of hidden treasure
	PoliticalUnrest  float64 // 0-1, faction tensions
	ResourceScarcity float64 // 0-1, how scarce resources are
}

// DynamicQuestType defines categories of generated quests.
type DynamicQuestType int

const (
	QuestTypeFetch DynamicQuestType = iota
	QuestTypeKill
	QuestTypeEscort
	QuestTypeInvestigate
	QuestTypeDeliver
	QuestTypeRescue
	QuestTypeSabotage
	QuestTypeNegotiate
	QuestTypeExplore
	QuestTypeDefend
)

// String returns the human-readable name of the quest type.
func (q DynamicQuestType) String() string {
	switch q {
	case QuestTypeFetch:
		return "fetch"
	case QuestTypeKill:
		return "kill"
	case QuestTypeEscort:
		return "escort"
	case QuestTypeInvestigate:
		return "investigate"
	case QuestTypeDeliver:
		return "deliver"
	case QuestTypeRescue:
		return "rescue"
	case QuestTypeSabotage:
		return "sabotage"
	case QuestTypeNegotiate:
		return "negotiate"
	case QuestTypeExplore:
		return "explore"
	case QuestTypeDefend:
		return "defend"
	default:
		return "unknown"
	}
}

// DynamicQuest represents a procedurally generated quest.
type DynamicQuest struct {
	ID           string
	Type         DynamicQuestType
	Title        string
	Description  string
	Giver        string // NPC name who gives the quest
	TargetName   string // Target NPC/location/item
	Reward       int    // Currency reward
	XPReward     int    // Experience reward
	Difficulty   int    // 1-5 difficulty rating
	TimeLimit    int    // Hours to complete (0 = no limit)
	WorldTrigger string // Which world state triggered this quest
}

// RadiantQuestTemplate defines a template for radiant quests.
type RadiantQuestTemplate struct {
	Type          DynamicQuestType
	TitlePattern  string // e.g., "Fetch %s for %s"
	DescPattern   string // Description template
	MinDifficulty int
	MaxDifficulty int
	BaseReward    int
	BaseXP        int
	Targets       []string // Possible targets for template
	Givers        []string // Possible quest givers
}

// RadiantQuestConfig holds templates for radiant quest generation.
type RadiantQuestConfig struct {
	Templates map[string][]RadiantQuestTemplate // genre -> templates
}

// DefaultRadiantConfig returns default radiant quest templates for all genres.
func DefaultRadiantConfig() *RadiantQuestConfig {
	return &RadiantQuestConfig{
		Templates: map[string][]RadiantQuestTemplate{
			"fantasy": {
				{
					Type:          QuestTypeFetch,
					TitlePattern:  "Gather %s",
					DescPattern:   "Collect %s and return to %s.",
					MinDifficulty: 1,
					MaxDifficulty: 3,
					BaseReward:    50,
					BaseXP:        100,
					Targets:       []string{"healing herbs", "mushrooms", "ore", "rare flowers", "gemstones"},
					Givers:        []string{"herbalist", "alchemist", "merchant", "blacksmith"},
				},
				{
					Type:          QuestTypeKill,
					TitlePattern:  "Slay the %s",
					DescPattern:   "A %s has been terrorizing the area. Eliminate the threat and return to %s.",
					MinDifficulty: 2,
					MaxDifficulty: 5,
					BaseReward:    100,
					BaseXP:        200,
					Targets:       []string{"wolves", "bandits", "goblin raiders", "troll", "dragon cultists"},
					Givers:        []string{"guard captain", "village elder", "innkeeper", "noble"},
				},
				{
					Type:          QuestTypeDeliver,
					TitlePattern:  "Deliver the %s",
					DescPattern:   "Transport a %s to the destination safely.",
					MinDifficulty: 1,
					MaxDifficulty: 2,
					BaseReward:    30,
					BaseXP:        75,
					Targets:       []string{"letter", "package", "medicine", "shipment", "artifact"},
					Givers:        []string{"courier", "merchant", "priest", "noble"},
				},
			},
			"sci-fi": {
				{
					Type:          QuestTypeFetch,
					TitlePattern:  "Salvage %s",
					DescPattern:   "Locate and retrieve %s from the wreckage. Report to %s.",
					MinDifficulty: 1,
					MaxDifficulty: 3,
					BaseReward:    50,
					BaseXP:        100,
					Targets:       []string{"power cells", "data cores", "circuit boards", "fuel rods", "quantum chips"},
					Givers:        []string{"engineer", "technician", "station commander", "scientist"},
				},
				{
					Type:          QuestTypeKill,
					TitlePattern:  "Neutralize %s",
					DescPattern:   "Hostile %s detected. Engage and neutralize. Report to %s.",
					MinDifficulty: 2,
					MaxDifficulty: 5,
					BaseReward:    100,
					BaseXP:        200,
					Targets:       []string{"rogue drones", "pirates", "alien hostiles", "mutants", "infected crew"},
					Givers:        []string{"commander", "security chief", "admiral", "station AI"},
				},
			},
			"horror": {
				{
					Type:          QuestTypeInvestigate,
					TitlePattern:  "Investigate the %s",
					DescPattern:   "Strange occurrences at the %s. Discover the truth and survive.",
					MinDifficulty: 2,
					MaxDifficulty: 4,
					BaseReward:    60,
					BaseXP:        150,
					Targets:       []string{"abandoned asylum", "haunted mansion", "ritual site", "crypt", "cursed village"},
					Givers:        []string{"survivor", "priest", "sheriff", "occult researcher"},
				},
				{
					Type:          QuestTypeRescue,
					TitlePattern:  "Save the %s",
					DescPattern:   "Someone is trapped in the %s. Find them before it's too late.",
					MinDifficulty: 3,
					MaxDifficulty: 5,
					BaseReward:    100,
					BaseXP:        200,
					Targets:       []string{"lost child", "missing researcher", "kidnapped survivor", "trapped explorer"},
					Givers:        []string{"grieving parent", "colleague", "priest", "survivor"},
				},
			},
			"cyberpunk": {
				{
					Type:          QuestTypeSabotage,
					TitlePattern:  "Hit the %s",
					DescPattern:   "The %s is the target. Infiltrate and sabotage their operations.",
					MinDifficulty: 3,
					MaxDifficulty: 5,
					BaseReward:    150,
					BaseXP:        250,
					Targets:       []string{"corpo server farm", "gang hideout", "megacorp HQ", "black clinic", "data vault"},
					Givers:        []string{"fixer", "netrunner", "gang boss", "corporate rival"},
				},
				{
					Type:          QuestTypeDeliver,
					TitlePattern:  "Move the %s",
					DescPattern:   "Transport the %s across the city. No questions asked.",
					MinDifficulty: 1,
					MaxDifficulty: 3,
					BaseReward:    75,
					BaseXP:        100,
					Targets:       []string{"data chip", "black market goods", "wetware", "stolen tech", "evidence"},
					Givers:        []string{"fixer", "street vendor", "smuggler", "info broker"},
				},
			},
			"post-apocalyptic": {
				{
					Type:          QuestTypeFetch,
					TitlePattern:  "Scavenge %s",
					DescPattern:   "We need %s for the settlement. Search the ruins.",
					MinDifficulty: 1,
					MaxDifficulty: 3,
					BaseReward:    40,
					BaseXP:        100,
					Targets:       []string{"food supplies", "medicine", "clean water", "fuel", "weapons"},
					Givers:        []string{"settlement leader", "trader", "doctor", "mechanic"},
				},
				{
					Type:          QuestTypeDefend,
					TitlePattern:  "Defend the %s",
					DescPattern:   "Raiders are targeting the %s. Help us hold them off!",
					MinDifficulty: 2,
					MaxDifficulty: 5,
					BaseReward:    100,
					BaseXP:        200,
					Targets:       []string{"water purifier", "settlement walls", "trading post", "caravan"},
					Givers:        []string{"settlement leader", "caravan master", "militia captain", "trader"},
				},
			},
		},
	}
}

// DynamicQuestGenerator generates quests based on world state.
type DynamicQuestGenerator struct {
	config   *RadiantQuestConfig
	seed     int64
	questSeq int64 // Sequence number for unique IDs
}

// NewDynamicQuestGenerator creates a new quest generator.
func NewDynamicQuestGenerator(seed int64) *DynamicQuestGenerator {
	return &DynamicQuestGenerator{
		config: DefaultRadiantConfig(),
		seed:   seed,
	}
}

// GenerateFromWorldState creates quests based on current world conditions.
func (g *DynamicQuestGenerator) GenerateFromWorldState(worldState *WorldState, genre string) []*DynamicQuest {
	var quests []*DynamicQuest

	triggers := []worldStateTrigger{
		{worldState.FamineLevel, "famine", g.generateFamineQuest},
		{worldState.WarIntensity, "war", g.generateWarQuest},
		{worldState.BanditActivity, "bandits", g.generateBanditQuest},
		{worldState.MonsterThreat, "monsters", g.generateMonsterQuest},
		{worldState.PoliticalUnrest, "politics", g.generatePoliticalQuest},
	}

	for _, t := range triggers {
		if quest := g.generateTriggeredQuest(t, genre); quest != nil {
			quests = append(quests, quest)
		}
	}

	return quests
}

// worldStateTrigger pairs a threshold value with its quest generator.
type worldStateTrigger struct {
	value     float64
	trigger   string
	generator func(string) *DynamicQuest
}

// generateTriggeredQuest creates a quest if the trigger threshold is exceeded.
func (g *DynamicQuestGenerator) generateTriggeredQuest(t worldStateTrigger, genre string) *DynamicQuest {
	if t.value <= WorldStateHighThreshold {
		return nil
	}
	quest := t.generator(genre)
	if quest != nil {
		quest.WorldTrigger = t.trigger
	}
	return quest
}

// generateFamineQuest creates a food-related quest.
func (g *DynamicQuestGenerator) generateFamineQuest(genre string) *DynamicQuest {
	g.questSeq++
	targets := map[string]string{
		"fantasy":          "food supplies",
		"sci-fi":           "ration packs",
		"horror":           "preserved food",
		"cyberpunk":        "synth-food crates",
		"post-apocalyptic": "food supplies",
	}
	target := targets[genre]
	if target == "" {
		target = "supplies"
	}

	return &DynamicQuest{
		ID:          g.generateQuestID("famine"),
		Type:        QuestTypeFetch,
		Title:       "Urgent: Gather " + target,
		Description: "The people are starving. Gather " + target + " from the surrounding area.",
		Giver:       "settlement leader",
		TargetName:  target,
		Reward:      75,
		XPReward:    150,
		Difficulty:  3,
		TimeLimit:   24,
	}
}

// generateWarQuest creates a combat quest related to war.
func (g *DynamicQuestGenerator) generateWarQuest(genre string) *DynamicQuest {
	g.questSeq++
	targets := map[string]string{
		"fantasy":          "enemy scouts",
		"sci-fi":           "hostile patrol",
		"horror":           "cultist warband",
		"cyberpunk":        "gang strike force",
		"post-apocalyptic": "raider war party",
	}
	target := targets[genre]
	if target == "" {
		target = "enemies"
	}

	return &DynamicQuest{
		ID:          g.generateQuestID("war"),
		Type:        QuestTypeKill,
		Title:       "Wartime: Eliminate " + target,
		Description: "The war effort requires action. Eliminate the " + target + " threatening our position.",
		Giver:       "commander",
		TargetName:  target,
		Reward:      150,
		XPReward:    300,
		Difficulty:  4,
		TimeLimit:   0,
	}
}

// generateBanditQuest creates an anti-bandit quest.
func (g *DynamicQuestGenerator) generateBanditQuest(genre string) *DynamicQuest {
	g.questSeq++
	return &DynamicQuest{
		ID:          g.generateQuestID("bandit"),
		Type:        QuestTypeKill,
		Title:       "Clear the Roads",
		Description: "Bandits are preying on travelers. Make the roads safe again.",
		Giver:       "merchant",
		TargetName:  "bandits",
		Reward:      100,
		XPReward:    200,
		Difficulty:  3,
		TimeLimit:   0,
	}
}

// generateMonsterQuest creates a monster hunting quest.
func (g *DynamicQuestGenerator) generateMonsterQuest(genre string) *DynamicQuest {
	g.questSeq++
	targets := map[string]string{
		"fantasy":          "the beast",
		"sci-fi":           "the alien creature",
		"horror":           "the abomination",
		"cyberpunk":        "the rogue cyborg",
		"post-apocalyptic": "the mutant horror",
	}
	target := targets[genre]
	if target == "" {
		target = "the monster"
	}

	return &DynamicQuest{
		ID:          g.generateQuestID("monster"),
		Type:        QuestTypeKill,
		Title:       "Hunt: " + target,
		Description: "A terrible creature threatens the area. Track it down and end its reign of terror.",
		Giver:       "hunter",
		TargetName:  target,
		Reward:      200,
		XPReward:    400,
		Difficulty:  5,
		TimeLimit:   0,
	}
}

// generatePoliticalQuest creates a diplomatic quest.
func (g *DynamicQuestGenerator) generatePoliticalQuest(genre string) *DynamicQuest {
	g.questSeq++
	return &DynamicQuest{
		ID:          g.generateQuestID("politics"),
		Type:        QuestTypeNegotiate,
		Title:       "Broker Peace",
		Description: "Tensions are rising. Negotiate a peace between the factions before war breaks out.",
		Giver:       "diplomat",
		TargetName:  "faction leaders",
		Reward:      100,
		XPReward:    250,
		Difficulty:  3,
		TimeLimit:   48,
	}
}

// generateQuestID creates a unique quest identifier.
func (g *DynamicQuestGenerator) generateQuestID(prefix string) string {
	return prefix + "_" + string(rune('A'+int(g.questSeq%26))) + string(rune('0'+int(g.questSeq/26)%10))
}

// RadiantQuestBoard represents a notice board that generates radiant quests.
type RadiantQuestBoard struct {
	LocationID   string
	Genre        string
	config       *RadiantQuestConfig
	questSeq     int64
	maxQuests    int
	activeQuests []*DynamicQuest
}

// NewRadiantQuestBoard creates a notice board at a location.
func NewRadiantQuestBoard(locationID, genre string) *RadiantQuestBoard {
	return &RadiantQuestBoard{
		LocationID:   locationID,
		Genre:        genre,
		config:       DefaultRadiantConfig(),
		maxQuests:    5,
		activeQuests: make([]*DynamicQuest, 0),
	}
}

// RefreshQuests generates new radiant quests for the board.
func (b *RadiantQuestBoard) RefreshQuests(seed int64) {
	b.activeQuests = make([]*DynamicQuest, 0, b.maxQuests)

	templates := b.config.Templates[b.Genre]
	if len(templates) == 0 {
		templates = b.config.Templates["fantasy"] // fallback
	}

	// Generate quests using seed for determinism
	for i := 0; i < b.maxQuests && i < len(templates)*2; i++ {
		templateIdx := int(seed+int64(i)) % len(templates)
		template := templates[templateIdx]

		quest := b.generateFromTemplate(template, seed+int64(i))
		if quest != nil {
			b.activeQuests = append(b.activeQuests, quest)
		}
	}
}

// generateFromTemplate creates a quest from a radiant template.
func (b *RadiantQuestBoard) generateFromTemplate(template RadiantQuestTemplate, seed int64) *DynamicQuest {
	b.questSeq++

	if len(template.Targets) == 0 || len(template.Givers) == 0 {
		return nil
	}

	targetIdx := int(seed) % len(template.Targets)
	giverIdx := int(seed/7) % len(template.Givers)

	target := template.Targets[targetIdx]
	giver := template.Givers[giverIdx]

	// Calculate difficulty based on seed
	diffRange := template.MaxDifficulty - template.MinDifficulty + 1
	difficulty := template.MinDifficulty + int(seed%int64(diffRange))

	// Calculate reward based on difficulty
	reward := template.BaseReward + (difficulty-1)*25
	xp := template.BaseXP + (difficulty-1)*50

	return &DynamicQuest{
		ID:          b.generateQuestID(),
		Type:        template.Type,
		Title:       formatPattern(template.TitlePattern, target, giver),
		Description: formatPattern(template.DescPattern, target, giver),
		Giver:       giver,
		TargetName:  target,
		Reward:      reward,
		XPReward:    xp,
		Difficulty:  difficulty,
		TimeLimit:   0,
	}
}

// generateQuestID creates a unique radiant quest ID.
func (b *RadiantQuestBoard) generateQuestID() string {
	return b.LocationID + "_radiant_" + string(rune('A'+int(b.questSeq%26))) + string(rune('0'+int(b.questSeq/26)%10))
}

// GetActiveQuests returns current available quests.
func (b *RadiantQuestBoard) GetActiveQuests() []*DynamicQuest {
	return b.activeQuests
}

// QuestCount returns number of active quests.
func (b *RadiantQuestBoard) QuestCount() int {
	return len(b.activeQuests)
}

// formatPattern replaces %s placeholders with values.
func formatPattern(pattern, target, giver string) string {
	result := pattern
	// Replace first %s with target
	idx := findPercentS(result)
	if idx >= 0 {
		result = result[:idx] + target + result[idx+2:]
	}
	// Replace second %s (now first in modified string) with giver
	idx = findPercentS(result)
	if idx >= 0 {
		result = result[:idx] + giver + result[idx+2:]
	}
	return result
}

// findPercentS finds the first occurrence of %s.
func findPercentS(s string) int {
	for i := 0; i < len(s)-1; i++ {
		if s[i] == '%' && s[i+1] == 's' {
			return i
		}
	}
	return -1
}
