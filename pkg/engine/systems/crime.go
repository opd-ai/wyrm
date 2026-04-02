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

// ============================================================================
// Guard Pursuit AI System
// ============================================================================

// GuardState represents a guard's current behavior state.
type GuardState int

const (
	GuardStatePatrol GuardState = iota
	GuardStateAlert
	GuardStatePursue
	GuardStateCombat
	GuardStateSearch
	GuardStateReturn
)

// GuardAI represents a guard's AI state for pursuit behavior.
type GuardAI struct {
	State          GuardState
	TargetEntity   ecs.Entity
	LastKnownX     float64
	LastKnownZ     float64
	SearchTimer    float64
	AlertTimer     float64
	PatrolPointIdx int
	PursuitSpeed   float64
	PatrolSpeed    float64
	SightRange     float64
	HearingRange   float64
	SearchDuration float64
}

// GuardPursuitSystem manages guard AI for pursuing criminals.
type GuardPursuitSystem struct {
	crimeSystem     *CrimeSystem
	pursuitSpeedMod float64 // Speed multiplier during pursuit
	alertDuration   float64 // How long guards stay alert
	searchDuration  float64 // How long guards search
	sightRange      float64 // Default sight range
	hearingRange    float64 // Default hearing range
}

// NewGuardPursuitSystem creates a new guard pursuit system.
func NewGuardPursuitSystem(crimeSystem *CrimeSystem) *GuardPursuitSystem {
	return &GuardPursuitSystem{
		crimeSystem:     crimeSystem,
		pursuitSpeedMod: 1.3,
		alertDuration:   5.0,
		searchDuration:  15.0,
		sightRange:      20.0,
		hearingRange:    15.0,
	}
}

// Update processes guard AI each tick.
func (s *GuardPursuitSystem) Update(w *ecs.World, dt float64) {
	for _, e := range w.Entities("Guard", "Position") {
		s.processGuard(w, e, dt)
	}
}

// processGuard updates a single guard's AI.
func (s *GuardPursuitSystem) processGuard(w *ecs.World, guard ecs.Entity, dt float64) {
	comp, ok := w.GetComponent(guard, "Guard")
	if !ok {
		return
	}
	guardComp := comp.(*components.Guard)

	// Initialize guard AI if needed
	if guardComp.SightRange <= 0 {
		guardComp.SightRange = s.sightRange
	}
	if guardComp.PursuitSpeed <= 0 {
		guardComp.PursuitSpeed = 4.0 // Default pursuit speed
	}

	switch guardComp.State {
	case int(GuardStatePatrol):
		s.handlePatrol(w, guard, guardComp, dt)
	case int(GuardStateAlert):
		s.handleAlert(w, guard, guardComp, dt)
	case int(GuardStatePursue):
		s.handlePursue(w, guard, guardComp, dt)
	case int(GuardStateCombat):
		s.handleCombat(w, guard, guardComp, dt)
	case int(GuardStateSearch):
		s.handleSearch(w, guard, guardComp, dt)
	case int(GuardStateReturn):
		s.handleReturn(w, guard, guardComp, dt)
	default:
		guardComp.State = int(GuardStatePatrol)
	}
}

// handlePatrol processes guard in patrol state.
func (s *GuardPursuitSystem) handlePatrol(w *ecs.World, guard ecs.Entity, guardComp *components.Guard, dt float64) {
	// Check for criminals in sight
	criminal := s.findVisibleCriminal(w, guard, guardComp)
	if criminal != 0 {
		guardComp.State = int(GuardStatePursue)
		guardComp.TargetEntity = uint64(criminal)
		s.updateLastKnownPosition(w, guardComp, criminal)
		return
	}

	// Continue patrol (movement handled by a separate movement system)
	guardComp.PatrolTimer += dt
}

