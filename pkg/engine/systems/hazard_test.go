package systems

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

func TestNewHazardSystem(t *testing.T) {
	sys := NewHazardSystem("fantasy")
	if sys == nil {
		t.Fatal("NewHazardSystem returned nil")
	}
	if sys.Genre != "fantasy" {
		t.Errorf("Genre = %s, expected fantasy", sys.Genre)
	}
	if sys.DamageMultiplier != 1.0 {
		t.Errorf("DamageMultiplier = %f, expected 1.0", sys.DamageMultiplier)
	}
}

func TestHazardSystemUpdate(t *testing.T) {
	sys := NewHazardSystem("fantasy")
	w := ecs.NewWorld()

	// Should not panic with empty world
	sys.Update(w, 0.016)
}

func TestHazardSystemCreateHazard(t *testing.T) {
	sys := NewHazardSystem("sci-fi")
	w := ecs.NewWorld()

	ent := sys.CreateHazard(w, components.HazardTypeFire, 10, 20, 0, 5.0, 0.8)

	if ent == 0 {
		t.Fatal("CreateHazard returned invalid entity")
	}

	// Check position
	posComp, ok := w.GetComponent(ent, "Position")
	if !ok {
		t.Fatal("Hazard should have Position component")
	}
	pos := posComp.(*components.Position)
	if pos.X != 10 || pos.Y != 20 {
		t.Errorf("Position = (%f,%f), expected (10,20)", pos.X, pos.Y)
	}

	// Check hazard component
	hazardComp, ok := w.GetComponent(ent, "EnvironmentalHazard")
	if !ok {
		t.Fatal("Hazard should have EnvironmentalHazard component")
	}
	hazard := hazardComp.(*components.EnvironmentalHazard)
	if hazard.HazardType != components.HazardTypeFire {
		t.Errorf("HazardType = %s, expected fire", hazard.HazardType)
	}
	if hazard.Radius != 5.0 {
		t.Errorf("Radius = %f, expected 5.0", hazard.Radius)
	}
	if hazard.Intensity != 0.8 {
		t.Errorf("Intensity = %f, expected 0.8", hazard.Intensity)
	}
	if !hazard.Active {
		t.Error("Hazard should be active")
	}
}

func TestHazardSystemCreatePermanentHazard(t *testing.T) {
	sys := NewHazardSystem("horror")
	w := ecs.NewWorld()

	ent := sys.CreatePermanentHazard(w, components.HazardTypeLava, 0, 0, 0, 3.0, 1.0)

	hazardComp, _ := w.GetComponent(ent, "EnvironmentalHazard")
	hazard := hazardComp.(*components.EnvironmentalHazard)

	if !hazard.Permanent {
		t.Error("Hazard should be permanent")
	}
}

func TestHazardSystemCreateTrap(t *testing.T) {
	sys := NewHazardSystem("cyberpunk")
	w := ecs.NewWorld()

	ent := sys.CreateTrap(w, "spike", 5, 10, 0, 25.0, 1.0)

	if ent == 0 {
		t.Fatal("CreateTrap returned invalid entity")
	}

	// Check trap component
	trapComp, ok := w.GetComponent(ent, "TrapMechanism")
	if !ok {
		t.Fatal("Trap should have TrapMechanism component")
	}
	trap := trapComp.(*components.TrapMechanism)
	if trap.TrapType != "spike" {
		t.Errorf("TrapType = %s, expected spike", trap.TrapType)
	}
	if trap.Damage != 25.0 {
		t.Errorf("Damage = %f, expected 25.0", trap.Damage)
	}
	if !trap.Armed {
		t.Error("Trap should be armed")
	}
	if trap.Triggered {
		t.Error("Trap should not be triggered")
	}
}

