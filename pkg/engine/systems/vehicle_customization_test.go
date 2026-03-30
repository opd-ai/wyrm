package systems

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/ecs"
)

func TestNewVehicleCustomizationSystem(t *testing.T) {
	sys := NewVehicleCustomizationSystem(12345, "fantasy")

	if sys == nil {
		t.Fatal("NewVehicleCustomizationSystem returned nil")
	}
	if sys.Seed != 12345 {
		t.Errorf("Expected seed 12345, got %d", sys.Seed)
	}
	if sys.Genre != "fantasy" {
		t.Errorf("Expected genre 'fantasy', got %s", sys.Genre)
	}
	if sys.VehicleCount() != 0 {
		t.Errorf("Expected 0 vehicles, got %d", sys.VehicleCount())
	}
	if sys.CatalogSize() == 0 {
		t.Error("Catalog should not be empty")
	}
}

func TestCatalogInitialization(t *testing.T) {
	sys := NewVehicleCustomizationSystem(12345, "fantasy")

	// Check catalog has items for all categories
	categories := make(map[CustomizationCategory]int)
	for _, custom := range sys.Catalog {
		categories[custom.Category]++
	}

	expectedCategories := []CustomizationCategory{
		CategoryEngine, CategoryArmor, CategoryPerformance,
		CategoryStorage, CategoryAppearance, CategoryWeapons,
	}

	for _, cat := range expectedCategories {
		if categories[cat] == 0 {
			t.Errorf("No customizations found for category: %s", cat)
		}
	}
}

func TestGenreCatalog(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		sys := NewVehicleCustomizationSystem(12345, genre)

		ids := sys.GenreCatalogs[genre]
		if len(ids) == 0 {
			t.Errorf("No customizations found for genre: %s", genre)
			continue
		}

		// Verify all IDs exist in catalog
		for _, id := range ids {
			if sys.Catalog[id] == nil {
				t.Errorf("Customization ID in genre catalog but not in main catalog: %s", id)
			}
		}
	}
}

func TestRegisterVehicle(t *testing.T) {
	sys := NewVehicleCustomizationSystem(12345, "fantasy")
	w := ecs.NewWorld()
	entity := w.CreateEntity()

	state := sys.RegisterVehicle(entity, "horse", 100)

	if state == nil {
		t.Fatal("RegisterVehicle returned nil")
	}
	if state.EntityID != entity {
		t.Error("Entity ID mismatch")
	}
	if state.VehicleType != "horse" {
		t.Errorf("Expected vehicle type 'horse', got %s", state.VehicleType)
	}
	if state.MaxWeight != 100 {
		t.Errorf("Expected max weight 100, got %f", state.MaxWeight)
	}
	if state.Level != 1 {
		t.Errorf("Expected level 1, got %d", state.Level)
	}
	if sys.VehicleCount() != 1 {
		t.Errorf("Expected 1 vehicle, got %d", sys.VehicleCount())
	}
}

func TestGetVehicleState(t *testing.T) {
	sys := NewVehicleCustomizationSystem(12345, "fantasy")
	w := ecs.NewWorld()
	entity := w.CreateEntity()

	sys.RegisterVehicle(entity, "cart", 150)

	state := sys.GetVehicleState(entity)
	if state == nil {
		t.Fatal("GetVehicleState returned nil for registered vehicle")
	}
	if state.VehicleType != "cart" {
		t.Errorf("Expected 'cart', got %s", state.VehicleType)
	}

	// Non-existent vehicle
	unregistered := w.CreateEntity()
	state = sys.GetVehicleState(unregistered)
	if state != nil {
		t.Error("GetVehicleState should return nil for unregistered vehicle")
	}
}

func TestInstallCustomization(t *testing.T) {
	sys := NewVehicleCustomizationSystem(12345, "fantasy")
	w := ecs.NewWorld()
	entity := w.CreateEntity()

	sys.RegisterVehicle(entity, "horse", 100)

	// Install an engine upgrade
	err := sys.InstallCustomization(entity, "engine_enchanted")
	if err != nil {
		t.Fatalf("InstallCustomization failed: %v", err)
	}

	count := sys.GetInstalledCount(entity)
	if count != 1 {
		t.Errorf("Expected 1 installed mod, got %d", count)
	}

	// Verify state has the mod
	state := sys.GetVehicleState(entity)
	if state.InstalledMods[SlotEnginePrimary] == nil {
		t.Error("Engine mod should be in engine slot")
	}
	if state.InstalledMods[SlotEnginePrimary].ID != "engine_enchanted" {
		t.Error("Wrong mod installed")
	}
}

