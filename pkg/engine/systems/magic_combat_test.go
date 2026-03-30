package systems

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

func TestMagicSystem_ManaRegeneration(t *testing.T) {
	w := ecs.NewWorld()
	magicSys := NewMagicSystem()

	// Create entity with mana
	entity := w.CreateEntity()
	mana := &components.Mana{
		Current:   50,
		Max:       100,
		RegenRate: 5.0,
	}
	w.AddComponent(entity, mana)

	// Simulate 2 seconds
	magicSys.Update(w, 2.0)

	// Should regenerate 10 mana (5 * 2)
	if mana.Current != 60 {
		t.Errorf("Expected mana 60, got %f", mana.Current)
	}
}

func TestMagicSystem_ManaRegenCap(t *testing.T) {
	w := ecs.NewWorld()
	magicSys := NewMagicSystem()

	entity := w.CreateEntity()
	mana := &components.Mana{
		Current:   95,
		Max:       100,
		RegenRate: 10.0,
	}
	w.AddComponent(entity, mana)

	magicSys.Update(w, 1.0)

	// Should cap at max
	if mana.Current != 100 {
		t.Errorf("Expected mana capped at 100, got %f", mana.Current)
	}
}

func TestMagicSystem_CastSpell_Basic(t *testing.T) {
	w := ecs.NewWorld()
	magicSys := NewMagicSystem()

	// Create caster
	caster := w.CreateEntity()
	w.AddComponent(caster, &components.Position{X: 0, Y: 0, Z: 0})
	w.AddComponent(caster, &components.Mana{Current: 100, Max: 100, RegenRate: 1.0})
	spellbook := &components.Spellbook{
		Spells: map[string]*components.Spell{
			"fireball": {
				ID:         "fireball",
				Name:       "Fireball",
				ManaCost:   20,
				Cooldown:   1.0,
				Range:      30.0,
				Magnitude:  50,
				EffectType: "damage",
				LastCast:   -1,
			},
		},
	}
	w.AddComponent(caster, spellbook)

	// Create target
	target := w.CreateEntity()
	w.AddComponent(target, &components.Position{X: 5, Y: 0, Z: 0})
	health := &components.Health{Current: 100, Max: 100}
	w.AddComponent(target, health)

	// Cast spell
	success := magicSys.CastSpell(w, caster, "fireball", target, nil)
	if !success {
		t.Error("Spell cast should succeed")
	}

	// Check mana consumed
	manaComp, _ := w.GetComponent(caster, "Mana")
	mana := manaComp.(*components.Mana)
	if mana.Current != 80 {
		t.Errorf("Expected mana 80 after spell, got %f", mana.Current)
	}

	// Check damage dealt
	if health.Current != 50 {
		t.Errorf("Expected health 50 after fireball, got %f", health.Current)
	}
}

func TestMagicSystem_CastSpell_InsufficientMana(t *testing.T) {
	w := ecs.NewWorld()
	magicSys := NewMagicSystem()

	caster := w.CreateEntity()
	w.AddComponent(caster, &components.Position{X: 0, Y: 0, Z: 0})
	w.AddComponent(caster, &components.Mana{Current: 10, Max: 100, RegenRate: 1.0})
	spellbook := &components.Spellbook{
		Spells: map[string]*components.Spell{
			"fireball": {
				ID:       "fireball",
				ManaCost: 20,
				Cooldown: 1.0,
				Range:    30.0,
			},
		},
	}
	w.AddComponent(caster, spellbook)

	target := w.CreateEntity()
	w.AddComponent(target, &components.Position{X: 5, Y: 0, Z: 0})

	success := magicSys.CastSpell(w, caster, "fireball", target, nil)
	if success {
		t.Error("Spell cast should fail with insufficient mana")
	}
}

