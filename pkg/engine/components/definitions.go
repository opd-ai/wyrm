// Package components defines all ECS component data types.
package components

import "github.com/opd-ai/wyrm/pkg/geom"

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

// FactionMembership tracks a player's membership, rank, and standing with multiple factions.
type FactionMembership struct {
	// Memberships maps faction ID to membership details.
	Memberships map[string]*FactionMemberInfo
}

// FactionMemberInfo represents a player's membership in a single faction.
// NOTE: This is a helper struct used within FactionMembership, not an ECS component.
type FactionMemberInfo struct {
	FactionID       string
	Rank            int             // 0 = not a member, 1-10 = member ranks
	RankTitle       string          // Rank-specific title (e.g., "Initiate", "Champion")
	Reputation      float64         // -100 to 100 standing with faction
	JoinedAt        float64         // Game time when joined
	LastPromoAt     float64         // Game time of last promotion
	XP              int             // Experience points toward next rank
	XPToNext        int             // XP required for next rank
	QuestsCompleted int             // Count of faction quests completed
	DonationTotal   int             // Total gold/resources donated
	IsExalted       bool            // Maximum rank achieved
	UnlockedContent map[string]bool // Exclusive content IDs that have been unlocked
}

// Type returns the component type identifier for FactionMembership.
func (f *FactionMembership) Type() string { return "FactionMembership" }

// GetMembership returns membership info for a faction, or nil if not a member.
func (f *FactionMembership) GetMembership(factionID string) *FactionMemberInfo {
	if f.Memberships == nil {
		return nil
	}
	return f.Memberships[factionID]
}

// GetRank returns the player's rank in a faction (0 if not a member).
func (f *FactionMembership) GetRank(factionID string) int {
	if info := f.GetMembership(factionID); info != nil {
		return info.Rank
	}
	return 0
}

// IsMember returns true if the player is a member of the faction.
func (f *FactionMembership) IsMember(factionID string) bool {
	return f.GetRank(factionID) > 0
}

// FactionTerritory represents a faction's claimed territory.
type FactionTerritory struct {
	FactionID string
	// Vertices defines the polygon boundary as X,Y coordinate pairs.
	Vertices []Point2D
	// KillTracker tracks kills by player entity ID.
	KillTracker map[uint64]int
	// ControlLevel represents how much control the faction has (0-1).
	ControlLevel float64
}

// Type returns the component type identifier for FactionTerritory.
func (f *FactionTerritory) Type() string { return "FactionTerritory" }

// Point2D represents a 2D point for territory polygons.
// NOTE: This is a helper struct used within components, not an ECS component.
type Point2D struct {
	X, Y float64
}

// ContainsPoint checks if a point is inside the territory polygon using ray casting.
// Delegates to geom.PointInPolygon for the actual algorithm.
func (f *FactionTerritory) ContainsPoint(x, y float64) bool {
	if len(f.Vertices) < 3 {
		return false
	}
	// Convert []Point2D to flat []float64 slice for geom.PointInPolygon
	flat := make([]float64, len(f.Vertices)*2)
	for i, v := range f.Vertices {
		flat[i*2] = v.X
		flat[i*2+1] = v.Y
	}
	return geom.PointInPolygon(x, y, flat)
}

// Schedule represents an NPC's daily activity schedule.
type Schedule struct {
	CurrentActivity string
	TimeSlots       map[int]string // hour -> activity
}

// Type returns the component type identifier for Schedule.
func (s *Schedule) Type() string { return "Schedule" }

// NPCPathfinding tracks NPC movement toward scheduled activity locations.
type NPCPathfinding struct {
	// TargetX, TargetY is the destination position.
	TargetX, TargetY float64
	// HasTarget indicates if the NPC has a destination.
	HasTarget bool
	// IsMoving indicates if the NPC is currently traveling.
	IsMoving bool
	// MoveSpeed is the NPC's movement speed (units per second).
	MoveSpeed float64
	// ActivityLocations maps activity names to world positions.
	ActivityLocations map[string]ActivityLocation
	// CurrentPath is the list of waypoints to follow.
	CurrentPath []Waypoint
	// CurrentWaypointIndex is the next waypoint to reach.
	CurrentWaypointIndex int
	// ArrivalThreshold is the distance at which the NPC is considered arrived.
	ArrivalThreshold float64
	// StuckTime tracks time spent not making progress.
	StuckTime float64
	// MaxStuckTime is the time before giving up on current path.
	MaxStuckTime float64
}

// Type returns the component type identifier for NPCPathfinding.
func (n *NPCPathfinding) Type() string { return "NPCPathfinding" }

// ActivityLocation represents a position for a scheduled activity.
// NOTE: This is a helper struct used within NPCPathfinding, not an ECS component.
type ActivityLocation struct {
	X, Y       float64
	LocationID string // Building or POI identifier
}

// Waypoint represents a point along a path.
// NOTE: This is a helper struct used within NPCPathfinding, not an ECS component.
type Waypoint struct {
	X, Y float64
}

// Inventory represents an entity's carried items.
type Inventory struct {
	Items    []string
	Capacity int
}

// Type returns the component type identifier for Inventory.
func (i *Inventory) Type() string { return "Inventory" }

// Currency represents an entity's monetary resources.
type Currency struct {
	// Gold is the primary currency amount.
	Gold int
	// Silver is a secondary currency (100 = 1 gold).
	Silver int
	// Copper is a tertiary currency (100 = 1 silver).
	Copper int
}

// Type returns the component type identifier for Currency.
func (c *Currency) Type() string { return "Currency" }

// Vehicle represents a vehicle component.
type Vehicle struct {
	VehicleType string
	Speed       float64
	Fuel        float64
	Direction   float64 // Heading angle in radians (0 = East, PI/2 = North)
}

// Type returns the component type identifier for Vehicle.
func (v *Vehicle) Type() string { return "Vehicle" }

// VehiclePhysics adds detailed physics simulation to a vehicle.
type VehiclePhysics struct {
	// CurrentSpeed is the current forward speed in units/second.
	CurrentSpeed float64
	// MaxSpeed is the maximum speed for this vehicle.
	MaxSpeed float64
	// Acceleration is the rate of speed increase per second.
	Acceleration float64
	// Deceleration is the rate of speed decrease per second (braking).
	Deceleration float64
	// FrictionDecel is the passive speed loss per second (no input).
	FrictionDecel float64
	// SteeringAngle is the current wheel/rudder angle in radians.
	SteeringAngle float64
	// MaxSteeringAngle is the maximum steering angle in radians.
	MaxSteeringAngle float64
	// SteeringSpeed is how fast steering changes (radians/second).
	SteeringSpeed float64
	// TurningRadius is the minimum turning circle radius.
	TurningRadius float64
	// Mass affects acceleration and handling.
	Mass float64
	// Throttle is current acceleration input (-1 to 1, negative = reverse).
	Throttle float64
	// Steering is current steering input (-1 = left, 1 = right).
	Steering float64
	// IsBraking indicates if brakes are applied.
	IsBraking bool
	// InReverse indicates if moving backward.
	InReverse bool
}

// Type returns the component type identifier for VehiclePhysics.
func (vp *VehiclePhysics) Type() string { return "VehiclePhysics" }

// VehicleState tracks operational state of a vehicle.
type VehicleState struct {
	// IsOccupied indicates if a driver is in the vehicle.
	IsOccupied bool
	// DriverEntity is the entity ID of the driver (0 = no driver).
	DriverEntity uint64
	// PassengerEntities lists passenger entity IDs.
	PassengerEntities []uint64
	// MaxPassengers is the maximum number of passengers.
	MaxPassengers int
	// InCockpitView indicates if player sees cockpit view.
	InCockpitView bool
	// EngineRunning indicates if the engine is on.
	EngineRunning bool
	// DamagePercent is vehicle damage (0 = pristine, 100 = destroyed).
	DamagePercent float64
}

// Type returns the component type identifier for VehicleState.
func (vs *VehicleState) Type() string { return "VehicleState" }

// MountInfo stores data about a mountable creature entity.
type MountInfo struct {
	// MountType is the type of mount (e.g., "horse", "wolf", "dragon").
	MountType string
	// Name is the custom name given to this mount.
	Name string
	// OwnerEntity is the entity ID of the owner (0 = wild/no owner).
	OwnerEntity uint64
	// RiderEntity is the current rider (0 = not mounted).
	RiderEntity uint64
	// IsMounted indicates if someone is currently riding.
	IsMounted bool
	// Speed is the movement speed when mounted.
	Speed float64
	// Stamina is current stamina.
	Stamina float64
	// MaxStamina is maximum stamina.
	MaxStamina float64
}

// Type returns the component type identifier for MountInfo.
func (mi *MountInfo) Type() string { return "MountInfo" }

// VehicleArchetype defines a vehicle template with genre-specific properties.
// NOTE: This is a helper struct used for vehicle configuration, not an ECS component.
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

// Guard represents a guard NPC with pursuit AI state.
type Guard struct {
	State        int     // Current AI state (patrol, alert, pursue, etc.)
	TargetEntity uint64  // Entity being pursued
	LastKnownX   float64 // Last known X position of target
	LastKnownZ   float64 // Last known Z position of target
	SearchTimer  float64 // Time remaining in search mode
	AlertTimer   float64 // Time remaining in alert mode
	PatrolTimer  float64 // Time spent patrolling
	PursuitSpeed float64 // Movement speed during pursuit
	PatrolSpeed  float64 // Movement speed during patrol
	SightRange   float64 // How far the guard can see
	HearingRange float64 // How far the guard can hear
}

// Type returns the component type identifier for Guard.
func (g *Guard) Type() string { return "Guard" }

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

// Weather represents the current weather state for an area.
// This is a singleton-style component typically attached to a world entity.
type Weather struct {
	WeatherType string  // Weather type: "clear", "overcast", "rain", "storm", "snow", "fog"
	CloudCover  float64 // Cloud coverage amount (0.0-1.0)
	Intensity   float64 // Weather intensity (0.0-1.0)
	WindSpeed   float64 // Wind speed in units per second
	WindAngle   float64 // Wind direction in radians
}

// Type returns the component type identifier for Weather.
func (w *Weather) Type() string { return "Weather" }

// NewWeather creates a new Weather component with default clear sky settings.
func NewWeather() *Weather {
	return &Weather{
		WeatherType: "clear",
		CloudCover:  0.0,
		Intensity:   0.0,
		WindSpeed:   0.0,
		WindAngle:   0.0,
	}
}

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
// NOTE: This is a helper struct used for skill configuration, not an ECS component.
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
	// EffectiveVolume is the computed volume after spatial attenuation.
	EffectiveVolume float64
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

// Weapon represents an entity's equipped weapon.
type Weapon struct {
	// Name is the weapon's display name.
	Name string
	// Damage is the base damage dealt per hit.
	Damage float64
	// Range is the maximum attack range in world units.
	Range float64
	// AttackSpeed is attacks per second.
	AttackSpeed float64
	// WeaponType categorizes the weapon (melee, ranged, magic).
	WeaponType string
}

// Type returns the component type identifier for Weapon.
func (w *Weapon) Type() string { return "Weapon" }

// CombatState tracks combat-related runtime state.
type CombatState struct {
	// LastAttackTime is the game time of the last attack.
	LastAttackTime float64
	// Cooldown is the remaining time before the next attack.
	Cooldown float64
	// IsAttacking indicates an attack is in progress.
	IsAttacking bool
	// TargetEntity is the current attack target (0 = none).
	TargetEntity uint64
	// InCombat indicates the entity is engaged in combat.
	InCombat bool
	// IsBlocking indicates the entity is blocking incoming attacks.
	IsBlocking bool
	// BlockReduction is the percentage of damage blocked (0.0 to 1.0).
	BlockReduction float64
	// IsDodging indicates the entity is performing a dodge roll.
	IsDodging bool
	// DodgeEndTime is the game time when the dodge ends.
	DodgeEndTime float64
	// DodgeInvulnerable indicates dodge grants invulnerability frames.
	DodgeInvulnerable bool
}

// Type returns the component type identifier for CombatState.
func (c *CombatState) Type() string { return "CombatState" }

