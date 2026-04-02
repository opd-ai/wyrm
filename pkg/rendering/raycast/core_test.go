//go:build noebiten

// Package raycast provides the first-person raycasting renderer.
// This test file tests core.go functions without requiring Ebiten.
package raycast

import (
	"image/color"
	"math"
	"testing"
)

func TestNewRendererCore(t *testing.T) {
	r := NewRenderer(1280, 720)
	if r == nil {
		t.Fatal("NewRenderer returned nil")
	}
	if r.Width != 1280 {
		t.Errorf("expected Width=1280, got %d", r.Width)
	}
	if r.Height != 720 {
		t.Errorf("expected Height=720, got %d", r.Height)
	}
	if r.PlayerX != DefaultPlayerX {
		t.Errorf("expected PlayerX=%f, got %f", DefaultPlayerX, r.PlayerX)
	}
	if r.PlayerY != DefaultPlayerY {
		t.Errorf("expected PlayerY=%f, got %f", DefaultPlayerY, r.PlayerY)
	}
	if r.FOV != DefaultFOV {
		t.Errorf("expected FOV=%f, got %f", DefaultFOV, r.FOV)
	}
	if r.WorldMap == nil {
		t.Error("WorldMap should not be nil")
	}
	if len(r.WorldMap) != DefaultMapSize {
		t.Errorf("expected WorldMap size %d, got %d", DefaultMapSize, len(r.WorldMap))
	}
}

func TestCalculateDeltaDist(t *testing.T) {
	tests := []struct {
		name         string
		rayDirX      float64
		rayDirY      float64
		expectLargeX bool
		expectLargeY bool
	}{
		{"right", 1.0, 0.0, false, true},
		{"left", -1.0, 0.0, false, true},
		{"up", 0.0, 1.0, true, false},
		{"down", 0.0, -1.0, true, false},
		{"diagonal", 0.707, 0.707, false, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dx, dy := calculateDeltaDist(tc.rayDirX, tc.rayDirY)
			if tc.expectLargeX && dx < 1e20 {
				t.Errorf("expected large deltaDistX for %s, got %f", tc.name, dx)
			}
			if tc.expectLargeY && dy < 1e20 {
				t.Errorf("expected large deltaDistY for %s, got %f", tc.name, dy)
			}
			if !tc.expectLargeX && dx <= 0 {
				t.Errorf("deltaDistX should be positive for %s, got %f", tc.name, dx)
			}
			if !tc.expectLargeY && dy <= 0 {
				t.Errorf("deltaDistY should be positive for %s, got %f", tc.name, dy)
			}
		})
	}
}

func TestCalculateSideDist(t *testing.T) {
	tests := []struct {
		name      string
		rayDirX   float64
		rayDirY   float64
		expectStX int
		expectStY int
	}{
		{"right_down", 1.0, 1.0, 1, 1},
		{"left_up", -1.0, -1.0, -1, -1},
		{"right_up", 1.0, -1.0, 1, -1},
		{"left_down", -1.0, 1.0, -1, 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			deltaDistX, deltaDistY := calculateDeltaDist(tc.rayDirX, tc.rayDirY)
			_, _, stepX, stepY := calculateSideDist(8.5, 8.5, 8, 8, tc.rayDirX, tc.rayDirY, deltaDistX, deltaDistY)
			if stepX != tc.expectStX {
				t.Errorf("expected stepX=%d, got %d", tc.expectStX, stepX)
			}
			if stepY != tc.expectStY {
				t.Errorf("expected stepY=%d, got %d", tc.expectStY, stepY)
			}
		})
	}
}

func TestCastRayKnownConfigurations(t *testing.T) {
	r := NewRenderer(640, 480)

	// Test 1: Player in clear area, ray toward boundary
	r.SetPlayerPos(12.0, 4.0, 0)
	dist, wallType := r.castRay(1.0, 0.0) // facing right
	if dist < 2.0 || dist > 4.0 {
		t.Errorf("expected distance ~3 to right boundary, got %f", dist)
	}
	if wallType != 1 {
		t.Errorf("expected boundary wallType=1, got %d", wallType)
	}

	// Test 2: Player facing interior wall type 2 at [4][4]
	r.SetPlayerPos(2.0, 4.5, 0)
	dist2, wallType2 := r.castRay(1.0, 0.0)
	if dist2 < 1.0 || dist2 > 3.0 {
		t.Errorf("expected distance ~2 to interior wall, got %f", dist2)
	}
	if wallType2 != 2 {
		t.Errorf("expected interior wallType=2, got %d", wallType2)
	}

	// Test 3: Player facing interior wall type 3 at [8][8]
	r.SetPlayerPos(6.0, 8.5, 0)
	dist3, wallType3 := r.castRay(1.0, 0.0)
	if dist3 < 1.0 || dist3 > 3.0 {
		t.Errorf("expected distance ~2 to interior wall, got %f", dist3)
	}
	if wallType3 != 3 {
		t.Errorf("expected interior wallType=3, got %d", wallType3)
	}
}

