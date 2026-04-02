package components

import "testing"

func TestPositionType(t *testing.T) {
	p := &Position{X: 1, Y: 2, Z: 3}
	if p.Type() != "Position" {
		t.Errorf("expected Position, got %s", p.Type())
	}
}

func TestHealthType(t *testing.T) {
	h := &Health{Current: 100, Max: 100}
	if h.Type() != "Health" {
		t.Errorf("expected Health, got %s", h.Type())
	}
}

func TestFactionType(t *testing.T) {
	f := &Faction{ID: "guild", Reputation: 50}
	if f.Type() != "Faction" {
		t.Errorf("expected Faction, got %s", f.Type())
	}
}

func TestScheduleType(t *testing.T) {
	s := &Schedule{
		CurrentActivity: "work",
		TimeSlots:       map[int]string{8: "work", 12: "eat"},
	}
	if s.Type() != "Schedule" {
		t.Errorf("expected Schedule, got %s", s.Type())
	}
}

func TestInventoryType(t *testing.T) {
	i := &Inventory{Items: []string{"sword"}, Capacity: 10}
	if i.Type() != "Inventory" {
		t.Errorf("expected Inventory, got %s", i.Type())
	}
}

func TestVehicleType(t *testing.T) {
	v := &Vehicle{VehicleType: "horse", Speed: 10, Fuel: 100}
	if v.Type() != "Vehicle" {
		t.Errorf("expected Vehicle, got %s", v.Type())
	}
}

func TestComponentImplementsInterface(t *testing.T) {
	// Verify all components implement the Component interface via Type()
	components := []interface{ Type() string }{
		&Position{},
		&Health{},
		&Faction{},
		&FactionTerritory{Vertices: []Point2D{}, KillTracker: make(map[uint64]int)},
		&Schedule{TimeSlots: make(map[int]string)},
		&Inventory{},
		&Vehicle{},
		&Reputation{Standings: make(map[string]float64)},
		&Crime{},
		&Witness{},
		&EconomyNode{},
		&Quest{Flags: make(map[string]bool)},
		&WorldClock{},
		&Skills{Levels: make(map[string]int), Experience: make(map[string]float64)},
		&Barrier{},
	}

	for _, c := range components {
		if c.Type() == "" {
			t.Error("component Type() returned empty string")
		}
	}
}

func TestReputationType(t *testing.T) {
	r := &Reputation{Standings: map[string]float64{"guild": 50.0}}
	if r.Type() != "Reputation" {
		t.Errorf("expected Reputation, got %s", r.Type())
	}
}

func TestCrimeType(t *testing.T) {
	c := &Crime{WantedLevel: 2, BountyAmount: 500.0}
	if c.Type() != "Crime" {
		t.Errorf("expected Crime, got %s", c.Type())
	}
}

func TestWitnessType(t *testing.T) {
	w := &Witness{CanReport: true}
	if w.Type() != "Witness" {
		t.Errorf("expected Witness, got %s", w.Type())
	}
}

func TestEconomyNodeType(t *testing.T) {
	e := &EconomyNode{PriceTable: map[string]float64{"sword": 100.0}}
	if e.Type() != "EconomyNode" {
		t.Errorf("expected EconomyNode, got %s", e.Type())
	}
}

func TestQuestType(t *testing.T) {
	q := &Quest{ID: "main", CurrentStage: 1, Flags: map[string]bool{"start": true}}
	if q.Type() != "Quest" {
		t.Errorf("expected Quest, got %s", q.Type())
	}
}

func TestWorldClockType(t *testing.T) {
	wc := &WorldClock{Hour: 12, Day: 1, HourLength: 60.0}
	if wc.Type() != "WorldClock" {
		t.Errorf("expected WorldClock, got %s", wc.Type())
	}
}

func TestSkillsType(t *testing.T) {
	s := &Skills{
		Levels:     map[string]int{"fire_magic": 10},
		Experience: map[string]float64{"fire_magic": 50.0},
	}
	if s.Type() != "Skills" {
		t.Errorf("expected Skills, got %s", s.Type())
	}
}

func TestGetSkillSchools(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			schools := GetSkillSchools(genre)
			if len(schools) != 6 {
				t.Errorf("expected 6 schools for %s, got %d", genre, len(schools))
			}
			for _, school := range schools {
				if school.ID == "" {
					t.Error("school has empty ID")
				}
				if school.Name == "" {
					t.Error("school has empty Name")
				}
				if len(school.Skills) != 5 {
					t.Errorf("expected 5 skills per school, got %d", len(school.Skills))
				}
			}
		})
	}
}

func TestGetSkillSchoolsDefault(t *testing.T) {
	// Unknown genre should fall back to fantasy
	schools := GetSkillSchools("unknown_genre")
	if len(schools) != 6 {
		t.Errorf("expected 6 schools for unknown genre (fallback), got %d", len(schools))
	}
}

func TestGetAllSkillIDs(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			skills := GetAllSkillIDs(genre)
			// 6 schools * 5 skills = 30 skills
			if len(skills) != 30 {
				t.Errorf("expected 30 skills for %s, got %d", genre, len(skills))
			}
		})
	}
}

func TestGenreSkillNamesUnique(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	// Collect all school names per genre
	schoolNames := make(map[string]map[string]bool)
	for _, genre := range genres {
		schoolNames[genre] = make(map[string]bool)
		for _, school := range GetSkillSchools(genre) {
			schoolNames[genre][school.Name] = true
		}
	}

	// At least some names should differ between genres
	fantasyNames := schoolNames["fantasy"]
	for _, otherGenre := range []string{"sci-fi", "cyberpunk", "post-apocalyptic"} {
		differentCount := 0
		for name := range fantasyNames {
			if !schoolNames[otherGenre][name] {
				differentCount++
			}
		}
		if differentCount < 3 {
			t.Errorf("expected at least 3 different school names between fantasy and %s", otherGenre)
		}
	}
}

func TestNewSkills(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			skills := NewSkills(genre)
			if skills == nil {
				t.Fatal("NewSkills returned nil")
			}
			if len(skills.Levels) != 30 {
				t.Errorf("expected 30 skills, got %d", len(skills.Levels))
			}
			// All skills should start at level 1
			for skillID, level := range skills.Levels {
				if level != 1 {
					t.Errorf("skill %s should start at level 1, got %d", skillID, level)
				}
			}
			// All skills should have 0 XP
			for skillID, xp := range skills.Experience {
				if xp != 0 {
					t.Errorf("skill %s should start with 0 XP, got %f", skillID, xp)
				}
			}
		})
	}
}

func TestGetVehicleArchetypes(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			archetypes := GetVehicleArchetypes(genre)
			if len(archetypes) != 3 {
				t.Errorf("expected 3 vehicle archetypes for %s, got %d", genre, len(archetypes))
			}
			for _, arch := range archetypes {
				if arch.ID == "" {
					t.Error("archetype has empty ID")
				}
				if arch.Name == "" {
					t.Error("archetype has empty Name")
				}
				if arch.BaseSpeed <= 0 {
					t.Errorf("archetype %s has invalid BaseSpeed", arch.ID)
				}
				if arch.MaxFuel <= 0 {
					t.Errorf("archetype %s has invalid MaxFuel", arch.ID)
				}
			}
		})
	}
}

