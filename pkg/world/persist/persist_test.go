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
		name    string
		before  *WorldSnapshot
		after   *WorldSnapshot
		maxDiff float64 // Maximum acceptable diff percentage
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

func TestCompareCrime(t *testing.T) {
	tests := []struct {
		name      string
		before    EntityData
		after     EntityData
		wantTotal int
		wantDiff  int
	}{
		{
			name:      "both have no crime",
			before:    EntityData{HasCrime: false},
			after:     EntityData{HasCrime: false},
			wantTotal: 0,
			wantDiff:  0,
		},
		{
			name:      "before has crime, after doesn't - compares actual values",
			before:    EntityData{HasCrime: true, WantedLevel: 2, BountyAmount: 100.0, InJail: false},
			after:     EntityData{HasCrime: false, WantedLevel: 0, BountyAmount: 0.0, InJail: false},
			wantTotal: 3,
			wantDiff:  2, // WantedLevel and BountyAmount differ; InJail is same (both false)
		},
		{
			name:      "after has crime, before doesn't",
			before:    EntityData{HasCrime: false},
			after:     EntityData{HasCrime: true, WantedLevel: 3, BountyAmount: 200.0, InJail: true},
			wantTotal: 3,
			wantDiff:  3,
		},
		{
			name:      "both have identical crime",
			before:    EntityData{HasCrime: true, WantedLevel: 2, BountyAmount: 100.0, InJail: false},
			after:     EntityData{HasCrime: true, WantedLevel: 2, BountyAmount: 100.0, InJail: false},
			wantTotal: 3,
			wantDiff:  0,
		},
		{
			name:      "wanted level differs",
			before:    EntityData{HasCrime: true, WantedLevel: 2, BountyAmount: 100.0, InJail: false},
			after:     EntityData{HasCrime: true, WantedLevel: 5, BountyAmount: 100.0, InJail: false},
			wantTotal: 3,
			wantDiff:  1,
		},
		{
			name:      "bounty differs",
			before:    EntityData{HasCrime: true, WantedLevel: 2, BountyAmount: 100.0, InJail: false},
			after:     EntityData{HasCrime: true, WantedLevel: 2, BountyAmount: 500.0, InJail: false},
			wantTotal: 3,
			wantDiff:  1,
		},
		{
			name:      "in jail differs",
			before:    EntityData{HasCrime: true, WantedLevel: 2, BountyAmount: 100.0, InJail: false},
			after:     EntityData{HasCrime: true, WantedLevel: 2, BountyAmount: 100.0, InJail: true},
			wantTotal: 3,
			wantDiff:  1,
		},
		{
			name:      "all crime fields differ",
			before:    EntityData{HasCrime: true, WantedLevel: 1, BountyAmount: 50.0, InJail: false},
			after:     EntityData{HasCrime: true, WantedLevel: 5, BountyAmount: 1000.0, InJail: true},
			wantTotal: 3,
			wantDiff:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareCrime(tt.before, tt.after)
			if result.total != tt.wantTotal {
				t.Errorf("total = %d, want %d", result.total, tt.wantTotal)
			}
			if result.diff != tt.wantDiff {
				t.Errorf("diff = %d, want %d", result.diff, tt.wantDiff)
			}
		})
	}
}

