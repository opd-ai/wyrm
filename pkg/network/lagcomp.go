// Package network provides client-server networking.
package network

import (
	"sync"
	"time"
)

// MaxRewindTime is the maximum time the server can rewind for hit detection.
// Per ROADMAP Phase 5 item 23: server rewinds entity state up to 500 ms.
const MaxRewindTime = 500 * time.Millisecond

// HistoryBufferSize is the number of state snapshots to keep.
const HistoryBufferSize = 64

// EntitySnapshot stores an entity's state at a point in time.
type EntitySnapshot struct {
	EntityID  uint64
	Timestamp time.Time
	Position  Position3D
	Angle     float32
	HitboxMin Position3D // Axis-aligned bounding box minimum
	HitboxMax Position3D // Axis-aligned bounding box maximum
}

// StateHistory stores a ring buffer of entity snapshots for lag compensation.
type StateHistory struct {
	mu        sync.RWMutex
	snapshots []EntitySnapshot
	writeIdx  int
	count     int
}

// NewStateHistory creates a new state history buffer.
func NewStateHistory() *StateHistory {
	return &StateHistory{
		snapshots: make([]EntitySnapshot, HistoryBufferSize),
	}
}

// Record stores an entity snapshot.
func (sh *StateHistory) Record(snapshot EntitySnapshot) {
	sh.mu.Lock()
	defer sh.mu.Unlock()

	sh.snapshots[sh.writeIdx] = snapshot
	sh.writeIdx = (sh.writeIdx + 1) % HistoryBufferSize
	if sh.count < HistoryBufferSize {
		sh.count++
	}
}

// GetAtTime returns the entity state closest to the given time.
func (sh *StateHistory) GetAtTime(entityID uint64, t time.Time) *EntitySnapshot {
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	var best *EntitySnapshot
	bestDiff := time.Duration(1<<63 - 1)

	for i := 0; i < sh.count; i++ {
		idx := (sh.writeIdx - 1 - i + HistoryBufferSize) % HistoryBufferSize
		snap := &sh.snapshots[idx]

		if snap.EntityID != entityID {
			continue
		}

		diff := t.Sub(snap.Timestamp)
		if diff < 0 {
			diff = -diff
		}

		if diff > MaxRewindTime {
			continue
		}

		if diff < bestDiff {
			bestDiff = diff
			copy := *snap
			best = &copy
		}
	}

	return best
}

// LagCompensator handles server-side lag compensation for hit detection.
// Per ROADMAP Phase 5 item 23:
// AC: Hit registration correct at 500 ms simulated RTT in automated test harness.
type LagCompensator struct {
	mu       sync.RWMutex
	entities map[uint64]*StateHistory

	// Tor-mode threshold: activates at RTT > 800ms
	torModeThreshold time.Duration
}

// NewLagCompensator creates a new lag compensator.
func NewLagCompensator() *LagCompensator {
	return &LagCompensator{
		entities:         make(map[uint64]*StateHistory),
		torModeThreshold: 800 * time.Millisecond,
	}
}

// RecordEntityState records an entity's current state for later rewind.
func (lc *LagCompensator) RecordEntityState(entityID uint64, pos Position3D, angle float32, hitboxMin, hitboxMax Position3D) {
	lc.mu.Lock()
	history, ok := lc.entities[entityID]
	if !ok {
		history = NewStateHistory()
		lc.entities[entityID] = history
	}
	lc.mu.Unlock()

	history.Record(EntitySnapshot{
		EntityID:  entityID,
		Timestamp: time.Now(),
		Position:  pos,
		Angle:     angle,
		HitboxMin: hitboxMin,
		HitboxMax: hitboxMax,
	})
}

// RemoveEntity removes an entity from tracking.
func (lc *LagCompensator) RemoveEntity(entityID uint64) {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	delete(lc.entities, entityID)
}

// HitResult contains the result of a lag-compensated hit test.
type HitResult struct {
	Hit           bool
	TargetID      uint64
	HitPosition   Position3D
	RewindTime    time.Duration
	ServerTime    time.Time
	CompensatedBy time.Duration
}

// HitTest performs a lag-compensated hit test.
func (lc *LagCompensator) HitTest(shooterID, targetID uint64, shotOrigin, shotDirection Position3D, clientTime time.Time, rtt time.Duration) *HitResult {
	lc.mu.RLock()
	targetHistory := lc.entities[targetID]
	lc.mu.RUnlock()

	if targetHistory == nil {
		return &HitResult{Hit: false}
	}

	rewindTime := clientTime.Add(-rtt / 2)

	now := time.Now()
	if now.Sub(rewindTime) > MaxRewindTime {
		rewindTime = now.Add(-MaxRewindTime)
	}

	targetState := targetHistory.GetAtTime(targetID, rewindTime)
	if targetState == nil {
		return &HitResult{Hit: false}
	}

	hit, distance := rayAABBIntersect(
		shotOrigin, shotDirection,
		targetState.HitboxMin, targetState.HitboxMax,
		targetState.Position,
	)

	if hit {
		hitPos := Position3D{
			X: shotOrigin.X + shotDirection.X*distance,
			Y: shotOrigin.Y + shotDirection.Y*distance,
			Z: shotOrigin.Z + shotDirection.Z*distance,
		}
		return &HitResult{
			Hit:           true,
			TargetID:      targetID,
			HitPosition:   hitPos,
			RewindTime:    now.Sub(rewindTime),
			ServerTime:    now,
			CompensatedBy: now.Sub(rewindTime),
		}
	}

	return &HitResult{Hit: false}
}