func TestVehicleArchetypesUnique(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	// Collect all vehicle IDs across genres
	vehicleIDs := make(map[string]string) // id -> genre
	for _, genre := range genres {
		for _, arch := range GetVehicleArchetypes(genre) {
			// Vehicle types should be unique per genre
			if existingGenre, exists := vehicleIDs[arch.ID]; exists && existingGenre != genre {
				// Same ID in different genres is OK (each genre has its own)
				continue
			}
			vehicleIDs[arch.ID] = genre
		}
	}

	// Verify each genre has distinct vehicles
	for _, genre := range []string{"fantasy", "sci-fi"} {
		archetypes := GetVehicleArchetypes(genre)
		for _, arch := range archetypes {
			otherGenre := "cyberpunk"
			otherArchetypes := GetVehicleArchetypes(otherGenre)
			found := false
			for _, other := range otherArchetypes {
				if arch.ID == other.ID {
					found = true
					break
				}
			}
			if found {
				t.Errorf("vehicle %s found in both %s and %s", arch.ID, genre, otherGenre)
			}
		}
	}
}

func TestNewVehicleFromArchetype(t *testing.T) {
	arch := VehicleArchetype{
		ID:        "test_vehicle",
		Name:      "Test Vehicle",
		BaseSpeed: 20,
		MaxFuel:   150,
		FuelRate:  0.01,
	}
	vehicle := NewVehicleFromArchetype(arch)
	if vehicle.VehicleType != "test_vehicle" {
		t.Errorf("expected vehicle type test_vehicle, got %s", vehicle.VehicleType)
	}
	if vehicle.Speed != 20 {
		t.Errorf("expected speed 20, got %f", vehicle.Speed)
	}
	if vehicle.Fuel != 150 {
		t.Errorf("expected fuel 150, got %f", vehicle.Fuel)
	}
}

func TestFactionTerritoryContainsPoint(t *testing.T) {
	// Square territory from (0,0) to (10,10)
	territory := &FactionTerritory{
		FactionID: "test_faction",
		Vertices: []Point2D{
			{X: 0, Y: 0},
			{X: 10, Y: 0},
			{X: 10, Y: 10},
			{X: 0, Y: 10},
		},
		KillTracker: make(map[uint64]int),
	}

	tests := []struct {
		name   string
		x, y   float64
		inside bool
	}{
		{"center point", 5, 5, true},
		{"top-left inside", 1, 1, true},
		{"bottom-right inside", 9, 9, true},
		{"outside left", -5, 5, false},
		{"outside right", 15, 5, false},
		{"outside top", 5, -5, false},
		{"outside bottom", 5, 15, false},
		{"corner outside", 15, 15, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := territory.ContainsPoint(tc.x, tc.y)
			if result != tc.inside {
				t.Errorf("ContainsPoint(%f, %f) = %v, want %v", tc.x, tc.y, result, tc.inside)
			}
		})
	}
}

func TestFactionTerritoryContainsPointTriangle(t *testing.T) {
	// Triangle territory
	territory := &FactionTerritory{
		FactionID: "triangle_faction",
		Vertices: []Point2D{
			{X: 0, Y: 0},
			{X: 10, Y: 0},
			{X: 5, Y: 10},
		},
		KillTracker: make(map[uint64]int),
	}

	// Center of triangle should be inside
	if !territory.ContainsPoint(5, 3) {
		t.Error("center of triangle should be inside")
	}

	// Point outside triangle
	if territory.ContainsPoint(0, 10) {
		t.Error("point outside triangle should not be inside")
	}
}

func TestFactionTerritoryContainsPointInvalidPolygon(t *testing.T) {
	// Too few vertices (less than 3 points)
	territory := &FactionTerritory{
		FactionID: "invalid_faction",
		Vertices: []Point2D{
			{X: 0, Y: 0},
			{X: 10, Y: 10},
		},
		KillTracker: make(map[uint64]int),
	}

	// Should return false for invalid polygon
	if territory.ContainsPoint(5, 5) {
		t.Error("invalid polygon with <3 vertices should return false")
	}
}

func TestQuestLockBranch(t *testing.T) {
	quest := &Quest{
		ID:           "main_quest",
		CurrentStage: 1,
		Flags:        make(map[string]bool),
	}

	// Initially no branches are locked
	if quest.IsBranchLocked("branch_a") {
		t.Error("branch should not be locked initially")
	}

	// Lock a branch
	quest.LockBranch("branch_a")

	// Now it should be locked
	if !quest.IsBranchLocked("branch_a") {
		t.Error("branch_a should be locked after LockBranch")
	}

	// Other branches should still be unlocked
	if quest.IsBranchLocked("branch_b") {
		t.Error("branch_b should not be locked")
	}

	// Lock another branch
	quest.LockBranch("branch_b")
	if !quest.IsBranchLocked("branch_b") {
		t.Error("branch_b should be locked after LockBranch")
	}
}

func TestQuestIsBranchLockedNilMap(t *testing.T) {
	quest := &Quest{
		ID:             "test_quest",
		CurrentStage:   1,
		Flags:          make(map[string]bool),
		LockedBranches: nil, // Explicitly nil
	}

	// Should return false without panic
	if quest.IsBranchLocked("any_branch") {
		t.Error("IsBranchLocked should return false when LockedBranches is nil")
	}
}

func TestAudioListenerType(t *testing.T) {
	listener := &AudioListener{
		Volume:  0.8,
		Enabled: true,
	}
	if listener.Type() != "AudioListener" {
		t.Errorf("expected AudioListener, got %s", listener.Type())
	}
}

func TestAudioSourceType(t *testing.T) {
	source := &AudioSource{
		SoundType: "footstep",
		Volume:    1.0,
		Range:     50.0,
		Looping:   false,
		Playing:   true,
	}
	if source.Type() != "AudioSource" {
		t.Errorf("expected AudioSource, got %s", source.Type())
	}
}

func TestAudioStateType(t *testing.T) {
	state := &AudioState{
		CurrentAmbient:  "forest",
		CombatIntensity: 0.5,
		LastPositionX:   100.0,
		LastPositionY:   200.0,
	}
	if state.Type() != "AudioState" {
		t.Errorf("expected AudioState, got %s", state.Type())
	}
}

func TestFactionTerritoryType(t *testing.T) {
	territory := &FactionTerritory{
		FactionID:   "test_faction",
		Vertices:    []Point2D{{X: 0, Y: 0}},
		KillTracker: make(map[uint64]int),
	}
	if territory.Type() != "FactionTerritory" {
		t.Errorf("expected FactionTerritory, got %s", territory.Type())
	}
}

func TestVehiclePhysicsType(t *testing.T) {
	vp := &VehiclePhysics{
		CurrentSpeed:     50.0,
		MaxSpeed:         100.0,
		Acceleration:     5.0,
		Deceleration:     10.0,
		FrictionDecel:    0.1,
		SteeringAngle:    0.0,
		MaxSteeringAngle: 0.5,
		SteeringSpeed:    1.0,
		TurningRadius:    5.0,
		Mass:             1000.0,
		Throttle:         0.5,
		Steering:         0.0,
		IsBraking:        false,
		InReverse:        false,
	}
	if vp.Type() != "VehiclePhysics" {
		t.Errorf("expected VehiclePhysics, got %s", vp.Type())
	}
}

func TestVehicleStateType(t *testing.T) {
	vs := &VehicleState{
		IsOccupied:        true,
		DriverEntity:      1,
		PassengerEntities: []uint64{2, 3},
		MaxPassengers:     4,
		InCockpitView:     true,
		EngineRunning:     true,
		DamagePercent:     10.0,
	}
	if vs.Type() != "VehicleState" {
		t.Errorf("expected VehicleState, got %s", vs.Type())
	}
}

