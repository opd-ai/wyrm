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

// ============================================================================
// Faction-Specific Quest Arcs
// ============================================================================

// FactionQuestArc represents a multi-quest story arc specific to a faction.
type FactionQuestArc struct {
	ID               string            // Unique arc identifier
	FactionID        string            // Owning faction
	Name             string            // Display name for the arc
	Description      string            // Arc description shown to player
	Quests           []FactionArcQuest // Ordered list of quests in the arc
	RequiredRank     int               // Minimum faction rank to start
	MutuallyExcludes []string          // Arc IDs that become unavailable if this arc is chosen
	IsExclusive      bool              // If true, player can only pursue one exclusive arc per faction
	RewardRank       int               // Rank gained upon completion
	RewardItems      []string          // Item IDs rewarded
	Genre            string            // Genre this arc is designed for
}

// FactionArcQuest represents a single quest within a faction arc.
type FactionArcQuest struct {
	ID           string            // Quest ID within the arc
	Title        string            // Quest title
	Description  string            // Quest description
	Objectives   []ArcQuestGoal    // Goals to complete
	RewardXP     int               // XP granted on completion
	RewardGold   int               // Gold granted on completion
	NextQuestID  string            // ID of the next quest (empty if last)
	BranchQuests []string          // Alternative quest IDs (branching paths)
	Consequences []ArcConsequence  // World state changes on completion
	DialogHints  map[string]string // NPC dialog hints for objectives
}

// ArcQuestGoal defines a single objective within a faction arc quest.
type ArcQuestGoal struct {
	Type        string // "kill", "fetch", "deliver", "talk", "explore", "defend"
	Target      string // Target identifier (NPC name, item ID, location)
	Count       int    // Number required (for kill/fetch)
	Description string // Human-readable goal description
}

// ArcConsequence represents a world state change from completing an arc quest.
type ArcConsequence struct {
	Type      string  // "reputation", "unlock", "spawn", "remove", "flag"
	Target    string  // Target faction/NPC/item/flag
	Value     float64 // Numeric value (reputation change, etc.)
	StringVal string  // String value (flag name, unlock ID)
}

// FactionArcManager manages faction-specific quest arcs.
type FactionArcManager struct {
	Arcs           map[string]*FactionQuestArc    // Arc ID -> Arc
	FactionArcs    map[string][]string            // Faction ID -> Arc IDs
	PlayerProgress map[uint64]*FactionArcProgress // Player entity -> progress
	genre          string
}

// FactionArcProgress tracks a player's progress through faction arcs.
type FactionArcProgress struct {
	ActiveArcs      map[string]string // Faction ID -> active arc ID
	CompletedArcs   map[string]bool   // Arc ID -> completed
	CurrentQuests   map[string]string // Arc ID -> current quest ID
	QuestProgress   map[string][]bool // Quest ID -> objective completion
	LockedArcs      map[string]bool   // Arc ID -> locked (mutually exclusive)
	FactionRanks    map[string]int    // Faction ID -> current rank
	UnlockedContent map[string]bool   // Content ID -> unlocked
}

// NewFactionArcManager creates a new faction arc manager with default arcs.
func NewFactionArcManager(genre string) *FactionArcManager {
	m := &FactionArcManager{
		Arcs:           make(map[string]*FactionQuestArc),
		FactionArcs:    make(map[string][]string),
		PlayerProgress: make(map[uint64]*FactionArcProgress),
		genre:          genre,
	}
	m.registerDefaultArcs()
	return m
}

// registerDefaultArcs adds genre-appropriate faction arcs.
func (m *FactionArcManager) registerDefaultArcs() {
	arcs := getGenreFactionArcs(m.genre)
	for _, arc := range arcs {
		m.RegisterArc(arc)
	}
}

// RegisterArc adds a faction arc to the manager.
func (m *FactionArcManager) RegisterArc(arc *FactionQuestArc) {
	m.Arcs[arc.ID] = arc
	m.FactionArcs[arc.FactionID] = append(m.FactionArcs[arc.FactionID], arc.ID)
}

// GetFactionArcs returns all arcs for a faction.
func (m *FactionArcManager) GetFactionArcs(factionID string) []*FactionQuestArc {
	arcIDs := m.FactionArcs[factionID]
	arcs := make([]*FactionQuestArc, 0, len(arcIDs))
	for _, id := range arcIDs {
		if arc, ok := m.Arcs[id]; ok {
			arcs = append(arcs, arc)
		}
	}
	return arcs
}

