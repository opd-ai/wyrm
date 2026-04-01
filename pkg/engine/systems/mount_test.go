package systems

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

func TestNewMountSystem(t *testing.T) {
	sys := NewMountSystem(12345, "fantasy")

	if sys == nil {
		t.Fatal("NewMountSystem returned nil")
	}
	if sys.Seed != 12345 {
		t.Errorf("Expected seed 12345, got %d", sys.Seed)
	}
	if sys.Genre != "fantasy" {
		t.Errorf("Expected genre 'fantasy', got %s", sys.Genre)
	}
	if sys.MountCount() != 0 {
		t.Errorf("Expected 0 mounts, got %d", sys.MountCount())
	}
	if len(sys.Archetypes) == 0 {
		t.Error("Archetypes should be initialized")
	}
}

func TestMountArchetypes(t *testing.T) {
	sys := NewMountSystem(12345, "fantasy")

	// Check fantasy archetypes
	horse := sys.GetArchetype(MountHorse)
	if horse == nil {
		t.Fatal("Horse archetype should exist")
	}
	if horse.Name != "War Horse" {
		t.Errorf("Expected 'War Horse', got %s", horse.Name)
	}
	if horse.BaseStats.Speed <= 0 {
		t.Error("Horse should have positive speed")
	}

	// Check griffin
	griffin := sys.GetArchetype(MountGriffin)
	if griffin == nil {
		t.Fatal("Griffin archetype should exist")
	}
	hasFlying := false
	for _, trait := range griffin.Traits {
		if trait == TraitFlying {
			hasFlying = true
			break
		}
	}
	if !hasFlying {
		t.Error("Griffin should have flying trait")
	}

	// Check non-existent
	fake := sys.GetArchetype("fake")
	if fake != nil {
		t.Error("Fake archetype should not exist")
	}
}

func TestGetGenreArchetypes(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		sys := NewMountSystem(12345, genre)
		archetypes := sys.GetGenreArchetypes()

		if len(archetypes) == 0 {
			t.Errorf("No archetypes found for genre: %s", genre)
			continue
		}

		for _, arch := range archetypes {
			if arch.Genre != genre {
				t.Errorf("Archetype %s has wrong genre: %s", arch.Name, arch.Genre)
			}
		}

		t.Logf("%s: %d archetypes", genre, len(archetypes))
	}
}

func TestTameMount(t *testing.T) {
	sys := NewMountSystem(12345, "fantasy")
	w := ecs.NewWorld()
	entity := w.CreateEntity()

	err := sys.TameMount(entity, 100, MountHorse, "Shadowmere")
	if err != nil {
		t.Fatalf("TameMount failed: %v", err)
	}

	mount := sys.GetMount(entity)
	if mount == nil {
		t.Fatal("Mount should be registered")
	}
	if mount.Name != "Shadowmere" {
		t.Errorf("Expected name 'Shadowmere', got %s", mount.Name)
	}
	if mount.OwnerID != 100 {
		t.Errorf("Expected owner 100, got %d", mount.OwnerID)
	}
	if mount.Level != 1 {
		t.Errorf("Expected level 1, got %d", mount.Level)
	}
	if sys.MountCount() != 1 {
		t.Errorf("Expected 1 mount, got %d", sys.MountCount())
	}

	// Duplicate tame should fail
	err = sys.TameMount(entity, 200, MountWolf, "Fang")
	if err == nil {
		t.Error("Duplicate tame should fail")
	}

	// Unknown mount type
	entity2 := w.CreateEntity()
	err = sys.TameMount(entity2, 100, "unknown", "Test")
	if err == nil {
		t.Error("Unknown mount type should fail")
	}
}

