package city

import (
	"testing"
)

func TestGenerate(t *testing.T) {
	c := Generate(12345, "fantasy")
	if c == nil {
		t.Fatal("Generate returned nil")
	}
	if c.Name == "" {
		t.Error("city name should not be empty")
	}
	if c.Seed != 12345 {
		t.Errorf("expected seed=12345, got %d", c.Seed)
	}
	if c.Genre != "fantasy" {
		t.Errorf("expected genre='fantasy', got %q", c.Genre)
	}
	if len(c.Districts) < 3 || len(c.Districts) > 6 {
		t.Errorf("expected 3-6 districts, got %d", len(c.Districts))
	}
}

func TestGenerateDeterminism(t *testing.T) {
	seed := int64(42)

	c1 := Generate(seed, "fantasy")
	c2 := Generate(seed, "fantasy")

	if c1.Name != c2.Name {
		t.Errorf("determinism fail: names differ %q vs %q", c1.Name, c2.Name)
	}
	if len(c1.Districts) != len(c2.Districts) {
		t.Error("determinism fail: district counts differ")
	}
	for i := range c1.Districts {
		if c1.Districts[i].Name != c2.Districts[i].Name {
			t.Errorf("determinism fail: district %d names differ", i)
		}
		if c1.Districts[i].CenterX != c2.Districts[i].CenterX {
			t.Errorf("determinism fail: district %d CenterX differs", i)
		}
		if c1.Districts[i].CenterY != c2.Districts[i].CenterY {
			t.Errorf("determinism fail: district %d CenterY differs", i)
		}
	}
}

func TestGenerateDifferentSeeds(t *testing.T) {
	c1 := Generate(12345, "fantasy")
	c2 := Generate(54321, "fantasy")

	if c1.Name == c2.Name {
		t.Error("different seeds should produce different city names")
	}
}

func TestGenerateAllGenres(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}
	seed := int64(12345)

	names := make(map[string]bool)
	for _, genre := range genres {
		c := Generate(seed, genre)
		if c.Genre != genre {
			t.Errorf("expected genre=%q, got %q", genre, c.Genre)
		}
		if len(c.Districts) < 3 {
			t.Errorf("genre %q: expected at least 3 districts", genre)
		}
		names[c.Name] = true
	}

	// Different genres with same seed should produce different names
	if len(names) == 1 {
		t.Error("different genres should produce different city names")
	}
}

func TestGenerateUnknownGenre(t *testing.T) {
	c := Generate(12345, "unknown-genre")
	if c == nil {
		t.Fatal("should handle unknown genre gracefully")
	}
	// Should fall back to fantasy
	if len(c.Districts) < 3 {
		t.Error("unknown genre should still generate valid city")
	}
}

func TestDistrictProperties(t *testing.T) {
	c := Generate(12345, "fantasy")

	for i, d := range c.Districts {
		if d.Name == "" {
			t.Errorf("district %d: name should not be empty", i)
		}
		if d.Buildings < 10 || d.Buildings > 50 {
			t.Errorf("district %d: buildings should be 10-50, got %d", i, d.Buildings)
		}
		if d.Type == "" {
			t.Errorf("district %d: type should not be empty", i)
		}
		// CenterX and CenterY should be in range [-500, 500]
		if d.CenterX < -500 || d.CenterX > 500 {
			t.Errorf("district %d: CenterX out of range: %f", i, d.CenterX)
		}
		if d.CenterY < -500 || d.CenterY > 500 {
			t.Errorf("district %d: CenterY out of range: %f", i, d.CenterY)
		}
	}
}

func TestGenreNameVariety(t *testing.T) {
	// Generate multiple cities to verify name variety
	names := make(map[string]bool)
	for seed := int64(1); seed <= 100; seed++ {
		c := Generate(seed, "fantasy")
		names[c.Name] = true
	}

	// With 8 prefixes and 8 suffixes = 64 possible combinations
	// 100 cities should have significant variety
	if len(names) < 20 {
		t.Errorf("expected more name variety, got only %d unique names", len(names))
	}
}

func TestGenreDistrictTypes(t *testing.T) {
	// Verify each genre has appropriate district types
	tests := []struct {
		genre    string
		expected string
	}{
		{"fantasy", "Market"},
		{"sci-fi", "Research"},
		{"horror", "Cemetery"},
		{"cyberpunk", "Corporate"},
		{"post-apocalyptic", "Salvage"},
	}

	for _, tc := range tests {
		found := false
		for seed := int64(1); seed <= 50; seed++ {
			c := Generate(seed, tc.genre)
			for _, d := range c.Districts {
				if d.Type == tc.expected {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			t.Errorf("genre %q: expected to find %q district type", tc.genre, tc.expected)
		}
	}
}

func BenchmarkGenerate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Generate(int64(i), "fantasy")
	}
}

func BenchmarkGenerateAllGenres(b *testing.B) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, genre := range genres {
			_ = Generate(int64(i), genre)
		}
	}
}
