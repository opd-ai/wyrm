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

	// Clear the world to test max steps
	for i := range r.WorldMap {
		for j := range r.WorldMap[i] {
			r.WorldMap[i][j] = 0
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
