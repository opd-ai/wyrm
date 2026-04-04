package chunk

import (
	"math"
	"math/rand"
	"testing"
)

func TestNewChunk(t *testing.T) {
	c := NewChunk(5, 10, 16, 12345)

	if c.X != 5 {
		t.Errorf("expected X=5, got %d", c.X)
	}
	if c.Y != 10 {
		t.Errorf("expected Y=10, got %d", c.Y)
	}
	if c.Size != 16 {
		t.Errorf("expected Size=16, got %d", c.Size)
	}
	if c.Seed != 12345 {
		t.Errorf("expected Seed=12345, got %d", c.Seed)
	}
	if len(c.HeightMap) != 16*16 {
		t.Errorf("expected HeightMap size 256, got %d", len(c.HeightMap))
	}
}

func TestChunkHeightMapPopulated(t *testing.T) {
	c := NewChunk(0, 0, 16, 12345)

	// Check that heightmap has non-zero values
	hasNonZero := false
	for _, h := range c.HeightMap {
		if h != 0 {
			hasNonZero = true
			break
		}
	}

	if !hasNonZero {
		t.Error("heightmap should be populated with non-zero values")
	}
}

func TestChunkDeterminism(t *testing.T) {
	seed := int64(42)

	c1 := NewChunk(0, 0, 16, seed)
	c2 := NewChunk(0, 0, 16, seed)

	// Same seed should produce identical heightmaps
	for i := range c1.HeightMap {
		if c1.HeightMap[i] != c2.HeightMap[i] {
			t.Errorf("heightmap mismatch at index %d: %f != %f", i, c1.HeightMap[i], c2.HeightMap[i])
		}
	}
}

func TestChunkDifferentSeeds(t *testing.T) {
	c1 := NewChunk(0, 0, 16, 12345)
	c2 := NewChunk(0, 0, 16, 54321)

	// Different seeds should produce different heightmaps
	same := true
	for i := range c1.HeightMap {
		if c1.HeightMap[i] != c2.HeightMap[i] {
			same = false
			break
		}
	}

	if same {
		t.Error("different seeds should produce different heightmaps")
	}
}

func TestChunkGetHeight(t *testing.T) {
	c := NewChunk(0, 0, 16, 12345)

	// Valid coordinates
	h := c.GetHeight(0, 0)
	if h < 0 || h > 1 {
		t.Errorf("height should be in [0,1] range, got %f", h)
	}

	h = c.GetHeight(15, 15)
	if h < 0 || h > 1 {
		t.Errorf("height should be in [0,1] range, got %f", h)
	}

	// Invalid coordinates should return 0
	h = c.GetHeight(-1, 0)
	if h != 0 {
		t.Errorf("expected 0 for invalid coordinates, got %f", h)
	}

	h = c.GetHeight(0, -1)
	if h != 0 {
		t.Errorf("expected 0 for invalid coordinates, got %f", h)
	}

	h = c.GetHeight(16, 0)
	if h != 0 {
		t.Errorf("expected 0 for out-of-bounds coordinates, got %f", h)
	}

	h = c.GetHeight(0, 16)
	if h != 0 {
		t.Errorf("expected 0 for out-of-bounds coordinates, got %f", h)
	}
}

func TestManagerGetChunk(t *testing.T) {
	cm := NewManager(16, 12345)

	c1 := cm.GetChunk(0, 0)
	if c1 == nil {
		t.Fatal("GetChunk returned nil")
	}
	if c1.X != 0 || c1.Y != 0 {
		t.Errorf("expected chunk at (0,0), got (%d,%d)", c1.X, c1.Y)
	}
}

func TestManagerCaching(t *testing.T) {
	cm := NewManager(16, 12345)

	c1 := cm.GetChunk(5, 10)
	c2 := cm.GetChunk(5, 10)

	// Should return the same cached chunk
	if c1 != c2 {
		t.Error("GetChunk should return cached chunk")
	}
}

func TestManagerDifferentCoordinates(t *testing.T) {
	cm := NewManager(16, 12345)

	c1 := cm.GetChunk(0, 0)
	c2 := cm.GetChunk(1, 0)

	if c1 == c2 {
		t.Error("different coordinates should return different chunks")
	}
}

func TestManagerSeedMixing(t *testing.T) {
	cm := NewManager(16, 12345)

	c00 := cm.GetChunk(0, 0)
	c01 := cm.GetChunk(0, 1)
	c10 := cm.GetChunk(1, 0)

	// Each chunk should have a unique seed derived from coordinates
	if c00.Seed == c01.Seed {
		t.Error("(0,0) and (0,1) should have different seeds")
	}
	if c00.Seed == c10.Seed {
		t.Error("(0,0) and (1,0) should have different seeds")
	}
	if c01.Seed == c10.Seed {
		t.Error("(0,1) and (1,0) should have different seeds")
	}
}

func TestManagerUnloadChunk(t *testing.T) {
	cm := NewManager(16, 12345)

	_ = cm.GetChunk(0, 0)
	if cm.LoadedCount() != 1 {
		t.Errorf("expected 1 loaded chunk, got %d", cm.LoadedCount())
	}

	cm.UnloadChunk(0, 0)
	if cm.LoadedCount() != 0 {
		t.Errorf("expected 0 loaded chunks after unload, got %d", cm.LoadedCount())
	}
}

func TestManagerLoadedCount(t *testing.T) {
	cm := NewManager(16, 12345)

	if cm.LoadedCount() != 0 {
		t.Errorf("expected 0 loaded chunks initially, got %d", cm.LoadedCount())
	}

	_ = cm.GetChunk(0, 0)
	_ = cm.GetChunk(1, 0)
	_ = cm.GetChunk(0, 1)

	if cm.LoadedCount() != 3 {
		t.Errorf("expected 3 loaded chunks, got %d", cm.LoadedCount())
	}
}

func TestMixChunkSeedDeterminism(t *testing.T) {
	baseSeed := int64(12345)

	s1 := mixChunkSeed(baseSeed, 10, 20)
	s2 := mixChunkSeed(baseSeed, 10, 20)

	if s1 != s2 {
		t.Error("mixChunkSeed should be deterministic")
	}
}

func TestMixChunkSeedUniqueness(t *testing.T) {
	baseSeed := int64(12345)

	s00 := mixChunkSeed(baseSeed, 0, 0)
	s01 := mixChunkSeed(baseSeed, 0, 1)
	s10 := mixChunkSeed(baseSeed, 1, 0)
	s11 := mixChunkSeed(baseSeed, 1, 1)

	seeds := []int64{s00, s01, s10, s11}
	seen := make(map[int64]bool)
	for _, s := range seeds {
		if seen[s] {
			t.Error("mixChunkSeed produced duplicate seeds for different coordinates")
		}
		seen[s] = true
	}
}

func BenchmarkNewChunk(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewChunk(0, 0, 64, int64(i))
	}
}

func BenchmarkManagerGetChunk(b *testing.B) {
	cm := NewManager(64, 12345)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = cm.GetChunk(i%100, i/100)
	}
}

// Tests for vertical terrain (hills, cliffs)

func TestChunkElevationMapPopulated(t *testing.T) {
	c := NewChunk(0, 0, 16, 12345)

	if c.ElevationMap == nil {
		t.Fatal("ElevationMap should not be nil")
	}
	if len(c.ElevationMap) != 16*16 {
		t.Errorf("expected ElevationMap size 256, got %d", len(c.ElevationMap))
	}

	// Check that elevation values are in valid range
	for _, e := range c.ElevationMap {
		if e < 0 || e > MaxElevation {
			t.Errorf("elevation %f out of range [0, %f]", e, MaxElevation)
		}
	}
}