func TestComparePosition(t *testing.T) {
	tests := []struct {
		name      string
		before    EntityData
		after     EntityData
		wantTotal int
		wantDiff  int
	}{
		{
			name:      "both no position",
			before:    EntityData{HasPosition: false},
			after:     EntityData{HasPosition: false},
			wantTotal: 0,
			wantDiff:  0,
		},
		{
			name:      "before has position, after doesn't",
			before:    EntityData{HasPosition: true, PosX: 10, PosY: 20, PosZ: 30, PosAngle: 1.5},
			after:     EntityData{HasPosition: false},
			wantTotal: 4,
			wantDiff:  4,
		},
		{
			name:      "identical positions",
			before:    EntityData{HasPosition: true, PosX: 10, PosY: 20, PosZ: 30, PosAngle: 1.5},
			after:     EntityData{HasPosition: true, PosX: 10, PosY: 20, PosZ: 30, PosAngle: 1.5},
			wantTotal: 4,
			wantDiff:  0,
		},
		{
			name:      "X position differs",
			before:    EntityData{HasPosition: true, PosX: 10, PosY: 20, PosZ: 30, PosAngle: 1.5},
			after:     EntityData{HasPosition: true, PosX: 15, PosY: 20, PosZ: 30, PosAngle: 1.5},
			wantTotal: 4,
			wantDiff:  1,
		},
		{
			name:      "all position fields differ",
			before:    EntityData{HasPosition: true, PosX: 10, PosY: 20, PosZ: 30, PosAngle: 1.5},
			after:     EntityData{HasPosition: true, PosX: 100, PosY: 200, PosZ: 300, PosAngle: 3.14},
			wantTotal: 4,
			wantDiff:  4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := comparePosition(tt.before, tt.after)
			if result.total != tt.wantTotal {
				t.Errorf("total = %d, want %d", result.total, tt.wantTotal)
			}
			if result.diff != tt.wantDiff {
				t.Errorf("diff = %d, want %d", result.diff, tt.wantDiff)
			}
		})
	}
}

func TestCompareHealth(t *testing.T) {
	tests := []struct {
		name      string
		before    EntityData
		after     EntityData
		wantTotal int
		wantDiff  int
	}{
		{
			name:      "both no health",
			before:    EntityData{HasHealth: false},
			after:     EntityData{HasHealth: false},
			wantTotal: 0,
			wantDiff:  0,
		},
		{
			name:      "identical health",
			before:    EntityData{HasHealth: true, HealthCurrent: 80, HealthMax: 100},
			after:     EntityData{HasHealth: true, HealthCurrent: 80, HealthMax: 100},
			wantTotal: 2,
			wantDiff:  0,
		},
		{
			name:      "current health differs",
			before:    EntityData{HasHealth: true, HealthCurrent: 80, HealthMax: 100},
			after:     EntityData{HasHealth: true, HealthCurrent: 50, HealthMax: 100},
			wantTotal: 2,
			wantDiff:  1,
		},
		{
			name:      "both differ",
			before:    EntityData{HasHealth: true, HealthCurrent: 80, HealthMax: 100},
			after:     EntityData{HasHealth: true, HealthCurrent: 50, HealthMax: 150},
			wantTotal: 2,
			wantDiff:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareHealth(tt.before, tt.after)
			if result.total != tt.wantTotal {
				t.Errorf("total = %d, want %d", result.total, tt.wantTotal)
			}
			if result.diff != tt.wantDiff {
				t.Errorf("diff = %d, want %d", result.diff, tt.wantDiff)
			}
		})
	}
}

func TestCompareEntities(t *testing.T) {
	tests := []struct {
		name      string
		before    []EntityData
		after     []EntityData
		wantTotal int
		wantDiff  int
	}{
		{
			name:      "both empty",
			before:    nil,
			after:     nil,
			wantTotal: 0,
			wantDiff:  0,
		},
		{
			name:      "different entity counts",
			before:    []EntityData{{ID: 1}},
			after:     []EntityData{{ID: 1}, {ID: 2}},
			wantTotal: 1, // At least counts as different
			wantDiff:  1,
		},
		{
			name: "identical entities",
			before: []EntityData{
				{ID: 1, HasPosition: true, PosX: 10, PosY: 20, PosZ: 30, PosAngle: 1.5},
			},
			after: []EntityData{
				{ID: 1, HasPosition: true, PosX: 10, PosY: 20, PosZ: 30, PosAngle: 1.5},
			},
			wantTotal: 4, // 4 position fields
			wantDiff:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total, diff := compareEntities(tt.before, tt.after)
			if total < tt.wantTotal {
				t.Errorf("total = %d, want >= %d", total, tt.wantTotal)
			}
			if diff < tt.wantDiff {
				t.Errorf("diff = %d, want >= %d", diff, tt.wantDiff)
			}
		})
	}
}

