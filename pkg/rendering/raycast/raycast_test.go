//go:build noebiten

package raycast

import (
	"testing"
)

func TestNewRenderer(t *testing.T) {
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
	if r.PlayerX != 8.0 {
		t.Errorf("expected PlayerX=8.0, got %f", r.PlayerX)
	}
	if r.PlayerY != 8.0 {
		t.Errorf("expected PlayerY=8.0, got %f", r.PlayerY)
	}
	if r.PlayerA != 0.0 {
		t.Errorf("expected PlayerA=0.0, got %f", r.PlayerA)
	}
	if r.WorldMap == nil {
		t.Error("WorldMap should not be nil")
	}
	if len(r.WorldMap) != 16 {
		t.Errorf("expected WorldMap size 16, got %d", len(r.WorldMap))
	}
}

func TestNewRendererWorldMapBoundaries(t *testing.T) {
	r := NewRenderer(640, 480)

	// Check boundary walls
	for i := 0; i < 16; i++ {
		if r.WorldMap[i][0] != 1 {
			t.Errorf("expected boundary wall at [%d][0]", i)
		}
		if r.WorldMap[i][15] != 1 {
			t.Errorf("expected boundary wall at [%d][15]", i)
		}
	}
	for j := 0; j < 16; j++ {
		if r.WorldMap[0][j] != 1 {
			t.Errorf("expected boundary wall at [0][%d]", j)
		}
		if r.WorldMap[15][j] != 1 {
			t.Errorf("expected boundary wall at [15][%d]", j)
		}
	}
}

func TestNewRendererInteriorWalls(t *testing.T) {
	r := NewRenderer(640, 480)

	// Check interior walls
	if r.WorldMap[4][4] != 2 {
		t.Error("expected wall type 2 at [4][4]")
	}
	if r.WorldMap[8][8] != 3 {
		t.Error("expected wall type 3 at [8][8]")
	}
}

func TestSetPlayerPos(t *testing.T) {
	r := NewRenderer(640, 480)

	r.SetPlayerPos(5.5, 7.2, 1.57)

	if r.PlayerX != 5.5 {
		t.Errorf("expected PlayerX=5.5, got %f", r.PlayerX)
	}
	if r.PlayerY != 7.2 {
		t.Errorf("expected PlayerY=7.2, got %f", r.PlayerY)
	}
	if r.PlayerA != 1.57 {
		t.Errorf("expected PlayerA=1.57, got %f", r.PlayerA)
	}
}

func TestCastRay(t *testing.T) {
	r := NewRenderer(640, 480)
	// Player at center (8, 8), facing right (angle 0)
	// Note: There's an interior wall at [8][8] (type 3)

	// Cast ray to the right - will hit the interior wall at [8][8]
	dist, wallType := r.castRay(1.0, 0.0)
	if dist <= 0 {
		t.Error("distance should be positive")
	}
	// Wall type should be either 3 (interior at [8][8]) or 1 (boundary)
	if wallType != 3 && wallType != 1 {
		t.Errorf("expected wallType=3 or 1, got %d", wallType)
	}

	// Move player to a clear position and test boundary hit
	r.SetPlayerPos(12.0, 4.0, 0) // Clear area, facing right toward boundary
	dist2, wallType2 := r.castRay(1.0, 0.0)
	if dist2 <= 0 {
		t.Error("distance should be positive")
	}
	if wallType2 != 1 {
		t.Errorf("expected wallType=1 (boundary), got %d", wallType2)
	}
	// Distance to boundary should be approximately 3 (from 12 to 15)
	if dist2 < 2 || dist2 > 4 {
		t.Errorf("distance to boundary seems wrong: %f", dist2)
	}
}

func TestCastRayDifferentDirections(t *testing.T) {
	r := NewRenderer(640, 480)

	tests := []struct {
		dirX, dirY float64
		desc       string
	}{
		{1.0, 0.0, "right"},
		{-1.0, 0.0, "left"},
		{0.0, 1.0, "down"},
		{0.0, -1.0, "up"},
		{0.707, 0.707, "diagonal"},
	}

	for _, tc := range tests {
		dist, wallType := r.castRay(tc.dirX, tc.dirY)
		if dist <= 0 || dist > 100 {
			t.Errorf("%s: invalid distance %f", tc.desc, dist)
		}
		if wallType < 0 {
			t.Errorf("%s: invalid wall type %d", tc.desc, wallType)
		}
	}
}