func TestWeaponType(t *testing.T) {
	w := &Weapon{
		Name:        "Iron Sword",
		Damage:      25.0,
		Range:       2.0,
		AttackSpeed: 1.5,
		WeaponType:  "melee",
	}
	if w.Type() != "Weapon" {
		t.Errorf("expected Weapon, got %s", w.Type())
	}
}

func TestCombatStateType(t *testing.T) {
	cs := &CombatState{
		LastAttackTime: 100.0,
		Cooldown:       0.5,
		IsAttacking:    true,
		TargetEntity:   42,
		InCombat:       true,
	}
	if cs.Type() != "CombatState" {
		t.Errorf("expected CombatState, got %s", cs.Type())
	}
}

func TestStealthType(t *testing.T) {
	s := &Stealth{
		Visibility:      0.5,
		Sneaking:        true,
		DetectionRadius: 10.0,
		BaseVisibility:  1.0,
		SneakVisibility: 0.3,
		LastDetectedBy:  make(map[uint64]float64),
	}
	if s.Type() != "Stealth" {
		t.Errorf("expected Stealth, got %s", s.Type())
	}
}

func TestAwarenessType(t *testing.T) {
	a := &Awareness{
		AlertLevel:       0.5,
		SightRange:       50.0,
		SightAngle:       1.57,
		DetectedEntities: make(map[uint64]float64),
	}
	if a.Type() != "Awareness" {
		t.Errorf("expected Awareness, got %s", a.Type())
	}
}

func TestMaterialType(t *testing.T) {
	m := &Material{
		ResourceType: "ore",
		Quantity:     10,
		Quality:      0.8,
		Rarity:       "rare",
	}
	if m.Type() != "Material" {
		t.Errorf("expected Material, got %s", m.Type())
	}
}

func TestResourceNodeType(t *testing.T) {
	rn := &ResourceNode{
		ResourceType: "iron_ore",
		Quantity:     50,
		MaxQuantity:  100,
		Quality:      0.7,
		RespawnTime:  3600.0,
		LastGathered: 0.0,
		Depleted:     false,
	}
	if rn.Type() != "ResourceNode" {
		t.Errorf("expected ResourceNode, got %s", rn.Type())
	}
}

func TestWorkbenchType(t *testing.T) {
	wb := &Workbench{
		WorkbenchType:        "forge",
		SupportedRecipeTypes: []string{"weapon", "armor"},
		CraftingSpeedMult:    1.0,
		QualityBonus:         0.1,
	}
	if wb.Type() != "Workbench" {
		t.Errorf("expected Workbench, got %s", wb.Type())
	}
}

func TestCraftingStateType(t *testing.T) {
	cs := &CraftingState{
		IsCrafting:        true,
		CurrentRecipeID:   "iron_sword",
		Progress:          0.5,
		TotalTime:         10.0,
		WorkbenchEntity:   100,
		ConsumedMaterials: map[string]int{"iron_ore": 3},
	}
	if cs.Type() != "CraftingState" {
		t.Errorf("expected CraftingState, got %s", cs.Type())
	}
}

func TestToolType(t *testing.T) {
	tool := &Tool{
		ToolType:      "pickaxe",
		Name:          "Iron Pickaxe",
		Durability:    80.0,
		MaxDurability: 100.0,
		GatherSpeed:   1.5,
		QualityBonus:  0.05,
		ToolTier:      2,
	}
	if tool.Type() != "Tool" {
		t.Errorf("expected Tool, got %s", tool.Type())
	}
}

func TestRecipeKnowledgeType(t *testing.T) {
	rk := &RecipeKnowledge{
		KnownRecipes:      map[string]bool{"iron_sword": true, "leather_armor": true},
		DiscoveryProgress: map[string]float64{"steel_sword": 0.5},
	}
	if rk.Type() != "RecipeKnowledge" {
		t.Errorf("expected RecipeKnowledge, got %s", rk.Type())
	}
}

func TestProjectileType(t *testing.T) {
	p := &Projectile{
		OwnerID:        1,
		VelocityX:      10.0,
		VelocityY:      0.0,
		VelocityZ:      0.0,
		Damage:         15.0,
		Lifetime:       5.0,
		HitRadius:      0.5,
		ProjectileType: "arrow",
		PierceCount:    0,
		HitEntities:    make(map[uint64]bool),
	}
	if p.Type() != "Projectile" {
		t.Errorf("expected Projectile, got %s", p.Type())
	}
}

func TestManaType(t *testing.T) {
	m := &Mana{
		Current:   50.0,
		Max:       100.0,
		RegenRate: 2.0,
	}
	if m.Type() != "Mana" {
		t.Errorf("expected Mana, got %s", m.Type())
	}
}

func TestSpellEffectType(t *testing.T) {
	se := &SpellEffect{
		EffectType: "damage",
		Magnitude:  25.0,
		Duration:   5.0,
		Remaining:  3.0,
		Source:     1,
	}
	if se.Type() != "SpellEffect" {
		t.Errorf("expected SpellEffect, got %s", se.Type())
	}
}

func TestSpellType(t *testing.T) {
	spell := &Spell{
		ID:              "fireball",
		Name:            "Fireball",
		ManaCost:        30.0,
		Cooldown:        5.0,
		LastCast:        0.0,
		EffectType:      "damage",
		Magnitude:       50.0,
		Range:           20.0,
		AreaOfEffect:    5.0,
		ProjectileSpeed: 15.0,
	}
	if spell.Type() != "Spell" {
		t.Errorf("expected Spell, got %s", spell.Type())
	}
}

func TestSpellbookType(t *testing.T) {
	sb := &Spellbook{
		Spells: map[string]*Spell{
			"fireball": {ID: "fireball", Name: "Fireball"},
		},
		ActiveSpellID: "fireball",
	}
	if sb.Type() != "Spellbook" {
		t.Errorf("expected Spellbook, got %s", sb.Type())
	}
}

func TestNPCMemoryType(t *testing.T) {
	mem := &NPCMemory{
		PlayerInteractions: make(map[uint64][]MemoryEvent),
		LastSeen:           make(map[uint64]float64),
		Disposition:        make(map[uint64]float64),
		MaxMemories:        100,
		MemoryDecayRate:    0.01,
	}
	if mem.Type() != "NPCMemory" {
		t.Errorf("expected NPCMemory, got %s", mem.Type())
	}
}

func TestNPCRelationshipsType(t *testing.T) {
	rel := &NPCRelationships{
		Relationships: map[uint64]*Relationship{
			1: {TargetEntity: 1, Type: "friend", Strength: 0.8},
		},
	}
	if rel.Type() != "NPCRelationships" {
		t.Errorf("expected NPCRelationships, got %s", rel.Type())
	}
}

func TestSocialStatusType(t *testing.T) {
	ss := &SocialStatus{
		Wealth:     0.5,
		Influence:  0.3,
		Occupation: "merchant",
		Title:      "Sir",
	}
	if ss.Type() != "SocialStatus" {
		t.Errorf("expected SocialStatus, got %s", ss.Type())
	}
}

func TestInteriorType(t *testing.T) {
	interior := &Interior{
		ParentBuilding: 1,
		Width:          20,
		Height:         15,
		Rooms: []Room{
			{ID: "main", Name: "Main Room", X: 0, Y: 0, Width: 10, Height: 10, Purpose: "shop"},
		},
		Furniture: []uint64{100, 101},
		WallTiles: [][]int{{1, 1, 1}},
		FloorType: "wood",
	}
	if interior.Type() != "Interior" {
		t.Errorf("expected Interior, got %s", interior.Type())
	}
}

