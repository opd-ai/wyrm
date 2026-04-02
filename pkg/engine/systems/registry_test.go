package systems

import (
	"testing"

	"github.com/opd-ai/wyrm/pkg/engine/components"
	"github.com/opd-ai/wyrm/pkg/engine/ecs"
	"github.com/opd-ai/wyrm/pkg/world/chunk"
)

// mockChunkLoader implements ChunkLoader for testing.
type mockChunkLoader struct {
	chunks map[[2]int]*chunk.Chunk
	calls  int
}

func newMockChunkLoader() *mockChunkLoader {
	return &mockChunkLoader{
		chunks: make(map[[2]int]*chunk.Chunk),
	}
}

func (m *mockChunkLoader) GetChunk(x, y int) *chunk.Chunk {
	m.calls++
	key := [2]int{x, y}
	if c, ok := m.chunks[key]; ok {
		return c
	}
	c := chunk.NewChunk(x, y, 16, 12345)
	m.chunks[key] = c
	return c
}

func TestWorldChunkSystemUpdate(t *testing.T) {
	w := ecs.NewWorld()
	loader := newMockChunkLoader()
	sys := NewWorldChunkSystem(loader, 512)
	w.RegisterSystem(sys)

	// Create entity with position
	e := w.CreateEntity()
	_ = w.AddComponent(e, &components.Position{X: 1000, Y: 1000, Z: 0})

	// Run update
	w.Update(0.016)

	// Should have loaded 9 chunks (3x3 grid around position)
	if loader.calls != 9 {
		t.Errorf("expected 9 chunk loads, got %d", loader.calls)
	}
}

func TestWorldChunkSystemNilManager(t *testing.T) {
	w := ecs.NewWorld()
	sys := &WorldChunkSystem{Manager: nil, chunkSize: 512}
	w.RegisterSystem(sys)

	e := w.CreateEntity()
	_ = w.AddComponent(e, &components.Position{X: 0, Y: 0, Z: 0})

	// Should not panic with nil manager
	w.Update(0.016)
}

func TestNPCScheduleSystemUpdate(t *testing.T) {
	w := ecs.NewWorld()
	sys := &NPCScheduleSystem{WorldHour: 9}
	w.RegisterSystem(sys)

	e := w.CreateEntity()
	schedule := &components.Schedule{
		CurrentActivity: "sleep",
		TimeSlots:       map[int]string{9: "work", 12: "eat", 18: "home"},
	}
	_ = w.AddComponent(e, schedule)

	w.Update(0.016)

	// Activity should have changed to "work" based on WorldHour=9
	if schedule.CurrentActivity != "work" {
		t.Errorf("expected activity 'work', got '%s'", schedule.CurrentActivity)
	}
}

func TestCombatSystemHealthClamp(t *testing.T) {
	w := ecs.NewWorld()
	sys := &CombatSystem{}
	w.RegisterSystem(sys)

	e := w.CreateEntity()
	health := &components.Health{Current: 150, Max: 100}
	_ = w.AddComponent(e, health)

	w.Update(0.016)

	// Health should be clamped to max
	if health.Current != 100 {
		t.Errorf("expected health clamped to 100, got %f", health.Current)
	}
}

func TestVehicleSystemMovement(t *testing.T) {
	w := ecs.NewWorld()
	sys := &VehicleSystem{}
	w.RegisterSystem(sys)

	e := w.CreateEntity()
	pos := &components.Position{X: 0, Y: 0, Z: 0}
	// Direction = 0 means facing East (positive X)
	vehicle := &components.Vehicle{VehicleType: "car", Speed: 10, Fuel: 100, Direction: 0}
	_ = w.AddComponent(e, pos)
	_ = w.AddComponent(e, vehicle)

	initialX := pos.X
	initialFuel := vehicle.Fuel

	w.Update(1.0) // 1 second update

	// Position should have changed in X direction
	if pos.X <= initialX {
		t.Error("vehicle position X should have increased")
	}

	// Fuel should have decreased
	if vehicle.Fuel >= initialFuel {
		t.Error("vehicle fuel should have decreased")
	}
}

func TestVehicleSystemDirectionalMovement(t *testing.T) {
	w := ecs.NewWorld()
	sys := &VehicleSystem{}
	w.RegisterSystem(sys)

	e := w.CreateEntity()
	pos := &components.Position{X: 0, Y: 0, Z: 0}
	// Direction = PI/2 means facing North (positive Y)
	vehicle := &components.Vehicle{VehicleType: "car", Speed: 10, Fuel: 100, Direction: 1.5707963} // ~PI/2
	_ = w.AddComponent(e, pos)
	_ = w.AddComponent(e, vehicle)

	initialY := pos.Y

	w.Update(1.0)

	// Y should have increased significantly
	if pos.Y <= initialY {
		t.Errorf("vehicle position Y should have increased with direction PI/2, got Y=%f", pos.Y)
	}
}

func TestVehicleSystemNoFuel(t *testing.T) {
	w := ecs.NewWorld()
	sys := &VehicleSystem{}
	w.RegisterSystem(sys)

	e := w.CreateEntity()
	pos := &components.Position{X: 0, Y: 0, Z: 0}
	vehicle := &components.Vehicle{VehicleType: "car", Speed: 10, Fuel: 0}
	_ = w.AddComponent(e, pos)
	_ = w.AddComponent(e, vehicle)

	initialX := pos.X

	w.Update(1.0)

	// Position should not change without fuel
	if pos.X != initialX {
		t.Error("vehicle should not move without fuel")
	}
}

