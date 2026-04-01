// Package systems implements ECS system logic.
package systems

import (
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/seedutil"
)

// ============================================================================
// Criminal Faction Questlines
// ============================================================================

// CriminalQuestType represents different types of criminal quests.
type CriminalQuestType int

const (
	// CriminalQuestTheft is a simple theft mission.
	CriminalQuestTheft CriminalQuestType = iota
	// CriminalQuestSmuggling is transporting illegal goods.
	CriminalQuestSmuggling
	// CriminalQuestHeist is a complex robbery.
	CriminalQuestHeist
	// CriminalQuestAssassination is eliminating a target.
	CriminalQuestAssassination
	// CriminalQuestExtortion is threatening for payment.
	CriminalQuestExtortion
	// CriminalQuestSabotage is destroying enemy resources.
	CriminalQuestSabotage
	// CriminalQuestInfiltration is placing a mole in an organization.
	CriminalQuestInfiltration
	// CriminalQuestJailbreak is freeing imprisoned members.
	CriminalQuestJailbreak
	// CriminalQuestTurf is territorial expansion.
	CriminalQuestTurf
	// CriminalQuestBoss is a leadership challenge.
	CriminalQuestBoss
)

// String returns the quest type name.
func (q CriminalQuestType) String() string {
	switch q {
	case CriminalQuestTheft:
		return "Theft"
	case CriminalQuestSmuggling:
		return "Smuggling"
	case CriminalQuestHeist:
		return "Heist"
	case CriminalQuestAssassination:
		return "Assassination"
	case CriminalQuestExtortion:
		return "Extortion"
	case CriminalQuestSabotage:
		return "Sabotage"
	case CriminalQuestInfiltration:
		return "Infiltration"
	case CriminalQuestJailbreak:
		return "Jailbreak"
	case CriminalQuestTurf:
		return "Turf War"
	case CriminalQuestBoss:
		return "Leadership Challenge"
	default:
		return "Unknown"
	}
}

// MinRank returns the minimum faction rank required for this quest type.
func (q CriminalQuestType) MinRank() int {
	switch q {
	case CriminalQuestTheft:
		return 1
	case CriminalQuestSmuggling:
		return 2
	case CriminalQuestExtortion:
		return 3
	case CriminalQuestSabotage:
		return 4
	case CriminalQuestHeist:
		return 5
	case CriminalQuestInfiltration:
		return 6
	case CriminalQuestAssassination:
		return 7
	case CriminalQuestJailbreak:
		return 8
	case CriminalQuestTurf:
		return 9
	case CriminalQuestBoss:
		return 10
	default:
		return 1
	}
}

// BaseXPReward returns the base XP reward for completing this quest type.
func (q CriminalQuestType) BaseXPReward() int {
	switch q {
	case CriminalQuestTheft:
		return 50
	case CriminalQuestSmuggling:
		return 75
	case CriminalQuestExtortion:
		return 100
	case CriminalQuestSabotage:
		return 150
	case CriminalQuestHeist:
		return 200
	case CriminalQuestInfiltration:
		return 250
	case CriminalQuestAssassination:
		return 300
	case CriminalQuestJailbreak:
		return 350
	case CriminalQuestTurf:
		return 400
	case CriminalQuestBoss:
		return 500
	default:
		return 50
	}
}

// BaseCurrencyReward returns the base currency reward.
func (q CriminalQuestType) BaseCurrencyReward() int {
	switch q {
	case CriminalQuestTheft:
		return 100
	case CriminalQuestSmuggling:
		return 200
	case CriminalQuestExtortion:
		return 300
	case CriminalQuestSabotage:
		return 250
	case CriminalQuestHeist:
		return 500
	case CriminalQuestInfiltration:
		return 350
	case CriminalQuestAssassination:
		return 450
	case CriminalQuestJailbreak:
		return 400
	case CriminalQuestTurf:
		return 600
	case CriminalQuestBoss:
		return 1000
	default:
		return 100
	}
}

// CriminalQuestState represents the current state of a quest.
type CriminalQuestState int

