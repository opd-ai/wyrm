package systems

import (
	"fmt"
	"math"
	"sync"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

// VehicleSystem manages vehicle movement and physics.
type VehicleSystem struct{}

// Update processes vehicle physics each tick.
func (s *VehicleSystem) Update(w *ecs.World, dt float64) {
	for _, e := range w.Entities("Vehicle", "Position") {
		s.updateVehicle(w, e, dt)
	}
}

// updateVehicle updates a single vehicle's position based on its physics.
func (s *VehicleSystem) updateVehicle(w *ecs.World, e ecs.Entity, dt float64) {
	vehicle, pos := s.getVehicleComponents(w, e)
	if vehicle == nil || pos == nil {
		return
	}
	if vehicle.Fuel > 0 && vehicle.Speed > 0 {
		s.applyVehicleMovement(vehicle, pos, dt)
	}
}

// getVehicleComponents retrieves vehicle and position components for an entity.
func (s *VehicleSystem) getVehicleComponents(w *ecs.World, e ecs.Entity) (*components.Vehicle, *components.Position) {
	vComp, ok := w.GetComponent(e, "Vehicle")
	if !ok {
		return nil, nil
	}
	pComp, ok := w.GetComponent(e, "Position")
	if !ok {
		return nil, nil
	}
	return vComp.(*components.Vehicle), pComp.(*components.Position)
}

// applyVehicleMovement updates position and consumes fuel.
func (s *VehicleSystem) applyVehicleMovement(vehicle *components.Vehicle, pos *components.Position, dt float64) {
	pos.X += math.Cos(vehicle.Direction) * vehicle.Speed * dt
	pos.Y += math.Sin(vehicle.Direction) * vehicle.Speed * dt
	vehicle.Fuel -= vehicle.Speed * dt * DefaultFuelConsumptionRate
	if vehicle.Fuel < MinFuelLevel {
		vehicle.Fuel = MinFuelLevel
	}
}

// VehicleWeaponType represents a type of vehicle weapon.
type VehicleWeaponType int

const (
	WeaponMachineGun VehicleWeaponType = iota
	WeaponCannon
	WeaponMissile
	WeaponLaser
	WeaponRam // For ramming attacks
)

// VehicleWeapon represents a weapon mounted on a vehicle.
type VehicleWeapon struct {
	Type         VehicleWeaponType
	Damage       float64
	Range        float64
	RateOfFire   float64 // Shots per second
	AmmoCapacity int
	AmmoCount    int
	LastFired    float64 // Time since last shot
	MountAngle   float64 // Angle offset from vehicle facing
	Enabled      bool
}

// VehicleCombatState represents the combat state of a vehicle.
type VehicleCombatState struct {
	EntityID    ecs.Entity
	Weapons     []*VehicleWeapon
	Armor       float64 // Damage reduction percentage
	ShieldPower float64 // Energy shield (0-100)
	ShieldRegen float64 // Shield regen per second
	Health      float64
	MaxHealth   float64
	InCombat    bool
	LastDamaged float64
	RamDamage   float64 // Damage dealt when ramming
	RamCooldown float64
}

// VehicleCombatSystem manages vehicle-to-vehicle combat.
type VehicleCombatSystem struct {
	vehicles map[ecs.Entity]*VehicleCombatState
}

// NewVehicleCombatSystem creates a new vehicle combat system.
func NewVehicleCombatSystem() *VehicleCombatSystem {
	return &VehicleCombatSystem{
		vehicles: make(map[ecs.Entity]*VehicleCombatState),
	}
}

// RegisterVehicle registers a vehicle for combat.
func (s *VehicleCombatSystem) RegisterVehicle(entity ecs.Entity, health, armor float64) *VehicleCombatState {
	state := &VehicleCombatState{
		EntityID:    entity,
		Weapons:     make([]*VehicleWeapon, 0),
		Armor:       armor,
		Health:      health,
		MaxHealth:   health,
		ShieldPower: 0,
		ShieldRegen: 5.0,
		RamDamage:   20.0,
		RamCooldown: 0,
	}
	s.vehicles[entity] = state
	return state
}

// AddWeapon adds a weapon to a vehicle.
func (s *VehicleCombatSystem) AddWeapon(entity ecs.Entity, weaponType VehicleWeaponType) bool {
	state, ok := s.vehicles[entity]
	if !ok {
		return false
	}
	weapon := s.createWeapon(weaponType)
	state.Weapons = append(state.Weapons, weapon)
	return true
}

// createWeapon creates a weapon with default stats based on type.
func (s *VehicleCombatSystem) createWeapon(weaponType VehicleWeaponType) *VehicleWeapon {
	switch weaponType {
	case WeaponMachineGun:
		return &VehicleWeapon{
			Type: WeaponMachineGun, Damage: 5, Range: 150,
			RateOfFire: 10, AmmoCapacity: 200, AmmoCount: 200, Enabled: true,
		}
	case WeaponCannon:
		return &VehicleWeapon{
			Type: WeaponCannon, Damage: 40, Range: 300,
			RateOfFire: 0.5, AmmoCapacity: 20, AmmoCount: 20, Enabled: true,
		}
	case WeaponMissile:
		return &VehicleWeapon{
			Type: WeaponMissile, Damage: 80, Range: 500,
			RateOfFire: 0.2, AmmoCapacity: 4, AmmoCount: 4, Enabled: true,
		}
	case WeaponLaser:
		return &VehicleWeapon{
			Type: WeaponLaser, Damage: 15, Range: 250,
			RateOfFire: 5, AmmoCapacity: 100, AmmoCount: 100, Enabled: true,
		}
	default:
		return &VehicleWeapon{
			Type: WeaponRam, Damage: 20, Range: 10,
			RateOfFire: 0.3, AmmoCapacity: -1, AmmoCount: -1, Enabled: true,
		}
	}
}

// Fire attempts to fire a weapon at a target.
func (s *VehicleCombatSystem) Fire(attacker ecs.Entity, weaponIdx int, target ecs.Entity, dist float64) float64 {
	state, ok := s.vehicles[attacker]
	if !ok || weaponIdx >= len(state.Weapons) {
		return 0
	}
	weapon := state.Weapons[weaponIdx]
	if !weapon.Enabled || weapon.AmmoCount == 0 || dist > weapon.Range {
		return 0
	}
	cooldown := 1.0 / weapon.RateOfFire
	if weapon.LastFired < cooldown {
		return 0
	}
	weapon.LastFired = 0
	if weapon.AmmoCount > 0 {
		weapon.AmmoCount--
	}
	return s.ApplyDamage(target, weapon.Damage)
}

// ApplyDamage applies damage to a vehicle, respecting armor and shields.
func (s *VehicleCombatSystem) ApplyDamage(entity ecs.Entity, damage float64) float64 {
	state, ok := s.vehicles[entity]
	if !ok {
		return 0
	}
	state.InCombat = true
	state.LastDamaged = 0
	// Shields absorb damage first
	if state.ShieldPower > 0 {
		if state.ShieldPower >= damage {
			state.ShieldPower -= damage
			return 0
		}
		damage -= state.ShieldPower
		state.ShieldPower = 0
	}
	// Apply armor reduction
	effectiveDamage := damage * (1.0 - state.Armor/100.0)
	state.Health -= effectiveDamage
	if state.Health < 0 {
		state.Health = 0
	}
	return effectiveDamage
}

// ProcessRam handles a ramming attack between vehicles.
func (s *VehicleCombatSystem) ProcessRam(attacker, target ecs.Entity, attackerSpeed float64) (float64, float64) {
	attackerState, ok1 := s.vehicles[attacker]
	_, ok2 := s.vehicles[target]
	if !ok1 || !ok2 {
		return 0, 0
	}
	if attackerState.RamCooldown > 0 {
		return 0, 0
	}
	attackerState.RamCooldown = 3.0 // 3 second cooldown
	// Damage scales with speed
	speedMultiplier := attackerSpeed / 50.0 // 50 units/sec = 1x damage
	attackerDamage := attackerState.RamDamage * speedMultiplier
	// Both vehicles take damage, attacker takes less
	targetDamage := s.ApplyDamage(target, attackerDamage)
	selfDamage := s.ApplyDamage(attacker, attackerDamage*0.3)
	return targetDamage, selfDamage
}

// Update processes combat state each tick.
func (s *VehicleCombatSystem) Update(w *ecs.World, dt float64) {
	for _, state := range s.vehicles {
		s.updateVehicleState(state, dt)
	}
}

// updateVehicleState updates cooldowns, shields, and combat status for one vehicle.
func (s *VehicleCombatSystem) updateVehicleState(state *VehicleCombatState, dt float64) {
	s.updateWeaponCooldowns(state, dt)
	s.updateRamCooldown(state, dt)
	s.regenerateShields(state, dt)
	s.updateCombatStatus(state, dt)
}

// updateWeaponCooldowns advances weapon cooldown timers.
func (s *VehicleCombatSystem) updateWeaponCooldowns(state *VehicleCombatState, dt float64) {
	for _, weapon := range state.Weapons {
		weapon.LastFired += dt
	}
}

// updateRamCooldown reduces ram cooldown timer.
func (s *VehicleCombatSystem) updateRamCooldown(state *VehicleCombatState, dt float64) {
	if state.RamCooldown > 0 {
		state.RamCooldown -= dt
	}
}

// regenerateShields restores shield power when out of combat.
func (s *VehicleCombatSystem) regenerateShields(state *VehicleCombatState, dt float64) {
	if state.ShieldPower < 100 && !state.InCombat {
		state.ShieldPower += state.ShieldRegen * dt
		if state.ShieldPower > 100 {
			state.ShieldPower = 100
		}
	}
}

// updateCombatStatus exits combat after no damage for 5 seconds.
func (s *VehicleCombatSystem) updateCombatStatus(state *VehicleCombatState, dt float64) {
	state.LastDamaged += dt
	if state.LastDamaged > 5.0 {
		state.InCombat = false
	}
}

// IsDestroyed returns whether a vehicle is destroyed.
func (s *VehicleCombatSystem) IsDestroyed(entity ecs.Entity) bool {
	state, ok := s.vehicles[entity]
	return ok && state.Health <= 0
}

// GetHealth returns the current health of a vehicle.
func (s *VehicleCombatSystem) GetHealth(entity ecs.Entity) float64 {
	state, ok := s.vehicles[entity]
	if !ok {
		return 0
	}
	return state.Health
}