func TestMountDismount(t *testing.T) {
	sys := NewMountSystem(12345, "fantasy")
	w := ecs.NewWorld()
	mountEntity := w.CreateEntity()
	riderEntity := w.CreateEntity()

	sys.TameMount(mountEntity, 100, MountHorse, "Test")

	// Mount
	err := sys.MountCreature(mountEntity, riderEntity)
	if err != nil {
		t.Fatalf("MountCreature failed: %v", err)
	}

	if !sys.IsMounted(mountEntity) {
		t.Error("Mount should be mounted")
	}

	// Double mount should fail
	rider2 := w.CreateEntity()
	err = sys.MountCreature(mountEntity, rider2)
	if err == nil {
		t.Error("Should not allow second rider")
	}

	// Dismount
	rider := sys.DismountCreature(mountEntity)
	if rider != riderEntity {
		t.Error("Dismount should return correct rider")
	}

	if sys.IsMounted(mountEntity) {
		t.Error("Mount should not be mounted after dismount")
	}

	// Dismount empty
	rider = sys.DismountCreature(mountEntity)
	if rider != 0 {
		t.Error("Dismount empty should return 0")
	}
}

func TestMountMoodAndHunger(t *testing.T) {
	sys := NewMountSystem(12345, "fantasy")
	w := ecs.NewWorld()
	entity := w.CreateEntity()

	sys.TameMount(entity, 100, MountHorse, "Test")
	mount := sys.GetMount(entity)

	initialMood := mount.Mood

	// Simulate time passing (hunger increases)
	sys.Update(nil, 100) // 100 seconds
	if mount.Hunger <= 0 {
		t.Error("Hunger should increase over time")
	}

	// Feed the mount
	err := sys.FeedMount(entity, 50)
	if err != nil {
		t.Fatalf("FeedMount failed: %v", err)
	}
	if mount.Hunger >= 50 {
		t.Error("Feeding should reduce hunger")
	}
	if mount.Mood <= initialMood {
		t.Error("Feeding should improve mood")
	}
}

func TestMountSprinting(t *testing.T) {
	sys := NewMountSystem(12345, "fantasy")
	w := ecs.NewWorld()
	entity := w.CreateEntity()
	rider := w.CreateEntity()

	sys.TameMount(entity, 100, MountHorse, "Test")
	sys.MountCreature(entity, rider)

	normalSpeed := sys.GetSpeed(entity)

	// Enable sprinting
	err := sys.SetSprinting(entity, true)
	if err != nil {
		t.Fatalf("SetSprinting failed: %v", err)
	}

	sprintSpeed := sys.GetSpeed(entity)
	if sprintSpeed <= normalSpeed {
		t.Error("Sprint speed should be faster than normal")
	}

	// Drain stamina
	mount := sys.GetMount(entity)
	mount.Stats.Stamina = 5 // Low stamina

	err = sys.SetSprinting(entity, true)
	if err == nil {
		t.Error("Should not sprint with low stamina")
	}
}

func TestMountRest(t *testing.T) {
	sys := NewMountSystem(12345, "fantasy")
	w := ecs.NewWorld()
	entity := w.CreateEntity()

	sys.TameMount(entity, 100, MountHorse, "Test")
	mount := sys.GetMount(entity)

	// Drain stamina
	mount.Stats.Stamina = 20

	// Rest
	err := sys.RestMount(entity, 10)
	if err != nil {
		t.Fatalf("RestMount failed: %v", err)
	}

	if mount.Stats.Stamina <= 20 {
		t.Error("Rest should recover stamina")
	}

	// Cannot rest while mounted
	rider := w.CreateEntity()
	sys.MountCreature(entity, rider)
	err = sys.RestMount(entity, 10)
	if err == nil {
		t.Error("Should not rest while mounted")
	}
}

func TestMountTraits(t *testing.T) {
	sys := NewMountSystem(12345, "fantasy")
	w := ecs.NewWorld()

	// Griffin has flying trait
	griffinEntity := w.CreateEntity()
	sys.TameMount(griffinEntity, 100, MountGriffin, "Griffin")

	if !sys.HasTrait(griffinEntity, TraitFlying) {
		t.Error("Griffin should have flying trait")
	}
	if !sys.CanFly(griffinEntity) {
		t.Error("Griffin should be able to fly")
	}

	// Horse does not have flying
	horseEntity := w.CreateEntity()
	sys.TameMount(horseEntity, 100, MountHorse, "Horse")

	if sys.HasTrait(horseEntity, TraitFlying) {
		t.Error("Horse should not have flying trait")
	}
	if sys.CanFly(horseEntity) {
		t.Error("Horse should not be able to fly")
	}
}

