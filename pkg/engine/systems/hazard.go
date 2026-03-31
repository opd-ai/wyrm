// Package systems contains ECS systems for Wyrm.
package systems

import (
	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// HazardSystem handles environmental hazard damage and effects.
type HazardSystem struct {
	// Genre affects hazard behavior and naming.
	Genre string
	// DamageMultiplier scales all hazard damage.
	DamageMultiplier float64
	// WorldClock provides current time for hazard timing.
	WorldClock *WorldClockSystem
	// IndoorChecker checks if positions are indoors (optional).
	IndoorChecker IndoorChecker
}

// IndoorChecker is an interface for checking if a position is indoors.
type IndoorChecker interface {
	IsIndoors(x, y, z float64) bool
}

// NewHazardSystem creates a new hazard system.
func NewHazardSystem(genre string) *HazardSystem {
	return &HazardSystem{
		Genre:            genre,
		DamageMultiplier: 1.0,
	}
}

// Update processes environmental hazards each tick.
func (s *HazardSystem) Update(w *ecs.World, dt float64) {
	// Process hazard zones affecting entities
	s.processHazardZones(w, dt)

	// Process active hazard effects on entities
	s.processHazardEffects(w, dt)

	// Process traps
	s.processTraps(w, dt)

	// Process weather hazards
	s.processWeatherHazards(w, dt)

	// Update temporary hazards
	s.updateTemporaryHazards(w, dt)
}

// processHazardZones checks entities in hazard zones and applies effects.
func (s *HazardSystem) processHazardZones(w *ecs.World, dt float64) {
	hazards := w.Entities("EnvironmentalHazard", "Position")
	targets := w.Entities("Position", "Health")

	for _, hazardEnt := range hazards {
		s.processHazardEntity(w, hazardEnt, targets, dt)
	}
}

// processHazardEntity handles damage application for a single hazard.
func (s *HazardSystem) processHazardEntity(w *ecs.World, hazardEnt ecs.Entity, targets []ecs.Entity, dt float64) {
	hazardComp, _ := w.GetComponent(hazardEnt, "EnvironmentalHazard")
	hazard := hazardComp.(*components.EnvironmentalHazard)

	if !hazard.Active || s.isOnCooldown(hazard) {
		return
	}

	hazardPos := s.getEntityPosition(w, hazardEnt)
	s.damageTargetsInRadius(w, hazardEnt, hazardPos, hazard, targets, dt)
	hazard.LastDamageTick = s.getCurrentTime()
}

// isOnCooldown checks if a hazard is still in its cooldown period.
func (s *HazardSystem) isOnCooldown(hazard *components.EnvironmentalHazard) bool {
	return hazard.LastDamageTick+hazard.CooldownTime > s.getCurrentTime()
}

// getEntityPosition retrieves an entity's position component.
func (s *HazardSystem) getEntityPosition(w *ecs.World, entity ecs.Entity) *components.Position {
	posComp, _ := w.GetComponent(entity, "Position")
	return posComp.(*components.Position)
}

// damageTargetsInRadius applies hazard damage to all valid targets within radius.
func (s *HazardSystem) damageTargetsInRadius(w *ecs.World, hazardEnt ecs.Entity, hazardPos *components.Position, hazard *components.EnvironmentalHazard, targets []ecs.Entity, dt float64) {
	for _, targetEnt := range targets {
		if targetEnt == hazardEnt {
			continue
		}
		if s.isInHazardRadius(w, targetEnt, hazardPos, hazard.Radius) {
			s.applyHazardDamage(w, targetEnt, hazard, dt)
		}
	}
}

// isInHazardRadius checks if a target is within the hazard's effect radius.
func (s *HazardSystem) isInHazardRadius(w *ecs.World, target ecs.Entity, hazardPos *components.Position, radius float64) bool {
	targetPos := s.getEntityPosition(w, target)
	dx := hazardPos.X - targetPos.X
	dy := hazardPos.Y - targetPos.Y
	dz := hazardPos.Z - targetPos.Z
	distSq := dx*dx + dy*dy + dz*dz
	return distSq <= radius*radius
}

// applyHazardDamage deals hazard damage to a target entity.
func (s *HazardSystem) applyHazardDamage(w *ecs.World, target ecs.Entity, hazard *components.EnvironmentalHazard, dt float64) {
	healthComp, ok := w.GetComponent(target, "Health")
	if !ok {
		return
	}
	health := healthComp.(*components.Health)

	damage := hazard.DamagePerSecond * hazard.Intensity * dt * s.DamageMultiplier

	// Check for resistance
	resistComp, hasResist := w.GetComponent(target, "HazardResistance")
	if hasResist {
		resist := resistComp.(*components.HazardResistance)
		resistance := resist.GetResistance(hazard.HazardType)
		damage *= (1.0 - resistance)
	}

	// Apply damage
	health.Current -= damage
	if health.Current < 0 {
		health.Current = 0
	}
}

// processHazardEffects updates ongoing hazard effects on entities.
func (s *HazardSystem) processHazardEffects(w *ecs.World, dt float64) {
	entities := w.Entities("HazardEffect", "Health")

	for _, ent := range entities {
		effectComp, _ := w.GetComponent(ent, "HazardEffect")
		effect := effectComp.(*components.HazardEffect)

		// Apply damage over time
		if effect.DamageOverTime > 0 {
			healthComp, _ := w.GetComponent(ent, "Health")
			health := healthComp.(*components.Health)

			damage := effect.DamageOverTime * float64(effect.StackCount) * dt * s.DamageMultiplier
			health.Current -= damage
			if health.Current < 0 {
				health.Current = 0
			}
		}

		// Reduce duration
		effect.RemainingDuration -= dt

		// Remove expired effects
		if effect.RemainingDuration <= 0 {
			w.RemoveComponent(ent, "HazardEffect")
		}
	}
}

// processTraps handles trap detection and triggering.
func (s *HazardSystem) processTraps(w *ecs.World, dt float64) {
	traps := w.Entities("TrapMechanism", "Position")
	targets := w.Entities("Position", "Health")

	for _, trapEnt := range traps {
		trapComp, _ := w.GetComponent(trapEnt, "TrapMechanism")
		trap := trapComp.(*components.TrapMechanism)

		if s.updateTrapReset(trap, dt) {
			continue
		}

		trapPosComp, _ := w.GetComponent(trapEnt, "Position")
		trapPos := trapPosComp.(*components.Position)
		s.checkTrapTriggers(w, trapEnt, trapPos, trap, targets)
	}
}

// updateTrapReset handles trap reset timing, returns true if trap is inactive.
func (s *HazardSystem) updateTrapReset(trap *components.TrapMechanism, dt float64) bool {
	if !trap.Armed || trap.Triggered {
		if trap.Triggered && trap.ResetTime > 0 {
			trap.ResetTime -= dt
			if trap.ResetTime <= 0 {
				trap.Triggered = false
				trap.Armed = true
			}
		}
		return true
	}
	return false
}

// checkTrapTriggers checks if any target triggers the trap.
func (s *HazardSystem) checkTrapTriggers(w *ecs.World, trapEnt ecs.Entity, trapPos *components.Position, trap *components.TrapMechanism, targets []ecs.Entity) {
	for _, targetEnt := range targets {
		if targetEnt == trapEnt {
			continue
		}

		targetPosComp, _ := w.GetComponent(targetEnt, "Position")
		targetPos := targetPosComp.(*components.Position)

		if s.isInTriggerRadius(trapPos, targetPos, trap.TriggerRadius) {
			s.triggerTrap(w, trapEnt, targetEnt, trap)
			return
		}
	}
}

// isInTriggerRadius checks if a target is within the trap's trigger radius.
func (s *HazardSystem) isInTriggerRadius(trapPos, targetPos *components.Position, triggerRadius float64) bool {
	dx := trapPos.X - targetPos.X
	dy := trapPos.Y - targetPos.Y
	distSq := dx*dx + dy*dy
	return distSq <= triggerRadius*triggerRadius
}

// triggerTrap activates a trap and deals damage.
func (s *HazardSystem) triggerTrap(w *ecs.World, trapEnt, targetEnt ecs.Entity, trap *components.TrapMechanism) {
	trap.Triggered = true
	trap.Armed = false

	healthComp, ok := w.GetComponent(targetEnt, "Health")
	if !ok {
		return
	}
	health := healthComp.(*components.Health)

	damage := trap.Damage * s.DamageMultiplier
	health.Current -= damage
	if health.Current < 0 {
		health.Current = 0
	}
}

// processWeatherHazards handles weather-based environmental damage.
func (s *HazardSystem) processWeatherHazards(w *ecs.World, dt float64) {
	weatherEnts := w.Entities("WeatherHazard")

	for _, weatherEnt := range weatherEnts {
		weatherComp, _ := w.GetComponent(weatherEnt, "WeatherHazard")
		weather := weatherComp.(*components.WeatherHazard)

		// Update duration
		weather.Duration -= dt

		// Remove expired weather
		if weather.Duration <= 0 {
			w.RemoveComponent(weatherEnt, "WeatherHazard")
			continue
		}

		// Apply damage to outdoor entities
		if weather.OutdoorDamage > 0 {
			s.applyWeatherDamage(w, weather, dt)
		}
	}
}

// applyWeatherDamage damages entities exposed to hazardous weather.
func (s *HazardSystem) applyWeatherDamage(w *ecs.World, weather *components.WeatherHazard, dt float64) {
	entities := w.Entities("Position", "Health")

	for _, ent := range entities {
		posComp, _ := w.GetComponent(ent, "Position")
		pos := posComp.(*components.Position)

		// Check if entity is indoors (shelter check)
		if s.isEntitySheltered(pos.X, pos.Y, pos.Z) {
			continue // Skip damage for sheltered entities
		}

		healthComp, _ := w.GetComponent(ent, "Health")
		health := healthComp.(*components.Health)

		damage := weather.OutdoorDamage * weather.Severity * dt * s.DamageMultiplier
		health.Current -= damage
		if health.Current < 0 {
			health.Current = 0
		}
	}
}

// isEntitySheltered checks if an entity at a position is sheltered from weather.
func (s *HazardSystem) isEntitySheltered(x, y, z float64) bool {
	if s.IndoorChecker == nil {
		return false // Default to outdoors if no checker available
	}
	return s.IndoorChecker.IsIndoors(x, y, z)
}

// updateTemporaryHazards reduces duration of temporary hazards.
func (s *HazardSystem) updateTemporaryHazards(w *ecs.World, dt float64) {
	hazards := w.Entities("EnvironmentalHazard")

	for _, ent := range hazards {
		hazardComp, _ := w.GetComponent(ent, "EnvironmentalHazard")
		hazard := hazardComp.(*components.EnvironmentalHazard)

		if hazard.Permanent {
			continue
		}

		hazard.Duration -= dt

		if hazard.Duration <= 0 {
			hazard.Active = false
		}
	}
}

// getCurrentTime returns the current world time.
func (s *HazardSystem) getCurrentTime() float64 {
	if s.WorldClock != nil {
		return s.WorldClock.ElapsedTime()
	}
	return 0
}

// CreateHazard creates a new environmental hazard entity.
func (s *HazardSystem) CreateHazard(w *ecs.World, hazardType components.HazardType, x, y, z, radius, intensity float64) ecs.Entity {
	ent := w.CreateEntity()

	w.AddComponent(ent, &components.Position{X: x, Y: y, Z: z})
	w.AddComponent(ent, &components.EnvironmentalHazard{
		HazardType:      hazardType,
		Intensity:       intensity,
		DamagePerSecond: s.getBaseDamageForHazard(hazardType),
		Radius:          radius,
		Active:          true,
		Visible:         true,
		Permanent:       false,
		Duration:        60.0, // Default 60 second duration
		CooldownTime:    1.0,  // Damage tick every second
	})

	return ent
}

// CreatePermanentHazard creates a permanent environmental hazard.
func (s *HazardSystem) CreatePermanentHazard(w *ecs.World, hazardType components.HazardType, x, y, z, radius, intensity float64) ecs.Entity {
	ent := s.CreateHazard(w, hazardType, x, y, z, radius, intensity)

	hazardComp, _ := w.GetComponent(ent, "EnvironmentalHazard")
	hazard := hazardComp.(*components.EnvironmentalHazard)
	hazard.Permanent = true

	return ent
}

// CreateTrap creates a new trap entity.
func (s *HazardSystem) CreateTrap(w *ecs.World, trapType string, x, y, z, damage, triggerRadius float64) ecs.Entity {
	ent := w.CreateEntity()

	w.AddComponent(ent, &components.Position{X: x, Y: y, Z: z})
	w.AddComponent(ent, &components.TrapMechanism{
		TrapType:            trapType,
		Triggered:           false,
		Armed:               true,
		DetectionDifficulty: 0.5,
		DisarmDifficulty:    0.5,
		Damage:              damage,
		ResetTime:           0, // Single use by default
		TriggerRadius:       triggerRadius,
	})

	return ent
}

// getBaseDamageForHazard returns the default DPS for a hazard type.
func (s *HazardSystem) getBaseDamageForHazard(hazardType components.HazardType) float64 {
	switch hazardType {
	case components.HazardTypeFire:
		return 10.0
	case components.HazardTypeLava:
		return 50.0
	case components.HazardTypeRadiation:
		return 5.0
	case components.HazardTypePoison:
		return 8.0
	case components.HazardTypeAcid:
		return 15.0
	case components.HazardTypeElectric:
		return 20.0
	case components.HazardTypeFreeze:
		return 7.0
	case components.HazardTypeMagic:
		return 12.0
	case components.HazardTypeGas:
		return 6.0
	default:
		return 10.0
	}
}

// GetGenreHazardName returns a genre-appropriate name for a hazard type.
func (s *HazardSystem) GetGenreHazardName(hazardType components.HazardType) string {
	names := map[components.HazardType]map[string]string{
		components.HazardTypeRadiation: {
			"fantasy":          "Cursed Ground",
			"sci-fi":           "Radiation Zone",
			"horror":           "Necrotic Field",
			"cyberpunk":        "Toxic Waste",
			"post-apocalyptic": "Hot Zone",
		},
		components.HazardTypeFire: {
			"fantasy":          "Dragon Fire",
			"sci-fi":           "Plasma Fire",
			"horror":           "Hellfire",
			"cyberpunk":        "Napalm",
			"post-apocalyptic": "Wildfire",
		},
		components.HazardTypeElectric: {
			"fantasy":          "Lightning Storm",
			"sci-fi":           "EMP Field",
			"horror":           "Shock Horror",
			"cyberpunk":        "Live Wire",
			"post-apocalyptic": "Power Surge",
		},
	}

	if typeNames, ok := names[hazardType]; ok {
		if name, ok := typeNames[s.Genre]; ok {
			return name
		}
	}

	return string(hazardType)
}
