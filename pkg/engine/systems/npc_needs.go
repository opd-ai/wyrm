// Package systems contains ECS system implementations for Wyrm.
// NPCNeedsSystem simulates basic NPC needs (hunger, energy, social, safety)
// that influence behavior and decision-making.
package systems

import (
	"math"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// NPCNeedsSystem processes NPC needs over time, affecting behavior.
type NPCNeedsSystem struct {
	// GameTime is the current game time in hours.
	GameTime float64
	// LastUpdateTime tracks when needs were last updated.
	LastUpdateTime float64
	// DefaultHungerRate is the base hunger rate per hour.
	DefaultHungerRate float64
	// DefaultEnergyRate is the base energy drain per hour.
	DefaultEnergyRate float64
	// DefaultSocialDecayRate is the base social decay per hour.
	DefaultSocialDecayRate float64
}

// NewNPCNeedsSystem creates a new NPCNeedsSystem with default rates.
func NewNPCNeedsSystem() *NPCNeedsSystem {
	return &NPCNeedsSystem{
		DefaultHungerRate:      0.04, // Get hungry in ~25 hours
		DefaultEnergyRate:      0.06, // Get tired in ~16 hours
		DefaultSocialDecayRate: 0.02, // Get lonely over ~50 hours
	}
}

// Update processes all NPC needs based on elapsed game time.
func (s *NPCNeedsSystem) Update(w *ecs.World, dt float64) {
	// Calculate hours elapsed since last update
	hoursElapsed := s.GameTime - s.LastUpdateTime
	if hoursElapsed <= 0 {
		return
	}
	s.LastUpdateTime = s.GameTime

	// Process all entities with NPCNeeds
	for _, e := range w.Entities("NPCNeeds") {
		s.updateEntityNeeds(w, e, hoursElapsed)
	}
}

// updateEntityNeeds updates a single entity's needs.
func (s *NPCNeedsSystem) updateEntityNeeds(w *ecs.World, e ecs.Entity, hoursElapsed float64) {
	comp, ok := w.GetComponent(e, "NPCNeeds")
	if !ok {
		return
	}
	needs := comp.(*components.NPCNeeds)

	// Apply hunger increase
	hungerRate := needs.HungerRate
	if hungerRate <= 0 {
		hungerRate = s.DefaultHungerRate
	}
	needs.Hunger = clampNeed(needs.Hunger + hungerRate*hoursElapsed)

	// Apply energy decrease (only when awake)
	if s.isAwake(w, e) {
		energyRate := needs.EnergyRate
		if energyRate <= 0 {
			energyRate = s.DefaultEnergyRate
		}
		needs.Energy = clampNeed(needs.Energy - energyRate*hoursElapsed)
	} else {
		// Sleeping restores energy
		needs.Energy = clampNeed(needs.Energy + 0.1*hoursElapsed)
	}

	// Apply social decay (only when alone)
	if !s.hasNearbyNPCs(w, e) {
		socialRate := needs.SocialDecayRate
		if socialRate <= 0 {
			socialRate = s.DefaultSocialDecayRate
		}
		needs.Social = clampNeed(needs.Social - socialRate*hoursElapsed)
	} else {
		// Social interaction restores social need
		needs.Social = clampNeed(needs.Social + 0.05*hoursElapsed)
	}

	// Safety is affected by nearby threats
	needs.Safety = s.calculateSafety(w, e)
}

// isAwake checks if an NPC is currently awake based on schedule.
func (s *NPCNeedsSystem) isAwake(w *ecs.World, e ecs.Entity) bool {
	comp, ok := w.GetComponent(e, "Schedule")
	if !ok {
		return true // Default to awake
	}
	schedule := comp.(*components.Schedule)
	return schedule.CurrentActivity != "sleep" && schedule.CurrentActivity != "rest"
}

// hasNearbyNPCs checks if there are other NPCs near this entity.
func (s *NPCNeedsSystem) hasNearbyNPCs(w *ecs.World, e ecs.Entity) bool {
	pos, ok := w.GetComponent(e, "Position")
	if !ok {
		return false
	}
	entityPos := pos.(*components.Position)

	socialRange := 10.0 // Units for social interaction
	for _, other := range w.Entities("Position", "NPCNeeds") {
		if other == e {
			continue
		}
		otherPos, ok := w.GetComponent(other, "Position")
		if !ok {
			continue
		}
		otherPosition := otherPos.(*components.Position)
		dx := entityPos.X - otherPosition.X
		dy := entityPos.Y - otherPosition.Y
		dz := entityPos.Z - otherPosition.Z
		distSq := dx*dx + dy*dy + dz*dz
		if distSq <= socialRange*socialRange {
			return true
		}
	}
	return false
}

// calculateSafety determines safety based on nearby threats.
func (s *NPCNeedsSystem) calculateSafety(w *ecs.World, e ecs.Entity) float64 {
	pos, ok := w.GetComponent(e, "Position")
	if !ok {
		return 1.0 // Default to safe
	}
	entityPos := pos.(*components.Position)

	// Check for nearby hostile entities
	threatRange := 20.0
	threatCount := 0
	for _, other := range w.Entities("Position", "CombatState") {
		if other == e {
			continue
		}
		otherPos, ok := w.GetComponent(other, "Position")
		if !ok {
			continue
		}
		combatComp, ok := w.GetComponent(other, "CombatState")
		if !ok {
			continue
		}
		combat := combatComp.(*components.CombatState)
		if !combat.InCombat {
			continue
		}

		otherPosition := otherPos.(*components.Position)
		dx := entityPos.X - otherPosition.X
		dy := entityPos.Y - otherPosition.Y
		dz := entityPos.Z - otherPosition.Z
		distSq := dx*dx + dy*dy + dz*dz
		if distSq <= threatRange*threatRange {
			threatCount++
		}
	}

	// Each nearby threat reduces safety
	safety := 1.0 - float64(threatCount)*0.2
	return clampNeed(safety)
}

// GetNeedPriority returns the most pressing need for an NPC.
func (s *NPCNeedsSystem) GetNeedPriority(needs *components.NPCNeeds) string {
	if needs.Safety < 0.3 {
		return "safety" // Flee or hide
	}
	if needs.Energy < 0.2 {
		return "sleep"
	}
	if needs.Hunger > 0.8 {
		return "eat"
	}
	if needs.Social < 0.3 {
		return "socialize"
	}
	return "none"
}

// GetNeedModifier returns a behavior modifier based on current needs.
func (s *NPCNeedsSystem) GetNeedModifier(needs *components.NPCNeeds) NeedModifier {
	return NeedModifier{
		SpeedModifier:     calculateSpeedModifier(needs),
		AlertnessModifier: calculateAlertnessModifier(needs),
		MoodModifier:      calculateMoodModifier(needs),
	}
}

// NeedModifier contains behavior modifiers based on needs state.
type NeedModifier struct {
	SpeedModifier     float64 // Movement speed multiplier
	AlertnessModifier float64 // Perception/awareness multiplier
	MoodModifier      float64 // Social interaction modifier
}

// calculateSpeedModifier returns speed adjustment based on hunger and energy.
func calculateSpeedModifier(needs *components.NPCNeeds) float64 {
	// Low energy reduces speed
	energyFactor := 0.5 + needs.Energy*0.5
	// Very hungry reduces speed slightly
	hungerFactor := 1.0
	if needs.Hunger > 0.7 {
		hungerFactor = 0.9
	}
	return energyFactor * hungerFactor
}

// calculateAlertnessModifier returns alertness adjustment.
func calculateAlertnessModifier(needs *components.NPCNeeds) float64 {
	// Low energy reduces alertness
	if needs.Energy < 0.3 {
		return 0.6
	}
	// Fear increases alertness
	if needs.Safety < 0.5 {
		return 1.3
	}
	return 1.0
}

// calculateMoodModifier returns mood adjustment for social interactions.
func calculateMoodModifier(needs *components.NPCNeeds) float64 {
	// Combine all needs into mood
	avg := (1.0 - needs.Hunger + needs.Energy + needs.Social + needs.Safety) / 4.0
	return 0.5 + avg*0.5
}

// clampNeed clamps a need value to [0, 1].
func clampNeed(v float64) float64 {
	return math.Max(0, math.Min(1, v))
}
