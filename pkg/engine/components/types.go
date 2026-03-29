// Package components defines all ECS component data types.
package components

// Position represents a 3D location and orientation in the world.
type Position struct {
	X, Y, Z float64
	Angle   float64 // Heading angle in radians (0 = East, PI/2 = North)
}

// Type returns the component type identifier for Position.
func (p *Position) Type() string { return "Position" }

// Health represents an entity's health state.
type Health struct {
	Current, Max float64
}

// Type returns the component type identifier for Health.
func (h *Health) Type() string { return "Health" }

// Faction represents an entity's faction allegiance and reputation.
type Faction struct {
	ID         string
	Reputation float64
}

// Type returns the component type identifier for Faction.
func (f *Faction) Type() string { return "Faction" }

// FactionTerritory represents a faction's claimed territory.
type FactionTerritory struct {
	FactionID string
	// Vertices defines the polygon boundary as X,Y coordinate pairs.
	Vertices []Point2D
	// KillTracker tracks kills by player entity ID.
	KillTracker map[uint64]int
}

// Type returns the component type identifier for FactionTerritory.
func (f *FactionTerritory) Type() string { return "FactionTerritory" }

// Point2D represents a 2D point for territory polygons.
type Point2D struct {
	X, Y float64
}

// ContainsPoint checks if a point is inside the territory polygon using ray casting.
func (f *FactionTerritory) ContainsPoint(x, y float64) bool {
	if len(f.Vertices) < 3 {
		return false
	}
	inside := false
	n := len(f.Vertices)
	j := n - 1
	for i := 0; i < n; i++ {
		xi, yi := f.Vertices[i].X, f.Vertices[i].Y
		xj, yj := f.Vertices[j].X, f.Vertices[j].Y
		if ((yi > y) != (yj > y)) && (x < (xj-xi)*(y-yi)/(yj-yi)+xi) {
			inside = !inside
		}
		j = i
	}
	return inside
}

// Schedule represents an NPC's daily activity schedule.
type Schedule struct {
	CurrentActivity string
	TimeSlots       map[int]string // hour -> activity
}

// Type returns the component type identifier for Schedule.
func (s *Schedule) Type() string { return "Schedule" }

// Inventory represents an entity's carried items.
type Inventory struct {
	Items    []string
	Capacity int
}

// Type returns the component type identifier for Inventory.
func (i *Inventory) Type() string { return "Inventory" }

// Vehicle represents a vehicle component.
type Vehicle struct {
	VehicleType string
	Speed       float64
	Fuel        float64
	Direction   float64 // Heading angle in radians (0 = East, PI/2 = North)
}

// Type returns the component type identifier for Vehicle.
func (v *Vehicle) Type() string { return "Vehicle" }

// VehicleArchetype defines a vehicle template with genre-specific properties.
type VehicleArchetype struct {
	ID          string
	Name        string
	BaseSpeed   float64
	MaxFuel     float64
	FuelRate    float64 // Fuel consumption per unit distance
	Description string
}

// GenreVehicleArchetypes maps genre to available vehicle archetypes.
var GenreVehicleArchetypes = map[string][]VehicleArchetype{
	"fantasy": {
		{ID: "horse", Name: "Horse", BaseSpeed: 15, MaxFuel: 200, FuelRate: 0.005, Description: "A trusty steed"},
		{ID: "cart", Name: "Horse Cart", BaseSpeed: 8, MaxFuel: 300, FuelRate: 0.008, Description: "Slow but carries cargo"},
		{ID: "ship", Name: "Sailing Ship", BaseSpeed: 12, MaxFuel: 500, FuelRate: 0.003, Description: "Ocean vessel for long voyages"},
	},
	"sci-fi": {
		{ID: "hoverbike", Name: "Hover-Bike", BaseSpeed: 30, MaxFuel: 150, FuelRate: 0.02, Description: "Fast anti-gravity cycle"},
		{ID: "shuttle", Name: "Shuttle", BaseSpeed: 50, MaxFuel: 400, FuelRate: 0.03, Description: "Short-range spacecraft"},
		{ID: "mech", Name: "Mech Walker", BaseSpeed: 12, MaxFuel: 300, FuelRate: 0.025, Description: "Armored bipedal walker"},
	},
	"horror": {
		{ID: "hearse", Name: "Hearse", BaseSpeed: 10, MaxFuel: 200, FuelRate: 0.01, Description: "Grim but reliable transport"},
		{ID: "bonecart", Name: "Bone Cart", BaseSpeed: 6, MaxFuel: 150, FuelRate: 0.007, Description: "Skeletal horse pulls a rattling cart"},
		{ID: "raft", Name: "Swamp Raft", BaseSpeed: 5, MaxFuel: 100, FuelRate: 0.002, Description: "Crude watercraft"},
	},
	"cyberpunk": {
		{ID: "motorbike", Name: "Street Bike", BaseSpeed: 25, MaxFuel: 120, FuelRate: 0.015, Description: "Neon-lit speed machine"},
		{ID: "apc", Name: "APC", BaseSpeed: 15, MaxFuel: 350, FuelRate: 0.025, Description: "Armored personnel carrier"},
		{ID: "drone", Name: "Personal Drone", BaseSpeed: 35, MaxFuel: 80, FuelRate: 0.04, Description: "Single-person aerial drone"},
	},
	"post-apocalyptic": {
		{ID: "buggy", Name: "Wasteland Buggy", BaseSpeed: 20, MaxFuel: 100, FuelRate: 0.02, Description: "Cobbled-together desert racer"},
		{ID: "truck", Name: "Armored Truck", BaseSpeed: 12, MaxFuel: 250, FuelRate: 0.03, Description: "Reinforced cargo hauler"},
		{ID: "gyroplane", Name: "Gyroplane", BaseSpeed: 28, MaxFuel: 90, FuelRate: 0.035, Description: "Jury-rigged autogyro"},
	},
}

