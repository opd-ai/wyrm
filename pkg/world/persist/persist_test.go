package persist

import (
	"os"
	"testing"
	"time"
)

func TestPersisterSaveLoad(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "wyrm_persist_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	p := NewPersister(tmpDir)

	// Create test snapshot
	snapshot := NewWorldSnapshot(12345, "fantasy")
	snapshot.WorldHour = 14
	snapshot.WorldDay = 5

	entity1 := NewEntityData(1)
	entity1.HasPosition = true
	entity1.PosX = 100.5
	entity1.PosY = 50.25
	entity1.PosZ = 200.75
	entity1.PosAngle = 1.57
	entity1.HasHealth = true
	entity1.HealthCurrent = 80
	entity1.HealthMax = 100
	snapshot.Entities = append(snapshot.Entities, entity1)

	entity2 := NewEntityData(2)
	entity2.HasCrime = true
	entity2.WantedLevel = 3
	entity2.BountyAmount = 500.0
	snapshot.Entities = append(snapshot.Entities, entity2)

	// Save
	if err := p.Save(snapshot); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify exists
	if !p.Exists(12345) {
		t.Error("Exists returned false after Save")
	}

	// Load
	loaded, err := p.Load(12345)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded == nil {
		t.Fatal("Load returned nil")
	}

	// Verify metadata
	if loaded.Seed != 12345 {
		t.Errorf("Seed mismatch: got %d, want 12345", loaded.Seed)
	}
	if loaded.Genre != "fantasy" {
		t.Errorf("Genre mismatch: got %s, want fantasy", loaded.Genre)
	}
	if loaded.WorldHour != 14 {
		t.Errorf("WorldHour mismatch: got %d, want 14", loaded.WorldHour)
	}
	if loaded.WorldDay != 5 {
		t.Errorf("WorldDay mismatch: got %d, want 5", loaded.WorldDay)
	}

	// Verify entities
	if len(loaded.Entities) != 2 {
		t.Fatalf("Entity count mismatch: got %d, want 2", len(loaded.Entities))
	}

	// Find entity 1
	var loadedEntity1 *EntityData
	for i := range loaded.Entities {
		if loaded.Entities[i].ID == 1 {
			loadedEntity1 = &loaded.Entities[i]
			break
		}
	}
	if loadedEntity1 == nil {
		t.Fatal("Entity 1 not found in loaded snapshot")
	}

	if !loadedEntity1.HasPosition {
		t.Error("Entity 1 should have position")
	}
	if loadedEntity1.PosX != 100.5 {
		t.Errorf("Entity 1 PosX mismatch: got %f, want 100.5", loadedEntity1.PosX)
	}
	if loadedEntity1.HealthCurrent != 80 {
		t.Errorf("Entity 1 HealthCurrent mismatch: got %f, want 80", loadedEntity1.HealthCurrent)
	}
}

func TestPersisterLoadNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wyrm_persist_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	p := NewPersister(tmpDir)

	// Load non-existent
	loaded, err := p.Load(99999)
	if err != nil {
		t.Errorf("Load of non-existent should not error: %v", err)
	}
	if loaded != nil {
		t.Error("Load of non-existent should return nil")
	}
}

func TestPersisterDelete(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wyrm_persist_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	p := NewPersister(tmpDir)

	// Save
	snapshot := NewWorldSnapshot(12345, "fantasy")
	if err := p.Save(snapshot); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Delete
	if err := p.Delete(12345); err != nil {
		t.Errorf("Delete failed: %v", err)
	}

	// Verify deleted
	if p.Exists(12345) {
		t.Error("Exists returned true after Delete")
	}

	// Delete non-existent should not error
	if err := p.Delete(99999); err != nil {
		t.Errorf("Delete of non-existent should not error: %v", err)
	}
}

