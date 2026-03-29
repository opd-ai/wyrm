package ecs

import (
	"testing"
)

// mockComponent is a test component implementation.
type mockComponent struct {
	value int
}

func (m *mockComponent) Type() string { return "Mock" }

// anotherComponent is a second test component.
type anotherComponent struct {
	data string
}

func (a *anotherComponent) Type() string { return "Another" }

// mockSystem is a test system implementation.
type mockSystem struct {
	updateCount int
	lastDT      float64
}

func (m *mockSystem) Update(w *World, dt float64) {
	m.updateCount++
	m.lastDT = dt
}

func TestNewWorld(t *testing.T) {
	w := NewWorld()
	if w == nil {
		t.Fatal("NewWorld returned nil")
	}
	if w.nextID != 1 {
		t.Errorf("expected nextID=1, got %d", w.nextID)
	}
	if w.components == nil {
		t.Error("components map not initialized")
	}
}

func TestCreateEntity(t *testing.T) {
	w := NewWorld()

	e1 := w.CreateEntity()
	if e1 != 1 {
		t.Errorf("expected first entity ID=1, got %d", e1)
	}

	e2 := w.CreateEntity()
	if e2 != 2 {
		t.Errorf("expected second entity ID=2, got %d", e2)
	}

	// Verify IDs are unique
	if e1 == e2 {
		t.Error("entity IDs should be unique")
	}
}

func TestDestroyEntity(t *testing.T) {
	w := NewWorld()
	e := w.CreateEntity()
	_ = w.AddComponent(e, &mockComponent{value: 42})

	w.DestroyEntity(e)

	// Verify entity is gone
	_, ok := w.GetComponent(e, "Mock")
	if ok {
		t.Error("destroyed entity should not have components")
	}
}

func TestAddComponent(t *testing.T) {
	w := NewWorld()
	e := w.CreateEntity()

	err := w.AddComponent(e, &mockComponent{value: 42})
	if err != nil {
		t.Fatalf("AddComponent failed: %v", err)
	}

	// Verify component was added
	c, ok := w.GetComponent(e, "Mock")
	if !ok {
		t.Fatal("component not found after add")
	}
	mc := c.(*mockComponent)
	if mc.value != 42 {
		t.Errorf("expected value=42, got %d", mc.value)
	}
}

func TestAddComponentEntityNotFound(t *testing.T) {
	w := NewWorld()
	err := w.AddComponent(999, &mockComponent{value: 1})
	if err != ErrEntityNotFound {
		t.Errorf("expected ErrEntityNotFound, got %v", err)
	}
}

func TestGetComponent(t *testing.T) {
	w := NewWorld()
	e := w.CreateEntity()
	_ = w.AddComponent(e, &mockComponent{value: 100})

	// Get existing component
	c, ok := w.GetComponent(e, "Mock")
	if !ok {
		t.Fatal("expected to find component")
	}
	mc := c.(*mockComponent)
	if mc.value != 100 {
		t.Errorf("expected value=100, got %d", mc.value)
	}

	// Get non-existent component type
	_, ok = w.GetComponent(e, "NonExistent")
	if ok {
		t.Error("expected false for non-existent component type")
	}

	// Get from non-existent entity
	_, ok = w.GetComponent(999, "Mock")
	if ok {
		t.Error("expected false for non-existent entity")
	}
}

func TestEntities(t *testing.T) {
	w := NewWorld()

	// Create entities with various component combinations
	e1 := w.CreateEntity()
	_ = w.AddComponent(e1, &mockComponent{value: 1})

	e2 := w.CreateEntity()
	_ = w.AddComponent(e2, &mockComponent{value: 2})
	_ = w.AddComponent(e2, &anotherComponent{data: "test"})

	e3 := w.CreateEntity()
	_ = w.AddComponent(e3, &anotherComponent{data: "only"})

	// Query for Mock component
	mockEntities := w.Entities("Mock")
	if len(mockEntities) != 2 {
		t.Errorf("expected 2 entities with Mock, got %d", len(mockEntities))
	}

	// Query for Another component
	anotherEntities := w.Entities("Another")
	if len(anotherEntities) != 2 {
		t.Errorf("expected 2 entities with Another, got %d", len(anotherEntities))
	}

	// Query for both components
	bothEntities := w.Entities("Mock", "Another")
	if len(bothEntities) != 1 {
		t.Errorf("expected 1 entity with both components, got %d", len(bothEntities))
	}
	if bothEntities[0] != e2 {
		t.Errorf("expected entity %d, got %d", e2, bothEntities[0])
	}

	// Query for non-existent component
	emptyEntities := w.Entities("NonExistent")
	if len(emptyEntities) != 0 {
		t.Errorf("expected 0 entities, got %d", len(emptyEntities))
	}

	// Verify sorted order
	allMock := w.Entities("Mock")
	for i := 1; i < len(allMock); i++ {
		if allMock[i-1] >= allMock[i] {
			t.Error("entities should be sorted by ID")
		}
	}
}

func TestRegisterSystem(t *testing.T) {
	w := NewWorld()
	s := &mockSystem{}

	w.RegisterSystem(s)

	if len(w.systems) != 1 {
		t.Errorf("expected 1 system, got %d", len(w.systems))
	}
}

func TestWorldUpdate(t *testing.T) {
	w := NewWorld()
	s1 := &mockSystem{}
	s2 := &mockSystem{}

	w.RegisterSystem(s1)
	w.RegisterSystem(s2)

	w.Update(0.016)

	if s1.updateCount != 1 {
		t.Errorf("expected s1 updateCount=1, got %d", s1.updateCount)
	}
	if s2.updateCount != 1 {
		t.Errorf("expected s2 updateCount=1, got %d", s2.updateCount)
	}
	if s1.lastDT != 0.016 {
		t.Errorf("expected dt=0.016, got %f", s1.lastDT)
	}
}

func TestWorldUpdateMultipleTicks(t *testing.T) {
	w := NewWorld()
	s := &mockSystem{}
	w.RegisterSystem(s)

	for i := 0; i < 10; i++ {
		w.Update(1.0 / 60.0)
	}

	if s.updateCount != 10 {
		t.Errorf("expected 10 updates, got %d", s.updateCount)
	}
}

func BenchmarkCreateDestroy(b *testing.B) {
	w := NewWorld()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Create and destroy 10,000 entities
		entities := make([]Entity, 10000)
		for j := 0; j < 10000; j++ {
			entities[j] = w.CreateEntity()
		}
		for j := 0; j < 10000; j++ {
			w.DestroyEntity(entities[j])
		}
	}
}

func BenchmarkAddComponent(b *testing.B) {
	w := NewWorld()
	entities := make([]Entity, b.N)
	for i := 0; i < b.N; i++ {
		entities[i] = w.CreateEntity()
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = w.AddComponent(entities[i], &mockComponent{value: i})
	}
}

func BenchmarkEntitiesQuery(b *testing.B) {
	w := NewWorld()
	// Create 10,000 entities with components
	for i := 0; i < 10000; i++ {
		e := w.CreateEntity()
		_ = w.AddComponent(e, &mockComponent{value: i})
		if i%2 == 0 {
			_ = w.AddComponent(e, &anotherComponent{data: "test"})
		}
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = w.Entities("Mock", "Another")
	}
}

func BenchmarkWorldUpdate(b *testing.B) {
	w := NewWorld()
	for i := 0; i < 10; i++ {
		w.RegisterSystem(&mockSystem{})
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w.Update(0.016)
	}
}
