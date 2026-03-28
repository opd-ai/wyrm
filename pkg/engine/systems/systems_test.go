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
	vehicle := &components.Vehicle{VehicleType: "car", Speed: 10, Fuel: 100}
	_ = w.AddComponent(e, pos)
	_ = w.AddComponent(e, vehicle)

	initialX := pos.X
	initialFuel := vehicle.Fuel

	w.Update(1.0) // 1 second update

	// Position should have changed
	if pos.X <= initialX {
		t.Error("vehicle position X should have increased")
	}

	// Fuel should have decreased
	if vehicle.Fuel >= initialFuel {
		t.Error("vehicle fuel should have decreased")
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
	sys := &WeatherSystem{}
	w.RegisterSystem(sys)

	if sys.CurrentWeather != "" {
		t.Error("initial weather should be empty")
	}

	w.Update(0.016)

	if sys.CurrentWeather != "clear" {
		t.Errorf("weather should be 'clear' after first update, got '%s'", sys.CurrentWeather)
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
		&FactionPoliticsSystem{},
		&CrimeSystem{},
		&EconomySystem{},
		&CombatSystem{},
		&VehicleSystem{},
		&QuestSystem{},
		&WeatherSystem{},
		&RenderSystem{},
		&AudioSystem{},
	}

	w := ecs.NewWorld()
	for _, s := range systems {
		// Each system should be able to Update without panic
		s.Update(w, 0.016)
	}
}