func TestMountExperience(t *testing.T) {
	sys := NewMountSystem(12345, "fantasy")
	w := ecs.NewWorld()
	entity := w.CreateEntity()

	sys.TameMount(entity, 100, MountHorse, "Test")

	// Add some XP (not enough to level)
	leveledUp, level := sys.AddExperience(entity, 50)
	if leveledUp {
		t.Error("Should not level up with 50 XP")
	}
	if level != 1 {
		t.Errorf("Expected level 1, got %d", level)
	}

	// Add enough to level up
	leveledUp, level = sys.AddExperience(entity, 60)
	if !leveledUp {
		t.Error("Should level up")
	}
	if level != 2 {
		t.Errorf("Expected level 2, got %d", level)
	}

	// Verify stats improved
	mount := sys.GetMount(entity)
	arch := sys.GetArchetype(MountHorse)
	if mount.Stats.MaxHealth <= arch.BaseStats.MaxHealth {
		t.Error("Max health should increase on level up")
	}
}

func TestMountBonding(t *testing.T) {
	sys := NewMountSystem(12345, "fantasy")
	w := ecs.NewWorld()
	entity := w.CreateEntity()

	sys.TameMount(entity, 100, MountHorse, "Test")
	mount := sys.GetMount(entity)

	// Initial bond level
	if mount.BondLevel != 1 {
		t.Errorf("Expected bond level 1, got %d", mount.BondLevel)
	}

	// Need high loyalty to improve bond
	mount.Loyalty = 25 // Enough for level 2

	improved := sys.ImproveBond(entity)
	if !improved {
		t.Error("Bond should improve with sufficient loyalty")
	}
	if mount.BondLevel != 2 {
		t.Errorf("Expected bond level 2, got %d", mount.BondLevel)
	}
}

func TestMountDamageAndHealing(t *testing.T) {
	sys := NewMountSystem(12345, "fantasy")
	w := ecs.NewWorld()
	entity := w.CreateEntity()

	sys.TameMount(entity, 100, MountHorse, "Test")
	mount := sys.GetMount(entity)

	initialHealth := mount.Stats.Health

	// Take damage
	damage := sys.DamageMount(entity, 20)
	if damage <= 0 {
		t.Error("Damage should be positive")
	}
	if mount.Stats.Health >= initialHealth {
		t.Error("Health should decrease")
	}

	// Heal
	healing := sys.HealMount(entity, 10)
	if healing <= 0 {
		t.Error("Healing should be positive")
	}

	// Check alive status
	if !sys.IsAlive(entity) {
		t.Error("Mount should be alive")
	}

	// Kill mount
	sys.DamageMount(entity, 1000)
	if sys.IsAlive(entity) {
		t.Error("Mount should be dead")
	}
}

func TestMountArmoredTrait(t *testing.T) {
	sys := NewMountSystem(12345, "fantasy")
	w := ecs.NewWorld()

	// Dragon has armored trait
	dragonEntity := w.CreateEntity()
	sys.TameMount(dragonEntity, 100, MountDragon, "Dragon")

	// Horse does not
	horseEntity := w.CreateEntity()
	sys.TameMount(horseEntity, 100, MountHorse, "Horse")

	dragon := sys.GetMount(dragonEntity)
	horse := sys.GetMount(horseEntity)

	// Set same health
	dragon.Stats.Health = 100
	horse.Stats.Health = 100

	// Apply same damage
	dragonDamage := sys.DamageMount(dragonEntity, 50)
	horseDamage := sys.DamageMount(horseEntity, 50)

	if dragonDamage >= horseDamage {
		t.Error("Dragon with armored trait should take less damage")
	}
}

func TestMountCargoCapacity(t *testing.T) {
	sys := NewMountSystem(12345, "fantasy")
	w := ecs.NewWorld()
	entity := w.CreateEntity()

	sys.TameMount(entity, 100, MountHorse, "Test")

	capacity := sys.GetCargoCapacity(entity)
	if capacity <= 0 {
		t.Error("Horse should have cargo capacity")
	}

	// Test non-existent
	fakeEntity := w.CreateEntity()
	capacity = sys.GetCargoCapacity(fakeEntity)
	if capacity != 0 {
		t.Error("Non-existent mount should have 0 capacity")
	}
}