func TestCastRayParallelWalls(t *testing.T) {
	r := NewRenderer(640, 480)

	// Ray parallel to Y axis (rayDirX = 0)
	r.SetPlayerPos(8.0, 8.0, 0)
	dist, wallType := r.castRay(0.0, 1.0)
	if dist <= 0 || dist > MaxRayDistance {
		t.Errorf("parallel ray should hit wall, got distance %f", dist)
	}
	if wallType == 0 {
		t.Error("parallel ray should hit a wall")
	}

	// Ray parallel to X axis (rayDirY = 0)
	dist2, wallType2 := r.castRay(1.0, 0.0)
	if dist2 <= 0 || dist2 > MaxRayDistance {
		t.Errorf("parallel ray should hit wall, got distance %f", dist2)
	}
	if wallType2 == 0 {
		t.Error("parallel ray should hit a wall")
	}
}

func TestCastRayCorners(t *testing.T) {
	r := NewRenderer(640, 480)

	// Position near corner of interior walls
	r.SetPlayerPos(3.0, 3.5, 0)
	// Diagonal ray toward corner at [4][4]
	dist, wallType := r.castRay(0.707, 0.707)
	if dist <= 0 || dist > MaxRayDistance {
		t.Errorf("diagonal ray should hit wall, got distance %f", dist)
	}
	if wallType == 0 {
		t.Error("diagonal ray should hit a wall")
	}
}

func TestCalculateWallDistanceEdgeCases(t *testing.T) {
	r := NewRenderer(640, 480)

	// Test side == 0 (X-axis hit)
	dist0, wt0 := r.calculateWallDistance(0, 5.0, 10.0, 1.0, 1.0, 4, 4)
	if dist0 != 4.0 { // sideDistX - deltaDistX = 5.0 - 1.0
		t.Errorf("expected distance 4.0 for side=0, got %f", dist0)
	}
	if wt0 != 2 { // wall type at [4][4] is 2
		t.Errorf("expected wallType=2, got %d", wt0)
	}

	// Test side == 1 (Y-axis hit)
	dist1, wt1 := r.calculateWallDistance(1, 10.0, 5.0, 1.0, 1.0, 4, 5)
	if dist1 != 4.0 { // sideDistY - deltaDistY = 5.0 - 1.0
		t.Errorf("expected distance 4.0 for side=1, got %f", dist1)
	}
	if wt1 != 2 { // wall type at [4][5] is 2
		t.Errorf("expected wallType=2, got %d", wt1)
	}

	// Test invalid position
	distInv, wtInv := r.calculateWallDistance(0, 5.0, 10.0, 1.0, 1.0, -1, -1)
	if wtInv != 0 {
		t.Errorf("expected wallType=0 for invalid position, got %d", wtInv)
	}
	if distInv != 4.0 { // distance calculation still works
		t.Errorf("distance should still be calculated, got %f", distInv)
	}
}

func TestIsValidMapPosition(t *testing.T) {
	r := NewRenderer(640, 480)

	tests := []struct {
		x, y   int
		expect bool
	}{
		{0, 0, true},
		{15, 15, true},
		{8, 8, true},
		{-1, 0, false},
		{0, -1, false},
		{16, 0, false},
		{0, 16, false},
		{-1, -1, false},
		{100, 100, false},
	}

	for _, tc := range tests {
		result := r.isValidMapPosition(tc.x, tc.y)
		if result != tc.expect {
			t.Errorf("isValidMapPosition(%d, %d) = %v, expected %v", tc.x, tc.y, result, tc.expect)
		}
	}
}

func TestHeightToWallType(t *testing.T) {
	threshold := 0.3

	tests := []struct {
		height   float64
		expected int
	}{
		{0.0, 0},  // Below threshold
		{0.3, 0},  // At threshold
		{0.31, 1}, // Just above threshold (low wall)
		{0.5, 1},  // Low wall
		{0.61, 2}, // Medium wall
		{0.7, 2},  // Medium wall
		{0.81, 3}, // High wall
		{1.0, 3},  // Maximum height
	}

	for _, tc := range tests {
		result := heightToWallType(tc.height, threshold)
		if result != tc.expected {
			t.Errorf("heightToWallType(%f, %f) = %d, expected %d", tc.height, threshold, result, tc.expected)
		}
	}
}

func TestSetWorldMap(t *testing.T) {
	r := NewRenderer(640, 480)

	heightMap := []float64{
		0.0, 0.5, 0.7, 0.9,
		0.2, 0.0, 0.65, 0.85,
		0.4, 0.6, 0.0, 0.0,
		0.8, 0.9, 0.3, 0.0,
	}

	r.SetWorldMap(heightMap, 4, 0.3)

	if len(r.WorldMap) != 4 {
		t.Errorf("expected WorldMap size 4, got %d", len(r.WorldMap))
	}

	// Check specific conversions
	expectedWalls := map[[2]int]int{
		{0, 0}: 0, // 0.0 <= 0.3
		{0, 1}: 1, // 0.5 > 0.3, <= 0.6
		{0, 2}: 2, // 0.7 > 0.6, <= 0.8
		{0, 3}: 3, // 0.9 > 0.8
	}

	for coord, expected := range expectedWalls {
		actual := r.WorldMap[coord[0]][coord[1]]
		if actual != expected {
			t.Errorf("WorldMap[%d][%d] = %d, expected %d", coord[0], coord[1], actual, expected)
		}
	}
}

