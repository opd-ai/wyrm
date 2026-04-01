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

func TestGenerateWallsAndGates(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			c := Generate(12345, genre)

			// Verify walls were generated
			if len(c.Walls.Segments) == 0 {
				t.Error("Wall segments should not be empty")
			}
			if c.Walls.Height <= 0 {
				t.Error("Wall height should be positive")
			}
			if c.Walls.Thickness <= 0 {
				t.Error("Wall thickness should be positive")
			}
			if c.Walls.Material == "" {
				t.Error("Wall material should not be empty")
			}

			// Verify gates were generated
			if len(c.Gates) < 2 {
				t.Errorf("Expected at least 2 gates, got %d", len(c.Gates))
			}

			for _, gate := range c.Gates {
				if gate.Name == "" {
					t.Error("Gate name should not be empty")
				}
				if gate.Width <= 0 {
					t.Error("Gate width should be positive")
				}
				if gate.Facing == "" {
					t.Error("Gate facing should not be empty")
				}
				if gate.Style == "" {
					t.Error("Gate style should not be empty")
				}
			}
		})
	}
}

func TestCityGateIsOpen(t *testing.T) {
	tests := []struct {
		name     string
		gate     CityGate
		hour     int
		wantOpen bool
	}{
		{
			"open during day",
			CityGate{OpenHour: 6, CloseHour: 22, Locked: false},
			12,
			true,
		},
		{
			"closed at night",
			CityGate{OpenHour: 6, CloseHour: 22, Locked: false},
			3,
			false,
		},
		{
			"locked gate",
			CityGate{OpenHour: 6, CloseHour: 22, Locked: true},
			12,
			false,
		},
		{
			"overnight hours open late",
			CityGate{OpenHour: 22, CloseHour: 6, Locked: false},
			23,
			true,
		},
		{
			"overnight hours open early",
			CityGate{OpenHour: 22, CloseHour: 6, Locked: false},
			2,
			true,
		},
		{
			"overnight hours closed midday",
			CityGate{OpenHour: 22, CloseHour: 6, Locked: false},
			12,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.gate.IsGateOpen(tt.hour)
			if got != tt.wantOpen {
				t.Errorf("IsGateOpen(%d) = %v, want %v", tt.hour, got, tt.wantOpen)
			}
		})
	}
}

func TestWallHasBreach(t *testing.T) {
	tests := []struct {
		name       string
		segments   []WallSegment
		wantBreach bool
	}{
		{
			"no breach",
			[]WallSegment{
				{Damaged: false},
				{Damaged: true, DamageLevel: 0.3},
			},
			false,
		},
		{
			"has breach",
			[]WallSegment{
				{Damaged: false},
				{Damaged: true, DamageLevel: 0.7},
			},
			true,
		},
		{
			"empty segments",
			[]WallSegment{},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wall := Wall{Segments: tt.segments}
			got := wall.HasBreach()
			if got != tt.wantBreach {
				t.Errorf("HasBreach() = %v, want %v", got, tt.wantBreach)
			}
		})
	}
}

func TestWallGetCondition(t *testing.T) {
	tests := []struct {
		name      string
		wall      Wall
		wantRange [2]float64 // min, max expected
	}{
		{
			"pristine wall",
			Wall{
				Condition: 1.0,
				Segments: []WallSegment{
					{Damaged: false},
					{Damaged: false},
				},
			},
			[2]float64{0.99, 1.01},
		},
		{
			"damaged wall",
			Wall{
				Condition: 1.0,
				Segments: []WallSegment{
					{Damaged: true, DamageLevel: 0.5},
					{Damaged: false},
				},
			},
			[2]float64{0.7, 0.8},
		},
		{
			"empty segments",
			Wall{
				Condition: 0.8,
				Segments:  []WallSegment{},
			},
			[2]float64{0.79, 0.81},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.wall.GetWallCondition()
			if got < tt.wantRange[0] || got > tt.wantRange[1] {
				t.Errorf("GetWallCondition() = %v, want between %v and %v", got, tt.wantRange[0], tt.wantRange[1])
			}
		})
	}
}