// GetAvailableArcs returns arcs available to a player based on rank and exclusivity.
func (m *FactionArcManager) GetAvailableArcs(playerEntity uint64, factionID string) []*FactionQuestArc {
	progress := m.getOrCreateProgress(playerEntity)
	allArcs := m.GetFactionArcs(factionID)
	available := make([]*FactionQuestArc, 0)

	playerRank := progress.FactionRanks[factionID]

	for _, arc := range allArcs {
		if m.isArcAvailable(arc, progress, playerRank) {
			available = append(available, arc)
		}
	}
	return available
}

// isArcAvailable checks if an arc is available to a player.
func (m *FactionArcManager) isArcAvailable(arc *FactionQuestArc, progress *FactionArcProgress, playerRank int) bool {
	// Already completed
	if progress.CompletedArcs[arc.ID] {
		return false
	}
	// Locked by exclusivity
	if progress.LockedArcs[arc.ID] {
		return false
	}
	// Rank requirement
	if playerRank < arc.RequiredRank {
		return false
	}
	// Already pursuing an exclusive arc for this faction
	if arc.IsExclusive && progress.ActiveArcs[arc.FactionID] != "" {
		activeArcID := progress.ActiveArcs[arc.FactionID]
		if activeArc, ok := m.Arcs[activeArcID]; ok && activeArc.IsExclusive {
			return false
		}
	}
	return true
}

// StartArc begins a faction arc for a player.
func (m *FactionArcManager) StartArc(playerEntity uint64, arcID string) bool {
	arc, ok := m.Arcs[arcID]
	if !ok || len(arc.Quests) == 0 {
		return false
	}

	progress := m.getOrCreateProgress(playerEntity)

	// Check availability
	playerRank := progress.FactionRanks[arc.FactionID]
	if !m.isArcAvailable(arc, progress, playerRank) {
		return false
	}

	// Lock mutually exclusive arcs
	for _, excludedID := range arc.MutuallyExcludes {
		progress.LockedArcs[excludedID] = true
	}

	// Start the arc
	progress.ActiveArcs[arc.FactionID] = arcID
	progress.CurrentQuests[arcID] = arc.Quests[0].ID

	// Initialize quest progress
	progress.QuestProgress[arc.Quests[0].ID] = make([]bool, len(arc.Quests[0].Objectives))

	return true
}

// GetCurrentQuest returns the current quest in an arc for a player.
func (m *FactionArcManager) GetCurrentQuest(playerEntity uint64, arcID string) *FactionArcQuest {
	arc, ok := m.Arcs[arcID]
	if !ok {
		return nil
	}

	progress := m.getOrCreateProgress(playerEntity)
	questID := progress.CurrentQuests[arcID]

	for i := range arc.Quests {
		if arc.Quests[i].ID == questID {
			return &arc.Quests[i]
		}
	}
	return nil
}

// CompleteObjective marks an objective as complete and checks for quest/arc completion.
func (m *FactionArcManager) CompleteObjective(playerEntity uint64, arcID string, objectiveIndex int) bool {
	arc, ok := m.Arcs[arcID]
	if !ok {
		return false
	}

	progress := m.getOrCreateProgress(playerEntity)
	questID := progress.CurrentQuests[arcID]

	// Find and update quest progress
	var quest *FactionArcQuest
	var questIndex int
	for i := range arc.Quests {
		if arc.Quests[i].ID == questID {
			quest = &arc.Quests[i]
			questIndex = i
			break
		}
	}
	if quest == nil {
		return false
	}

	if objectiveIndex < 0 || objectiveIndex >= len(quest.Objectives) {
		return false
	}

	// Mark objective complete
	if progress.QuestProgress[questID] == nil {
		progress.QuestProgress[questID] = make([]bool, len(quest.Objectives))
	}
	progress.QuestProgress[questID][objectiveIndex] = true

	// Check if all objectives complete
	allComplete := true
	for _, done := range progress.QuestProgress[questID] {
		if !done {
			allComplete = false
			break
		}
	}

	if allComplete {
		return m.completeQuest(playerEntity, arc, quest, questIndex, progress)
	}
	return true
}

// completeQuest finishes a quest and advances the arc.
func (m *FactionArcManager) completeQuest(playerEntity uint64, arc *FactionQuestArc, quest *FactionArcQuest, questIndex int, progress *FactionArcProgress) bool {
	// Check if this was the last quest
	if questIndex >= len(arc.Quests)-1 || quest.NextQuestID == "" {
		return m.completeArc(playerEntity, arc, progress)
	}

	// Advance to next quest
	nextQuestID := quest.NextQuestID
	progress.CurrentQuests[arc.ID] = nextQuestID

	// Initialize next quest progress
	for i := range arc.Quests {
		if arc.Quests[i].ID == nextQuestID {
			progress.QuestProgress[nextQuestID] = make([]bool, len(arc.Quests[i].Objectives))
			break
		}
	}

	return true
}