func TestSetWorldMapInvalidSize(t *testing.T) {
	r := NewRenderer(640, 480)
	originalLen := len(r.WorldMap)

	r.SetWorldMap(nil, 0, 0.3)
	if len(r.WorldMap) != originalLen {
		t.Error("WorldMap should not change with invalid size")
	}

	r.SetWorldMap(nil, -1, 0.3)
	if len(r.WorldMap) != originalLen {
		t.Error("WorldMap should not change with negative size")
	}
}

func TestSetWorldMapDirect(t *testing.T) {
	r := NewRenderer(640, 480)

	customMap := [][]int{
		{1, 1, 1, 1, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 2, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 1, 1, 1, 1},
	}

	r.SetWorldMapDirect(customMap)

	if len(r.WorldMap) != 5 {
		t.Errorf("expected WorldMap size 5, got %d", len(r.WorldMap))
	}
	if r.WorldMap[2][2] != 2 {
		t.Errorf("expected WorldMap[2][2]=2, got %d", r.WorldMap[2][2])
	}
}

func TestGetWallColorDistanceFog(t *testing.T) {
	r := NewRenderer(640, 480)

	// Test fog progression
	colors := make([]int, 5)
	distances := []float64{1.0, 4.0, 8.0, 12.0, 16.0}

	for i, d := range distances {
		c := r.getWallColor(1, d)
		colors[i] = int(c.R) + int(c.G) + int(c.B)
	}

	// Each subsequent color should be darker or equal
	for i := 1; i < len(colors); i++ {
		if colors[i] > colors[i-1] {
			t.Errorf("brightness should decrease with distance: %d at d=%f > %d at d=%f",
				colors[i], distances[i], colors[i-1], distances[i-1])
		}
	}
}

func TestGetWallColorMinFog(t *testing.T) {
	r := NewRenderer(640, 480)

	// Very far wall should still have minimum brightness
	c := r.getWallColor(1, 100.0)
	brightness := int(c.R) + int(c.G) + int(c.B)
	if brightness == 0 {
		t.Error("wall at extreme distance should have minimum brightness")
	}
}

func TestApplyDistanceFog(t *testing.T) {
	c := applyDistanceFog(color.RGBA{R: 100, G: 100, B: 100, A: 255}, 0.0)
	if c.R != 100 || c.G != 100 || c.B != 100 {
		t.Error("fog at distance 0 should not darken")
	}

	c2 := applyDistanceFog(color.RGBA{R: 100, G: 100, B: 100, A: 255}, FogDistance)
	if c2.R >= 100 || c2.G >= 100 || c2.B >= 100 {
		t.Error("fog at max distance should darken")
	}

	if c2.A != 255 {
		t.Error("alpha should always be 255")
	}
}

func TestSetGenreRegeneration(t *testing.T) {
	r := NewRendererWithGenre(640, 480, "fantasy", 12345)
	if r.Genre != "fantasy" {
		t.Errorf("expected genre fantasy, got %s", r.Genre)
	}

	// Same genre and seed should not regenerate
	tex := r.WallTextures[0]
	r.SetGenre("fantasy", 12345)
	if r.WallTextures[0] != tex {
		t.Error("same genre+seed should not regenerate textures")
	}

	// Different genre should regenerate
	r.SetGenre("cyberpunk", 12345)
	if r.Genre != "cyberpunk" {
		t.Errorf("expected genre cyberpunk, got %s", r.Genre)
	}
}

func TestTextureSampling(t *testing.T) {
	r := NewRendererWithGenre(640, 480, "fantasy", 12345)

	// Test wall texture sampling
	c := r.GetWallTextureColor(1, 0.5, 0.5, 5.0)
	if c.A != 255 {
		t.Errorf("wall texture alpha should be 255, got %d", c.A)
	}

	// Test floor texture sampling
	fc := r.GetFloorTextureColor(0.5, 0.5, 5.0)
	if fc.A != 255 {
		t.Errorf("floor texture alpha should be 255, got %d", fc.A)
	}

	// Test ceiling texture sampling
	cc := r.GetCeilingTextureColor(0.5, 0.5, 5.0)
	if cc.A != 255 {
		t.Errorf("ceiling texture alpha should be 255, got %d", cc.A)
	}
}

func TestTextureWrapping(t *testing.T) {
	r := NewRendererWithGenre(640, 480, "fantasy", 12345)

	// Test coordinates wrapping (negative and > 1.0)
	c1 := r.GetWallTextureColor(1, 0.5, 0.5, 5.0)
	c2 := r.GetWallTextureColor(1, 1.5, 1.5, 5.0) // Should wrap to 0.5, 0.5
	if c1 != c2 {
		t.Errorf("texture coordinates should wrap: %v != %v", c1, c2)
	}

	// Negative coordinates should also wrap correctly
	c3 := r.GetWallTextureColor(1, -0.5, -0.5, 5.0)
	if c3.A != 255 {
		t.Error("negative texture coordinates should produce valid color")
	}
}

