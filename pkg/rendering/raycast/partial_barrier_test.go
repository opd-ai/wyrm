//go:build noebiten

// Package raycast provides the first-person raycasting renderer.
// This test file tests partial barrier rendering: alpha blending, gap patterns, and transparency.
package raycast

import (
	"image/color"
	"testing"
)

// getPixelColor is a test helper that returns a color.RGBA from the framebuffer.
func getPixelColor(r *Renderer, x, y int) color.RGBA {
	rr, g, b, a := r.GetPixel(x, y)
	return color.RGBA{R: rr, G: g, B: b, A: a}
}

// ============================================================
// Alpha Blending Tests
// ============================================================

func TestBlendPixelAlpha(t *testing.T) {
	r := NewRenderer(64, 64)
	if r == nil {
		t.Fatal("NewRenderer returned nil")
	}

	// Set a background color
	bgColor := color.RGBA{R: 100, G: 100, B: 100, A: 255}
	r.SetPixelColor(10, 10, bgColor)

	// Blend a semi-transparent foreground
	fgColor := color.RGBA{R: 200, G: 50, B: 50, A: 128}
	r.BlendPixelColor(10, 10, fgColor)

	// Read back the blended result
	result := getPixelColor(r, 10, 10)

	// Expected: blend of bg and fg at ~50% alpha
	// blend = (fg * alpha + bg * (255 - alpha)) / 255
	expectedR := uint8((int(fgColor.R)*int(fgColor.A) + int(bgColor.R)*(255-int(fgColor.A))) / 255)
	expectedG := uint8((int(fgColor.G)*int(fgColor.A) + int(bgColor.G)*(255-int(fgColor.A))) / 255)
	expectedB := uint8((int(fgColor.B)*int(fgColor.A) + int(bgColor.B)*(255-int(fgColor.A))) / 255)

	// Allow small tolerance for rounding
	tolerance := uint8(2)
	if absDiff(result.R, expectedR) > tolerance {
		t.Errorf("expected R=%d, got %d", expectedR, result.R)
	}
	if absDiff(result.G, expectedG) > tolerance {
		t.Errorf("expected G=%d, got %d", expectedG, result.G)
	}
	if absDiff(result.B, expectedB) > tolerance {
		t.Errorf("expected B=%d, got %d", expectedB, result.B)
	}
}

func TestBlendPixelFullyTransparent(t *testing.T) {
	r := NewRenderer(64, 64)
	if r == nil {
		t.Fatal("NewRenderer returned nil")
	}

	// Set a background color
	bgColor := color.RGBA{R: 100, G: 150, B: 200, A: 255}
	r.SetPixelColor(5, 5, bgColor)

	// Blend a fully transparent foreground
	fgColor := color.RGBA{R: 255, G: 0, B: 0, A: 0}
	r.BlendPixelColor(5, 5, fgColor)

	// Result should be unchanged (background)
	result := getPixelColor(r, 5, 5)
	if result.R != bgColor.R || result.G != bgColor.G || result.B != bgColor.B {
		t.Errorf("fully transparent blend should not change background, got %v", result)
	}
}

func TestBlendPixelFullyOpaque(t *testing.T) {
	r := NewRenderer(64, 64)
	if r == nil {
		t.Fatal("NewRenderer returned nil")
	}

	// Set a background color
	bgColor := color.RGBA{R: 100, G: 150, B: 200, A: 255}
	r.SetPixelColor(5, 5, bgColor)

	// Blend a fully opaque foreground
	fgColor := color.RGBA{R: 50, G: 60, B: 70, A: 255}
	r.BlendPixelColor(5, 5, fgColor)

	// Result should be the foreground
	result := getPixelColor(r, 5, 5)
	if result.R != fgColor.R || result.G != fgColor.G || result.B != fgColor.B {
		t.Errorf("fully opaque blend should replace background, got %v expected %v", result, fgColor)
	}
}

func TestBlendPixelBoundsChecking(t *testing.T) {
	r := NewRenderer(64, 64)
	if r == nil {
		t.Fatal("NewRenderer returned nil")
	}

	// These should not panic
	r.BlendPixelColor(-1, 0, color.RGBA{R: 255, G: 0, B: 0, A: 128})
	r.BlendPixelColor(0, -1, color.RGBA{R: 255, G: 0, B: 0, A: 128})
	r.BlendPixelColor(64, 0, color.RGBA{R: 255, G: 0, B: 0, A: 128})
	r.BlendPixelColor(0, 64, color.RGBA{R: 255, G: 0, B: 0, A: 128})
	r.BlendPixelColor(100, 100, color.RGBA{R: 255, G: 0, B: 0, A: 128})
}