// handleAlert processes guard in alert state.
func (s *GuardPursuitSystem) handleAlert(w *ecs.World, guard ecs.Entity, guardComp *components.Guard, dt float64) {
	guardComp.AlertTimer -= dt

	// Check for criminals
	criminal := s.findVisibleCriminal(w, guard, guardComp)
	if criminal != 0 {
		guardComp.State = int(GuardStatePursue)
		guardComp.TargetEntity = uint64(criminal)
		s.updateLastKnownPosition(w, guardComp, criminal)
		return
	}

	// Alert timer expired
	if guardComp.AlertTimer <= 0 {
		guardComp.State = int(GuardStatePatrol)
	}
}

// handlePursue processes guard in pursuit state.
func (s *GuardPursuitSystem) handlePursue(w *ecs.World, guard ecs.Entity, guardComp *components.Guard, dt float64) {
	target := ecs.Entity(guardComp.TargetEntity)

	// Check if target is still visible
	if s.canSeeTarget(w, guard, target, guardComp) {
		s.updateLastKnownPosition(w, guardComp, target)
		s.moveToward(w, guard, guardComp.LastKnownX, guardComp.LastKnownZ, guardComp.PursuitSpeed*s.pursuitSpeedMod, dt)

		// Check if in combat range
		if s.getDistanceToTarget(w, guard, target) < 2.0 {
			guardComp.State = int(GuardStateCombat)
		}
	} else {
		// Lost sight, switch to search
		guardComp.State = int(GuardStateSearch)
		guardComp.SearchTimer = s.searchDuration
	}
}

// handleCombat processes guard in combat state.
func (s *GuardPursuitSystem) handleCombat(w *ecs.World, guard ecs.Entity, guardComp *components.Guard, dt float64) {
	target := ecs.Entity(guardComp.TargetEntity)

	// Check if target escaped
	dist := s.getDistanceToTarget(w, guard, target)
	if dist > 3.0 {
		guardComp.State = int(GuardStatePursue)
		return
	}

	// Combat is handled by CombatSystem
	// Guard applies arrest effect here
	s.attemptArrest(w, target)
}

// handleSearch processes guard in search state.
func (s *GuardPursuitSystem) handleSearch(w *ecs.World, guard ecs.Entity, guardComp *components.Guard, dt float64) {
	guardComp.SearchTimer -= dt

	// Move toward last known position
	s.moveToward(w, guard, guardComp.LastKnownX, guardComp.LastKnownZ, guardComp.PursuitSpeed*0.7, dt)

	// Check for criminal again
	criminal := s.findVisibleCriminal(w, guard, guardComp)
	if criminal != 0 {
		guardComp.State = int(GuardStatePursue)
		guardComp.TargetEntity = uint64(criminal)
		s.updateLastKnownPosition(w, guardComp, criminal)
		return
	}

	// Search timer expired
	if guardComp.SearchTimer <= 0 {
		guardComp.State = int(GuardStateReturn)
	}
}

// handleReturn processes guard returning to patrol.
func (s *GuardPursuitSystem) handleReturn(w *ecs.World, guard ecs.Entity, guardComp *components.Guard, dt float64) {
	// Return to patrol start (simple version: just switch state)
	guardComp.State = int(GuardStatePatrol)
	guardComp.TargetEntity = 0
}

// findVisibleCriminal finds a criminal entity visible to the guard.
func (s *GuardPursuitSystem) findVisibleCriminal(w *ecs.World, guard ecs.Entity, guardComp *components.Guard) ecs.Entity {
	guardPos := s.getPosition(w, guard)

	for _, e := range w.Entities("Crime", "Position") {
		crimeComp, ok := w.GetComponent(e, "Crime")
		if !ok {
			continue
		}
		crime := crimeComp.(*components.Crime)

		// Only pursue wanted criminals
		if crime.WantedLevel <= 0 || crime.InJail {
			continue
		}

		pos := s.getPosition(w, e)
		dist := s.distance(guardPos, pos)

		if dist <= guardComp.SightRange {
			return e
		}
	}
	return 0
}