func TestInvalidWallType(t *testing.T) {
	r := NewRendererWithGenre(640, 480, "fantasy", 12345)

	// Invalid wall types should return fallback color
	c := r.GetWallTextureColor(100, 0.5, 0.5, 5.0)
	if c.A != 255 {
		t.Error("invalid wall type should return valid fallback color")
	}

	c2 := r.GetWallTextureColor(-1, 0.5, 0.5, 5.0)
	if c2.A != 255 {
		t.Error("negative wall type should return valid fallback color")
	}
}

func TestDeterministicTextureGeneration(t *testing.T) {
	r1 := NewRendererWithGenre(640, 480, "fantasy", 42)
	r2 := NewRendererWithGenre(640, 480, "fantasy", 42)

	// Same seed should produce identical textures
	for wallType := 0; wallType < 4; wallType++ {
		for i := 0; i < 50; i++ {
			texX := float64(i) / 50.0
			texY := float64(i) / 50.0
			c1 := r1.GetWallTextureColor(wallType, texX, texY, 3.0)
			c2 := r2.GetWallTextureColor(wallType, texX, texY, 3.0)
			if c1 != c2 {
				t.Errorf("determinism failed: wallType=%d at (%f,%f): %v != %v",
					wallType, texX, texY, c1, c2)
				return
			}
		}
	}
}

func TestPerformDDAMaxSteps(t *testing.T) {
	r := NewRenderer(640, 480)

	// Clear the world to test max steps (both legacy and enhanced maps)
	for i := range r.WorldMap {
		for j := range r.WorldMap[i] {
			r.WorldMap[i][j] = 0
		}
	}
	for i := range r.WorldMapCells {
		for j := range r.WorldMapCells[i] {
			r.WorldMapCells[i][j] = DefaultMapCell()
		}
	}

	r.SetPlayerPos(8.0, 8.0, 0)
	dist, wallType := r.castRay(1.0, 0.0)

	// Should hit max distance when no walls
	if dist != MaxRayDistance {
		t.Errorf("expected MaxRayDistance when no walls, got %f", dist)
	}
	if wallType != 0 {
		t.Errorf("expected wallType=0 when no hit, got %d", wallType)
	}
}

func BenchmarkCastRayCore(b *testing.B) {
	r := NewRenderer(1280, 720)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.castRay(1.0, 0.0)
	}
}

func BenchmarkCalculateDeltaDist(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = calculateDeltaDist(0.707, 0.707)
	}
}

func BenchmarkGetWallColorCore(b *testing.B) {
	r := NewRenderer(1280, 720)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.getWallColor(1, 5.0)
	}
}

func BenchmarkApplyDistanceFog(b *testing.B) {
	c := color.RGBA{R: 128, G: 64, B: 192, A: 255}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = applyDistanceFog(c, 8.0)
	}
}

func BenchmarkIsValidMapPosition(b *testing.B) {
	r := NewRenderer(1280, 720)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.isValidMapPosition(8, 8)
	}
}

func BenchmarkHeightToWallType(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = heightToWallType(0.65, 0.3)
	}
}

func BenchmarkGetWallTextureColorCore(b *testing.B) {
	r := NewRendererWithGenre(1280, 720, "fantasy", 12345)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.GetWallTextureColor(1, 0.5, 0.5, 5.0)
	}
}

func BenchmarkNewRendererCore(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewRenderer(1280, 720)
	}
}

// Test fisheye correction math for texture coordinate calculations
func TestTextureCoordinateSeams(t *testing.T) {
	r := NewRendererWithGenre(640, 480, "fantasy", 12345)

	// Test texture seams at 0 and 1 boundaries
	c0 := r.GetWallTextureColor(1, 0.0, 0.5, 5.0)
	c1 := r.GetWallTextureColor(1, 1.0, 0.5, 5.0)

	// Both should produce valid colors (alpha = 255)
	if c0.A != 255 || c1.A != 255 {
		t.Error("texture seam should produce valid colors")
	}

	// Test vertical seam
	cv0 := r.GetWallTextureColor(1, 0.5, 0.0, 5.0)
	cv1 := r.GetWallTextureColor(1, 0.5, 1.0, 5.0)

	if cv0.A != 255 || cv1.A != 255 {
		t.Error("vertical texture seam should produce valid colors")
	}
}

// Test that different genres produce different colors
func TestGenreColorDifferences(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}
	colorSums := make(map[string]int)

	for _, genre := range genres {
		r := NewRendererWithGenre(640, 480, genre, 42)
		sum := 0
		for i := 0; i < 10; i++ {
			c := r.GetWallTextureColor(1, float64(i)/10, float64(i)/10, 3.0)
			sum += int(c.R) + int(c.G) + int(c.B)
		}
		colorSums[genre] = sum
	}

	// Count unique color sums
	unique := make(map[int]bool)
	for _, sum := range colorSums {
		unique[sum] = true
	}

	// At least 3 of 5 genres should have different color totals
	if len(unique) < 3 {
		t.Errorf("expected at least 3 distinct genre color profiles, got %d", len(unique))
	}
}