// ============================================================
// Gap Pattern Tests
// ============================================================

func TestSemiOpaqueGapPattern(t *testing.T) {
	// Test the gap pattern produces consistent results
	gapCount := 0
	solidCount := 0

	// Sample across the entire texture space
	samples := 256
	for i := 0; i < samples; i++ {
		texX := float64(i%16) / 16.0
		texY := float64(i/16) / 16.0
		if isSemiOpaqueGap(texX, texY) {
			gapCount++
		} else {
			solidCount++
		}
	}

	// Should have both gaps and solid areas for a fence-like pattern
	if gapCount == 0 {
		t.Error("gap pattern should have some gaps")
	}
	if solidCount == 0 {
		t.Error("gap pattern should have some solid areas")
	}

	// Gap ratio should be roughly 25% (1 in 4 for the modulo 4 pattern)
	gapRatio := float64(gapCount) / float64(samples)
	if gapRatio < 0.1 || gapRatio > 0.5 {
		t.Errorf("gap ratio %f seems unreasonable (expected ~25%%)", gapRatio)
	}
}

func TestSemiOpaqueGapPatternDeterministic(t *testing.T) {
	// The same texture coordinates should always produce the same gap decision
	testCoords := []struct {
		texX, texY float64
	}{
		{0.0, 0.0},
		{0.5, 0.5},
		{0.125, 0.375},
		{0.99, 0.01},
	}

	for _, tc := range testCoords {
		result1 := isSemiOpaqueGap(tc.texX, tc.texY)
		result2 := isSemiOpaqueGap(tc.texX, tc.texY)
		result3 := isSemiOpaqueGap(tc.texX, tc.texY)

		if result1 != result2 || result2 != result3 {
			t.Errorf("gap pattern not deterministic at (%f, %f)", tc.texX, tc.texY)
		}
	}
}

func TestSemiOpaqueGapPatternSymmetry(t *testing.T) {
	// The pattern should be symmetric in certain ways
	// Test that the pattern creates a regular grid

	// Check that adjacent cells differ
	adjacentDiffers := 0
	for x := 0; x < 8; x++ {
		for y := 0; y < 8; y++ {
			texX := float64(x) / 8.0
			texY := float64(y) / 8.0
			texXNext := float64(x+1) / 8.0

			current := isSemiOpaqueGap(texX, texY)
			next := isSemiOpaqueGap(texXNext, texY)

			if current != next {
				adjacentDiffers++
			}
		}
	}

	// Should have some variation (not all same)
	if adjacentDiffers == 0 {
		t.Error("gap pattern should have some variation between adjacent cells")
	}
}

// ============================================================
// Transparency Rendering Tests
// ============================================================

func TestGetTransparencyForAllFlags(t *testing.T) {
	tests := []struct {
		name     string
		flags    CellFlags
		minAlpha float64
		maxAlpha float64
	}{
		{"no flags", 0, 1.0, 1.0},
		{"solid only", FlagSolid, 1.0, 1.0},
		{"passable", FlagPassable, 1.0, 1.0},
		{"transparent", FlagTransparent, 0.4, 0.6},
		{"semi-opaque", FlagSemiOpaque, 0.85, 0.95},
		{"climbable", FlagClimbable, 1.0, 1.0},
		{"destructible", FlagDestructible, 1.0, 1.0},
		{"transparent + solid", FlagTransparent | FlagSolid, 0.4, 0.6},
		{"semi-opaque + destructible", FlagSemiOpaque | FlagDestructible, 0.85, 0.95},
		{"all flags", FlagSolid | FlagPassable | FlagTransparent | FlagClimbable | FlagDestructible | FlagSemiOpaque, 0.4, 0.6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTransparencyForFlags(tt.flags)
			if result < tt.minAlpha || result > tt.maxAlpha {
				t.Errorf("transparency for %s: expected between %f and %f, got %f",
					tt.name, tt.minAlpha, tt.maxAlpha, result)
			}
		})
	}
}