func TestHazardSystemGetBaseDamageForHazard(t *testing.T) {
	sys := NewHazardSystem("fantasy")

	tests := []struct {
		hazardType components.HazardType
		minDamage  float64
	}{
		{components.HazardTypeFire, 5.0},
		{components.HazardTypeLava, 40.0},
		{components.HazardTypeRadiation, 1.0},
		{components.HazardTypePoison, 5.0},
		{components.HazardTypeAcid, 10.0},
		{components.HazardTypeElectric, 15.0},
		{components.HazardTypeFreeze, 5.0},
		{components.HazardTypeMagic, 10.0},
		{components.HazardTypeGas, 1.0},
	}

	for _, tt := range tests {
		damage := sys.getBaseDamageForHazard(tt.hazardType)
		if damage < tt.minDamage {
			t.Errorf("Base damage for %s = %f, expected at least %f", tt.hazardType, damage, tt.minDamage)
		}
	}
}

func TestHazardSystemGetGenreHazardName(t *testing.T) {
	tests := []struct {
		genre      string
		hazardType components.HazardType
		expected   string
	}{
		{"fantasy", components.HazardTypeRadiation, "Cursed Ground"},
		{"sci-fi", components.HazardTypeRadiation, "Radiation Zone"},
		{"horror", components.HazardTypeRadiation, "Necrotic Field"},
		{"cyberpunk", components.HazardTypeRadiation, "Toxic Waste"},
		{"post-apocalyptic", components.HazardTypeRadiation, "Hot Zone"},
		{"fantasy", components.HazardTypeFire, "Dragon Fire"},
		{"cyberpunk", components.HazardTypeFire, "Napalm"},
	}

	for _, tt := range tests {
		sys := NewHazardSystem(tt.genre)
		name := sys.GetGenreHazardName(tt.hazardType)
		if name != tt.expected {
			t.Errorf("GetGenreHazardName(%s, %s) = %s, expected %s", tt.genre, tt.hazardType, name, tt.expected)
		}
	}
}

func TestHazardSystemApplyHazardDamage(t *testing.T) {
	sys := NewHazardSystem("fantasy")
	w := ecs.NewWorld()

	// Create target with health
	target := w.CreateEntity()
	w.AddComponent(target, &components.Health{Current: 100, Max: 100})

	hazard := &components.EnvironmentalHazard{
		HazardType:      components.HazardTypeFire,
		Intensity:       1.0,
		DamagePerSecond: 10.0,
	}

	sys.applyHazardDamage(w, target, hazard, 1.0)

	healthComp, _ := w.GetComponent(target, "Health")
	health := healthComp.(*components.Health)

	if health.Current >= 100 {
		t.Error("Health should be reduced after hazard damage")
	}
}

func TestHazardSystemApplyHazardDamageWithResistance(t *testing.T) {
	sys := NewHazardSystem("fantasy")
	w := ecs.NewWorld()

	// Create target with health and fire resistance
	target := w.CreateEntity()
	w.AddComponent(target, &components.Health{Current: 100, Max: 100})
	w.AddComponent(target, &components.HazardResistance{
		Resistances: map[components.HazardType]float64{
			components.HazardTypeFire: 0.5, // 50% resistance
		},
	})

	hazard := &components.EnvironmentalHazard{
		HazardType:      components.HazardTypeFire,
		Intensity:       1.0,
		DamagePerSecond: 10.0,
	}

	sys.applyHazardDamage(w, target, hazard, 1.0)

	healthComp, _ := w.GetComponent(target, "Health")
	health := healthComp.(*components.Health)

	// With 50% resistance, should take ~5 damage instead of 10
	expectedHealth := 95.0
	if health.Current < expectedHealth-0.1 || health.Current > expectedHealth+0.1 {
		t.Errorf("Health = %f, expected ~%f with 50%% resistance", health.Current, expectedHealth)
	}
}

