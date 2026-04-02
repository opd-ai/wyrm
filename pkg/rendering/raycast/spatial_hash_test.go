//go:build noebiten

package raycast

import (
	"testing"
)

func TestSpatialHashInsertAndQuery(t *testing.T) {
	sh := NewSpatialHash(5.0)

	e1 := &SpriteEntity{X: 2.5, Y: 2.5}
	e2 := &SpriteEntity{X: 7.5, Y: 2.5}
	e3 := &SpriteEntity{X: 2.5, Y: 7.5}

	sh.Insert(e1)
	sh.Insert(e2)
	sh.Insert(e3)

	if sh.Count() != 3 {
		t.Errorf("Count() = %d, want 3", sh.Count())
	}

	// Query cell containing e1
	result := sh.Query(2.5, 2.5)
	if len(result) != 1 || result[0] != e1 {
		t.Error("Query should return e1")
	}

	// Query cell containing e2
	result = sh.Query(7.5, 2.5)
	if len(result) != 1 || result[0] != e2 {
		t.Error("Query should return e2")
	}
}

func TestSpatialHashRemove(t *testing.T) {
	sh := NewSpatialHash(5.0)

	e1 := &SpriteEntity{X: 2.5, Y: 2.5}
	e2 := &SpriteEntity{X: 2.5, Y: 2.5}

	sh.Insert(e1)
	sh.Insert(e2)

	if sh.Count() != 2 {
		t.Errorf("Count() = %d, want 2", sh.Count())
	}

	sh.Remove(e1)

	if sh.Count() != 1 {
		t.Errorf("Count() = %d, want 1", sh.Count())
	}

	result := sh.Query(2.5, 2.5)
	if len(result) != 1 || result[0] != e2 {
		t.Error("Query should return only e2 after removing e1")
	}
}

func TestSpatialHashUpdate(t *testing.T) {
	sh := NewSpatialHash(5.0)

	e := &SpriteEntity{X: 2.5, Y: 2.5}
	sh.Insert(e)

	// Update to same cell - should be no-op
	e.X = 3.0
	sh.Update(e)

	result := sh.Query(2.5, 2.5)
	if len(result) != 1 {
		t.Error("Entity should still be in same cell")
	}

	// Update to different cell
	e.X = 7.5
	sh.Update(e)

	result = sh.Query(2.5, 2.5)
	if len(result) != 0 {
		t.Error("Old cell should be empty after move")
	}

	result = sh.Query(7.5, 2.5)
	if len(result) != 1 || result[0] != e {
		t.Error("Entity should be in new cell")
	}
}

func TestSpatialHashQueryRadius(t *testing.T) {
	sh := NewSpatialHash(5.0)

	// Place entities in a grid pattern
	e1 := &SpriteEntity{X: 0.0, Y: 0.0}
	e2 := &SpriteEntity{X: 3.0, Y: 0.0}
	e3 := &SpriteEntity{X: 10.0, Y: 0.0}
	e4 := &SpriteEntity{X: 20.0, Y: 0.0}

	sh.Insert(e1)
	sh.Insert(e2)
	sh.Insert(e3)
	sh.Insert(e4)

	// Query with radius 5 from origin
	result := sh.QueryRadius(0.0, 0.0, 5.0)
	if len(result) != 2 {
		t.Errorf("QueryRadius(0,0,5) = %d entities, want 2", len(result))
	}

	// Query with radius 15 from origin
	result = sh.QueryRadius(0.0, 0.0, 15.0)
	if len(result) != 3 {
		t.Errorf("QueryRadius(0,0,15) = %d entities, want 3", len(result))
	}
}

func TestSpatialHashQueryRect(t *testing.T) {
	sh := NewSpatialHash(5.0)

	e1 := &SpriteEntity{X: 2.5, Y: 2.5}
	e2 := &SpriteEntity{X: 7.5, Y: 2.5}
	e3 := &SpriteEntity{X: 2.5, Y: 7.5}
	e4 := &SpriteEntity{X: 12.5, Y: 12.5}

	sh.Insert(e1)
	sh.Insert(e2)
	sh.Insert(e3)
	sh.Insert(e4)

	// Query rectangle containing first 3 entities
	result := sh.QueryRect(0.0, 0.0, 10.0, 10.0)
	if len(result) != 3 {
		t.Errorf("QueryRect(0,0,10,10) = %d entities, want 3", len(result))
	}

	// Query smaller rectangle
	result = sh.QueryRect(0.0, 0.0, 5.0, 5.0)
	if len(result) != 1 {
		t.Errorf("QueryRect(0,0,5,5) = %d entities, want 1", len(result))
	}
}

func TestSpatialHashClear(t *testing.T) {
	sh := NewSpatialHash(5.0)

	sh.Insert(&SpriteEntity{X: 0.0, Y: 0.0})
	sh.Insert(&SpriteEntity{X: 5.0, Y: 5.0})
	sh.Insert(&SpriteEntity{X: 10.0, Y: 10.0})

	if sh.Count() != 3 {
		t.Errorf("Count() = %d, want 3", sh.Count())
	}

	sh.Clear()

	if sh.Count() != 0 {
		t.Errorf("Count() after Clear() = %d, want 0", sh.Count())
	}
	if sh.CellCount() != 0 {
		t.Errorf("CellCount() after Clear() = %d, want 0", sh.CellCount())
	}
}

