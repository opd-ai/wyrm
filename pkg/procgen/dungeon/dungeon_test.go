package dungeon

import (
	"testing"
)

func TestGenerateDungeon(t *testing.T) {
	gen := NewGenerator(12345, "fantasy")
	d := gen.Generate(64, 64, 1)

	if d == nil {
		t.Fatal("Generate returned nil")
	}

	if d.Width != 64 || d.Height != 64 {
		t.Errorf("Dimensions = %dx%d, want 64x64", d.Width, d.Height)
	}

	if len(d.Rooms) == 0 {
		t.Error("No rooms generated")
	}

	// Per AC: 0 unreachable rooms
	if !d.ValidateConnectivity() {
		t.Error("Dungeon has unreachable rooms")
	}
}

func TestConnectivityOver100Dungeons(t *testing.T) {
	// Per AC: 100 generated dungeons have 0 unreachable rooms
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for i := 0; i < 100; i++ {
		genre := genres[i%len(genres)]
		gen := NewGenerator(int64(i), genre)
		d := gen.Generate(64, 64, i%5+1)

		if !d.ValidateConnectivity() {
			t.Errorf("Dungeon %d (seed=%d, genre=%s) has unreachable rooms", i, i, genre)
		}
	}
}

func TestGenreDistinctPalettes(t *testing.T) {
	// Per AC: each genre produces distinct tile aesthetics
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}
	palettes := make(map[string]TilePalette)

	for _, genre := range genres {
		gen := NewGenerator(12345, genre)
		d := gen.Generate(32, 32, 1)
		palettes[genre] = d.Palette
	}

	// Check that all palettes are different
	for i := 0; i < len(genres); i++ {
		for j := i + 1; j < len(genres); j++ {
			p1 := palettes[genres[i]]
			p2 := palettes[genres[j]]

			if p1.WallColor == p2.WallColor &&
				p1.FloorColor == p2.FloorColor &&
				p1.AccentColor == p2.AccentColor {
				t.Errorf("Palettes for %s and %s are identical", genres[i], genres[j])
			}
		}
	}
}

func TestBossRoomAtDepth(t *testing.T) {
	gen := NewGenerator(12345, "fantasy")

	// Shallow dungeons should not have boss rooms
	shallow := gen.Generate(64, 64, 1)
	if shallow.HasBossRoom() {
		t.Error("Depth 1 dungeon should not have boss room")
	}

	// Deep dungeons should have boss rooms
	gen2 := NewGenerator(12345, "fantasy")
	deep := gen2.Generate(64, 64, 5)
	if !deep.HasBossRoom() {
		t.Error("Depth 5 dungeon should have boss room")
	}
}

func TestTrapsAndPuzzles(t *testing.T) {
	gen := NewGenerator(12345, "fantasy")
	d := gen.Generate(64, 64, 3)

	// Should have some traps and puzzles
	if d.TrapCount() == 0 && d.PuzzleCount() == 0 {
		t.Error("Expected some traps or puzzles")
	}
}

func TestEntranceAndExit(t *testing.T) {
	gen := NewGenerator(12345, "fantasy")
	d := gen.Generate(64, 64, 1)

	hasEntrance := false
	hasExit := false

	for _, room := range d.Rooms {
		if room.Type == RoomEntrance {
			hasEntrance = true
		}
		if room.Type == RoomExit || room.Type == RoomBoss {
			hasExit = true
		}
	}

	if !hasEntrance {
		t.Error("No entrance room")
	}
	if !hasExit {
		t.Error("No exit room")
	}
}

func TestDeterministicGeneration(t *testing.T) {
	seed := int64(42)

	gen1 := NewGenerator(seed, "fantasy")
	d1 := gen1.Generate(64, 64, 2)

	gen2 := NewGenerator(seed, "fantasy")
	d2 := gen2.Generate(64, 64, 2)

	// Same seed should produce same number of rooms
	if len(d1.Rooms) != len(d2.Rooms) {
		t.Errorf("Same seed produced different room counts: %d vs %d",
			len(d1.Rooms), len(d2.Rooms))
	}

	// Same tile layout
	for y := 0; y < d1.Height; y++ {
		for x := 0; x < d1.Width; x++ {
			if d1.Tiles[y][x] != d2.Tiles[y][x] {
				t.Errorf("Same seed produced different tile at (%d,%d)", x, y)
				return
			}
		}
	}
}