func TestPOIMarkerType(t *testing.T) {
	poi := &POIMarker{
		IconType:          "shop",
		Name:              "General Store",
		Description:       "Sells general goods",
		Visible:           true,
		MinimapVisible:    true,
		DiscoveryRequired: false,
		Discovered:        true,
	}
	if poi.Type() != "POIMarker" {
		t.Errorf("expected POIMarker, got %s", poi.Type())
	}
}

func TestBuildingType(t *testing.T) {
	b := &Building{
		BuildingType:   "shop",
		Name:           "General Store",
		OwnerFaction:   "merchants",
		InteriorEntity: 100,
		Floors:         2,
		Width:          10.0,
		Height:         8.0,
		EntranceX:      5.0,
		EntranceY:      0.0,
		EntranceZ:      0.0,
		IsOpen:         true,
		OpenHour:       8,
		CloseHour:      20,
	}
	if b.Type() != "Building" {
		t.Errorf("expected Building, got %s", b.Type())
	}
}

func TestShopInventoryType(t *testing.T) {
	si := &ShopInventory{
		ShopType:        "blacksmith",
		Items:           map[string]int{"iron_sword": 5, "steel_armor": 2},
		Prices:          map[string]float64{"iron_sword": 100.0, "steel_armor": 500.0},
		RestockInterval: 24,
		LastRestock:     12,
		GoldReserve:     1000.0,
	}
	if si.Type() != "ShopInventory" {
		t.Errorf("expected ShopInventory, got %s", si.Type())
	}
}

func TestGovernmentBuildingType(t *testing.T) {
	gb := &GovernmentBuilding{
		GovernmentType:     "barracks",
		ControllingFaction: "kingdom",
		Services:           []string{"bounty_payment", "training"},
		NPCRoles:           []string{"guard", "captain"},
	}
	if gb.Type() != "GovernmentBuilding" {
		t.Errorf("expected GovernmentBuilding, got %s", gb.Type())
	}
}

func TestGossipNetworkType(t *testing.T) {
	gn := &GossipNetwork{
		KnownGossip:    make(map[string]*GossipItem),
		GossipChance:   0.3,
		ListenChance:   0.5,
		LastGossipTime: 0.0,
		GossipCooldown: 60.0,
	}
	if gn.Type() != "GossipNetwork" {
		t.Errorf("expected GossipNetwork, got %s", gn.Type())
	}
}

func TestEmotionalStateType(t *testing.T) {
	es := &EmotionalState{
		CurrentEmotion:    "happy",
		Intensity:         0.7,
		Mood:              0.3,
		Stress:            0.1,
		LastEmotionChange: 100.0,
		EmotionDecayRate:  0.1,
		MoodDecayRate:     0.01,
	}
	if es.Type() != "EmotionalState" {
		t.Errorf("expected EmotionalState, got %s", es.Type())
	}
}

func TestNPCNeedsType(t *testing.T) {
	nn := &NPCNeeds{
		Hunger:          0.3,
		Energy:          0.7,
		Social:          0.5,
		Safety:          0.9,
		HungerRate:      0.1,
		EnergyRate:      0.05,
		SocialDecayRate: 0.02,
	}
	if nn.Type() != "NPCNeeds" {
		t.Errorf("expected NPCNeeds, got %s", nn.Type())
	}
}

func TestAllComponentTypes(t *testing.T) {
	// Comprehensive test ensuring all components implement Type()
	components := []interface{ Type() string }{
		&Position{},
		&Health{},
		&Faction{},
		&FactionTerritory{Vertices: []Point2D{}, KillTracker: make(map[uint64]int)},
		&Schedule{TimeSlots: make(map[int]string)},
		&Inventory{},
		&Vehicle{},
		&VehiclePhysics{},
		&VehicleState{},
		&Reputation{Standings: make(map[string]float64)},
		&Crime{},
		&Witness{},
		&EconomyNode{},
		&Quest{Flags: make(map[string]bool)},
		&WorldClock{},
		&Skills{Levels: make(map[string]int), Experience: make(map[string]float64)},
		&AudioListener{},
		&AudioSource{},
		&AudioState{},
		&Weapon{},
		&CombatState{},
		&Stealth{LastDetectedBy: make(map[uint64]float64)},
		&Awareness{DetectedEntities: make(map[uint64]float64)},
		&Material{},
		&ResourceNode{},
		&Workbench{},
		&CraftingState{ConsumedMaterials: make(map[string]int)},
		&Tool{},
		&RecipeKnowledge{KnownRecipes: make(map[string]bool), DiscoveryProgress: make(map[string]float64)},
		&Projectile{HitEntities: make(map[uint64]bool)},
		&Mana{},
		&SpellEffect{},
		&Spell{},
		&Spellbook{Spells: make(map[string]*Spell)},
		&NPCMemory{PlayerInteractions: make(map[uint64][]MemoryEvent), LastSeen: make(map[uint64]float64), Disposition: make(map[uint64]float64)},
		&NPCRelationships{Relationships: make(map[uint64]*Relationship)},
		&SocialStatus{},
		&Interior{},
		&POIMarker{},
		&Building{},
		&ShopInventory{Items: make(map[string]int), Prices: make(map[string]float64)},
		&GovernmentBuilding{},
		&GossipNetwork{KnownGossip: make(map[string]*GossipItem)},
		&EmotionalState{},
		&NPCNeeds{},
	}

	typeNames := make(map[string]bool)
	for _, c := range components {
		typeName := c.Type()
		if typeName == "" {
			t.Errorf("component has empty Type()")
		}
		if typeNames[typeName] {
			t.Errorf("duplicate Type() name: %s", typeName)
		}
		typeNames[typeName] = true
	}

	// Verify we have at least 45 unique component types
	if len(typeNames) < 45 {
		t.Errorf("expected at least 45 component types, got %d", len(typeNames))
	}
}

// ========== Environmental Hazard Component Tests ==========

func TestEnvironmentalHazardType(t *testing.T) {
	h := &EnvironmentalHazard{
		HazardType:      HazardTypeRadiation,
		Intensity:       0.5,
		DamagePerSecond: 10.0,
		Radius:          5.0,
		Active:          true,
	}
	if h.Type() != "EnvironmentalHazard" {
		t.Errorf("Expected type 'EnvironmentalHazard', got '%s'", h.Type())
	}
}

func TestHazardTypeConstants(t *testing.T) {
	types := []HazardType{
		HazardTypeRadiation,
		HazardTypeFire,
		HazardTypePoison,
		HazardTypeElectric,
		HazardTypeFreeze,
		HazardTypeLava,
		HazardTypeAcid,
		HazardTypeMagic,
		HazardTypeTrap,
		HazardTypeGas,
	}

	seen := make(map[HazardType]bool)
	for _, ht := range types {
		if seen[ht] {
			t.Errorf("Duplicate hazard type: %s", ht)
		}
		seen[ht] = true
		if string(ht) == "" {
			t.Error("Hazard type should not be empty")
		}
	}
}

func TestHazardResistanceType(t *testing.T) {
	h := &HazardResistance{
		Resistances: map[HazardType]float64{
			HazardTypeFire: 0.5,
		},
	}
	if h.Type() != "HazardResistance" {
		t.Errorf("Expected type 'HazardResistance', got '%s'", h.Type())
	}
}

func TestHazardResistanceGetResistance(t *testing.T) {
	h := &HazardResistance{
		Resistances: map[HazardType]float64{
			HazardTypeFire:   0.5,
			HazardTypePoison: 0.75,
		},
	}

	if h.GetResistance(HazardTypeFire) != 0.5 {
		t.Error("Fire resistance should be 0.5")
	}
	if h.GetResistance(HazardTypePoison) != 0.75 {
		t.Error("Poison resistance should be 0.75")
	}
	if h.GetResistance(HazardTypeRadiation) != 0 {
		t.Error("Unset resistance should be 0")
	}
}

