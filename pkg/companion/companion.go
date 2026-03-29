// Package companion provides persistent companion NPCs with AI behavior.
// Per ROADMAP Phase 6 item 27:
// AC: Companion uses class-appropriate abilities;
// dialog references player actions from last 10 events.
package companion

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Personality represents a companion's personality type.
type Personality int

const (
	PersonalityBrave Personality = iota
	PersonalityCautious
	PersonalityLoyal
	PersonalityAggressive
	PersonalityWise
)

// String returns the human-readable name of the personality type.
func (p Personality) String() string {
	switch p {
	case PersonalityBrave:
		return "brave"
	case PersonalityCautious:
		return "cautious"
	case PersonalityLoyal:
		return "loyal"
	case PersonalityAggressive:
		return "aggressive"
	case PersonalityWise:
		return "wise"
	default:
		return "unknown"
	}
}

// CombatRole defines the companion's combat behavior.
type CombatRole int

const (
	RoleTank CombatRole = iota
	RoleDPS
	RoleHealer
	RoleSupport
	RoleRanged
)

// String returns the human-readable name of the combat role.
func (r CombatRole) String() string {
	switch r {
	case RoleTank:
		return "tank"
	case RoleDPS:
		return "damage"
	case RoleHealer:
		return "healer"
	case RoleSupport:
		return "support"
	case RoleRanged:
		return "ranged"
	default:
		return "unknown"
	}
}

// ActionEvent represents a player action that companions can remember.
type ActionEvent struct {
	EventType   string
	Description string
	Target      string
	Outcome     string
	Timestamp   time.Time
}

// Companion represents a persistent companion NPC.
type Companion struct {
	ID          uint64
	Name        string
	Genre       string
	Personality Personality
	Role        CombatRole
	Backstory   string

	// Class and abilities
	Class     string
	Abilities []Ability

	// Relationship with player
	Loyalty    float64 // 0-100
	Morale     float64 // 0-100
	TrustLevel float64 // 0-100

	// Memory of player actions
	PlayerEvents []ActionEvent

	// Current state
	InCombat     bool
	Following    bool
	CurrentOrder Order
}

// Order represents a command given to the companion.
type Order int

const (
	OrderFollow Order = iota
	OrderStay
	OrderAttack
	OrderDefend
	OrderHeal
	OrderWait
)

// Ability represents a companion ability.
type Ability struct {
	ID          string
	Name        string
	Description string
	Cooldown    time.Duration
	DamageType  string
	IsSupport   bool
	IsHealing   bool
}

// GenreCompanionTemplates defines genre-specific companion archetypes.
var GenreCompanionTemplates = map[string][]CompanionTemplate{
	"fantasy": {
		{Role: RoleTank, Class: "Knight", Backstory: "A disgraced knight seeking redemption"},
		{Role: RoleDPS, Class: "Rogue", Backstory: "A street thief with a heart of gold"},
		{Role: RoleHealer, Class: "Cleric", Backstory: "A wandering healer who lost their faith"},
		{Role: RoleSupport, Class: "Bard", Backstory: "A traveling minstrel collecting stories"},
		{Role: RoleRanged, Class: "Ranger", Backstory: "A forest guardian exiled from their home"},
	},
	"sci-fi": {
		{Role: RoleTank, Class: "Marine", Backstory: "A veteran soldier haunted by past missions"},
		{Role: RoleDPS, Class: "Operative", Backstory: "A corporate assassin who went rogue"},
		{Role: RoleHealer, Class: "Medic", Backstory: "A combat medic sworn to save lives"},
		{Role: RoleSupport, Class: "Hacker", Backstory: "A digital ghost hiding from megacorps"},
		{Role: RoleRanged, Class: "Sniper", Backstory: "A precision shooter seeking justice"},
	},
	"horror": {
		{Role: RoleTank, Class: "Survivor", Backstory: "The last survivor of a doomed expedition"},
		{Role: RoleDPS, Class: "Hunter", Backstory: "A monster hunter tracking an ancient evil"},
		{Role: RoleHealer, Class: "Doctor", Backstory: "A physician who has seen too much death"},
		{Role: RoleSupport, Class: "Medium", Backstory: "A psychic cursed with visions of doom"},
		{Role: RoleRanged, Class: "Investigator", Backstory: "A detective who stumbled onto the truth"},
	},
	"cyberpunk": {
		{Role: RoleTank, Class: "Street Samurai", Backstory: "A chromed enforcer tired of corpo wars"},
		{Role: RoleDPS, Class: "Solo", Backstory: "A freelance mercenary with a code"},
		{Role: RoleHealer, Class: "Ripperdoc", Backstory: "A back-alley surgeon fixing the broken"},
		{Role: RoleSupport, Class: "Netrunner", Backstory: "A hacker who lost their crew to ICE"},
		{Role: RoleRanged, Class: "Techie", Backstory: "A weapons engineer on the run"},
	},
	"post-apocalyptic": {
		{Role: RoleTank, Class: "Bruiser", Backstory: "A wastelander who survived the bombs"},
		{Role: RoleDPS, Class: "Scavenger", Backstory: "A ruthless survivor who found purpose"},
		{Role: RoleHealer, Class: "Herbalist", Backstory: "A healer preserving old-world medicine"},
		{Role: RoleSupport, Class: "Scout", Backstory: "A guide who knows the dead zones"},
		{Role: RoleRanged, Class: "Sharpshooter", Backstory: "A caravan guard protecting survivors"},
	},
}