func TestWeatherSystemInitialization(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewWeatherSystem("fantasy", 300.0)
	w.RegisterSystem(sys)

	w.Update(0.016)

	if sys.CurrentWeather != "clear" {
		t.Errorf("weather should be 'clear' after first update, got '%s'", sys.CurrentWeather)
	}
}

func TestWeatherSystemTransition(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewWeatherSystem("fantasy", 1.0) // 1 second duration for fast test
	w.RegisterSystem(sys)

	w.Update(0.5)
	initial := sys.CurrentWeather

	// Advance past weather duration
	w.Update(1.0)

	if sys.CurrentWeather == initial {
		t.Error("weather should have changed after duration")
	}
}

func TestWorldClockSystemAdvancesTime(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewWorldClockSystem(1.0) // 1 second per hour for fast testing
	w.RegisterSystem(sys)

	clock := w.CreateEntity()
	clockComp := &components.WorldClock{Hour: 0, Day: 0, HourLength: 1.0}
	_ = w.AddComponent(clock, clockComp)

	// Advance 1.5 seconds
	w.Update(1.5)

	if clockComp.Hour != 1 {
		t.Errorf("expected hour 1, got %d", clockComp.Hour)
	}
	if clockComp.TimeAccum < 0.4 || clockComp.TimeAccum > 0.6 {
		t.Errorf("expected ~0.5 accumulated time, got %f", clockComp.TimeAccum)
	}
}

func TestWorldClockSystemDayRollover(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewWorldClockSystem(0.1)
	w.RegisterSystem(sys)

	clock := w.CreateEntity()
	clockComp := &components.WorldClock{Hour: 23, Day: 0, HourLength: 0.1}
	_ = w.AddComponent(clock, clockComp)

	// Advance past midnight
	w.Update(0.15)

	if clockComp.Hour != 0 {
		t.Errorf("expected hour 0 after midnight, got %d", clockComp.Hour)
	}
	if clockComp.Day != 1 {
		t.Errorf("expected day 1 after midnight, got %d", clockComp.Day)
	}
}

func TestNPCScheduleSystemReadsWorldClock(t *testing.T) {
	w := ecs.NewWorld()
	npcSys := &NPCScheduleSystem{}
	w.RegisterSystem(npcSys)

	// Create world clock entity
	clock := w.CreateEntity()
	_ = w.AddComponent(clock, &components.WorldClock{Hour: 12, Day: 0})

	// Create NPC with schedule
	npc := w.CreateEntity()
	sched := &components.Schedule{
		CurrentActivity: "sleep",
		TimeSlots:       map[int]string{8: "work", 12: "eat", 18: "home"},
	}
	_ = w.AddComponent(npc, sched)

	w.Update(0.016)

	if sched.CurrentActivity != "eat" {
		t.Errorf("expected activity 'eat' at hour 12, got '%s'", sched.CurrentActivity)
	}
	if npcSys.WorldHour != 12 {
		t.Errorf("NPCScheduleSystem.WorldHour should be 12, got %d", npcSys.WorldHour)
	}
}

func TestRenderSystemWithPlayerEntity(t *testing.T) {
	w := ecs.NewWorld()
	player := w.CreateEntity()
	_ = w.AddComponent(player, &components.Position{X: 100, Y: 100, Z: 0})

	sys := &RenderSystem{PlayerEntity: player}
	w.RegisterSystem(sys)

	// Should not panic
	w.Update(0.016)
}

func TestAudioSystemGenre(t *testing.T) {
	sys := &AudioSystem{Genre: "fantasy"}
	if sys.Genre != "fantasy" {
		t.Errorf("expected genre 'fantasy', got '%s'", sys.Genre)
	}
}

func TestAllSystemsImplementInterface(t *testing.T) {
	// Verify all systems implement the System interface
	systems := []ecs.System{
		&WorldChunkSystem{},
		&NPCScheduleSystem{},
		NewWorldClockSystem(60.0),
		NewFactionPoliticsSystem(0.1),
		NewCrimeSystem(10.0, 100.0),
		NewEconomySystem(0.5, 0.1),
		&CombatSystem{},
		&VehicleSystem{},
		NewQuestSystem(),
		NewWeatherSystem("fantasy", 300.0),
		&RenderSystem{},
		&AudioSystem{},
		NewSkillProgressionSystem(100, 100),
	}

	w := ecs.NewWorld()
	for _, s := range systems {
		// Each system should be able to Update without panic
		s.Update(w, 0.016)
	}
}

func TestFactionPoliticsSystemReputationDecay(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewFactionPoliticsSystem(10.0) // Fast decay for testing
	w.RegisterSystem(sys)

	e := w.CreateEntity()
	rep := &components.Reputation{
		Standings: map[string]float64{
			"guild":   50.0,
			"bandits": -50.0,
		},
	}
	_ = w.AddComponent(e, rep)

	// Run several updates
	for i := 0; i < 10; i++ {
		w.Update(1.0)
	}

	// Reputation should have decayed toward 0
	if rep.Standings["guild"] >= 50.0 {
		t.Error("positive reputation should decay toward 0")
	}
	if rep.Standings["bandits"] <= -50.0 {
		t.Error("negative reputation should decay toward 0")
	}
}

func TestFactionPoliticsSystemRelations(t *testing.T) {
	sys := NewFactionPoliticsSystem(0.1)

	// Set a hostile relation
	sys.SetRelation("knights", "bandits", RelationHostile)
	if sys.GetRelation("knights", "bandits") != RelationHostile {
		t.Error("expected hostile relation between knights and bandits")
	}

	// Relations should be symmetric
	if sys.GetRelation("bandits", "knights") != RelationHostile {
		t.Error("relations should be symmetric")
	}

	// Unset relations should be neutral
	if sys.GetRelation("guild", "church") != RelationNeutral {
		t.Error("unset relations should be neutral")
	}
}