// Test edge cases in FOV ray calculation
func TestFOVConstants(t *testing.T) {
	// Verify FOV is approximately 60 degrees
	expectedFOV := math.Pi / 3
	if math.Abs(DefaultFOV-expectedFOV) > 0.001 {
		t.Errorf("DefaultFOV should be pi/3, got %f", DefaultFOV)
	}
}

// === MapCell Tests ===

func TestDefaultMapCell(t *testing.T) {
	cell := DefaultMapCell()
	if cell.WallType != 0 {
		t.Errorf("expected WallType=0, got %d", cell.WallType)
	}
	if cell.WallHeight != DefaultWallHeight {
		t.Errorf("expected WallHeight=%f, got %f", DefaultWallHeight, cell.WallHeight)
	}
	if cell.FloorH != 0.0 {
		t.Errorf("expected FloorH=0.0, got %f", cell.FloorH)
	}
	if !cell.IsEmpty() {
		t.Error("default cell should be empty")
	}
	if cell.IsSolid() {
		t.Error("default cell should not be solid")
	}
}

func TestWallMapCell(t *testing.T) {
	tests := []struct {
		wallType int
	}{
		{1},
		{2},
		{3},
	}
	for _, tc := range tests {
		cell := WallMapCell(tc.wallType)
		if cell.WallType != tc.wallType {
			t.Errorf("expected WallType=%d, got %d", tc.wallType, cell.WallType)
		}
		if cell.WallHeight != DefaultWallHeight {
			t.Errorf("expected WallHeight=%f, got %f", DefaultWallHeight, cell.WallHeight)
		}
		if cell.IsEmpty() {
			t.Error("wall cell should not be empty")
		}
		if !cell.IsSolid() {
			t.Error("wall cell should be solid")
		}
	}
}

func TestWallMapCellWithHeight(t *testing.T) {
	tests := []struct {
		name           string
		wallType       int
		height         float64
		expectedHeight float64
	}{
		{"standard height", 1, 1.0, 1.0},
		{"double height", 2, 2.0, 2.0},
		{"half height", 3, 0.5, 0.5},
		{"triple height", 1, 3.0, MaxWallHeight},
		{"below minimum", 2, 0.1, MinWallHeight},
		{"above maximum", 3, 5.0, MaxWallHeight},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cell := WallMapCellWithHeight(tc.wallType, tc.height)
			if cell.WallType != tc.wallType {
				t.Errorf("expected WallType=%d, got %d", tc.wallType, cell.WallType)
			}
			if cell.WallHeight != tc.expectedHeight {
				t.Errorf("expected WallHeight=%f, got %f", tc.expectedHeight, cell.WallHeight)
			}
			if !cell.IsSolid() {
				t.Error("wall cell should be solid")
			}
		})
	}
}

func TestCellFlags(t *testing.T) {
	tests := []struct {
		name           string
		flags          CellFlags
		isTransparent  bool
		isClimbable    bool
		isDestructible bool
	}{
		{"no flags", 0, false, false, false},
		{"transparent only", FlagTransparent, true, false, false},
		{"climbable only", FlagClimbable, false, true, false},
		{"destructible only", FlagDestructible, false, false, true},
		{"all flags", FlagTransparent | FlagClimbable | FlagDestructible, true, true, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cell := MapCell{Flags: tc.flags}
			if cell.IsTransparent() != tc.isTransparent {
				t.Errorf("IsTransparent: expected %v, got %v", tc.isTransparent, cell.IsTransparent())
			}
			if cell.IsClimbable() != tc.isClimbable {
				t.Errorf("IsClimbable: expected %v, got %v", tc.isClimbable, cell.IsClimbable())
			}
			if cell.IsDestructible() != tc.isDestructible {
				t.Errorf("IsDestructible: expected %v, got %v", tc.isDestructible, cell.IsDestructible())
			}
		})
	}
}

func TestEffectiveHeight(t *testing.T) {
	tests := []struct {
		name           string
		wallHeight     float64
		floorH         float64
		ceilH          float64
		expectedHeight float64
	}{
		{"default ceiling", 1.0, 0.0, 0.0, 1.0},
		{"explicit ceiling", 1.0, 0.0, 2.0, 2.0},
		{"elevated floor", 1.0, 0.5, 1.5, 1.0},
		{"tall building", 2.0, 0.0, 3.0, 3.0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cell := MapCell{
				WallHeight: tc.wallHeight,
				FloorH:     tc.floorH,
				CeilH:      tc.ceilH,
			}
			if cell.EffectiveHeight() != tc.expectedHeight {
				t.Errorf("expected EffectiveHeight=%f, got %f", tc.expectedHeight, cell.EffectiveHeight())
			}
		})
	}
}

func TestGetMapCell(t *testing.T) {
	r := NewRenderer(640, 480)

	// Test getting a cell from WorldMapCells
	cell := r.GetMapCell(4, 4)
	if cell.WallType != 2 {
		t.Errorf("expected WallType=2 at (4,4), got %d", cell.WallType)
	}

	// Test out of bounds returns default
	cell = r.GetMapCell(-1, 0)
	if cell.WallType != 0 {
		t.Errorf("expected WallType=0 for out of bounds, got %d", cell.WallType)
	}

	cell = r.GetMapCell(100, 100)
	if cell.WallType != 0 {
		t.Errorf("expected WallType=0 for out of bounds, got %d", cell.WallType)
	}
}