// Stealth represents an entity's stealth state for sneaking mechanics.
type Stealth struct {
	// Visibility is the current visibility level (0.0 = invisible, 1.0 = fully visible).
	Visibility float64
	// Sneaking indicates if the entity is actively sneaking.
	Sneaking bool
	// DetectionRadius is how far NPCs can detect this entity when sneaking.
	DetectionRadius float64
	// BaseVisibility is the default visibility when not sneaking.
	BaseVisibility float64
	// SneakVisibility is the visibility when actively sneaking.
	SneakVisibility float64
	// LastDetectedBy tracks which entities have detected this one (entity ID -> time).
	LastDetectedBy map[uint64]float64
}

// Type returns the component type identifier for Stealth.
func (s *Stealth) Type() string { return "Stealth" }

// Awareness tracks an NPC's awareness of stealthy players.
type Awareness struct {
	// AlertLevel is the current alert state (0.0 = unaware, 1.0 = fully alert).
	AlertLevel float64
	// SightRange is the maximum distance the NPC can see.
	SightRange float64
	// SightAngle is the field of view angle in radians.
	SightAngle float64
	// DetectedEntities tracks entities this NPC is aware of (entity ID -> alert level).
	DetectedEntities map[uint64]float64
}

// Type returns the component type identifier for Awareness.
func (a *Awareness) Type() string { return "Awareness" }

// Material represents a gatherable or craftable material.
type Material struct {
	// ResourceType identifies the material category (ore, herb, wood, etc.).
	ResourceType string
	// Quantity is the amount of this material.
	Quantity int
	// Quality is the material quality (0.0-1.0, affects crafted item quality).
	Quality float64
	// Rarity indicates how rare this material is (common, uncommon, rare, epic, legendary).
	Rarity string
}

// Type returns the component type identifier for Material.
func (m *Material) Type() string { return "Material" }

// ResourceNode represents a gatherable resource in the world.
type ResourceNode struct {
	// ResourceType identifies what material this node yields.
	ResourceType string
	// Quantity is the remaining amount available.
	Quantity int
	// MaxQuantity is the maximum this node can hold.
	MaxQuantity int
	// Quality is the base quality of materials from this node.
	Quality float64
	// RespawnTime is seconds until the node respawns after depletion.
	RespawnTime float64
	// LastGathered is the game time when the node was last gathered.
	LastGathered float64
	// Depleted indicates the node is currently empty.
	Depleted bool
}

// Type returns the component type identifier for ResourceNode.
func (r *ResourceNode) Type() string { return "ResourceNode" }

// Workbench represents a crafting station.
type Workbench struct {
	// WorkbenchType identifies the station type (forge, alchemy_table, enchanting_table, etc.).
	WorkbenchType string
	// SupportedRecipeTypes lists what recipe categories this workbench can craft.
	SupportedRecipeTypes []string
	// CraftingSpeedMult is a multiplier on crafting time (1.0 = normal, 0.5 = twice as fast).
	CraftingSpeedMult float64
	// QualityBonus is added to crafted item quality.
	QualityBonus float64
}

// Type returns the component type identifier for Workbench.
func (w *Workbench) Type() string { return "Workbench" }

// CraftingState tracks an entity's ongoing crafting activity.
type CraftingState struct {
	// IsCrafting indicates if the entity is currently crafting.
	IsCrafting bool
	// CurrentRecipeID is the ID of the recipe being crafted.
	CurrentRecipeID string
	// Progress is 0.0-1.0 representing completion percentage.
	Progress float64
	// TotalTime is the total crafting time in seconds.
	TotalTime float64
	// WorkbenchEntity is the entity ID of the workbench being used.
	WorkbenchEntity uint64
	// ConsumedMaterials tracks materials already consumed for this craft.
	ConsumedMaterials map[string]int
}

// Type returns the component type identifier for CraftingState.
func (c *CraftingState) Type() string { return "CraftingState" }

// Tool represents an equipped tool with durability.
type Tool struct {
	// ToolType identifies the tool category (pickaxe, axe, hammer, etc.).
	ToolType string
	// Name is the tool's display name.
	Name string
	// Durability is the current durability (0 = broken).
	Durability float64
	// MaxDurability is the maximum durability.
	MaxDurability float64
	// GatherSpeed is a multiplier on gathering time.
	GatherSpeed float64
	// QualityBonus is added to gathered material quality.
	QualityBonus float64
	// ToolTier affects what resources can be gathered (1=basic, 5=legendary).
	ToolTier int
}

// Type returns the component type identifier for Tool.
func (t *Tool) Type() string { return "Tool" }

// Equipment holds equipped items in various slots.
type Equipment struct {
	// Slots maps slot names to equipped items.
	Slots map[string]*EquipmentSlot
}

// EquipmentSlot represents a single equipment slot.
// NOTE: This is a helper struct used within Equipment, not an ECS component.
type EquipmentSlot struct {
	ItemID        string
	Name          string
	Durability    float64
	MaxDurability float64
}

// Type returns the component type identifier for Equipment.
func (e *Equipment) Type() string { return "Equipment" }

// RecipeKnowledge tracks which recipes an entity has discovered.
type RecipeKnowledge struct {
	// KnownRecipes is a set of recipe IDs the entity can craft.
	KnownRecipes map[string]bool
	// DiscoveryProgress tracks partial discovery progress for recipes.
	DiscoveryProgress map[string]float64
}

// Type returns the component type identifier for RecipeKnowledge.
func (r *RecipeKnowledge) Type() string { return "RecipeKnowledge" }

// Projectile represents a moving projectile entity (arrow, bullet, spell).
type Projectile struct {
	// OwnerID is the entity that fired this projectile.
	OwnerID uint64
	// VelocityX is the X component of velocity (units per second).
	VelocityX float64
	// VelocityY is the Y component of velocity (units per second).
	VelocityY float64
	// VelocityZ is the Z component of velocity (units per second).
	VelocityZ float64
	// Damage is the damage dealt on impact.
	Damage float64
	// Lifetime is the remaining time before despawn (seconds).
	Lifetime float64
	// HitRadius is the collision radius for hit detection.
	HitRadius float64
	// ProjectileType identifies the projectile type (arrow, bullet, spell, etc.).
	ProjectileType string
	// PierceCount is how many targets the projectile can hit before despawning (0 = unlimited).
	PierceCount int
	// HitEntities tracks which entities have already been hit (for pierce mechanics).
	HitEntities map[uint64]bool
}

// Type returns the component type identifier for Projectile.
func (p *Projectile) Type() string { return "Projectile" }

// Mana represents an entity's magical energy pool.
type Mana struct {
	// Current is the current mana level.
	Current float64
	// Max is the maximum mana capacity.
	Max float64
	// RegenRate is mana regenerated per second.
	RegenRate float64
}

// Type returns the component type identifier for Mana.
func (m *Mana) Type() string { return "Mana" }

// Stamina represents an entity's stamina resource for physical actions.
type Stamina struct {
	// Current is the current stamina level.
	Current float64
	// Max is the maximum stamina capacity.
	Max float64
	// RegenRate is stamina regenerated per second.
	RegenRate float64
}

// Type returns the component type identifier for Stamina.
func (s *Stamina) Type() string { return "Stamina" }

// DeathState tracks an entity's death status and respawn information.
type DeathState struct {
	// IsDead indicates whether the entity is currently dead.
	IsDead bool
	// DeathTime is the game time when death occurred.
	DeathTime float64
	// RespawnTime is the game time when respawn is available.
	RespawnTime float64
	// RespawnPosition is where the entity will respawn.
	RespawnX, RespawnY, RespawnZ float64
	// PenaltiesApplied indicates death penalties have been processed.
	PenaltiesApplied bool
}

// Type returns the component type identifier for DeathState.
func (d *DeathState) Type() string { return "DeathState" }

// Corpse represents a dead entity's remains containing lootable items.
type Corpse struct {
	// OwnerEntity is the entity ID of the original owner.
	OwnerEntity uint64
	// DeathTime is when the entity died.
	DeathTime float64
	// DecayTime is when the corpse will despawn.
	DecayTime float64
	// LootedBy tracks which entities have looted this corpse.
	LootedBy map[uint64]bool
}

// Type returns the component type identifier for Corpse.
func (c *Corpse) Type() string { return "Corpse" }

// SpellEffect represents an active status effect on an entity.
type SpellEffect struct {
	// EffectType identifies the effect (damage, heal, buff, debuff).
	EffectType string
	// Magnitude is the strength of the effect.
	Magnitude float64
	// Duration is the total duration in seconds.
	Duration float64
	// Remaining is the remaining duration in seconds.
	Remaining float64
	// Source is the entity that applied this effect.
	Source uint64
}

// Type returns the component type identifier for SpellEffect.
func (s *SpellEffect) Type() string { return "SpellEffect" }

// Spell represents a castable spell or ability.
type Spell struct {
	// ID is the unique spell identifier.
	ID string
	// Name is the display name.
	Name string
	// ManaCost is the mana required to cast.
	ManaCost float64
	// Cooldown is the time between casts (seconds).
	Cooldown float64
	// LastCast is the game time when last cast.
	LastCast float64
	// EffectType is the spell effect type.
	EffectType string
	// Magnitude is the spell's power.
	Magnitude float64
	// Range is the maximum casting distance.
	Range float64
	// AreaOfEffect is the radius for AoE spells (0 = single target).
	AreaOfEffect float64
	// ProjectileSpeed is the speed if this is a projectile spell (0 = instant).
	ProjectileSpeed float64
}

// Type returns the component type identifier for Spell.
func (sp *Spell) Type() string { return "Spell" }

// Spellbook contains an entity's known spells.
type Spellbook struct {
	// Spells maps spell ID to spell data.
	Spells map[string]*Spell
	// ActiveSpellID is the currently selected spell.
	ActiveSpellID string
}

// Type returns the component type identifier for Spellbook.
func (sb *Spellbook) Type() string { return "Spellbook" }

// MemoryEvent represents a single interaction event in NPC memory.
// NOTE: This is a helper struct used within NPCMemory, not an ECS component.
type MemoryEvent struct {
	// EventType categorizes the interaction ("gift", "attack", "dialog", "quest_complete", "theft").
	EventType string
	// Timestamp is the game time when this event occurred.
	Timestamp float64
	// Impact is the disposition change caused by this event (-1.0 to +1.0).
	Impact float64
	// Details contains additional context about the event.
	Details string
}

// NPCMemory stores an NPC's memories of player interactions.
type NPCMemory struct {
	// PlayerInteractions maps player entity IDs to their interaction history.
	PlayerInteractions map[uint64][]MemoryEvent
	// LastSeen maps player entity IDs to the last game time they were seen.
	LastSeen map[uint64]float64
	// Disposition maps player entity IDs to how the NPC feels about them (-1.0 = hostile, +1.0 = friendly).
	Disposition map[uint64]float64
	// MaxMemories is the maximum number of events to remember per player.
	MaxMemories int
	// MemoryDecayRate is how fast old memories fade (disposition per second).
	MemoryDecayRate float64
}

// Type returns the component type identifier for NPCMemory.
func (m *NPCMemory) Type() string { return "NPCMemory" }

// Relationship tracks a social bond between two entities.
// NOTE: This is a helper struct used within NPCRelationships, not an ECS component.
type Relationship struct {
	// TargetEntity is the entity this relationship is with.
	TargetEntity uint64
	// Type classifies the relationship ("friend", "enemy", "neutral", "family", "employer").
	Type string
	// Strength indicates how strong the bond is (0.0 to 1.0).
	Strength float64
	// History tracks significant events in this relationship.
	History []MemoryEvent
}

// NPCRelationships stores an NPC's relationships with other entities.
type NPCRelationships struct {
	// Relationships maps entity IDs to relationship data.
	Relationships map[uint64]*Relationship
}

// Type returns the component type identifier for NPCRelationships.
func (r *NPCRelationships) Type() string { return "NPCRelationships" }

// SocialStatus represents an NPC's standing in society.
type SocialStatus struct {
	// Wealth indicates economic status (0.0 = destitute, 1.0 = wealthy).
	Wealth float64
	// Influence indicates social/political power (0.0 = none, 1.0 = high).
	Influence float64
	// Occupation is the NPC's job or role.
	Occupation string
	// Title is any honorific or rank.
	Title string
}