// canSeeTarget checks if the guard can see the target.
func (s *GuardPursuitSystem) canSeeTarget(w *ecs.World, guard, target ecs.Entity, guardComp *components.Guard) bool {
	dist := s.getDistanceToTarget(w, guard, target)
	return dist <= guardComp.SightRange
}

// updateLastKnownPosition updates the guard's record of the target's position.
func (s *GuardPursuitSystem) updateLastKnownPosition(w *ecs.World, guardComp *components.Guard, target ecs.Entity) {
	pos := s.getPosition(w, target)
	guardComp.LastKnownX = pos[0]
	guardComp.LastKnownZ = pos[2]
}

// moveToward moves the guard toward a target position.
func (s *GuardPursuitSystem) moveToward(w *ecs.World, guard ecs.Entity, targetX, targetZ, speed, dt float64) {
	posComp, ok := w.GetComponent(guard, "Position")
	if !ok {
		return
	}
	pos := posComp.(*components.Position)

	dx := targetX - pos.X
	dz := targetZ - pos.Z
	dist := s.sqrt(dx*dx + dz*dz)

	if dist < 0.1 {
		return
	}

	// Normalize and scale by speed
	moveX := (dx / dist) * speed * dt
	moveZ := (dz / dist) * speed * dt

	pos.X += moveX
	pos.Z += moveZ
}

// getDistanceToTarget returns the distance from guard to target.
func (s *GuardPursuitSystem) getDistanceToTarget(w *ecs.World, guard, target ecs.Entity) float64 {
	guardPos := s.getPosition(w, guard)
	targetPos := s.getPosition(w, target)
	return s.distance(guardPos, targetPos)
}

// attemptArrest attempts to arrest a criminal.
func (s *GuardPursuitSystem) attemptArrest(w *ecs.World, criminal ecs.Entity) {
	if s.crimeSystem == nil {
		return
	}
	// Send to jail
	s.crimeSystem.GoToJail(w, criminal, 60.0) // 1 minute jail time
}

// getPosition returns an entity's position.
func (s *GuardPursuitSystem) getPosition(w *ecs.World, e ecs.Entity) [3]float64 {
	comp, ok := w.GetComponent(e, "Position")
	if !ok {
		return [3]float64{}
	}
	pos := comp.(*components.Position)
	return [3]float64{pos.X, pos.Y, pos.Z}
}

// distance calculates Euclidean distance between two 3D points.
func (s *GuardPursuitSystem) distance(a, b [3]float64) float64 {
	dx := a[0] - b[0]
	dy := a[1] - b[1]
	dz := a[2] - b[2]
	return s.sqrt(dx*dx + dy*dy + dz*dz)
}

// sqrt is a simple square root approximation.
func (s *GuardPursuitSystem) sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	// Newton's method
	guess := x / 2
	for i := 0; i < 10; i++ {
		guess = (guess + x/guess) / 2
	}
	return guess
}

// AlertGuardsNearby alerts all guards within hearing range of a crime.
func (s *GuardPursuitSystem) AlertGuardsNearby(w *ecs.World, crimeX, crimeZ float64) int {
	alertedCount := 0
	for _, e := range w.Entities("Guard", "Position") {
		comp, ok := w.GetComponent(e, "Guard")
		if !ok {
			continue
		}
		guardComp := comp.(*components.Guard)

		pos := s.getPosition(w, e)
		dist := s.sqrt((pos[0]-crimeX)*(pos[0]-crimeX) + (pos[2]-crimeZ)*(pos[2]-crimeZ))

		if dist <= s.hearingRange {
			guardComp.State = int(GuardStateAlert)
			guardComp.AlertTimer = s.alertDuration
			guardComp.LastKnownX = crimeX
			guardComp.LastKnownZ = crimeZ
			alertedCount++
		}
	}
	return alertedCount
}

// SetSearchDuration sets how long guards search for criminals.
func (s *GuardPursuitSystem) SetSearchDuration(duration float64) {
	s.searchDuration = duration
}