func TestHasWall(t *testing.T) {
	r := NewRenderer(640, 480)

	// Test wall positions
	if !r.HasWall(4, 4) {
		t.Error("expected wall at (4,4)")
	}
	if !r.HasWall(0, 0) {
		t.Error("expected boundary wall at (0,0)")
	}

	// Test empty positions
	if r.HasWall(7, 7) {
		t.Error("expected no wall at (7,7)")
	}

	// Test out of bounds
	if r.HasWall(-1, 0) {
		t.Error("expected no wall for out of bounds")
	}
}

func TestSetWorldMapCells(t *testing.T) {
	r := NewRenderer(640, 480)

	// Create a custom map with variable heights
	customMap := [][]MapCell{
		{WallMapCell(1), WallMapCellWithHeight(2, 2.0)},
		{DefaultMapCell(), WallMapCellWithHeight(3, 0.5)},
	}

	r.SetWorldMapCells(customMap)

	// Verify WorldMapCells was set
	if len(r.WorldMapCells) != 2 {
		t.Errorf("expected WorldMapCells length=2, got %d", len(r.WorldMapCells))
	}

	// Verify legacy WorldMap was also created
	if len(r.WorldMap) != 2 {
		t.Errorf("expected WorldMap length=2, got %d", len(r.WorldMap))
	}

	// Check cell at (0,1)
	cell := r.GetMapCell(0, 1)
	if cell.WallHeight != 2.0 {
		t.Errorf("expected WallHeight=2.0 at (0,1), got %f", cell.WallHeight)
	}

	// Check legacy map reflects wall type
	if r.WorldMap[0][1] != 2 {
		t.Errorf("expected WorldMap[0][1]=2, got %d", r.WorldMap[0][1])
	}
}

func TestRendererNewFieldsInitialized(t *testing.T) {
	r := NewRenderer(1280, 720)

	// Test new fields added to Renderer
	if r.PlayerZ != 0.5 {
		t.Errorf("expected PlayerZ=0.5, got %f", r.PlayerZ)
	}
	if r.PlayerPitch != 0.0 {
		t.Errorf("expected PlayerPitch=0.0, got %f", r.PlayerPitch)
	}
	if r.WorldMapCells == nil {
		t.Error("WorldMapCells should not be nil")
	}
	if len(r.WorldMapCells) != DefaultMapSize {
		t.Errorf("expected WorldMapCells size %d, got %d", DefaultMapSize, len(r.WorldMapCells))
	}

	// Verify WorldMapCells has varying heights in default map
	// Position (4,5) should have height 1.5, (8,8) should have height 2.0
	cell := r.GetMapCell(4, 5)
	if cell.WallHeight != 1.5 {
		t.Errorf("expected WallHeight=1.5 at (4,5), got %f", cell.WallHeight)
	}
	cell = r.GetMapCell(8, 8)
	if cell.WallHeight != 2.0 {
		t.Errorf("expected WallHeight=2.0 at (8,8), got %f", cell.WallHeight)
	}
	cell = r.GetMapCell(9, 8)
	if cell.WallHeight != 0.5 {
		t.Errorf("expected WallHeight=0.5 at (9,8), got %f", cell.WallHeight)
	}
}

func TestSetWorldMapWithWallHeights(t *testing.T) {
	r := NewRenderer(640, 480)

	// Create test heightmap (4x4)
	heightMap := []float64{
		0.7, 0.7, 0.2, 0.2, // Row 0: walls, empty
		0.7, 0.7, 0.2, 0.2, // Row 1: walls, empty
		0.2, 0.2, 0.7, 0.7, // Row 2: empty, walls
		0.2, 0.2, 0.7, 0.7, // Row 3: empty, walls
	}

	// Create test wall heights
	wallHeights := []float64{
		2.0, 1.5, 1.0, 1.0, // Row 0
		2.0, 1.5, 1.0, 1.0, // Row 1
		1.0, 1.0, 0.5, 3.0, // Row 2
		1.0, 1.0, 0.5, 3.0, // Row 3
	}

	r.SetWorldMapWithWallHeights(heightMap, wallHeights, 4, 0.5)

	// Verify WorldMap size
	if len(r.WorldMap) != 4 {
		t.Fatalf("expected WorldMap length=4, got %d", len(r.WorldMap))
	}

	// Verify WorldMapCells size
	if len(r.WorldMapCells) != 4 {
		t.Fatalf("expected WorldMapCells length=4, got %d", len(r.WorldMapCells))
	}

	// Check wall heights are applied correctly
	// Position (0,0) has height 0.7 > 0.5 threshold, wall height 2.0
	cell := r.GetMapCell(0, 0)
	if cell.WallType == 0 {
		t.Error("expected wall at (0,0)")
	}
	if cell.WallHeight != 2.0 {
		t.Errorf("expected WallHeight=2.0 at (0,0), got %f", cell.WallHeight)
	}

	// Position (2,0) has height 0.2 < 0.5 threshold, should be empty
	cell = r.GetMapCell(2, 0)
	if cell.WallType != 0 {
		t.Errorf("expected no wall at (2,0), got wallType=%d", cell.WallType)
	}

	// Position (3,3) has height 0.7 > 0.5 threshold, wall height 3.0
	cell = r.GetMapCell(3, 3)
	if cell.WallType == 0 {
		t.Error("expected wall at (3,3)")
	}
	if cell.WallHeight != MaxWallHeight {
		t.Errorf("expected WallHeight=%f at (3,3), got %f", MaxWallHeight, cell.WallHeight)
	}

	// Position (2,2) has height 0.7 > 0.5 threshold, wall height 0.5
	cell = r.GetMapCell(2, 2)
	if cell.WallType == 0 {
		t.Error("expected wall at (2,2)")
	}
	if cell.WallHeight != 0.5 {
		t.Errorf("expected WallHeight=0.5 at (2,2), got %f", cell.WallHeight)
	}
}