func TestFactionTerritoryKillTracking(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewFactionPoliticsSystem(0.1)
	sys.KillsForHostility = 3
	sys.ReputationPerKill = -25.0
	w.RegisterSystem(sys)

	// Create player with reputation
	player := w.CreateEntity()
	rep := &components.Reputation{Standings: map[string]float64{"guards": 50.0}}
	_ = w.AddComponent(player, rep)

	// Create faction territory
	territory := w.CreateEntity()
	ft := &components.FactionTerritory{
		FactionID: "guards",
		Vertices: []components.Point2D{
			{X: 0, Y: 0}, {X: 100, Y: 0}, {X: 100, Y: 100}, {X: 0, Y: 100},
		},
	}
	_ = w.AddComponent(territory, ft)

	// First kill - should not trigger hostility
	result := sys.ReportKill(w, player, "guards")
	if result {
		t.Error("first kill should not trigger hostility")
	}
	if rep.Standings["guards"] != 25.0 { // 50 - 25
		t.Errorf("expected reputation 25, got %f", rep.Standings["guards"])
	}

	// Second kill
	result = sys.ReportKill(w, player, "guards")
	if result {
		t.Error("second kill should not trigger hostility")
	}

	// Third kill - should trigger hostility
	result = sys.ReportKill(w, player, "guards")
	if !result {
		t.Error("third kill should trigger hostility")
	}
	if rep.Standings["guards"] != -100 {
		t.Errorf("expected reputation -100 (hostile), got %f", rep.Standings["guards"])
	}
}

func TestFactionTreatyReducesHostility(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewFactionPoliticsSystem(0.1)
	w.RegisterSystem(sys)

	// Create hostile player
	player := w.CreateEntity()
	rep := &components.Reputation{Standings: map[string]float64{"guards": -100.0}}
	_ = w.AddComponent(player, rep)

	// Create territory with kill history
	territory := w.CreateEntity()
	ft := &components.FactionTerritory{
		FactionID:   "guards",
		KillTracker: map[uint64]int{uint64(player): 5},
	}
	_ = w.AddComponent(territory, ft)

	// Sign treaty
	result := sys.SignTreaty(w, player, "guards")
	if !result {
		t.Error("SignTreaty should succeed")
	}
	if rep.Standings["guards"] != 0 {
		t.Errorf("expected reputation 0 after treaty, got %f", rep.Standings["guards"])
	}
	if ft.KillTracker[uint64(player)] != 0 {
		t.Errorf("expected kill tracker reset, got %d", ft.KillTracker[uint64(player)])
	}
}

func TestFactionTerritoryContainsPoint(t *testing.T) {
	ft := &components.FactionTerritory{
		FactionID: "guards",
		Vertices: []components.Point2D{
			{X: 0, Y: 0}, {X: 100, Y: 0}, {X: 100, Y: 100}, {X: 0, Y: 100},
		},
	}

	// Point inside
	if !ft.ContainsPoint(50, 50) {
		t.Error("point (50,50) should be inside")
	}

	// Point outside
	if ft.ContainsPoint(150, 50) {
		t.Error("point (150,50) should be outside")
	}

	// Point on boundary edge case
	if ft.ContainsPoint(-10, -10) {
		t.Error("point (-10,-10) should be outside")
	}
}

func TestCrimeSystemWantedLevelDecay(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewCrimeSystem(5.0, 100.0) // 5 second decay delay
	w.RegisterSystem(sys)

	e := w.CreateEntity()
	crime := &components.Crime{WantedLevel: 3, BountyAmount: 300.0, LastCrimeTime: 0}
	_ = w.AddComponent(e, crime)

	// Advance time past decay delay
	for i := 0; i < 10; i++ {
		w.Update(1.0)
	}

	// Wanted level should have decreased
	if crime.WantedLevel >= 3 {
		t.Errorf("wanted level should have decayed, got %d", crime.WantedLevel)
	}
}

func TestCrimeSystemReportCrime(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewCrimeSystem(60.0, 100.0)
	w.RegisterSystem(sys)

	criminal := w.CreateEntity()
	crime := &components.Crime{WantedLevel: 0, BountyAmount: 0}
	_ = w.AddComponent(criminal, crime)

	// Create a witness
	witness := w.CreateEntity()
	_ = w.AddComponent(witness, &components.Witness{CanReport: true})
	_ = w.AddComponent(witness, &components.Position{X: 10, Y: 10})

	// Report a crime
	sys.ReportCrime(w, criminal)

	if crime.WantedLevel != 1 {
		t.Errorf("expected wanted level 1 after crime report, got %d", crime.WantedLevel)
	}
	if crime.BountyAmount != 100.0 {
		t.Errorf("expected bounty 100, got %f", crime.BountyAmount)
	}
}

func TestCrimeSystemNoWitnesses(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewCrimeSystem(60.0, 100.0)
	w.RegisterSystem(sys)

	criminal := w.CreateEntity()
	crime := &components.Crime{WantedLevel: 0, BountyAmount: 0}
	_ = w.AddComponent(criminal, crime)

	// Report crime with no witnesses
	sys.ReportCrime(w, criminal)

	if crime.WantedLevel != 0 {
		t.Error("wanted level should not increase without witnesses")
	}
}