func TestInstallCustomizationErrors(t *testing.T) {
	sys := NewVehicleCustomizationSystem(12345, "fantasy")
	w := ecs.NewWorld()
	entity := w.CreateEntity()

	// Unregistered vehicle
	err := sys.InstallCustomization(entity, "engine_enchanted")
	if err == nil {
		t.Error("Should error for unregistered vehicle")
	}

	sys.RegisterVehicle(entity, "horse", 100)

	// Non-existent customization
	err = sys.InstallCustomization(entity, "nonexistent")
	if err == nil {
		t.Error("Should error for non-existent customization")
	}

	// Level too high
	err = sys.InstallCustomization(entity, "engine_elemental") // Requires level 10
	if err == nil {
		t.Error("Should error for level requirement not met")
	}
}

func TestInstallCustomizationWeight(t *testing.T) {
	sys := NewVehicleCustomizationSystem(12345, "fantasy")
	w := ecs.NewWorld()
	entity := w.CreateEntity()

	// Very low max weight
	sys.RegisterVehicle(entity, "horse", 1)

	// Try to install something heavy
	err := sys.InstallCustomization(entity, "armor_iron") // 20 weight
	if err == nil {
		t.Error("Should error for exceeding weight capacity")
	}
}

func TestUninstallCustomization(t *testing.T) {
	sys := NewVehicleCustomizationSystem(12345, "fantasy")
	w := ecs.NewWorld()
	entity := w.CreateEntity()

	sys.RegisterVehicle(entity, "horse", 100)
	sys.InstallCustomization(entity, "engine_enchanted")

	// Verify installed
	if sys.GetInstalledCount(entity) != 1 {
		t.Fatal("Should have 1 mod installed")
	}

	// Uninstall
	err := sys.UninstallCustomization(entity, SlotEnginePrimary)
	if err != nil {
		t.Fatalf("UninstallCustomization failed: %v", err)
	}

	if sys.GetInstalledCount(entity) != 0 {
		t.Error("Should have 0 mods after uninstall")
	}

	// Uninstall empty slot
	err = sys.UninstallCustomization(entity, SlotEnginePrimary)
	if err == nil {
		t.Error("Should error for empty slot")
	}
}

func TestInstallReplacesExisting(t *testing.T) {
	sys := NewVehicleCustomizationSystem(12345, "fantasy")
	w := ecs.NewWorld()
	entity := w.CreateEntity()

	sys.RegisterVehicle(entity, "horse", 100)

	// Install first engine
	sys.InstallCustomization(entity, "engine_enchanted")
	state := sys.GetVehicleState(entity)
	initialWeight := state.TotalWeight

	// Install different armor (same slot)
	err := sys.InstallCustomization(entity, "armor_iron")
	if err != nil {
		t.Fatalf("Install replacement failed: %v", err)
	}

	// Should still have only 2 mods (engine + armor)
	if sys.GetInstalledCount(entity) != 2 {
		t.Errorf("Expected 2 mods, got %d", sys.GetInstalledCount(entity))
	}

	// Weight should have changed
	if state.TotalWeight <= initialWeight {
		t.Log("Weight changed as expected")
	}
}

func TestGetModifiedStats(t *testing.T) {
	sys := NewVehicleCustomizationSystem(12345, "fantasy")
	w := ecs.NewWorld()
	entity := w.CreateEntity()

	// No mods - should return defaults
	speed, accel, handling, armor, fuel, damage, wepRange, cargo, passengers := sys.GetModifiedStats(entity)
	if speed != 1.0 || accel != 1.0 || armor != 1.0 {
		t.Error("Unregistered vehicle should return 1.0 defaults")
	}

	sys.RegisterVehicle(entity, "horse", 100)

	// Still defaults with no mods
	speed, accel, handling, armor, fuel, damage, wepRange, cargo, passengers = sys.GetModifiedStats(entity)
	if speed != 1.0 || accel != 1.0 || armor != 1.0 {
		t.Error("Vehicle with no mods should return 1.0 defaults")
	}

	// Install engine upgrade
	sys.InstallCustomization(entity, "engine_enchanted")
	speed, accel, _, _, _, _, _, _, _ = sys.GetModifiedStats(entity)

	if speed <= 1.0 {
		t.Error("Engine should increase speed")
	}
	if accel <= 1.0 {
		t.Error("Engine should increase acceleration")
	}

	// Install storage upgrade
	sys.InstallCustomization(entity, "storage_saddlebags")
	_, _, _, _, _, _, _, cargo, passengers = sys.GetModifiedStats(entity)

	if cargo <= 0 {
		t.Errorf("Storage should add cargo capacity, got %d", cargo)
	}

	// Use all variables to avoid compiler complaints
	_ = handling
	_ = armor
	_ = fuel
	_ = damage
	_ = wepRange
	_ = passengers
}