func TestRendererSkyboxInitialization(t *testing.T) {
	r := NewRenderer(640, 480)

	// Skybox should be initialized automatically
	if r.Skybox == nil {
		t.Fatal("expected Skybox to be initialized")
	}

	// Should default to fantasy genre
	if r.Skybox.GetConfig().Genre != "fantasy" {
		t.Errorf("expected genre=fantasy, got %s", r.Skybox.GetConfig().Genre)
	}
}

func TestRendererSkyboxGenreInheritance(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		r := NewRendererWithGenre(640, 480, genre, 12345)
		if r.Skybox == nil {
			t.Fatalf("expected Skybox for genre %s", genre)
		}
		if r.Skybox.GetConfig().Genre != genre {
			t.Errorf("expected skybox genre=%s, got %s", genre, r.Skybox.GetConfig().Genre)
		}
	}
}

func TestRendererSetSkybox(t *testing.T) {
	r := NewRenderer(640, 480)

	// Create custom skybox
	customSkybox := NewSkybox()
	customSkybox.SetTimeOfDay(18.0)

	r.SetSkybox(customSkybox)

	if r.GetSkybox() != customSkybox {
		t.Error("SetSkybox did not set the skybox")
	}
	if r.Skybox.GetTimeOfDay() != 18.0 {
		t.Errorf("expected TimeOfDay=18.0, got %f", r.Skybox.GetTimeOfDay())
	}
}

func TestGetHorizonLineDefault(t *testing.T) {
	r := NewRenderer(640, 480)

	// With no pitch, horizon should be at half height
	horizonY := r.getHorizonLine()
	expectedHorizon := 480 / 2

	if horizonY != expectedHorizon {
		t.Errorf("expected horizon at %d, got %d", expectedHorizon, horizonY)
	}
}

func TestGetHorizonLineWithPitch(t *testing.T) {
	r := NewRenderer(640, 480)
	halfHeight := 480 / 2

	// Test looking up (positive pitch moves horizon down)
	r.PlayerPitch = MaxPitchAngle
	horizonUp := r.getHorizonLine()
	if horizonUp <= halfHeight {
		t.Errorf("expected horizon below half when looking up, got %d", horizonUp)
	}

	// Test looking down (negative pitch moves horizon up)
	r.PlayerPitch = -MaxPitchAngle
	horizonDown := r.getHorizonLine()
	if horizonDown >= halfHeight {
		t.Errorf("expected horizon above half when looking down, got %d", horizonDown)
	}

	// Test horizon line clamping at screen edges
	r.PlayerPitch = MaxPitchAngle * 2 // Exceeds max
	horizonMax := r.getHorizonLine()
	if horizonMax > 480 {
		t.Errorf("horizon should be clamped to screen height, got %d", horizonMax)
	}

	r.PlayerPitch = -MaxPitchAngle * 2 // Exceeds max negative
	horizonMin := r.getHorizonLine()
	if horizonMin < 0 {
		t.Errorf("horizon should be clamped to 0, got %d", horizonMin)
	}
}

func TestMaxPitchAngleConstant(t *testing.T) {
	// MaxPitchAngle should be 85 degrees in radians
	expected := 85.0 * math.Pi / 180.0
	if math.Abs(MaxPitchAngle-expected) > 0.0001 {
		t.Errorf("expected MaxPitchAngle=%f radians (85°), got %f", expected, MaxPitchAngle)
	}
}

// ============================================================
// Transparent Wall Rendering Tests
// ============================================================

func TestGetTransparencyForFlags(t *testing.T) {
	tests := []struct {
		name     string
		flags    CellFlags
		expected float64
	}{
		{"no flags", 0, 1.0},
		{"solid only", FlagSolid, 1.0},
		{"transparent", FlagTransparent, 0.5},
		{"semi-opaque", FlagSemiOpaque, 0.9},
		{"transparent and solid", FlagTransparent | FlagSolid, 0.5},
		{"semi-opaque and solid", FlagSemiOpaque | FlagSolid, 0.9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTransparencyForFlags(tt.flags)
			if math.Abs(result-tt.expected) > 0.001 {
				t.Errorf("expected transparency=%f, got %f", tt.expected, result)
			}
		})
	}
}