func TestMountEquipment(t *testing.T) {
	sys := NewMountSystem(12345, "fantasy")
	w := ecs.NewWorld()
	entity := w.CreateEntity()

	sys.TameMount(entity, 100, MountHorse, "Test")

	// Set equipment
	err := sys.SetMountEquipment(entity, "saddle", "iron_saddle")
	if err != nil {
		t.Fatalf("SetMountEquipment failed: %v", err)
	}

	// Get equipment
	eq := sys.GetMountEquipment(entity, "saddle")
	if eq != "iron_saddle" {
		t.Errorf("Expected 'iron_saddle', got '%s'", eq)
	}

	// Non-existent slot
	eq = sys.GetMountEquipment(entity, "nonexistent")
	if eq != "" {
		t.Error("Non-existent slot should return empty string")
	}
}

func TestReleaseMount(t *testing.T) {
	sys := NewMountSystem(12345, "fantasy")
	w := ecs.NewWorld()
	entity := w.CreateEntity()
	rider := w.CreateEntity()

	sys.TameMount(entity, 100, MountHorse, "Test")

	// Cannot release while mounted
	sys.MountCreature(entity, rider)
	err := sys.ReleaseMount(entity)
	if err == nil {
		t.Error("Should not release while mounted")
	}

	// Dismount first
	sys.DismountCreature(entity)

	// Now release
	err = sys.ReleaseMount(entity)
	if err != nil {
		t.Fatalf("ReleaseMount failed: %v", err)
	}

	if sys.MountCount() != 0 {
		t.Error("Mount should be removed")
	}

	// Release non-existent
	err = sys.ReleaseMount(entity)
	if err == nil {
		t.Error("Should error for non-existent mount")
	}
}

func TestMountStaminaDrain(t *testing.T) {
	sys := NewMountSystem(12345, "fantasy")
	w := ecs.NewWorld()
	entity := w.CreateEntity()
	rider := w.CreateEntity()

	sys.TameMount(entity, 100, MountHorse, "Test")
	mount := sys.GetMount(entity)

	initialStamina := mount.Stats.Stamina

	// Mount and ride
	sys.MountCreature(entity, rider)
	sys.Update(nil, 10) // 10 seconds of riding

	if mount.Stats.Stamina >= initialStamina {
		t.Error("Stamina should drain while riding")
	}

	// Sprint drains faster
	mount.Stats.Stamina = 100
	sys.SetSprinting(entity, true)
	sys.Update(nil, 10)

	// Should drain more while sprinting
	if mount.Stats.Stamina > 50 { // Sprinting drains 5/sec
		t.Error("Sprinting should drain stamina faster")
	}
}

func TestMountSpeedModifiers(t *testing.T) {
	sys := NewMountSystem(12345, "fantasy")
	w := ecs.NewWorld()
	entity := w.CreateEntity()

	sys.TameMount(entity, 100, MountHorse, "Test")
	mount := sys.GetMount(entity)

	// Full mood and stamina
	mount.Mood = 100
	mount.Stats.Stamina = mount.Stats.MaxStamina
	normalSpeed := sys.GetSpeed(entity)

	// Low mood
	mount.Mood = 10
	lowMoodSpeed := sys.GetSpeed(entity)
	if lowMoodSpeed >= normalSpeed {
		t.Error("Low mood should reduce speed")
	}

	// High hunger
	mount.Mood = 100
	mount.Hunger = 90
	hungrySpeed := sys.GetSpeed(entity)
	if hungrySpeed >= normalSpeed {
		t.Error("Hunger should reduce speed")
	}

	// Low stamina
	mount.Hunger = 0
	mount.Stats.Stamina = mount.Stats.MaxStamina * 0.1
	exhaustedSpeed := sys.GetSpeed(entity)
	if exhaustedSpeed >= normalSpeed {
		t.Error("Low stamina should reduce speed")
	}
}