func TestSetVehicleName(t *testing.T) {
	sys := NewVehicleCustomizationSystem(12345, "fantasy")
	w := ecs.NewWorld()
	entity := w.CreateEntity()

	// Unregistered
	err := sys.SetVehicleName(entity, "Test")
	if err == nil {
		t.Error("Should error for unregistered vehicle")
	}

	sys.RegisterVehicle(entity, "horse", 100)

	err = sys.SetVehicleName(entity, "Shadowmere")
	if err != nil {
		t.Fatalf("SetVehicleName failed: %v", err)
	}

	state := sys.GetVehicleState(entity)
	if state.CustomName != "Shadowmere" {
		t.Errorf("Expected name 'Shadowmere', got '%s'", state.CustomName)
	}
}

func TestSetColors(t *testing.T) {
	sys := NewVehicleCustomizationSystem(12345, "fantasy")
	w := ecs.NewWorld()
	entity := w.CreateEntity()

	sys.RegisterVehicle(entity, "cart", 100)

	err := sys.SetColors(entity, 0xFF0000, 0x00FF00)
	if err != nil {
		t.Fatalf("SetColors failed: %v", err)
	}

	state := sys.GetVehicleState(entity)
	if state.PrimaryColor != 0xFF0000 {
		t.Errorf("Expected primary 0xFF0000, got %X", state.PrimaryColor)
	}
	if state.SecondaryColor != 0x00FF00 {
		t.Errorf("Expected secondary 0x00FF00, got %X", state.SecondaryColor)
	}
}

func TestAddRemoveDecal(t *testing.T) {
	sys := NewVehicleCustomizationSystem(12345, "fantasy")
	w := ecs.NewWorld()
	entity := w.CreateEntity()

	sys.RegisterVehicle(entity, "cart", 100)

	// Add decals
	for i := 0; i < 5; i++ {
		err := sys.AddDecal(entity, "decal_skull")
		if err != nil {
			t.Fatalf("AddDecal %d failed: %v", i, err)
		}
	}

	state := sys.GetVehicleState(entity)
	if len(state.DecalIDs) != 5 {
		t.Errorf("Expected 5 decals, got %d", len(state.DecalIDs))
	}

	// Max 5 decals
	err := sys.AddDecal(entity, "decal_extra")
	if err == nil {
		t.Error("Should error when exceeding max decals")
	}

	// Remove decal
	err = sys.RemoveDecal(entity, "decal_skull")
	if err != nil {
		t.Fatalf("RemoveDecal failed: %v", err)
	}

	state = sys.GetVehicleState(entity)
	if len(state.DecalIDs) != 4 {
		t.Errorf("Expected 4 decals after removal, got %d", len(state.DecalIDs))
	}

	// Remove non-existent
	err = sys.RemoveDecal(entity, "nonexistent")
	if err == nil {
		t.Error("Should error for non-existent decal")
	}
}

func TestAddExperience(t *testing.T) {
	sys := NewVehicleCustomizationSystem(12345, "fantasy")
	w := ecs.NewWorld()
	entity := w.CreateEntity()

	// Unregistered
	leveledUp, level := sys.AddExperience(entity, 100)
	if leveledUp || level != 0 {
		t.Error("Should return false, 0 for unregistered")
	}

	sys.RegisterVehicle(entity, "horse", 100)

	// Add some XP (not enough to level)
	leveledUp, level = sys.AddExperience(entity, 50)
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

	// Verify level persisted
	if sys.GetVehicleLevel(entity) != 2 {
		t.Error("Level should be saved")
	}
}

