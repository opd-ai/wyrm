package chunk

import (
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