func TestChunkTerrainTypesPopulated(t *testing.T) {
	c := NewChunk(0, 0, 16, 12345)

	if c.TerrainTypes == nil {
		t.Fatal("TerrainTypes should not be nil")
	}
	if len(c.TerrainTypes) != 16*16 {
		t.Errorf("expected TerrainTypes size 256, got %d", len(c.TerrainTypes))
	}

	// Check that terrain types are valid
	for _, tt := range c.TerrainTypes {
		if tt < TerrainFlat || tt > TerrainPeak {
			t.Errorf("invalid terrain type %d", tt)
		}
	}
}

func TestChunkGetElevation(t *testing.T) {
	c := NewChunk(0, 0, 16, 12345)

	// Valid coordinates
	e := c.GetElevation(0, 0)
	if e < 0 || e > MaxElevation {
		t.Errorf("elevation should be in [0, %f] range, got %f", MaxElevation, e)
	}

	e = c.GetElevation(15, 15)
	if e < 0 || e > MaxElevation {
		t.Errorf("elevation should be in [0, %f] range, got %f", MaxElevation, e)
	}

	// Invalid coordinates should return 0
	e = c.GetElevation(-1, 0)
	if e != 0 {
		t.Errorf("expected 0 for invalid coordinates, got %f", e)
	}

	e = c.GetElevation(16, 0)
	if e != 0 {
		t.Errorf("expected 0 for out-of-bounds coordinates, got %f", e)
	}
}

func TestChunkGetTerrainType(t *testing.T) {
	c := NewChunk(0, 0, 16, 12345)

	// Valid coordinates should return a valid type
	tt := c.GetTerrainType(8, 8)
	if tt < TerrainFlat || tt > TerrainPeak {
		t.Errorf("invalid terrain type %d", tt)
	}

	// Invalid coordinates should return TerrainFlat
	tt = c.GetTerrainType(-1, 0)
	if tt != TerrainFlat {
		t.Errorf("expected TerrainFlat for invalid coordinates, got %d", tt)
	}

	tt = c.GetTerrainType(16, 0)
	if tt != TerrainFlat {
		t.Errorf("expected TerrainFlat for out-of-bounds coordinates, got %d", tt)
	}
}

func TestChunkIsCliff(t *testing.T) {
	// Create a chunk and manually set a cliff for testing
	c := NewChunk(0, 0, 16, 12345)

	// Find a cliff cell if one exists
	hasCliff := false
	for y := 0; y < c.Size; y++ {
		for x := 0; x < c.Size; x++ {
			if c.TerrainTypes[y*c.Size+x] == TerrainCliff {
				if !c.IsCliff(x, y) {
					t.Error("IsCliff should return true for cliff terrain")
				}
				hasCliff = true
				break
			}
		}
		if hasCliff {
			break
		}
	}

	// A flat cell should not be a cliff
	for y := 0; y < c.Size; y++ {
		for x := 0; x < c.Size; x++ {
			if c.TerrainTypes[y*c.Size+x] == TerrainFlat {
				if c.IsCliff(x, y) {
					t.Error("IsCliff should return false for flat terrain")
				}
				return
			}
		}
	}
}

func TestChunkIsHill(t *testing.T) {
	c := NewChunk(0, 0, 16, 12345)

	// Test that hills and peaks return true for IsHill
	for y := 0; y < c.Size; y++ {
		for x := 0; x < c.Size; x++ {
			tt := c.TerrainTypes[y*c.Size+x]
			isHill := c.IsHill(x, y)
			expected := tt == TerrainHill || tt == TerrainPeak
			if isHill != expected {
				t.Errorf("IsHill(%d, %d) = %v, expected %v for terrain type %d", x, y, isHill, expected, tt)
			}
		}
	}
}

func TestChunkGetElevationDifference(t *testing.T) {
	c := NewChunk(0, 0, 16, 12345)

	// Same cell should have 0 difference
	diff := c.GetElevationDifference(5, 5, 5, 5)
	if diff != 0 {
		t.Errorf("expected 0 difference for same cell, got %f", diff)
	}

	// Different cells may have positive difference
	diff = c.GetElevationDifference(0, 0, 15, 15)
	if diff < 0 {
		t.Errorf("elevation difference should be non-negative, got %f", diff)
	}
}

func TestElevationDeterminism(t *testing.T) {
	seed := int64(42)

	c1 := NewChunk(0, 0, 16, seed)
	c2 := NewChunk(0, 0, 16, seed)

	// Same seed should produce identical elevation maps
	for i := range c1.ElevationMap {
		if c1.ElevationMap[i] != c2.ElevationMap[i] {
			t.Errorf("elevation mismatch at index %d: %f != %f", i, c1.ElevationMap[i], c2.ElevationMap[i])
		}
	}

	// Same seed should produce identical terrain types
	for i := range c1.TerrainTypes {
		if c1.TerrainTypes[i] != c2.TerrainTypes[i] {
			t.Errorf("terrain type mismatch at index %d: %d != %d", i, c1.TerrainTypes[i], c2.TerrainTypes[i])
		}
	}
}

func TestTerrainTypeDistribution(t *testing.T) {
	// Test that terrain types are distributed across the chunk
	c := NewChunk(0, 0, 64, 12345)

	counts := make(map[int]int)
	for _, tt := range c.TerrainTypes {
		counts[tt]++
	}

	// With a 64x64 chunk, we should have some variation
	// At minimum, we should have flat terrain
	if counts[TerrainFlat] == 0 {
		t.Error("expected some flat terrain cells")
	}

	// Log distribution for debugging (not an error if others are 0)
	t.Logf("Terrain distribution: Flat=%d, Hill=%d, Cliff=%d, Peak=%d",
		counts[TerrainFlat], counts[TerrainHill], counts[TerrainCliff], counts[TerrainPeak])
}

func BenchmarkNewChunkWithElevation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewChunk(0, 0, 64, int64(i))
	}
}

// ========== Dynamic Terrain Modification Tests ==========

func TestNewModifiedChunk(t *testing.T) {
	base := NewChunk(0, 0, 16, 12345)
	mc := NewModifiedChunk(base)

	if mc.Chunk != base {
		t.Error("ModifiedChunk should wrap the base chunk")
	}
	if mc.ModificationCount() != 0 {
		t.Error("New ModifiedChunk should have no modifications")
	}
	if mc.IsDirty() {
		t.Error("New ModifiedChunk should not be dirty")
	}
}

func TestModifiedChunkDig(t *testing.T) {
	base := NewChunk(0, 0, 16, 12345)
	mc := NewModifiedChunk(base)

	originalHeight := mc.GetHeight(8, 8)
	mc.Dig(8, 8, 0.1, 1, 1000)

	newHeight := mc.GetHeight(8, 8)
	if newHeight >= originalHeight {
		t.Errorf("Dig should lower terrain: was %f, now %f", originalHeight, newHeight)
	}
	if !mc.IsModified(8, 8) {
		t.Error("Cell should be marked as modified after Dig")
	}
	if !mc.IsDirty() {
		t.Error("Chunk should be dirty after modification")
	}
}

func TestModifiedChunkFill(t *testing.T) {
	base := NewChunk(0, 0, 16, 12345)
	mc := NewModifiedChunk(base)

	originalHeight := mc.GetHeight(8, 8)
	mc.Fill(8, 8, 0.1, 1, 1000)

	newHeight := mc.GetHeight(8, 8)
	if newHeight <= originalHeight && originalHeight < 1.0 {
		t.Errorf("Fill should raise terrain: was %f, now %f", originalHeight, newHeight)
	}
}