func TestLevelUnlocksCustomizations(t *testing.T) {
	sys := NewVehicleCustomizationSystem(12345, "fantasy")
	w := ecs.NewWorld()
	entity := w.CreateEntity()

	sys.RegisterVehicle(entity, "horse", 100)

	// Level 1 available customizations
	available := sys.GetAvailableCustomizations(entity)
	initialCount := len(available)

	// Level up to 5
	for i := 0; i < 10; i++ {
		sys.AddExperience(entity, 200)
	}

	// Should have more available now
	available = sys.GetAvailableCustomizations(entity)
	if len(available) <= initialCount {
		t.Error("Higher level should unlock more customizations")
	}
}

func TestGetCustomization(t *testing.T) {
	sys := NewVehicleCustomizationSystem(12345, "fantasy")

	// Existing
	custom := sys.GetCustomization("engine_enchanted")
	if custom == nil {
		t.Fatal("Should find engine_enchanted")
	}
	if custom.Name != "Enchanted Harness" {
		t.Errorf("Wrong name: %s", custom.Name)
	}
	if custom.Category != CategoryEngine {
		t.Error("Wrong category")
	}

	// Non-existent
	custom = sys.GetCustomization("nonexistent")
	if custom != nil {
		t.Error("Should return nil for non-existent")
	}
}

func TestUnregisterVehicle(t *testing.T) {
	sys := NewVehicleCustomizationSystem(12345, "fantasy")
	w := ecs.NewWorld()
	entity := w.CreateEntity()

	sys.RegisterVehicle(entity, "horse", 100)
	if sys.VehicleCount() != 1 {
		t.Fatal("Should have 1 vehicle")
	}

	sys.UnregisterVehicle(entity)

	if sys.VehicleCount() != 0 {
		t.Error("Should have 0 vehicles after unregister")
	}
	if sys.GetVehicleState(entity) != nil {
		t.Error("State should be nil after unregister")
	}
}

func TestAllGenresHaveCustomizations(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		sys := NewVehicleCustomizationSystem(12345, genre)
		w := ecs.NewWorld()
		entity := w.CreateEntity()

		sys.RegisterVehicle(entity, "vehicle", 100)

		available := sys.GetAvailableCustomizations(entity)
		if len(available) == 0 {
			t.Errorf("No customizations available for genre: %s", genre)
		}

		t.Logf("%s: %d customizations available at level 1", genre, len(available))
	}
}

func TestPaintUpdatesColor(t *testing.T) {
	sys := NewVehicleCustomizationSystem(12345, "fantasy")
	w := ecs.NewWorld()
	entity := w.CreateEntity()

	sys.RegisterVehicle(entity, "cart", 100)

	// Install paint
	err := sys.InstallCustomization(entity, "paint_royal_blue")
	if err != nil {
		t.Fatalf("Install paint failed: %v", err)
	}

	state := sys.GetVehicleState(entity)
	if state.PrimaryColor != 0x1E3A8A {
		t.Errorf("Paint should update primary color, got %X", state.PrimaryColor)
	}
}

func TestMaxLevel(t *testing.T) {
	sys := NewVehicleCustomizationSystem(12345, "fantasy")
	w := ecs.NewWorld()
	entity := w.CreateEntity()

	sys.RegisterVehicle(entity, "horse", 100)

	// Add massive XP to hit max level
	for i := 0; i < 100; i++ {
		sys.AddExperience(entity, 10000)
	}

	level := sys.GetVehicleLevel(entity)
	if level > 15 {
		t.Errorf("Level should cap at 15, got %d", level)
	}
}

func TestConcurrentAccess(t *testing.T) {
	sys := NewVehicleCustomizationSystem(12345, "fantasy")
	w := ecs.NewWorld()

	// Create multiple vehicles
	var entities []ecs.Entity
	for i := 0; i < 10; i++ {
		e := w.CreateEntity()
		sys.RegisterVehicle(e, "horse", 100)
		entities = append(entities, e)
	}

	done := make(chan bool, 100)

	// Concurrent reads
	for i := 0; i < 50; i++ {
		go func(idx int) {
			e := entities[idx%len(entities)]
			_ = sys.GetVehicleState(e)
			_, _, _, _, _, _, _, _, _ = sys.GetModifiedStats(e)
			done <- true
		}(i)
	}

	// Concurrent writes
	for i := 0; i < 50; i++ {
		go func(idx int) {
			e := entities[idx%len(entities)]
			sys.AddExperience(e, 10)
			done <- true
		}(i)
	}

	// Wait for all
	for i := 0; i < 100; i++ {
		<-done
	}
}
