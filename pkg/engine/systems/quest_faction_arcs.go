package systems

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

	// Find the quest
	quest, questIndex := m.findQuest(arc, questID)
	if quest == nil {
		return false
	}

	// Validate objective index
	if objectiveIndex < 0 || objectiveIndex >= len(quest.Objectives) {
		return false
	}

	// Mark objective complete
	m.markObjectiveComplete(progress, questID, objectiveIndex, len(quest.Objectives))

	// Check if quest is complete
	if m.isQuestComplete(progress, questID) {
		return m.completeQuest(playerEntity, arc, quest, questIndex, progress)
	}
	return true
}

// findQuest locates a quest in an arc by ID.
func (m *FactionArcManager) findQuest(arc *FactionQuestArc, questID string) (*FactionArcQuest, int) {
	for i := range arc.Quests {
		if arc.Quests[i].ID == questID {
			return &arc.Quests[i], i
		}
	}
	return nil, -1
}

// markObjectiveComplete marks a specific objective as done.
func (m *FactionArcManager) markObjectiveComplete(progress *FactionArcProgress, questID string, objectiveIndex, totalObjectives int) {
	if progress.QuestProgress[questID] == nil {
		progress.QuestProgress[questID] = make([]bool, totalObjectives)
	}
	progress.QuestProgress[questID][objectiveIndex] = true
}

// isQuestComplete checks if all objectives are done.
func (m *FactionArcManager) isQuestComplete(progress *FactionArcProgress, questID string) bool {
	for _, done := range progress.QuestProgress[questID] {
		if !done {
			return false
		}
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
	objectivesTotal, objectivesDone = m.countQuestObjectives(arc, questID, progress)
	return questID, objectivesDone, objectivesTotal
}

// countQuestObjectives counts completed and total objectives for a quest.
func (m *FactionArcManager) countQuestObjectives(arc *FactionQuestArc, questID string, progress *FactionArcProgress) (total, done int) {
	for i := range arc.Quests {
		if arc.Quests[i].ID != questID {
			continue
		}
		total = len(arc.Quests[i].Objectives)
		done = m.countCompletedObjectives(progress, questID)
		return total, done
	}
	return 0, 0
}

// countCompletedObjectives counts how many objectives are completed in a quest.
func (m *FactionArcManager) countCompletedObjectives(progress *FactionArcProgress, questID string) int {
	prog, ok := progress.QuestProgress[questID]
	if !ok {
		return 0
	}
	count := 0
	for _, done := range prog {
		if done {
			count++
		}
	}
	return count
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