func TestHazardSystemTrapTriggering(t *testing.T) {
	sys := NewHazardSystem("fantasy")
	w := ecs.NewWorld()

	// Create trap
	trapEnt := sys.CreateTrap(w, "spike", 5, 5, 0, 20.0, 2.0)

	// Create target in trigger radius
	target := w.CreateEntity()
	w.AddComponent(target, &components.Position{X: 5, Y: 5, Z: 0})
	w.AddComponent(target, &components.Health{Current: 100, Max: 100})

	// Process traps
	sys.processTraps(w, 0.016)

	// Check trap was triggered
	trapComp, _ := w.GetComponent(trapEnt, "TrapMechanism")
	trap := trapComp.(*components.TrapMechanism)

	if !trap.Triggered {
		t.Error("Trap should be triggered")
	}
	if trap.Armed {
		t.Error("Trap should not be armed after triggering")
	}

	// Check target took damage
	healthComp, _ := w.GetComponent(target, "Health")
	health := healthComp.(*components.Health)

	if health.Current >= 100 {
		t.Error("Target should have taken trap damage")
	}
}

func TestHazardSystemHazardEffectProcessing(t *testing.T) {
	sys := NewHazardSystem("horror")
	w := ecs.NewWorld()

	// Create entity with hazard effect
	ent := w.CreateEntity()
	w.AddComponent(ent, &components.Health{Current: 100, Max: 100})
	w.AddComponent(ent, &components.HazardEffect{
		SourceHazard:      components.HazardTypePoison,
		StackCount:        1,
		MaxStacks:         3,
		RemainingDuration: 10.0,
		DamageOverTime:    5.0,
	})

	// Process effects
	sys.processHazardEffects(w, 1.0)

	// Check damage was applied
	healthComp, _ := w.GetComponent(ent, "Health")
	health := healthComp.(*components.Health)

	if health.Current >= 100 {
		t.Error("Health should be reduced by damage over time")
	}

	// Check duration was reduced
	effectComp, _ := w.GetComponent(ent, "HazardEffect")
	effect := effectComp.(*components.HazardEffect)

	if effect.RemainingDuration >= 10.0 {
		t.Error("Effect duration should be reduced")
	}
}

func TestHazardSystemEffectExpiration(t *testing.T) {
	sys := NewHazardSystem("sci-fi")
	w := ecs.NewWorld()

	// Create entity with nearly-expired effect
	ent := w.CreateEntity()
	w.AddComponent(ent, &components.Health{Current: 100, Max: 100})
	w.AddComponent(ent, &components.HazardEffect{
		SourceHazard:      components.HazardTypeRadiation,
		StackCount:        1,
		RemainingDuration: 0.5,
		DamageOverTime:    1.0,
	})

	// Process with enough dt to expire the effect
	sys.processHazardEffects(w, 1.0)

	// Effect should be removed
	_, hasEffect := w.GetComponent(ent, "HazardEffect")
	if hasEffect {
		t.Error("Expired effect should be removed")
	}
}

func TestHazardSystemTemporaryHazardExpiration(t *testing.T) {
	sys := NewHazardSystem("post-apocalyptic")
	w := ecs.NewWorld()

	// Create temporary hazard
	ent := sys.CreateHazard(w, components.HazardTypeGas, 0, 0, 0, 5.0, 0.5)

	hazardComp, _ := w.GetComponent(ent, "EnvironmentalHazard")
	hazard := hazardComp.(*components.EnvironmentalHazard)
	hazard.Duration = 1.0 // 1 second remaining

	// Process with enough time to expire
	sys.updateTemporaryHazards(w, 2.0)

	// Hazard should be inactive
	if hazard.Active {
		t.Error("Expired hazard should be inactive")
	}
}

func BenchmarkHazardSystemUpdate(b *testing.B) {
	sys := NewHazardSystem("fantasy")
	w := ecs.NewWorld()

	// Create many hazards and targets
	for i := 0; i < 20; i++ {
		sys.CreateHazard(w, components.HazardTypeFire, float64(i*10), float64(i*10), 0, 5.0, 0.5)
	}
	for i := 0; i < 100; i++ {
		ent := w.CreateEntity()
		w.AddComponent(ent, &components.Position{X: float64(i * 5), Y: float64(i * 5), Z: 0})
		w.AddComponent(ent, &components.Health{Current: 100, Max: 100})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(w, 0.016)
	}
}