func TestTransparentMapCellCreation(t *testing.T) {
	// Create a transparent barrier cell
	cell := MapCell{
		WallType:   1,
		WallHeight: 1.0,
		FloorH:     0.0,
		CeilH:      1.0,
		MaterialID: 1,
		Flags:      FlagSolid | FlagTransparent,
	}

	// Verify flags
	if cell.Flags&FlagTransparent == 0 {
		t.Error("cell should have FlagTransparent set")
	}
	if cell.Flags&FlagSolid == 0 {
		t.Error("cell should have FlagSolid set")
	}

	// Verify transparency value
	transparency := getTransparencyForFlags(cell.Flags)
	if transparency < 0.4 || transparency > 0.6 {
		t.Errorf("transparent cell should have ~50%% opacity, got %f", transparency)
	}
}

func TestSemiOpaqueMapCellCreation(t *testing.T) {
	// Create a semi-opaque barrier cell (fence, grate)
	cell := MapCell{
		WallType:   2,
		WallHeight: 0.8,
		FloorH:     0.0,
		CeilH:      0.8,
		MaterialID: 2,
		Flags:      FlagSolid | FlagSemiOpaque,
	}

	// Verify flags
	if cell.Flags&FlagSemiOpaque == 0 {
		t.Error("cell should have FlagSemiOpaque set")
	}

	// Verify opacity value (semi-opaque should be mostly solid)
	transparency := getTransparencyForFlags(cell.Flags)
	if transparency < 0.85 || transparency > 0.95 {
		t.Errorf("semi-opaque cell should have ~90%% opacity, got %f", transparency)
	}
}

// ============================================================
// Partial Barrier Integration Tests
// ============================================================

func TestPartialBarrierInWorldMap(t *testing.T) {
	r := NewRenderer(320, 240)
	if r == nil {
		t.Fatal("NewRenderer returned nil")
	}

	// Set up world map with transparent and semi-opaque cells
	cells := make([][]MapCell, 16)
	for i := range cells {
		cells[i] = make([]MapCell, 16)
		for j := range cells[i] {
			cells[i][j] = DefaultMapCell()
		}
	}

	// Add a transparent wall
	cells[5][5] = MapCell{
		WallType:   1,
		WallHeight: 1.0,
		Flags:      FlagSolid | FlagTransparent,
	}

	// Add a semi-opaque wall (fence)
	cells[5][6] = MapCell{
		WallType:   2,
		WallHeight: 0.6,
		Flags:      FlagSolid | FlagSemiOpaque,
	}

	// Add a border wall
	cells[0][5] = WallMapCell(1)

	r.SetWorldMapCells(cells)

	// Verify cells are set correctly
	if r.WorldMapCells == nil {
		t.Fatal("WorldMapCells should not be nil after SetWorldMapCells")
	}

	// Check transparent cell
	if r.WorldMapCells[5][5].Flags&FlagTransparent == 0 {
		t.Error("cell [5][5] should have FlagTransparent")
	}

	// Check semi-opaque cell
	if r.WorldMapCells[5][6].Flags&FlagSemiOpaque == 0 {
		t.Error("cell [5][6] should have FlagSemiOpaque")
	}
}

func TestPartialBarrierFlagCombinations(t *testing.T) {
	// Test various flag combinations that might occur with partial barriers
	testCases := []struct {
		name        string
		flags       CellFlags
		shouldBlock bool
		shouldBlend bool
		shouldGap   bool
	}{
		{"transparent glass", FlagSolid | FlagTransparent, true, true, false},
		{"fence", FlagSolid | FlagSemiOpaque, true, false, true},
		{"force field", FlagTransparent | FlagDestructible, false, true, false},
		{"climbable fence", FlagSolid | FlagSemiOpaque | FlagClimbable, true, false, true},
		{"broken glass", FlagSolid | FlagTransparent | FlagDestructible, true, true, false},
		{"tall grass", FlagPassable | FlagSemiOpaque, false, false, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Check blocking
			blocks := tc.flags&FlagSolid != 0
			if blocks != tc.shouldBlock {
				t.Errorf("%s: shouldBlock=%v, got blocking=%v", tc.name, tc.shouldBlock, blocks)
			}

			// Check blending (transparent flag)
			blends := tc.flags&FlagTransparent != 0
			if blends != tc.shouldBlend {
				t.Errorf("%s: shouldBlend=%v, got blends=%v", tc.name, tc.shouldBlend, blends)
			}

			// Check gap pattern (semi-opaque flag)
			gaps := tc.flags&FlagSemiOpaque != 0
			if gaps != tc.shouldGap {
				t.Errorf("%s: shouldGap=%v, got gaps=%v", tc.name, tc.shouldGap, gaps)
			}
		})
	}
}