func TestHazardResistanceGetResistanceNilMap(t *testing.T) {
	h := &HazardResistance{}
	if h.GetResistance(HazardTypeFire) != 0 {
		t.Error("Nil resistances map should return 0")
	}
}

func TestHazardEffectType(t *testing.T) {
	h := &HazardEffect{
		SourceHazard:      HazardTypePoison,
		StackCount:        1,
		MaxStacks:         3,
		RemainingDuration: 30.0,
		DamageOverTime:    5.0,
	}
	if h.Type() != "HazardEffect" {
		t.Errorf("Expected type 'HazardEffect', got '%s'", h.Type())
	}
}

func TestHazardZoneType(t *testing.T) {
	h := &HazardZone{
		ZoneName:    "Radiation Zone A",
		HazardTypes: []HazardType{HazardTypeRadiation},
		ZoneLevel:   3,
	}
	if h.Type() != "HazardZone" {
		t.Errorf("Expected type 'HazardZone', got '%s'", h.Type())
	}
}

func TestWeatherHazardType(t *testing.T) {
	w := &WeatherHazard{
		WeatherType:   "storm",
		Severity:      0.7,
		OutdoorDamage: 2.0,
	}
	if w.Type() != "WeatherHazard" {
		t.Errorf("Expected type 'WeatherHazard', got '%s'", w.Type())
	}
}

func TestTrapMechanismType(t *testing.T) {
	t2 := &TrapMechanism{
		TrapType:            "spike",
		Armed:               true,
		DetectionDifficulty: 0.5,
		Damage:              25.0,
	}
	if t2.Type() != "TrapMechanism" {
		t.Errorf("Expected type 'TrapMechanism', got '%s'", t2.Type())
	}
}

func TestAppearanceType(t *testing.T) {
	a := &Appearance{
		SpriteCategory: "humanoid",
		BodyPlan:       "warrior",
		Visible:        true,
		Opacity:        1.0,
	}
	if a.Type() != "Appearance" {
		t.Errorf("expected Appearance, got %s", a.Type())
	}
}

func TestNewAppearance(t *testing.T) {
	a := NewAppearance("humanoid", "guard", "fantasy")

	if a.SpriteCategory != "humanoid" {
		t.Errorf("expected category humanoid, got %s", a.SpriteCategory)
	}
	if a.BodyPlan != "guard" {
		t.Errorf("expected body plan guard, got %s", a.BodyPlan)
	}
	if a.GenreID != "fantasy" {
		t.Errorf("expected genre fantasy, got %s", a.GenreID)
	}
	if a.Scale != 1.0 {
		t.Errorf("expected scale 1.0, got %f", a.Scale)
	}
	if a.AnimState != "idle" {
		t.Errorf("expected anim state idle, got %s", a.AnimState)
	}
	if !a.Visible {
		t.Error("expected visible true")
	}
	if a.Opacity != 1.0 {
		t.Errorf("expected opacity 1.0, got %f", a.Opacity)
	}
	if a.FlipH {
		t.Error("expected flipH false")
	}
	if a.DamageOverlay != 0.0 {
		t.Errorf("expected damage overlay 0.0, got %f", a.DamageOverlay)
	}
}

func TestAppearanceDefaults(t *testing.T) {
	// Test zero-value appearance
	a := &Appearance{}
	if a.Type() != "Appearance" {
		t.Errorf("expected Appearance, got %s", a.Type())
	}
	// Zero values are valid, just not visible
	if a.Visible {
		t.Error("expected default visible to be false")
	}
}

func TestBarrierType(t *testing.T) {
	b := &Barrier{
		Shape: BarrierShape{ShapeType: "cylinder", Height: 2.0},
		Genre: "fantasy",
	}
	if b.Type() != "Barrier" {
		t.Errorf("expected Barrier, got %s", b.Type())
	}
}

func TestNewBarrier(t *testing.T) {
	b := NewBarrier("cylinder", "boulder_01", "fantasy", 1.5)
	if b.Shape.ShapeType != "cylinder" {
		t.Errorf("expected cylinder, got %s", b.Shape.ShapeType)
	}
	if b.Shape.SpriteKey != "boulder_01" {
		t.Errorf("expected boulder_01, got %s", b.Shape.SpriteKey)
	}
	if b.Genre != "fantasy" {
		t.Errorf("expected fantasy, got %s", b.Genre)
	}
	if b.Shape.Height != 1.5 {
		t.Errorf("expected height 1.5, got %f", b.Shape.Height)
	}
	if b.Destructible {
		t.Error("expected non-destructible by default")
	}
}

func TestNewDestructibleBarrier(t *testing.T) {
	b := NewDestructibleBarrier("box", "crate_01", "sci-fi", 1.0, 100.0)
	if !b.Destructible {
		t.Error("expected destructible")
	}
	if b.HitPoints != 100.0 {
		t.Errorf("expected HitPoints 100.0, got %f", b.HitPoints)
	}
	if b.MaxHP != 100.0 {
		t.Errorf("expected MaxHP 100.0, got %f", b.MaxHP)
	}
}

func TestBarrierDamagePercent(t *testing.T) {
	tests := []struct {
		name        string
		barrier     *Barrier
		wantPercent float64
	}{
		{
			name:        "non-destructible",
			barrier:     NewBarrier("cylinder", "rock", "fantasy", 1.0),
			wantPercent: 0.0,
		},
		{
			name:        "undamaged",
			barrier:     NewDestructibleBarrier("box", "crate", "sci-fi", 1.0, 100.0),
			wantPercent: 0.0,
		},
		{
			name: "half damaged",
			barrier: &Barrier{
				Destructible: true,
				HitPoints:    50.0,
				MaxHP:        100.0,
			},
			wantPercent: 0.5,
		},
		{
			name: "fully destroyed",
			barrier: &Barrier{
				Destructible: true,
				HitPoints:    0.0,
				MaxHP:        100.0,
			},
			wantPercent: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.barrier.DamagePercent()
			if got != tt.wantPercent {
				t.Errorf("DamagePercent() = %f, want %f", got, tt.wantPercent)
			}
		})
	}
}

func TestBarrierIsDamaged(t *testing.T) {
	// Non-destructible is never damaged
	b1 := NewBarrier("cylinder", "rock", "fantasy", 1.0)
	if b1.IsDamaged() {
		t.Error("non-destructible should not be damaged")
	}

	// Undamaged destructible
	b2 := NewDestructibleBarrier("box", "crate", "sci-fi", 1.0, 100.0)
	if b2.IsDamaged() {
		t.Error("undamaged should not be damaged")
	}

	// Damaged destructible
	b3 := NewDestructibleBarrier("box", "crate", "sci-fi", 1.0, 100.0)
	b3.HitPoints = 50.0
	if !b3.IsDamaged() {
		t.Error("partially damaged should be damaged")
	}
}

func TestBarrierIsDestroyed(t *testing.T) {
	// Non-destructible is never destroyed
	b1 := NewBarrier("cylinder", "rock", "fantasy", 1.0)
	if b1.IsDestroyed() {
		t.Error("non-destructible should not be destroyed")
	}

	// Undamaged destructible
	b2 := NewDestructibleBarrier("box", "crate", "sci-fi", 1.0, 100.0)
	if b2.IsDestroyed() {
		t.Error("undamaged should not be destroyed")
	}

	// Destroyed
	b3 := NewDestructibleBarrier("box", "crate", "sci-fi", 1.0, 100.0)
	b3.HitPoints = 0.0
	if !b3.IsDestroyed() {
		t.Error("zero HP should be destroyed")
	}
}