func TestSpatialHashNearestEntity(t *testing.T) {
	sh := NewSpatialHash(5.0)

	e1 := &SpriteEntity{X: 0.0, Y: 0.0}
	e2 := &SpriteEntity{X: 3.0, Y: 0.0}
	e3 := &SpriteEntity{X: 10.0, Y: 0.0}

	sh.Insert(e1)
	sh.Insert(e2)
	sh.Insert(e3)

	// Nearest to (2.0, 0.0) should be e2
	nearest := sh.NearestEntity(2.0, 0.0, 10.0)
	if nearest != e2 {
		t.Error("NearestEntity should return e2 (closest to 2.0, 0.0)")
	}

	// Nearest to (0.0, 0.0) should be e1
	nearest = sh.NearestEntity(0.0, 0.0, 10.0)
	if nearest != e1 {
		t.Error("NearestEntity should return e1 (at origin)")
	}

	// No entity within 0.5 of (100.0, 100.0)
	nearest = sh.NearestEntity(100.0, 100.0, 0.5)
	if nearest != nil {
		t.Error("NearestEntity should return nil when no entity in range")
	}
}

func TestSpatialHashNearestInteractable(t *testing.T) {
	sh := NewSpatialHash(5.0)

	e1 := &SpriteEntity{X: 0.0, Y: 0.0, IsInteractable: false, Visible: true}
	e2 := &SpriteEntity{X: 3.0, Y: 0.0, IsInteractable: true, Visible: true}
	e3 := &SpriteEntity{X: 5.0, Y: 0.0, IsInteractable: true, Visible: false}

	sh.Insert(e1)
	sh.Insert(e2)
	sh.Insert(e3)

	// e1 is closest but not interactable
	// e3 is interactable but not visible
	// e2 is interactable and visible
	nearest := sh.NearestInteractable(0.0, 0.0, 10.0)
	if nearest != e2 {
		t.Error("NearestInteractable should return e2 (interactable and visible)")
	}
}

func TestSpatialHashInsertAll(t *testing.T) {
	sh := NewSpatialHash(5.0)

	entities := []*SpriteEntity{
		{X: 0.0, Y: 0.0},
		{X: 5.0, Y: 0.0},
		{X: 10.0, Y: 0.0},
	}

	sh.InsertAll(entities)

	if sh.Count() != 3 {
		t.Errorf("Count() = %d, want 3", sh.Count())
	}
}

func TestSpatialHashCellKey(t *testing.T) {
	// Test that different coordinates produce different keys
	key1 := cellKey(0, 0)
	key2 := cellKey(1, 0)
	key3 := cellKey(0, 1)
	key4 := cellKey(-1, 0)

	if key1 == key2 {
		t.Error("Different cell X should produce different key")
	}
	if key1 == key3 {
		t.Error("Different cell Y should produce different key")
	}
	if key2 == key4 {
		t.Error("Positive and negative X should produce different key")
	}
}

func TestSpatialHashDefaultCellSize(t *testing.T) {
	sh := NewSpatialHash(0) // Should use default
	if sh.CellSize != 5.0 {
		t.Errorf("Default CellSize = %f, want 5.0", sh.CellSize)
	}

	sh = NewSpatialHash(-1) // Should use default
	if sh.CellSize != 5.0 {
		t.Errorf("Default CellSize for negative input = %f, want 5.0", sh.CellSize)
	}
}

func TestSpatialHashNilEntity(t *testing.T) {
	sh := NewSpatialHash(5.0)

	// These should not panic
	sh.Insert(nil)
	sh.Remove(nil)
	sh.Update(nil)

	if sh.Count() != 0 {
		t.Error("Nil insertions should not change count")
	}
}

func BenchmarkSpatialHashInsert(b *testing.B) {
	sh := NewSpatialHash(5.0)
	entities := make([]*SpriteEntity, 1000)
	for i := range entities {
		entities[i] = &SpriteEntity{X: float64(i % 100), Y: float64(i / 100)}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sh.Clear()
		for _, e := range entities {
			sh.Insert(e)
		}
	}
}

func BenchmarkSpatialHashQueryRadius(b *testing.B) {
	sh := NewSpatialHash(5.0)
	for i := 0; i < 1000; i++ {
		sh.Insert(&SpriteEntity{X: float64(i % 100), Y: float64(i / 100)})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sh.QueryRadius(50.0, 5.0, 10.0)
	}
}

func BenchmarkSpatialHashNearestEntity(b *testing.B) {
	sh := NewSpatialHash(5.0)
	for i := 0; i < 1000; i++ {
		sh.Insert(&SpriteEntity{X: float64(i % 100), Y: float64(i / 100)})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sh.NearestEntity(50.0, 5.0, 10.0)
	}
}
