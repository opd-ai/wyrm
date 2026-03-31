package systems

import (
	"math"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

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
	pos, awareness := s.getNPCDistractionComponents(w, npc)
	if pos == nil || awareness == nil {
		return nil
	}

	var best *Distraction
	bestScore := 0.0

	for _, distraction := range s.ActiveDistractions {
		score := s.evaluateDistraction(distraction, pos, awareness)
		if score > bestScore {
			bestScore = score
			best = distraction
		}
	}
	return best
}

// getNPCDistractionComponents retrieves position and awareness for distraction evaluation.
func (s *DistractionSystem) getNPCDistractionComponents(w *ecs.World, npc ecs.Entity) (*components.Position, *components.Awareness) {
	posComp, ok := w.GetComponent(npc, "Position")
	if !ok {
		return nil, nil
	}
	awarenessComp, ok := w.GetComponent(npc, "Awareness")
	if !ok {
		return nil, nil
	}
	return posComp.(*components.Position), awarenessComp.(*components.Awareness)
}

// evaluateDistraction scores a distraction for an NPC, returns 0 if invalid.
func (s *DistractionSystem) evaluateDistraction(d *Distraction, pos *components.Position, awareness *components.Awareness) float64 {
	if !s.isValidDistraction(d) {
		return 0
	}
	if d.CanBeIgnored && awareness.AlertLevel > AwarenessThreshold {
		return 0
	}

	dist := s.calculateDistance(d.Position, pos)
	if dist > d.Radius {
		return 0
	}

	return s.calculateDistractionScore(d, dist)
}

// isValidDistraction checks if a distraction can be considered.
func (s *DistractionSystem) isValidDistraction(d *Distraction) bool {
	return d.TimeRemaining > 0 && d.Position != nil
}

// calculateDistance computes euclidean distance between two positions.
func (s *DistractionSystem) calculateDistance(a, b *components.Position) float64 {
	return math.Sqrt(math.Pow(a.X-b.X, 2) + math.Pow(a.Y-b.Y, 2))
}

// calculateDistractionScore computes the score based on priority, intensity, and proximity.
func (s *DistractionSystem) calculateDistractionScore(d *Distraction, dist float64) float64 {
	proximityScore := 1.0 - (dist / d.Radius)
	return float64(d.Priority) * d.Intensity * proximityScore
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