// Type returns the component type identifier for SocialStatus.
func (s *SocialStatus) Type() string { return "SocialStatus" }

// Interior represents the inside of a building.
type Interior struct {
	// ParentBuilding is the entity ID of the building containing this interior.
	ParentBuilding uint64
	// Width is the interior width in units.
	Width int
	// Height is the interior height in units.
	Height int
	// Rooms contains room definitions within the interior.
	Rooms []Room
	// Furniture lists furniture entity IDs placed in this interior.
	Furniture []uint64
	// WallTiles defines the wall layout (1 = wall, 0 = empty).
	WallTiles [][]int
	// FloorType determines the floor texture/material.
	FloorType string
}

// Room defines a single room within an interior.
// NOTE: This is a helper struct used within Interior, not an ECS component.
type Room struct {
	// ID uniquely identifies this room within the interior.
	ID string
	// Name is the display name of the room.
	Name string
	// X, Y are the room's top-left coordinates.
	X, Y int
	// Width, Height are the room dimensions.
	Width, Height int
	// Purpose describes the room's function ("shop", "bedroom", "storage", etc.).
	Purpose string
}

// Type returns the component type identifier for Interior.
func (i *Interior) Type() string { return "Interior" }

// POIMarker marks an entity as a Point of Interest on the map.
type POIMarker struct {
	// IconType determines the map icon ("shop", "quest", "danger", "guild", "inn", "blacksmith").
	IconType string
	// Name is the display name for this POI.
	Name string
	// Description provides additional info when hovering/selecting.
	Description string
	// Visible determines if this POI appears on the map.
	Visible bool
	// MinimapVisible determines if this POI appears on the minimap.
	MinimapVisible bool
	// DiscoveryRequired means the POI must be discovered before showing.
	DiscoveryRequired bool
	// Discovered indicates if the player has found this POI.
	Discovered bool
}

// Type returns the component type identifier for POIMarker.
func (p *POIMarker) Type() string { return "POIMarker" }

// Building represents a building structure in the world.
type Building struct {
	// BuildingType classifies the building ("shop", "residence", "government", "industrial", "inn").
	BuildingType string
	// Name is the building's display name.
	Name string
	// OwnerFaction is the faction ID that owns this building.
	OwnerFaction string
	// InteriorEntity links to the interior entity (0 = no interior).
	InteriorEntity uint64
	// Floors is the number of floors/stories.
	Floors int
	// Width, Height are the exterior dimensions.
	Width, Height float64
	// EntranceX, EntranceY, EntranceZ are the door coordinates.
	EntranceX, EntranceY, EntranceZ float64
	// IsOpen indicates if the building is currently accessible.
	IsOpen bool
	// OpenHour, CloseHour define operating hours (0-23).
	OpenHour, CloseHour int
}

// Type returns the component type identifier for Building.
func (b *Building) Type() string { return "Building" }

// ShopInventory represents a shop's available goods.
type ShopInventory struct {
	// ShopType classifies the shop ("general", "blacksmith", "alchemist", "tailor", "weapons").
	ShopType string
	// Items maps item IDs to quantities.
	Items map[string]int
	// Prices maps item IDs to current prices (may differ from base economy).
	Prices map[string]float64
	// RestockInterval is hours between restocking.
	RestockInterval int
	// LastRestock is the game hour when last restocked.
	LastRestock int
	// GoldReserve is the shop's available gold for buying from players.
	GoldReserve float64
}

// Type returns the component type identifier for ShopInventory.
func (s *ShopInventory) Type() string { return "ShopInventory" }

// GovernmentBuilding holds data specific to government/faction buildings.
type GovernmentBuilding struct {
	// GovernmentType classifies the building ("barracks", "courthouse", "guild_hall", "palace", "prison").
	GovernmentType string
	// ControllingFaction is the faction ID in control.
	ControllingFaction string
	// Services lists available services ("bounty_payment", "quest_board", "training", "storage").
	Services []string
	// NPCRoles lists NPC roles stationed here ("guard", "clerk", "leader").
	NPCRoles []string
}

// Type returns the component type identifier for GovernmentBuilding.
func (g *GovernmentBuilding) Type() string { return "GovernmentBuilding" }

// GossipItem represents a piece of gossip that can spread through the NPC network.
// NOTE: This is a helper struct used within GossipNetwork, not an ECS component.
type GossipItem struct {
	// ID uniquely identifies this gossip.
	ID string
	// Topic categorizes the gossip ("crime", "romance", "business", "politics", "danger").
	Topic string
	// Content is the gossip text/description.
	Content string
	// SubjectEntity is the entity the gossip is about (0 = general/no subject).
	SubjectEntity uint64
	// OriginTime is when the gossip originated.
	OriginTime float64
	// Spread indicates how many NPCs have heard this (0.0 to 1.0 of local population).
	Spread float64
	// Truthfulness indicates accuracy (0.0 = lie, 1.0 = true).
	Truthfulness float64
	// ImpactOnReputation is how this gossip affects the subject's reputation.
	ImpactOnReputation float64
}

// GossipNetwork stores gossip an NPC knows and can spread.
type GossipNetwork struct {
	// KnownGossip maps gossip IDs to the gossip items this NPC knows.
	KnownGossip map[string]*GossipItem
	// GossipChance is the probability of sharing gossip during social interactions (0.0-1.0).
	GossipChance float64
	// ListenChance is the probability of remembering heard gossip (0.0-1.0).
	ListenChance float64
	// LastGossipTime tracks when this NPC last gossiped.
	LastGossipTime float64
	// GossipCooldown is the minimum time between gossip sharing.
	GossipCooldown float64
}

// Type returns the component type identifier for GossipNetwork.
func (g *GossipNetwork) Type() string { return "GossipNetwork" }

// EmotionalState represents an NPC's current emotional condition.
type EmotionalState struct {
	// CurrentEmotion is the dominant emotion ("neutral", "happy", "sad", "angry", "fearful", "disgusted", "surprised").
	CurrentEmotion string
	// Intensity indicates emotion strength (0.0 = calm, 1.0 = intense).
	Intensity float64
	// Mood is the longer-term emotional baseline (-1.0 = depressed, +1.0 = elated).
	Mood float64
	// Stress accumulates from negative events (0.0 to 1.0, high = breakdown risk).
	Stress float64
	// LastEmotionChange is when the emotion last changed.
	LastEmotionChange float64
	// EmotionDecayRate is how fast emotions return to neutral.
	EmotionDecayRate float64
	// MoodDecayRate is how fast mood returns to neutral.
	MoodDecayRate float64
}

// Type returns the component type identifier for EmotionalState.
func (e *EmotionalState) Type() string { return "EmotionalState" }

// NPCNeeds tracks an NPC's basic needs that drive behavior.
type NPCNeeds struct {
	// Hunger ranges from 0.0 (full) to 1.0 (starving).
	Hunger float64
	// Energy ranges from 0.0 (exhausted) to 1.0 (fully rested).
	Energy float64
	// Social ranges from 0.0 (lonely) to 1.0 (socially fulfilled).
	Social float64
	// Safety ranges from 0.0 (terrified) to 1.0 (completely safe).
	Safety float64
	// HungerRate is how fast hunger increases per hour.
	HungerRate float64
	// EnergyRate is how fast energy decreases per hour when awake.
	EnergyRate float64
	// SocialDecayRate is how fast social need decreases per hour alone.
	SocialDecayRate float64
}

// Type returns the component type identifier for NPCNeeds.
func (n *NPCNeeds) Type() string { return "NPCNeeds" }

// ========== Environmental Hazard Components ==========

// HazardType represents different types of environmental dangers.
type HazardType string

const (
	HazardTypeRadiation HazardType = "radiation"
	HazardTypeFire      HazardType = "fire"
	HazardTypePoison    HazardType = "poison"
	HazardTypeElectric  HazardType = "electric"
	HazardTypeFreeze    HazardType = "freeze"
	HazardTypeLava      HazardType = "lava"
	HazardTypeAcid      HazardType = "acid"
	HazardTypeMagic     HazardType = "magic"
	HazardTypeTrap      HazardType = "trap"
	HazardTypeGas       HazardType = "gas"
)

// EnvironmentalHazard represents a dangerous area in the world.
type EnvironmentalHazard struct {
	// HazardType identifies the type of hazard.
	HazardType HazardType
	// Intensity is the strength of the hazard effect (0.0-1.0).
	Intensity float64
	// DamagePerSecond is the base damage dealt per second of exposure.
	DamagePerSecond float64
	// Radius is the effect radius in world units (0 = point source).
	Radius float64
	// Active indicates if the hazard is currently dangerous.
	Active bool
	// Visible indicates if the hazard can be seen by players.
	Visible bool
	// Permanent indicates if the hazard persists indefinitely.
	Permanent bool
	// Duration is the remaining time for temporary hazards (seconds).
	Duration float64
	// CooldownTime is the time between damage ticks.
	CooldownTime float64
	// LastDamageTick tracks the last damage application time.
	LastDamageTick float64
	// GenreOverride allows different behavior based on genre.
	GenreOverride string
}

// Type returns the component type identifier for EnvironmentalHazard.
func (e *EnvironmentalHazard) Type() string { return "EnvironmentalHazard" }

// HazardResistance tracks an entity's resistance to various hazards.
type HazardResistance struct {
	// Resistances maps hazard types to resistance values (0.0 = none, 1.0 = immune).
	Resistances map[HazardType]float64
	// ProtectionEquipment maps hazard types to equipped protection items.
	ProtectionEquipment map[HazardType]string
	// ActiveEffects maps hazard types to currently active effect durations.
	ActiveEffects map[HazardType]float64
}

// Type returns the component type identifier for HazardResistance.
func (h *HazardResistance) Type() string { return "HazardResistance" }

// GetResistance returns the resistance value for a hazard type.
func (h *HazardResistance) GetResistance(hazardType HazardType) float64 {
	if h.Resistances == nil {
		return 0
	}
	return h.Resistances[hazardType]
}

// HazardEffect represents an ongoing effect from hazard exposure.
type HazardEffect struct {
	// SourceHazard identifies the hazard type causing this effect.
	SourceHazard HazardType
	// StackCount is how many times the effect has stacked.
	StackCount int
	// MaxStacks is the maximum allowed stacks.
	MaxStacks int
	// RemainingDuration is the time left on this effect (seconds).
	RemainingDuration float64
	// DamageOverTime is the damage dealt per second while active.
	DamageOverTime float64
	// MovementPenalty reduces movement speed (0.0 = none, 1.0 = immobile).
	MovementPenalty float64
	// VisualEffect identifies the visual effect to display.
	VisualEffect string
	// Curable indicates if the effect can be cured/cleansed.
	Curable bool
}

// Type returns the component type identifier for HazardEffect.
func (h *HazardEffect) Type() string { return "HazardEffect" }

// HazardZone marks a region as containing environmental hazards.
type HazardZone struct {
	// ZoneName identifies this hazard zone.
	ZoneName string
	// HazardTypes lists all hazard types present in this zone.
	HazardTypes []HazardType
	// ZoneLevel indicates hazard severity (1 = minor, 5 = lethal).
	ZoneLevel int
	// RequiredProtection lists equipment needed to safely traverse.
	RequiredProtection []string
	// WarningDisplayed tracks if the player has been warned about this zone.
	WarningDisplayed bool
	// EnterTime tracks when the entity entered this zone.
	EnterTime float64
	// SafePoints lists coordinates of safe spots within the zone.
	SafePoints [][2]float64
}

// Type returns the component type identifier for HazardZone.
func (h *HazardZone) Type() string { return "HazardZone" }

// WeatherHazard represents hazardous weather conditions.
type WeatherHazard struct {
	// WeatherType identifies the hazardous weather ("storm", "blizzard", "sandstorm", "acid_rain", "meteor_shower").
	WeatherType string
	// Severity indicates intensity (0.0 = mild, 1.0 = extreme).
	Severity float64
	// StartTime is when the weather hazard began.
	StartTime float64
	// Duration is how long the weather will last (seconds).
	Duration float64
	// OutdoorDamage is damage per second when outside shelter.
	OutdoorDamage float64
	// VisibilityReduction reduces view distance (0.0 = none, 1.0 = blind).
	VisibilityReduction float64
	// MovementPenalty affects movement speed (0.0 = none, 1.0 = immobile).
	MovementPenalty float64
}