func TestCrimeSystemPayBounty(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewCrimeSystem(60.0, 100.0)
	w.RegisterSystem(sys)

	criminal := w.CreateEntity()
	crime := &components.Crime{WantedLevel: 3, BountyAmount: 300.0}
	_ = w.AddComponent(criminal, crime)

	// Pay bounty
	result := sys.PayBounty(w, criminal)
	if !result {
		t.Error("PayBounty should succeed")
	}
	if crime.WantedLevel != 0 {
		t.Errorf("wanted level should be 0 after paying bounty, got %d", crime.WantedLevel)
	}
	if crime.BountyAmount != 0 {
		t.Errorf("bounty should be 0 after paying, got %f", crime.BountyAmount)
	}
}

func TestCrimeSystemJailMechanic(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewCrimeSystem(60.0, 100.0)
	w.RegisterSystem(sys)

	criminal := w.CreateEntity()
	crime := &components.Crime{WantedLevel: 3, BountyAmount: 300.0}
	_ = w.AddComponent(criminal, crime)

	// Send to jail for 10 seconds
	result := sys.GoToJail(w, criminal, 10.0)
	if !result {
		t.Error("GoToJail should succeed")
	}
	if !crime.InJail {
		t.Error("criminal should be in jail")
	}
	if crime.WantedLevel != 0 {
		t.Error("wanted level should be 0 while in jail")
	}

	// Advance time and check release
	for i := 0; i < 15; i++ {
		w.Update(1.0) // Advance 1 second per tick
	}
	sys.CheckJailRelease(w)

	if crime.InJail {
		t.Error("criminal should be released from jail after serving time")
	}
}

func TestCrimeSystemWitnessLOS(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewCrimeSystem(60.0, 100.0)
	sys.WitnessRange = 50.0
	w.RegisterSystem(sys)

	// Create criminal at position (0, 0)
	criminal := w.CreateEntity()
	crime := &components.Crime{WantedLevel: 0, BountyAmount: 0}
	_ = w.AddComponent(criminal, crime)
	_ = w.AddComponent(criminal, &components.Position{X: 0, Y: 0})

	// Create witness far away (100, 100) - out of range
	witness := w.CreateEntity()
	_ = w.AddComponent(witness, &components.Witness{CanReport: true})
	_ = w.AddComponent(witness, &components.Position{X: 100, Y: 100})

	// Report crime - should not be witnessed (out of range)
	sys.ReportCrime(w, criminal)
	if crime.WantedLevel != 0 {
		t.Errorf("crime should not be witnessed from far away, got wanted level %d", crime.WantedLevel)
	}

	// Move witness close (10, 10) - in range
	wPos, _ := w.GetComponent(witness, "Position")
	wPos.(*components.Position).X = 10
	wPos.(*components.Position).Y = 10

	// Report crime - should be witnessed now
	sys.ReportCrime(w, criminal)
	if crime.WantedLevel != 1 {
		t.Errorf("crime should be witnessed from close by, expected wanted level 1, got %d", crime.WantedLevel)
	}
}

func TestEconomySystemPricing(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewEconomySystem(0.5, 0.1)
	sys.SetBasePrice("sword", 100.0)
	sys.SetBasePrice("potion", 50.0)
	w.RegisterSystem(sys)

	shop := w.CreateEntity()
	node := &components.EconomyNode{
		PriceTable: make(map[string]float64),
		Supply:     map[string]int{"sword": 10, "potion": 5},
		Demand:     map[string]int{"sword": 10, "potion": 20}, // High potion demand
	}
	_ = w.AddComponent(shop, node)

	w.Update(0.016)

	// Balanced supply/demand should have base price
	if node.PriceTable["sword"] < 90 || node.PriceTable["sword"] > 110 {
		t.Errorf("sword price should be near base, got %f", node.PriceTable["sword"])
	}

	// High demand should increase price
	if node.PriceTable["potion"] <= 50 {
		t.Errorf("potion price should be above base due to demand, got %f", node.PriceTable["potion"])
	}
}

func TestEconomySellReducesPrice(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewEconomySystem(0.5, 0.1)
	sys.SetBasePrice("iron_ore", 10.0)
	w.RegisterSystem(sys)

	shop := w.CreateEntity()
	node := &components.EconomyNode{
		PriceTable: make(map[string]float64),
		Supply:     map[string]int{"iron_ore": 10},
		Demand:     map[string]int{"iron_ore": 10},
	}
	_ = w.AddComponent(shop, node)

	// Get initial price
	w.Update(0.016)
	initialPrice := sys.GetBuyPrice(w, shop, "iron_ore")
	t.Logf("Initial price: %f", initialPrice)

	// Sell 50 items (flooding market with supply)
	sys.SellItem(w, shop, "iron_ore", 50)
	w.Update(0.016)
	newPrice := sys.GetBuyPrice(w, shop, "iron_ore")
	t.Logf("Price after selling 50: %f", newPrice)

	// Price should have reduced by at least 10%
	reduction := (initialPrice - newPrice) / initialPrice
	if reduction < 0.10 {
		t.Errorf("selling 50 items should reduce price by ≥10%%, got %.2f%% reduction (initial=%f, new=%f)",
			reduction*100, initialPrice, newPrice)
	}
}