func TestMagicSystem_CastSpell_Cooldown(t *testing.T) {
	w := ecs.NewWorld()
	magicSys := NewMagicSystem()

	caster := w.CreateEntity()
	w.AddComponent(caster, &components.Position{X: 0, Y: 0, Z: 0})
	w.AddComponent(caster, &components.Mana{Current: 100, Max: 100, RegenRate: 1.0})
	spellbook := &components.Spellbook{
		Spells: map[string]*components.Spell{
			"fireball": {
				ID:         "fireball",
				ManaCost:   20,
				Cooldown:   2.0,
				Range:      30.0,
				Magnitude:  50,
				EffectType: "damage",
				LastCast:   -1, // Never cast before
			},
		},
	}
	w.AddComponent(caster, spellbook)

	target := w.CreateEntity()
	w.AddComponent(target, &components.Position{X: 5, Y: 0, Z: 0})
	w.AddComponent(target, &components.Health{Current: 100, Max: 100})

	// First cast should succeed
	success1 := magicSys.CastSpell(w, caster, "fireball", target, nil)
	if !success1 {
		t.Error("First cast should succeed")
	}

	// Immediate second cast should fail (cooldown)
	success2 := magicSys.CastSpell(w, caster, "fireball", target, nil)
	if success2 {
		t.Error("Second cast should fail due to cooldown")
	}

	// Advance time past cooldown
	magicSys.Update(w, 2.5)

	// Now cast should succeed again
	success3 := magicSys.CastSpell(w, caster, "fireball", target, nil)
	if !success3 {
		t.Error("Third cast should succeed after cooldown")
	}
}

func TestMagicSystem_CastSpell_OutOfRange(t *testing.T) {
	w := ecs.NewWorld()
	magicSys := NewMagicSystem()

	caster := w.CreateEntity()
	w.AddComponent(caster, &components.Position{X: 0, Y: 0, Z: 0})
	w.AddComponent(caster, &components.Mana{Current: 100, Max: 100, RegenRate: 1.0})
	spellbook := &components.Spellbook{
		Spells: map[string]*components.Spell{
			"fireball": {
				ID:       "fireball",
				ManaCost: 20,
				Cooldown: 1.0,
				Range:    10.0, // Short range
			},
		},
	}
	w.AddComponent(caster, spellbook)

	target := w.CreateEntity()
	w.AddComponent(target, &components.Position{X: 50, Y: 0, Z: 0}) // Too far

	success := magicSys.CastSpell(w, caster, "fireball", target, nil)
	if success {
		t.Error("Spell cast should fail when target out of range")
	}
}

func TestMagicSystem_HealSpell(t *testing.T) {
	w := ecs.NewWorld()
	magicSys := NewMagicSystem()

	caster := w.CreateEntity()
	w.AddComponent(caster, &components.Position{X: 0, Y: 0, Z: 0})
	w.AddComponent(caster, &components.Mana{Current: 100, Max: 100, RegenRate: 1.0})
	spellbook := &components.Spellbook{
		Spells: map[string]*components.Spell{
			"heal": {
				ID:         "heal",
				ManaCost:   15,
				Cooldown:   1.0,
				Range:      20.0,
				Magnitude:  30,
				EffectType: "heal",
				LastCast:   -1,
			},
		},
	}
	w.AddComponent(caster, spellbook)

	target := w.CreateEntity()
	w.AddComponent(target, &components.Position{X: 5, Y: 0, Z: 0})
	health := &components.Health{Current: 50, Max: 100}
	w.AddComponent(target, health)

	success := magicSys.CastSpell(w, caster, "heal", target, nil)
	if !success {
		t.Error("Heal cast should succeed")
	}

	if health.Current != 80 {
		t.Errorf("Expected health 80 after heal, got %f", health.Current)
	}
}

func TestMagicSystem_HealSpell_Cap(t *testing.T) {
	w := ecs.NewWorld()
	magicSys := NewMagicSystem()

	caster := w.CreateEntity()
	w.AddComponent(caster, &components.Position{X: 0, Y: 0, Z: 0})
	w.AddComponent(caster, &components.Mana{Current: 100, Max: 100, RegenRate: 1.0})
	spellbook := &components.Spellbook{
		Spells: map[string]*components.Spell{
			"heal": {
				ID:         "heal",
				ManaCost:   15,
				Cooldown:   1.0,
				Range:      20.0,
				Magnitude:  50,
				EffectType: "heal",
				LastCast:   -1,
			},
		},
	}
	w.AddComponent(caster, spellbook)

	target := w.CreateEntity()
	w.AddComponent(target, &components.Position{X: 5, Y: 0, Z: 0})
	health := &components.Health{Current: 80, Max: 100}
	w.AddComponent(target, health)

	magicSys.CastSpell(w, caster, "heal", target, nil)

	// Should cap at max
	if health.Current != 100 {
		t.Errorf("Expected health capped at 100, got %f", health.Current)
	}
}