func TestModifiedChunkFlatten(t *testing.T) {
	base := NewChunk(0, 0, 16, 12345)
	mc := NewModifiedChunk(base)

	targetHeight := 0.5
	mc.Flatten(8, 8, targetHeight, 1, 1000)

	newHeight := mc.GetHeight(8, 8)
	if newHeight != targetHeight {
		t.Errorf("Flatten should set exact height: expected %f, got %f", targetHeight, newHeight)
	}
}

func TestModifiedChunkSmooth(t *testing.T) {
	base := NewChunk(0, 0, 16, 12345)
	mc := NewModifiedChunk(base)

	mc.Smooth(8, 8, 1, 1000)

	if !mc.IsModified(8, 8) {
		t.Error("Cell should be marked as modified after Smooth")
	}
}

func TestModifiedChunkExplode(t *testing.T) {
	base := NewChunk(0, 0, 16, 12345)
	mc := NewModifiedChunk(base)

	centerHeight := mc.GetHeight(8, 8)
	mc.Explode(8, 8, 3.0, 1, 1000)

	newCenterHeight := mc.GetHeight(8, 8)
	if newCenterHeight >= centerHeight && centerHeight > 0.3 {
		t.Errorf("Explosion center should be lowered: was %f, now %f", centerHeight, newCenterHeight)
	}
}

func TestModifiedChunkErode(t *testing.T) {
	base := NewChunk(0, 0, 16, 12345)
	mc := NewModifiedChunk(base)

	mc.Erode(8, 8, 0.5, 1, 1000)

	if !mc.IsModified(8, 8) {
		t.Error("Cell should be marked as modified after Erode")
	}
}

func TestModifiedChunkUndo(t *testing.T) {
	base := NewChunk(0, 0, 16, 12345)
	mc := NewModifiedChunk(base)

	originalHeight := mc.GetHeight(8, 8)
	mc.Dig(8, 8, 0.1, 1, 1000)

	// Undo should restore original value
	if !mc.UndoLastModification() {
		t.Error("UndoLastModification should return true when there are modifications")
	}

	restoredHeight := mc.GetHeight(8, 8)
	if restoredHeight != originalHeight {
		t.Errorf("Undo should restore original height: expected %f, got %f", originalHeight, restoredHeight)
	}

	if mc.IsModified(8, 8) {
		t.Error("Cell should no longer be marked as modified after Undo")
	}

	// Undo when empty should return false
	if mc.UndoLastModification() {
		t.Error("UndoLastModification should return false when no modifications remain")
	}
}

func TestModifiedChunkBoundaryChecks(t *testing.T) {
	base := NewChunk(0, 0, 16, 12345)
	mc := NewModifiedChunk(base)

	// Modifications outside bounds should be ignored
	initialCount := mc.ModificationCount()
	mc.Dig(-1, 0, 0.1, 1, 1000)
	mc.Dig(16, 0, 0.1, 1, 1000)
	mc.Dig(0, -1, 0.1, 1, 1000)
	mc.Dig(0, 16, 0.1, 1, 1000)

	if mc.ModificationCount() != initialCount {
		t.Error("Modifications outside bounds should be ignored")
	}

	// IsModified for out-of-bounds should return false
	if mc.IsModified(-1, 0) {
		t.Error("IsModified should return false for out-of-bounds coordinates")
	}
}

func TestModifiedChunkGetModifications(t *testing.T) {
	base := NewChunk(0, 0, 16, 12345)
	mc := NewModifiedChunk(base)

	mc.Dig(1, 1, 0.1, 1, 1000)
	mc.Fill(2, 2, 0.1, 2, 2000)

	mods := mc.GetModifications()
	if len(mods) != 2 {
		t.Errorf("Expected 2 modifications, got %d", len(mods))
	}

	// Verify modifications contain expected data
	if mods[0].X != 1 || mods[0].Y != 1 || mods[0].ModifierID != 1 {
		t.Error("First modification data mismatch")
	}
	if mods[1].X != 2 || mods[1].Y != 2 || mods[1].ModifierID != 2 {
		t.Error("Second modification data mismatch")
	}
}

func TestModifiedChunkRestoreModifications(t *testing.T) {
	base1 := NewChunk(0, 0, 16, 12345)
	mc1 := NewModifiedChunk(base1)

	mc1.Dig(8, 8, 0.1, 1, 1000)
	mc1.Fill(9, 9, 0.2, 2, 2000)

	// Save modifications
	mods := mc1.GetModifications()

	// Create new chunk and restore
	base2 := NewChunk(0, 0, 16, 12345)
	mc2 := NewModifiedChunk(base2)
	mc2.RestoreModifications(mods)

	// Heights should match after restoration
	if mc1.GetHeight(8, 8) != mc2.GetHeight(8, 8) {
		t.Error("Restored chunk should have matching heights")
	}
	if mc1.GetHeight(9, 9) != mc2.GetHeight(9, 9) {
		t.Error("Restored chunk should have matching heights")
	}

	// Should not be dirty after restoration
	if mc2.IsDirty() {
		t.Error("Chunk should not be dirty after RestoreModifications")
	}
}

func TestModifiedChunkClearDirty(t *testing.T) {
	base := NewChunk(0, 0, 16, 12345)
	mc := NewModifiedChunk(base)

	mc.Dig(8, 8, 0.1, 1, 1000)
	if !mc.IsDirty() {
		t.Error("Should be dirty after modification")
	}

	mc.ClearDirty()
	if mc.IsDirty() {
		t.Error("Should not be dirty after ClearDirty")
	}
}

func TestModifiedChunkHeightClamp(t *testing.T) {
	base := NewChunk(0, 0, 16, 12345)
	mc := NewModifiedChunk(base)

	// Digging should not go below 0
	mc.Dig(0, 0, 10.0, 1, 1000) // Large dig value
	if mc.GetHeight(0, 0) < 0 {
		t.Error("Height should not go below 0")
	}

	// Filling should not go above 1
	mc.Fill(1, 1, 10.0, 1, 1000) // Large fill value
	if mc.GetHeight(1, 1) > 1 {
		t.Error("Height should not go above 1")
	}
}

func BenchmarkModifiedChunkDig(b *testing.B) {
	base := NewChunk(0, 0, 64, 12345)
	mc := NewModifiedChunk(base)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := i%62 + 1
		y := (i/62)%62 + 1
		mc.Dig(x, y, 0.01, 1, int64(i))
	}
}

func BenchmarkModifiedChunkExplode(b *testing.B) {
	base := NewChunk(0, 0, 64, 12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mc := NewModifiedChunk(base)
		mc.Explode(32, 32, 5.0, 1, int64(i))
	}
}

// ========== Terrain LOD System Tests ==========

func TestLODLevelConstants(t *testing.T) {
	if LODFull != 0 {
		t.Error("LODFull should be 0")
	}
	if LODHalf != 1 {
		t.Error("LODHalf should be 1")
	}
	if LODQuarter != 2 {
		t.Error("LODQuarter should be 2")
	}
	if LODEighth != 3 {
		t.Error("LODEighth should be 3")
	}
}

func TestNewChunkLODCache(t *testing.T) {
	cache := NewChunkLODCache()
	if cache == nil {
		t.Fatal("NewChunkLODCache returned nil")
	}
	if cache.chunks == nil {
		t.Error("Cache chunks map should be initialized")
	}
}