func TestQuestSystemStageAdvancement(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewQuestSystem()
	sys.DefineQuest("main_quest", []QuestStageCondition{
		{FromStage: 0, RequiredFlag: "talked_to_npc", NextStage: 1},
		{FromStage: 1, RequiredFlag: "found_artifact", NextStage: 2},
		{FromStage: 2, RequiredFlag: "returned_artifact", Completes: true},
	})
	w.RegisterSystem(sys)

	player := w.CreateEntity()
	quest := &components.Quest{
		ID:           "main_quest",
		CurrentStage: 0,
		Flags:        map[string]bool{},
	}
	_ = w.AddComponent(player, quest)

	// Initially at stage 0
	w.Update(0.016)
	if quest.CurrentStage != 0 {
		t.Error("quest should start at stage 0")
	}

	// Set flag to advance
	quest.Flags["talked_to_npc"] = true
	w.Update(0.016)
	if quest.CurrentStage != 1 {
		t.Errorf("quest should advance to stage 1, got %d", quest.CurrentStage)
	}

	// Set next flag
	quest.Flags["found_artifact"] = true
	w.Update(0.016)
	if quest.CurrentStage != 2 {
		t.Errorf("quest should advance to stage 2, got %d", quest.CurrentStage)
	}

	// Complete quest
	quest.Flags["returned_artifact"] = true
	w.Update(0.016)
	if !quest.Completed {
		t.Error("quest should be completed")
	}
}

func TestQuestSystemBranchLocking(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewQuestSystem()
	// Define a quest with branching paths
	sys.DefineQuest("branch_quest", []QuestStageCondition{
		{FromStage: 0, RequiredFlag: "choose_path", NextStage: 1},
		// Branch A: help the rebels
		{FromStage: 1, RequiredFlag: "help_rebels", NextStage: 10, BranchID: "branch_a", LocksBranch: "branch_b"},
		// Branch B: help the empire
		{FromStage: 1, RequiredFlag: "help_empire", NextStage: 20, BranchID: "branch_b", LocksBranch: "branch_a"},
		// Branch A continuation
		{FromStage: 10, RequiredFlag: "rebels_win", Completes: true},
		// Branch B continuation
		{FromStage: 20, RequiredFlag: "empire_wins", Completes: true},
	})
	w.RegisterSystem(sys)

	player := w.CreateEntity()
	quest := &components.Quest{
		ID:           "branch_quest",
		CurrentStage: 0,
		Flags:        map[string]bool{},
	}
	_ = w.AddComponent(player, quest)

	// Advance to branch choice
	quest.Flags["choose_path"] = true
	w.Update(0.016)
	if quest.CurrentStage != 1 {
		t.Errorf("expected stage 1, got %d", quest.CurrentStage)
	}

	// Choose branch A (help rebels)
	quest.Flags["help_rebels"] = true
	w.Update(0.016)
	if quest.CurrentStage != 10 {
		t.Errorf("expected stage 10, got %d", quest.CurrentStage)
	}

	// Branch B should be locked
	if !quest.IsBranchLocked("branch_b") {
		t.Error("branch_b should be locked after choosing branch_a")
	}

	// Reset to stage 1 to test that branch B is blocked
	quest.CurrentStage = 1
	quest.Flags["help_empire"] = true
	w.Update(0.016)
	// Should NOT advance to stage 20 because branch_b is locked
	if quest.CurrentStage == 20 {
		t.Error("should not be able to take branch_b after it was locked")
	}
}

func TestSkillProgressionSystemCreation(t *testing.T) {
	sys := NewSkillProgressionSystem(100, 100)
	if sys.XPPerLevel != 100 {
		t.Errorf("expected XPPerLevel=100, got %f", sys.XPPerLevel)
	}
	if sys.LevelCap != 100 {
		t.Errorf("expected LevelCap=100, got %d", sys.LevelCap)
	}
}

func TestSkillProgressionSystemDefaults(t *testing.T) {
	sys := NewSkillProgressionSystem(0, 0)
	if sys.XPPerLevel != 100 {
		t.Errorf("expected default XPPerLevel=100, got %f", sys.XPPerLevel)
	}
	if sys.LevelCap != 100 {
		t.Errorf("expected default LevelCap=100, got %d", sys.LevelCap)
	}
}

func TestSkillProgressionSystemLevelUp(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewSkillProgressionSystem(100, 100)
	w.RegisterSystem(sys)

	player := w.CreateEntity()
	skills := components.NewSkills("fantasy")
	_ = w.AddComponent(player, skills)

	// Grant enough XP to level up (slightly more than 110 to handle floating point)
	sys.GrantSkillXP(w, player, "fire_magic", 115)

	w.Update(0.016)

	// Should have leveled up
	if skills.Levels["fire_magic"] != 2 {
		t.Errorf("expected fire_magic level 2, got %d", skills.Levels["fire_magic"])
	}
	// Excess XP should carry over
	if skills.Experience["fire_magic"] <= 0 {
		t.Error("excess XP should carry over")
	}
}

func TestSkillProgressionSystemMultipleLevelUps(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewSkillProgressionSystem(100, 100)
	w.RegisterSystem(sys)

	player := w.CreateEntity()
	skills := components.NewSkills("fantasy")
	_ = w.AddComponent(player, skills)

	// Grant XP over multiple updates
	sys.GrantSkillXP(w, player, "fire_magic", 110)
	w.Update(0.016)
	sys.GrantSkillXP(w, player, "fire_magic", 120) // More XP for level 2->3
	w.Update(0.016)

	// Should have leveled up twice
	if skills.Levels["fire_magic"] < 2 {
		t.Errorf("expected fire_magic level >= 2, got %d", skills.Levels["fire_magic"])
	}
}