func TestCastRayInteriorWall(t *testing.T) {
	r := NewRenderer(640, 480)

	// Position player to face an interior wall
	r.SetPlayerPos(2.0, 4.5, 0) // facing right, should hit wall at [4][4]

	dist, wallType := r.castRay(1.0, 0.0)
	if dist < 1 || dist > 3 {
		t.Errorf("expected distance ~2 to interior wall, got %f", dist)
	}
	if wallType != 2 {
		t.Errorf("expected wallType=2 (interior), got %d", wallType)
	}
}

func TestGetWallColor(t *testing.T) {
	r := NewRenderer(640, 480)

	tests := []struct {
		wallType int
		distance float64
	}{
		{1, 1.0},
		{2, 5.0},
		{3, 10.0},
		{0, 2.0}, // default
	}

	for _, tc := range tests {
		color := r.getWallColor(tc.wallType, tc.distance)
		if color.A != 255 {
			t.Errorf("wallType=%d: alpha should be 255", tc.wallType)
		}
		// Color should be dimmer with distance
		nearColor := r.getWallColor(tc.wallType, 1.0)
		farColor := r.getWallColor(tc.wallType, 10.0)
		if farColor.R > nearColor.R && nearColor.R > 0 {
			t.Error("far walls should be darker")
		}
	}
}

func TestGetWallColorDifferentTypes(t *testing.T) {
	r := NewRenderer(640, 480)
	dist := 3.0

	c1 := r.getWallColor(1, dist) // red-brown
	c2 := r.getWallColor(2, dist) // green
	c3 := r.getWallColor(3, dist) // blue

	// Different wall types should have different colors
	if c1.R == c2.R && c1.G == c2.G && c1.B == c2.B {
		t.Error("wall types 1 and 2 should have different colors")
	}
	if c2.R == c3.R && c2.G == c3.G && c2.B == c3.B {
		t.Error("wall types 2 and 3 should have different colors")
	}
}

func TestGetWallColorDistanceFogBrightness(t *testing.T) {
	r := NewRenderer(640, 480)

	near := r.getWallColor(1, 1.0)
	mid := r.getWallColor(1, 8.0)
	far := r.getWallColor(1, 15.0)

	// Near should be brightest
	nearBrightness := int(near.R) + int(near.G) + int(near.B)
	midBrightness := int(mid.R) + int(mid.G) + int(mid.B)
	farBrightness := int(far.R) + int(far.G) + int(far.B)

	if nearBrightness <= midBrightness {
		t.Error("near walls should be brighter than mid")
	}
	if midBrightness <= farBrightness {
		t.Error("mid walls should be brighter than far")
	}
}

func TestFOV(t *testing.T) {
	r := NewRenderer(640, 480)

	// Default FOV should be approximately 60 degrees (pi/3)
	expectedFOV := 1.047 // pi/3
	if r.FOV < expectedFOV-0.1 || r.FOV > expectedFOV+0.1 {
		t.Errorf("expected FOV near pi/3, got %f", r.FOV)
	}
}

func BenchmarkCastRay(b *testing.B) {
	r := NewRenderer(1280, 720)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.castRay(1.0, 0.0)
	}
}

func BenchmarkGetWallColor(b *testing.B) {
	r := NewRenderer(1280, 720)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.getWallColor(1, 5.0)
	}
}

func BenchmarkNewRenderer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewRenderer(1280, 720)
	}
}

func TestNewRendererWithGenre(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}
	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			r := NewRendererWithGenre(640, 480, genre, 12345)
			if r == nil {
				t.Fatal("NewRendererWithGenre returned nil")
			}
			if r.Genre != genre {
				t.Errorf("expected Genre=%s, got %s", genre, r.Genre)
			}
			if r.WallTextures == nil {
				t.Error("WallTextures should not be nil")
			}
			if len(r.WallTextures) != 4 {
				t.Errorf("expected 4 wall textures, got %d", len(r.WallTextures))
			}
			if r.FloorTexture == nil {
				t.Error("FloorTexture should not be nil")
			}
			if r.CeilTexture == nil {
				t.Error("CeilTexture should not be nil")
			}
		})
	}
}

func TestSetGenre(t *testing.T) {
	r := NewRendererWithGenre(640, 480, "fantasy", 12345)
	originalTexture := r.WallTextures[0]

	// Change genre
	r.SetGenre("cyberpunk", 12345)

	if r.Genre != "cyberpunk" {
		t.Errorf("expected Genre=cyberpunk, got %s", r.Genre)
	}

	// Textures should be regenerated (different genre = different colors)
	// Note: The texture object itself changes
	if r.WallTextures[0] == originalTexture {
		t.Error("expected textures to be regenerated")
	}
}