// Type returns the component type identifier for WeatherHazard.
func (w *WeatherHazard) Type() string { return "WeatherHazard" }

// TrapMechanism represents mechanical or magical traps.
type TrapMechanism struct {
	// TrapType identifies the trap ("spike", "dart", "pit", "flame", "magic", "alarm").
	TrapType string
	// Triggered indicates if the trap has been activated.
	Triggered bool
	// Armed indicates if the trap is ready to trigger.
	Armed bool
	// DetectionDifficulty is the skill needed to detect (0.0-1.0).
	DetectionDifficulty float64
	// DisarmDifficulty is the skill needed to disarm (0.0-1.0).
	DisarmDifficulty float64
	// Damage is the damage dealt on trigger.
	Damage float64
	// ResetTime is the time until the trap resets (0 = single use).
	ResetTime float64
	// TriggerRadius is the activation radius.
	TriggerRadius float64
}

// Type returns the component type identifier for TrapMechanism.
func (t *TrapMechanism) Type() string { return "TrapMechanism" }

// CityEvent represents a dynamic event occurring in a city.
type CityEvent struct {
	// EventType identifies the event category.
	EventType string
	// Name is the human-readable event name.
	Name string
	// Description provides details about the event.
	Description string
	// CityID identifies the city where this event is occurring.
	CityID string
	// DistrictName is the affected district (if applicable).
	DistrictName string
	// StartTime is the game time when the event started.
	StartTime float64
	// Duration is how long the event lasts in game hours.
	Duration float64
	// Severity is the event impact level (0.0-1.0).
	Severity float64
	// Active indicates if the event is currently happening.
	Active bool
	// Effects contains modifiers applied during the event.
	Effects CityEventEffects
	// ParticipantRequirements specifies who can participate.
	ParticipantRequirements CityEventRequirements
}

// Type returns the component type identifier for CityEvent.
func (c *CityEvent) Type() string { return "CityEvent" }

// CityEventEffects contains gameplay modifiers during an event.
// NOTE: This is a helper struct used within CityEvent, not an ECS component.
type CityEventEffects struct {
	// ShopPriceMultiplier affects shop prices (1.0 = normal).
	ShopPriceMultiplier float64
	// CrimePenaltyMultiplier affects crime penalties (1.0 = normal).
	CrimePenaltyMultiplier float64
	// NPCActivityChange overrides NPC schedules if non-empty.
	NPCActivityChange string
	// SpawnRateMultiplier affects hostile NPC spawn rate (1.0 = normal).
	SpawnRateMultiplier float64
	// QuestRewardMultiplier affects quest rewards (1.0 = normal).
	QuestRewardMultiplier float64
	// GuardPatrolMultiplier affects guard presence (1.0 = normal).
	GuardPatrolMultiplier float64
}

// CityEventRequirements specifies participation requirements.
// NOTE: This is a helper struct used within CityEvent, not an ECS component.
type CityEventRequirements struct {
	// MinFactionReputation required to participate.
	MinFactionReputation float64
	// RequiredFaction limits participation to a faction (empty = any).
	RequiredFaction string
	// MinLevel is the minimum player level required.
	MinLevel int
	// RequiredItems lists items needed to participate.
	RequiredItems []string
}

// NPCOccupation defines an NPC's job role and work behaviors.
type NPCOccupation struct {
	// OccupationType is the type of occupation (merchant, guard, blacksmith, etc.)
	OccupationType string
	// SkillLevel is the NPC's proficiency at their job (0.0-1.0).
	SkillLevel float64
	// WorkplaceID is the entity ID or location ID where the NPC works.
	WorkplaceID string
	// CurrentTask is what the NPC is currently doing at work.
	CurrentTask string
	// TaskProgress is progress on current task (0.0-1.0).
	TaskProgress float64
	// TaskDuration is the base duration of tasks in seconds.
	TaskDuration float64
	// GoldPerHour is base income when working.
	GoldPerHour float64
	// CanTrade indicates if NPC can trade with players.
	CanTrade bool
	// CanCraft indicates if NPC can craft items.
	CanCraft bool
	// CanProvideQuests indicates if NPC can give quests.
	CanProvideQuests bool
	// CanTrain indicates if NPC can train player skills.
	CanTrain bool
	// TrainableSkills lists skills this NPC can teach.
	TrainableSkills []string
	// Inventory tracks work-related items (for merchants, craftsmen).
	WorkInventory []OccupationItem
	// CustomerQueue tracks entities waiting for service.
	CustomerQueue []uint64
	// IsWorking indicates if NPC is actively working.
	IsWorking bool
	// WorkEfficiency modifies task speed (1.0 = normal).
	WorkEfficiency float64
	// Fatigue reduces efficiency over time (0.0 = fresh, 1.0 = exhausted).
	Fatigue float64
	// LastRestTime is when the NPC last rested.
	LastRestTime float64
}

// Type returns the component type identifier for NPCOccupation.
func (o *NPCOccupation) Type() string { return "NPCOccupation" }

// OccupationItem represents an item in an NPC's work inventory.
// NOTE: This is a helper struct used within NPCOccupation, not an ECS component.
type OccupationItem struct {
	// ItemID is the item type identifier.
	ItemID string
	// Quantity is how many of this item the NPC has.
	Quantity int
	// BasePrice is the base price for this item.
	BasePrice float64
	// Quality affects price and effectiveness (0.0-1.0).
	Quality float64
}

// DialogState tracks an entity's current dialog status and available options.
type DialogState struct {
	// IsInDialog indicates if the entity is currently in a conversation.
	IsInDialog bool
	// ConversationPartner is the entity ID of who they're talking to.
	ConversationPartner uint64
	// CurrentTopicID is the current topic or node in the dialog tree.
	CurrentTopicID string
	// AvailableResponses are the current dialog options.
	AvailableResponses []DialogOption
	// DialogHistory tracks recent exchanges in this conversation.
	DialogHistory []DialogExchange
	// StartTime is when the conversation began.
	StartTime float64
}

// Type returns the component type identifier for DialogState.
func (d *DialogState) Type() string { return "DialogState" }

// DialogOption represents a single dialog choice.
// NOTE: This is a helper struct used within DialogState, not an ECS component.
type DialogOption struct {
	// ID uniquely identifies this option.
	ID string
	// Text is the displayed response text.
	Text string
	// NextTopicID is where this option leads.
	NextTopicID string
	// Requirements are conditions needed to show this option.
	Requirements DialogRequirements
	// Consequences are effects triggered when selected.
	Consequences DialogConsequences
	// IsVisible indicates if the option is currently shown.
	IsVisible bool
	// IsEnabled indicates if the option can be selected.
	IsEnabled bool
}

// DialogRequirements specifies conditions for a dialog option.
// NOTE: This is a helper struct used within DialogOption, not an ECS component.
type DialogRequirements struct {
	// MinReputation is minimum faction reputation needed.
	MinReputation float64
	// RequiredItems are items the player must have.
	RequiredItems []string
	// RequiredSkill is a skill check (skill_name:level).
	RequiredSkill string
	// RequiredQuest is a quest that must be active/complete.
	RequiredQuest string
	// RequiredFlag is a world flag that must be set.
	RequiredFlag string
	// MinGold is minimum gold required.
	MinGold float64
}

// DialogConsequences defines effects of selecting a dialog option.
// NOTE: This is a helper struct used within DialogOption, not an ECS component.
type DialogConsequences struct {
	// ReputationChange adjusts faction reputation.
	ReputationChange float64
	// FactionID specifies which faction's reputation changes.
	FactionID string
	// GoldChange adds/removes gold (positive = gain).
	GoldChange float64
	// ItemsGiven are items the player receives.
	ItemsGiven []string
	// ItemsTaken are items removed from the player.
	ItemsTaken []string
	// QuestStart is a quest ID to begin.
	QuestStart string
	// QuestProgress is a quest stage to advance to.
	QuestProgress string
	// QuestComplete is a quest ID to mark complete.
	QuestComplete string
	// FlagSet is a world flag to set.
	FlagSet string
	// FlagClear is a world flag to clear.
	FlagClear string
	// RelationshipChange adjusts NPC relationship.
	RelationshipChange float64
	// TriggerCombat starts combat with this NPC.
	TriggerCombat bool
	// NPCMood changes the NPC's emotional state.
	NPCMood string
	// SpawnEntity creates a new entity (NPC, item, etc).
	SpawnEntity string
	// TeleportPlayer moves the player to a location.
	TeleportPlayer string
}

// DialogExchange records a single exchange in dialog history.
// NOTE: This is a helper struct used within DialogState, not an ECS component.
type DialogExchange struct {
	// Speaker is the entity ID who spoke.
	Speaker uint64
	// Text is what was said.
	Text string
	// OptionID is the ID of the selected option (for player).
	OptionID string
	// Timestamp is when this exchange occurred.
	Timestamp float64
}

// DialogMemory tracks an NPC's memory of past conversations.
type DialogMemory struct {
	// ConversationCount is total conversations with this entity.
	ConversationCount int
	// LastConversationTime is when the last conversation ended.
	LastConversationTime float64
	// TopicsDiscussed maps topic IDs to times discussed.
	TopicsDiscussed map[string]int
	// ImportantEvents records significant dialog outcomes.
	ImportantEvents []DialogMemoryEvent
	// Attitude is the NPC's overall attitude toward this entity (-1 to 1).
	Attitude float64
	// KnownFacts are facts the NPC knows about the player.
	KnownFacts map[string]bool
	// GiftsReceived tracks gifts given by the player.
	GiftsReceived []string
	// PromisesMade tracks promises made in conversation.
	PromisesMade []DialogPromise
}

// Type returns the component type identifier for DialogMemory.
func (m *DialogMemory) Type() string { return "DialogMemory" }

// DialogMemoryEvent records a significant dialog outcome.
// NOTE: This is a helper struct used within DialogMemory, not an ECS component.
type DialogMemoryEvent struct {
	// EventType categorizes the event (quest_given, insulted, helped, etc).
	EventType string
	// Description is a brief description of what happened.
	Description string
	// Timestamp is when this event occurred.
	Timestamp float64
	// Sentiment is how the NPC felt about this (-1 to 1).
	Sentiment float64
}

// DialogPromise tracks a promise made during conversation.
// NOTE: This is a helper struct used within DialogMemory, not an ECS component.
type DialogPromise struct {
	// PromiseID uniquely identifies this promise.
	PromiseID string
	// Description is what was promised.
	Description string
	// Deadline is when the promise should be fulfilled (0 = no deadline).
	Deadline float64
	// IsFulfilled indicates if the promise was kept.
	IsFulfilled bool
	// IsBroken indicates if the promise was broken.
	IsBroken bool
}

// Appearance defines the visual representation of an entity in the first-person view.
// It is pure data — rendering logic belongs in the SpriteRenderSystem.
type Appearance struct {
	// SpriteCategory selects the generation algorithm.
	// One of: "humanoid", "creature", "vehicle", "object", "effect"
	SpriteCategory string

	// BodyPlan selects the silhouette template within the category.
	// Examples: "warrior", "merchant", "wolf", "dragon", "horse", "buggy"
	BodyPlan string

	// PrimaryColor and SecondaryColor are packed RGBA values (genre-derived).
	PrimaryColor   uint32
	SecondaryColor uint32

	// AccentColor for details (belt, trim, insignia).
	AccentColor uint32

	// Scale multiplier relative to default entity height (1.0 = standard humanoid).
	// Range: 0.25 (small critter) to 4.0 (dragon/mech).
	Scale float64

	// AnimState is the current animation state identifier.
	// One of: "idle", "walk", "run", "attack", "cast", "sneak", "dead", "sit", "work"
	AnimState string

	// AnimFrame is the current frame index within the animation.
	AnimFrame int

	// AnimTimer accumulates dt for frame advancement.
	AnimTimer float64

	// Visible controls whether the entity is rendered at all.
	// Set to false for hidden/despawned entities.
	Visible bool

	// Opacity controls alpha blending (0.0 = invisible, 1.0 = opaque).
	// Driven by Stealth.Visibility when sneaking.
	Opacity float64

	// FlipH mirrors the sprite horizontally (for facing direction).
	FlipH bool

	// Decorations are additional overlay identifiers (armor, hat, weapon held).
	Decorations []string

	// DamageOverlay intensity (0.0 = pristine, 1.0 = heavily damaged).
	// Driven by Health.Current/Health.Max or VehicleState.DamagePercent.
	DamageOverlay float64

	// GenreID is stored for sprite generation cache keying.
	GenreID string
}

