package systems

import (
	"math"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

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

// hidingSpotConfig holds configuration values for a hiding spot type.
type hidingSpotConfig struct {
	Radius              float64
	VisibilityReduction float64
	DetectionMultiplier float64
	Capacity            int
	CanBeSearched       bool
	SearchDifficulty    float64
	RequiredSkillLevel  int
}

// hidingSpotConfigs maps hiding spot types to their default configurations.
var hidingSpotConfigs = map[HidingSpotType]hidingSpotConfig{
	HidingSpotShadow:        {3.0, 0.6, 0.4, 2, false, 0.9, 0},
	HidingSpotFoliage:       {4.0, 0.7, 0.3, 3, true, 0.6, 0},
	HidingSpotContainer:     {1.5, 0.95, 0.1, 1, true, 0.3, 5},
	HidingSpotFurniture:     {2.0, 0.8, 0.2, 1, true, 0.4, 0},
	HidingSpotArchitectural: {2.5, 0.7, 0.35, 2, false, 0.7, 0},
	HidingSpotRooftop:       {5.0, 0.5, 0.6, 2, false, 0.8, 20},
	HidingSpotUnderwater:    {4.0, 0.9, 0.15, 1, false, 0.95, 10},
}

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

	// Basic availability checks
	if !s.isSpotAvailable(entity, spot) {
		return false
	}

	// Check skill requirement
	if !s.hasRequiredSkill(w, entity, spot) {
		return false
	}

	// Check position proximity
	return s.isWithinRange(w, entity, spot)
}

// isSpotAvailable checks if the spot has room and entity isn't already hiding.
func (s *HidingSpotSystem) isSpotAvailable(entity ecs.Entity, spot *HidingSpot) bool {
	// Check capacity
	if len(spot.Occupants) >= spot.Capacity {
		return false
	}

	// Check if already hiding elsewhere
	entityID := uint64(entity)
	if _, hiding := s.EntityHiding[entityID]; hiding {
		return false
	}

	return true
}

// hasRequiredSkill checks if the entity meets skill requirements.
func (s *HidingSpotSystem) hasRequiredSkill(w *ecs.World, entity ecs.Entity, spot *HidingSpot) bool {
	if spot.RequiredSkillLevel <= 0 {
		return true
	}

	skillsComp, ok := w.GetComponent(entity, "Skills")
	if !ok {
		return false
	}

	skills := skillsComp.(*components.Skills)
	return skills.Levels != nil && skills.Levels["sneak"] >= spot.RequiredSkillLevel
}

// isWithinRange checks if the entity is close enough to the spot.
func (s *HidingSpotSystem) isWithinRange(w *ecs.World, entity ecs.Entity, spot *HidingSpot) bool {
	if spot.Position == nil {
		return true // No position requirement
	}

	posComp, ok := w.GetComponent(entity, "Position")
	if !ok {
		return false
	}
	pos := posComp.(*components.Position)

	dist := math.Sqrt(
		math.Pow(pos.X-spot.Position.X, 2) +
			math.Pow(pos.Y-spot.Position.Y, 2))
	return dist <= spot.Radius*2
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

	searcherSkill := s.getEntitySkillLevel(w, searcher, "perception")
	found := make([]ecs.Entity, 0)

	for _, occupant := range spot.Occupants {
		if s.isOccupantFound(w, occupant, searcherSkill, spot.SearchDifficulty) {
			found = append(found, occupant)
			s.ExitHidingSpot(w, occupant)
		}
	}
	return found
}

// getEntitySkillLevel retrieves a specific skill level for an entity.
func (s *HidingSpotSystem) getEntitySkillLevel(w *ecs.World, entity ecs.Entity, skillName string) int {
	skillsComp, ok := w.GetComponent(entity, "Skills")
	if !ok {
		return 0
	}
	skills := skillsComp.(*components.Skills)
	if skills.Levels == nil {
		return 0
	}
	return skills.Levels[skillName]
}

// isOccupantFound checks if a searcher discovers a hidden occupant.
func (s *HidingSpotSystem) isOccupantFound(w *ecs.World, occupant ecs.Entity, searcherSkill int, searchDifficulty float64) bool {
	hiderSkill := s.getEntitySkillLevel(w, occupant, "sneak")
	effectiveHiding := float64(hiderSkill) * searchDifficulty
	return float64(searcherSkill) > effectiveHiding
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

	// Set type-specific defaults using config map
	if cfg, ok := hidingSpotConfigs[spotType]; ok {
		spot.Radius = cfg.Radius
		spot.VisibilityReduction = cfg.VisibilityReduction
		spot.DetectionMultiplier = cfg.DetectionMultiplier
		spot.Capacity = cfg.Capacity
		spot.CanBeSearched = cfg.CanBeSearched
		spot.SearchDifficulty = cfg.SearchDifficulty
		spot.RequiredSkillLevel = cfg.RequiredSkillLevel
	} else {
		// Default values for unknown types
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