// completeArc finishes a faction arc and applies rewards.
func (m *FactionArcManager) completeArc(playerEntity uint64, arc *FactionQuestArc, progress *FactionArcProgress) bool {
	progress.CompletedArcs[arc.ID] = true
	delete(progress.ActiveArcs, arc.FactionID)

	// Apply rank reward
	if arc.RewardRank > 0 {
		progress.FactionRanks[arc.FactionID] += arc.RewardRank
	}

	// Unlock exclusive content
	progress.UnlockedContent["arc_"+arc.ID] = true

	return true
}

// getOrCreateProgress returns or creates progress tracking for a player.
func (m *FactionArcManager) getOrCreateProgress(playerEntity uint64) *FactionArcProgress {
	if progress, ok := m.PlayerProgress[playerEntity]; ok {
		return progress
	}

	progress := &FactionArcProgress{
		ActiveArcs:      make(map[string]string),
		CompletedArcs:   make(map[string]bool),
		CurrentQuests:   make(map[string]string),
		QuestProgress:   make(map[string][]bool),
		LockedArcs:      make(map[string]bool),
		FactionRanks:    make(map[string]int),
		UnlockedContent: make(map[string]bool),
	}
	m.PlayerProgress[playerEntity] = progress
	return progress
}

// GetArcProgress returns a player's progress in a specific arc.
func (m *FactionArcManager) GetArcProgress(playerEntity uint64, arcID string) (questID string, objectivesDone, objectivesTotal int) {
	arc, ok := m.Arcs[arcID]
	if !ok {
		return "", 0, 0
	}

	progress := m.getOrCreateProgress(playerEntity)
	questID = progress.CurrentQuests[arcID]

	for i := range arc.Quests {
		if arc.Quests[i].ID == questID {
			objectivesTotal = len(arc.Quests[i].Objectives)
			if prog, ok := progress.QuestProgress[questID]; ok {
				for _, done := range prog {
					if done {
						objectivesDone++
					}
				}
			}
			break
		}
	}
	return questID, objectivesDone, objectivesTotal
}

// IsArcComplete checks if a player has completed an arc.
func (m *FactionArcManager) IsArcComplete(playerEntity uint64, arcID string) bool {
	progress := m.getOrCreateProgress(playerEntity)
	return progress.CompletedArcs[arcID]
}

// GetCompletedArcs returns all completed arc IDs for a player.
func (m *FactionArcManager) GetCompletedArcs(playerEntity uint64) []string {
	progress := m.getOrCreateProgress(playerEntity)
	completed := make([]string, 0)
	for arcID, done := range progress.CompletedArcs {
		if done {
			completed = append(completed, arcID)
		}
	}
	return completed
}

// getGenreFactionArcs returns default faction arcs for a genre.
func getGenreFactionArcs(genre string) []*FactionQuestArc {
	switch genre {
	case "fantasy":
		return getFantasyFactionArcs()
	case "sci-fi":
		return getSciFiFactionArcs()
	case "horror":
		return getHorrorFactionArcs()
	case "cyberpunk":
		return getCyberpunkFactionArcs()
	case "post-apocalyptic":
		return getPostApocFactionArcs()
	default:
		return getFantasyFactionArcs()
	}
}

