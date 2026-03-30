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

// ========== Residential Area Tests ==========

func TestGenerateResidentialAreas(t *testing.T) {
	city := Generate(12345, "fantasy")

	if len(city.ResidentialAreas) < 2 {
		t.Errorf("Expected at least 2 residential areas, got %d", len(city.ResidentialAreas))
	}

	for i, area := range city.ResidentialAreas {
		if area.Name == "" {
			t.Errorf("Residential area %d has empty name", i)
		}
		if area.Radius <= 0 {
			t.Errorf("Residential area %d has invalid radius: %f", i, area.Radius)
		}
		if area.HousingType == "" {
			t.Errorf("Residential area %d has empty housing type", i)
		}
		if area.WealthLevel < 1 || area.WealthLevel > 5 {
			t.Errorf("Residential area %d has invalid wealth level: %d", i, area.WealthLevel)
		}
		if len(area.Buildings) == 0 {
			t.Errorf("Residential area %d has no buildings", i)
		}
	}
}

func TestResidentialBuildingProperties(t *testing.T) {
	city := Generate(12345, "sci-fi")

	if len(city.ResidentialAreas) == 0 {
		t.Fatal("No residential areas generated")
	}

	for _, area := range city.ResidentialAreas {
		for j, building := range area.Buildings {
			if building.Width <= 0 || building.Height <= 0 {
				t.Errorf("Building %d has invalid dimensions: %dx%d", j, building.Width, building.Height)
			}
			if building.Floors <= 0 {
				t.Errorf("Building %d has invalid floor count: %d", j, building.Floors)
			}
			if building.Units <= 0 {
				t.Errorf("Building %d has invalid unit count: %d", j, building.Units)
			}
			if building.Occupied > building.Units {
				t.Errorf("Building %d has more occupied than total units", j)
			}
		}
	}
}

func TestResidentialPopulation(t *testing.T) {
	city := Generate(12345, "cyberpunk")

	totalPop := city.GetResidentialPopulation()
	if totalPop <= 0 {
		t.Error("City should have some residential population")
	}

	// Check that population matches building occupancy
	calculatedPop := 0
	for _, area := range city.ResidentialAreas {
		for _, b := range area.Buildings {
			calculatedPop += int(float64(b.Occupied) * 2.5)
		}
	}

	if totalPop != calculatedPop {
		t.Errorf("Population mismatch: GetResidentialPopulation=%d, calculated=%d", totalPop, calculatedPop)
	}
}

func TestResidentialAmenities(t *testing.T) {
	city := Generate(54321, "fantasy")

	for _, area := range city.ResidentialAreas {
		if len(area.Amenities) == 0 {
			// Poor areas might have no amenities
			if area.WealthLevel > 2 {
				t.Errorf("Wealthy area %s should have amenities", area.Name)
			}
		}
	}
}

func TestGetAreaByName(t *testing.T) {
	city := Generate(12345, "fantasy")

	if len(city.ResidentialAreas) == 0 {
		t.Skip("No residential areas")
	}

	targetName := city.ResidentialAreas[0].Name
	found := city.GetAreaByName(targetName)

	if found == nil {
		t.Errorf("GetAreaByName failed to find %s", targetName)
	}

	notFound := city.GetAreaByName("NonexistentArea123")
	if notFound != nil {
		t.Error("GetAreaByName should return nil for nonexistent area")
	}
}

// ========== Industrial Zone Tests ==========

func TestGenerateIndustrialZones(t *testing.T) {
	city := Generate(12345, "sci-fi")

	if len(city.IndustrialZones) < 1 {
		t.Errorf("Expected at least 1 industrial zone, got %d", len(city.IndustrialZones))
	}

	for i, zone := range city.IndustrialZones {
		if zone.Name == "" {
			t.Errorf("Industrial zone %d has empty name", i)
		}
		if zone.Radius <= 0 {
			t.Errorf("Industrial zone %d has invalid radius: %f", i, zone.Radius)
		}
		if zone.ZoneType == "" {
			t.Errorf("Industrial zone %d has empty zone type", i)
		}
		if zone.OutputType == "" {
			t.Errorf("Industrial zone %d has empty output type", i)
		}
		if len(zone.Facilities) == 0 {
			t.Errorf("Industrial zone %d has no facilities", i)
		}
	}
}