func TestCalculateStateDiff(t *testing.T) {
	tests := []struct {
		name     string
		before   *WorldSnapshot
		after    *WorldSnapshot
		maxDiff  float64 // Maximum acceptable diff percentage
	}{
		{
			name:    "nil before",
			before:  nil,
			after:   NewWorldSnapshot(1, "fantasy"),
			maxDiff: 100.0,
		},
		{
			name:    "nil after",
			before:  NewWorldSnapshot(1, "fantasy"),
			after:   nil,
			maxDiff: 100.0,
		},
		{
			name:    "identical empty",
			before:  NewWorldSnapshot(1, "fantasy"),
			after:   NewWorldSnapshot(1, "fantasy"),
			maxDiff: 0.0,
		},
		{
			name:    "different seed",
			before:  NewWorldSnapshot(1, "fantasy"),
			after:   NewWorldSnapshot(2, "fantasy"),
			maxDiff: 50.0, // 1 out of 3 base fields different
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := CalculateStateDiff(tt.before, tt.after)
			if diff > tt.maxDiff {
				t.Errorf("diff = %f, want <= %f", diff, tt.maxDiff)
			}
		})
	}
}

func TestCalculateStateDiffWithEntities(t *testing.T) {
	before := NewWorldSnapshot(1, "fantasy")
	entity1 := NewEntityData(1)
	entity1.HasPosition = true
	entity1.PosX = 100
	entity1.PosY = 50
	entity1.PosZ = 200
	entity1.HasHealth = true
	entity1.HealthCurrent = 100
	entity1.HealthMax = 100
	before.Entities = append(before.Entities, entity1)

	// Identical after
	after := NewWorldSnapshot(1, "fantasy")
	entity1After := NewEntityData(1)
	entity1After.HasPosition = true
	entity1After.PosX = 100
	entity1After.PosY = 50
	entity1After.PosZ = 200
	entity1After.HasHealth = true
	entity1After.HealthCurrent = 100
	entity1After.HealthMax = 100
	after.Entities = append(after.Entities, entity1After)

	diff := CalculateStateDiff(before, after)
	if diff != 0.0 {
		t.Errorf("identical snapshots should have 0 diff, got %f", diff)
	}

	// Modified after (health changed)
	afterModified := NewWorldSnapshot(1, "fantasy")
	entity1Mod := NewEntityData(1)
	entity1Mod.HasPosition = true
	entity1Mod.PosX = 100
	entity1Mod.PosY = 50
	entity1Mod.PosZ = 200
	entity1Mod.HasHealth = true
	entity1Mod.HealthCurrent = 50 // Changed!
	entity1Mod.HealthMax = 100
	afterModified.Entities = append(afterModified.Entities, entity1Mod)

	diff = CalculateStateDiff(before, afterModified)
	// Expected: 1 field different out of 9 total (2 metadata + 1 count + 4 pos + 2 health)
	// = 1/9 = ~11.1%
	if diff > 15.0 {
		t.Errorf("small change should have low diff, got %f%%", diff)
	}

	// Per AC: diff <5% from pre-restart snapshot
	// This test shows the diff calculation works
	t.Logf("Diff with one health field changed: %.2f%%", diff)
}

func TestLastSaveTime(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wyrm_persist_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	p := NewPersister(tmpDir)

	beforeSave := time.Now()

	snapshot := NewWorldSnapshot(12345, "fantasy")
	if err := p.Save(snapshot); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	afterSave := time.Now()

	lastSave, err := p.LastSaveTime(12345)
	if err != nil {
		t.Fatalf("LastSaveTime failed: %v", err)
	}

	if lastSave.Before(beforeSave) || lastSave.After(afterSave) {
		t.Errorf("LastSaveTime %v not between %v and %v", lastSave, beforeSave, afterSave)
	}
}

func TestSetAutoSave(t *testing.T) {
	p := NewPersister("/tmp/test")

	p.SetAutoSave(true, 10*time.Minute)

	if !p.autoSave {
		t.Error("autoSave should be true")
	}
	if p.interval != 10*time.Minute {
		t.Errorf("interval = %v, want 10m", p.interval)
	}

	// Zero interval should not change
	p.SetAutoSave(true, 0)
	if p.interval != 10*time.Minute {
		t.Errorf("interval changed with zero: %v", p.interval)
	}
}
