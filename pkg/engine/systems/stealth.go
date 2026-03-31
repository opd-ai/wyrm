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