// getFantasyFactionArcs returns fantasy genre faction arcs.
func getFantasyFactionArcs() []*FactionQuestArc {
	return []*FactionQuestArc{
		{
			ID:           "thieves_guild_heist",
			FactionID:    "thieves_guild",
			Name:         "The Great Heist",
			Description:  "Rise through the ranks by pulling off increasingly daring heists.",
			RequiredRank: 1,
			IsExclusive:  true,
			RewardRank:   3,
			Genre:        "fantasy",
			Quests: []FactionArcQuest{
				{
					ID:          "heist_1_recon",
					Title:       "Eyes on the Prize",
					Description: "Scout the merchant's vault and report back.",
					Objectives: []ArcQuestGoal{
						{Type: "explore", Target: "merchant_vault", Count: 1, Description: "Scout the vault entrance"},
						{Type: "talk", Target: "fence_npc", Count: 1, Description: "Report to the fence"},
					},
					RewardXP:    200,
					RewardGold:  50,
					NextQuestID: "heist_2_setup",
				},
				{
					ID:          "heist_2_setup",
					Title:       "Tools of the Trade",
					Description: "Acquire the tools needed for the heist.",
					Objectives: []ArcQuestGoal{
						{Type: "fetch", Target: "lockpicks", Count: 3, Description: "Obtain master lockpicks"},
						{Type: "fetch", Target: "smoke_bomb", Count: 2, Description: "Get smoke bombs"},
					},
					RewardXP:    300,
					RewardGold:  75,
					NextQuestID: "heist_3_execution",
				},
				{
					ID:          "heist_3_execution",
					Title:       "The Big Score",
					Description: "Execute the heist and escape with the loot.",
					Objectives: []ArcQuestGoal{
						{Type: "explore", Target: "merchant_vault_interior", Count: 1, Description: "Break into the vault"},
						{Type: "fetch", Target: "golden_chalice", Count: 1, Description: "Steal the Golden Chalice"},
						{Type: "deliver", Target: "fence_npc", Count: 1, Description: "Deliver the goods"},
					},
					RewardXP:   500,
					RewardGold: 200,
				},
			},
		},
		{
			ID:               "mages_guild_arcane",
			FactionID:        "mages_guild",
			Name:             "Path of the Arcane",
			Description:      "Uncover ancient magical secrets and prove your worth to the guild.",
			RequiredRank:     0,
			MutuallyExcludes: []string{"mages_guild_forbidden"},
			IsExclusive:      true,
			RewardRank:       2,
			Genre:            "fantasy",
			Quests: []FactionArcQuest{
				{
					ID:          "arcane_1_initiation",
					Title:       "The First Spell",
					Description: "Demonstrate basic magical aptitude.",
					Objectives: []ArcQuestGoal{
						{Type: "talk", Target: "archmage", Count: 1, Description: "Speak with the Archmage"},
						{Type: "kill", Target: "training_golem", Count: 3, Description: "Defeat training golems with magic"},
					},
					RewardXP:    150,
					RewardGold:  30,
					NextQuestID: "arcane_2_research",
				},
				{
					ID:          "arcane_2_research",
					Title:       "Lost Knowledge",
					Description: "Recover ancient tomes from the ruined library.",
					Objectives: []ArcQuestGoal{
						{Type: "explore", Target: "ruined_library", Count: 1, Description: "Find the ruined library"},
						{Type: "fetch", Target: "ancient_tome", Count: 2, Description: "Recover the tomes"},
					},
					RewardXP:   350,
					RewardGold: 100,
				},
			},
		},
	}
}

// getSciFiFactionArcs returns sci-fi genre faction arcs.
func getSciFiFactionArcs() []*FactionQuestArc {
	return []*FactionQuestArc{
		{
			ID:           "corporate_espionage",
			FactionID:    "megacorp",
			Name:         "Corporate Warfare",
			Description:  "Engage in high-stakes corporate espionage against rival corporations.",
			RequiredRank: 1,
			IsExclusive:  true,
			RewardRank:   2,
			Genre:        "sci-fi",
			Quests: []FactionArcQuest{
				{
					ID:          "corp_1_infiltrate",
					Title:       "Inside Job",
					Description: "Infiltrate the rival corp's research facility.",
					Objectives: []ArcQuestGoal{
						{Type: "explore", Target: "rival_facility", Count: 1, Description: "Enter the facility"},
						{Type: "fetch", Target: "access_card", Count: 1, Description: "Acquire security clearance"},
					},
					RewardXP:    250,
					RewardGold:  100,
					NextQuestID: "corp_2_steal",
				},
				{
					ID:          "corp_2_steal",
					Title:       "Data Heist",
					Description: "Download the prototype schematics.",
					Objectives: []ArcQuestGoal{
						{Type: "explore", Target: "server_room", Count: 1, Description: "Access the server room"},
						{Type: "fetch", Target: "prototype_data", Count: 1, Description: "Download the data"},
						{Type: "kill", Target: "security_bot", Count: 5, Description: "Eliminate security"},
					},
					RewardXP:   500,
					RewardGold: 300,
				},
			},
		},
	}
}