// CompanionTemplate defines a companion archetype.
type CompanionTemplate struct {
	Role      CombatRole
	Class     string
	Backstory string
}

// RoleAbilities maps combat roles to abilities.
var RoleAbilities = map[CombatRole][]Ability{
	RoleTank: {
		{ID: "shield_wall", Name: "Shield Wall", Description: "Block incoming damage", Cooldown: 30 * time.Second},
		{ID: "taunt", Name: "Taunt", Description: "Draw enemy attention", Cooldown: 15 * time.Second},
		{ID: "fortify", Name: "Fortify", Description: "Increase defense temporarily", Cooldown: 60 * time.Second},
	},
	RoleDPS: {
		{ID: "power_strike", Name: "Power Strike", Description: "High damage attack", Cooldown: 10 * time.Second, DamageType: "physical"},
		{ID: "flurry", Name: "Flurry", Description: "Multiple quick attacks", Cooldown: 20 * time.Second, DamageType: "physical"},
		{ID: "execute", Name: "Execute", Description: "Finish low health enemies", Cooldown: 45 * time.Second, DamageType: "physical"},
	},
	RoleHealer: {
		{ID: "heal", Name: "Heal", Description: "Restore health", Cooldown: 10 * time.Second, IsHealing: true},
		{ID: "cleanse", Name: "Cleanse", Description: "Remove negative effects", Cooldown: 30 * time.Second, IsSupport: true},
		{ID: "revive", Name: "Revive", Description: "Bring back fallen allies", Cooldown: 120 * time.Second, IsHealing: true},
	},
	RoleSupport: {
		{ID: "buff", Name: "Empower", Description: "Increase ally damage", Cooldown: 45 * time.Second, IsSupport: true},
		{ID: "debuff", Name: "Weaken", Description: "Reduce enemy defense", Cooldown: 30 * time.Second, IsSupport: true},
		{ID: "inspire", Name: "Inspire", Description: "Boost ally morale", Cooldown: 60 * time.Second, IsSupport: true},
	},
	RoleRanged: {
		{ID: "aimed_shot", Name: "Aimed Shot", Description: "Precise ranged attack", Cooldown: 8 * time.Second, DamageType: "physical"},
		{ID: "volley", Name: "Volley", Description: "Area attack", Cooldown: 25 * time.Second, DamageType: "physical"},
		{ID: "crippling_shot", Name: "Crippling Shot", Description: "Slow enemy movement", Cooldown: 20 * time.Second, DamageType: "physical"},
	},
}

// CompanionManager handles companion NPCs.
type CompanionManager struct {
	mu         sync.RWMutex
	companions map[uint64]*Companion // CompanionID -> Companion
	playerComps map[uint64]uint64    // PlayerID -> CompanionID
	rng        *rand.Rand
}