// ============================================================
// Barrier Rendering Quality Tests
// ============================================================

func TestAlphaBlendingColorAccuracy(t *testing.T) {
	r := NewRenderer(32, 32)
	if r == nil {
		t.Fatal("NewRenderer returned nil")
	}

	// Test various alpha values
	alphaTests := []uint8{0, 64, 128, 192, 255}
	bg := color.RGBA{R: 100, G: 100, B: 100, A: 255}
	fg := color.RGBA{R: 200, G: 50, B: 50, A: 255}

	for i, alpha := range alphaTests {
		x, y := i, 0
		r.SetPixelColor(x, y, bg)

		blendColor := color.RGBA{R: fg.R, G: fg.G, B: fg.B, A: alpha}
		r.BlendPixelColor(x, y, blendColor)

		result := getPixelColor(r, x, y)

		// Calculate expected result
		a := float64(alpha) / 255.0
		expectedR := uint8(float64(fg.R)*a + float64(bg.R)*(1-a))
		expectedG := uint8(float64(fg.G)*a + float64(bg.G)*(1-a))
		expectedB := uint8(float64(fg.B)*a + float64(bg.B)*(1-a))

		tolerance := uint8(2)
		if absDiff(result.R, expectedR) > tolerance ||
			absDiff(result.G, expectedG) > tolerance ||
			absDiff(result.B, expectedB) > tolerance {
			t.Errorf("alpha=%d: expected (%d,%d,%d), got (%d,%d,%d)",
				alpha, expectedR, expectedG, expectedB, result.R, result.G, result.B)
		}
	}
}

func TestGapPatternVisualConsistency(t *testing.T) {
	// Verify the gap pattern creates a visually reasonable distribution
	// by checking that gaps don't cluster too much

	const gridSize = 32
	gapMap := make([][]bool, gridSize)
	for i := range gapMap {
		gapMap[i] = make([]bool, gridSize)
	}

	// Fill the gap map
	for x := 0; x < gridSize; x++ {
		for y := 0; y < gridSize; y++ {
			texX := float64(x) / float64(gridSize)
			texY := float64(y) / float64(gridSize)
			gapMap[x][y] = isSemiOpaqueGap(texX, texY)
		}
	}

	// Check for reasonable gap distribution
	totalGaps := 0
	for x := 0; x < gridSize; x++ {
		for y := 0; y < gridSize; y++ {
			if gapMap[x][y] {
				totalGaps++
			}
		}
	}

	gapPercentage := float64(totalGaps) / float64(gridSize*gridSize) * 100
	if gapPercentage < 10 || gapPercentage > 50 {
		t.Errorf("gap percentage %f%% seems unreasonable for a fence pattern", gapPercentage)
	}
}

// ============================================================
// Benchmark Tests
// ============================================================

func BenchmarkAlphaBlend(b *testing.B) {
	r := NewRenderer(64, 64)
	bg := color.RGBA{R: 100, G: 100, B: 100, A: 255}
	fg := color.RGBA{R: 200, G: 50, B: 50, A: 128}

	// Pre-set background
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			r.SetPixelColor(x, y, bg)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x, y := i%64, (i/64)%64
		r.BlendPixelColor(x, y, fg)
	}
}

func BenchmarkSemiOpaqueGapCheck(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		texX := float64(i%100) / 100.0
		texY := float64(i/100) / 100.0
		_ = isSemiOpaqueGap(texX, texY)
	}
}

func BenchmarkTransparencyForFlags(b *testing.B) {
	flags := []CellFlags{0, FlagSolid, FlagTransparent, FlagSemiOpaque, FlagSolid | FlagTransparent}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getTransparencyForFlags(flags[i%len(flags)])
	}
}

// ============================================================
// Helper Functions
// ============================================================

func absDiff(a, b uint8) uint8 {
	if a > b {
		return a - b
	}
	return b - a
}