const (
	// CriminalQuestAvailable means the quest can be accepted.
	CriminalQuestAvailable CriminalQuestState = iota
	// CriminalQuestActive means the quest is in progress.
	CriminalQuestActive
	// CriminalQuestCompleted means the quest was successfully finished.
	CriminalQuestCompleted
	// CriminalQuestFailed means the quest was failed.
	CriminalQuestFailed
	// CriminalQuestExpired means the quest timed out.
	CriminalQuestExpired
)

// CriminalQuest represents a criminal faction quest.
type CriminalQuest struct {
	ID                 string             // Unique quest identifier
	FactionID          string             // Criminal faction offering the quest
	Type               CriminalQuestType  // Type of criminal activity
	State              CriminalQuestState // Current quest state
	Title              string             // Quest title
	Description        string             // Quest description
	Objectives         []QuestObjective   // List of objectives
	CurrentStage       int                // Current stage (0-indexed)
	AssignedTo         uint64             // Entity ID of player
	GiverID            uint64             // NPC who gave the quest
	TargetID           uint64             // Target entity (if applicable)
	TargetLocationX    float64            // Target X coordinate
	TargetLocationZ    float64            // Target Z coordinate
	StartTime          float64            // When the quest was accepted
	TimeLimit          float64            // Time limit in seconds (0 = unlimited)
	Difficulty         int                // 1-5 difficulty rating
	IsStealthy         bool               // Whether stealth is required
	RivalFactionID     string             // Rival faction involved (if any)
	RewardXP           int                // XP reward
	RewardCurrency     int                // Currency reward
	RewardItems        []string           // Item rewards
	RewardReputation   float64            // Reputation gain with faction
	ConsequencesOnFail []string           // What happens if failed
}

// QuestObjective represents a single quest objective.
type QuestObjective struct {
	ID          string // Objective identifier
	Description string // What the player must do
	IsCompleted bool   // Whether this objective is done
	IsOptional  bool   // Whether this objective is optional
	Progress    int    // Current progress count
	Required    int    // Required count (for collect/kill quests)
	TargetType  string // Type of target (item, NPC, location, etc.)
	TargetID    string // Specific target identifier
}

// CriminalFactionQuestSystem manages criminal faction questlines.
type CriminalFactionQuestSystem struct {
	factionRankSystem *FactionRankSystem
	// Quest storage
	quests          map[string]*CriminalQuest
	questsByFaction map[string][]string // FactionID -> QuestIDs
	questsByPlayer  map[uint64][]string // PlayerID -> QuestIDs
	activeQuests    map[uint64]string   // PlayerID -> current active quest
	// Generation settings
	genre string
	// Tracking
	gameTime    float64
	nextQuestID int
	// Random generator for determinism
	rng *PseudoRandomLCG
	// Quest generation settings
	QuestsPerRankTier     int     // Number of quests available per rank tier
	QuestRefreshTime      float64 // Time before new quests generate
	FailureReputationLoss float64 // Reputation lost on failure
}

// NewCriminalFactionQuestSystem creates a new criminal quest system.
func NewCriminalFactionQuestSystem(factionRankSystem *FactionRankSystem, genre string, seed int64) *CriminalFactionQuestSystem {
	return &CriminalFactionQuestSystem{
		factionRankSystem:     factionRankSystem,
		quests:                make(map[string]*CriminalQuest),
		questsByFaction:       make(map[string][]string),
		questsByPlayer:        make(map[uint64][]string),
		activeQuests:          make(map[uint64]string),
		genre:                 genre,
		rng:                   NewPseudoRandomLCG(seed),
		QuestsPerRankTier:     3,
		QuestRefreshTime:      3600.0, // 1 hour
		FailureReputationLoss: -10.0,
	}
}

// Update processes active quests and generates new ones.
func (s *CriminalFactionQuestSystem) Update(w *ecs.World, dt float64) {
	s.gameTime += dt
	s.processActiveQuests(w)
}

// processActiveQuests checks time limits and updates quest states.
func (s *CriminalFactionQuestSystem) processActiveQuests(w *ecs.World) {
	for _, quest := range s.quests {
		if quest.State != CriminalQuestActive {
			continue
		}
		// Check time limit
		if quest.TimeLimit > 0 {
			elapsed := s.gameTime - quest.StartTime
			if elapsed > quest.TimeLimit {
				s.failQuest(w, quest, "Time expired")
			}
		}
	}
}