func TestCompareMetadata(t *testing.T) {
	tests := []struct {
		name      string
		before    *WorldSnapshot
		after     *WorldSnapshot
		wantTotal int
		wantDiff  int
	}{
		{
			name:      "identical metadata",
			before:    &WorldSnapshot{Seed: 123, Genre: "fantasy", Entities: []EntityData{}},
			after:     &WorldSnapshot{Seed: 123, Genre: "fantasy", Entities: []EntityData{}},
			wantTotal: 3,
			wantDiff:  0,
		},
		{
			name:      "different seed",
			before:    &WorldSnapshot{Seed: 123, Genre: "fantasy", Entities: []EntityData{}},
			after:     &WorldSnapshot{Seed: 456, Genre: "fantasy", Entities: []EntityData{}},
			wantTotal: 3,
			wantDiff:  1,
		},
		{
			name:      "different genre",
			before:    &WorldSnapshot{Seed: 123, Genre: "fantasy", Entities: []EntityData{}},
			after:     &WorldSnapshot{Seed: 123, Genre: "sci-fi", Entities: []EntityData{}},
			wantTotal: 3,
			wantDiff:  1,
		},
		{
			name:      "different entity count",
			before:    &WorldSnapshot{Seed: 123, Genre: "fantasy", Entities: []EntityData{}},
			after:     &WorldSnapshot{Seed: 123, Genre: "fantasy", Entities: []EntityData{{ID: 1}}},
			wantTotal: 3,
			wantDiff:  1,
		},
		{
			name:      "all metadata differs",
			before:    &WorldSnapshot{Seed: 123, Genre: "fantasy", Entities: []EntityData{}},
			after:     &WorldSnapshot{Seed: 999, Genre: "horror", Entities: []EntityData{{ID: 1}, {ID: 2}}},
			wantTotal: 3,
			wantDiff:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total, diff := compareMetadata(tt.before, tt.after)
			if total != tt.wantTotal {
				t.Errorf("total = %d, want %d", total, tt.wantTotal)
			}
			if diff != tt.wantDiff {
				t.Errorf("diff = %d, want %d", diff, tt.wantDiff)
			}
		})
	}
}

func TestLastSaveTimeNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wyrm_persist_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	p := NewPersister(tmpDir)

	// LastSaveTime for non-existent seed
	_, err = p.LastSaveTime(99999)
	if err == nil {
		t.Error("LastSaveTime should return error for non-existent seed")
	}
}

// WorldConsequenceTracker tests

func TestNewWorldConsequenceTracker(t *testing.T) {
	tracker := NewWorldConsequenceTracker()
	if tracker == nil {
		t.Fatal("NewWorldConsequenceTracker returned nil")
	}
	if tracker.Consequences == nil {
		t.Error("Consequences map should be initialized")
	}
	if tracker.ByChunk == nil {
		t.Error("ByChunk map should be initialized")
	}
	if tracker.ByType == nil {
		t.Error("ByType map should be initialized")
	}
	if tracker.ByPlayer == nil {
		t.Error("ByPlayer map should be initialized")
	}
}

func TestRecordConsequence(t *testing.T) {
	tracker := NewWorldConsequenceTracker()

	c := &WorldConsequence{
		Type:           ConsequenceNPCKilled,
		CausedByPlayer: 100,
		AffectedEntity: 200,
		ChunkX:         5,
		ChunkY:         10,
	}
	tracker.RecordConsequence(c)

	if c.ID == "" {
		t.Error("Consequence should have generated ID")
	}
	if c.Timestamp.IsZero() {
		t.Error("Consequence should have timestamp")
	}

	// Check stored in main map
	stored := tracker.GetConsequence(c.ID)
	if stored == nil {
		t.Fatal("Consequence should be retrievable")
	}
	if stored.Type != ConsequenceNPCKilled {
		t.Errorf("Type = %v, want %v", stored.Type, ConsequenceNPCKilled)
	}
}