// IsTorMode returns whether the given RTT indicates Tor-level latency.
func (lc *LagCompensator) IsTorMode(rtt time.Duration) bool {
	return rtt > lc.torModeThreshold
}

// EntityCount returns the number of tracked entities.
func (lc *LagCompensator) EntityCount() int {
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	return len(lc.entities)
}

// translateBox computes the world-space AABB min/max by adding boxCenter to boxMin/boxMax.
func translateBox(boxMin, boxMax, boxCenter Position3D) (Position3D, Position3D) {
	return Position3D{
			X: boxMin.X + boxCenter.X,
			Y: boxMin.Y + boxCenter.Y,
			Z: boxMin.Z + boxCenter.Z,
		}, Position3D{
			X: boxMax.X + boxCenter.X,
			Y: boxMax.Y + boxCenter.Y,
			Z: boxMax.Z + boxCenter.Z,
		}
}

// slabIntersect computes the intersection interval for one axis slab.
// Returns updated (tmin, tmax) and whether the ray still intersects.
func slabIntersect(origin, direction, minVal, maxVal, tmin, tmax float32) (float32, float32, bool) {
	if direction != 0 {
		return slabIntersectMoving(origin, direction, minVal, maxVal, tmin, tmax)
	}
	return slabIntersectParallel(origin, minVal, maxVal, tmin, tmax)
}

// slabIntersectMoving handles slab intersection when ray is not parallel to axis.
func slabIntersectMoving(origin, direction, minVal, maxVal, tmin, tmax float32) (float32, float32, bool) {
	t1 := (minVal - origin) / direction
	t2 := (maxVal - origin) / direction
	if t1 > t2 {
		t1, t2 = t2, t1
	}
	if t1 > tmin {
		tmin = t1
	}
	if t2 < tmax {
		tmax = t2
	}
	return tmin, tmax, tmin <= tmax
}

// slabIntersectParallel handles slab intersection when ray is parallel to axis.
func slabIntersectParallel(origin, minVal, maxVal, tmin, tmax float32) (float32, float32, bool) {
	if origin < minVal || origin > maxVal {
		return tmin, tmax, false
	}
	return tmin, tmax, true
}

// rayAABBIntersect performs ray-AABB intersection test.
func rayAABBIntersect(origin, direction, boxMin, boxMax, boxCenter Position3D) (bool, float32) {
	min, max := translateBox(boxMin, boxMax, boxCenter)

	tmin, tmax := float32(0), float32(1000.0)
	var ok bool

	tmin, tmax, ok = slabIntersect(origin.X, direction.X, min.X, max.X, tmin, tmax)
	if !ok {
		return false, 0
	}

	tmin, tmax, ok = slabIntersect(origin.Y, direction.Y, min.Y, max.Y, tmin, tmax)
	if !ok {
		return false, 0
	}

	tmin, _, ok = slabIntersect(origin.Z, direction.Z, min.Z, max.Z, tmin, tmax)
	if !ok {
		return false, 0
	}

	return tmin >= 0, tmin
}

// PvPCombatType represents a type of PvP combat action.
type PvPCombatType int

const (
	PvPMeleeAttack PvPCombatType = iota
	PvPRangedAttack
	PvPMagicAttack
	PvPAreaEffect
	PvPStatusEffect
)

// PvPCombatAction represents a PvP combat action from a client.
type PvPCombatAction struct {
	AttackerID  uint64
	TargetID    uint64
	ActionType  PvPCombatType
	DamageClaim float64
	ClientTime  time.Time
	Position    Position3D
	Direction   Position3D
	AbilityID   string
}

// PvPValidationResult contains the result of validating a PvP action.
type PvPValidationResult struct {
	Valid             bool
	ActualDamage      float64
	RejectionReason   string
	ServerTime        time.Time
	PositionCorrected bool
	CorrectedPosition Position3D
}

// PvPValidator validates PvP combat actions.
type PvPValidator struct {
	mu             sync.RWMutex
	lagComp        *LagCompensator
	maxDamageRates map[PvPCombatType]float64       // Max damage per second per type
	cooldowns      map[uint64]map[string]time.Time // entityID -> abilityID -> lastUse
	zoneConfig     map[string]bool                 // Zone names where PvP is enabled
}

// NewPvPValidator creates a new PvP validator.
func NewPvPValidator(lagComp *LagCompensator) *PvPValidator {
	pv := &PvPValidator{
		lagComp:        lagComp,
		maxDamageRates: make(map[PvPCombatType]float64),
		cooldowns:      make(map[uint64]map[string]time.Time),
		zoneConfig:     make(map[string]bool),
	}
	pv.initDefaults()
	return pv
}