// Type returns the component type identifier for Appearance.
func (a *Appearance) Type() string { return "Appearance" }

// NewAppearance creates a new Appearance with default visible settings.
func NewAppearance(category, bodyPlan, genre string) *Appearance {
	return &Appearance{
		SpriteCategory: category,
		BodyPlan:       bodyPlan,
		Scale:          1.0,
		AnimState:      "idle",
		AnimFrame:      0,
		AnimTimer:      0.0,
		Visible:        true,
		Opacity:        1.0,
		FlipH:          false,
		Decorations:    nil,
		DamageOverlay:  0.0,
		GenreID:        genre,
	}
}

// NewObjectAppearance creates an Appearance for interactive objects like doors, crates, etc.
// The bodyPlan should match the object type: "door", "crate", "chest", "lever", etc.
func NewObjectAppearance(bodyPlan, genre string) *Appearance {
	return NewAppearance("object", bodyPlan, genre)
}

// NewCreatureAppearance creates an Appearance for creatures and enemies.
// The bodyPlan should match creature types: "wolf", "dragon", "spider", etc.
func NewCreatureAppearance(bodyPlan, genre string) *Appearance {
	return NewAppearance("creature", bodyPlan, genre)
}

// NewHumanoidAppearance creates an Appearance for humanoid NPCs.
// The bodyPlan should match humanoid types: "warrior", "merchant", "mage", etc.
func NewHumanoidAppearance(bodyPlan, genre string) *Appearance {
	return NewAppearance("humanoid", bodyPlan, genre)
}

// NewVehicleAppearance creates an Appearance for vehicles.
// The bodyPlan should match vehicle types: "horse", "buggy", "mech", etc.
func NewVehicleAppearance(bodyPlan, genre string) *Appearance {
	return NewAppearance("vehicle", bodyPlan, genre)
}

// NewEffectAppearance creates an Appearance for visual effects like explosions, magic, etc.
// The bodyPlan should match effect types: "explosion", "fireball", "smoke", etc.
func NewEffectAppearance(bodyPlan, genre string) *Appearance {
	return NewAppearance("effect", bodyPlan, genre)
}

// BarrierShape defines the collision and visual profile of a barrier.
// NOTE: This is a helper struct used within Barrier, not an ECS component.
type BarrierShape struct {
	// ShapeType is the collision shape type: "cylinder", "box", "polygon", "billboard".
	ShapeType string
	// Radius is used for cylinder shapes.
	Radius float64
	// Width is used for box shapes.
	Width float64
	// Depth is used for box shapes.
	Depth float64
	// Height is the world-space height of the barrier.
	Height float64
	// Vertices defines polygon shapes as [x0,y0, x1,y1, ...] relative to center.
	Vertices []float64
	// SpriteKey is the key into sprite cache for visual representation.
	SpriteKey string
	// MaterialID indexes into MaterialRegistry for collision sound/effects.
	MaterialID int
	// ClimbHeight is the maximum height the player can climb over (0 = not climbable).
	// Barriers with Height <= ClimbHeight can be climbed by the player.
	ClimbHeight float64
}

// Barrier is an ECS component for environmental barriers.
type Barrier struct {
	// Shape defines the collision and visual profile.
	Shape BarrierShape
	// Genre is the genre that generated this barrier.
	Genre string
	// Destructible indicates if the barrier can be destroyed.
	Destructible bool
	// HitPoints is the current health for destructible barriers.
	HitPoints float64
	// MaxHP is the maximum health for destructible barriers.
	MaxHP float64
	// MaterialType describes the barrier's material (wood, stone, metal, glass, ice).
	MaterialType string
	// DestructionProcessed tracks if the destruction has been handled by the system.
	DestructionProcessed bool
}

// Type returns the component type identifier for Barrier.
func (b *Barrier) Type() string { return "Barrier" }

// NewBarrier creates a new Barrier component with the given shape and genre.
func NewBarrier(shapeType, spriteKey, genre string, height float64) *Barrier {
	return &Barrier{
		Shape: BarrierShape{
			ShapeType: shapeType,
			Height:    height,
			SpriteKey: spriteKey,
		},
		Genre:        genre,
		Destructible: false,
		HitPoints:    0,
		MaxHP:        0,
	}
}

// NewDestructibleBarrier creates a destructible Barrier with health.
func NewDestructibleBarrier(shapeType, spriteKey, genre string, height, maxHP float64) *Barrier {
	return &Barrier{
		Shape: BarrierShape{
			ShapeType: shapeType,
			Height:    height,
			SpriteKey: spriteKey,
		},
		Genre:        genre,
		Destructible: true,
		HitPoints:    maxHP,
		MaxHP:        maxHP,
	}
}

// IsDamaged returns true if the barrier has taken damage.
func (b *Barrier) IsDamaged() bool {
	return b.Destructible && b.HitPoints < b.MaxHP
}

// IsDestroyed returns true if the barrier has been destroyed.
func (b *Barrier) IsDestroyed() bool {
	return b.Destructible && b.HitPoints <= 0
}

// DamagePercent returns the damage percentage (0.0 = pristine, 1.0 = destroyed).
func (b *Barrier) DamagePercent() float64 {
	if !b.Destructible || b.MaxHP <= 0 {
		return 0
	}
	return 1.0 - (b.HitPoints / b.MaxHP)
}

// BarrierCategory represents a category of barrier (natural, constructed, organic).
type BarrierCategory string

const (
	// BarrierCategoryNatural represents natural terrain obstacles.
	BarrierCategoryNatural BarrierCategory = "natural"
	// BarrierCategoryConstructed represents man-made structural barriers.
	BarrierCategoryConstructed BarrierCategory = "constructed"
	// BarrierCategoryOrganic represents living or once-living barriers.
	BarrierCategoryOrganic BarrierCategory = "organic"
)

// BarrierArchetype defines a template for generating barriers of a specific type.
type BarrierArchetype struct {
	// ID is the unique identifier for this archetype.
	ID string
	// Name is the display name for this barrier type.
	Name string
	// Category is the barrier category (natural, constructed, organic).
	Category BarrierCategory
	// Genre is the genre this archetype belongs to.
	Genre string
	// ShapeType is the default collision shape type.
	ShapeType string
	// BaseRadius is the default radius for cylinder shapes.
	BaseRadius float64
	// BaseWidth is the default width for box shapes.
	BaseWidth float64
	// BaseDepth is the default depth for box shapes.
	BaseDepth float64
	// BaseHeight is the default height.
	BaseHeight float64
	// Destructible indicates if barriers of this type can be destroyed.
	Destructible bool
	// BaseHP is the base hit points for destructible barriers.
	BaseHP float64
	// SpawnWeight is the relative spawn probability (higher = more common).
	SpawnWeight float64
	// MaterialID is the default material ID.
	MaterialID int
}

// BarrierArchetypeRegistry holds all registered barrier archetypes by genre.
type BarrierArchetypeRegistry struct {
	archetypes map[string][]BarrierArchetype
}

// NewBarrierArchetypeRegistry creates a registry pre-populated with genre archetypes.
func NewBarrierArchetypeRegistry() *BarrierArchetypeRegistry {
	r := &BarrierArchetypeRegistry{
		archetypes: make(map[string][]BarrierArchetype),
	}
	r.registerFantasyArchetypes()
	r.registerSciFiArchetypes()
	r.registerHorrorArchetypes()
	r.registerCyberpunkArchetypes()
	r.registerPostApocArchetypes()
	return r
}

// GetArchetypes returns all archetypes for a given genre.
func (r *BarrierArchetypeRegistry) GetArchetypes(genre string) []BarrierArchetype {
	return r.archetypes[genre]
}

// GetArchetypesByCategory returns archetypes for a genre filtered by category.
func (r *BarrierArchetypeRegistry) GetArchetypesByCategory(genre string, category BarrierCategory) []BarrierArchetype {
	all := r.archetypes[genre]
	var filtered []BarrierArchetype
	for _, a := range all {
		if a.Category == category {
			filtered = append(filtered, a)
		}
	}
	return filtered
}

// GetArchetypeByID returns a specific archetype by ID and genre.
func (r *BarrierArchetypeRegistry) GetArchetypeByID(genre, id string) (BarrierArchetype, bool) {
	for _, a := range r.archetypes[genre] {
		if a.ID == id {
			return a, true
		}
	}
	return BarrierArchetype{}, false
}

func (r *BarrierArchetypeRegistry) registerFantasyArchetypes() {
	r.archetypes["fantasy"] = []BarrierArchetype{
		// Natural
		{ID: "boulder", Name: "Boulder", Category: BarrierCategoryNatural, Genre: "fantasy", ShapeType: "cylinder", BaseRadius: 0.8, BaseHeight: 1.2, SpawnWeight: 3.0, MaterialID: 1},
		{ID: "ancient_tree", Name: "Ancient Tree", Category: BarrierCategoryNatural, Genre: "fantasy", ShapeType: "cylinder", BaseRadius: 1.2, BaseHeight: 4.0, SpawnWeight: 2.0, MaterialID: 2},
		{ID: "crystal_formation", Name: "Crystal Formation", Category: BarrierCategoryNatural, Genre: "fantasy", ShapeType: "polygon", BaseHeight: 2.5, SpawnWeight: 1.0, MaterialID: 3},
		// Constructed
		{ID: "stone_pillar", Name: "Stone Pillar", Category: BarrierCategoryConstructed, Genre: "fantasy", ShapeType: "cylinder", BaseRadius: 0.5, BaseHeight: 3.0, SpawnWeight: 2.0, MaterialID: 1},
		{ID: "archway", Name: "Stone Archway", Category: BarrierCategoryConstructed, Genre: "fantasy", ShapeType: "box", BaseWidth: 2.0, BaseDepth: 0.5, BaseHeight: 3.5, SpawnWeight: 1.0, MaterialID: 1},
		{ID: "statue", Name: "Statue", Category: BarrierCategoryConstructed, Genre: "fantasy", ShapeType: "cylinder", BaseRadius: 0.6, BaseHeight: 2.5, Destructible: true, BaseHP: 150, SpawnWeight: 1.5, MaterialID: 1},
		// Organic
		{ID: "hedgerow", Name: "Hedgerow", Category: BarrierCategoryOrganic, Genre: "fantasy", ShapeType: "box", BaseWidth: 2.0, BaseDepth: 0.6, BaseHeight: 1.8, Destructible: true, BaseHP: 50, SpawnWeight: 2.5, MaterialID: 2},
		{ID: "thornbush", Name: "Thornbush", Category: BarrierCategoryOrganic, Genre: "fantasy", ShapeType: "cylinder", BaseRadius: 0.7, BaseHeight: 1.2, Destructible: true, BaseHP: 30, SpawnWeight: 2.0, MaterialID: 2},
		{ID: "vine_wall", Name: "Vine Wall", Category: BarrierCategoryOrganic, Genre: "fantasy", ShapeType: "box", BaseWidth: 3.0, BaseDepth: 0.3, BaseHeight: 2.5, Destructible: true, BaseHP: 80, SpawnWeight: 1.5, MaterialID: 2},
	}
}