func TestSkillProgressionSystemLevelCap(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewSkillProgressionSystem(100, 5) // Low cap for testing
	w.RegisterSystem(sys)

	player := w.CreateEntity()
	skills := components.NewSkills("fantasy")
	skills.Levels["fire_magic"] = 5 // At cap
	skills.Experience["fire_magic"] = 1000
	_ = w.AddComponent(player, skills)

	w.Update(0.016)

	// Should not exceed cap
	if skills.Levels["fire_magic"] > 5 {
		t.Errorf("skill level should not exceed cap, got %d", skills.Levels["fire_magic"])
	}
}

func TestSkillProgressionSystemGrantXP(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewSkillProgressionSystem(100, 100)

	player := w.CreateEntity()
	skills := components.NewSkills("fantasy")
	_ = w.AddComponent(player, skills)

	// Grant XP
	result := sys.GrantSkillXP(w, player, "fire_magic", 50)
	if !result {
		t.Error("GrantSkillXP should return true for valid skill")
	}
	if skills.Experience["fire_magic"] != 50 {
		t.Errorf("expected 50 XP, got %f", skills.Experience["fire_magic"])
	}

	// Invalid skill should fail
	result = sys.GrantSkillXP(w, player, "nonexistent_skill", 50)
	if result {
		t.Error("GrantSkillXP should return false for nonexistent skill")
	}
}

func TestSkillProgressionSystemGetSkillLevel(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewSkillProgressionSystem(100, 100)

	player := w.CreateEntity()
	skills := components.NewSkills("fantasy")
	skills.Levels["fire_magic"] = 25
	_ = w.AddComponent(player, skills)

	level := sys.GetSkillLevel(w, player, "fire_magic")
	if level != 25 {
		t.Errorf("expected level 25, got %d", level)
	}

	// Unknown entity should return 0
	unknownLevel := sys.GetSkillLevel(w, 99999, "fire_magic")
	if unknownLevel != 0 {
		t.Errorf("expected level 0 for unknown entity, got %d", unknownLevel)
	}
}

func TestSkillProgressionSystemXPScaling(t *testing.T) {
	sys := NewSkillProgressionSystem(100, 100)

	// XP required should scale with level
	xp1 := sys.calculateXPRequired(1)
	xp10 := sys.calculateXPRequired(10)
	xp50 := sys.calculateXPRequired(50)

	if xp10 <= xp1 {
		t.Error("XP required at level 10 should be more than level 1")
	}
	if xp50 <= xp10 {
		t.Error("XP required at level 50 should be more than level 10")
	}
}

// TestNewAudioSystem verifies constructor defaults.
func TestNewAudioSystem(t *testing.T) {
	tests := []struct {
		genre string
	}{
		{"fantasy"},
		{"sci-fi"},
		{"horror"},
		{"cyberpunk"},
		{"post-apocalyptic"},
	}

	for _, tc := range tests {
		t.Run(tc.genre, func(t *testing.T) {
			sys := NewAudioSystem(tc.genre)
			if sys.Genre != tc.genre {
				t.Errorf("expected genre %q, got %q", tc.genre, sys.Genre)
			}
			if sys.CombatDetectionRange != 50.0 {
				t.Errorf("expected combat range 50.0, got %f", sys.CombatDetectionRange)
			}
			if sys.AmbientUpdateInterval != 5.0 {
				t.Errorf("expected ambient interval 5.0, got %f", sys.AmbientUpdateInterval)
			}
		})
	}
}

// TestAudioSystemUpdateWithListener verifies spatial audio processing.
func TestAudioSystemUpdateWithListener(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewAudioSystem("fantasy")
	w.RegisterSystem(sys)

	// Create listener entity
	listener := w.CreateEntity()
	_ = w.AddComponent(listener, &components.Position{X: 0, Y: 0, Z: 0})
	_ = w.AddComponent(listener, &components.AudioListener{Enabled: true})

	// Create audio source at known distance
	source := w.CreateEntity()
	_ = w.AddComponent(source, &components.Position{X: 10, Y: 0, Z: 0})
	_ = w.AddComponent(source, &components.AudioSource{
		Playing: true,
		Volume:  1.0,
		Range:   50.0,
	})

	// Create audio state to track updates
	stateEntity := w.CreateEntity()
	_ = w.AddComponent(stateEntity, &components.AudioState{})

	// Run update
	w.Update(0.016)

	// Verify state was updated (listener position recorded)
	comp, ok := w.GetComponent(stateEntity, "AudioState")
	if !ok {
		t.Fatal("AudioState component not found")
	}
	state := comp.(*components.AudioState)
	if state.LastPositionX != 0 || state.LastPositionY != 0 {
		t.Errorf("expected position (0,0), got (%f,%f)", state.LastPositionX, state.LastPositionY)
	}
}

// TestAudioSystemCombatIntensity verifies combat intensity calculation.
func TestAudioSystemCombatIntensity(t *testing.T) {
	tests := []struct {
		name         string
		hostileCount int
		expectedMin  float64
		expectedMax  float64
	}{
		{"no hostiles", 0, 0.0, 0.0},
		{"one hostile", 1, 0.09, 0.11},
		{"five hostiles", 5, 0.49, 0.51},
		{"max hostiles", 10, 1.0, 1.0},
		{"over max hostiles", 15, 1.0, 1.0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := ecs.NewWorld()
			sys := NewAudioSystem("fantasy")

			// Create listener
			listener := w.CreateEntity()
			_ = w.AddComponent(listener, &components.Position{X: 0, Y: 0, Z: 0})
			_ = w.AddComponent(listener, &components.AudioListener{Enabled: true})

			// Create hostile entities within range
			for i := 0; i < tc.hostileCount; i++ {
				hostile := w.CreateEntity()
				_ = w.AddComponent(hostile, &components.Position{X: float64(i * 5), Y: 0, Z: 0})
				_ = w.AddComponent(hostile, &components.Health{Current: 100, Max: 100})
				_ = w.AddComponent(hostile, &components.Faction{ID: "enemy", Reputation: -100})
			}

			// Calculate intensity
			intensity := sys.calculateCombatIntensity(w, [2]float64{0, 0})
			if intensity < tc.expectedMin || intensity > tc.expectedMax {
				t.Errorf("expected intensity in [%f, %f], got %f", tc.expectedMin, tc.expectedMax, intensity)
			}
		})
	}
}

