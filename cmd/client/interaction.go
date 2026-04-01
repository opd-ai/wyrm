//go:build !noebiten

// Package main provides the interaction system for player-world interaction.
package main

import (
	"math"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// InteractionType defines the kind of interaction available with an entity.
type InteractionType int

const (
	InteractionNone InteractionType = iota
	InteractionNPC
	InteractionItem
	InteractionWorkbench
	InteractionDoor
	InteractionContainer
)

// InteractionResult holds information about an interactable entity.
type InteractionResult struct {
	Entity   ecs.Entity
	Type     InteractionType
	Distance float64
	Name     string
	Prompt   string
}

// InteractionSystem handles player interaction with world entities.
type InteractionSystem struct {
	maxRange       float64 // Maximum interaction range in world units
	playerEntity   ecs.Entity
	currentTarget  *InteractionResult
	interactionKey bool // Whether interaction key is pressed
}

// NewInteractionSystem creates a new interaction system.
func NewInteractionSystem(playerEntity ecs.Entity, maxRange float64) *InteractionSystem {
	return &InteractionSystem{
		maxRange:     maxRange,
		playerEntity: playerEntity,
	}
}

// Update checks for interactable entities in the player's line of sight.
func (is *InteractionSystem) Update(world *ecs.World) *InteractionResult {
	if world == nil || is.playerEntity == 0 {
		return nil
	}

	// Get player position and angle
	posComp, ok := world.GetComponent(is.playerEntity, "Position")
	if !ok {
		return nil
	}
	playerPos := posComp.(*components.Position)

	// Cast a ray from player position in look direction
	is.currentTarget = is.findInteractableInRay(world, playerPos)
	return is.currentTarget
}

// findInteractableInRay performs a ray cast to find interactable entities.
func (is *InteractionSystem) findInteractableInRay(world *ecs.World, playerPos *components.Position) *InteractionResult {
	// Direction vector from player angle
	dirX := math.Cos(playerPos.Angle)
	dirY := math.Sin(playerPos.Angle)

	var closest *InteractionResult

	// Check all entities with Position component
	for _, entity := range world.Entities("Position") {
		if entity == is.playerEntity {
			continue
		}

		posComp, ok := world.GetComponent(entity, "Position")
		if !ok {
			continue
		}
		entPos := posComp.(*components.Position)

		// Calculate distance and direction to entity
		dx := entPos.X - playerPos.X
		dy := entPos.Y - playerPos.Y
		distance := math.Sqrt(dx*dx + dy*dy)

		// Skip if too far
		if distance > is.maxRange {
			continue
		}

		// Check if entity is roughly in front of player (dot product check)
		if distance > 0.1 {
			normalizedDX := dx / distance
			normalizedDY := dy / distance
			dot := dirX*normalizedDX + dirY*normalizedDY
			if dot < 0.5 { // Roughly 60 degree cone in front
				continue
			}
		}

		// Determine interaction type
		interactionType := is.getEntityInteractionType(world, entity)
		if interactionType == InteractionNone {
			continue
		}

		// Get entity name and prompt
		name, prompt := is.getInteractionInfo(world, entity, interactionType)

		result := &InteractionResult{
			Entity:   entity,
			Type:     interactionType,
			Distance: distance,
			Name:     name,
			Prompt:   prompt,
		}

		// Keep the closest interactable
		if closest == nil || distance < closest.Distance {
			closest = result
		}
	}

	return closest
}

// getEntityInteractionType determines what kind of interaction an entity supports.
func (is *InteractionSystem) getEntityInteractionType(world *ecs.World, entity ecs.Entity) InteractionType {
	// Check for NPC (has Schedule or DialogState)
	if _, ok := world.GetComponent(entity, "Schedule"); ok {
		return InteractionNPC
	}
	if _, ok := world.GetComponent(entity, "DialogState"); ok {
		return InteractionNPC
	}

	// Check for workbench
	if _, ok := world.GetComponent(entity, "Workbench"); ok {
		return InteractionWorkbench
	}

	// Check for inventory (items that can be picked up)
	if inv, ok := world.GetComponent(entity, "Inventory"); ok {
		inventory := inv.(*components.Inventory)
		// Loose items have items but low/no capacity (containers have capacity)
		if len(inventory.Items) > 0 && inventory.Capacity <= len(inventory.Items) {
			return InteractionItem
		}
		// Containers have capacity
		if inventory.Capacity > len(inventory.Items) {
			return InteractionContainer
		}
	}

	return InteractionNone
}

// getInteractionInfo returns the display name and prompt for an interactable entity.
func (is *InteractionSystem) getInteractionInfo(world *ecs.World, entity ecs.Entity, iType InteractionType) (name, prompt string) {
	switch iType {
	case InteractionNPC:
		name = is.getNPCName(world, entity)
		prompt = "Press E to talk to " + name
	case InteractionItem:
		name = is.getItemName(world, entity)
		prompt = "Press E to pick up " + name
	case InteractionWorkbench:
		name = is.getWorkbenchName(world, entity)
		prompt = "Press E to use " + name
	case InteractionDoor:
		name = "Door"
		prompt = "Press E to open"
	case InteractionContainer:
		name = "Container"
		prompt = "Press E to open"
	default:
		name = "Unknown"
		prompt = ""
	}
	return name, prompt
}

// getNPCName attempts to get an NPC's name from various components.
func (is *InteractionSystem) getNPCName(world *ecs.World, entity ecs.Entity) string {
	// Try to get name from NPCMemory (which might have stored name)
	if memComp, ok := world.GetComponent(entity, "NPCMemory"); ok {
		_ = memComp // Would access name if stored
	}

	// Try to get name from Faction
	if factionComp, ok := world.GetComponent(entity, "Faction"); ok {
		faction := factionComp.(*components.Faction)
		if faction.ID != "" && faction.ID != "neutral" {
			return faction.ID + " Member"
		}
	}

	// Default name based on entity ID
	return "Villager"
}

// getItemName gets the name of a pickup item.
func (is *InteractionSystem) getItemName(world *ecs.World, entity ecs.Entity) string {
	if invComp, ok := world.GetComponent(entity, "Inventory"); ok {
		inv := invComp.(*components.Inventory)
		if len(inv.Items) > 0 {
			return inv.Items[0]
		}
	}
	return "Item"
}

// getWorkbenchName gets the name of a workbench.
func (is *InteractionSystem) getWorkbenchName(world *ecs.World, entity ecs.Entity) string {
	if wbComp, ok := world.GetComponent(entity, "Workbench"); ok {
		wb := wbComp.(*components.Workbench)
		if wb.WorkbenchType != "" {
			return wb.WorkbenchType
		}
	}
	return "Workbench"
}

// GetCurrentTarget returns the currently targeted interactable entity.
func (is *InteractionSystem) GetCurrentTarget() *InteractionResult {
	return is.currentTarget
}

// ClearTarget clears the current interaction target.
func (is *InteractionSystem) ClearTarget() {
	is.currentTarget = nil
}