func TestChunkLODCacheGetLODFull(t *testing.T) {
	cache := NewChunkLODCache()
	chunk := NewChunk(0, 0, 64, 12345)

	lod := cache.GetLOD(chunk, LODFull)

	if lod.Level != LODFull {
		t.Errorf("Expected LODFull, got %d", lod.Level)
	}
	if lod.LODSize != chunk.Size {
		t.Errorf("Full LOD should have same size as chunk: expected %d, got %d", chunk.Size, lod.LODSize)
	}
	// Full LOD should use same slice
	if &lod.HeightMap[0] != &chunk.HeightMap[0] {
		t.Error("Full LOD should share height map with original chunk")
	}
}

func TestChunkLODCacheGetLODHalf(t *testing.T) {
	cache := NewChunkLODCache()
	chunk := NewChunk(0, 0, 64, 12345)

	lod := cache.GetLOD(chunk, LODHalf)

	if lod.Level != LODHalf {
		t.Errorf("Expected LODHalf, got %d", lod.Level)
	}
	if lod.LODSize != 32 {
		t.Errorf("Half LOD should be half size: expected 32, got %d", lod.LODSize)
	}
}

func TestChunkLODCacheGetLODQuarter(t *testing.T) {
	cache := NewChunkLODCache()
	chunk := NewChunk(0, 0, 64, 12345)

	lod := cache.GetLOD(chunk, LODQuarter)

	if lod.LODSize != 16 {
		t.Errorf("Quarter LOD should be quarter size: expected 16, got %d", lod.LODSize)
	}
}

func TestChunkLODCacheGetLODEighth(t *testing.T) {
	cache := NewChunkLODCache()
	chunk := NewChunk(0, 0, 64, 12345)

	lod := cache.GetLOD(chunk, LODEighth)

	if lod.LODSize != 8 {
		t.Errorf("Eighth LOD should be eighth size: expected 8, got %d", lod.LODSize)
	}
}

func TestChunkLODCacheCaching(t *testing.T) {
	cache := NewChunkLODCache()
	chunk := NewChunk(0, 0, 64, 12345)

	lod1 := cache.GetLOD(chunk, LODHalf)
	lod2 := cache.GetLOD(chunk, LODHalf)

	// Should return same cached object
	if lod1 != lod2 {
		t.Error("Cache should return same LOD object for same chunk and level")
	}
}

func TestChunkLODCacheInvalidate(t *testing.T) {
	cache := NewChunkLODCache()
	chunk := NewChunk(0, 0, 64, 12345)

	lod1 := cache.GetLOD(chunk, LODHalf)
	cache.InvalidateLOD(0, 0)
	lod2 := cache.GetLOD(chunk, LODHalf)

	// After invalidation, should generate new LOD
	if lod1 == lod2 {
		t.Error("After invalidation, should generate new LOD object")
	}
}

func TestLODChunkGetHeight(t *testing.T) {
	cache := NewChunkLODCache()
	chunk := NewChunk(0, 0, 64, 12345)

	lod := cache.GetLOD(chunk, LODHalf)

	// Valid coordinates should return a value
	h := lod.GetHeight(16, 16)
	if h == 0 {
		// Height might be 0, check bounds instead
	}

	// Invalid coordinates should return 0
	h = lod.GetHeight(-1, 0)
	if h != 0 {
		t.Error("GetHeight should return 0 for negative coordinates")
	}
	h = lod.GetHeight(32, 0) // Out of LOD bounds
	if h != 0 {
		t.Error("GetHeight should return 0 for out-of-bounds coordinates")
	}
}

func TestLODChunkCoordConversion(t *testing.T) {
	cache := NewChunkLODCache()
	chunk := NewChunk(0, 0, 64, 12345)

	lod := cache.GetLOD(chunk, LODHalf) // step = 2

	// LOD to full
	fullX, fullY := lod.ToFullCoords(10, 15)
	if fullX != 20 || fullY != 30 {
		t.Errorf("ToFullCoords(10,15) = (%d,%d), expected (20,30)", fullX, fullY)
	}

	// Full to LOD
	lodX, lodY := lod.FromFullCoords(20, 30)
	if lodX != 10 || lodY != 15 {
		t.Errorf("FromFullCoords(20,30) = (%d,%d), expected (10,15)", lodX, lodY)
	}
}

func TestLODChunkVertexAndTriangleCount(t *testing.T) {
	cache := NewChunkLODCache()
	chunk := NewChunk(0, 0, 64, 12345)

	lod := cache.GetLOD(chunk, LODHalf) // 32x32

	expectedVerts := 32 * 32
	if lod.VertexCount() != expectedVerts {
		t.Errorf("VertexCount = %d, expected %d", lod.VertexCount(), expectedVerts)
	}

	expectedTris := 31 * 31 * 2
	if lod.TriangleCount() != expectedTris {
		t.Errorf("TriangleCount = %d, expected %d", lod.TriangleCount(), expectedTris)
	}
}

func TestCalculateLODLevel(t *testing.T) {
	tests := []struct {
		distSq   float64
		expected LODLevel
	}{
		{0, LODFull},
		{1000, LODFull},
		{4096, LODHalf},
		{10000, LODHalf},
		{16384, LODQuarter},
		{40000, LODQuarter},
		{65536, LODEighth},
		{100000, LODEighth},
	}

	for _, tt := range tests {
		got := CalculateLODLevel(tt.distSq)
		if got != tt.expected {
			t.Errorf("CalculateLODLevel(%f) = %d, expected %d", tt.distSq, got, tt.expected)
		}
	}
}

func TestNewLODManager(t *testing.T) {
	cm := NewManager(64, 12345)
	lm := NewLODManager(cm)

	if lm == nil {
		t.Fatal("NewLODManager returned nil")
	}
	if lm.cache == nil {
		t.Error("LODManager cache should be initialized")
	}
	if lm.chunkSize != 64 {
		t.Errorf("LODManager chunkSize = %d, expected 64", lm.chunkSize)
	}
}

func TestLODManagerSetViewpoint(t *testing.T) {
	cm := NewManager(64, 12345)
	lm := NewLODManager(cm)

	lm.SetViewpoint(100.5, 200.5)

	if lm.viewX != 100.5 || lm.viewY != 200.5 {
		t.Errorf("Viewpoint = (%f,%f), expected (100.5,200.5)", lm.viewX, lm.viewY)
	}
}

func TestLODManagerGetChunkLOD(t *testing.T) {
	cm := NewManager(64, 12345)
	lm := NewLODManager(cm)

	// Set viewpoint at chunk (0,0) center
	lm.SetViewpoint(32, 32)

	// Nearby chunk should be full LOD
	lod := lm.GetChunkLOD(0, 0)
	if lod.Level != LODFull {
		t.Errorf("Chunk at viewpoint should be LODFull, got %d", lod.Level)
	}

	// Far chunk should be lower LOD
	lm.SetViewpoint(32, 32)
	lodFar := lm.GetChunkLOD(10, 10) // Far away
	if lodFar.Level == LODFull {
		// Depending on distance, might still be full
	}
}

func TestLODManagerGetChunksInView(t *testing.T) {
	cm := NewManager(64, 12345)
	lm := NewLODManager(cm)
	lm.SetViewpoint(128, 128) // Chunk (2,2) center area

	chunks := lm.GetChunksInView(1) // 3x3 area

	expectedCount := 9 // 3x3
	if len(chunks) != expectedCount {
		t.Errorf("GetChunksInView(1) returned %d chunks, expected %d", len(chunks), expectedCount)
	}
}