func TestConsequenceIndexes(t *testing.T) {
	tracker := NewWorldConsequenceTracker()

	c1 := &WorldConsequence{
		Type:           ConsequenceNPCKilled,
		CausedByPlayer: 100,
		ChunkX:         5,
		ChunkY:         10,
	}
	c2 := &WorldConsequence{
		Type:           ConsequenceBuildingDestroyed,
		CausedByPlayer: 100,
		ChunkX:         5,
		ChunkY:         10,
	}
	c3 := &WorldConsequence{
		Type:           ConsequenceNPCKilled,
		CausedByPlayer: 200,
		ChunkX:         8,
		ChunkY:         15,
	}

	tracker.RecordConsequence(c1)
	tracker.RecordConsequence(c2)
	tracker.RecordConsequence(c3)

	// Test chunk index
	chunkCons := tracker.GetChunkConsequences(5, 10)
	if len(chunkCons) != 2 {
		t.Errorf("Chunk (5,10) has %d consequences, want 2", len(chunkCons))
	}

	// Test type index
	npcKilledCons := tracker.GetTypeConsequences(ConsequenceNPCKilled)
	if len(npcKilledCons) != 2 {
		t.Errorf("NPC_KILLED has %d consequences, want 2", len(npcKilledCons))
	}

	// Test player index
	player100Cons := tracker.GetPlayerConsequences(100)
	if len(player100Cons) != 2 {
		t.Errorf("Player 100 has %d consequences, want 2", len(player100Cons))
	}
}

func TestReverseConsequence(t *testing.T) {
	tracker := NewWorldConsequenceTracker()

	c1 := &WorldConsequence{
		Type:         ConsequenceBuildingDestroyed,
		IsReversible: true,
	}
	c2 := &WorldConsequence{
		Type:         ConsequenceNPCKilled,
		IsReversible: false,
	}

	tracker.RecordConsequence(c1)
	tracker.RecordConsequence(c2)

	// Should succeed for reversible
	if !tracker.ReverseConsequence(c1.ID) {
		t.Error("Should be able to reverse reversible consequence")
	}
	if tracker.GetConsequence(c1.ID).ReversedAt.IsZero() {
		t.Error("ReversedAt should be set")
	}

	// Should fail for non-reversible
	if tracker.ReverseConsequence(c2.ID) {
		t.Error("Should not be able to reverse non-reversible consequence")
	}

	// Should fail for non-existent
	if tracker.ReverseConsequence("fake_id") {
		t.Error("Should not be able to reverse non-existent consequence")
	}
}

func TestGetActiveConsequences(t *testing.T) {
	tracker := NewWorldConsequenceTracker()

	c1 := &WorldConsequence{Type: "type1", IsReversible: true}
	c2 := &WorldConsequence{Type: "type2"}
	c3 := &WorldConsequence{Type: "type3", IsReversible: true}

	tracker.RecordConsequence(c1)
	tracker.RecordConsequence(c2)
	tracker.RecordConsequence(c3)
	tracker.ReverseConsequence(c1.ID)

	active := tracker.GetActiveConsequences()
	if len(active) != 2 {
		t.Errorf("Active consequences = %d, want 2", len(active))
	}

	// Check reversed one is not in active
	for _, a := range active {
		if a.ID == c1.ID {
			t.Error("Reversed consequence should not be in active list")
		}
	}
}

func TestRecordNPCKilled(t *testing.T) {
	tracker := NewWorldConsequenceTracker()

	RecordNPCKilled(tracker, 200, 100, 50.0, 60.0, 0.0, 5, 6, "Guard Bob", "guard")

	cons := tracker.GetTypeConsequences(ConsequenceNPCKilled)
	if len(cons) != 1 {
		t.Fatalf("Should have 1 NPC killed consequence, got %d", len(cons))
	}

	c := cons[0]
	if c.CausedByPlayer != 100 {
		t.Errorf("CausedByPlayer = %d, want 100", c.CausedByPlayer)
	}
	if c.AffectedEntity != 200 {
		t.Errorf("AffectedEntity = %d, want 200", c.AffectedEntity)
	}
	if c.IsReversible {
		t.Error("NPC death should not be reversible")
	}
	if c.Data["npc_name"] != "Guard Bob" {
		t.Errorf("npc_name = %v, want 'Guard Bob'", c.Data["npc_name"])
	}
}

func TestRecordBuildingDestroyed(t *testing.T) {
	tracker := NewWorldConsequenceTracker()

	RecordBuildingDestroyed(tracker, 300, 100, 100.0, 200.0, 10, 20, "tavern")

	cons := tracker.GetTypeConsequences(ConsequenceBuildingDestroyed)
	if len(cons) != 1 {
		t.Fatalf("Should have 1 building destroyed consequence, got %d", len(cons))
	}

	c := cons[0]
	if !c.IsReversible {
		t.Error("Building destruction should be reversible")
	}
	if c.Data["building_type"] != "tavern" {
		t.Errorf("building_type = %v, want 'tavern'", c.Data["building_type"])
	}
}