// GetSearchDuration returns the current search duration.
func (s *GuardPursuitSystem) GetSearchDuration() float64 {
	return s.searchDuration
}

// ============================================================================
// Bribery System
// ============================================================================

// BribeTarget represents who can be bribed.
type BribeTarget int

const (
	// BribeTargetGuard bribes a pursuing guard.
	BribeTargetGuard BribeTarget = iota
	// BribeTargetWitness bribes a witness to forget.
	BribeTargetWitness
	// BribeTargetOfficial bribes an official to reduce bounty.
	BribeTargetOfficial
	// BribeTargetJailer bribes a jailer for early release.
	BribeTargetJailer
)

// BribeResult represents the outcome of a bribery attempt.
type BribeResult int

const (
	// BribeResultSuccess means the bribe was accepted.
	BribeResultSuccess BribeResult = iota
	// BribeResultFailed means the bribe was rejected.
	BribeResultFailed
	// BribeResultInsufficient means not enough money offered.
	BribeResultInsufficient
	// BribeResultReported means the target reported the bribe attempt.
	BribeResultReported
	// BribeResultNoTarget means there's no valid bribery target.
	BribeResultNoTarget
)

// BriberySystem handles bribery attempts to reduce wanted level and bounty.
type BriberySystem struct {
	crimeSystem        *CrimeSystem
	guardPursuitSystem *GuardPursuitSystem
	// Base costs per wanted level
	BaseBribeCostPerLevel float64
	// Success chance modifiers
	GuardSuccessBase    float64
	WitnessSuccessBase  float64
	OfficialSuccessBase float64
	JailerSuccessBase   float64
	// Chance to be reported when bribe fails
	ReportChanceOnFailure float64
	// Multipliers for different wanted levels
	HighWantedMultiplier float64 // Applied at 4-5 stars
	// Random generator for determinism
	rng *PseudoRandomLCG
}

// NewBriberySystem creates a new bribery system.
func NewBriberySystem(crimeSystem *CrimeSystem, guardPursuitSystem *GuardPursuitSystem, seed int64) *BriberySystem {
	return &BriberySystem{
		crimeSystem:           crimeSystem,
		guardPursuitSystem:    guardPursuitSystem,
		BaseBribeCostPerLevel: 100.0,
		GuardSuccessBase:      0.6, // 60% base success
		WitnessSuccessBase:    0.8, // 80% base success
		OfficialSuccessBase:   0.5, // 50% base success
		JailerSuccessBase:     0.4, // 40% base success
		ReportChanceOnFailure: 0.3, // 30% chance to report failed bribe
		HighWantedMultiplier:  2.0, // Double cost at high wanted level
		rng:                   NewPseudoRandomLCG(seed),
	}
}

// pseudoRandom generates a deterministic pseudo-random number 0.0-1.0.
func (s *BriberySystem) pseudoRandom() float64 {
	return s.rng.Float64()
}

// CalculateBribeCost calculates the cost to bribe a target.
func (s *BriberySystem) CalculateBribeCost(w *ecs.World, entity ecs.Entity, target BribeTarget) float64 {
	comp, ok := w.GetComponent(entity, "Crime")
	if !ok {
		return 0
	}
	crime := comp.(*components.Crime)

	baseCost := s.BaseBribeCostPerLevel * float64(crime.WantedLevel)

	// Apply target-specific multipliers
	switch target {
	case BribeTargetGuard:
		baseCost *= 1.0 // Standard
	case BribeTargetWitness:
		baseCost *= 0.5 // Witnesses are cheaper
	case BribeTargetOfficial:
		baseCost *= 2.0 // Officials cost more
	case BribeTargetJailer:
		baseCost *= 1.5 // Jailers are in-between
	}

	// High wanted level increases cost
	if crime.WantedLevel >= 4 {
		baseCost *= s.HighWantedMultiplier
	}

	return baseCost
}