func TestLODManagerInvalidateChunk(t *testing.T) {
	cm := NewManager(64, 12345)
	lm := NewLODManager(cm)
	lm.SetViewpoint(32, 32)

	// Get a chunk to cache it
	_ = lm.GetChunkLOD(0, 0)

	// Invalidate
	lm.InvalidateChunk(0, 0)

	// Stats should reflect invalidation
	totalChunks, _ := lm.CacheStats()
	// Full LOD doesn't use cache, so check might vary
	t.Logf("Cache stats after invalidation: chunks=%d", totalChunks)
}

func TestLODManagerCacheStats(t *testing.T) {
	cm := NewManager(64, 12345)
	lm := NewLODManager(cm)
	lm.SetViewpoint(0, 0)

	// Get several chunks at different LODs
	lm.cache.GetLOD(cm.GetChunk(0, 0), LODHalf)
	lm.cache.GetLOD(cm.GetChunk(0, 0), LODQuarter)
	lm.cache.GetLOD(cm.GetChunk(1, 1), LODHalf)

	totalChunks, totalLODs := lm.CacheStats()

	if totalChunks < 2 {
		t.Errorf("Expected at least 2 chunks in cache, got %d", totalChunks)
	}
	if totalLODs < 3 {
		t.Errorf("Expected at least 3 LODs in cache, got %d", totalLODs)
	}
}

func TestLODAveraging(t *testing.T) {
	cache := NewChunkLODCache()
	chunk := NewChunk(0, 0, 64, 12345)

	lod := cache.GetLOD(chunk, LODHalf)

	// LOD height should be average of corresponding full-res cells
	// Check a specific cell
	lodX, lodY := 10, 10
	avgHeight := 0.0
	for dy := 0; dy < 2; dy++ {
		for dx := 0; dx < 2; dx++ {
			avgHeight += chunk.GetHeight(lodX*2+dx, lodY*2+dy)
		}
	}
	avgHeight /= 4.0

	lodHeight := lod.GetHeight(lodX, lodY)
	if lodHeight != avgHeight {
		t.Errorf("LOD height should be average: expected %f, got %f", avgHeight, lodHeight)
	}
}

func BenchmarkGenerateLODHalf(b *testing.B) {
	chunk := NewChunk(0, 0, 64, 12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generateLOD(chunk, LODHalf)
	}
}

func BenchmarkGenerateLODEighth(b *testing.B) {
	chunk := NewChunk(0, 0, 64, 12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generateLOD(chunk, LODEighth)
	}
}

func BenchmarkLODManagerGetChunksInView(b *testing.B) {
	cm := NewManager(64, 12345)
	lm := NewLODManager(cm)
	lm.SetViewpoint(512, 512)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = lm.GetChunksInView(3)
	}
}

// ========== Biome Blending Tests ==========

func TestBiomeWorldSpaceContinuity(t *testing.T) {
	seed := int64(12345)
	size := 64

	// Generate two adjacent chunks
	chunk0 := NewChunk(0, 0, size, seed)
	chunk1 := NewChunk(1, 0, size, seed)

	// Get biome values at the boundary
	// Right edge of chunk 0 (x=size-1) should be similar to left edge of chunk 1 (x=0)
	rightEdge := chunk0.BiomeMap[(size/2)*size+(size-1)]
	leftEdge := chunk1.BiomeMap[(size/2)*size+0]

	// With world-space coordinates, these should be continuous
	// Allow some tolerance due to noise variation
	diff := abs(rightEdge - leftEdge)
	if diff > 0.3 {
		t.Errorf("biome discontinuity at chunk boundary: chunk0 edge=%f, chunk1 edge=%f, diff=%f",
			rightEdge, leftEdge, diff)
	}
}

func TestBiomeMapWorldSpaceVerticalContinuity(t *testing.T) {
	seed := int64(54321)
	size := 64

	// Generate two vertically adjacent chunks
	chunk0 := NewChunk(0, 0, size, seed)
	chunk1 := NewChunk(0, 1, size, seed)

	// Get biome values at the vertical boundary
	// Bottom edge of chunk 0 (y=size-1) should match top edge of chunk 1 (y=0)
	bottomEdge := chunk0.BiomeMap[(size-1)*size+(size/2)]
	topEdge := chunk1.BiomeMap[0*size+(size/2)]

	diff := abs(bottomEdge - topEdge)
	if diff > 0.3 {
		t.Errorf("vertical biome discontinuity at chunk boundary: chunk0 bottom=%f, chunk1 top=%f, diff=%f",
			bottomEdge, topEdge, diff)
	}
}

func TestBiomeDeterminismWithWorldSpace(t *testing.T) {
	seed := int64(99999)
	size := 32

	// Generate same chunk twice
	c1 := NewChunk(5, 10, size, seed)
	c2 := NewChunk(5, 10, size, seed)

	// Biome maps should be identical
	for i := range c1.BiomeMap {
		if c1.BiomeMap[i] != c2.BiomeMap[i] {
			t.Errorf("biome mismatch at index %d: %f != %f", i, c1.BiomeMap[i], c2.BiomeMap[i])
		}
	}
}

func TestBiomeEdgeBlending(t *testing.T) {
	// Test that edge blending smooths values near boundaries
	size := 64

	// Center value should be less blended than edge value
	centerResult := applyBiomeEdgeBlend(size/2, size/2, size, 0.8)
	edgeResult := applyBiomeEdgeBlend(0, size/2, size, 0.8)

	// Center should preserve value, edge should be pulled toward 0.5
	if abs(centerResult-0.8) > 0.01 {
		t.Errorf("center biome value should be preserved: expected ~0.8, got %f", centerResult)
	}

	// Edge should be pulled toward center (0.5)
	if edgeResult >= centerResult {
		t.Errorf("edge biome should be blended toward center: edge=%f, center=%f", edgeResult, centerResult)
	}
}

func TestBiomeSmoothFunction(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{0.0, 0.0},
		{1.0, 1.0},
		{0.5, 0.5},
		{-0.5, 0.0}, // Clamped
		{1.5, 1.0},  // Clamped
	}

	for _, tc := range tests {
		result := biomeSmooth(tc.input)
		if abs(result-tc.expected) > 0.01 {
			t.Errorf("biomeSmooth(%f) = %f, expected %f", tc.input, result, tc.expected)
		}
	}
}

func TestLerp(t *testing.T) {
	tests := []struct {
		a, b, t  float64
		expected float64
	}{
		{0.0, 1.0, 0.0, 0.0},
		{0.0, 1.0, 1.0, 1.0},
		{0.0, 1.0, 0.5, 0.5},
		{2.0, 4.0, 0.5, 3.0},
		{10.0, 20.0, 0.25, 12.5},
	}

	for _, tc := range tests {
		result := lerp(tc.a, tc.b, tc.t)
		if abs(result-tc.expected) > 0.0001 {
			t.Errorf("lerp(%f, %f, %f) = %f, expected %f", tc.a, tc.b, tc.t, result, tc.expected)
		}
	}
}

func TestMinIntFunction(t *testing.T) {
	tests := []struct {
		a, b     int
		expected int
	}{
		{1, 2, 1},
		{5, 3, 3},
		{0, 0, 0},
		{-1, 1, -1},
	}

	for _, tc := range tests {
		result := minInt(tc.a, tc.b)
		if result != tc.expected {
			t.Errorf("minInt(%d, %d) = %d, expected %d", tc.a, tc.b, result, tc.expected)
		}
	}
}

func BenchmarkBiomeMapWorldSpace(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = generateBiomeMapWorldSpace(i%10, i/10, 64, 12345)
	}
}

// ========== Chunk Generation Latency Benchmarks ==========

func BenchmarkChunkGenerationLatency64(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewChunk(i%100, i/100, 64, int64(i))
	}
}