func (r *BarrierArchetypeRegistry) registerSciFiArchetypes() {
	r.archetypes["sci-fi"] = []BarrierArchetype{
		// Natural
		{ID: "alien_rock", Name: "Alien Rock", Category: BarrierCategoryNatural, Genre: "sci-fi", ShapeType: "polygon", BaseHeight: 1.5, SpawnWeight: 2.5, MaterialID: 4},
		{ID: "fungal_growth", Name: "Fungal Growth", Category: BarrierCategoryNatural, Genre: "sci-fi", ShapeType: "cylinder", BaseRadius: 0.9, BaseHeight: 1.8, Destructible: true, BaseHP: 40, SpawnWeight: 2.0, MaterialID: 5},
		{ID: "crystal_node", Name: "Crystal Node", Category: BarrierCategoryNatural, Genre: "sci-fi", ShapeType: "cylinder", BaseRadius: 0.4, BaseHeight: 2.0, SpawnWeight: 1.5, MaterialID: 3},
		// Constructed
		{ID: "steel_beam", Name: "Steel Beam", Category: BarrierCategoryConstructed, Genre: "sci-fi", ShapeType: "box", BaseWidth: 0.3, BaseDepth: 0.3, BaseHeight: 4.0, SpawnWeight: 2.0, MaterialID: 6},
		{ID: "energy_pylon", Name: "Energy Pylon", Category: BarrierCategoryConstructed, Genre: "sci-fi", ShapeType: "cylinder", BaseRadius: 0.5, BaseHeight: 3.5, SpawnWeight: 1.5, MaterialID: 6},
		{ID: "antenna_array", Name: "Antenna Array", Category: BarrierCategoryConstructed, Genre: "sci-fi", ShapeType: "box", BaseWidth: 1.5, BaseDepth: 1.5, BaseHeight: 5.0, Destructible: true, BaseHP: 200, SpawnWeight: 1.0, MaterialID: 6},
		// Organic
		{ID: "bio_pod", Name: "Bio-Pod", Category: BarrierCategoryOrganic, Genre: "sci-fi", ShapeType: "cylinder", BaseRadius: 0.8, BaseHeight: 1.5, Destructible: true, BaseHP: 60, SpawnWeight: 2.0, MaterialID: 5},
		{ID: "growth_membrane", Name: "Growth Membrane", Category: BarrierCategoryOrganic, Genre: "sci-fi", ShapeType: "box", BaseWidth: 2.5, BaseDepth: 0.2, BaseHeight: 2.0, Destructible: true, BaseHP: 25, SpawnWeight: 1.5, MaterialID: 5},
		{ID: "tendril_curtain", Name: "Tendril Curtain", Category: BarrierCategoryOrganic, Genre: "sci-fi", ShapeType: "box", BaseWidth: 3.0, BaseDepth: 0.3, BaseHeight: 3.0, Destructible: true, BaseHP: 40, SpawnWeight: 1.0, MaterialID: 5},
	}
}

func (r *BarrierArchetypeRegistry) registerHorrorArchetypes() {
	r.archetypes["horror"] = []BarrierArchetype{
		// Natural
		{ID: "gnarled_tree", Name: "Gnarled Tree", Category: BarrierCategoryNatural, Genre: "horror", ShapeType: "cylinder", BaseRadius: 1.0, BaseHeight: 3.5, SpawnWeight: 2.5, MaterialID: 2},
		{ID: "bone_pile", Name: "Bone Pile", Category: BarrierCategoryNatural, Genre: "horror", ShapeType: "cylinder", BaseRadius: 0.9, BaseHeight: 0.8, Destructible: true, BaseHP: 30, SpawnWeight: 2.0, MaterialID: 7},
		{ID: "pulsing_hive", Name: "Pulsing Hive", Category: BarrierCategoryNatural, Genre: "horror", ShapeType: "cylinder", BaseRadius: 1.1, BaseHeight: 1.5, Destructible: true, BaseHP: 80, SpawnWeight: 1.0, MaterialID: 5},
		// Constructed
		{ID: "iron_gate", Name: "Iron Gate", Category: BarrierCategoryConstructed, Genre: "horror", ShapeType: "box", BaseWidth: 2.5, BaseDepth: 0.2, BaseHeight: 3.0, SpawnWeight: 1.5, MaterialID: 6},
		{ID: "tombstone", Name: "Tombstone", Category: BarrierCategoryConstructed, Genre: "horror", ShapeType: "box", BaseWidth: 0.8, BaseDepth: 0.3, BaseHeight: 1.5, Destructible: true, BaseHP: 100, SpawnWeight: 3.0, MaterialID: 1},
		{ID: "ritual_circle", Name: "Ritual Circle", Category: BarrierCategoryConstructed, Genre: "horror", ShapeType: "cylinder", BaseRadius: 1.5, BaseHeight: 0.1, SpawnWeight: 0.5, MaterialID: 1},
		// Organic
		{ID: "flesh_wall", Name: "Flesh Wall", Category: BarrierCategoryOrganic, Genre: "horror", ShapeType: "box", BaseWidth: 3.0, BaseDepth: 0.5, BaseHeight: 2.5, Destructible: true, BaseHP: 100, SpawnWeight: 1.0, MaterialID: 5},
		{ID: "web_cluster", Name: "Web Cluster", Category: BarrierCategoryOrganic, Genre: "horror", ShapeType: "box", BaseWidth: 2.0, BaseDepth: 2.0, BaseHeight: 2.5, Destructible: true, BaseHP: 20, SpawnWeight: 2.0, MaterialID: 8},
		{ID: "fungal_mass", Name: "Fungal Mass", Category: BarrierCategoryOrganic, Genre: "horror", ShapeType: "cylinder", BaseRadius: 1.2, BaseHeight: 1.0, Destructible: true, BaseHP: 50, SpawnWeight: 1.5, MaterialID: 5},
	}
}

func (r *BarrierArchetypeRegistry) registerCyberpunkArchetypes() {
	r.archetypes["cyberpunk"] = []BarrierArchetype{
		// Natural (urban "natural")
		{ID: "toxic_drum", Name: "Toxic Waste Drum", Category: BarrierCategoryNatural, Genre: "cyberpunk", ShapeType: "cylinder", BaseRadius: 0.4, BaseHeight: 1.0, Destructible: true, BaseHP: 50, SpawnWeight: 3.0, MaterialID: 6},
		{ID: "mutant_flora", Name: "Mutant Flora", Category: BarrierCategoryNatural, Genre: "cyberpunk", ShapeType: "cylinder", BaseRadius: 0.7, BaseHeight: 1.5, Destructible: true, BaseHP: 30, SpawnWeight: 1.5, MaterialID: 5},
		{ID: "debris_pile", Name: "Debris Pile", Category: BarrierCategoryNatural, Genre: "cyberpunk", ShapeType: "polygon", BaseHeight: 0.8, Destructible: true, BaseHP: 40, SpawnWeight: 2.5, MaterialID: 9},
		// Constructed
		{ID: "neon_sign", Name: "Neon Sign", Category: BarrierCategoryConstructed, Genre: "cyberpunk", ShapeType: "box", BaseWidth: 2.0, BaseDepth: 0.2, BaseHeight: 3.0, Destructible: true, BaseHP: 60, SpawnWeight: 2.0, MaterialID: 6},
		{ID: "holographic_wall", Name: "Holographic Wall", Category: BarrierCategoryConstructed, Genre: "cyberpunk", ShapeType: "box", BaseWidth: 3.0, BaseDepth: 0.1, BaseHeight: 2.5, SpawnWeight: 1.0, MaterialID: 10},
		{ID: "vending_machine", Name: "Vending Machine", Category: BarrierCategoryConstructed, Genre: "cyberpunk", ShapeType: "box", BaseWidth: 1.0, BaseDepth: 0.8, BaseHeight: 2.0, Destructible: true, BaseHP: 150, SpawnWeight: 2.5, MaterialID: 6},
		// Organic
		{ID: "graffiti_barrier", Name: "Graffiti Barrier", Category: BarrierCategoryOrganic, Genre: "cyberpunk", ShapeType: "box", BaseWidth: 2.5, BaseDepth: 0.3, BaseHeight: 2.0, Destructible: true, BaseHP: 80, SpawnWeight: 2.0, MaterialID: 9},
		{ID: "plant_wall", Name: "Urban Plant Wall", Category: BarrierCategoryOrganic, Genre: "cyberpunk", ShapeType: "box", BaseWidth: 3.0, BaseDepth: 0.4, BaseHeight: 2.5, Destructible: true, BaseHP: 40, SpawnWeight: 1.0, MaterialID: 2},
		{ID: "bio_wire_tangle", Name: "Bio-Wire Tangle", Category: BarrierCategoryOrganic, Genre: "cyberpunk", ShapeType: "cylinder", BaseRadius: 1.0, BaseHeight: 1.5, Destructible: true, BaseHP: 35, SpawnWeight: 1.5, MaterialID: 10},
	}
}

func (r *BarrierArchetypeRegistry) registerPostApocArchetypes() {
	r.archetypes["post-apocalyptic"] = []BarrierArchetype{
		// Natural
		{ID: "rubble_mound", Name: "Rubble Mound", Category: BarrierCategoryNatural, Genre: "post-apocalyptic", ShapeType: "polygon", BaseHeight: 1.2, SpawnWeight: 3.0, MaterialID: 9},
		{ID: "burnt_tree", Name: "Burnt Tree", Category: BarrierCategoryNatural, Genre: "post-apocalyptic", ShapeType: "cylinder", BaseRadius: 0.6, BaseHeight: 2.5, SpawnWeight: 2.0, MaterialID: 2},
		{ID: "crater_rim", Name: "Crater Rim", Category: BarrierCategoryNatural, Genre: "post-apocalyptic", ShapeType: "polygon", BaseHeight: 0.6, SpawnWeight: 1.0, MaterialID: 9},
		// Constructed
		{ID: "barricade", Name: "Barricade", Category: BarrierCategoryConstructed, Genre: "post-apocalyptic", ShapeType: "box", BaseWidth: 2.5, BaseDepth: 0.5, BaseHeight: 1.5, Destructible: true, BaseHP: 100, SpawnWeight: 2.5, MaterialID: 9},
		{ID: "wrecked_car", Name: "Wrecked Car", Category: BarrierCategoryConstructed, Genre: "post-apocalyptic", ShapeType: "box", BaseWidth: 2.0, BaseDepth: 4.0, BaseHeight: 1.5, Destructible: true, BaseHP: 200, SpawnWeight: 2.0, MaterialID: 6},
		{ID: "makeshift_wall", Name: "Makeshift Wall", Category: BarrierCategoryConstructed, Genre: "post-apocalyptic", ShapeType: "box", BaseWidth: 3.0, BaseDepth: 0.3, BaseHeight: 2.0, Destructible: true, BaseHP: 80, SpawnWeight: 2.0, MaterialID: 9},
		// Organic
		{ID: "overgrown_ruin", Name: "Overgrown Ruin", Category: BarrierCategoryOrganic, Genre: "post-apocalyptic", ShapeType: "polygon", BaseHeight: 2.0, SpawnWeight: 1.5, MaterialID: 9},
		{ID: "thorn_thicket", Name: "Thorn Thicket", Category: BarrierCategoryOrganic, Genre: "post-apocalyptic", ShapeType: "cylinder", BaseRadius: 1.0, BaseHeight: 1.5, Destructible: true, BaseHP: 40, SpawnWeight: 2.0, MaterialID: 2},
		{ID: "rad_fungus", Name: "Radioactive Fungus", Category: BarrierCategoryOrganic, Genre: "post-apocalyptic", ShapeType: "cylinder", BaseRadius: 0.8, BaseHeight: 1.0, Destructible: true, BaseHP: 25, SpawnWeight: 1.5, MaterialID: 5},
	}
}

// CreateBarrierFromArchetype creates a Barrier component from an archetype template.
func CreateBarrierFromArchetype(arch BarrierArchetype, spriteKey string) *Barrier {
	shape := BarrierShape{
		ShapeType:  arch.ShapeType,
		Radius:     arch.BaseRadius,
		Width:      arch.BaseWidth,
		Depth:      arch.BaseDepth,
		Height:     arch.BaseHeight,
		SpriteKey:  spriteKey,
		MaterialID: arch.MaterialID,
	}

	return &Barrier{
		Shape:        shape,
		Genre:        arch.Genre,
		Destructible: arch.Destructible,
		HitPoints:    arch.BaseHP,
		MaxHP:        arch.BaseHP,
	}
}

// ============================================================
// Environment Object Components (Phase 5)
// ============================================================

// ObjectCategory defines the type of environment object for rendering and interaction.
type ObjectCategory string

const (
	// ObjectCategoryInventoriable items can be picked up and stored in inventory.
	ObjectCategoryInventoriable ObjectCategory = "inventoriable"
	// ObjectCategoryInteractive items can be interacted with but not picked up (doors, levers, chests).
	ObjectCategoryInteractive ObjectCategory = "interactive"
	// ObjectCategoryDecorative items are purely visual with no interaction.
	ObjectCategoryDecorative ObjectCategory = "decorative"
)

