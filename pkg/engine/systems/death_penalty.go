package systems

import (
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// DeathPenaltyConfig holds death penalty settings from difficulty config.
type DeathPenaltyConfig struct {
	PermaDeath        bool    // If true, death is permanent
	XPLossPercent     float64 // 0.0-1.0, portion of XP lost
	GoldLossPercent   float64 // 0.0-1.0, portion of gold lost
	DropItems         bool    // Whether items drop on death
	RespawnAtGrave    bool    // Respawn at death location or checkpoint
	DurabilityLoss    float64 // 0.0-1.0, equipment durability lost
	CorpseRetrievable bool    // Can retrieve items from corpse
}

// DeathPenaltySystem applies configurable penalties when entities die.
type DeathPenaltySystem struct {
	Config   DeathPenaltyConfig
	GameTime float64
}

// NewDeathPenaltySystem creates a new death penalty system with config.
func NewDeathPenaltySystem(config DeathPenaltyConfig) *DeathPenaltySystem {
	return &DeathPenaltySystem{
		Config:   config,
		GameTime: 0,
	}
}

// NewDefaultDeathPenaltySystem creates a death penalty system with default (normal) settings.
func NewDefaultDeathPenaltySystem() *DeathPenaltySystem {
	return &DeathPenaltySystem{
		Config: DeathPenaltyConfig{
			PermaDeath:        false,
			XPLossPercent:     0.1, // Lose 10% XP
			GoldLossPercent:   0.1, // Lose 10% gold
			DropItems:         false,
			RespawnAtGrave:    false, // Respawn at checkpoint
			DurabilityLoss:    0.1,   // 10% durability loss
			CorpseRetrievable: true,
		},
		GameTime: 0,
	}
}

// Update checks for dead entities and applies penalties.
func (s *DeathPenaltySystem) Update(w *ecs.World, dt float64) {
	s.GameTime += dt

	// Find entities that just died
	for _, e := range w.Entities("Health") {
		healthComp, ok := w.GetComponent(e, "Health")
		if !ok {
			continue
		}
		health := healthComp.(*components.Health)

		// Check if entity is dead and hasn't been processed
		if health.Current <= 0 && !s.hasDeathProcessed(w, e) {
			s.applyDeathPenalties(w, e)
			s.markDeathProcessed(w, e)
		}
	}
}

// hasDeathProcessed checks if death penalties have been applied.
func (s *DeathPenaltySystem) hasDeathProcessed(w *ecs.World, entity ecs.Entity) bool {
	// Check for DeathState component
	deathComp, ok := w.GetComponent(entity, "DeathState")
	if !ok {
		return false
	}
	death := deathComp.(*components.DeathState)
	return death.PenaltiesApplied
}

// markDeathProcessed marks that death penalties have been applied.
func (s *DeathPenaltySystem) markDeathProcessed(w *ecs.World, entity ecs.Entity) {
	deathComp, ok := w.GetComponent(entity, "DeathState")
	if !ok {
		// Create DeathState if not exists
		w.AddComponent(entity, &components.DeathState{
			IsDead:           true,
			DeathTime:        s.GameTime,
			PenaltiesApplied: true,
		})
		return
	}
	death := deathComp.(*components.DeathState)
	death.IsDead = true
	death.DeathTime = s.GameTime
	death.PenaltiesApplied = true
}

// applyDeathPenalties applies all configured penalties to the dead entity.
func (s *DeathPenaltySystem) applyDeathPenalties(w *ecs.World, entity ecs.Entity) {
	// Apply XP loss
	if s.Config.XPLossPercent > 0 {
		s.applyXPLoss(w, entity)
	}

	// Apply gold loss
	if s.Config.GoldLossPercent > 0 {
		s.applyGoldLoss(w, entity)
	}

	// Apply equipment durability loss
	if s.Config.DurabilityLoss > 0 {
		s.applyDurabilityLoss(w, entity)
	}

	// Create corpse for item retrieval if configured
	if s.Config.DropItems && s.Config.CorpseRetrievable {
		s.createCorpse(w, entity)
	}
}

// applyXPLoss reduces entity's experience points.
func (s *DeathPenaltySystem) applyXPLoss(w *ecs.World, entity ecs.Entity) {
	skillsComp, ok := w.GetComponent(entity, "Skills")
	if !ok {
		return
	}
	skills := skillsComp.(*components.Skills)

	// Reduce XP in all skills
	if skills.Experience != nil {
		for skillID, xp := range skills.Experience {
			lossAmount := xp * s.Config.XPLossPercent
			skills.Experience[skillID] = xp - lossAmount
			if skills.Experience[skillID] < 0 {
				skills.Experience[skillID] = 0
			}
		}
	}
}

// applyGoldLoss reduces entity's gold.
func (s *DeathPenaltySystem) applyGoldLoss(w *ecs.World, entity ecs.Entity) {
	currencyComp, ok := w.GetComponent(entity, "Currency")
	if !ok {
		return
	}
	currency := currencyComp.(*components.Currency)

	// Reduce gold
	if currency.Gold > 0 {
		lossAmount := int(float64(currency.Gold) * s.Config.GoldLossPercent)
		currency.Gold -= lossAmount
		if currency.Gold < 0 {
			currency.Gold = 0
		}
	}
}

// applyDurabilityLoss reduces equipment durability.
func (s *DeathPenaltySystem) applyDurabilityLoss(w *ecs.World, entity ecs.Entity) {
	equipmentComp, ok := w.GetComponent(entity, "Equipment")
	if !ok {
		return
	}
	equipment := equipmentComp.(*components.Equipment)

	// Reduce durability of all equipped items
	for _, slot := range equipment.Slots {
		if slot.Durability > 0 {
			loss := slot.MaxDurability * s.Config.DurabilityLoss
			slot.Durability -= loss
			if slot.Durability < 0 {
				slot.Durability = 0
			}
		}
	}
}

// createCorpse creates a corpse entity with dropped items.
func (s *DeathPenaltySystem) createCorpse(w *ecs.World, entity ecs.Entity) {
	// Get position for corpse
	posComp, ok := w.GetComponent(entity, "Position")
	if !ok {
		return
	}
	pos := posComp.(*components.Position)

	// Create corpse entity
	corpse := w.CreateEntity()
	w.AddComponent(corpse, &components.Position{X: pos.X, Y: pos.Y, Z: pos.Z})
	w.AddComponent(corpse, &components.Corpse{
		OwnerEntity: uint64(entity),
		DeathTime:   s.GameTime,
		DecayTime:   s.GameTime + 300, // 5 minutes
	})

	// Move items to corpse if DropItems is enabled
	if s.Config.DropItems {
		s.moveItemsToCorpse(w, entity, corpse)
	}
}

// moveItemsToCorpse transfers items from dead entity to corpse.
func (s *DeathPenaltySystem) moveItemsToCorpse(w *ecs.World, source, corpse ecs.Entity) {
	inventoryComp, ok := w.GetComponent(source, "Inventory")
	if !ok {
		return
	}
	inventory := inventoryComp.(*components.Inventory)

	// Create corpse inventory
	corpseInventory := &components.Inventory{
		Items:    inventory.Items,
		Capacity: inventory.Capacity,
	}
	w.AddComponent(corpse, corpseInventory)

	// Clear source inventory
	inventory.Items = nil
}

// SetConfig updates the death penalty configuration.
func (s *DeathPenaltySystem) SetConfig(config DeathPenaltyConfig) {
	s.Config = config
}

// IsPermaDeath returns whether permanent death is enabled.
func (s *DeathPenaltySystem) IsPermaDeath() bool {
	return s.Config.PermaDeath
}