func TestMagicSystem_SpellEffect_DamageOverTime(t *testing.T) {
	w := ecs.NewWorld()
	magicSys := NewMagicSystem()

	entity := w.CreateEntity()
	health := &components.Health{Current: 100, Max: 100}
	w.AddComponent(entity, health)

	effect := &components.SpellEffect{
		EffectType: "burn",
		Magnitude:  10, // 10 DPS
		Duration:   3.0,
		Remaining:  3.0,
	}
	w.AddComponent(entity, effect)

	// Simulate 1 second
	magicSys.Update(w, 1.0)

	if health.Current != 90 {
		t.Errorf("Expected health 90 after 1s burn, got %f", health.Current)
	}

	// Effect should have 2s remaining
	if effect.Remaining != 2.0 {
		t.Errorf("Expected 2s remaining, got %f", effect.Remaining)
	}
}

func TestMagicSystem_SpellEffect_HealOverTime(t *testing.T) {
	w := ecs.NewWorld()
	magicSys := NewMagicSystem()

	entity := w.CreateEntity()
	health := &components.Health{Current: 50, Max: 100}
	w.AddComponent(entity, health)

	effect := &components.SpellEffect{
		EffectType: "regen",
		Magnitude:  5, // 5 HPS
		Duration:   4.0,
		Remaining:  4.0,
	}
	w.AddComponent(entity, effect)

	magicSys.Update(w, 2.0)

	if health.Current != 60 {
		t.Errorf("Expected health 60 after 2s regen, got %f", health.Current)
	}
}

func TestMagicSystem_SpellEffect_Expiration(t *testing.T) {
	w := ecs.NewWorld()
	magicSys := NewMagicSystem()

	entity := w.CreateEntity()
	w.AddComponent(entity, &components.Health{Current: 100, Max: 100})

	effect := &components.SpellEffect{
		EffectType: "burn",
		Magnitude:  5,
		Duration:   2.0,
		Remaining:  2.0,
	}
	w.AddComponent(entity, effect)

	// Simulate past duration
	magicSys.Update(w, 3.0)

	// Effect should be removed
	_, hasEffect := w.GetComponent(entity, "SpellEffect")
	if hasEffect {
		t.Error("Expired spell effect should be removed")
	}
}

func TestMagicSystem_LearnSpell(t *testing.T) {
	w := ecs.NewWorld()
	magicSys := NewMagicSystem()

	entity := w.CreateEntity()

	spell := &components.Spell{
		ID:        "frostbolt",
		Name:      "Frostbolt",
		ManaCost:  15,
		Cooldown:  0.8,
		Range:     25.0,
		Magnitude: 35,
	}

	// Should create spellbook if not exists
	success := magicSys.LearnSpell(w, entity, spell)
	if !success {
		t.Error("LearnSpell should succeed")
	}

	// Verify spell was learned
	spellbookComp, ok := w.GetComponent(entity, "Spellbook")
	if !ok {
		t.Error("Spellbook should exist after learning spell")
	}
	spellbook := spellbookComp.(*components.Spellbook)

	if _, exists := spellbook.Spells["frostbolt"]; !exists {
		t.Error("Frostbolt should be in spellbook")
	}
}

func TestMagicSystem_SetActiveSpell(t *testing.T) {
	w := ecs.NewWorld()
	magicSys := NewMagicSystem()

	entity := w.CreateEntity()
	spellbook := &components.Spellbook{
		Spells: map[string]*components.Spell{
			"fireball":  {ID: "fireball"},
			"frostbolt": {ID: "frostbolt"},
		},
	}
	w.AddComponent(entity, spellbook)

	success := magicSys.SetActiveSpell(w, entity, "frostbolt")
	if !success {
		t.Error("SetActiveSpell should succeed")
	}

	if spellbook.ActiveSpellID != "frostbolt" {
		t.Errorf("Expected active spell frostbolt, got %s", spellbook.ActiveSpellID)
	}
}