// initDefaults sets up default damage rate limits.
func (pv *PvPValidator) initDefaults() {
	pv.maxDamageRates[PvPMeleeAttack] = 50.0  // 50 damage per second max
	pv.maxDamageRates[PvPRangedAttack] = 40.0 // 40 damage per second max
	pv.maxDamageRates[PvPMagicAttack] = 60.0  // 60 damage per second max
	pv.maxDamageRates[PvPAreaEffect] = 30.0   // 30 damage per second max
	pv.maxDamageRates[PvPStatusEffect] = 20.0 // 20 damage per second max
}

// SetZonePvPEnabled configures PvP enabled status for a zone.
func (pv *PvPValidator) SetZonePvPEnabled(zoneName string, enabled bool) {
	pv.mu.Lock()
	defer pv.mu.Unlock()
	pv.zoneConfig[zoneName] = enabled
}

// IsZonePvPEnabled returns whether PvP is enabled in a zone.
func (pv *PvPValidator) IsZonePvPEnabled(zoneName string) bool {
	pv.mu.RLock()
	defer pv.mu.RUnlock()
	enabled, ok := pv.zoneConfig[zoneName]
	return ok && enabled
}

// ValidateAction validates a PvP combat action.
func (pv *PvPValidator) ValidateAction(action *PvPCombatAction, rtt time.Duration, zoneName string) *PvPValidationResult {
	pv.mu.Lock()
	defer pv.mu.Unlock()

	result := &PvPValidationResult{
		ServerTime: time.Now(),
	}

	// Check if PvP is allowed in this zone
	if !pv.zoneConfig[zoneName] {
		result.Valid = false
		result.RejectionReason = "PvP not enabled in zone"
		return result
	}

	// Validate cooldowns
	if !pv.validateCooldown(action) {
		result.Valid = false
		result.RejectionReason = "ability on cooldown"
		return result
	}

	// Validate damage claim against rate limits
	if !pv.validateDamageRate(action) {
		result.Valid = false
		result.RejectionReason = "damage rate exceeded"
		result.ActualDamage = pv.maxDamageRates[action.ActionType]
		return result
	}

	// Use lag compensation to validate hit
	hitResult := pv.lagComp.HitTest(
		action.AttackerID,
		action.TargetID,
		action.Position,
		action.Direction,
		action.ClientTime,
		rtt,
	)

	if !hitResult.Hit {
		result.Valid = false
		result.RejectionReason = "hit not confirmed"
		return result
	}

	// Record cooldown
	pv.recordCooldown(action)

	result.Valid = true
	result.ActualDamage = action.DamageClaim
	return result
}

// validateCooldown checks if an ability is off cooldown.
func (pv *PvPValidator) validateCooldown(action *PvPCombatAction) bool {
	if action.AbilityID == "" {
		return true // Basic attacks have no cooldown
	}
	entityCooldowns := pv.cooldowns[action.AttackerID]
	if entityCooldowns == nil {
		return true
	}
	lastUse, ok := entityCooldowns[action.AbilityID]
	if !ok {
		return true
	}
	// Default 1 second cooldown
	return time.Since(lastUse) > time.Second
}

// validateDamageRate checks if damage claim is within acceptable rates.
func (pv *PvPValidator) validateDamageRate(action *PvPCombatAction) bool {
	maxRate := pv.maxDamageRates[action.ActionType]
	// Allow instant damage up to max rate (assumes 1 action per second)
	return action.DamageClaim <= maxRate
}

// recordCooldown records when an ability was used.
func (pv *PvPValidator) recordCooldown(action *PvPCombatAction) {
	if action.AbilityID == "" {
		return
	}
	if pv.cooldowns[action.AttackerID] == nil {
		pv.cooldowns[action.AttackerID] = make(map[string]time.Time)
	}
	pv.cooldowns[action.AttackerID][action.AbilityID] = time.Now()
}

// CleanupCooldowns removes old cooldown entries.
func (pv *PvPValidator) CleanupCooldowns(olderThan time.Duration) {
	pv.mu.Lock()
	defer pv.mu.Unlock()
	cutoff := time.Now().Add(-olderThan)
	for entityID, abilities := range pv.cooldowns {
		for abilityID, lastUse := range abilities {
			if lastUse.Before(cutoff) {
				delete(abilities, abilityID)
			}
		}
		if len(abilities) == 0 {
			delete(pv.cooldowns, entityID)
		}
	}
}

// GetCooldownRemaining returns the remaining cooldown for an ability.
func (pv *PvPValidator) GetCooldownRemaining(entityID uint64, abilityID string) time.Duration {
	pv.mu.RLock()
	defer pv.mu.RUnlock()
	entityCooldowns := pv.cooldowns[entityID]
	if entityCooldowns == nil {
		return 0
	}
	lastUse, ok := entityCooldowns[abilityID]
	if !ok {
		return 0
	}
	remaining := time.Second - time.Since(lastUse)
	if remaining < 0 {
		return 0
	}
	return remaining
}