// InteractionType defines what kind of interaction an object supports.
type InteractionType string

const (
	// InteractionNone indicates no interaction is possible.
	InteractionNone InteractionType = "none"
	// InteractionPickup allows the player to take the item into inventory.
	InteractionPickup InteractionType = "pickup"
	// InteractionOpen opens a container (chest, door, cabinet).
	InteractionOpen InteractionType = "open"
	// InteractionUse activates the object (lever, button, switch).
	InteractionUse InteractionType = "use"
	// InteractionRead displays text (signs, books, notes).
	InteractionRead InteractionType = "read"
	// InteractionTalk initiates dialog (for interactive NPCs/terminals).
	InteractionTalk InteractionType = "talk"
	// InteractionPush allows the player to push/move the object.
	InteractionPush InteractionType = "push"
	// InteractionExamine displays detailed description.
	InteractionExamine InteractionType = "examine"
)

// EnvironmentObject represents an interactable or decorative object in the world.
type EnvironmentObject struct {
	// Category determines rendering and interaction behavior.
	Category ObjectCategory

	// ObjectType is the specific object identifier (e.g., "health_potion", "wooden_chest", "stone_pillar").
	ObjectType string

	// DisplayName is the name shown to the player.
	DisplayName string

	// SpriteKey indexes into the sprite cache for visual representation.
	SpriteKey string

	// Scale is the world-space scale multiplier (1.0 = normal item size).
	Scale float64

	// InteractionType is the primary interaction this object supports.
	InteractionType InteractionType

	// InteractionRange is the maximum distance (in world units) for interaction.
	InteractionRange float64

	// HighlightState indicates the current visual highlight state.
	// 0 = no highlight, 1 = in range, 2 = targeted
	HighlightState int

	// IsTargeted indicates if this object is currently targeted by a player.
	IsTargeted bool

	// ItemID is the inventory item ID for inventoriable objects.
	ItemID string

	// Quantity is the stack count for inventoriable objects.
	Quantity int

	// ContainerContents holds item IDs for container objects.
	ContainerContents []string

	// IsLocked indicates if the object (door/chest) is locked.
	IsLocked bool

	// LockDifficulty is the skill check difficulty to unlock (0-100).
	LockDifficulty int

	// RequiredKeyID is the key item needed to unlock, if any.
	RequiredKeyID string

	// IsOpen indicates if the object (door/chest) is currently open.
	IsOpen bool

	// IsUsed indicates if the object has been used/activated.
	IsUsed bool

	// UseText is the action text shown (e.g., "Pick up", "Open", "Read").
	UseText string

	// ExamineText is the detailed description shown on examination.
	ExamineText string

	// Genre is used for appropriate sprite generation.
	Genre string

	// MaterialID indexes into MaterialRegistry for sounds and effects.
	MaterialID int
}

// Type returns the component type identifier for EnvironmentObject.
func (e *EnvironmentObject) Type() string { return "EnvironmentObject" }

// CanInteract returns true if the object supports any interaction.
func (e *EnvironmentObject) CanInteract() bool {
	return e.InteractionType != InteractionNone
}

// IsPickupable returns true if the object can be picked up.
func (e *EnvironmentObject) IsPickupable() bool {
	return e.Category == ObjectCategoryInventoriable && e.InteractionType == InteractionPickup
}

// IsContainer returns true if the object is a container (chest, crate, etc).
func (e *EnvironmentObject) IsContainer() bool {
	return e.InteractionType == InteractionOpen && len(e.ContainerContents) > 0
}

// IsDoor returns true if the object is a door.
func (e *EnvironmentObject) IsDoor() bool {
	return e.InteractionType == InteractionOpen && e.ObjectType == "door"
}

// IsUsable returns true if the object can be used/activated.
func (e *EnvironmentObject) IsUsable() bool {
	return e.InteractionType == InteractionUse
}

// NeedsKey returns true if the object requires a specific key to unlock.
func (e *EnvironmentObject) NeedsKey() bool {
	return e.IsLocked && e.RequiredKeyID != ""
}

// NewInventoriableObject creates a new pickupable item.
func NewInventoriableObject(objectType, displayName, itemID, genre string, quantity int) *EnvironmentObject {
	return &EnvironmentObject{
		Category:         ObjectCategoryInventoriable,
		ObjectType:       objectType,
		DisplayName:      displayName,
		Scale:            1.0,
		InteractionType:  InteractionPickup,
		InteractionRange: 2.0,
		ItemID:           itemID,
		Quantity:         quantity,
		UseText:          "Pick up",
		Genre:            genre,
	}
}

// NewInteractiveObject creates a new interactive object (door, lever, etc).
func NewInteractiveObject(objectType, displayName string, interactionType InteractionType, genre string) *EnvironmentObject {
	useText := "Use"
	switch interactionType {
	case InteractionOpen:
		useText = "Open"
	case InteractionUse:
		useText = "Activate"
	case InteractionRead:
		useText = "Read"
	case InteractionPush:
		useText = "Push"
	case InteractionExamine:
		useText = "Examine"
	}

	return &EnvironmentObject{
		Category:         ObjectCategoryInteractive,
		ObjectType:       objectType,
		DisplayName:      displayName,
		Scale:            1.0,
		InteractionType:  interactionType,
		InteractionRange: 2.5,
		UseText:          useText,
		Genre:            genre,
	}
}

// NewDecorativeObject creates a purely visual object with no interaction.
func NewDecorativeObject(objectType, displayName, genre string) *EnvironmentObject {
	return &EnvironmentObject{
		Category:         ObjectCategoryDecorative,
		ObjectType:       objectType,
		DisplayName:      displayName,
		Scale:            1.0,
		InteractionType:  InteractionNone,
		InteractionRange: 0,
		Genre:            genre,
	}
}

// NewContainer creates a container object (chest, crate, etc).
func NewContainer(objectType, displayName, genre string, contents []string) *EnvironmentObject {
	return &EnvironmentObject{
		Category:          ObjectCategoryInteractive,
		ObjectType:        objectType,
		DisplayName:       displayName,
		Scale:             1.0,
		InteractionType:   InteractionOpen,
		InteractionRange:  2.0,
		ContainerContents: contents,
		UseText:           "Open",
		Genre:             genre,
	}
}

// NewLockedContainer creates a locked container object.
func NewLockedContainer(objectType, displayName, genre string, contents []string, lockDifficulty int, keyID string) *EnvironmentObject {
	obj := NewContainer(objectType, displayName, genre, contents)
	obj.IsLocked = true
	obj.LockDifficulty = lockDifficulty
	obj.RequiredKeyID = keyID
	obj.UseText = "Unlock"
	return obj
}

// NewDoor creates a door object.
func NewDoor(displayName, genre string, isLocked bool, lockDifficulty int, keyID string) *EnvironmentObject {
	useText := "Open"
	if isLocked {
		useText = "Unlock"
	}
	return &EnvironmentObject{
		Category:         ObjectCategoryInteractive,
		ObjectType:       "door",
		DisplayName:      displayName,
		Scale:            1.0,
		InteractionType:  InteractionOpen,
		InteractionRange: 2.0,
		IsLocked:         isLocked,
		LockDifficulty:   lockDifficulty,
		RequiredKeyID:    keyID,
		IsOpen:           false,
		UseText:          useText,
		Genre:            genre,
	}
}

// ObjectHighlightNone indicates no highlight.
const ObjectHighlightNone = 0

// ObjectHighlightInRange indicates the object is within interaction range.
const ObjectHighlightInRange = 1

// ObjectHighlightTargeted indicates the object is being targeted by the player.
const ObjectHighlightTargeted = 2

// PhysicsBody enables physics simulation for an entity.
// Used for pushable objects, swinging elements, and basic physics interactions.
type PhysicsBody struct {
	// VelocityX is the X component of velocity (units per second).
	VelocityX float64
	// VelocityY is the Y component of velocity (units per second).
	VelocityY float64
	// VelocityZ is the Z component of velocity (units per second).
	VelocityZ float64
	// Mass affects how the object responds to forces. 0 = infinite mass (immovable).
	Mass float64
	// Friction determines how quickly the object slows down (0-1).
	Friction float64
	// Bounciness determines how much velocity is retained on collision (0-1).
	Bounciness float64
	// IsKinematic when true, the object is moved by code, not physics forces.
	IsKinematic bool
	// IsPushable allows the object to be pushed by player or other entities.
	IsPushable bool
	// IsSwinging indicates the object can swing (like a door or hanging sign).
	IsSwinging bool
	// SwingAngle is the current swing angle in radians (for swinging objects).
	SwingAngle float64
	// SwingVelocity is the current angular velocity for swinging (radians/second).
	SwingVelocity float64
	// SwingDamping reduces swing velocity over time (0-1).
	SwingDamping float64
	// MaxSwingAngle is the maximum swing angle in radians.
	MaxSwingAngle float64
	// PivotOffsetX is the X offset of the pivot point from the object's position.
	PivotOffsetX float64
	// PivotOffsetY is the Y offset of the pivot point from the object's position.
	PivotOffsetY float64
	// CollisionRadius is the radius for collision detection.
	CollisionRadius float64
	// Grounded indicates if the object is resting on a surface.
	Grounded bool
}

// Type returns the component type identifier for PhysicsBody.
func (p *PhysicsBody) Type() string { return "PhysicsBody" }

// ApplyForce adds a force to the physics body, converting to velocity based on mass.
func (p *PhysicsBody) ApplyForce(forceX, forceY, forceZ float64) {
	if p.Mass <= 0 || p.IsKinematic {
		return // Immovable or kinematic objects don't respond to forces
	}
	p.VelocityX += forceX / p.Mass
	p.VelocityY += forceY / p.Mass
	p.VelocityZ += forceZ / p.Mass
}

// ApplyImpulse adds an instant velocity change.
func (p *PhysicsBody) ApplyImpulse(impulseX, impulseY, impulseZ float64) {
	if p.IsKinematic {
		return
	}
	p.VelocityX += impulseX
	p.VelocityY += impulseY
	p.VelocityZ += impulseZ
}

// ApplySwingImpulse adds angular velocity to a swinging object.
func (p *PhysicsBody) ApplySwingImpulse(angularImpulse float64) {
	if !p.IsSwinging {
		return
	}
	p.SwingVelocity += angularImpulse
}

// NewPushableBody creates a physics body for a pushable object.
func NewPushableBody(mass, friction float64) *PhysicsBody {
	return &PhysicsBody{
		Mass:            mass,
		Friction:        friction,
		Bounciness:      0.2,
		IsPushable:      true,
		CollisionRadius: 0.5,
		Grounded:        true,
	}
}

// NewSwingingBody creates a physics body for a swinging object (door, sign).
func NewSwingingBody(maxAngle, damping float64) *PhysicsBody {
	return &PhysicsBody{
		Mass:          1.0,
		IsSwinging:    true,
		SwingDamping:  damping,
		MaxSwingAngle: maxAngle,
		IsKinematic:   true, // Position controlled by swing, not linear physics
	}
}

// ============================================================
// Interactable Component
// ============================================================

// Interactable represents an entity that can be interacted with by the player.
// This component is used for doors, containers, NPCs, switches, and other interactive objects.
type Interactable struct {
	// InteractionType describes what kind of interaction is available.
	// Common types: "use", "open", "pickup", "talk", "read", "activate", "unlock"
	InteractionType string

	// Range is the maximum distance (in world units) from which the player can interact.
	Range float64

	// Prompt is the text shown to the player when in range (e.g., "Press E to open").
	Prompt string

	// Cooldown is the minimum time (in seconds) between interactions.
	Cooldown float64

	// LastInteractTime tracks when the entity was last interacted with.
	LastInteractTime float64

	// Locked indicates whether the object requires a key or condition to interact.
	Locked bool

	// KeyID is the item ID required to unlock (if Locked is true).
	KeyID string

	// RequiredSkill is the skill check required (e.g., "lockpicking", "hacking").
	RequiredSkill string

	// SkillDifficulty is the skill level required to pass the check.
	SkillDifficulty int

	// SingleUse indicates whether the interaction can only happen once.
	SingleUse bool

	// Used tracks whether a single-use interaction has been triggered.
	Used bool

	// Sound is the sound effect to play on interaction.
	Sound string

	// Animation is the animation to trigger on the interactable entity.
	Animation string

	// TargetEntity is the entity ID affected by this interaction (for switches/levers).
	TargetEntity uint64

	// DialogID links to a dialog tree (for NPCs or readable objects).
	DialogID string

	// QuestTrigger is the quest ID triggered by this interaction.
	QuestTrigger string
}