// GenerateQuestsForFaction creates available quests for a criminal faction.
func (s *CriminalFactionQuestSystem) GenerateQuestsForFaction(factionID string, playerRank int) []*CriminalQuest {
	var generated []*CriminalQuest
	// Generate quests for ranks at or below player's rank
	questTypes := s.getAvailableQuestTypes(playerRank)
	for i := 0; i < s.QuestsPerRankTier && i < len(questTypes); i++ {
		questType := questTypes[i]
		quest := s.generateQuest(factionID, questType)
		s.quests[quest.ID] = quest
		s.questsByFaction[factionID] = append(s.questsByFaction[factionID], quest.ID)
		generated = append(generated, quest)
	}
	return generated
}

// getAvailableQuestTypes returns quest types available for a given rank.
func (s *CriminalFactionQuestSystem) getAvailableQuestTypes(rank int) []CriminalQuestType {
	allTypes := []CriminalQuestType{
		CriminalQuestTheft, CriminalQuestSmuggling, CriminalQuestExtortion,
		CriminalQuestSabotage, CriminalQuestHeist, CriminalQuestInfiltration,
		CriminalQuestAssassination, CriminalQuestJailbreak, CriminalQuestTurf,
		CriminalQuestBoss,
	}
	var available []CriminalQuestType
	for _, qt := range allTypes {
		if qt.MinRank() <= rank {
			available = append(available, qt)
		}
	}
	// Shuffle based on RNG
	for i := len(available) - 1; i > 0; i-- {
		j := int(s.pseudoRandom() * float64(i+1))
		available[i], available[j] = available[j], available[i]
	}
	return available
}

// generateQuest creates a new quest of the specified type.
func (s *CriminalFactionQuestSystem) generateQuest(factionID string, questType CriminalQuestType) *CriminalQuest {
	s.nextQuestID++
	id := formatQuestID(s.nextQuestID)
	difficulty := questType.MinRank()/2 + 1
	if difficulty > 5 {
		difficulty = 5
	}
	quest := &CriminalQuest{
		ID:               id,
		FactionID:        factionID,
		Type:             questType,
		State:            CriminalQuestAvailable,
		Title:            s.generateQuestTitle(questType),
		Description:      s.generateQuestDescription(questType),
		Objectives:       s.generateObjectives(questType),
		CurrentStage:     0,
		Difficulty:       difficulty,
		IsStealthy:       s.isStealthyQuest(questType),
		RewardXP:         questType.BaseXPReward() * difficulty,
		RewardCurrency:   questType.BaseCurrencyReward() * difficulty,
		RewardReputation: float64(difficulty) * 5.0,
		TimeLimit:        s.getTimeLimit(questType),
	}
	return quest
}

// formatQuestID creates a quest ID string.
func formatQuestID(n int) string {
	return seedutil.FormatPrefixedID("CQ", n)
}

// isStealthyQuest determines if a quest requires stealth.
func (s *CriminalFactionQuestSystem) isStealthyQuest(questType CriminalQuestType) bool {
	switch questType {
	case CriminalQuestTheft, CriminalQuestSmuggling, CriminalQuestInfiltration, CriminalQuestHeist:
		return true
	default:
		return false
	}
}

// getTimeLimit returns the time limit for a quest type.
func (s *CriminalFactionQuestSystem) getTimeLimit(questType CriminalQuestType) float64 {
	switch questType {
	case CriminalQuestTheft:
		return 1800.0 // 30 minutes
	case CriminalQuestSmuggling:
		return 3600.0 // 1 hour
	case CriminalQuestExtortion:
		return 2400.0 // 40 minutes
	case CriminalQuestHeist:
		return 3600.0 // 1 hour
	case CriminalQuestAssassination:
		return 7200.0 // 2 hours
	default:
		return 0 // No time limit
	}
}