// CalculateSuccessChance calculates bribe success probability.
func (s *BriberySystem) CalculateSuccessChance(w *ecs.World, entity ecs.Entity, target BribeTarget, offerMultiplier float64) float64 {
	comp, ok := w.GetComponent(entity, "Crime")
	if !ok {
		return 0
	}
	crime := comp.(*components.Crime)

	var baseChance float64
	switch target {
	case BribeTargetGuard:
		baseChance = s.GuardSuccessBase
	case BribeTargetWitness:
		baseChance = s.WitnessSuccessBase
	case BribeTargetOfficial:
		baseChance = s.OfficialSuccessBase
	case BribeTargetJailer:
		baseChance = s.JailerSuccessBase
	}

	// Offering more than base cost increases success chance
	if offerMultiplier > 1.0 {
		baseChance += (offerMultiplier - 1.0) * 0.1 // +10% per 100% extra
	}

	// Higher wanted level reduces success chance
	baseChance -= float64(crime.WantedLevel) * 0.05 // -5% per wanted level

	// Clamp to 5-95%
	if baseChance < 0.05 {
		baseChance = 0.05
	}
	if baseChance > 0.95 {
		baseChance = 0.95
	}

	return baseChance
}

// AttemptBribe attempts to bribe a target to reduce criminal status.
func (s *BriberySystem) AttemptBribe(w *ecs.World, entity ecs.Entity, target BribeTarget, offerAmount float64) BribeResult {
	comp, ok := w.GetComponent(entity, "Crime")
	if !ok {
		return BribeResultNoTarget
	}
	crime := comp.(*components.Crime)

	requiredCost := s.CalculateBribeCost(w, entity, target)
	if requiredCost <= 0 {
		return BribeResultNoTarget
	}

	if s.isOfferTooLow(offerAmount, requiredCost) {
		return BribeResultInsufficient
	}

	return s.resolveBribeAttempt(w, entity, crime, target, offerAmount, requiredCost)
}

// isOfferTooLow checks if the bribe offer is insufficiently low.
func (s *BriberySystem) isOfferTooLow(offer, required float64) bool {
	return offer < required*0.5
}

// resolveBribeAttempt determines the outcome of a bribe attempt.
func (s *BriberySystem) resolveBribeAttempt(w *ecs.World, entity ecs.Entity, crime *components.Crime, target BribeTarget, offer, required float64) BribeResult {
	offerMultiplier := offer / required
	successChance := s.CalculateSuccessChance(w, entity, target, offerMultiplier)

	if s.pseudoRandom() < successChance {
		s.applyBribeEffect(w, entity, crime, target)
		return BribeResultSuccess
	}

	return s.handleFailedBribe(crime)
}

// handleFailedBribe determines the consequence of a failed bribe.
func (s *BriberySystem) handleFailedBribe(crime *components.Crime) BribeResult {
	if s.pseudoRandom() < s.ReportChanceOnFailure {
		s.increaseCrimeSeverity(crime)
		return BribeResultReported
	}
	return BribeResultFailed
}

// increaseCrimeSeverity raises wanted level and bounty after a reported bribe.
func (s *BriberySystem) increaseCrimeSeverity(crime *components.Crime) {
	crime.WantedLevel++
	if crime.WantedLevel > MaxWantedLevel {
		crime.WantedLevel = MaxWantedLevel
	}
	crime.BountyAmount += s.BaseBribeCostPerLevel
}

// applyBribeEffect applies the effect of a successful bribe.
func (s *BriberySystem) applyBribeEffect(w *ecs.World, entity ecs.Entity, crime *components.Crime, target BribeTarget) {
	switch target {
	case BribeTargetGuard:
		// Guard stops pursuing - reduce wanted level by 1
		crime.WantedLevel--
		if crime.WantedLevel < 0 {
			crime.WantedLevel = 0
		}
		// Also resets any guards in pursuit (would need to track guard state)

	case BribeTargetWitness:
		// Witness forgets - prevent wanted level increase for a time
		// In full implementation, would mark witness as "bribed"
		crime.WantedLevel--
		if crime.WantedLevel < 0 {
			crime.WantedLevel = 0
		}

	case BribeTargetOfficial:
		// Official reduces bounty significantly
		crime.WantedLevel -= 2
		if crime.WantedLevel < 0 {
			crime.WantedLevel = 0
		}
		crime.BountyAmount /= 2

	case BribeTargetJailer:
		// Jailer releases from jail early
		if crime.InJail {
			crime.InJail = false
			crime.JailReleaseTime = 0
		}
	}
}

