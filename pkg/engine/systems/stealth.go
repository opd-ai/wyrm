package systems

import (
	"math"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// StealthSystem manages stealth detection, sneaking, and awareness.
type StealthSystem struct {
	// BackstabMultiplier is the damage multiplier for attacking unaware targets.
	BackstabMultiplier float64
	// SneakSpeedReduction is the movement speed penalty when sneaking (0.5 = 50% slower).
	SneakSpeedReduction float64
	// AlertDecayRate is how fast alert levels decay per second.
	AlertDecayRate float64
	// GameTime tracks current game time for detection timing.
	GameTime float64
}

// NewStealthSystem creates a new stealth system with default settings.
func NewStealthSystem() *StealthSystem {
	return &StealthSystem{
		BackstabMultiplier:  BackstabDamageMultiplier,
		SneakSpeedReduction: DefaultSneakSpeedReduction,
		AlertDecayRate:      DefaultAlertDecayRate,
		GameTime:            0,
	}
}

// Update processes stealth detection and awareness each tick.
func (s *StealthSystem) Update(w *ecs.World, dt float64) {
	s.GameTime += dt
	s.updateStealthVisibility(w)
	s.updateDetection(w)
	s.decayAlertLevels(w, dt)
}

// updateStealthVisibility updates visibility for sneaking entities.
func (s *StealthSystem) updateStealthVisibility(w *ecs.World) {
	for _, e := range w.Entities("Stealth") {
		comp, ok := w.GetComponent(e, "Stealth")
		if !ok {
			continue
		}
		stealth := comp.(*components.Stealth)
		s.updateEntityVisibility(stealth)
	}
}

// updateEntityVisibility sets visibility based on sneak state.
func (s *StealthSystem) updateEntityVisibility(stealth *components.Stealth) {
	if stealth.Sneaking {
		stealth.Visibility = stealth.SneakVisibility
	} else {
		stealth.Visibility = stealth.BaseVisibility
	}
}

// updateDetection checks if NPCs detect sneaking entities.
func (s *StealthSystem) updateDetection(w *ecs.World) {
	// Get all NPCs with awareness
	for _, npc := range w.Entities("Awareness", "Position") {
		s.checkNPCDetection(w, npc)
	}
}

// checkNPCDetection checks if an NPC detects any stealthy entities.
func (s *StealthSystem) checkNPCDetection(w *ecs.World, npc ecs.Entity) {
	awarenessComp, ok := w.GetComponent(npc, "Awareness")
	if !ok {
		return
	}
	awareness := awarenessComp.(*components.Awareness)

	npcPos := s.getPosition(w, npc)
	if npcPos == nil {
		return
	}

	// Check all stealthy entities
	for _, target := range w.Entities("Stealth", "Position") {
		if target == npc {
			continue
		}
		s.checkDetection(w, npc, target, awareness, npcPos)
	}
}

// checkDetection determines if an NPC detects a specific target.
func (s *StealthSystem) checkDetection(w *ecs.World, npc, target ecs.Entity, awareness *components.Awareness, npcPos *components.Position) {
	stealthComp, ok := w.GetComponent(target, "Stealth")
	if !ok {
		return
	}
	stealth := stealthComp.(*components.Stealth)

	targetPos := s.getPosition(w, target)
	if targetPos == nil {
		return
	}

	// Check if in range and sight cone
	if s.canDetect(npcPos, targetPos, awareness, stealth) {
		s.applyDetection(awareness, stealth, target)
	}
}

// canDetect checks if an NPC can detect a stealthy target.
func (s *StealthSystem) canDetect(npcPos, targetPos *components.Position, awareness *components.Awareness, stealth *components.Stealth) bool {
	// Calculate distance
	dx := targetPos.X - npcPos.X
	dy := targetPos.Y - npcPos.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	// Check range (modified by target visibility)
	effectiveRange := awareness.SightRange * stealth.Visibility
	if dist > effectiveRange {
		return false
	}

	// Check sight cone
	angleToTarget := math.Atan2(dy, dx)
	angleDiff := normalizeAngle(angleToTarget - npcPos.Angle)
	halfFOV := awareness.SightAngle / 2

	return math.Abs(angleDiff) <= halfFOV
}

// normalizeAngle wraps an angle to [-PI, PI].
func normalizeAngle(angle float64) float64 {
	for angle > math.Pi {
		angle -= 2 * math.Pi
	}
	for angle < -math.Pi {
		angle += 2 * math.Pi
	}
	return angle
}

// applyDetection records that an NPC detected a target.
func (s *StealthSystem) applyDetection(awareness *components.Awareness, stealth *components.Stealth, target ecs.Entity) {
	if awareness.DetectedEntities == nil {
		awareness.DetectedEntities = make(map[uint64]float64)
	}
	if stealth.LastDetectedBy == nil {
		stealth.LastDetectedBy = make(map[uint64]float64)
	}

	// Increase alert level based on visibility
	awareness.AlertLevel += stealth.Visibility * AlertIncreasePerDetection
	if awareness.AlertLevel > MaxAlertLevel {
		awareness.AlertLevel = MaxAlertLevel
	}

	awareness.DetectedEntities[uint64(target)] = awareness.AlertLevel
	stealth.LastDetectedBy[uint64(target)] = s.GameTime
}

// decayAlertLevels reduces alert levels over time.
func (s *StealthSystem) decayAlertLevels(w *ecs.World, dt float64) {
	for _, e := range w.Entities("Awareness") {
		comp, ok := w.GetComponent(e, "Awareness")
		if !ok {
			continue
		}
		awareness := comp.(*components.Awareness)
		awareness.AlertLevel -= s.AlertDecayRate * dt
		if awareness.AlertLevel < 0 {
			awareness.AlertLevel = 0
		}
	}
}

// IsTargetUnaware checks if a target is unaware of the attacker (for backstab).
func (s *StealthSystem) IsTargetUnaware(w *ecs.World, attacker, target ecs.Entity) bool {
	// Check if target has awareness
	awarenessComp, ok := w.GetComponent(target, "Awareness")
	if !ok {
		return true // No awareness = always unaware
	}
	awareness := awarenessComp.(*components.Awareness)

	// Check if target has detected attacker
	if awareness.DetectedEntities == nil {
		return true
	}
	alertLevel, detected := awareness.DetectedEntities[uint64(attacker)]
	return !detected || alertLevel < AwarenessThreshold // Below threshold = unaware
}

// GetBackstabDamage calculates damage with backstab multiplier if applicable.
func (s *StealthSystem) GetBackstabDamage(w *ecs.World, baseDamage float64, attacker, target ecs.Entity) float64 {
	if s.IsTargetUnaware(w, attacker, target) {
		return baseDamage * s.BackstabMultiplier
	}
	return baseDamage
}

// SetSneaking sets an entity's sneak state.
func (s *StealthSystem) SetSneaking(w *ecs.World, entity ecs.Entity, sneaking bool) bool {
	stealthComp, ok := w.GetComponent(entity, "Stealth")
	if !ok {
		return false
	}
	stealth := stealthComp.(*components.Stealth)
	stealth.Sneaking = sneaking
	return true
}

// CanPickpocket checks if a pickpocket attempt can succeed.
func (s *StealthSystem) CanPickpocket(w *ecs.World, thief, target ecs.Entity) bool {
	// Check thief is sneaking
	stealthComp, ok := w.GetComponent(thief, "Stealth")
	if !ok || !stealthComp.(*components.Stealth).Sneaking {
		return false
	}

	// Check target is unaware
	return s.IsTargetUnaware(w, thief, target)
}

// AttemptPickpocket performs a pickpocket with skill check.
func (s *StealthSystem) AttemptPickpocket(w *ecs.World, thief, target ecs.Entity, difficulty float64) bool {
	if !s.CanPickpocket(w, thief, target) {
		return false
	}

	// Get thief's pickpocket skill
	skillLevel := s.getPickpocketSkill(w, thief)

	// Skill check: skill level >= difficulty
	return float64(skillLevel) >= difficulty*PickpocketDifficultyMultiplier
}

// getPickpocketSkill returns the pickpocket skill level for an entity.
func (s *StealthSystem) getPickpocketSkill(w *ecs.World, entity ecs.Entity) int {
	skillsComp, ok := w.GetComponent(entity, "Skills")
	if !ok {
		return 0
	}
	skills := skillsComp.(*components.Skills)
	if skills.Levels == nil {
		return 0
	}

	// Check for pickpocket-related skills
	for _, skillID := range []string{"pickpocket", "stealth", "thievery", "sneak"} {
		if level, exists := skills.Levels[skillID]; exists {
			return level
		}
	}
	return 0
}

// getPosition retrieves an entity's position component.
func (s *StealthSystem) getPosition(w *ecs.World, entity ecs.Entity) *components.Position {
	posComp, ok := w.GetComponent(entity, "Position")
	if !ok {
		return nil
	}
	return posComp.(*components.Position)
}

// ============================================================================
// Hiding Spot System
// ============================================================================

// HidingSpotType categorizes different types of hiding locations.
type HidingSpotType string

const (
	// HidingSpotShadow represents dark areas with low visibility.
	HidingSpotShadow HidingSpotType = "shadow"
	// HidingSpotFoliage represents bushes, plants, and vegetation.
	HidingSpotFoliage HidingSpotType = "foliage"
	// HidingSpotContainer represents barrels, crates, and large containers.
	HidingSpotContainer HidingSpotType = "container"
	// HidingSpotFurniture represents beds, curtains, and other furniture.
	HidingSpotFurniture HidingSpotType = "furniture"
	// HidingSpotArchitectural represents alcoves, pillars, and structural cover.
	HidingSpotArchitectural HidingSpotType = "architectural"
	// HidingSpotRooftop represents elevated hiding positions.
	HidingSpotRooftop HidingSpotType = "rooftop"
	// HidingSpotUnderwater represents submerged positions.
	HidingSpotUnderwater HidingSpotType = "underwater"
)

// HidingSpot defines a location where entities can hide.
type HidingSpot struct {
	// ID uniquely identifies this hiding spot.
	ID string
	// Type categorizes the hiding spot.
	Type HidingSpotType
	// Position is the world position of the hiding spot.
	Position *components.Position
	// Radius is the area of effect for the hiding spot.
	Radius float64
	// VisibilityReduction reduces entity visibility while hiding (0.0-1.0).
	VisibilityReduction float64
	// DetectionMultiplier modifies how easily entities are detected (lower = harder to detect).
	DetectionMultiplier float64
	// Capacity is the maximum number of entities that can hide here.
	Capacity int
	// Occupants tracks entities currently using this spot.
	Occupants []ecs.Entity
	// RequiredSkillLevel is the minimum sneak skill to use this spot.
	RequiredSkillLevel int
	// CanBeSearched indicates if NPCs can search this location.
	CanBeSearched bool
	// SearchDifficulty is how hard it is to find someone here (0.0-1.0).
	SearchDifficulty float64
	// IsTemporary indicates if the spot disappears after use.
	IsTemporary bool
	// Genre restricts this spot to specific genres ("" for all).
	Genre string
}

// HidingSpotSystem manages hiding spots and their effects on stealth.
type HidingSpotSystem struct {
	// HidingSpots is the registry of all active hiding spots.
	HidingSpots map[string]*HidingSpot
	// EntityHiding tracks which entity is hiding in which spot (entityID -> spotID).
	EntityHiding map[uint64]string
	// SpotsByChunk organizes spots by chunk coordinates for efficient lookup.
	SpotsByChunk map[int64]map[int64][]*HidingSpot
	// ChunkSize is the world chunk size for spatial organization.
	ChunkSize float64
	// DefaultCapacity is the default capacity for hiding spots.
	DefaultCapacity int
}

// NewHidingSpotSystem creates a new hiding spot system.
func NewHidingSpotSystem(chunkSize float64) *HidingSpotSystem {
	if chunkSize <= 0 {
		chunkSize = 512
	}
	return &HidingSpotSystem{
		HidingSpots:     make(map[string]*HidingSpot),
		EntityHiding:    make(map[uint64]string),
		SpotsByChunk:    make(map[int64]map[int64][]*HidingSpot),
		ChunkSize:       chunkSize,
		DefaultCapacity: 1,
	}
}

// RegisterHidingSpot adds a hiding spot to the system.
func (s *HidingSpotSystem) RegisterHidingSpot(spot *HidingSpot) {
	if spot == nil || spot.ID == "" {
		return
	}
	if spot.Capacity <= 0 {
		spot.Capacity = s.DefaultCapacity
	}
	if spot.Occupants == nil {
		spot.Occupants = make([]ecs.Entity, 0)
	}

	s.HidingSpots[spot.ID] = spot
	s.indexByChunk(spot)
}

// indexByChunk adds a spot to the chunk-based spatial index.
func (s *HidingSpotSystem) indexByChunk(spot *HidingSpot) {
	if spot.Position == nil {
		return
	}
	cx := int64(spot.Position.X / s.ChunkSize)
	cy := int64(spot.Position.Y / s.ChunkSize)

	if s.SpotsByChunk[cx] == nil {
		s.SpotsByChunk[cx] = make(map[int64][]*HidingSpot)
	}
	if s.SpotsByChunk[cx][cy] == nil {
		s.SpotsByChunk[cx][cy] = make([]*HidingSpot, 0)
	}
	s.SpotsByChunk[cx][cy] = append(s.SpotsByChunk[cx][cy], spot)
}

// GetHidingSpot retrieves a hiding spot by ID.
func (s *HidingSpotSystem) GetHidingSpot(spotID string) *HidingSpot {
	return s.HidingSpots[spotID]
}

// GetSpotsNear returns hiding spots within range of a position.
func (s *HidingSpotSystem) GetSpotsNear(x, y, radius float64) []*HidingSpot {
	spots := make([]*HidingSpot, 0)

	// Check nearby chunks
	cx := int64(x / s.ChunkSize)
	cy := int64(y / s.ChunkSize)

	for dx := int64(-1); dx <= 1; dx++ {
		for dy := int64(-1); dy <= 1; dy++ {
			chunkSpots := s.SpotsByChunk[cx+dx][cy+dy]
			for _, spot := range chunkSpots {
				if spot.Position == nil {
					continue
				}
				dist := math.Sqrt(
					math.Pow(spot.Position.X-x, 2) +
						math.Pow(spot.Position.Y-y, 2))
				if dist <= radius+spot.Radius {
					spots = append(spots, spot)
				}
			}
		}
	}
	return spots
}

// GetSpotsByType returns all hiding spots of a specific type.
func (s *HidingSpotSystem) GetSpotsByType(spotType HidingSpotType) []*HidingSpot {
	spots := make([]*HidingSpot, 0)
	for _, spot := range s.HidingSpots {
		if spot.Type == spotType {
			spots = append(spots, spot)
		}
	}
	return spots
}

// Update processes hiding spot effects each tick.
func (s *HidingSpotSystem) Update(w *ecs.World, dt float64) {
	s.updateHidingEffects(w)
	s.cleanupTemporarySpots()
}

// updateHidingEffects applies visibility bonuses to hiding entities.
func (s *HidingSpotSystem) updateHidingEffects(w *ecs.World) {
	for entityID, spotID := range s.EntityHiding {
		spot := s.HidingSpots[spotID]
		if spot == nil {
			continue
		}

		entity := ecs.Entity(entityID)
		stealthComp, ok := w.GetComponent(entity, "Stealth")
		if !ok {
			continue
		}
		stealth := stealthComp.(*components.Stealth)

		// Apply hiding spot visibility reduction
		if stealth.Sneaking {
			stealth.Visibility = stealth.SneakVisibility * (1.0 - spot.VisibilityReduction)
		}
	}
}

// cleanupTemporarySpots removes temporary spots that have been used.
func (s *HidingSpotSystem) cleanupTemporarySpots() {
	for id, spot := range s.HidingSpots {
		if spot.IsTemporary && len(spot.Occupants) == 0 {
			// Check if recently vacated (would need timestamp for full impl)
			delete(s.HidingSpots, id)
		}
	}
}

// CanHide checks if an entity can hide in a specific spot.
func (s *HidingSpotSystem) CanHide(w *ecs.World, entity ecs.Entity, spotID string) bool {
	spot := s.HidingSpots[spotID]
	if spot == nil {
		return false
	}

	// Check capacity
	if len(spot.Occupants) >= spot.Capacity {
		return false
	}

	// Check if already hiding elsewhere
	entityID := uint64(entity)
	if _, hiding := s.EntityHiding[entityID]; hiding {
		return false
	}

	// Check skill requirement
	if spot.RequiredSkillLevel > 0 {
		skillsComp, ok := w.GetComponent(entity, "Skills")
		if !ok {
			return false
		}
		skills := skillsComp.(*components.Skills)
		if skills.Levels == nil || skills.Levels["sneak"] < spot.RequiredSkillLevel {
			return false
		}
	}

	// Check position proximity
	posComp, ok := w.GetComponent(entity, "Position")
	if !ok {
		return false
	}
	pos := posComp.(*components.Position)

	if spot.Position != nil {
		dist := math.Sqrt(
			math.Pow(pos.X-spot.Position.X, 2) +
				math.Pow(pos.Y-spot.Position.Y, 2))
		if dist > spot.Radius*2 {
			return false // Too far away
		}
	}

	return true
}

// EnterHidingSpot puts an entity into a hiding spot.
func (s *HidingSpotSystem) EnterHidingSpot(w *ecs.World, entity ecs.Entity, spotID string) bool {
	if !s.CanHide(w, entity, spotID) {
		return false
	}

	spot := s.HidingSpots[spotID]
	entityID := uint64(entity)

	spot.Occupants = append(spot.Occupants, entity)
	s.EntityHiding[entityID] = spotID

	// Auto-enable sneaking
	stealthComp, ok := w.GetComponent(entity, "Stealth")
	if ok {
		stealth := stealthComp.(*components.Stealth)
		stealth.Sneaking = true
	}

	return true
}

// ExitHidingSpot removes an entity from their hiding spot.
func (s *HidingSpotSystem) ExitHidingSpot(w *ecs.World, entity ecs.Entity) bool {
	entityID := uint64(entity)
	spotID, hiding := s.EntityHiding[entityID]
	if !hiding {
		return false
	}

	spot := s.HidingSpots[spotID]
	if spot != nil {
		// Remove from occupants
		newOccupants := make([]ecs.Entity, 0)
		for _, occupant := range spot.Occupants {
			if occupant != entity {
				newOccupants = append(newOccupants, occupant)
			}
		}
		spot.Occupants = newOccupants
	}

	delete(s.EntityHiding, entityID)
	return true
}

// IsHiding checks if an entity is currently hiding.
func (s *HidingSpotSystem) IsHiding(entity ecs.Entity) bool {
	entityID := uint64(entity)
	_, hiding := s.EntityHiding[entityID]
	return hiding
}

// GetCurrentSpot returns the hiding spot an entity is using (nil if not hiding).
func (s *HidingSpotSystem) GetCurrentSpot(entity ecs.Entity) *HidingSpot {
	entityID := uint64(entity)
	spotID, hiding := s.EntityHiding[entityID]
	if !hiding {
		return nil
	}
	return s.HidingSpots[spotID]
}

// SearchHidingSpot attempts to find hidden entities in a spot.
func (s *HidingSpotSystem) SearchHidingSpot(w *ecs.World, searcher ecs.Entity, spotID string) []ecs.Entity {
	spot := s.HidingSpots[spotID]
	if spot == nil || !spot.CanBeSearched {
		return nil
	}

	// Get searcher's perception skill
	searcherSkill := 0
	skillsComp, ok := w.GetComponent(searcher, "Skills")
	if ok {
		skills := skillsComp.(*components.Skills)
		if skills.Levels != nil {
			searcherSkill = skills.Levels["perception"]
		}
	}

	found := make([]ecs.Entity, 0)
	for _, occupant := range spot.Occupants {
		// Get hider's sneak skill
		hiderSkill := 0
		hiderSkillsComp, ok := w.GetComponent(occupant, "Skills")
		if ok {
			hiderSkills := hiderSkillsComp.(*components.Skills)
			if hiderSkills.Levels != nil {
				hiderSkill = hiderSkills.Levels["sneak"]
			}
		}

		// Check if found: searcher perception vs hider sneak * search difficulty
		effectiveHiding := float64(hiderSkill) * spot.SearchDifficulty
		if float64(searcherSkill) > effectiveHiding {
			found = append(found, occupant)
			// Reveal found entities
			s.ExitHidingSpot(w, occupant)
		}
	}
	return found
}

// GetNearestAvailableSpot finds the closest usable hiding spot.
func (s *HidingSpotSystem) GetNearestAvailableSpot(w *ecs.World, entity ecs.Entity) *HidingSpot {
	posComp, ok := w.GetComponent(entity, "Position")
	if !ok {
		return nil
	}
	pos := posComp.(*components.Position)

	var nearest *HidingSpot
	nearestDist := math.MaxFloat64

	for spotID, spot := range s.HidingSpots {
		if !s.CanHide(w, entity, spotID) {
			continue
		}
		if spot.Position == nil {
			continue
		}

		dist := math.Sqrt(
			math.Pow(pos.X-spot.Position.X, 2) +
				math.Pow(pos.Y-spot.Position.Y, 2))

		if dist < nearestDist {
			nearestDist = dist
			nearest = spot
		}
	}
	return nearest
}

// CreateHidingSpotFromEnvironment generates a hiding spot at a position.
func (s *HidingSpotSystem) CreateHidingSpotFromEnvironment(id string, spotType HidingSpotType, x, y, z float64) *HidingSpot {
	spot := &HidingSpot{
		ID:   id,
		Type: spotType,
		Position: &components.Position{
			X: x, Y: y, Z: z,
		},
		Occupants: make([]ecs.Entity, 0),
	}

	// Set type-specific defaults
	switch spotType {
	case HidingSpotShadow:
		spot.Radius = 3.0
		spot.VisibilityReduction = 0.6
		spot.DetectionMultiplier = 0.4
		spot.Capacity = 2
		spot.CanBeSearched = false
		spot.SearchDifficulty = 0.9
	case HidingSpotFoliage:
		spot.Radius = 4.0
		spot.VisibilityReduction = 0.7
		spot.DetectionMultiplier = 0.3
		spot.Capacity = 3
		spot.CanBeSearched = true
		spot.SearchDifficulty = 0.6
	case HidingSpotContainer:
		spot.Radius = 1.5
		spot.VisibilityReduction = 0.95
		spot.DetectionMultiplier = 0.1
		spot.Capacity = 1
		spot.CanBeSearched = true
		spot.SearchDifficulty = 0.3
		spot.RequiredSkillLevel = 5
	case HidingSpotFurniture:
		spot.Radius = 2.0
		spot.VisibilityReduction = 0.8
		spot.DetectionMultiplier = 0.2
		spot.Capacity = 1
		spot.CanBeSearched = true
		spot.SearchDifficulty = 0.4
	case HidingSpotArchitectural:
		spot.Radius = 2.5
		spot.VisibilityReduction = 0.7
		spot.DetectionMultiplier = 0.35
		spot.Capacity = 2
		spot.CanBeSearched = false
		spot.SearchDifficulty = 0.7
	case HidingSpotRooftop:
		spot.Radius = 5.0
		spot.VisibilityReduction = 0.5
		spot.DetectionMultiplier = 0.6
		spot.Capacity = 2
		spot.CanBeSearched = false
		spot.SearchDifficulty = 0.8
		spot.RequiredSkillLevel = 20
	case HidingSpotUnderwater:
		spot.Radius = 4.0
		spot.VisibilityReduction = 0.9
		spot.DetectionMultiplier = 0.15
		spot.Capacity = 1
		spot.CanBeSearched = false
		spot.SearchDifficulty = 0.95
		spot.RequiredSkillLevel = 10
	default:
		spot.Radius = 2.0
		spot.VisibilityReduction = 0.5
		spot.DetectionMultiplier = 0.5
		spot.Capacity = 1
		spot.CanBeSearched = true
		spot.SearchDifficulty = 0.5
	}

	return spot
}

// SpotCount returns the total number of registered hiding spots.
func (s *HidingSpotSystem) SpotCount() int {
	return len(s.HidingSpots)
}

// GetOccupiedSpots returns all spots that have occupants.
func (s *HidingSpotSystem) GetOccupiedSpots() []*HidingSpot {
	spots := make([]*HidingSpot, 0)
	for _, spot := range s.HidingSpots {
		if len(spot.Occupants) > 0 {
			spots = append(spots, spot)
		}
	}
	return spots
}

// GetAvailableSpots returns all spots with available capacity.
func (s *HidingSpotSystem) GetAvailableSpots() []*HidingSpot {
	spots := make([]*HidingSpot, 0)
	for _, spot := range s.HidingSpots {
		if len(spot.Occupants) < spot.Capacity {
			spots = append(spots, spot)
		}
	}
	return spots
}

// ============================================================================
// Distraction System
// ============================================================================

// DistractionType categorizes different kinds of distractions.
type DistractionType string

const (
	// DistractionSound represents noise-based distractions (thrown objects, whistles).
	DistractionSound DistractionType = "sound"
	// DistractionVisual represents visible distractions (light, movement).
	DistractionVisual DistractionType = "visual"
	// DistractionTactical represents strategic distractions (fires, explosions).
	DistractionTactical DistractionType = "tactical"
	// DistractionSocial represents NPC-based distractions (conversations, arguments).
	DistractionSocial DistractionType = "social"
	// DistractionEnvironmental represents world-based distractions (weather, animals).
	DistractionEnvironmental DistractionType = "environmental"
)

// Distraction represents an event that draws NPC attention.
type Distraction struct {
	// ID uniquely identifies this distraction.
	ID string
	// Type categorizes the distraction.
	Type DistractionType
	// Position is where the distraction occurs.
	Position *components.Position
	// Radius is the area of effect for detecting the distraction.
	Radius float64
	// Intensity affects how strongly NPCs are drawn to investigate (0.0-1.0).
	Intensity float64
	// Duration is how long the distraction lasts (seconds).
	Duration float64
	// TimeRemaining tracks remaining active time.
	TimeRemaining float64
	// Creator is the entity that created the distraction (nil if environmental).
	Creator ecs.Entity
	// Priority determines which distraction NPCs prefer (higher = more important).
	Priority int
	// CanBeIgnored indicates if high-alert NPCs might ignore this.
	CanBeIgnored bool
	// InvestigationTime is how long NPCs spend investigating (seconds).
	InvestigationTime float64
	// IsRepeating indicates if the distraction continues making noise.
	IsRepeating bool
	// RepeatInterval is time between sounds/flashes for repeating distractions.
	RepeatInterval float64
	// Genre restricts this distraction to specific genres ("" for all).
	Genre string
}

// DistractionSystem manages distractions and NPC reactions.
type DistractionSystem struct {
	// ActiveDistractions are currently active distractions.
	ActiveDistractions map[string]*Distraction
	// NPCInvestigating tracks which NPC is investigating which distraction.
	NPCInvestigating map[uint64]string
	// InvestigationProgress tracks how long an NPC has been investigating.
	InvestigationProgress map[uint64]float64
	// DistractionTemplates are predefined distraction types.
	DistractionTemplates map[string]*Distraction
	// NextID is used for generating unique IDs.
	NextID int
}

// NewDistractionSystem creates a new distraction system.
func NewDistractionSystem() *DistractionSystem {
	s := &DistractionSystem{
		ActiveDistractions:    make(map[string]*Distraction),
		NPCInvestigating:      make(map[uint64]string),
		InvestigationProgress: make(map[uint64]float64),
		DistractionTemplates:  make(map[string]*Distraction),
		NextID:                1,
	}
	s.registerDefaultTemplates()
	return s
}

// registerDefaultTemplates adds standard distraction types.
func (s *DistractionSystem) registerDefaultTemplates() {
	// Sound distractions
	s.RegisterTemplate(&Distraction{
		ID: "thrown_rock", Type: DistractionSound,
		Radius: 15, Intensity: 0.5, Duration: 3, Priority: 1,
		CanBeIgnored: true, InvestigationTime: 5,
	})
	s.RegisterTemplate(&Distraction{
		ID: "whistle", Type: DistractionSound,
		Radius: 25, Intensity: 0.7, Duration: 2, Priority: 2,
		CanBeIgnored: false, InvestigationTime: 8,
	})
	s.RegisterTemplate(&Distraction{
		ID: "broken_glass", Type: DistractionSound,
		Radius: 20, Intensity: 0.8, Duration: 1, Priority: 3,
		CanBeIgnored: false, InvestigationTime: 10,
	})
	s.RegisterTemplate(&Distraction{
		ID: "animal_call", Type: DistractionSound,
		Radius: 30, Intensity: 0.3, Duration: 5, Priority: 1,
		CanBeIgnored: true, InvestigationTime: 3, IsRepeating: true, RepeatInterval: 2,
	})
	s.RegisterTemplate(&Distraction{
		ID: "explosion", Type: DistractionSound,
		Radius: 50, Intensity: 1.0, Duration: 1, Priority: 5,
		CanBeIgnored: false, InvestigationTime: 15,
	})

	// Visual distractions
	s.RegisterTemplate(&Distraction{
		ID: "torch_flicker", Type: DistractionVisual,
		Radius: 10, Intensity: 0.3, Duration: 4, Priority: 1,
		CanBeIgnored: true, InvestigationTime: 2, IsRepeating: true, RepeatInterval: 1,
	})
	s.RegisterTemplate(&Distraction{
		ID: "smoke_signal", Type: DistractionVisual,
		Radius: 40, Intensity: 0.6, Duration: 30, Priority: 2,
		CanBeIgnored: true, InvestigationTime: 10,
	})
	s.RegisterTemplate(&Distraction{
		ID: "bright_flash", Type: DistractionVisual,
		Radius: 25, Intensity: 0.9, Duration: 0.5, Priority: 4,
		CanBeIgnored: false, InvestigationTime: 5,
	})

	// Tactical distractions
	s.RegisterTemplate(&Distraction{
		ID: "small_fire", Type: DistractionTactical,
		Radius: 15, Intensity: 0.8, Duration: 60, Priority: 4,
		CanBeIgnored: false, InvestigationTime: 20,
	})
	s.RegisterTemplate(&Distraction{
		ID: "alarm_bell", Type: DistractionTactical,
		Radius: 100, Intensity: 1.0, Duration: 30, Priority: 5,
		CanBeIgnored: false, InvestigationTime: 30, IsRepeating: true, RepeatInterval: 2,
	})
	s.RegisterTemplate(&Distraction{
		ID: "trap_triggered", Type: DistractionTactical,
		Radius: 20, Intensity: 0.7, Duration: 5, Priority: 3,
		CanBeIgnored: false, InvestigationTime: 15,
	})

	// Social distractions
	s.RegisterTemplate(&Distraction{
		ID: "loud_argument", Type: DistractionSocial,
		Radius: 20, Intensity: 0.6, Duration: 20, Priority: 2,
		CanBeIgnored: true, InvestigationTime: 10,
	})
	s.RegisterTemplate(&Distraction{
		ID: "bard_performance", Type: DistractionSocial,
		Radius: 30, Intensity: 0.4, Duration: 120, Priority: 1,
		CanBeIgnored: true, InvestigationTime: 60, IsRepeating: true, RepeatInterval: 10,
	})
	s.RegisterTemplate(&Distraction{
		ID: "scream", Type: DistractionSocial,
		Radius: 40, Intensity: 0.9, Duration: 2, Priority: 4,
		CanBeIgnored: false, InvestigationTime: 15,
	})

	// Environmental distractions
	s.RegisterTemplate(&Distraction{
		ID: "thunder", Type: DistractionEnvironmental,
		Radius: 100, Intensity: 0.2, Duration: 1, Priority: 1,
		CanBeIgnored: true, InvestigationTime: 0,
	})
	s.RegisterTemplate(&Distraction{
		ID: "falling_debris", Type: DistractionEnvironmental,
		Radius: 15, Intensity: 0.5, Duration: 2, Priority: 2,
		CanBeIgnored: true, InvestigationTime: 5,
	})
	s.RegisterTemplate(&Distraction{
		ID: "animal_panic", Type: DistractionEnvironmental,
		Radius: 25, Intensity: 0.4, Duration: 10, Priority: 2,
		CanBeIgnored: true, InvestigationTime: 8,
	})
}

// RegisterTemplate adds a distraction template.
func (s *DistractionSystem) RegisterTemplate(template *Distraction) {
	if template == nil || template.ID == "" {
		return
	}
	s.DistractionTemplates[template.ID] = template
}

// CreateDistraction spawns a new distraction at a position.
func (s *DistractionSystem) CreateDistraction(templateID string, x, y, z float64, creator ecs.Entity) *Distraction {
	template := s.DistractionTemplates[templateID]
	if template == nil {
		return nil
	}

	id := s.generateID()
	distraction := &Distraction{
		ID:                id,
		Type:              template.Type,
		Position:          &components.Position{X: x, Y: y, Z: z},
		Radius:            template.Radius,
		Intensity:         template.Intensity,
		Duration:          template.Duration,
		TimeRemaining:     template.Duration,
		Creator:           creator,
		Priority:          template.Priority,
		CanBeIgnored:      template.CanBeIgnored,
		InvestigationTime: template.InvestigationTime,
		IsRepeating:       template.IsRepeating,
		RepeatInterval:    template.RepeatInterval,
		Genre:             template.Genre,
	}

	s.ActiveDistractions[id] = distraction
	return distraction
}

// generateID creates a unique distraction ID.
func (s *DistractionSystem) generateID() string {
	id := s.NextID
	s.NextID++
	return "distraction_" + string(rune('0'+id%10)) + string(rune('0'+id/10%10)) + string(rune('0'+id/100%10))
}

// Update processes distractions and NPC reactions.
func (s *DistractionSystem) Update(w *ecs.World, dt float64) {
	s.updateDistractionTimers(dt)
	s.updateNPCReactions(w, dt)
	s.cleanupExpiredDistractions()
}

// updateDistractionTimers decrements time remaining on distractions.
func (s *DistractionSystem) updateDistractionTimers(dt float64) {
	for _, distraction := range s.ActiveDistractions {
		distraction.TimeRemaining -= dt
	}
}

// updateNPCReactions processes NPC awareness of and reactions to distractions.
func (s *DistractionSystem) updateNPCReactions(w *ecs.World, dt float64) {
	for _, npc := range w.Entities("Awareness", "Position") {
		s.processNPCDistraction(w, npc, dt)
	}
}

// processNPCDistraction handles a single NPC's distraction response.
func (s *DistractionSystem) processNPCDistraction(w *ecs.World, npc ecs.Entity, dt float64) {
	entityID := uint64(npc)

	// Check if already investigating
	if currentID, investigating := s.NPCInvestigating[entityID]; investigating {
		distraction := s.ActiveDistractions[currentID]
		if distraction == nil || distraction.TimeRemaining <= 0 {
			// Distraction ended, stop investigating
			delete(s.NPCInvestigating, entityID)
			delete(s.InvestigationProgress, entityID)
		} else {
			// Continue investigating
			s.InvestigationProgress[entityID] += dt
			if s.InvestigationProgress[entityID] >= distraction.InvestigationTime {
				// Investigation complete
				delete(s.NPCInvestigating, entityID)
				delete(s.InvestigationProgress, entityID)
			}
		}
		return
	}

	// Find best distraction to investigate
	bestDistraction := s.findBestDistraction(w, npc)
	if bestDistraction != nil {
		s.NPCInvestigating[entityID] = bestDistraction.ID
		s.InvestigationProgress[entityID] = 0
	}
}

// findBestDistraction selects the highest priority distraction for an NPC.
func (s *DistractionSystem) findBestDistraction(w *ecs.World, npc ecs.Entity) *Distraction {
	posComp, ok := w.GetComponent(npc, "Position")
	if !ok {
		return nil
	}
	pos := posComp.(*components.Position)

	awarenessComp, ok := w.GetComponent(npc, "Awareness")
	if !ok {
		return nil
	}
	awareness := awarenessComp.(*components.Awareness)

	var best *Distraction
	bestScore := 0.0

	for _, distraction := range s.ActiveDistractions {
		if distraction.TimeRemaining <= 0 {
			continue
		}
		if distraction.Position == nil {
			continue
		}

		// Check if high-alert NPC might ignore low-priority distractions
		if distraction.CanBeIgnored && awareness.AlertLevel > AwarenessThreshold {
			continue
		}

		// Check range
		dist := math.Sqrt(
			math.Pow(distraction.Position.X-pos.X, 2) +
				math.Pow(distraction.Position.Y-pos.Y, 2))
		if dist > distraction.Radius {
			continue
		}

		// Calculate score based on priority, intensity, and proximity
		proximityScore := 1.0 - (dist / distraction.Radius)
		score := float64(distraction.Priority) * distraction.Intensity * proximityScore

		if score > bestScore {
			bestScore = score
			best = distraction
		}
	}

	return best
}

// cleanupExpiredDistractions removes distractions that have ended.
func (s *DistractionSystem) cleanupExpiredDistractions() {
	for id, distraction := range s.ActiveDistractions {
		if distraction.TimeRemaining <= 0 {
			delete(s.ActiveDistractions, id)
		}
	}
}

// IsNPCDistracted checks if an NPC is currently investigating a distraction.
func (s *DistractionSystem) IsNPCDistracted(npc ecs.Entity) bool {
	entityID := uint64(npc)
	_, distracted := s.NPCInvestigating[entityID]
	return distracted
}

// GetNPCDistraction returns the distraction an NPC is investigating.
func (s *DistractionSystem) GetNPCDistraction(npc ecs.Entity) *Distraction {
	entityID := uint64(npc)
	distractionID, ok := s.NPCInvestigating[entityID]
	if !ok {
		return nil
	}
	return s.ActiveDistractions[distractionID]
}

// GetInvestigationProgress returns how far an NPC is through investigating (0.0-1.0).
func (s *DistractionSystem) GetInvestigationProgress(npc ecs.Entity) float64 {
	entityID := uint64(npc)
	distractionID, ok := s.NPCInvestigating[entityID]
	if !ok {
		return 0
	}
	distraction := s.ActiveDistractions[distractionID]
	if distraction == nil || distraction.InvestigationTime <= 0 {
		return 0
	}
	progress := s.InvestigationProgress[entityID]
	return progress / distraction.InvestigationTime
}

// CancelDistraction removes an active distraction.
func (s *DistractionSystem) CancelDistraction(distractionID string) {
	delete(s.ActiveDistractions, distractionID)
	// Clear any NPCs investigating this distraction
	for entityID, id := range s.NPCInvestigating {
		if id == distractionID {
			delete(s.NPCInvestigating, entityID)
			delete(s.InvestigationProgress, entityID)
		}
	}
}

// GetActiveDistractions returns all currently active distractions.
func (s *DistractionSystem) GetActiveDistractions() []*Distraction {
	distractions := make([]*Distraction, 0)
	for _, d := range s.ActiveDistractions {
		if d.TimeRemaining > 0 {
			distractions = append(distractions, d)
		}
	}
	return distractions
}

// GetDistractionsNear returns active distractions within range of a position.
func (s *DistractionSystem) GetDistractionsNear(x, y, radius float64) []*Distraction {
	distractions := make([]*Distraction, 0)
	for _, d := range s.ActiveDistractions {
		if d.TimeRemaining <= 0 || d.Position == nil {
			continue
		}
		dist := math.Sqrt(math.Pow(d.Position.X-x, 2) + math.Pow(d.Position.Y-y, 2))
		if dist <= radius+d.Radius {
			distractions = append(distractions, d)
		}
	}
	return distractions
}

// GetDistractionsByType returns active distractions of a specific type.
func (s *DistractionSystem) GetDistractionsByType(distractionType DistractionType) []*Distraction {
	distractions := make([]*Distraction, 0)
	for _, d := range s.ActiveDistractions {
		if d.Type == distractionType && d.TimeRemaining > 0 {
			distractions = append(distractions, d)
		}
	}
	return distractions
}

// GetTemplateCount returns the number of registered templates.
func (s *DistractionSystem) GetTemplateCount() int {
	return len(s.DistractionTemplates)
}

// GetActiveCount returns the number of active distractions.
func (s *DistractionSystem) GetActiveCount() int {
	count := 0
	for _, d := range s.ActiveDistractions {
		if d.TimeRemaining > 0 {
			count++
		}
	}
	return count
}

// ThrowDistraction creates a thrown-object distraction (commonly used for stealth).
func (s *DistractionSystem) ThrowDistraction(w *ecs.World, thrower ecs.Entity, targetX, targetY, targetZ float64) *Distraction {
	return s.CreateDistraction("thrown_rock", targetX, targetY, targetZ, thrower)
}

// CreateWhistle creates a whistle distraction.
func (s *DistractionSystem) CreateWhistle(w *ecs.World, creator ecs.Entity, x, y, z float64) *Distraction {
	return s.CreateDistraction("whistle", x, y, z, creator)
}

// TriggerAlarm creates an alarm distraction (high priority, large radius).
func (s *DistractionSystem) TriggerAlarm(w *ecs.World, x, y, z float64) *Distraction {
	return s.CreateDistraction("alarm_bell", x, y, z, 0)
}

// CreateFire creates a fire distraction that lasts longer.
func (s *DistractionSystem) CreateFire(w *ecs.World, x, y, z float64, creator ecs.Entity) *Distraction {
	return s.CreateDistraction("small_fire", x, y, z, creator)
}