// GetVehicleArchetypes returns available vehicle archetypes for a genre.
func GetVehicleArchetypes(genre string) []VehicleArchetype {
	archetypes, ok := GenreVehicleArchetypes[genre]
	if !ok {
		return GenreVehicleArchetypes["fantasy"]
	}
	return archetypes
}

// NewVehicleFromArchetype creates a Vehicle component from an archetype.
func NewVehicleFromArchetype(archetype VehicleArchetype) *Vehicle {
	return &Vehicle{
		VehicleType: archetype.ID,
		Speed:       archetype.BaseSpeed,
		Fuel:        archetype.MaxFuel,
		Direction:   0,
	}
}

// Reputation represents an entity's standing with various factions.
type Reputation struct {
	// Standings maps faction ID to reputation value (-100 to 100).
	Standings map[string]float64
}

// Type returns the component type identifier for Reputation.
func (r *Reputation) Type() string { return "Reputation" }

// Crime represents an entity's criminal status.
type Crime struct {
	WantedLevel     int     // 0-5 stars
	BountyAmount    float64 // currency owed
	LastCrimeTime   float64 // game time of last offense
	InJail          bool    // whether entity is currently in jail
	JailReleaseTime float64 // game time when entity is released from jail
}

// Type returns the component type identifier for Crime.
func (c *Crime) Type() string { return "Crime" }

// Witness is a tag component marking NPCs that can report crimes.
type Witness struct {
	CanReport bool
}

// Type returns the component type identifier for Witness.
func (w *Witness) Type() string { return "Witness" }

// EconomyNode represents a location with supply/demand pricing.
type EconomyNode struct {
	// PriceTable maps item type to current price.
	PriceTable map[string]float64
	// Supply maps item type to available quantity.
	Supply map[string]int
	// Demand maps item type to desired quantity.
	Demand map[string]int
}

// Type returns the component type identifier for EconomyNode.
func (e *EconomyNode) Type() string { return "EconomyNode" }

// Quest represents an active quest with branching state.
type Quest struct {
	ID           string
	CurrentStage int
	Flags        map[string]bool
	Completed    bool
	// LockedBranches contains branch IDs that can no longer be taken.
	LockedBranches map[string]bool
}

// Type returns the component type identifier for Quest.
func (q *Quest) Type() string { return "Quest" }

// LockBranch marks a quest branch as unavailable.
func (q *Quest) LockBranch(branchID string) {
	if q.LockedBranches == nil {
		q.LockedBranches = make(map[string]bool)
	}
	q.LockedBranches[branchID] = true
}

// IsBranchLocked checks if a quest branch is locked.
func (q *Quest) IsBranchLocked(branchID string) bool {
	if q.LockedBranches == nil {
		return false
	}
	return q.LockedBranches[branchID]
}

// WorldClock represents the global game time state.
type WorldClock struct {
	Hour       int     // 0-23
	Day        int     // Day count from world start
	TimeAccum  float64 // Accumulated time toward next hour
	HourLength float64 // Real seconds per game hour
}