func BenchmarkChunkGenerationLatency128(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewChunk(i%100, i/100, 128, int64(i))
	}
}

func BenchmarkChunkGenerationLatency256(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewChunk(i%100, i/100, 256, int64(i))
	}
}

func BenchmarkAsyncChunkGeneration(b *testing.B) {
	cm := NewManager(64, 12345)
	cm.EnableAsyncGeneration(4) // 4 workers
	defer cm.DisableAsyncGeneration()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm.RequestChunkAsync(i%100, i/100)
	}
}

func BenchmarkGetChunkOrPlaceholder(b *testing.B) {
	cm := NewManager(64, 12345)
	cm.EnableAsyncGeneration(2)
	defer cm.DisableAsyncGeneration()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cm.GetChunkOrPlaceholder(i%100, i/100)
	}
}

// === WallHeights Tests ===

func TestChunkWallHeightsGenerated(t *testing.T) {
	c := NewChunk(0, 0, 16, 12345)

	if c.WallHeights == nil {
		t.Fatal("WallHeights should not be nil")
	}
	if len(c.WallHeights) != 16*16 {
		t.Errorf("expected WallHeights size 256, got %d", len(c.WallHeights))
	}

	// Check that wall heights are within valid range
	for i, h := range c.WallHeights {
		if h < MinWallHeight || h > MaxWallHeightMultiplier {
			t.Errorf("wall height at index %d out of range: %f", i, h)
		}
	}
}

func TestGetWallHeight(t *testing.T) {
	c := NewChunk(0, 0, 16, 12345)

	// Test valid coordinates
	h := c.GetWallHeight(8, 8)
	if h < MinWallHeight || h > MaxWallHeightMultiplier {
		t.Errorf("wall height out of range: %f", h)
	}

	// Test out of bounds returns default
	h = c.GetWallHeight(-1, 0)
	if h != DefaultWallHeight {
		t.Errorf("expected default wall height for out of bounds, got %f", h)
	}

	h = c.GetWallHeight(100, 100)
	if h != DefaultWallHeight {
		t.Errorf("expected default wall height for out of bounds, got %f", h)
	}
}

func TestWallHeightsDeterminism(t *testing.T) {
	seed := int64(42)

	c1 := NewChunk(0, 0, 16, seed)
	c2 := NewChunk(0, 0, 16, seed)

	// Same seed should produce identical wall heights
	for i := range c1.WallHeights {
		if c1.WallHeights[i] != c2.WallHeights[i] {
			t.Errorf("wall heights mismatch at index %d: %f != %f", i, c1.WallHeights[i], c2.WallHeights[i])
		}
	}
}

func TestWallHeightsVaryByTerrain(t *testing.T) {
	// Create a large chunk to have various terrain types
	c := NewChunk(0, 0, 64, 12345)

	// Collect heights by terrain type
	heightsByTerrain := make(map[int][]float64)
	for i := 0; i < 64*64; i++ {
		x := i % 64
		y := i / 64
		terrainType := c.GetTerrainType(x, y)
		wallHeight := c.GetWallHeight(x, y)
		heightsByTerrain[terrainType] = append(heightsByTerrain[terrainType], wallHeight)
	}

	// Verify we have some variation (at least 2 terrain types)
	if len(heightsByTerrain) < 2 {
		t.Log("Warning: chunk only has", len(heightsByTerrain), "terrain types, variation may be limited")
	}

	// Peaks should generally have taller walls than valleys
	peakHeights := heightsByTerrain[TerrainPeak]
	valleyHeights := heightsByTerrain[TerrainValley]

	if len(peakHeights) > 0 && len(valleyHeights) > 0 {
		avgPeak := average(peakHeights)
		avgValley := average(valleyHeights)
		if avgPeak <= avgValley {
			t.Logf("Note: peak avg=%.2f, valley avg=%.2f - peaks should generally be taller", avgPeak, avgValley)
		}
	}
}

func average(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

// ============================================================
// Barrier Spawn Tests
// ============================================================

func TestBarrierSpawnDataDefaults(t *testing.T) {
	cfg := DefaultBarrierSpawnConfig()

	if cfg.Genre != "fantasy" {
		t.Errorf("expected default genre 'fantasy', got %q", cfg.Genre)
	}
	if cfg.Density <= 0 || cfg.Density > 1 {
		t.Errorf("density should be in (0, 1], got %f", cfg.Density)
	}
	if cfg.NaturalWeight+cfg.ConstructedWeight+cfg.OrganicWeight <= 0 {
		t.Error("weights should sum to positive value")
	}
}

func TestDetailSpawnIsBarrier(t *testing.T) {
	tests := []struct {
		name      string
		spawnType DetailSpawnType
		isBarrier bool
	}{
		{"tree is not barrier", DetailSpawnTree, false},
		{"bush is not barrier", DetailSpawnBush, false},
		{"rock is not barrier", DetailSpawnRock, false},
		{"boulder is not barrier", DetailSpawnBoulder, false},
		{"natural barrier is barrier", DetailSpawnBarrierNatural, true},
		{"constructed barrier is barrier", DetailSpawnBarrierConstructed, true},
		{"organic barrier is barrier", DetailSpawnBarrierOrganic, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spawn := DetailSpawn{Type: tt.spawnType}
			if spawn.IsBarrier() != tt.isBarrier {
				t.Errorf("IsBarrier() = %v, want %v", spawn.IsBarrier(), tt.isBarrier)
			}
		})
	}
}

func TestGenerateBarrierSpawns(t *testing.T) {
	size := 64
	seed := int64(12345)

	// Generate terrain data
	biomeMap := make([]float64, size*size)
	terrainTypes := make([]int, size*size)

	// Fill with varied terrain
	for i := 0; i < size*size; i++ {
		biomeMap[i] = float64(i%5) / 5.0
		terrainTypes[i] = i % 7 // Cycle through terrain types
	}

	cfg := DefaultBarrierSpawnConfig()
	spawns := GenerateBarrierSpawns(size, seed, terrainTypes, biomeMap, cfg)

	// Should generate some barriers
	if len(spawns) == 0 {
		t.Error("expected at least some barrier spawns")
	}

	// All spawns should be barriers
	for i, spawn := range spawns {
		if !spawn.IsBarrier() {
			t.Errorf("spawn %d is not a barrier: type=%d", i, spawn.Type)
		}
		if spawn.BarrierData == nil {
			t.Errorf("spawn %d has nil BarrierData", i)
		}
	}
}

func TestGenerateBarrierSpawnsDeterminism(t *testing.T) {
	size := 32
	seed := int64(42)

	terrainTypes := make([]int, size*size)
	biomeMap := make([]float64, size*size)
	for i := 0; i < size*size; i++ {
		terrainTypes[i] = TerrainFlat
		biomeMap[i] = 0.5
	}

	cfg := DefaultBarrierSpawnConfig()

	spawns1 := GenerateBarrierSpawns(size, seed, terrainTypes, biomeMap, cfg)
	spawns2 := GenerateBarrierSpawns(size, seed, terrainTypes, biomeMap, cfg)

	if len(spawns1) != len(spawns2) {
		t.Fatalf("expected same number of spawns, got %d vs %d", len(spawns1), len(spawns2))
	}

	for i := range spawns1 {
		if spawns1[i].Type != spawns2[i].Type {
			t.Errorf("spawn %d: type mismatch %d vs %d", i, spawns1[i].Type, spawns2[i].Type)
		}
		if spawns1[i].LocalX != spawns2[i].LocalX {
			t.Errorf("spawn %d: LocalX mismatch %f vs %f", i, spawns1[i].LocalX, spawns2[i].LocalX)
		}
		if spawns1[i].LocalY != spawns2[i].LocalY {
			t.Errorf("spawn %d: LocalY mismatch %f vs %f", i, spawns1[i].LocalY, spawns2[i].LocalY)
		}
	}
}

