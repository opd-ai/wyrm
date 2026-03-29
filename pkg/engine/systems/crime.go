package systems

import (
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// CrimeSystem tracks crimes, wanted levels, witnesses, and bounties.
type CrimeSystem struct {
	// DecayDelay is seconds without new crime before wanted level decreases.
	DecayDelay float64
	// BountyPerLevel is bounty added per wanted level.
	BountyPerLevel float64
	// GameTime is the current game time for tracking decay.
	GameTime float64
	// WitnessRange is the maximum distance for witnesses to observe crimes.
	WitnessRange float64
}

// NewCrimeSystem creates a new crime system with specified decay delay.
func NewCrimeSystem(decayDelay, bountyPerLevel float64) *CrimeSystem {
	return &CrimeSystem{
		DecayDelay:     decayDelay,
		BountyPerLevel: bountyPerLevel,
		WitnessRange:   DefaultWitnessRange,
	}
}

// Update processes crime detection and bounty updates each tick.
func (s *CrimeSystem) Update(w *ecs.World, dt float64) {
	s.GameTime += dt
	for _, e := range w.Entities("Crime") {
		s.processCrimeEntity(w, e)
	}
}

// processCrimeEntity updates a single entity's crime state.
func (s *CrimeSystem) processCrimeEntity(w *ecs.World, e ecs.Entity) {
	comp, ok := w.GetComponent(e, "Crime")
	if !ok {
		return
	}
	crime := comp.(*components.Crime)

	if crime.InJail {
		return
	}

	s.decayWantedLevel(crime)
	s.clampWantedLevel(crime)
}

// decayWantedLevel decreases wanted level if enough time has passed.
func (s *CrimeSystem) decayWantedLevel(crime *components.Crime) {
	if crime.WantedLevel <= 0 {
		return
	}
	timeSinceCrime := s.GameTime - crime.LastCrimeTime
	if timeSinceCrime >= s.DecayDelay {
		crime.WantedLevel--
		crime.LastCrimeTime = s.GameTime
	}
}

// clampWantedLevel constrains wanted level to valid range.
func (s *CrimeSystem) clampWantedLevel(crime *components.Crime) {
	if crime.WantedLevel < MinWantedLevel {
		crime.WantedLevel = MinWantedLevel
	}
	if crime.WantedLevel > MaxWantedLevel {
		crime.WantedLevel = MaxWantedLevel
	}
}

// ReportCrime increases wanted level for an entity if witnessed.
func (s *CrimeSystem) ReportCrime(w *ecs.World, criminal ecs.Entity) {
	comp, ok := w.GetComponent(criminal, "Crime")
	if !ok {
		return
	}
	crime := comp.(*components.Crime)
	criminalPos := s.getEntityPosition(w, criminal)
	if !s.isWitnessed(w, criminalPos) {
		return
	}
	crime.WantedLevel++
	crime.BountyAmount += s.BountyPerLevel
	crime.LastCrimeTime = s.GameTime
}

// isWitnessed checks if any witness can observe the given position.
func (s *CrimeSystem) isWitnessed(w *ecs.World, pos [3]float64) bool {
	for _, witness := range w.Entities("Witness", "Position") {
		if s.canWitnessObserve(w, witness, pos) {
			return true
		}
	}
	return false
}

// canWitnessObserve checks if a specific witness can observe a position.
func (s *CrimeSystem) canWitnessObserve(w *ecs.World, witness ecs.Entity, pos [3]float64) bool {
	wComp, ok := w.GetComponent(witness, "Witness")
	if !ok {
		return false
	}
	wState := wComp.(*components.Witness)
	if !wState.CanReport {
		return false
	}
	witnessPos := s.getEntityPosition(w, witness)
	return s.canWitnessSee(pos, witnessPos)
}

// canWitnessSee checks if a witness can see the crime location (LOS + range).
func (s *CrimeSystem) canWitnessSee(crimePos, witnessPos [3]float64) bool {
	// Simple distance-based check (future: actual line-of-sight)
	dx := crimePos[0] - witnessPos[0]
	dy := crimePos[1] - witnessPos[1]
	distSq := dx*dx + dy*dy
	return distSq <= s.WitnessRange*s.WitnessRange
}

// getEntityPosition returns an entity's position or zero if not found.
func (s *CrimeSystem) getEntityPosition(w *ecs.World, e ecs.Entity) [3]float64 {
	comp, ok := w.GetComponent(e, "Position")
	if !ok {
		return [3]float64{}
	}
	pos := comp.(*components.Position)
	return [3]float64{pos.X, pos.Y, pos.Z}
}

// PayBounty allows an entity to pay off their bounty and reset wanted level.
func (s *CrimeSystem) PayBounty(w *ecs.World, entity ecs.Entity) bool {
	comp, ok := w.GetComponent(entity, "Crime")
	if !ok {
		return false
	}
	crime := comp.(*components.Crime)
	// Could integrate with Economy system for actual payment
	crime.WantedLevel = 0
	crime.BountyAmount = 0
	crime.InJail = false
	return true
}

// GoToJail sends an entity to jail (sets flag, would teleport in full impl).
func (s *CrimeSystem) GoToJail(w *ecs.World, entity ecs.Entity, jailTime float64) bool {
	comp, ok := w.GetComponent(entity, "Crime")
	if !ok {
		return false
	}
	crime := comp.(*components.Crime)
	crime.InJail = true
	crime.JailReleaseTime = s.GameTime + jailTime
	// Note: bounty remains, wanted level clears after jail time
	crime.WantedLevel = 0
	return true
}

// CheckJailRelease checks if entity should be released from jail.
func (s *CrimeSystem) CheckJailRelease(w *ecs.World) {
	for _, e := range w.Entities("Crime") {
		comp, ok := w.GetComponent(e, "Crime")
		if !ok {
			continue
		}
		crime := comp.(*components.Crime)
		if crime.InJail && s.GameTime >= crime.JailReleaseTime {
			crime.InJail = false
		}
	}
}