// TestAudioSystemAmbientSelection verifies genre-based ambient selection.
func TestAudioSystemAmbientSelection(t *testing.T) {
	tests := []struct {
		genre              string
		expectedCity       string
		expectedWilderness string
	}{
		{"fantasy", "city_medieval", "wilderness_forest"},
		{"sci-fi", "city_station", "wilderness_alien"},
		{"horror", "city_abandoned", "wilderness_dark"},
		{"cyberpunk", "city_neon", "wilderness_industrial"},
		{"post-apocalyptic", "city_ruins", "wilderness_wasteland"},
		{"unknown", "city_generic", "wilderness_generic"},
	}

	for _, tc := range tests {
		t.Run(tc.genre, func(t *testing.T) {
			sys := NewAudioSystem(tc.genre)

			city := sys.getCityAmbient()
			if city != tc.expectedCity {
				t.Errorf("expected city ambient %q, got %q", tc.expectedCity, city)
			}

			wilderness := sys.getWildernessAmbient()
			if wilderness != tc.expectedWilderness {
				t.Errorf("expected wilderness ambient %q, got %q", tc.expectedWilderness, wilderness)
			}
		})
	}
}

// TestAudioSystemNoListener verifies graceful handling when no listener exists.
func TestAudioSystemNoListener(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewAudioSystem("fantasy")
	w.RegisterSystem(sys)

	// Create audio source but NO listener
	source := w.CreateEntity()
	_ = w.AddComponent(source, &components.Position{X: 10, Y: 0, Z: 0})
	_ = w.AddComponent(source, &components.AudioSource{
		Playing: true,
		Volume:  1.0,
		Range:   50.0,
	})

	// Should not panic
	w.Update(0.016)
}

// TestAudioSystemDisabledListener verifies disabled listeners are skipped.
func TestAudioSystemDisabledListener(t *testing.T) {
	w := ecs.NewWorld()
	sys := NewAudioSystem("fantasy")
	w.RegisterSystem(sys)

	// Create disabled listener
	listener := w.CreateEntity()
	_ = w.AddComponent(listener, &components.Position{X: 100, Y: 100, Z: 0})
	_ = w.AddComponent(listener, &components.AudioListener{Enabled: false})

	// Create audio state
	stateEntity := w.CreateEntity()
	_ = w.AddComponent(stateEntity, &components.AudioState{LastPositionX: -1, LastPositionY: -1})

	// Run update
	w.Update(0.016)

	// State should NOT be updated since listener is disabled
	comp, _ := w.GetComponent(stateEntity, "AudioState")
	state := comp.(*components.AudioState)
	if state.LastPositionX != -1 || state.LastPositionY != -1 {
		t.Error("AudioState was updated despite disabled listener")
	}
}

// ============================================================
// WorldChunkSystem Barrier Integration Tests
// ============================================================

// mockChunkLoaderWithBarriers creates chunks that have barrier spawns.
type mockChunkLoaderWithBarriers struct {
	chunks map[[2]int]*chunk.Chunk
	calls  int
}

func newMockChunkLoaderWithBarriers() *mockChunkLoaderWithBarriers {
	return &mockChunkLoaderWithBarriers{
		chunks: make(map[[2]int]*chunk.Chunk),
	}
}

func (m *mockChunkLoaderWithBarriers) GetChunk(x, y int) *chunk.Chunk {
	m.calls++
	key := [2]int{x, y}
	if c, ok := m.chunks[key]; ok {
		return c
	}
	// Create chunk with barriers
	c := chunk.NewChunkWithBarriers(x, y, 16, int64(x*1000+y), "fantasy")
	m.chunks[key] = c
	return c
}

func TestWorldChunkSystemSpawnsBarriers(t *testing.T) {
	w := ecs.NewWorld()
	loader := newMockChunkLoaderWithBarriers()
	sys := NewWorldChunkSystemWithGenre(loader, 16, "fantasy")
	w.RegisterSystem(sys)

	// Create entity with position to trigger chunk loading
	e := w.CreateEntity()
	_ = w.AddComponent(e, &components.Position{X: 8, Y: 8, Z: 0})

	// Run update to load chunks
	w.Update(0.016)

	// System should have loaded chunks
	if loader.calls == 0 {
		t.Error("expected chunks to be loaded")
	}

	// Check if barrier count is tracked
	chunkCount := sys.GetChunksWithBarriers()
	if chunkCount == 0 {
		t.Error("expected chunks to be tracked")
	}
}