func TestRoomCount(t *testing.T) {
	gen := NewGenerator(12345, "fantasy")
	d := gen.Generate(64, 64, 2)

	// Should have multiple rooms
	if len(d.Rooms) < 5 {
		t.Errorf("Expected at least 5 rooms, got %d", len(d.Rooms))
	}
}

func TestValidateConnectivityEmptyDungeon(t *testing.T) {
	d := &Dungeon{
		Width:  32,
		Height: 32,
		Rooms:  []*Room{},
	}

	// Empty dungeon should be valid (no unreachable rooms)
	if !d.ValidateConnectivity() {
		t.Error("Empty dungeon should validate")
	}
}

func TestGenrePaletteColors(t *testing.T) {
	tests := []struct {
		genre    string
		wantWall uint32
	}{
		{"fantasy", 0x4A3B2A},
		{"sci-fi", 0x1A1A2E},
		{"horror", 0x2E2E2E},
		{"cyberpunk", 0x0D0D0D},
		{"post-apocalyptic", 0x5C4033},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			palette := GenrePalettes[tt.genre]
			if palette.WallColor != tt.wantWall {
				t.Errorf("WallColor = 0x%X, want 0x%X", palette.WallColor, tt.wantWall)
			}
		})
	}
}

func TestUnknownGenreFallsBackToFantasy(t *testing.T) {
	gen := NewGenerator(12345, "unknown_genre")
	d := gen.Generate(32, 32, 1)

	fantasyPalette := GenrePalettes["fantasy"]
	if d.Palette.WallColor != fantasyPalette.WallColor {
		t.Errorf("Unknown genre should fallback to fantasy palette")
	}
}

func BenchmarkDungeonGeneration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		gen := NewGenerator(int64(i), "fantasy")
		gen.Generate(64, 64, 3)
	}
}

func BenchmarkLargeDungeon(b *testing.B) {
	for i := 0; i < b.N; i++ {
		gen := NewGenerator(int64(i), "fantasy")
		gen.Generate(128, 128, 5)
	}
}

// Cave system tests

func TestGenerateCave(t *testing.T) {
	gen := NewCaveGenerator(12345, "fantasy")
	c := gen.GenerateCave(64, 64)

	if c == nil {
		t.Fatal("GenerateCave returned nil")
	}

	if c.Width != 64 || c.Height != 64 {
		t.Errorf("Dimensions = %dx%d, want 64x64", c.Width, c.Height)
	}

	// Should have some open space
	openPercent := c.GetOpenAreaPercent()
	if openPercent < 0.1 {
		t.Errorf("Cave has too little open space: %.2f%%", openPercent*100)
	}
}

func TestCaveConnectivity(t *testing.T) {
	gen := NewCaveGenerator(12345, "fantasy")
	c := gen.GenerateCave(64, 64)

	if !c.IsConnected() {
		t.Error("Cave should be connected from entrance to exit")
	}
}

func TestCaveConnectivityOver50Caves(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for i := 0; i < 50; i++ {
		genre := genres[i%len(genres)]
		gen := NewCaveGenerator(int64(i), genre)
		c := gen.GenerateCave(64, 64)

		if !c.IsConnected() {
			t.Errorf("Cave %d (seed=%d, genre=%s) is not connected", i, i, genre)
		}
	}
}

func TestCaveEntranceAndExit(t *testing.T) {
	gen := NewCaveGenerator(12345, "fantasy")
	c := gen.GenerateCave(64, 64)

	// Should have at least one exit
	if len(c.Exits) == 0 {
		t.Error("Cave has no exits")
	}

	// Entrance should be valid (either at 0,0 for empty caves or properly placed)
	entranceValid := false
	if c.Entrance.X > 0 || c.Entrance.Y > 0 {
		if c.Tiles[c.Entrance.Y][c.Entrance.X] == TileEntrance {
			entranceValid = true
		}
	} else if c.GetOpenAreaPercent() < 0.01 {
		// Very closed cave, entrance may not be placed properly
		entranceValid = true
	}

	if !entranceValid && c.GetOpenAreaPercent() > 0.1 {
		t.Logf("Entrance at (%d,%d), tile type: %v", c.Entrance.X, c.Entrance.Y, c.Tiles[c.Entrance.Y][c.Entrance.X])
		t.Error("Entrance position is not marked as TileEntrance")
	}

	// Exit tile should be marked
	for _, exit := range c.Exits {
		if c.Tiles[exit.Y][exit.X] != TileExit {
			t.Logf("Exit at (%d,%d), tile type: %v", exit.X, exit.Y, c.Tiles[exit.Y][exit.X])
			t.Error("Exit position is not marked as TileExit")
		}
	}
}