// generateQuestTitle creates a genre-appropriate quest title.
func (s *CriminalFactionQuestSystem) generateQuestTitle(questType CriminalQuestType) string {
	switch s.genre {
	case "fantasy":
		return s.fantasyQuestTitle(questType)
	case "sci-fi":
		return s.sciFiQuestTitle(questType)
	case "horror":
		return s.horrorQuestTitle(questType)
	case "cyberpunk":
		return s.cyberpunkQuestTitle(questType)
	case "post-apocalyptic":
		return s.postApocQuestTitle(questType)
	default:
		return s.fantasyQuestTitle(questType)
	}
}

func (s *CriminalFactionQuestSystem) fantasyQuestTitle(questType CriminalQuestType) string {
	switch questType {
	case CriminalQuestTheft:
		return "A Simple Acquisition"
	case CriminalQuestSmuggling:
		return "Moonlight Cargo"
	case CriminalQuestExtortion:
		return "Protection Money"
	case CriminalQuestHeist:
		return "The Vault Job"
	case CriminalQuestAssassination:
		return "Silent Blade"
	case CriminalQuestSabotage:
		return "Breaking the Competition"
	case CriminalQuestInfiltration:
		return "The Inside Man"
	case CriminalQuestJailbreak:
		return "Spring the Rogue"
	case CriminalQuestTurf:
		return "Claiming the Streets"
	case CriminalQuestBoss:
		return "The Throne of Shadows"
	default:
		return "Shady Business"
	}
}

func (s *CriminalFactionQuestSystem) sciFiQuestTitle(questType CriminalQuestType) string {
	switch questType {
	case CriminalQuestTheft:
		return "Data Extraction"
	case CriminalQuestSmuggling:
		return "Black Market Run"
	case CriminalQuestExtortion:
		return "Corporate Leverage"
	case CriminalQuestHeist:
		return "Station Break-In"
	case CriminalQuestAssassination:
		return "Target Elimination"
	case CriminalQuestSabotage:
		return "System Corruption"
	case CriminalQuestInfiltration:
		return "Deep Cover"
	case CriminalQuestJailbreak:
		return "Prison Transport Intercept"
	case CriminalQuestTurf:
		return "Sector Control"
	case CriminalQuestBoss:
		return "Syndicate Takeover"
	default:
		return "Illegal Operations"
	}
}

func (s *CriminalFactionQuestSystem) horrorQuestTitle(questType CriminalQuestType) string {
	switch questType {
	case CriminalQuestTheft:
		return "Forbidden Relics"
	case CriminalQuestSmuggling:
		return "Cursed Cargo"
	case CriminalQuestExtortion:
		return "Blood Money"
	case CriminalQuestHeist:
		return "Tomb Robbery"
	case CriminalQuestAssassination:
		return "The Ritual Sacrifice"
	case CriminalQuestSabotage:
		return "Desecration"
	case CriminalQuestInfiltration:
		return "Among the Cultists"
	case CriminalQuestJailbreak:
		return "Release the Damned"
	case CriminalQuestTurf:
		return "Unholy Ground"
	case CriminalQuestBoss:
		return "The Dark Throne"
	default:
		return "Sinister Dealings"
	}
}

func (s *CriminalFactionQuestSystem) cyberpunkQuestTitle(questType CriminalQuestType) string {
	switch questType {
	case CriminalQuestTheft:
		return "Smash and Grab"
	case CriminalQuestSmuggling:
		return "Contraband Courier"
	case CriminalQuestExtortion:
		return "Digital Blackmail"
	case CriminalQuestHeist:
		return "The Big Score"
	case CriminalQuestAssassination:
		return "Flatline Contract"
	case CriminalQuestSabotage:
		return "Corporate Sabotage"
	case CriminalQuestInfiltration:
		return "Netrunner Mole"
	case CriminalQuestJailbreak:
		return "Bust Out"
	case CriminalQuestTurf:
		return "Neon Territory"
	case CriminalQuestBoss:
		return "King of the Underworld"
	default:
		return "Street Crime"
	}
}

