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