func TestMagicSystem_SetActiveSpell_Unknown(t *testing.T) {
	w := ecs.NewWorld()
	magicSys := NewMagicSystem()

	entity := w.CreateEntity()
	spellbook := &components.Spellbook{
		Spells: map[string]*components.Spell{},
	}
	w.AddComponent(entity, spellbook)

	success := magicSys.SetActiveSpell(w, entity, "unknown")
	if success {
		t.Error("SetActiveSpell should fail for unknown spell")
	}
}

func TestMagicSystem_AoESpell(t *testing.T) {
	w := ecs.NewWorld()
	magicSys := NewMagicSystem()

	caster := w.CreateEntity()
	w.AddComponent(caster, &components.Position{X: 0, Y: 0, Z: 0})
	w.AddComponent(caster, &components.Mana{Current: 100, Max: 100, RegenRate: 1.0})
	spellbook := &components.Spellbook{
		Spells: map[string]*components.Spell{
			"explosion": {
				ID:           "explosion",
				ManaCost:     30,
				Cooldown:     3.0,
				Range:        20.0,
				Magnitude:    40,
				AreaOfEffect: 5.0,
				EffectType:   "damage",
				LastCast:     -1,
			},
		},
	}
	w.AddComponent(caster, spellbook)

	// Create targets at various distances from AoE center (10, 0, 0)
	target1 := w.CreateEntity()
	w.AddComponent(target1, &components.Position{X: 10, Y: 0, Z: 0})
	health1 := &components.Health{Current: 100, Max: 100}
	w.AddComponent(target1, health1)

	target2 := w.CreateEntity()
	w.AddComponent(target2, &components.Position{X: 12, Y: 0, Z: 0}) // 2 units from center
	health2 := &components.Health{Current: 100, Max: 100}
	w.AddComponent(target2, health2)

	target3 := w.CreateEntity()
	w.AddComponent(target3, &components.Position{X: 20, Y: 0, Z: 0}) // Outside AoE
	health3 := &components.Health{Current: 100, Max: 100}
	w.AddComponent(target3, health3)

	// Cast AoE at position (10, 0, 0)
	success := magicSys.CastSpellAtPosition(w, caster, "explosion", 10, 0, 0, nil)
	if !success {
		t.Error("AoE spell cast should succeed")
	}

	// Target at center should take full damage
	if health1.Current >= 100 {
		t.Error("Target at center should take damage")
	}

	// Target near center should take reduced damage
	if health2.Current >= 100 {
		t.Error("Target near center should take damage")
	}
	if health2.Current <= health1.Current {
		t.Error("Target further from center should take less damage")
	}

	// Target outside radius should be unharmed
	if health3.Current != 100 {
		t.Error("Target outside AoE should not take damage")
	}
}

func TestMagicSystem_GetMagicSkillModifier(t *testing.T) {
	w := ecs.NewWorld()
	magicSys := NewMagicSystem()

	entity := w.CreateEntity()
	skills := &components.Skills{
		Levels: map[string]int{
			"destruction": 10,
		},
	}
	w.AddComponent(entity, skills)

	modifier := magicSys.GetMagicSkillModifier(w, entity)

	// 1.0 + (10 * 0.02) = 1.2
	expected := 1.2
	if modifier != expected {
		t.Errorf("Expected modifier %f, got %f", expected, modifier)
	}
}

func TestMagicSystem_GetMagicSkillModifier_NoSkills(t *testing.T) {
	w := ecs.NewWorld()
	magicSys := NewMagicSystem()

	entity := w.CreateEntity()

	modifier := magicSys.GetMagicSkillModifier(w, entity)
	if modifier != 1.0 {
		t.Errorf("Expected modifier 1.0 without skills, got %f", modifier)
	}
}