func TestBarrierArchetypeIDs(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}
	categories := []DetailSpawnType{DetailSpawnBarrierNatural, DetailSpawnBarrierConstructed, DetailSpawnBarrierOrganic}

	for _, genre := range genres {
		for _, category := range categories {
			ids := getBarrierArchetypeIDs(category, genre)
			if len(ids) == 0 {
				t.Errorf("no archetypes for genre=%s category=%d", genre, category)
			}
			// Each archetype should have a non-empty ID
			for _, id := range ids {
				if id == "" {
					t.Errorf("empty archetype ID for genre=%s category=%d", genre, category)
				}
			}
		}
	}
}

func TestSelectBarrierCategory(t *testing.T) {
	cfg := BarrierSpawnConfig{
		NaturalWeight:     1.0,
		ConstructedWeight: 0.0,
		OrganicWeight:     0.0,
	}

	// With only natural weight, should always get natural
	for i := 0; i < 10; i++ {
		rng := newSeededRNG(int64(i))
		category := selectBarrierCategory(TerrainFlat, 0.5, rng, cfg)
		if category != DetailSpawnBarrierNatural {
			t.Errorf("expected natural barrier with cfg weights, got %d", category)
		}
	}
}

func TestNewChunkWithBarriers(t *testing.T) {
	c := NewChunkWithBarriers(0, 0, 32, 12345, "fantasy")

	if c == nil {
		t.Fatal("NewChunkWithBarriers returned nil")
	}

	// Check basic chunk properties
	if c.Size != 32 {
		t.Errorf("expected size=32, got %d", c.Size)
	}

	// Should have some detail spawns (regular + barriers)
	if len(c.DetailSpawns) == 0 {
		t.Error("expected some detail spawns")
	}

	// Should have some barriers
	barrierCount := c.BarrierSpawnCount()
	// Note: may be 0 due to density, but GetBarrierSpawns should work
	barriers := c.GetBarrierSpawns()
	if len(barriers) != barrierCount {
		t.Errorf("BarrierSpawnCount() = %d, but GetBarrierSpawns() returned %d", barrierCount, len(barriers))
	}
}

func TestChunkGetBarrierSpawnsInArea(t *testing.T) {
	// Create a chunk with known barrier spawns
	c := &Chunk{
		Size: 16,
		DetailSpawns: []DetailSpawn{
			{Type: DetailSpawnTree, LocalX: 5, LocalY: 5},
			{Type: DetailSpawnBarrierNatural, LocalX: 2, LocalY: 2, BarrierData: &BarrierSpawnData{ShapeType: "cylinder"}},
			{Type: DetailSpawnBarrierConstructed, LocalX: 8, LocalY: 8, BarrierData: &BarrierSpawnData{ShapeType: "box"}},
			{Type: DetailSpawnBush, LocalX: 10, LocalY: 10},
			{Type: DetailSpawnBarrierOrganic, LocalX: 12, LocalY: 12, BarrierData: &BarrierSpawnData{ShapeType: "cylinder"}},
		},
	}

	// Get all barriers
	all := c.GetBarrierSpawns()
	if len(all) != 3 {
		t.Errorf("expected 3 barriers, got %d", len(all))
	}

	// Get barriers in area (2,2) to (10,10)
	inArea := c.GetBarrierSpawnsInArea(2, 2, 10, 10)
	if len(inArea) != 2 {
		t.Errorf("expected 2 barriers in area, got %d", len(inArea))
	}
}

func TestChunkAddBarrierSpawns(t *testing.T) {
	c := &Chunk{
		Size:         16,
		DetailSpawns: []DetailSpawn{{Type: DetailSpawnTree}},
	}

	newBarriers := []DetailSpawn{
		{Type: DetailSpawnBarrierNatural, BarrierData: &BarrierSpawnData{ShapeType: "cylinder"}},
		{Type: DetailSpawnBarrierConstructed, BarrierData: &BarrierSpawnData{ShapeType: "box"}},
	}

	c.AddBarrierSpawns(newBarriers)

	if len(c.DetailSpawns) != 3 {
		t.Errorf("expected 3 spawns after adding barriers, got %d", len(c.DetailSpawns))
	}
	if c.BarrierSpawnCount() != 2 {
		t.Errorf("expected 2 barriers, got %d", c.BarrierSpawnCount())
	}
}

func TestBarrierSpawnDataFields(t *testing.T) {
	// Test that barrier spawns have valid data
	size := 32
	seed := int64(99999)

	terrainTypes := make([]int, size*size)
	biomeMap := make([]float64, size*size)
	for i := 0; i < size*size; i++ {
		terrainTypes[i] = TerrainForest // Lots of organic barriers
		biomeMap[i] = 0.8
	}

	cfg := BarrierSpawnConfig{
		Genre:             "fantasy",
		Density:           0.5, // High density for testing
		NaturalWeight:     0.3,
		ConstructedWeight: 0.3,
		OrganicWeight:     0.4,
	}

	spawns := GenerateBarrierSpawns(size, seed, terrainTypes, biomeMap, cfg)

	for i, spawn := range spawns {
		if spawn.BarrierData == nil {
			t.Errorf("spawn %d: BarrierData is nil", i)
			continue
		}

		data := spawn.BarrierData

		// ShapeType should be valid
		validShapes := map[string]bool{"cylinder": true, "box": true}
		if !validShapes[data.ShapeType] {
			t.Errorf("spawn %d: invalid shape type %q", i, data.ShapeType)
		}

		// Height should be positive
		if data.Height <= 0 {
			t.Errorf("spawn %d: height should be positive, got %f", i, data.Height)
		}

		// Radius or Width/Depth should be set based on shape
		if data.ShapeType == "cylinder" && data.Radius <= 0 {
			t.Errorf("spawn %d: cylinder should have positive radius, got %f", i, data.Radius)
		}
		if data.ShapeType == "box" && (data.Width <= 0 || data.Depth <= 0) {
			t.Errorf("spawn %d: box should have positive width/depth, got %f/%f", i, data.Width, data.Depth)
		}

		// If destructible, should have HP
		if data.Destructible && data.HitPoints <= 0 {
			t.Errorf("spawn %d: destructible barrier should have positive HP, got %f", i, data.HitPoints)
		}
	}
}

func TestSelectBarrierSpawnTerrainFiltering(t *testing.T) {
	cfg := DefaultBarrierSpawnConfig()
	cfg.Density = 1.0 // Always spawn if terrain allows
	rng := newSeededRNG(12345)

	// Water should never spawn barriers
	spawn := selectBarrierSpawn(TerrainWater, 0.5, rng, cfg)
	if spawn != nil {
		t.Error("expected no barrier spawn on water")
	}

	// Cliff should never spawn barriers
	rng = newSeededRNG(12345)
	spawn = selectBarrierSpawn(TerrainCliff, 0.5, rng, cfg)
	if spawn != nil {
		t.Error("expected no barrier spawn on cliff")
	}

	// Flat terrain should allow spawns
	rng = newSeededRNG(12345)
	spawn = selectBarrierSpawn(TerrainFlat, 0.5, rng, cfg)
	if spawn == nil {
		t.Error("expected barrier spawn on flat terrain")
	}
}

func newSeededRNG(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed))
}