// Type returns the component type identifier for WorldClock.
func (wc *WorldClock) Type() string { return "WorldClock" }

// Skills represents an entity's skill levels and experience.
// Skills improve through use (Elder Scrolls-style progression).
type Skills struct {
	// Levels maps skill ID to current level (0-100).
	Levels map[string]int
	// Experience maps skill ID to XP toward next level.
	Experience map[string]float64
	// SchoolBonuses maps school name to bonus percentage.
	SchoolBonuses map[string]float64
}

// Type returns the component type identifier for Skills.
func (s *Skills) Type() string { return "Skills" }

// SkillSchool defines a category of related skills.
type SkillSchool struct {
	ID          string
	Name        string   // Genre-specific display name
	Description string   // Genre-specific description
	Skills      []string // Skill IDs in this school
}

// GenreSkillSchools maps genre to school definitions with genre-appropriate names.
var GenreSkillSchools = map[string][]SkillSchool{
	"fantasy": {
		{ID: "combat", Name: "Warrior Arts", Description: "Martial combat and weaponry", Skills: []string{"one_handed", "two_handed", "block", "archery", "heavy_armor"}},
		{ID: "magic", Name: "Destruction", Description: "Offensive magical arts", Skills: []string{"fire_magic", "frost_magic", "shock_magic", "conjuration", "enchanting"}},
		{ID: "stealth", Name: "Shadow Arts", Description: "Subterfuge and cunning", Skills: []string{"sneak", "lockpicking", "pickpocket", "speech", "alchemy"}},
		{ID: "crafting", Name: "Artisan Crafts", Description: "Creation and smithing", Skills: []string{"smithing", "leatherworking", "woodworking", "cooking", "herbalism"}},
		{ID: "knowledge", Name: "Alteration", Description: "Protective and utility magic", Skills: []string{"restoration", "illusion", "divination", "lore", "inscription"}},
		{ID: "physical", Name: "Athletics", Description: "Physical prowess", Skills: []string{"running", "swimming", "climbing", "acrobatics", "endurance"}},
	},
	"sci-fi": {
		{ID: "combat", Name: "Weaponry", Description: "Ranged and energy weapons", Skills: []string{"rifles", "pistols", "heavy_weapons", "tactical_armor", "grenades"}},
		{ID: "magic", Name: "Psi-Ops", Description: "Psionic abilities", Skills: []string{"telekinesis", "mind_control", "precognition", "psi_shield", "mind_link"}},
		{ID: "stealth", Name: "Infiltration", Description: "Covert operations", Skills: []string{"stealth_tech", "hacking", "social_engineering", "disguise", "demolitions"}},
		{ID: "crafting", Name: "Engineering", Description: "Tech construction", Skills: []string{"weapon_mods", "armor_tech", "cybernetics", "drones", "medicine"}},
		{ID: "knowledge", Name: "Biotech", Description: "Biological sciences", Skills: []string{"first_aid", "stims", "genetics", "xenobiology", "research"}},
		{ID: "physical", Name: "Combat Training", Description: "Physical conditioning", Skills: []string{"zero_g", "sprinting", "climbing", "martial_arts", "stamina"}},
	},
	"horror": {
		{ID: "combat", Name: "Survival Combat", Description: "Desperate fighting", Skills: []string{"melee", "firearms", "improvised_weapons", "evasion", "fortification"}},
		{ID: "magic", Name: "Occult", Description: "Dark rituals", Skills: []string{"blood_magic", "summoning", "curses", "warding", "spirit_binding"}},
		{ID: "stealth", Name: "Survival", Description: "Staying hidden", Skills: []string{"hiding", "tracking", "traps", "scavenging", "alertness"}},
		{ID: "crafting", Name: "Improvisation", Description: "Makeshift creation", Skills: []string{"barricading", "medicine_crafting", "trap_making", "repair", "preservation"}},
		{ID: "knowledge", Name: "Forbidden Lore", Description: "Eldritch knowledge", Skills: []string{"occult_lore", "monster_knowledge", "ritual_casting", "sanity", "investigation"}},
		{ID: "physical", Name: "Endurance", Description: "Physical survival", Skills: []string{"running", "holding_breath", "pain_tolerance", "night_vision", "constitution"}},
	},
	"cyberpunk": {
		{ID: "combat", Name: "Street Combat", Description: "Urban warfare", Skills: []string{"firearms", "blades", "martial_arts", "street_armor", "explosives"}},
		{ID: "magic", Name: "Netrunning", Description: "Matrix skills", Skills: []string{"hacking", "ice_breaking", "daemon_control", "data_mining", "system_crash"}},
		{ID: "stealth", Name: "Wetwork", Description: "Assassination and infiltration", Skills: []string{"stealth", "disguise", "social_hacking", "surveillance", "escape"}},
		{ID: "crafting", Name: "Tech", Description: "Cybernetic engineering", Skills: []string{"cyberware", "weapons_tech", "vehicle_mod", "electronics", "bioware"}},
		{ID: "knowledge", Name: "Street Smarts", Description: "Urban survival", Skills: []string{"contacts", "negotiation", "intimidation", "streetwise", "corporate_knowledge"}},
		{ID: "physical", Name: "Chrome", Description: "Enhanced physique", Skills: []string{"reflex", "strength_aug", "endurance_aug", "speed", "combat_sense"}},
	},
	"post-apocalyptic": {
		{ID: "combat", Name: "Wasteland Combat", Description: "Survival fighting", Skills: []string{"guns", "melee", "thrown_weapons", "scrap_armor", "mounted_combat"}},
		{ID: "magic", Name: "Mutations", Description: "Radiation abilities", Skills: []string{"rad_resistance", "mutation_control", "toxic_immunity", "regeneration", "sixth_sense"}},
		{ID: "stealth", Name: "Scavenging", Description: "Finding and hiding", Skills: []string{"sneak", "lockpicking", "scavenging", "tracking", "camouflage"}},
		{ID: "crafting", Name: "Jury-Rigging", Description: "Makeshift creation", Skills: []string{"weapon_crafting", "armor_repair", "vehicle_repair", "medicine", "cooking"}},
		{ID: "knowledge", Name: "Wasteland Lore", Description: "Survival knowledge", Skills: []string{"navigation", "weather_sense", "creature_lore", "trade_routes", "old_world_tech"}},
		{ID: "physical", Name: "Hardened", Description: "Wasteland toughness", Skills: []string{"endurance", "radiation_tolerance", "disease_resistance", "sprinting", "brawling"}},
	},
}