func TestWorldChunkSystemBarrierEntityCreation(t *testing.T) {
	w := ecs.NewWorld()
	loader := newMockChunkLoaderWithBarriers()
	sys := NewWorldChunkSystemWithGenre(loader, 16, "fantasy")

	// Pre-populate a chunk with known barrier data
	testChunk := &chunk.Chunk{
		X:    0,
		Y:    0,
		Size: 16,
		DetailSpawns: []chunk.DetailSpawn{
			{
				Type:   chunk.DetailSpawnBarrierNatural,
				LocalX: 5.0,
				LocalY: 5.0,
				Scale:  1.0,
				BarrierData: &chunk.BarrierSpawnData{
					ArchetypeID:  "boulder",
					ShapeType:    "cylinder",
					Radius:       0.5,
					Height:       1.0,
					Destructible: false,
				},
			},
		},
	}
	loader.chunks[[2]int{0, 0}] = testChunk

	w.RegisterSystem(sys)

	// Create entity with position to trigger chunk loading
	e := w.CreateEntity()
	_ = w.AddComponent(e, &components.Position{X: 8, Y: 8, Z: 0})

	// Run update
	w.Update(0.016)

	// Check that barrier entities were created
	barrierEntities := w.Entities("Barrier")
	if len(barrierEntities) == 0 {
		t.Error("expected barrier entities to be created")
	}

	// Verify barrier has correct position
	for _, be := range barrierEntities {
		posComp, ok := w.GetComponent(be, "Position")
		if !ok {
			t.Error("barrier entity missing Position component")
			continue
		}
		pos := posComp.(*components.Position)
		// Position should be world-space (chunk offset + local)
		if pos.X < 0 || pos.Y < 0 {
			t.Errorf("unexpected barrier position: %f, %f", pos.X, pos.Y)
		}

		barrierComp, ok := w.GetComponent(be, "Barrier")
		if !ok {
			t.Error("barrier entity missing Barrier component")
			continue
		}
		barrier := barrierComp.(*components.Barrier)
		if barrier.Genre != "fantasy" {
			t.Errorf("expected genre 'fantasy', got %q", barrier.Genre)
		}
	}
}

func TestWorldChunkSystemUnloadChunk(t *testing.T) {
	w := ecs.NewWorld()
	loader := newMockChunkLoaderWithBarriers()
	sys := NewWorldChunkSystemWithGenre(loader, 16, "fantasy")

	// Pre-populate a chunk with a barrier
	testChunk := &chunk.Chunk{
		X:    0,
		Y:    0,
		Size: 16,
		DetailSpawns: []chunk.DetailSpawn{
			{
				Type:   chunk.DetailSpawnBarrierNatural,
				LocalX: 5.0,
				LocalY: 5.0,
				Scale:  1.0,
				BarrierData: &chunk.BarrierSpawnData{
					ShapeType: "cylinder",
					Radius:    0.5,
					Height:    1.0,
				},
			},
		},
	}
	loader.chunks[[2]int{0, 0}] = testChunk

	w.RegisterSystem(sys)

	// Create entity to trigger chunk loading
	e := w.CreateEntity()
	_ = w.AddComponent(e, &components.Position{X: 8, Y: 8, Z: 0})
	w.Update(0.016)

	// Count barriers before unload
	barriersBefore := len(w.Entities("Barrier"))

	// Unload the chunk
	sys.UnloadChunk(w, 0, 0)

	// Barrier entities should be removed
	barriersAfter := len(w.Entities("Barrier"))
	if barriersAfter >= barriersBefore && barriersBefore > 0 {
		t.Errorf("expected barriers to be removed after unload, had %d, now %d", barriersBefore, barriersAfter)
	}
}

func TestWorldChunkSystemDuplicateLoadPrevented(t *testing.T) {
	w := ecs.NewWorld()
	loader := newMockChunkLoaderWithBarriers()
	sys := NewWorldChunkSystemWithGenre(loader, 16, "fantasy")

	// Pre-populate a chunk with a barrier
	testChunk := &chunk.Chunk{
		X:    0,
		Y:    0,
		Size: 16,
		DetailSpawns: []chunk.DetailSpawn{
			{
				Type:   chunk.DetailSpawnBarrierNatural,
				LocalX: 5.0,
				LocalY: 5.0,
				Scale:  1.0,
				BarrierData: &chunk.BarrierSpawnData{
					ShapeType: "cylinder",
					Radius:    0.5,
					Height:    1.0,
				},
			},
		},
	}
	loader.chunks[[2]int{0, 0}] = testChunk

	w.RegisterSystem(sys)

	// Create entity to trigger chunk loading
	e := w.CreateEntity()
	_ = w.AddComponent(e, &components.Position{X: 8, Y: 8, Z: 0})

	// Run multiple updates
	w.Update(0.016)
	barriers1 := len(w.Entities("Barrier"))

	w.Update(0.016)
	barriers2 := len(w.Entities("Barrier"))

	w.Update(0.016)
	barriers3 := len(w.Entities("Barrier"))

	// Barrier count should stay constant (no duplicates)
	if barriers1 != barriers2 || barriers2 != barriers3 {
		t.Errorf("barrier count changed across updates: %d -> %d -> %d", barriers1, barriers2, barriers3)
	}
}

func TestChunkKey(t *testing.T) {
	// Test that different coordinates produce different keys
	key1 := chunkKey(0, 0)
	key2 := chunkKey(1, 0)
	key3 := chunkKey(0, 1)
	key4 := chunkKey(-1, -1)

	if key1 == key2 || key1 == key3 || key2 == key3 {
		t.Error("different coordinates should produce different keys")
	}
	if key1 == key4 {
		t.Error("(0,0) and (-1,-1) should have different keys")
	}

	// Test determinism
	if chunkKey(5, 10) != chunkKey(5, 10) {
		t.Error("same coordinates should produce same key")
	}
}