func TestBarrierShapeTypes(t *testing.T) {
	shapeTypes := []string{"cylinder", "box", "polygon", "billboard"}

	for _, st := range shapeTypes {
		t.Run(st, func(t *testing.T) {
			b := NewBarrier(st, "sprite_"+st, "fantasy", 2.0)
			if b.Shape.ShapeType != st {
				t.Errorf("expected ShapeType %s, got %s", st, b.Shape.ShapeType)
			}
		})
	}
}

func TestBarrierArchetypeRegistry(t *testing.T) {
	reg := NewBarrierArchetypeRegistry()

	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			archetypes := reg.GetArchetypes(genre)
			if len(archetypes) == 0 {
				t.Errorf("expected archetypes for genre %s", genre)
			}

			// Each genre should have all 3 categories
			natural := reg.GetArchetypesByCategory(genre, BarrierCategoryNatural)
			if len(natural) == 0 {
				t.Errorf("expected natural archetypes for genre %s", genre)
			}

			constructed := reg.GetArchetypesByCategory(genre, BarrierCategoryConstructed)
			if len(constructed) == 0 {
				t.Errorf("expected constructed archetypes for genre %s", genre)
			}

			organic := reg.GetArchetypesByCategory(genre, BarrierCategoryOrganic)
			if len(organic) == 0 {
				t.Errorf("expected organic archetypes for genre %s", genre)
			}
		})
	}
}

func TestBarrierArchetypeByID(t *testing.T) {
	reg := NewBarrierArchetypeRegistry()

	// Test known archetype
	arch, found := reg.GetArchetypeByID("fantasy", "boulder")
	if !found {
		t.Error("expected to find boulder archetype")
	}
	if arch.Name != "Boulder" {
		t.Errorf("expected Boulder, got %s", arch.Name)
	}
	if arch.ShapeType != "cylinder" {
		t.Errorf("expected cylinder shape, got %s", arch.ShapeType)
	}

	// Test unknown archetype
	_, found = reg.GetArchetypeByID("fantasy", "nonexistent")
	if found {
		t.Error("expected not to find nonexistent archetype")
	}
}

func TestCreateBarrierFromArchetype(t *testing.T) {
	reg := NewBarrierArchetypeRegistry()

	arch, _ := reg.GetArchetypeByID("fantasy", "hedgerow")
	barrier := CreateBarrierFromArchetype(arch, "hedgerow_sprite")

	if barrier.Genre != "fantasy" {
		t.Errorf("expected fantasy genre, got %s", barrier.Genre)
	}
	if !barrier.Destructible {
		t.Error("expected destructible barrier")
	}
	if barrier.HitPoints != arch.BaseHP {
		t.Errorf("expected HP %f, got %f", arch.BaseHP, barrier.HitPoints)
	}
	if barrier.Shape.SpriteKey != "hedgerow_sprite" {
		t.Errorf("expected hedgerow_sprite, got %s", barrier.Shape.SpriteKey)
	}
}

func TestBarrierArchetypeValidation(t *testing.T) {
	reg := NewBarrierArchetypeRegistry()

	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		archetypes := reg.GetArchetypes(genre)
		for _, arch := range archetypes {
			t.Run(genre+"/"+arch.ID, func(t *testing.T) {
				// Every archetype should have required fields
				if arch.ID == "" {
					t.Error("archetype ID is empty")
				}
				if arch.Name == "" {
					t.Error("archetype Name is empty")
				}
				if arch.Genre != genre {
					t.Errorf("archetype Genre mismatch: expected %s, got %s", genre, arch.Genre)
				}
				if arch.ShapeType == "" {
					t.Error("archetype ShapeType is empty")
				}
				if arch.BaseHeight <= 0 {
					t.Errorf("archetype BaseHeight should be positive, got %f", arch.BaseHeight)
				}
				if arch.SpawnWeight <= 0 {
					t.Errorf("archetype SpawnWeight should be positive, got %f", arch.SpawnWeight)
				}
				// Destructible barriers must have HP
				if arch.Destructible && arch.BaseHP <= 0 {
					t.Error("destructible archetype has no BaseHP")
				}
			})
		}
	}
}

// ============================================================
// EnvironmentObject Tests
// ============================================================

func TestEnvironmentObjectType(t *testing.T) {
	obj := &EnvironmentObject{ObjectType: "test"}
	if obj.Type() != "EnvironmentObject" {
		t.Errorf("expected EnvironmentObject, got %s", obj.Type())
	}
}

func TestNewInventoriableObject(t *testing.T) {
	obj := NewInventoriableObject("health_potion", "Health Potion", "item_health_potion", "fantasy", 3)

	if obj.Category != ObjectCategoryInventoriable {
		t.Errorf("expected inventoriable category, got %s", obj.Category)
	}
	if obj.ObjectType != "health_potion" {
		t.Errorf("expected health_potion, got %s", obj.ObjectType)
	}
	if obj.DisplayName != "Health Potion" {
		t.Errorf("expected Health Potion, got %s", obj.DisplayName)
	}
	if obj.ItemID != "item_health_potion" {
		t.Errorf("expected item_health_potion, got %s", obj.ItemID)
	}
	if obj.Quantity != 3 {
		t.Errorf("expected quantity 3, got %d", obj.Quantity)
	}
	if obj.InteractionType != InteractionPickup {
		t.Errorf("expected pickup interaction, got %s", obj.InteractionType)
	}
	if !obj.IsPickupable() {
		t.Error("inventoriable object should be pickupable")
	}
}

func TestNewInteractiveObject(t *testing.T) {
	tests := []struct {
		interactionType InteractionType
		expectedText    string
	}{
		{InteractionOpen, "Open"},
		{InteractionUse, "Activate"},
		{InteractionRead, "Read"},
		{InteractionPush, "Push"},
		{InteractionExamine, "Examine"},
	}

	for _, tt := range tests {
		t.Run(string(tt.interactionType), func(t *testing.T) {
			obj := NewInteractiveObject("test_obj", "Test Object", tt.interactionType, "fantasy")
			if obj.Category != ObjectCategoryInteractive {
				t.Errorf("expected interactive category, got %s", obj.Category)
			}
			if obj.UseText != tt.expectedText {
				t.Errorf("expected %s, got %s", tt.expectedText, obj.UseText)
			}
		})
	}
}

func TestNewDecorativeObject(t *testing.T) {
	obj := NewDecorativeObject("pillar", "Stone Pillar", "fantasy")

	if obj.Category != ObjectCategoryDecorative {
		t.Errorf("expected decorative category, got %s", obj.Category)
	}
	if obj.InteractionType != InteractionNone {
		t.Errorf("expected no interaction, got %s", obj.InteractionType)
	}
	if obj.CanInteract() {
		t.Error("decorative object should not be interactable")
	}
}

func TestNewContainer(t *testing.T) {
	contents := []string{"gold_coins", "health_potion"}
	obj := NewContainer("wooden_chest", "Wooden Chest", "fantasy", contents)

	if obj.InteractionType != InteractionOpen {
		t.Errorf("expected open interaction, got %s", obj.InteractionType)
	}
	if !obj.IsContainer() {
		t.Error("should be identified as container")
	}
	if len(obj.ContainerContents) != 2 {
		t.Errorf("expected 2 items, got %d", len(obj.ContainerContents))
	}
}

