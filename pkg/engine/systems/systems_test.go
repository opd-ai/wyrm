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