// Type returns the component type identifier for Interactable.
func (i *Interactable) Type() string { return "Interactable" }

// CanInteract checks if the entity can be interacted with right now.
func (i *Interactable) CanInteract(currentTime float64) bool {
	if i.Used && i.SingleUse {
		return false
	}
	if currentTime-i.LastInteractTime < i.Cooldown {
		return false
	}
	return true
}

// TriggerInteraction marks the interaction as used and updates the timestamp.
func (i *Interactable) TriggerInteraction(currentTime float64) {
	i.LastInteractTime = currentTime
	if i.SingleUse {
		i.Used = true
	}
}

// NewSimpleInteractable creates a basic interactable with common defaults.
func NewSimpleInteractable(interactionType, prompt string, interactRange float64) *Interactable {
	return &Interactable{
		InteractionType: interactionType,
		Range:           interactRange,
		Prompt:          prompt,
		Cooldown:        0.5, // Default half-second cooldown
	}
}

// NewLockedInteractable creates an interactable that requires a key.
func NewLockedInteractable(interactionType, prompt, keyID string, interactRange float64) *Interactable {
	return &Interactable{
		InteractionType: interactionType,
		Range:           interactRange,
		Prompt:          prompt,
		Cooldown:        0.5,
		Locked:          true,
		KeyID:           keyID,
	}
}

// NewSkillCheckInteractable creates an interactable that requires a skill check.
func NewSkillCheckInteractable(interactionType, prompt, skill string, difficulty int, interactRange float64) *Interactable {
	return &Interactable{
		InteractionType: interactionType,
		Range:           interactRange,
		Prompt:          prompt,
		Cooldown:        0.5,
		RequiredSkill:   skill,
		SkillDifficulty: difficulty,
	}
}

// ============================================================
// WorldItem Component
// ============================================================

// WorldItem represents an item that exists in the world and can be picked up.
// This component is used for loot drops, resource nodes, and placed items.
type WorldItem struct {
	// ItemID is the unique identifier for the item type.
	ItemID string

	// Quantity is the number of items in this stack.
	Quantity int

	// SpawnTime is the game time when this item was spawned/dropped.
	SpawnTime float64

	// Respawnable indicates whether this item will respawn after being picked up.
	Respawnable bool

	// RespawnTime is the delay (in seconds) before respawning (if Respawnable).
	RespawnTime float64

	// LastPickupTime tracks when the item was last picked up (for respawn calculation).
	LastPickupTime float64

	// Quality is the item quality level (0 = common, 1 = uncommon, 2 = rare, etc.).
	Quality int

	// Durability is the current durability of the item (if applicable).
	Durability float64

	// MaxDurability is the maximum durability of the item.
	MaxDurability float64

	// Owner is the entity ID that owns this item (0 = no owner, anyone can pick up).
	Owner uint64

	// OwnerExpiry is the game time when ownership expires (others can then pick up).
	OwnerExpiry float64

	// Highlight indicates whether this item should be visually highlighted.
	Highlight bool

	// StackLimit is the maximum quantity per stack for this item type.
	StackLimit int

	// Value is the base gold/currency value of the item.
	Value int

	// Weight is the weight of the item (affects carry capacity).
	Weight float64

	// Category categorizes the item (weapon, armor, consumable, material, quest, etc.).
	Category string

	// Rarity affects drop chance and visual effects (common, uncommon, rare, epic, legendary).
	Rarity string

	// PickupSound is the sound effect to play when picked up.
	PickupSound string

	// LevelRequirement is the minimum player level to pick up/use this item.
	LevelRequirement int

	// BoundOnPickup indicates whether the item becomes bound to the player when picked up.
	BoundOnPickup bool
}

// Type returns the component type identifier for WorldItem.
func (w *WorldItem) Type() string { return "WorldItem" }

// CanPickup checks if the item can be picked up by the given entity at the current time.
func (w *WorldItem) CanPickup(entityID uint64, currentTime float64) bool {
	// Check if respawn is still pending
	if w.Respawnable && w.LastPickupTime > 0 {
		if currentTime-w.LastPickupTime < w.RespawnTime {
			return false
		}
	}
	// Check ownership
	if w.Owner != 0 && w.Owner != entityID {
		if currentTime < w.OwnerExpiry {
			return false
		}
	}
	return true
}

// Pickup marks the item as picked up.
func (w *WorldItem) Pickup(currentTime float64) {
	w.LastPickupTime = currentTime
}

// IsRespawned checks if a respawnable item has respawned.
func (w *WorldItem) IsRespawned(currentTime float64) bool {
	if !w.Respawnable {
		return false
	}
	if w.LastPickupTime == 0 {
		return true // Never picked up
	}
	return currentTime-w.LastPickupTime >= w.RespawnTime
}

// NewWorldItem creates a basic world item with common defaults.
func NewWorldItem(itemID string, quantity int, category string) *WorldItem {
	return &WorldItem{
		ItemID:     itemID,
		Quantity:   quantity,
		Category:   category,
		Quality:    0,
		StackLimit: 99,
		Rarity:     "common",
	}
}

// NewRespawnableItem creates a world item that will respawn after being picked up.
func NewRespawnableItem(itemID string, quantity int, respawnTime float64) *WorldItem {
	return &WorldItem{
		ItemID:      itemID,
		Quantity:    quantity,
		Respawnable: true,
		RespawnTime: respawnTime,
		StackLimit:  99,
		Rarity:      "common",
	}
}

// NewOwnedDrop creates a world item with temporary ownership (like loot drops).
func NewOwnedDrop(itemID string, quantity int, ownerID uint64, ownershipDuration, currentTime float64) *WorldItem {
	return &WorldItem{
		ItemID:      itemID,
		Quantity:    quantity,
		SpawnTime:   currentTime,
		Owner:       ownerID,
		OwnerExpiry: currentTime + ownershipDuration,
		StackLimit:  99,
		Rarity:      "common",
	}
}

// ============================================================
// Particle Component
// ============================================================

// Particle represents a visual particle for effects like debris, sparks, smoke, etc.
// Particles are typically short-lived entities managed by a particle system.
type Particle struct {
	// ParticleID identifies the particle type for rendering/pooling.
	ParticleID string

	// Color is the RGBA color of the particle.
	Color [4]uint8

	// Size is the current size of the particle (may change over lifetime).
	Size float64

	// InitialSize is the size at spawn (for scaling effects).
	InitialSize float64

	// Lifetime is the total duration this particle will exist (seconds).
	Lifetime float64

	// Age is how long this particle has existed (seconds).
	Age float64

	// FadeOut indicates whether the particle should fade as it ages.
	FadeOut bool

	// ShrinkOut indicates whether the particle should shrink as it ages.
	ShrinkOut bool

	// GrowOut indicates whether the particle should grow as it ages.
	GrowOut bool

	// RotationSpeed is the angular velocity in radians/second.
	RotationSpeed float64

	// Rotation is the current rotation angle in radians.
	Rotation float64

	// EmitterID links to the emitter that spawned this particle.
	EmitterID uint64

	// TextureKey is an optional texture to render instead of a simple shape.
	TextureKey string

	// BlendMode controls how the particle blends with the background.
	// 0 = normal, 1 = additive, 2 = multiply
	BlendMode int
}

// Type returns the component type identifier for Particle.
func (p *Particle) Type() string { return "Particle" }

// IsExpired returns true if the particle has exceeded its lifetime.
func (p *Particle) IsExpired() bool {
	return p.Age >= p.Lifetime
}

// LifetimeRatio returns a value from 0.0 (just spawned) to 1.0 (about to expire).
func (p *Particle) LifetimeRatio() float64 {
	if p.Lifetime <= 0 {
		return 1.0
	}
	ratio := p.Age / p.Lifetime
	if ratio > 1.0 {
		return 1.0
	}
	return ratio
}

// CurrentAlpha returns the current alpha based on fade settings.
func (p *Particle) CurrentAlpha() uint8 {
	if !p.FadeOut {
		return p.Color[3]
	}
	alpha := float64(p.Color[3]) * (1.0 - p.LifetimeRatio())
	if alpha < 0 {
		alpha = 0
	}
	return uint8(alpha)
}

// CurrentSize returns the current size based on shrink/grow settings.
func (p *Particle) CurrentSize() float64 {
	ratio := p.LifetimeRatio()
	if p.ShrinkOut {
		return p.InitialSize * (1.0 - ratio)
	}
	if p.GrowOut {
		return p.InitialSize * (1.0 + ratio)
	}
	return p.Size
}

// NewParticle creates a basic particle with common defaults.
func NewParticle(id string, color [4]uint8, size, lifetime float64) *Particle {
	return &Particle{
		ParticleID:  id,
		Color:       color,
		Size:        size,
		InitialSize: size,
		Lifetime:    lifetime,
		Age:         0,
		FadeOut:     true,
	}
}

// NewDebrisParticle creates a debris particle with physics-friendly defaults.
func NewDebrisParticle(color [4]uint8, size, lifetime float64) *Particle {
	return &Particle{
		ParticleID:  "debris",
		Color:       color,
		Size:        size,
		InitialSize: size,
		Lifetime:    lifetime,
		Age:         0,
		FadeOut:     true,
		ShrinkOut:   true,
	}
}

// ============================================================
// SoundEvent Component
// ============================================================

// SoundEvent represents a one-shot or continuous sound effect to be played.
// The audio system processes these and plays the appropriate sounds.
type SoundEvent struct {
	// SoundID is the identifier for the sound to play.
	SoundID string

	// X, Y, Z are the world position where the sound originates.
	X, Y, Z float64

	// Volume is the base volume (0.0 to 1.0).
	Volume float64

	// Radius is the distance at which the sound is no longer audible.
	Radius float64

	// Pitch is the pitch multiplier (1.0 = normal).
	Pitch float64

	// OneShot indicates the sound plays once then the event is removed.
	OneShot bool

	// Loop indicates the sound should loop until stopped.
	Loop bool

	// Processed indicates the audio system has handled this event.
	Processed bool

	// Priority affects which sounds play when too many are active.
	Priority int

	// OwnerEntity is the entity this sound is attached to (for moving sounds).
	OwnerEntity uint64

	// FollowOwner indicates the sound position should track the owner entity.
	FollowOwner bool

	// FadeInTime is the duration to fade in (seconds).
	FadeInTime float64

	// FadeOutTime is the duration to fade out when stopping (seconds).
	FadeOutTime float64
}

// Type returns the component type identifier for SoundEvent.
func (s *SoundEvent) Type() string { return "SoundEvent" }

// NewSoundEvent creates a basic one-shot sound event.
func NewSoundEvent(soundID string, x, y, z, volume, radius float64) *SoundEvent {
	return &SoundEvent{
		SoundID: soundID,
		X:       x,
		Y:       y,
		Z:       z,
		Volume:  volume,
		Radius:  radius,
		Pitch:   1.0,
		OneShot: true,
	}
}

// NewPositionalSound creates a sound attached to an entity.
func NewPositionalSound(soundID string, owner uint64, volume, radius float64) *SoundEvent {
	return &SoundEvent{
		SoundID:     soundID,
		Volume:      volume,
		Radius:      radius,
		Pitch:       1.0,
		OneShot:     true,
		OwnerEntity: owner,
		FollowOwner: true,
	}
}

// NewAmbientLoop creates a looping ambient sound at a position.
func NewAmbientLoop(soundID string, x, y, z, volume, radius float64) *SoundEvent {
	return &SoundEvent{
		SoundID: soundID,
		X:       x,
		Y:       y,
		Z:       z,
		Volume:  volume,
		Radius:  radius,
		Pitch:   1.0,
		Loop:    true,
	}
}