func TestRecordFactionWar(t *testing.T) {
	tracker := NewWorldConsequenceTracker()

	RecordFactionWar(tracker, "guards", "thieves", 100)

	cons := tracker.GetTypeConsequences(ConsequenceFactionWar)
	if len(cons) != 1 {
		t.Fatalf("Should have 1 faction war consequence, got %d", len(cons))
	}

	c := cons[0]
	if c.Data["faction1"] != "guards" {
		t.Errorf("faction1 = %v, want 'guards'", c.Data["faction1"])
	}
	if c.Data["faction2"] != "thieves" {
		t.Errorf("faction2 = %v, want 'thieves'", c.Data["faction2"])
	}
}

func TestRecordQuestCompleted(t *testing.T) {
	tracker := NewWorldConsequenceTracker()

	RecordQuestCompleted(tracker, 100, "kill_dragon", "gold")

	cons := tracker.GetTypeConsequences(ConsequenceQuestCompleted)
	if len(cons) != 1 {
		t.Fatalf("Should have 1 quest completed consequence, got %d", len(cons))
	}

	c := cons[0]
	if c.CausedByPlayer != 100 {
		t.Errorf("CausedByPlayer = %d, want 100", c.CausedByPlayer)
	}
	if c.Data["quest_id"] != "kill_dragon" {
		t.Errorf("quest_id = %v, want 'kill_dragon'", c.Data["quest_id"])
	}
}

func TestSerializeAndLoadConsequences(t *testing.T) {
	tracker := NewWorldConsequenceTracker()

	c1 := &WorldConsequence{
		Type:           ConsequenceNPCKilled,
		CausedByPlayer: 100,
		AffectedEntity: 200,
		ChunkX:         5,
		ChunkY:         10,
	}
	c2 := &WorldConsequence{
		Type:         ConsequenceBuildingDestroyed,
		IsReversible: true,
		ChunkX:       8,
		ChunkY:       15,
	}
	tracker.RecordConsequence(c1)
	tracker.RecordConsequence(c2)

	// Serialize
	data := tracker.SerializeConsequences()
	if len(data) != 2 {
		t.Fatalf("Serialized %d consequences, want 2", len(data))
	}

	// Load into new tracker
	newTracker := NewWorldConsequenceTracker()
	newTracker.LoadConsequences(data)

	if newTracker.TotalCount() != 2 {
		t.Errorf("Loaded %d consequences, want 2", newTracker.TotalCount())
	}

	// Verify indexes were rebuilt
	npcKilled := newTracker.GetTypeConsequences(ConsequenceNPCKilled)
	if len(npcKilled) != 1 {
		t.Errorf("Type index has %d NPC_KILLED, want 1", len(npcKilled))
	}
}

func TestCountByType(t *testing.T) {
	tracker := NewWorldConsequenceTracker()

	tracker.RecordConsequence(&WorldConsequence{Type: ConsequenceNPCKilled})
	tracker.RecordConsequence(&WorldConsequence{Type: ConsequenceNPCKilled})
	tracker.RecordConsequence(&WorldConsequence{Type: ConsequenceBuildingDestroyed})

	counts := tracker.CountByType()
	if counts[ConsequenceNPCKilled] != 2 {
		t.Errorf("NPC_KILLED count = %d, want 2", counts[ConsequenceNPCKilled])
	}
	if counts[ConsequenceBuildingDestroyed] != 1 {
		t.Errorf("BUILDING_DESTROYED count = %d, want 1", counts[ConsequenceBuildingDestroyed])
	}
}

func TestTotalCount(t *testing.T) {
	tracker := NewWorldConsequenceTracker()

	if tracker.TotalCount() != 0 {
		t.Error("Empty tracker should have 0 count")
	}

	tracker.RecordConsequence(&WorldConsequence{Type: "type1"})
	tracker.RecordConsequence(&WorldConsequence{Type: "type2"})
	tracker.RecordConsequence(&WorldConsequence{Type: "type3"})

	if tracker.TotalCount() != 3 {
		t.Errorf("TotalCount = %d, want 3", tracker.TotalCount())
	}
}