// NewCompanionManager creates a new companion manager.
func NewCompanionManager(seed int64) *CompanionManager {
	return &CompanionManager{
		companions:  make(map[uint64]*Companion),
		playerComps: make(map[uint64]uint64),
		rng:         rand.New(rand.NewSource(seed)),
	}
}

// CreateCompanion generates a new companion for a player.
func (cm *CompanionManager) CreateCompanion(playerID uint64, genre string, preferredRole CombatRole) *Companion {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	templates := GenreCompanionTemplates[genre]
	if templates == nil {
		templates = GenreCompanionTemplates["fantasy"]
	}

	// Find template matching preferred role
	var template CompanionTemplate
	for _, t := range templates {
		if t.Role == preferredRole {
			template = t
			break
		}
	}
	if template.Class == "" {
		template = templates[cm.rng.Intn(len(templates))]
	}

	// Generate companion ID
	compID := uint64(cm.rng.Int63())

	companion := &Companion{
		ID:           compID,
		Name:         cm.generateName(genre),
		Genre:        genre,
		Personality:  Personality(cm.rng.Intn(5)),
		Role:         template.Role,
		Backstory:    template.Backstory,
		Class:        template.Class,
		Abilities:    RoleAbilities[template.Role],
		Loyalty:      50,
		Morale:       75,
		TrustLevel:   30,
		PlayerEvents: make([]ActionEvent, 0),
		Following:    true,
		CurrentOrder: OrderFollow,
	}

	cm.companions[compID] = companion
	cm.playerComps[playerID] = compID

	return companion
}

// generateName creates a genre-appropriate name.
func (cm *CompanionManager) generateName(genre string) string {
	names := map[string][]string{
		"fantasy":          {"Aldric", "Lyra", "Thorin", "Seraphina", "Gareth"},
		"sci-fi":           {"Nova", "Rex", "Luna", "Orion", "Zeta"},
		"horror":           {"Victor", "Helena", "Edgar", "Lilith", "Mortimer"},
		"cyberpunk":        {"Razor", "Neon", "Chrome", "Pixel", "Binary"},
		"post-apocalyptic": {"Rust", "Ash", "Storm", "Ember", "Flint"},
	}

	genreNames := names[genre]
	if genreNames == nil {
		genreNames = names["fantasy"]
	}

	return genreNames[cm.rng.Intn(len(genreNames))]
}

// GetCompanion returns a companion by ID.
func (cm *CompanionManager) GetCompanion(compID uint64) *Companion {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.companions[compID]
}

// GetPlayerCompanion returns the companion assigned to a player.
func (cm *CompanionManager) GetPlayerCompanion(playerID uint64) *Companion {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	compID, ok := cm.playerComps[playerID]
	if !ok {
		return nil
	}
	return cm.companions[compID]
}

// RecordPlayerAction stores a player action in companion memory.
// Per AC: dialog references player actions from last 10 events.
func (cm *CompanionManager) RecordPlayerAction(compID uint64, event ActionEvent) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	comp := cm.companions[compID]
	if comp == nil {
		return
	}

	event.Timestamp = time.Now()
	comp.PlayerEvents = append(comp.PlayerEvents, event)

	// Keep only last 10 events (per AC)
	if len(comp.PlayerEvents) > 10 {
		comp.PlayerEvents = comp.PlayerEvents[len(comp.PlayerEvents)-10:]
	}

	// Adjust loyalty based on event type
	cm.adjustLoyalty(comp, event)
}

// adjustLoyalty modifies companion loyalty based on player actions.
func (cm *CompanionManager) adjustLoyalty(comp *Companion, event ActionEvent) {
	switch event.EventType {
	case "helped_npc":
		comp.Loyalty += 2
	case "killed_innocent":
		comp.Loyalty -= 10
	case "shared_loot":
		comp.Loyalty += 5
	case "protected_companion":
		comp.Loyalty += 8
	case "abandoned_companion":
		comp.Loyalty -= 15
	}

	// Clamp loyalty
	if comp.Loyalty > 100 {
		comp.Loyalty = 100
	}
	if comp.Loyalty < 0 {
		comp.Loyalty = 0
	}
}