func TestIsSemiOpaqueGap(t *testing.T) {
	// Test that the semi-opaque gap pattern creates gaps
	gapCount := 0
	solidCount := 0

	// Sample the pattern
	for i := 0; i < 100; i++ {
		texX := float64(i%10) / 10.0
		texY := float64(i/10) / 10.0
		if isSemiOpaqueGap(texX, texY) {
			gapCount++
		} else {
			solidCount++
		}
	}

	// Should have a mix of gaps and solid areas
	if gapCount == 0 {
		t.Error("expected some gaps in semi-opaque pattern")
	}
	if solidCount == 0 {
		t.Error("expected some solid areas in semi-opaque pattern")
	}
}

func TestMapCellTransparencyFlags(t *testing.T) {
	// Test that MapCell flags work correctly
	tests := []struct {
		name           string
		cell           MapCell
		isTransparent  bool
		isClimbable    bool
		isDestructible bool
	}{
		{
			name:           "default cell",
			cell:           DefaultMapCell(),
			isTransparent:  false,
			isClimbable:    false,
			isDestructible: false,
		},
		{
			name: "transparent cell",
			cell: MapCell{
				WallType:   1,
				WallHeight: 1.0,
				Flags:      FlagSolid | FlagTransparent,
			},
			isTransparent:  true,
			isClimbable:    false,
			isDestructible: false,
		},
		{
			name: "climbable cell",
			cell: MapCell{
				WallType:   1,
				WallHeight: 0.5,
				Flags:      FlagSolid | FlagClimbable,
			},
			isTransparent:  false,
			isClimbable:    true,
			isDestructible: false,
		},
		{
			name: "destructible cell",
			cell: MapCell{
				WallType:   1,
				WallHeight: 1.0,
				Flags:      FlagSolid | FlagDestructible,
			},
			isTransparent:  false,
			isClimbable:    false,
			isDestructible: true,
		},
		{
			name: "all flags",
			cell: MapCell{
				WallType:   1,
				WallHeight: 1.0,
				Flags:      FlagSolid | FlagTransparent | FlagClimbable | FlagDestructible,
			},
			isTransparent:  true,
			isClimbable:    true,
			isDestructible: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cell.IsTransparent() != tt.isTransparent {
				t.Errorf("IsTransparent() = %v, want %v", tt.cell.IsTransparent(), tt.isTransparent)
			}
			if tt.cell.IsClimbable() != tt.isClimbable {
				t.Errorf("IsClimbable() = %v, want %v", tt.cell.IsClimbable(), tt.isClimbable)
			}
			if tt.cell.IsDestructible() != tt.isDestructible {
				t.Errorf("IsDestructible() = %v, want %v", tt.cell.IsDestructible(), tt.isDestructible)
			}
		})
	}
}

func TestTransparentWallMapCell(t *testing.T) {
	// Helper to create a transparent wall cell
	cell := MapCell{
		WallType:   3,
		WallHeight: 1.0,
		Flags:      FlagSolid | FlagTransparent,
		MaterialID: 2,
	}

	if !cell.IsSolid() {
		t.Error("transparent wall should still be solid for collision")
	}
	if !cell.IsTransparent() {
		t.Error("transparent wall should be transparent")
	}
	if cell.WallType != 3 {
		t.Errorf("expected WallType=3, got %d", cell.WallType)
	}
}

func TestSemiOpaqueBarrierCell(t *testing.T) {
	// Helper to create a semi-opaque barrier cell (fence, grate)
	cell := MapCell{
		WallType:   4,
		WallHeight: 1.5,
		Flags:      FlagSolid | FlagSemiOpaque,
		MaterialID: 3,
	}

	if !cell.IsSolid() {
		t.Error("semi-opaque barrier should be solid")
	}
	if cell.Flags&FlagSemiOpaque == 0 {
		t.Error("semi-opaque flag should be set")
	}
}

func TestApplySideDarkening(t *testing.T) {
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}

	// No darkening
	result := applySideDarkening(white, 1.0)
	if result.R != 255 || result.G != 255 || result.B != 255 {
		t.Error("factor 1.0 should not change color")
	}
	if result.A != 255 {
		t.Error("alpha should not change")
	}

	// 80% darkening
	result = applySideDarkening(white, 0.8)
	if result.R != 204 || result.G != 204 || result.B != 204 {
		t.Errorf("expected RGB=(204,204,204), got (%d,%d,%d)", result.R, result.G, result.B)
	}
	if result.A != 255 {
		t.Error("alpha should not change during darkening")
	}
}

func TestGetSideDarkenFactor(t *testing.T) {
	// Side 0 (horizontal wall) should have no darkening
	factor0 := getSideDarkenFactor(0)
	if factor0 != 1.0 {
		t.Errorf("side 0 should have factor 1.0, got %f", factor0)
	}

	// Side 1 (vertical wall) should be darkened
	factor1 := getSideDarkenFactor(1)
	if factor1 != 0.8 {
		t.Errorf("side 1 should have factor 0.8, got %f", factor1)
	}
}