// getHorrorFactionArcs returns horror genre faction arcs.
func getHorrorFactionArcs() []*FactionQuestArc {
	return []*FactionQuestArc{
		{
			ID:           "cult_investigation",
			FactionID:    "survivors",
			Name:         "Into the Darkness",
			Description:  "Investigate the cult that plagues the land and uncover their dark secrets.",
			RequiredRank: 0,
			IsExclusive:  false,
			RewardRank:   1,
			Genre:        "horror",
			Quests: []FactionArcQuest{
				{
					ID:          "cult_1_clues",
					Title:       "Gathering Evidence",
					Description: "Find evidence of cult activity in the abandoned church.",
					Objectives: []ArcQuestGoal{
						{Type: "explore", Target: "abandoned_church", Count: 1, Description: "Search the church"},
						{Type: "fetch", Target: "cult_symbol", Count: 3, Description: "Collect cult artifacts"},
					},
					RewardXP:    200,
					RewardGold:  40,
					NextQuestID: "cult_2_ritual",
				},
				{
					ID:          "cult_2_ritual",
					Title:       "Disrupting the Ritual",
					Description: "Stop the cult before they complete their summoning.",
					Objectives: []ArcQuestGoal{
						{Type: "kill", Target: "cultist", Count: 10, Description: "Eliminate the cultists"},
						{Type: "explore", Target: "ritual_chamber", Count: 1, Description: "Reach the ritual chamber"},
						{Type: "kill", Target: "cult_leader", Count: 1, Description: "Stop the High Priest"},
					},
					RewardXP:   600,
					RewardGold: 150,
				},
			},
		},
	}
}

// getCyberpunkFactionArcs returns cyberpunk genre faction arcs.
func getCyberpunkFactionArcs() []*FactionQuestArc {
	return []*FactionQuestArc{
		{
			ID:           "gang_takeover",
			FactionID:    "street_gang",
			Name:         "Rise to Power",
			Description:  "Help the gang expand their territory and take down rivals.",
			RequiredRank: 0,
			IsExclusive:  true,
			RewardRank:   2,
			Genre:        "cyberpunk",
			Quests: []FactionArcQuest{
				{
					ID:          "gang_1_muscle",
					Title:       "Proving Ground",
					Description: "Show the gang you're worth their time.",
					Objectives: []ArcQuestGoal{
						{Type: "kill", Target: "rival_thug", Count: 5, Description: "Take out rival thugs"},
						{Type: "talk", Target: "gang_boss", Count: 1, Description: "Report to the boss"},
					},
					RewardXP:    200,
					RewardGold:  75,
					NextQuestID: "gang_2_territory",
				},
				{
					ID:          "gang_2_territory",
					Title:       "Hostile Takeover",
					Description: "Seize control of the rival gang's operations.",
					Objectives: []ArcQuestGoal{
						{Type: "explore", Target: "rival_hideout", Count: 1, Description: "Storm the hideout"},
						{Type: "kill", Target: "rival_boss", Count: 1, Description: "Eliminate the rival boss"},
						{Type: "defend", Target: "gang_territory", Count: 1, Description: "Hold the position"},
					},
					RewardXP:   500,
					RewardGold: 250,
				},
			},
		},
	}
}

// getPostApocFactionArcs returns post-apocalyptic genre faction arcs.
func getPostApocFactionArcs() []*FactionQuestArc {
	return []*FactionQuestArc{
		{
			ID:           "settlement_defense",
			FactionID:    "settlement",
			Name:         "Defender of the Wastes",
			Description:  "Protect the settlement from raiders and secure vital resources.",
			RequiredRank: 0,
			IsExclusive:  false,
			RewardRank:   1,
			Genre:        "post-apocalyptic",
			Quests: []FactionArcQuest{
				{
					ID:          "settle_1_scout",
					Title:       "Eyes in the Wastes",
					Description: "Scout the surrounding area for threats.",
					Objectives: []ArcQuestGoal{
						{Type: "explore", Target: "raider_camp", Count: 1, Description: "Locate the raider camp"},
						{Type: "talk", Target: "settlement_leader", Count: 1, Description: "Report your findings"},
					},
					RewardXP:    150,
					RewardGold:  30,
					NextQuestID: "settle_2_strike",
				},
				{
					ID:          "settle_2_strike",
					Title:       "Preemptive Strike",
					Description: "Attack the raiders before they attack you.",
					Objectives: []ArcQuestGoal{
						{Type: "kill", Target: "raider", Count: 15, Description: "Eliminate the raiders"},
						{Type: "fetch", Target: "supplies", Count: 5, Description: "Recover stolen supplies"},
					},
					RewardXP:    400,
					RewardGold:  120,
					NextQuestID: "settle_3_fortify",
				},
				{
					ID:          "settle_3_fortify",
					Title:       "Last Stand",
					Description: "Prepare and defend against the raider counterattack.",
					Objectives: []ArcQuestGoal{
						{Type: "fetch", Target: "building_materials", Count: 10, Description: "Gather fortification materials"},
						{Type: "defend", Target: "settlement_gate", Count: 1, Description: "Survive the raid"},
					},
					RewardXP:   600,
					RewardGold: 200,
				},
			},
		},
	}
}