// GetSkillSchools returns the skill schools for a given genre.
func GetSkillSchools(genre string) []SkillSchool {
	schools, ok := GenreSkillSchools[genre]
	if !ok {
		return GenreSkillSchools["fantasy"]
	}
	return schools
}

// GetAllSkillIDs returns all skill IDs for a given genre.
func GetAllSkillIDs(genre string) []string {
	schools := GetSkillSchools(genre)
	var skills []string
	for _, school := range schools {
		skills = append(skills, school.Skills...)
	}
	return skills
}

// NewSkills creates a new Skills component with all skills at level 0.
func NewSkills(genre string) *Skills {
	skills := &Skills{
		Levels:        make(map[string]int),
		Experience:    make(map[string]float64),
		SchoolBonuses: make(map[string]float64),
	}
	for _, skillID := range GetAllSkillIDs(genre) {
		skills.Levels[skillID] = 1
		skills.Experience[skillID] = 0
	}
	return skills
}

// AudioListener marks an entity as the audio listener for 3D spatial audio.
// Typically attached to the player entity.
type AudioListener struct {
	// Volume is the master volume multiplier (0.0 to 1.0).
	Volume float64
	// Enabled controls whether the listener is active.
	Enabled bool
}

// Type returns the component type identifier for AudioListener.
func (a *AudioListener) Type() string { return "AudioListener" }

// AudioSource represents a sound-emitting entity in the world.
type AudioSource struct {
	// SoundType identifies the category of sound (footstep, ambient, etc.).
	SoundType string
	// Volume is the source volume multiplier (0.0 to 1.0).
	Volume float64
	// Range is the maximum audible distance in world units.
	Range float64
	// Looping indicates whether the sound should repeat.
	Looping bool
	// Playing indicates whether the sound is currently active.
	Playing bool
}

// Type returns the component type identifier for AudioSource.
func (a *AudioSource) Type() string { return "AudioSource" }

// AudioState holds runtime audio playback state for the AudioSystem.
type AudioState struct {
	// CurrentAmbient is the currently playing ambient sound type.
	CurrentAmbient string
	// CombatIntensity is the current combat music intensity (0.0 to 1.0).
	CombatIntensity float64
	// LastPositionX caches the listener's last X position.
	LastPositionX float64
	// LastPositionY caches the listener's last Y position.
	LastPositionY float64
}

// Type returns the component type identifier for AudioState.
func (a *AudioState) Type() string { return "AudioState" }