func (s *CriminalFactionQuestSystem) postApocQuestTitle(questType CriminalQuestType) string {
	switch questType {
	case CriminalQuestTheft:
		return "Scavenger Run"
	case CriminalQuestSmuggling:
		return "Wasteland Smuggler"
	case CriminalQuestExtortion:
		return "Toll Collection"
	case CriminalQuestHeist:
		return "Vault Raid"
	case CriminalQuestAssassination:
		return "Wasteland Justice"
	case CriminalQuestSabotage:
		return "Sabotage the Settlers"
	case CriminalQuestInfiltration:
		return "Spy in the Bunker"
	case CriminalQuestJailbreak:
		return "Free the Raiders"
	case CriminalQuestTurf:
		return "Territory Expansion"
	case CriminalQuestBoss:
		return "Warlord's Challenge"
	default:
		return "Raider Business"
	}
}

// generateQuestDescription creates a detailed quest description.
func (s *CriminalFactionQuestSystem) generateQuestDescription(questType CriminalQuestType) string {
	switch questType {
	case CriminalQuestTheft:
		return "Acquire the target item and bring it back without getting caught."
	case CriminalQuestSmuggling:
		return "Transport the contraband to the drop point while avoiding authorities."
	case CriminalQuestExtortion:
		return "Convince the target to pay protection money through intimidation."
	case CriminalQuestHeist:
		return "Plan and execute a complex robbery of a high-security target."
	case CriminalQuestAssassination:
		return "Eliminate the designated target permanently."
	case CriminalQuestSabotage:
		return "Destroy or disable the enemy's valuable assets."
	case CriminalQuestInfiltration:
		return "Gain the trust of the rival organization and gather intelligence."
	case CriminalQuestJailbreak:
		return "Free our imprisoned associate from custody."
	case CriminalQuestTurf:
		return "Expand our territory by taking control of the contested area."
	case CriminalQuestBoss:
		return "Challenge the current leadership and prove your worth."
	default:
		return "Complete the assigned criminal task."
	}
}

// generateObjectives creates objectives for a quest type.
func (s *CriminalFactionQuestSystem) generateObjectives(questType CriminalQuestType) []QuestObjective {
	switch questType {
	case CriminalQuestTheft:
		return []QuestObjective{
			{ID: "locate", Description: "Locate the target item", Required: 1},
			{ID: "acquire", Description: "Acquire the item", Required: 1},
			{ID: "escape", Description: "Escape without detection", Required: 1},
			{ID: "deliver", Description: "Deliver to the fence", Required: 1},
		}
	case CriminalQuestSmuggling:
		return []QuestObjective{
			{ID: "pickup", Description: "Pick up the cargo", Required: 1},
			{ID: "avoid", Description: "Avoid patrol checkpoints", Required: 3, IsOptional: true},
			{ID: "deliver", Description: "Deliver to destination", Required: 1},
		}
	case CriminalQuestExtortion:
		return []QuestObjective{
			{ID: "confront", Description: "Confront the target", Required: 1},
			{ID: "threaten", Description: "Make your demands clear", Required: 1},
			{ID: "collect", Description: "Collect payment", Required: 1},
		}
	case CriminalQuestHeist:
		return []QuestObjective{
			{ID: "scout", Description: "Scout the location", Required: 1},
			{ID: "disable", Description: "Disable security systems", Required: 1},
			{ID: "breach", Description: "Enter the secure area", Required: 1},
			{ID: "loot", Description: "Acquire the valuables", Required: 1},
			{ID: "escape", Description: "Escape with the loot", Required: 1},
		}
	case CriminalQuestAssassination:
		return []QuestObjective{
			{ID: "locate", Description: "Locate the target", Required: 1},
			{ID: "approach", Description: "Get close to the target", Required: 1},
			{ID: "eliminate", Description: "Eliminate the target", Required: 1},
			{ID: "escape", Description: "Leave no witnesses", Required: 1, IsOptional: true},
		}
	case CriminalQuestSabotage:
		return []QuestObjective{
			{ID: "infiltrate", Description: "Infiltrate the target location", Required: 1},
			{ID: "destroy", Description: "Destroy the equipment", Required: 3},
			{ID: "escape", Description: "Escape before discovery", Required: 1},
		}
	case CriminalQuestInfiltration:
		return []QuestObjective{
			{ID: "contact", Description: "Make initial contact", Required: 1},
			{ID: "trust", Description: "Build trust with the target", Required: 3},
			{ID: "intel", Description: "Gather intelligence", Required: 1},
			{ID: "report", Description: "Report back to your handler", Required: 1},
		}
	case CriminalQuestJailbreak:
		return []QuestObjective{
			{ID: "plan", Description: "Plan the escape route", Required: 1},
			{ID: "distract", Description: "Create a distraction", Required: 1},
			{ID: "free", Description: "Free the prisoner", Required: 1},
			{ID: "escape", Description: "Escape to safety", Required: 1},
		}
	case CriminalQuestTurf:
		return []QuestObjective{
			{ID: "intimidate", Description: "Intimidate local businesses", Required: 5},
			{ID: "drive_out", Description: "Drive out rival presence", Required: 3},
			{ID: "establish", Description: "Establish control point", Required: 1},
		}
	case CriminalQuestBoss:
		return []QuestObjective{
			{ID: "challenge", Description: "Issue the challenge", Required: 1},
			{ID: "duel", Description: "Defeat the current leader", Required: 1},
			{ID: "claim", Description: "Claim leadership", Required: 1},
		}
	default:
		return []QuestObjective{
			{ID: "complete", Description: "Complete the task", Required: 1},
		}
	}
}