func TestGetWallTextureColor(t *testing.T) {
	r := NewRendererWithGenre(640, 480, "fantasy", 12345)

	tests := []struct {
		wallType int
		texX     float64
		texY     float64
		distance float64
	}{
		{0, 0.0, 0.0, 1.0},
		{1, 0.5, 0.5, 5.0},
		{2, 0.25, 0.75, 10.0},
		{3, 0.9, 0.1, 15.0},
	}

	for _, tc := range tests {
		color := r.GetWallTextureColor(tc.wallType, tc.texX, tc.texY, tc.distance)
		if color.A != 255 {
			t.Errorf("wallType=%d: alpha should be 255, got %d", tc.wallType, color.A)
		}
	}
}

func TestGetFloorTextureColor(t *testing.T) {
	r := NewRendererWithGenre(640, 480, "fantasy", 12345)

	nearColor := r.GetFloorTextureColor(0.5, 0.5, 1.0)
	farColor := r.GetFloorTextureColor(0.5, 0.5, 15.0)

	if nearColor.A != 255 || farColor.A != 255 {
		t.Error("alpha should be 255")
	}

	// Near should be brighter
	nearBrightness := int(nearColor.R) + int(nearColor.G) + int(nearColor.B)
	farBrightness := int(farColor.R) + int(farColor.G) + int(farColor.B)
	if farBrightness > nearBrightness && nearBrightness > 0 {
		t.Error("near floor should be brighter than far")
	}
}

func TestGetCeilingTextureColor(t *testing.T) {
	r := NewRendererWithGenre(640, 480, "fantasy", 12345)

	nearColor := r.GetCeilingTextureColor(0.5, 0.5, 1.0)
	farColor := r.GetCeilingTextureColor(0.5, 0.5, 15.0)

	if nearColor.A != 255 || farColor.A != 255 {
		t.Error("alpha should be 255")
	}

	// Near should be brighter
	nearBrightness := int(nearColor.R) + int(nearColor.G) + int(nearColor.B)
	farBrightness := int(farColor.R) + int(farColor.G) + int(farColor.B)
	if farBrightness > nearBrightness && nearBrightness > 0 {
		t.Error("near ceiling should be brighter than far")
	}
}

func TestTextureDeterminism(t *testing.T) {
	r1 := NewRendererWithGenre(640, 480, "fantasy", 42)
	r2 := NewRendererWithGenre(640, 480, "fantasy", 42)

	// Same seed should produce identical textures
	for wallType := 0; wallType < 4; wallType++ {
		for i := 0; i < 100; i++ {
			c1 := r1.GetWallTextureColor(wallType, float64(i)/100, float64(i)/100, 5.0)
			c2 := r2.GetWallTextureColor(wallType, float64(i)/100, float64(i)/100, 5.0)
			if c1 != c2 {
				t.Errorf("wallType=%d, idx=%d: colors differ (%v vs %v)", wallType, i, c1, c2)
				break
			}
		}
	}
}

func TestGenrePaletteDifference(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}
	colors := make(map[string][]uint32)

	for _, genre := range genres {
		r := NewRendererWithGenre(640, 480, genre, 12345)
		// Sample some colors from wall texture
		var samples []uint32
		for i := 0; i < 10; i++ {
			c := r.GetWallTextureColor(1, float64(i)/10, float64(i)/10, 3.0)
			samples = append(samples, uint32(c.R)<<16|uint32(c.G)<<8|uint32(c.B))
		}
		colors[genre] = samples
	}

	// Each genre should have at least some different colors
	distinctGenres := 0
	for i := 0; i < len(genres); i++ {
		for j := i + 1; j < len(genres); j++ {
			different := false
			for k := 0; k < len(colors[genres[i]]); k++ {
				if colors[genres[i]][k] != colors[genres[j]][k] {
					different = true
					break
				}
			}
			if different {
				distinctGenres++
			}
		}
	}

	// At least 80% of genre pairs should have different colors
	totalPairs := len(genres) * (len(genres) - 1) / 2
	if distinctGenres < totalPairs*8/10 {
		t.Errorf("expected most genres to have distinct colors, got %d/%d different", distinctGenres, totalPairs)
	}
}

func BenchmarkGetWallTextureColor(b *testing.B) {
	r := NewRendererWithGenre(1280, 720, "fantasy", 12345)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.GetWallTextureColor(1, 0.5, 0.5, 5.0)
	}
}

func BenchmarkGetFloorTextureColor(b *testing.B) {
	r := NewRendererWithGenre(1280, 720, "fantasy", 12345)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.GetFloorTextureColor(0.5, 0.5, 5.0)
	}
}
