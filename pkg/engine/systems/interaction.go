package systems

import (
	"math"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// InteractionSystem handles proximity detection and interaction feedback
// for environment objects. It updates highlight states based on player
// distance and targeting, and processes interaction attempts.
type InteractionSystem struct {
	// DefaultInteractionRange is the default range for interactions if not specified.
	DefaultInteractionRange float64
	// TargetingTolerance is the angle tolerance (radians) for look-based targeting.
	TargetingTolerance float64
	// MaxTargetDistance is the maximum distance for targeting entities.
	MaxTargetDistance float64
	// PlayerEntity is the entity ID of the current player (for client-side).
	PlayerEntity ecs.Entity
	// CurrentTarget is the currently targeted entity (if any).
	CurrentTarget ecs.Entity
	// CurrentTargetValid indicates if CurrentTarget is a valid target.
	CurrentTargetValid bool
	// pendingInteraction stores a pending interaction request.
	pendingInteraction *interactionRequest
}

// interactionRequest stores details of a requested interaction.
type interactionRequest struct {
	TargetEntity ecs.Entity
	Timestamp    float64
}

// InteractionResult holds the outcome of an interaction attempt.
type InteractionResult struct {
	Success      bool
	Message      string
	ItemPickedup ecs.Entity // Set if item was picked up
}

// NewInteractionSystem creates a new interaction system with default settings.
func NewInteractionSystem() *InteractionSystem {
	return &InteractionSystem{
		DefaultInteractionRange: 2.5,
		TargetingTolerance:      0.3, // ~17 degrees
		MaxTargetDistance:       10.0,
		CurrentTargetValid:      false,
	}
}

// Update processes interaction proximity and targeting each tick.
func (s *InteractionSystem) Update(w *ecs.World, dt float64) {
	s.updateProximityHighlights(w)
	s.updateTargeting(w)
	s.processInteractions(w)
}

// interactableEntityInfo holds extracted info for an interactable entity.
type interactableEntityInfo struct {
	Entity   ecs.Entity
	EnvObj   *components.EnvironmentObject
	Position *components.Position
}

// forEachInteractable iterates over all interactable entities with positions.
func (s *InteractionSystem) forEachInteractable(w *ecs.World, fn func(info interactableEntityInfo)) {
	for _, e := range w.Entities("EnvironmentObject", "Position") {
		envComp, ok := w.GetComponent(e, "EnvironmentObject")
		if !ok {
			continue
		}
		envObj := envComp.(*components.EnvironmentObject)
		if !envObj.CanInteract() {
			continue
		}

		posComp, ok := w.GetComponent(e, "Position")
		if !ok {
			continue
		}
		pos := posComp.(*components.Position)

		fn(interactableEntityInfo{
			Entity:   e,
			EnvObj:   envObj,
			Position: pos,
		})
	}
}

// updateProximityHighlights updates highlight states for entities near the player.
func (s *InteractionSystem) updateProximityHighlights(w *ecs.World) {
	playerPos := s.getPlayerPosition(w)
	if playerPos == nil {
		return
	}

	s.forEachInteractable(w, func(info interactableEntityInfo) {
		s.updateEntityHighlight(info, playerPos)
	})
}

// updateEntityHighlight updates the highlight state for a single entity.
func (s *InteractionSystem) updateEntityHighlight(info interactableEntityInfo, playerPos *components.Position) {
	dist := s.distanceBetween(playerPos, info.Position)
	interactionRange := s.getEffectiveRange(info.EnvObj.InteractionRange)
	info.EnvObj.HighlightState = s.calculateHighlightState(info.Entity, dist, interactionRange)
}

// getEffectiveRange returns the interaction range, using default if zero.
func (s *InteractionSystem) getEffectiveRange(specifiedRange float64) float64 {
	if specifiedRange <= 0 {
		return s.DefaultInteractionRange
	}
	return specifiedRange
}

// calculateHighlightState determines highlight level based on targeting and distance.
func (s *InteractionSystem) calculateHighlightState(entity ecs.Entity, dist, interactionRange float64) int {
	if entity == s.CurrentTarget && s.CurrentTargetValid {
		return 2 // Targeted
	}
	if dist <= interactionRange {
		return 1 // InRange
	}
	return 0 // None
}

// targetCandidate holds info about a potential targeting candidate.
type targetCandidate struct {
	entity   ecs.Entity
	distance float64
	angle    float64
}

// updateTargeting finds the entity the player is looking at.
func (s *InteractionSystem) updateTargeting(w *ecs.World) {
	s.CurrentTargetValid = false
	s.CurrentTarget = 0

	playerPos := s.getPlayerPosition(w)
	playerLook := s.getPlayerLookDirection(w)
	if playerPos == nil || playerLook == nil {
		return
	}

	best := s.findBestTarget(w, playerPos, playerLook)
	if best.entity != 0 {
		s.CurrentTarget = best.entity
		s.CurrentTargetValid = true
	}
}

// findBestTarget finds the best targeting candidate among interactable entities.
func (s *InteractionSystem) findBestTarget(w *ecs.World, playerPos *components.Position, playerLook []float64) targetCandidate {
	best := targetCandidate{
		distance: s.MaxTargetDistance + 1,
		angle:    s.TargetingTolerance + 1,
	}

	s.forEachInteractable(w, func(info interactableEntityInfo) {
		candidate := s.evaluateTargetCandidate(info, playerPos, playerLook)
		if candidate.entity != 0 && s.isBetterTarget(candidate, best) {
			best = candidate
		}
	})

	return best
}

// evaluateTargetCandidate checks if an entity qualifies as a target candidate.
func (s *InteractionSystem) evaluateTargetCandidate(info interactableEntityInfo, playerPos *components.Position, playerLook []float64) targetCandidate {
	dist := s.distanceBetween(playerPos, info.Position)
	if dist > s.MaxTargetDistance || dist <= 0.1 {
		return targetCandidate{}
	}

	angle := s.angleTo(playerPos, info.Position, playerLook)
	if angle > s.TargetingTolerance {
		return targetCandidate{}
	}

	return targetCandidate{entity: info.Entity, distance: dist, angle: angle}
}

// isBetterTarget returns true if candidate is a better target than current best.
func (s *InteractionSystem) isBetterTarget(candidate, best targetCandidate) bool {
	return candidate.distance < best.distance ||
		(candidate.distance == best.distance && candidate.angle < best.angle)
}

// processInteractions handles pending interaction requests.
func (s *InteractionSystem) processInteractions(w *ecs.World) {
	if s.pendingInteraction == nil {
		return
	}

	target := s.pendingInteraction.TargetEntity
	s.pendingInteraction = nil

	// Validate the interaction is still possible
	envComp, ok := w.GetComponent(target, "EnvironmentObject")
	if !ok {
		return
	}
	envObj := envComp.(*components.EnvironmentObject)

	if !envObj.CanInteract() {
		return
	}

	// Check player is still in range
	playerPos := s.getPlayerPosition(w)
	if playerPos == nil {
		return
	}

	posComp, ok := w.GetComponent(target, "Position")
	if !ok {
		return
	}
	targetPos := posComp.(*components.Position)

	dist := s.distanceBetween(playerPos, targetPos)
	interactionRange := envObj.InteractionRange
	if interactionRange <= 0 {
		interactionRange = s.DefaultInteractionRange
	}

	if dist > interactionRange {
		return
	}

	// Perform the interaction based on type
	s.performInteraction(w, target, envObj)
}

// performInteraction executes the interaction on the target entity.
func (s *InteractionSystem) performInteraction(w *ecs.World, target ecs.Entity, envObj *components.EnvironmentObject) {
	switch envObj.InteractionType {
	case components.InteractionPickup:
		s.handlePickup(w, target, envObj)
	case components.InteractionOpen:
		s.handleOpen(w, target, envObj)
	case components.InteractionUse:
		s.handleUse(w, target, envObj)
	case components.InteractionRead:
		s.handleRead(w, target, envObj)
	case components.InteractionTalk:
		s.handleTalk(w, target, envObj)
	case components.InteractionPush:
		s.handlePush(w, target, envObj)
	case components.InteractionExamine:
		s.handleExamine(w, target, envObj)
	}
}

// handlePickup processes picking up an item.
func (s *InteractionSystem) handlePickup(w *ecs.World, target ecs.Entity, envObj *components.EnvironmentObject) {
	// Get player inventory
	invComp, ok := w.GetComponent(s.PlayerEntity, "Inventory")
	if !ok {
		return
	}
	inventory := invComp.(*components.Inventory)

	// Add the item to inventory using its ItemID
	if envObj.ItemID != "" {
		inventory.Items = append(inventory.Items, envObj.ItemID)
	}

	// Mark as picked up (could remove from world or flag as collected)
	envObj.InteractionType = components.InteractionNone

	// Hide the entity (could also destroy it)
	if appearComp, ok := w.GetComponent(target, "Appearance"); ok {
		appearance := appearComp.(*components.Appearance)
		appearance.Visible = false
	}
}

// handleOpen processes opening a container or door.
func (s *InteractionSystem) handleOpen(w *ecs.World, target ecs.Entity, envObj *components.EnvironmentObject) {
	// Check if locked
	if envObj.NeedsKey() {
		// Check if player has the key
		invComp, ok := w.GetComponent(s.PlayerEntity, "Inventory")
		if !ok {
			return
		}
		inventory := invComp.(*components.Inventory)

		hasKey := false
		for _, item := range inventory.Items {
			if item == envObj.RequiredKeyID {
				hasKey = true
				break
			}
		}

		if !hasKey {
			return // Can't open without key
		}

		// Unlock it
		envObj.RequiredKeyID = ""
		envObj.IsLocked = false
	}

	// Toggle open state
	envObj.IsOpen = !envObj.IsOpen

	// If the door has a physics body, trigger swing animation
	if envObj.IsDoor() {
		s.triggerDoorSwing(w, target, envObj.IsOpen)
	}
}

// triggerDoorSwing initiates swing animation for a door entity.
// Opens doors swing to ~90° (pi/2), closing doors swing back to 0°.
func (s *InteractionSystem) triggerDoorSwing(w *ecs.World, doorEntity ecs.Entity, opening bool) {
	physComp, hasPhys := w.GetComponent(doorEntity, "PhysicsBody")
	if !hasPhys {
		return
	}

	phys := physComp.(*components.PhysicsBody)
	if !phys.IsSwinging {
		// Not a swinging door - enable swing physics
		phys.IsSwinging = true
		phys.MaxSwingAngle = math.Pi / 2 // 90 degrees
		phys.SwingDamping = 4.0          // Dampen quickly for door feel
	}

	// Apply angular impulse to swing the door
	// The swing velocity is set to achieve ~90° over ~0.5 seconds
	swingSpeed := 3.5 // radians per second
	if opening {
		// Opening: swing to positive angle
		phys.SwingVelocity = swingSpeed
	} else {
		// Closing: swing to negative/zero angle
		phys.SwingVelocity = -swingSpeed
	}
}

// handleUse processes using an interactive object.
func (s *InteractionSystem) handleUse(w *ecs.World, target ecs.Entity, envObj *components.EnvironmentObject) {
	// Mark as used - specific behavior depends on object type
	envObj.IsUsed = true
}

// handleRead processes reading a book or sign.
func (s *InteractionSystem) handleRead(w *ecs.World, target ecs.Entity, envObj *components.EnvironmentObject) {
	// Mark as read - UI system should display the content
	envObj.IsUsed = true
}

// handleTalk processes talking to an NPC.
func (s *InteractionSystem) handleTalk(w *ecs.World, target ecs.Entity, envObj *components.EnvironmentObject) {
	// Trigger dialog - dialog system should pick this up
	// Set a flag or create a dialog event
}

// handlePush processes pushing a movable object.
func (s *InteractionSystem) handlePush(w *ecs.World, target ecs.Entity, envObj *components.EnvironmentObject) {
	// Get player direction to determine push direction
	playerPos := s.getPlayerPosition(w)
	if playerPos == nil {
		return
	}

	// Calculate push force direction
	pushForce := 2.0
	forceX := math.Cos(playerPos.Angle) * pushForce
	forceY := math.Sin(playerPos.Angle) * pushForce

	// Try to apply physics-based push first
	physComp, hasPhysics := w.GetComponent(target, "PhysicsBody")
	if hasPhysics {
		phys := physComp.(*components.PhysicsBody)
		if phys.IsPushable {
			phys.ApplyImpulse(forceX, forceY, 0)
			envObj.IsUsed = true
			return
		}
		if phys.IsSwinging {
			// Apply swing impulse (direction determines swing direction)
			swingImpulse := 1.5
			if forceX > 0 {
				phys.ApplySwingImpulse(swingImpulse)
			} else {
				phys.ApplySwingImpulse(-swingImpulse)
			}
			envObj.IsUsed = true
			return
		}
	}

	// Fallback: directly move the object's position (for objects without PhysicsBody)
	objPosComp, ok := w.GetComponent(target, "Position")
	if !ok {
		return
	}
	objPos := objPosComp.(*components.Position)

	// Apply a simple push by moving the object in player's facing direction
	smallPush := 0.5
	objPos.X += math.Cos(playerPos.Angle) * smallPush
	objPos.Y += math.Sin(playerPos.Angle) * smallPush

	// Mark as used
	envObj.IsUsed = true
}

// handleExamine processes examining an object.
func (s *InteractionSystem) handleExamine(w *ecs.World, target ecs.Entity, envObj *components.EnvironmentObject) {
	// Mark as examined - UI should show description
	envObj.IsUsed = true
}

// RequestInteraction queues an interaction with the current target.
func (s *InteractionSystem) RequestInteraction(gameTime float64) {
	if !s.CurrentTargetValid {
		return
	}

	s.pendingInteraction = &interactionRequest{
		TargetEntity: s.CurrentTarget,
		Timestamp:    gameTime,
	}
}

// RequestInteractionWith queues an interaction with a specific entity.
func (s *InteractionSystem) RequestInteractionWith(target ecs.Entity, gameTime float64) {
	s.pendingInteraction = &interactionRequest{
		TargetEntity: target,
		Timestamp:    gameTime,
	}
}

// GetCurrentTarget returns the currently targeted entity and whether it's valid.
func (s *InteractionSystem) GetCurrentTarget() (ecs.Entity, bool) {
	return s.CurrentTarget, s.CurrentTargetValid
}

// SetPlayerEntity sets the player entity for the interaction system.
func (s *InteractionSystem) SetPlayerEntity(e ecs.Entity) {
	s.PlayerEntity = e
}

// getPlayerPosition retrieves the player's position component.
func (s *InteractionSystem) getPlayerPosition(w *ecs.World) *components.Position {
	if s.PlayerEntity == 0 {
		return nil
	}

	comp, ok := w.GetComponent(s.PlayerEntity, "Position")
	if !ok {
		return nil
	}
	return comp.(*components.Position)
}

// getPlayerLookDirection retrieves the player's look direction as [2]float64.
func (s *InteractionSystem) getPlayerLookDirection(w *ecs.World) []float64 {
	if s.PlayerEntity == 0 {
		return nil
	}

	// Get from Position component which has Angle field
	comp, ok := w.GetComponent(s.PlayerEntity, "Position")
	if !ok {
		return nil
	}
	pos := comp.(*components.Position)

	// Position.Angle is the direction angle in radians
	// Convert to unit vector
	return []float64{math.Cos(pos.Angle), math.Sin(pos.Angle)}
}

// distanceBetween calculates the distance between two positions.
func (s *InteractionSystem) distanceBetween(a, b *components.Position) float64 {
	dx := b.X - a.X
	dy := b.Y - a.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// angleTo calculates the angle between the look direction and the direction to target.
func (s *InteractionSystem) angleTo(from, to *components.Position, lookDir []float64) float64 {
	if len(lookDir) < 2 {
		return math.Pi // Max angle if no look direction
	}

	// Direction to target
	dx := to.X - from.X
	dy := to.Y - from.Y
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist < 0.0001 {
		return 0 // At the same position
	}
	dx /= dist
	dy /= dist

	// Dot product gives cosine of angle
	dot := dx*lookDir[0] + dy*lookDir[1]

	// Clamp to valid range for acos
	if dot > 1 {
		dot = 1
	}
	if dot < -1 {
		dot = -1
	}

	return math.Acos(dot)
}

// GetInteractableInRange returns all interactable entities within range of the player.
func (s *InteractionSystem) GetInteractableInRange(w *ecs.World) []ecs.Entity {
	playerPos := s.getPlayerPosition(w)
	if playerPos == nil {
		return nil
	}

	var result []ecs.Entity

	s.forEachInteractable(w, func(info interactableEntityInfo) {
		interactionRange := info.EnvObj.InteractionRange
		if interactionRange <= 0 {
			interactionRange = s.DefaultInteractionRange
		}

		if s.distanceBetween(playerPos, info.Position) <= interactionRange {
			result = append(result, info.Entity)
		}
	})

	return result
}