// GetRecentEvents returns the companion's memory of recent player actions.
func (cm *CompanionManager) GetRecentEvents(compID uint64) []ActionEvent {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	comp := cm.companions[compID]
	if comp == nil {
		return nil
	}
	return comp.PlayerEvents
}

// findAbilityByPredicate returns the first ability matching the predicate, or nil.
func findAbilityByPredicate(abilities []Ability, pred func(*Ability) bool) *Ability {
	for i := range abilities {
		if pred(&abilities[i]) {
			return &abilities[i]
		}
	}
	return nil
}

// selectHealerAbility selects an appropriate ability for healer role.
func selectHealerAbility(abilities []Ability, allyLowHealth bool) *Ability {
	if allyLowHealth {
		return findAbilityByPredicate(abilities, func(a *Ability) bool { return a.IsHealing })
	}
	return nil
}

// selectDPSAbility selects an appropriate ability for DPS role.
func selectDPSAbility(abilities []Ability, targetLowHealth bool) *Ability {
	if targetLowHealth {
		return findAbilityByPredicate(abilities, func(a *Ability) bool { return a.ID == "execute" })
	}
	return nil
}

// selectTankAbility selects an appropriate ability for tank role.
func selectTankAbility(abilities []Ability) *Ability {
	return findAbilityByPredicate(abilities, func(a *Ability) bool { return a.ID == "shield_wall" })
}

// SelectAbility chooses an appropriate ability based on combat situation.
// Per AC: Companion uses class-appropriate abilities.
func (cm *CompanionManager) SelectAbility(compID uint64, targetLowHealth bool, allyLowHealth bool) *Ability {
	cm.mu.RLock()
	comp := cm.companions[compID]
	cm.mu.RUnlock()

	if comp == nil || len(comp.Abilities) == 0 {
		return nil
	}

	// Role-based ability selection
	var selected *Ability
	switch comp.Role {
	case RoleHealer:
		selected = selectHealerAbility(comp.Abilities, allyLowHealth)
	case RoleDPS:
		selected = selectDPSAbility(comp.Abilities, targetLowHealth)
	case RoleTank:
		selected = selectTankAbility(comp.Abilities)
	}

	if selected != nil {
		return selected
	}

	// Default: use first ability
	return &comp.Abilities[0]
}

// GenerateDialogResponse creates a companion dialog that references past events.
func (cm *CompanionManager) GenerateDialogResponse(compID uint64, topic string) string {
	cm.mu.RLock()
	comp := cm.companions[compID]
	cm.mu.RUnlock()

	if comp == nil {
		return "..."
	}

	// Per AC: dialog references player actions from last 10 events
	response := fmt.Sprintf("[%s - %s %s] ", comp.Name, comp.Personality, comp.Role)

	// Reference recent events
	if len(comp.PlayerEvents) > 0 {
		lastEvent := comp.PlayerEvents[len(comp.PlayerEvents)-1]
		response += fmt.Sprintf("I remember when you %s. ", lastEvent.Description)
	}

	// Add topic-specific response
	response += fmt.Sprintf("About %s... ", topic)

	// Add personality-based flavor
	switch comp.Personality {
	case PersonalityBrave:
		response += "We should face this head-on!"
	case PersonalityCautious:
		response += "Let's be careful about this..."
	case PersonalityLoyal:
		response += "I'll support whatever you decide."
	case PersonalityAggressive:
		response += "Let's crush our enemies!"
	case PersonalityWise:
		response += "Consider all options before acting."
	}

	return response
}

// SetOrder gives a command to the companion.
func (cm *CompanionManager) SetOrder(compID uint64, order Order) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	comp := cm.companions[compID]
	if comp == nil {
		return
	}

	comp.CurrentOrder = order
	comp.Following = (order == OrderFollow)
}

// SetCombatState updates the companion's combat status.
func (cm *CompanionManager) SetCombatState(compID uint64, inCombat bool) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	comp := cm.companions[compID]
	if comp == nil {
		return
	}
	comp.InCombat = inCombat
}

// CompanionCount returns the total number of companions.
func (cm *CompanionManager) CompanionCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.companions)
}