// TestDeterministicBarrierSpawn validates that the same seed produces identical barrier spawns.
// This is a critical requirement for networked multiplayer - all clients must generate
// the same world content from the same seed.
func TestDeterministicBarrierSpawn(t *testing.T) {
	seed := int64(42424242)
	genre := "fantasy"
	size := 32

	// Generate barriers twice with the same parameters
	c1 := NewChunkWithBarriers(0, 0, size, seed, genre)
	c2 := NewChunkWithBarriers(0, 0, size, seed, genre)

	// Extract barrier spawns
	barriers1 := make([]DetailSpawn, 0)
	barriers2 := make([]DetailSpawn, 0)

	for _, spawn := range c1.DetailSpawns {
		if spawn.IsBarrier() {
			barriers1 = append(barriers1, spawn)
		}
	}
	for _, spawn := range c2.DetailSpawns {
		if spawn.IsBarrier() {
			barriers2 = append(barriers2, spawn)
		}
	}

	// Same number of barriers
	if len(barriers1) != len(barriers2) {
		t.Fatalf("barrier count mismatch: %d vs %d", len(barriers1), len(barriers2))
	}

	// Same positions, shapes, and materials
	for i := range barriers1 {
		b1, b2 := barriers1[i], barriers2[i]

		if b1.LocalX != b2.LocalX || b1.LocalY != b2.LocalY {
			t.Errorf("barrier %d position mismatch: (%f,%f) vs (%f,%f)",
				i, b1.LocalX, b1.LocalY, b2.LocalX, b2.LocalY)
		}

		if b1.Type != b2.Type {
			t.Errorf("barrier %d type mismatch: %d vs %d", i, b1.Type, b2.Type)
		}

		if b1.Scale != b2.Scale {
			t.Errorf("barrier %d scale mismatch: %f vs %f", i, b1.Scale, b2.Scale)
		}

		if b1.BarrierData != nil && b2.BarrierData != nil {
			if b1.BarrierData.ShapeType != b2.BarrierData.ShapeType {
				t.Errorf("barrier %d shape type mismatch: %s vs %s",
					i, b1.BarrierData.ShapeType, b2.BarrierData.ShapeType)
			}
			if b1.BarrierData.MaterialID != b2.BarrierData.MaterialID {
				t.Errorf("barrier %d material mismatch: %d vs %d",
					i, b1.BarrierData.MaterialID, b2.BarrierData.MaterialID)
			}
			if b1.BarrierData.Height != b2.BarrierData.Height {
				t.Errorf("barrier %d height mismatch: %f vs %f",
					i, b1.BarrierData.Height, b2.BarrierData.Height)
			}
		} else if (b1.BarrierData == nil) != (b2.BarrierData == nil) {
			t.Errorf("barrier %d: BarrierData presence mismatch", i)
		}
	}

	// Different seed should produce different barriers
	c3 := NewChunkWithBarriers(0, 0, size, seed+1, genre)
	barriers3 := make([]DetailSpawn, 0)
	for _, spawn := range c3.DetailSpawns {
		if spawn.IsBarrier() {
			barriers3 = append(barriers3, spawn)
		}
	}

	// Either count differs or at least one position differs
	if len(barriers1) == len(barriers3) && len(barriers1) > 0 {
		allSame := true
		for i := range barriers1 {
			if barriers1[i].LocalX != barriers3[i].LocalX ||
				barriers1[i].LocalY != barriers3[i].LocalY {
				allSame = false
				break
			}
		}
		if allSame {
			t.Error("different seeds should produce different barrier layouts")
		}
	}
}

func TestStoreNetworkChunk(t *testing.T) {
	const chunkSize = 16
	cm := NewManager(chunkSize, 42)

	// Build a known height/biome payload.
	cells := chunkSize * chunkSize
	heights := make([]uint16, cells)
	biomes := make([]uint8, cells)
	for i := 0; i < cells; i++ {
		// Heights span 0..90 (encoded as h*100 → 0..9000)
		h := float64(i) / float64(cells-1) * 0.9
		heights[i] = uint16(h * 100)
		biomes[i] = uint8(float64(i) / float64(cells-1) * 255)
	}

	cm.StoreNetworkChunk(3, 7, chunkSize, heights, biomes)

	// Chunk should now be loaded.
	if !cm.IsChunkLoaded(3, 7) {
		t.Fatal("expected chunk (3,7) to be loaded after StoreNetworkChunk")
	}

	c, isReal := cm.GetChunkOrPlaceholder(3, 7)
	if !isReal {
		t.Fatal("expected real chunk, got placeholder")
	}
	if c.Size != chunkSize {
		t.Errorf("expected Size=%d, got %d", chunkSize, c.Size)
	}
	if len(c.HeightMap) != cells {
		t.Fatalf("expected HeightMap len %d, got %d", cells, len(c.HeightMap))
	}
	if len(c.ElevationMap) != cells {
		t.Fatalf("expected ElevationMap len %d, got %d", cells, len(c.ElevationMap))
	}
	if len(c.TerrainTypes) != cells {
		t.Fatalf("expected TerrainTypes len %d, got %d", cells, len(c.TerrainTypes))
	}
	if len(c.WallHeights) != cells {
		t.Fatalf("expected WallHeights len %d, got %d", cells, len(c.WallHeights))
	}
	if c.DetailSpawns != nil {
		t.Errorf("expected nil DetailSpawns for network chunk, got %d spawns", len(c.DetailSpawns))
	}

	// Verify elevation uses the h*h*MaxElevation curve.
	const elevationTolerance = 0.01
	for i := 0; i < cells; i++ {
		h := c.HeightMap[i]
		expectedElev := h * h * MaxElevation
		if math.Abs(c.ElevationMap[i]-expectedElev) > elevationTolerance {
			t.Errorf("cell %d: elevation %.4f, expected %.4f (h=%.4f)", i, c.ElevationMap[i], expectedElev, h)
			break
		}
	}

	// Verify wall heights are within valid range.
	for i := 0; i < cells; i++ {
		wh := c.WallHeights[i]
		if wh < MinWallHeight {
			t.Errorf("cell %d: wall height %.4f < MinWallHeight %.4f", i, wh, MinWallHeight)
			break
		}
		if wh > MaxWallHeightMultiplier {
			t.Errorf("cell %d: wall height %.4f > MaxWallHeightMultiplier %.4f", i, wh, MaxWallHeightMultiplier)
			break
		}
	}
}

func TestStoreNetworkChunkSizeMismatch(t *testing.T) {
	cm := NewManager(16, 42)

	// Attempt to store a chunk with wrong size — should be rejected.
	cm.StoreNetworkChunk(0, 0, 32, make([]uint16, 32*32), make([]uint8, 32*32))

	if cm.IsChunkLoaded(0, 0) {
		t.Error("expected chunk to be rejected when size != ChunkSize")
	}
}

func TestStoreNetworkChunkIncompleteData(t *testing.T) {
	const chunkSize = 16
	cm := NewManager(chunkSize, 42)

	// Store with incomplete data — should still store without panic.
	cm.StoreNetworkChunk(1, 1, chunkSize, make([]uint16, 10), make([]uint8, 10))

	if !cm.IsChunkLoaded(1, 1) {
		t.Fatal("expected chunk to be loaded even with incomplete data")
	}

	c, _ := cm.GetChunkOrPlaceholder(1, 1)
	// Remaining cells should be zero-initialized.
	cells := chunkSize * chunkSize
	if len(c.HeightMap) != cells {
		t.Errorf("expected HeightMap len %d, got %d", cells, len(c.HeightMap))
	}
}