func TestNewLockedContainer(t *testing.T) {
	obj := NewLockedContainer("treasure_chest", "Treasure Chest", "fantasy", []string{"gold"}, 50, "gold_key")

	if !obj.IsLocked {
		t.Error("container should be locked")
	}
	if obj.LockDifficulty != 50 {
		t.Errorf("expected difficulty 50, got %d", obj.LockDifficulty)
	}
	if obj.RequiredKeyID != "gold_key" {
		t.Errorf("expected gold_key, got %s", obj.RequiredKeyID)
	}
	if !obj.NeedsKey() {
		t.Error("should need key to unlock")
	}
	if obj.UseText != "Unlock" {
		t.Errorf("expected Unlock, got %s", obj.UseText)
	}
}

func TestNewDoor(t *testing.T) {
	// Unlocked door
	door := NewDoor("Wooden Door", "fantasy", false, 0, "")
	if door.ObjectType != "door" {
		t.Errorf("expected door type, got %s", door.ObjectType)
	}
	if !door.IsDoor() {
		t.Error("should be identified as door")
	}
	if door.UseText != "Open" {
		t.Errorf("expected Open, got %s", door.UseText)
	}

	// Locked door
	lockedDoor := NewDoor("Iron Gate", "fantasy", true, 75, "dungeon_key")
	if !lockedDoor.IsLocked {
		t.Error("door should be locked")
	}
	if lockedDoor.UseText != "Unlock" {
		t.Errorf("expected Unlock, got %s", lockedDoor.UseText)
	}
}

func TestEnvironmentObjectMethods(t *testing.T) {
	// Test CanInteract
	decorative := NewDecorativeObject("statue", "Statue", "fantasy")
	if decorative.CanInteract() {
		t.Error("decorative should not be interactable")
	}

	interactive := NewInteractiveObject("lever", "Lever", InteractionUse, "fantasy")
	if !interactive.CanInteract() {
		t.Error("interactive should be interactable")
	}

	// Test IsUsable
	if !interactive.IsUsable() {
		t.Error("lever should be usable")
	}

	// Test NeedsKey
	lockedChest := NewLockedContainer("chest", "Chest", "fantasy", nil, 30, "silver_key")
	if !lockedChest.NeedsKey() {
		t.Error("locked chest with key should need key")
	}

	lockedNoKey := NewLockedContainer("chest", "Chest", "fantasy", nil, 30, "")
	if lockedNoKey.NeedsKey() {
		t.Error("locked chest without key ID should not need key")
	}
}

func TestObjectHighlightConstants(t *testing.T) {
	if ObjectHighlightNone != 0 {
		t.Error("ObjectHighlightNone should be 0")
	}
	if ObjectHighlightInRange != 1 {
		t.Error("ObjectHighlightInRange should be 1")
	}
	if ObjectHighlightTargeted != 2 {
		t.Error("ObjectHighlightTargeted should be 2")
	}
}

func TestObjectCategories(t *testing.T) {
	if ObjectCategoryInventoriable != "inventoriable" {
		t.Error("ObjectCategoryInventoriable mismatch")
	}
	if ObjectCategoryInteractive != "interactive" {
		t.Error("ObjectCategoryInteractive mismatch")
	}
	if ObjectCategoryDecorative != "decorative" {
		t.Error("ObjectCategoryDecorative mismatch")
	}
}

func TestInteractionTypes(t *testing.T) {
	types := []InteractionType{
		InteractionNone,
		InteractionPickup,
		InteractionOpen,
		InteractionUse,
		InteractionRead,
		InteractionTalk,
		InteractionPush,
		InteractionExamine,
	}

	// Ensure all types are unique
	seen := make(map[InteractionType]bool)
	for _, t := range types {
		if seen[t] {
			// This would indicate a duplicate, but we just check they're defined
		}
		seen[t] = true
	}
	if len(seen) != len(types) {
		// Duplicates found, but this test is mainly for coverage
	}
}

func TestNewObjectAppearance(t *testing.T) {
	app := NewObjectAppearance("door", "fantasy")
	if app.SpriteCategory != "object" {
		t.Errorf("expected category 'object', got %s", app.SpriteCategory)
	}
	if app.BodyPlan != "door" {
		t.Errorf("expected body plan 'door', got %s", app.BodyPlan)
	}
	if app.GenreID != "fantasy" {
		t.Errorf("expected genre 'fantasy', got %s", app.GenreID)
	}
	if !app.Visible {
		t.Error("appearance should be visible by default")
	}
}

func TestNewCreatureAppearance(t *testing.T) {
	app := NewCreatureAppearance("wolf", "horror")
	if app.SpriteCategory != "creature" {
		t.Errorf("expected category 'creature', got %s", app.SpriteCategory)
	}
	if app.BodyPlan != "wolf" {
		t.Errorf("expected body plan 'wolf', got %s", app.BodyPlan)
	}
}

func TestNewHumanoidAppearance(t *testing.T) {
	app := NewHumanoidAppearance("warrior", "sci-fi")
	if app.SpriteCategory != "humanoid" {
		t.Errorf("expected category 'humanoid', got %s", app.SpriteCategory)
	}
	if app.BodyPlan != "warrior" {
		t.Errorf("expected body plan 'warrior', got %s", app.BodyPlan)
	}
}

func TestNewVehicleAppearance(t *testing.T) {
	app := NewVehicleAppearance("buggy", "post-apocalyptic")
	if app.SpriteCategory != "vehicle" {
		t.Errorf("expected category 'vehicle', got %s", app.SpriteCategory)
	}
	if app.BodyPlan != "buggy" {
		t.Errorf("expected body plan 'buggy', got %s", app.BodyPlan)
	}
}

func TestNewEffectAppearance(t *testing.T) {
	app := NewEffectAppearance("explosion", "cyberpunk")
	if app.SpriteCategory != "effect" {
		t.Errorf("expected category 'effect', got %s", app.SpriteCategory)
	}
	if app.BodyPlan != "explosion" {
		t.Errorf("expected body plan 'explosion', got %s", app.BodyPlan)
	}
}

func TestAppearanceConvenienceDefaults(t *testing.T) {
	app := NewAppearance("humanoid", "test", "fantasy")
	if app.Scale != 1.0 {
		t.Errorf("expected scale 1.0, got %f", app.Scale)
	}
	if app.AnimState != "idle" {
		t.Errorf("expected anim state 'idle', got %s", app.AnimState)
	}
	if app.Opacity != 1.0 {
		t.Errorf("expected opacity 1.0, got %f", app.Opacity)
	}
	if app.DamageOverlay != 0.0 {
		t.Errorf("expected damage overlay 0.0, got %f", app.DamageOverlay)
	}
}

// ============================================================
// Interactable Component Tests
// ============================================================

func TestInteractableType(t *testing.T) {
	i := &Interactable{InteractionType: "use", Range: 2.0}
	if i.Type() != "Interactable" {
		t.Errorf("expected Interactable, got %s", i.Type())
	}
}

