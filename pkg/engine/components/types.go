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