// BribeGuard attempts to bribe a specific pursuing guard.
func (s *BriberySystem) BribeGuard(w *ecs.World, criminal, guard ecs.Entity, offerAmount float64) BribeResult {
	// Verify guard is pursuing the criminal
	guardComp, ok := w.GetComponent(guard, "Guard")
	if !ok {
		return BribeResultNoTarget
	}
	guardState := guardComp.(*components.Guard)

	// Guard must be in pursuit or combat with this criminal
	if guardState.State != int(GuardStatePursue) && guardState.State != int(GuardStateCombat) {
		return BribeResultNoTarget
	}
	if ecs.Entity(guardState.TargetEntity) != criminal {
		return BribeResultNoTarget
	}

	result := s.AttemptBribe(w, criminal, BribeTargetGuard, offerAmount)

	// On success, guard returns to patrol
	if result == BribeResultSuccess {
		guardState.State = int(GuardStateReturn)
		guardState.TargetEntity = 0
	}

	return result
}

// BribeWitnessesNearby attempts to bribe all witnesses near a position.
func (s *BriberySystem) BribeWitnessesNearby(w *ecs.World, criminal ecs.Entity, x, z, radius, offerPerWitness float64) (successes, failures int) {
	for _, e := range w.Entities("Witness", "Position") {
		posComp, ok := w.GetComponent(e, "Position")
		if !ok {
			continue
		}
		pos := posComp.(*components.Position)

		// Check if witness is in range
		dx := pos.X - x
		dz := pos.Z - z
		dist := dx*dx + dz*dz
		if dist > radius*radius {
			continue
		}

		// Attempt bribe
		result := s.AttemptBribe(w, criminal, BribeTargetWitness, offerPerWitness)
		if result == BribeResultSuccess {
			successes++
			// Mark witness as bribed (would set CanReport = false temporarily)
			witnessComp, ok := w.GetComponent(e, "Witness")
			if ok {
				witness := witnessComp.(*components.Witness)
				witness.CanReport = false
			}
		} else {
			failures++
		}
	}
	return successes, failures
}

// GetBribeDescription returns a description of the bribe target.
func (s *BriberySystem) GetBribeDescription(target BribeTarget) string {
	switch target {
	case BribeTargetGuard:
		return "Bribe a guard to stop pursuit and look the other way."
	case BribeTargetWitness:
		return "Bribe witnesses to forget what they saw."
	case BribeTargetOfficial:
		return "Bribe an official to reduce your bounty and wanted level."
	case BribeTargetJailer:
		return "Bribe the jailer for early release from prison."
	default:
		return "Unknown bribe target."
	}
}

// GetAvailableBribeTargets returns which bribe options are currently available.
func (s *BriberySystem) GetAvailableBribeTargets(w *ecs.World, entity ecs.Entity) []BribeTarget {
	comp, ok := w.GetComponent(entity, "Crime")
	if !ok {
		return nil
	}
	crime := comp.(*components.Crime)

	var available []BribeTarget

	// Guard bribe available if wanted
	if crime.WantedLevel > 0 && !crime.InJail {
		available = append(available, BribeTargetGuard)
		available = append(available, BribeTargetWitness)
		available = append(available, BribeTargetOfficial)
	}

	// Jailer bribe only available in jail
	if crime.InJail {
		available = append(available, BribeTargetJailer)
	}

	return available
}

// Update processes bribery cooldowns and effects each tick.
func (s *BriberySystem) Update(w *ecs.World, dt float64) {
	// Bribery is event-driven; no per-tick processing required
}