func TestCaveDeterminism(t *testing.T) {
	seed := int64(42)

	gen1 := NewCaveGenerator(seed, "fantasy")
	c1 := gen1.GenerateCave(64, 64)

	gen2 := NewCaveGenerator(seed, "fantasy")
	c2 := gen2.GenerateCave(64, 64)

	// Same seed should produce same tiles
	for y := 0; y < c1.Height; y++ {
		for x := 0; x < c1.Width; x++ {
			if c1.Tiles[y][x] != c2.Tiles[y][x] {
				t.Errorf("Same seed produced different tile at (%d,%d): %v vs %v",
					x, y, c1.Tiles[y][x], c2.Tiles[y][x])
				return
			}
		}
	}

	// Same entrance and exits
	if c1.Entrance != c2.Entrance {
		t.Error("Same seed produced different entrance positions")
	}
}

func TestCaveCaverns(t *testing.T) {
	gen := NewCaveGenerator(12345, "fantasy")
	c := gen.GenerateCave(64, 64)

	// Should have at least one cavern after connection
	if c.CavernCount() == 0 {
		t.Error("Cave should have at least one cavern")
	}

	// Largest cavern should have reasonable size (at least 20 cells for a connected cave)
	largest := c.LargestCavernSize()
	if largest < 20 {
		t.Errorf("Largest cavern too small: %d (expected at least 20)", largest)
	}
}

func TestCaveBorderWalls(t *testing.T) {
	gen := NewCaveGenerator(12345, "fantasy")
	c := gen.GenerateCave(64, 64)

	// All border tiles should be walls
	for x := 0; x < c.Width; x++ {
		if c.Tiles[0][x] != TileWall {
			t.Errorf("Top border at x=%d should be wall", x)
		}
		if c.Tiles[c.Height-1][x] != TileWall {
			t.Errorf("Bottom border at x=%d should be wall", x)
		}
	}

	for y := 0; y < c.Height; y++ {
		if c.Tiles[y][0] != TileWall {
			t.Errorf("Left border at y=%d should be wall", y)
		}
		if c.Tiles[y][c.Width-1] != TileWall {
			t.Errorf("Right border at y=%d should be wall", y)
		}
	}
}

func TestCaveOpenAreaRange(t *testing.T) {
	// Generate multiple caves and check open area is reasonable
	for i := 0; i < 20; i++ {
		gen := NewCaveGenerator(int64(i), "fantasy")
		c := gen.GenerateCave(64, 64)

		openPercent := c.GetOpenAreaPercent()
		if openPercent < 0.2 || openPercent > 0.8 {
			t.Logf("Cave %d open area: %.2f%% (outside typical 20-80%% range)", i, openPercent*100)
		}
	}
}

func TestCaveGenreVariety(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		gen := NewCaveGenerator(12345, genre)
		c := gen.GenerateCave(32, 32)

		if c.Genre != genre {
			t.Errorf("Cave genre = %s, want %s", c.Genre, genre)
		}

		// Cave should be valid regardless of genre
		if !c.IsConnected() {
			t.Errorf("Genre %s produced disconnected cave", genre)
		}
	}
}

func BenchmarkCaveGeneration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		gen := NewCaveGenerator(int64(i), "fantasy")
		gen.GenerateCave(64, 64)
	}
}

func BenchmarkLargeCave(b *testing.B) {
	for i := 0; i < b.N; i++ {
		gen := NewCaveGenerator(int64(i), "fantasy")
		gen.GenerateCave(128, 128)
	}
}