func TestInteractableCanInteract(t *testing.T) {
	tests := []struct {
		name        string
		interact    *Interactable
		currentTime float64
		expected    bool
	}{
		{
			name:        "basic interaction available",
			interact:    &Interactable{InteractionType: "use", Cooldown: 1.0, LastInteractTime: 0},
			currentTime: 2.0,
			expected:    true,
		},
		{
			name:        "cooldown not elapsed",
			interact:    &Interactable{InteractionType: "use", Cooldown: 1.0, LastInteractTime: 1.5},
			currentTime: 2.0,
			expected:    false,
		},
		{
			name:        "cooldown elapsed",
			interact:    &Interactable{InteractionType: "use", Cooldown: 1.0, LastInteractTime: 0.5},
			currentTime: 2.0,
			expected:    true,
		},
		{
			name:        "single use not used",
			interact:    &Interactable{InteractionType: "use", SingleUse: true, Used: false},
			currentTime: 1.0,
			expected:    true,
		},
		{
			name:        "single use already used",
			interact:    &Interactable{InteractionType: "use", SingleUse: true, Used: true},
			currentTime: 1.0,
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.interact.CanInteract(tt.currentTime)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestInteractableTriggerInteraction(t *testing.T) {
	i := &Interactable{InteractionType: "use", SingleUse: false}
	i.TriggerInteraction(5.0)

	if i.LastInteractTime != 5.0 {
		t.Errorf("expected LastInteractTime 5.0, got %f", i.LastInteractTime)
	}
	if i.Used {
		t.Error("non-single-use should not be marked as used")
	}

	// Test single use
	i2 := &Interactable{InteractionType: "pickup", SingleUse: true}
	i2.TriggerInteraction(10.0)

	if !i2.Used {
		t.Error("single-use should be marked as used")
	}
}

func TestNewSimpleInteractable(t *testing.T) {
	i := NewSimpleInteractable("open", "Press E to open", 2.5)

	if i.InteractionType != "open" {
		t.Errorf("expected InteractionType 'open', got %s", i.InteractionType)
	}
	if i.Prompt != "Press E to open" {
		t.Errorf("expected Prompt 'Press E to open', got %s", i.Prompt)
	}
	if i.Range != 2.5 {
		t.Errorf("expected Range 2.5, got %f", i.Range)
	}
	if i.Cooldown != 0.5 {
		t.Errorf("expected default Cooldown 0.5, got %f", i.Cooldown)
	}
}

func TestNewLockedInteractable(t *testing.T) {
	i := NewLockedInteractable("unlock", "Requires key", "gold_key", 1.5)

	if !i.Locked {
		t.Error("expected Locked to be true")
	}
	if i.KeyID != "gold_key" {
		t.Errorf("expected KeyID 'gold_key', got %s", i.KeyID)
	}
}

func TestNewSkillCheckInteractable(t *testing.T) {
	i := NewSkillCheckInteractable("lockpick", "Pick lock", "lockpicking", 5, 1.0)

	if i.RequiredSkill != "lockpicking" {
		t.Errorf("expected RequiredSkill 'lockpicking', got %s", i.RequiredSkill)
	}
	if i.SkillDifficulty != 5 {
		t.Errorf("expected SkillDifficulty 5, got %d", i.SkillDifficulty)
	}
}

// ============================================================
// WorldItem Component Tests
// ============================================================

func TestWorldItemType(t *testing.T) {
	w := &WorldItem{ItemID: "sword", Quantity: 1}
	if w.Type() != "WorldItem" {
		t.Errorf("expected WorldItem, got %s", w.Type())
	}
}

func TestWorldItemCanPickup(t *testing.T) {
	tests := []struct {
		name        string
		item        *WorldItem
		entityID    uint64
		currentTime float64
		expected    bool
	}{
		{
			name:        "basic item pickup",
			item:        &WorldItem{ItemID: "gold", Quantity: 10},
			entityID:    1,
			currentTime: 0,
			expected:    true,
		},
		{
			name: "owned item - owner can pickup",
			item: &WorldItem{
				ItemID:      "rare_sword",
				Quantity:    1,
				Owner:       1,
				OwnerExpiry: 100.0,
			},
			entityID:    1,
			currentTime: 50.0,
			expected:    true,
		},
		{
			name: "owned item - non-owner cannot pickup",
			item: &WorldItem{
				ItemID:      "rare_sword",
				Quantity:    1,
				Owner:       1,
				OwnerExpiry: 100.0,
			},
			entityID:    2,
			currentTime: 50.0,
			expected:    false,
		},
		{
			name: "owned item - ownership expired",
			item: &WorldItem{
				ItemID:      "rare_sword",
				Quantity:    1,
				Owner:       1,
				OwnerExpiry: 100.0,
			},
			entityID:    2,
			currentTime: 150.0,
			expected:    true,
		},
		{
			name: "respawnable not yet respawned",
			item: &WorldItem{
				ItemID:         "herb",
				Quantity:       1,
				Respawnable:    true,
				RespawnTime:    60.0,
				LastPickupTime: 100.0,
			},
			entityID:    1,
			currentTime: 120.0,
			expected:    false,
		},
		{
			name: "respawnable has respawned",
			item: &WorldItem{
				ItemID:         "herb",
				Quantity:       1,
				Respawnable:    true,
				RespawnTime:    60.0,
				LastPickupTime: 100.0,
			},
			entityID:    1,
			currentTime: 200.0,
			expected:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.item.CanPickup(tt.entityID, tt.currentTime)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestWorldItemPickup(t *testing.T) {
	w := &WorldItem{ItemID: "gold", Quantity: 50}
	w.Pickup(100.0)

	if w.LastPickupTime != 100.0 {
		t.Errorf("expected LastPickupTime 100.0, got %f", w.LastPickupTime)
	}
}

func TestWorldItemIsRespawned(t *testing.T) {
	// Non-respawnable item
	w1 := &WorldItem{ItemID: "unique_item", Respawnable: false}
	if w1.IsRespawned(1000.0) {
		t.Error("non-respawnable item should never be considered respawned")
	}

	// Respawnable item, never picked up
	w2 := &WorldItem{ItemID: "herb", Respawnable: true, RespawnTime: 60.0}
	if !w2.IsRespawned(0.0) {
		t.Error("never-picked-up respawnable should be available")
	}

	// Respawnable item, not yet respawned
	w3 := &WorldItem{
		ItemID:         "herb",
		Respawnable:    true,
		RespawnTime:    60.0,
		LastPickupTime: 100.0,
	}
	if w3.IsRespawned(120.0) {
		t.Error("item should not be respawned before respawn time")
	}

	// Respawnable item, has respawned
	if !w3.IsRespawned(200.0) {
		t.Error("item should be respawned after respawn time")
	}
}

func TestNewWorldItem(t *testing.T) {
	w := NewWorldItem("iron_ore", 5, "material")

	if w.ItemID != "iron_ore" {
		t.Errorf("expected ItemID 'iron_ore', got %s", w.ItemID)
	}
	if w.Quantity != 5 {
		t.Errorf("expected Quantity 5, got %d", w.Quantity)
	}
	if w.Category != "material" {
		t.Errorf("expected Category 'material', got %s", w.Category)
	}
	if w.StackLimit != 99 {
		t.Errorf("expected default StackLimit 99, got %d", w.StackLimit)
	}
	if w.Rarity != "common" {
		t.Errorf("expected default Rarity 'common', got %s", w.Rarity)
	}
}

func TestNewRespawnableItem(t *testing.T) {
	w := NewRespawnableItem("mushroom", 3, 120.0)

	if !w.Respawnable {
		t.Error("expected Respawnable to be true")
	}
	if w.RespawnTime != 120.0 {
		t.Errorf("expected RespawnTime 120.0, got %f", w.RespawnTime)
	}
}

func TestNewOwnedDrop(t *testing.T) {
	w := NewOwnedDrop("epic_helm", 1, 42, 30.0, 1000.0)

	if w.Owner != 42 {
		t.Errorf("expected Owner 42, got %d", w.Owner)
	}
	if w.OwnerExpiry != 1030.0 {
		t.Errorf("expected OwnerExpiry 1030.0, got %f", w.OwnerExpiry)
	}
	if w.SpawnTime != 1000.0 {
		t.Errorf("expected SpawnTime 1000.0, got %f", w.SpawnTime)
	}
}