func TestMountStats(t *testing.T) {
	sys := NewMountSystem(12345, "fantasy")
	w := ecs.NewWorld()
	entity := w.CreateEntity()

	sys.TameMount(entity, 100, MountHorse, "Test")

	stats := sys.GetMountStats(entity)
	if stats == nil {
		t.Fatal("GetMountStats returned nil")
	}
	if stats.Speed <= 0 {
		t.Error("Speed should be positive")
	}
	if stats.MaxHealth <= 0 {
		t.Error("MaxHealth should be positive")
	}

	// Non-existent
	fakeEntity := w.CreateEntity()
	stats = sys.GetMountStats(fakeEntity)
	if stats != nil {
		t.Error("Stats should be nil for non-existent mount")
	}
}

func TestMountLevelAndBondHelpers(t *testing.T) {
	sys := NewMountSystem(12345, "fantasy")
	w := ecs.NewWorld()
	entity := w.CreateEntity()

	sys.TameMount(entity, 100, MountHorse, "Test")

	level := sys.GetMountLevel(entity)
	if level != 1 {
		t.Errorf("Expected level 1, got %d", level)
	}

	bond := sys.GetBondLevel(entity)
	if bond != 1 {
		t.Errorf("Expected bond 1, got %d", bond)
	}

	// Non-existent
	fakeEntity := w.CreateEntity()
	if sys.GetMountLevel(fakeEntity) != 0 {
		t.Error("Non-existent mount level should be 0")
	}
	if sys.GetBondLevel(fakeEntity) != 0 {
		t.Error("Non-existent mount bond should be 0")
	}
}

func TestMountExhaustedCannotRide(t *testing.T) {
	sys := NewMountSystem(12345, "fantasy")
	w := ecs.NewWorld()
	entity := w.CreateEntity()
	rider := w.CreateEntity()

	sys.TameMount(entity, 100, MountHorse, "Test")
	mount := sys.GetMount(entity)

	// Exhaust stamina
	mount.Stats.Stamina = 0

	err := sys.MountCreature(entity, rider)
	if err == nil {
		t.Error("Should not be able to mount exhausted creature")
	}
}

func TestMountLowMoodRefuses(t *testing.T) {
	sys := NewMountSystem(12345, "fantasy")
	w := ecs.NewWorld()
	entity := w.CreateEntity()
	rider := w.CreateEntity()

	sys.TameMount(entity, 100, MountHorse, "Test")
	mount := sys.GetMount(entity)

	// Very low mood
	mount.Mood = 5

	err := sys.MountCreature(entity, rider)
	if err == nil {
		t.Error("Mount with very low mood should refuse")
	}
}

func TestMountMaxLevel(t *testing.T) {
	sys := NewMountSystem(12345, "fantasy")
	w := ecs.NewWorld()
	entity := w.CreateEntity()

	sys.TameMount(entity, 100, MountHorse, "Test")

	// Add massive XP
	for i := 0; i < 100; i++ {
		sys.AddExperience(entity, 10000)
	}

	level := sys.GetMountLevel(entity)
	if level > 20 {
		t.Errorf("Level should cap at 20, got %d", level)
	}
}

func TestConcurrentMountAccess(t *testing.T) {
	sys := NewMountSystem(12345, "fantasy")
	w := ecs.NewWorld()

	// Create multiple mounts
	var entities []ecs.Entity
	for i := 0; i < 5; i++ {
		e := w.CreateEntity()
		sys.TameMount(e, uint64(100+i), MountHorse, "Test")
		entities = append(entities, e)
	}

	done := make(chan bool, 100)

	// Concurrent reads
	for i := 0; i < 50; i++ {
		go func(idx int) {
			e := entities[idx%len(entities)]
			_ = sys.GetMount(e)
			_ = sys.GetSpeed(e)
			_ = sys.GetMountStats(e)
			done <- true
		}(i)
	}

	// Concurrent updates
	for i := 0; i < 50; i++ {
		go func(idx int) {
			e := entities[idx%len(entities)]
			sys.AddExperience(e, 10)
			sys.FeedMount(e, 5)
			done <- true
		}(i)
	}

	// Wait for all
	for i := 0; i < 100; i++ {
		<-done
	}
}
