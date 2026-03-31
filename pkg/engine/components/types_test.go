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