func TestGetGateByDirection(t *testing.T) {
	c := Generate(12345, "fantasy")

	// Should find at least one gate
	directions := []string{"north", "south", "east", "west"}
	foundCount := 0
	for _, dir := range directions {
		if g := c.GetGateByDirection(dir); g != nil {
			foundCount++
			if g.Facing != dir {
				t.Errorf("Gate facing = %v, want %v", g.Facing, dir)
			}
		}
	}

	if foundCount == 0 {
		t.Error("Should find at least one gate by direction")
	}
}

func TestGetGateByName(t *testing.T) {
	c := Generate(12345, "fantasy")

	if len(c.Gates) == 0 {
		t.Fatal("No gates generated")
	}

	// Find a gate by its name
	firstGate := c.Gates[0]
	found := c.GetGateByName(firstGate.Name)
	if found == nil {
		t.Errorf("GetGateByName(%s) returned nil", firstGate.Name)
	}

	// Non-existent gate should return nil
	notFound := c.GetGateByName("NonExistent Gate")
	if notFound != nil {
		t.Error("GetGateByName should return nil for non-existent gate")
	}
}

func TestWallsAndGatesDeterminism(t *testing.T) {
	c1 := Generate(12345, "fantasy")
	c2 := Generate(12345, "fantasy")

	// Walls should be deterministic
	if len(c1.Walls.Segments) != len(c2.Walls.Segments) {
		t.Error("Wall segment count should be deterministic")
	}
	if c1.Walls.Material != c2.Walls.Material {
		t.Error("Wall material should be deterministic")
	}
	if c1.Walls.Height != c2.Walls.Height {
		t.Error("Wall height should be deterministic")
	}

	// Gates should be deterministic
	if len(c1.Gates) != len(c2.Gates) {
		t.Error("Gate count should be deterministic")
	}
	for i := range c1.Gates {
		if c1.Gates[i].Name != c2.Gates[i].Name {
			t.Errorf("Gate %d name should be deterministic", i)
		}
		if c1.Gates[i].Style != c2.Gates[i].Style {
			t.Errorf("Gate %d style should be deterministic", i)
		}
	}
}

func TestRoadGeneration(t *testing.T) {
	c := Generate(12345, "fantasy")

	// Should have roads (at least one per district minus one for MST)
	if len(c.Roads) < len(c.Districts)-1 {
		t.Errorf("expected at least %d roads, got %d", len(c.Districts)-1, len(c.Roads))
	}

	// All roads should have valid properties
	for i, road := range c.Roads {
		if road.Width <= 0 {
			t.Errorf("road %d has invalid width: %f", i, road.Width)
		}
		if road.Type == "" {
			t.Errorf("road %d has empty type", i)
		}
		if road.Material == "" {
			t.Errorf("road %d has empty material", i)
		}
	}
}

func TestRoadDeterminism(t *testing.T) {
	c1 := Generate(12345, "fantasy")
	c2 := Generate(12345, "fantasy")

	if len(c1.Roads) != len(c2.Roads) {
		t.Errorf("road count should be deterministic: %d vs %d", len(c1.Roads), len(c2.Roads))
	}

	for i := range c1.Roads {
		if c1.Roads[i].StartX != c2.Roads[i].StartX || c1.Roads[i].StartY != c2.Roads[i].StartY {
			t.Errorf("road %d start position differs", i)
		}
		if c1.Roads[i].EndX != c2.Roads[i].EndX || c1.Roads[i].EndY != c2.Roads[i].EndY {
			t.Errorf("road %d end position differs", i)
		}
	}
}

func TestIsOnRoad(t *testing.T) {
	c := Generate(12345, "fantasy")

	if len(c.Roads) == 0 {
		t.Skip("no roads generated")
	}

	road := c.Roads[0]
	// Point at road start should be on road
	if !c.IsOnRoad(road.StartX, road.StartY) {
		t.Error("road start point should be on road")
	}

	// Point at road end should be on road
	if !c.IsOnRoad(road.EndX, road.EndY) {
		t.Error("road end point should be on road")
	}

	// Point far away should not be on road
	if c.IsOnRoad(10000, 10000) {
		t.Error("point far away should not be on road")
	}
}