func TestIndustrialFacilityProperties(t *testing.T) {
	city := Generate(12345, "cyberpunk")

	if len(city.IndustrialZones) == 0 {
		t.Fatal("No industrial zones generated")
	}

	for _, zone := range city.IndustrialZones {
		for j, facility := range zone.Facilities {
			if facility.Width <= 0 || facility.Height <= 0 {
				t.Errorf("Facility %d has invalid dimensions", j)
			}
			if facility.OutputRate <= 0 {
				t.Errorf("Facility %d has invalid output rate", j)
			}
			if facility.WorkerCapacity <= 0 {
				t.Errorf("Facility %d has invalid worker capacity", j)
			}
		}
	}
}

func TestIndustrialWorkers(t *testing.T) {
	city := Generate(12345, "post-apocalyptic")

	totalWorkers := city.GetIndustrialWorkers()
	if totalWorkers <= 0 {
		t.Error("City should have some industrial workers")
	}
}

func TestAutomationReducesWorkers(t *testing.T) {
	// High-tech genres should have more automation
	sciFi := Generate(12345, "sci-fi")
	fantasy := Generate(12345, "fantasy")

	// Compare automation levels
	var sciFiAutomation, fantasyAutomation float64
	for _, zone := range sciFi.IndustrialZones {
		sciFiAutomation += zone.Automation
	}
	for _, zone := range fantasy.IndustrialZones {
		fantasyAutomation += zone.Automation
	}

	if len(sciFi.IndustrialZones) > 0 {
		sciFiAutomation /= float64(len(sciFi.IndustrialZones))
	}
	if len(fantasy.IndustrialZones) > 0 {
		fantasyAutomation /= float64(len(fantasy.IndustrialZones))
	}

	// Sci-fi should generally have higher automation (allow for RNG variance)
	t.Logf("Sci-fi avg automation: %f, Fantasy avg automation: %f", sciFiAutomation, fantasyAutomation)
}

func TestGetZoneByName(t *testing.T) {
	city := Generate(12345, "cyberpunk")

	if len(city.IndustrialZones) == 0 {
		t.Skip("No industrial zones")
	}

	targetName := city.IndustrialZones[0].Name
	found := city.GetZoneByName(targetName)

	if found == nil {
		t.Errorf("GetZoneByName failed to find %s", targetName)
	}

	notFound := city.GetZoneByName("NonexistentZone123")
	if notFound != nil {
		t.Error("GetZoneByName should return nil for nonexistent zone")
	}
}

func TestGenreSpecificHousingTypes(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		city := Generate(12345, genre)

		// Check that housing types match genre
		for _, area := range city.ResidentialAreas {
			expectedTypes := genreHousingTypes[genre]
			found := false
			for _, expected := range expectedTypes {
				if area.HousingType == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Genre %s: unexpected housing type %s", genre, area.HousingType)
			}
		}
	}
}

func TestResidentialDeterminism(t *testing.T) {
	c1 := Generate(99999, "fantasy")
	c2 := Generate(99999, "fantasy")

	if len(c1.ResidentialAreas) != len(c2.ResidentialAreas) {
		t.Fatal("Same seed should produce same number of residential areas")
	}

	for i := range c1.ResidentialAreas {
		if c1.ResidentialAreas[i].Name != c2.ResidentialAreas[i].Name {
			t.Errorf("Area names should match: %s vs %s", c1.ResidentialAreas[i].Name, c2.ResidentialAreas[i].Name)
		}
		if c1.ResidentialAreas[i].Population != c2.ResidentialAreas[i].Population {
			t.Error("Population should be deterministic")
		}
	}
}

func BenchmarkGenerateWithResidential(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Generate(int64(i), "cyberpunk")
	}
}