// AcceptQuest assigns a quest to a player.
func (s *CriminalFactionQuestSystem) AcceptQuest(w *ecs.World, questID string, playerID uint64) bool {
	quest, ok := s.quests[questID]
	if !ok || quest.State != CriminalQuestAvailable {
		return false
	}
	if !s.canPlayerAcceptQuest(w, playerID, quest) {
		return false
	}
	s.activateQuest(quest, playerID, questID)
	return true
}

// canPlayerAcceptQuest checks if a player meets all requirements to accept a quest.
func (s *CriminalFactionQuestSystem) canPlayerAcceptQuest(w *ecs.World, playerID uint64, quest *CriminalQuest) bool {
	// Check if player already has an active quest
	if _, hasActive := s.activeQuests[playerID]; hasActive {
		return false
	}
	return s.hasRequiredRank(w, playerID, quest)
}

// hasRequiredRank checks if the player has sufficient faction rank for the quest.
func (s *CriminalFactionQuestSystem) hasRequiredRank(w *ecs.World, playerID uint64, quest *CriminalQuest) bool {
	if s.factionRankSystem == nil {
		return true
	}
	comp, ok := w.GetComponent(ecs.Entity(playerID), "FactionMembership")
	if !ok {
		return true
	}
	membership := comp.(*components.FactionMembership)
	return membership.GetRank(quest.FactionID) >= quest.Type.MinRank()
}

// activateQuest transitions a quest to the active state for the player.
func (s *CriminalFactionQuestSystem) activateQuest(quest *CriminalQuest, playerID uint64, questID string) {
	quest.State = CriminalQuestActive
	quest.AssignedTo = playerID
	quest.StartTime = s.gameTime
	s.activeQuests[playerID] = questID
	s.questsByPlayer[playerID] = append(s.questsByPlayer[playerID], questID)
}

// CompleteObjective marks an objective as completed.
func (s *CriminalFactionQuestSystem) CompleteObjective(questID, objectiveID string) bool {
	quest, ok := s.quests[questID]
	if !ok || quest.State != CriminalQuestActive {
		return false
	}
	for i := range quest.Objectives {
		if quest.Objectives[i].ID == objectiveID && !quest.Objectives[i].IsCompleted {
			quest.Objectives[i].Progress++
			if quest.Objectives[i].Progress >= quest.Objectives[i].Required {
				quest.Objectives[i].IsCompleted = true
			}
			return true
		}
	}
	return false
}

// CheckQuestComplete checks if all required objectives are done.
func (s *CriminalFactionQuestSystem) CheckQuestComplete(questID string) bool {
	quest, ok := s.quests[questID]
	if !ok {
		return false
	}
	for _, obj := range quest.Objectives {
		if !obj.IsOptional && !obj.IsCompleted {
			return false
		}
	}
	return true
}

// CompleteQuest marks a quest as completed and awards rewards.
func (s *CriminalFactionQuestSystem) CompleteQuest(w *ecs.World, questID string) bool {
	quest, ok := s.quests[questID]
	if !ok || quest.State != CriminalQuestActive {
		return false
	}
	if !s.CheckQuestComplete(questID) {
		return false
	}
	quest.State = CriminalQuestCompleted
	// Remove from active quests
	delete(s.activeQuests, quest.AssignedTo)
	// Award rewards
	s.awardRewards(w, quest)
	return true
}

// awardRewards gives the player quest rewards.
func (s *CriminalFactionQuestSystem) awardRewards(w *ecs.World, quest *CriminalQuest) {
	if s.factionRankSystem != nil {
		// Add XP to faction
		s.factionRankSystem.AddXP(w, ecs.Entity(quest.AssignedTo), quest.FactionID, quest.RewardXP)
	}
	// In a full implementation, would also:
	// - Add currency to player inventory
	// - Add items to player inventory
	// - Update reputation
}

// failQuest handles quest failure.
func (s *CriminalFactionQuestSystem) failQuest(w *ecs.World, quest *CriminalQuest, reason string) {
	quest.State = CriminalQuestFailed
	quest.ConsequencesOnFail = append(quest.ConsequencesOnFail, reason)
	// Remove from active quests
	delete(s.activeQuests, quest.AssignedTo)
	// Apply reputation loss
	if s.factionRankSystem != nil {
		comp, ok := w.GetComponent(ecs.Entity(quest.AssignedTo), "FactionMembership")
		if ok {
			membership := comp.(*components.FactionMembership)
			info := membership.GetMembership(quest.FactionID)
			if info != nil {
				info.Reputation += s.FailureReputationLoss
				if info.Reputation < -100 {
					info.Reputation = -100
				}
			}
		}
	}
}

// AbandonQuest allows player to give up on a quest.
func (s *CriminalFactionQuestSystem) AbandonQuest(w *ecs.World, questID string) bool {
	quest, ok := s.quests[questID]
	if !ok || quest.State != CriminalQuestActive {
		return false
	}
	s.failQuest(w, quest, "Abandoned")
	return true
}

// GetQuestProgress returns progress information for a quest.
func (s *CriminalFactionQuestSystem) GetQuestProgress(questID string) (completed, total int) {
	quest, ok := s.quests[questID]
	if !ok {
		return 0, 0
	}
	for _, obj := range quest.Objectives {
		if !obj.IsOptional {
			total++
			if obj.IsCompleted {
				completed++
			}
		}
	}
	return completed, total
}

// GetQuest returns a quest by ID.
func (s *CriminalFactionQuestSystem) GetQuest(questID string) *CriminalQuest {
	return s.quests[questID]
}

// GetActiveQuest returns the player's active quest.
func (s *CriminalFactionQuestSystem) GetActiveQuest(playerID uint64) *CriminalQuest {
	questID, ok := s.activeQuests[playerID]
	if !ok {
		return nil
	}
	return s.quests[questID]
}

// GetAvailableQuests returns available quests for a faction.
func (s *CriminalFactionQuestSystem) GetAvailableQuests(factionID string) []*CriminalQuest {
	var available []*CriminalQuest
	for _, questID := range s.questsByFaction[factionID] {
		quest := s.quests[questID]
		if quest.State == CriminalQuestAvailable {
			available = append(available, quest)
		}
	}
	return available
}

// GetPlayerQuestHistory returns all quests a player has taken.
func (s *CriminalFactionQuestSystem) GetPlayerQuestHistory(playerID uint64) []*CriminalQuest {
	var history []*CriminalQuest
	for _, questID := range s.questsByPlayer[playerID] {
		history = append(history, s.quests[questID])
	}
	return history
}

// GetRemainingTime returns time remaining on a timed quest.
func (s *CriminalFactionQuestSystem) GetRemainingTime(questID string) float64 {
	quest, ok := s.quests[questID]
	if !ok || quest.TimeLimit <= 0 {
		return -1 // No time limit
	}
	elapsed := s.gameTime - quest.StartTime
	remaining := quest.TimeLimit - elapsed
	if remaining < 0 {
		return 0
	}
	return remaining
}

// pseudoRandom generates a deterministic pseudo-random number 0.0-1.0.
func (s *CriminalFactionQuestSystem) pseudoRandom() float64 {
	return s.rng.Float64()
}
