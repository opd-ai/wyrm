//go:build noebiten

package adapters

import (
	"testing"
)

func TestGetGenreBiomeDistribution(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			dist := GetGenreBiomeDistribution(genre)
			if dist == nil {
				t.Fatalf("GetGenreBiomeDistribution(%s) returned nil", genre)
			}

			// Verify we have primary biomes
			if len(dist.PrimaryBiomes) == 0 {
				t.Errorf("Genre %s has no primary biomes", genre)
			}

			// Verify weights sum to approximately 1.0
			totalWeight := 0.0
			for _, weight := range dist.Weights {
				totalWeight += weight
			}
			if totalWeight < 0.99 || totalWeight > 1.01 {
				t.Errorf("Genre %s weights sum to %f, expected ~1.0", genre, totalWeight)
			}

			// Verify all primary biomes have weights
			for _, biome := range dist.PrimaryBiomes {
				if _, ok := dist.Weights[biome]; !ok {
					t.Errorf("Genre %s primary biome %d has no weight", genre, biome)
				}
			}
		})
	}
}

func TestGetGenreBiomeDistributionUnknownGenre(t *testing.T) {
	dist := GetGenreBiomeDistribution("unknown_genre")
	if dist == nil {
		t.Fatal("GetGenreBiomeDistribution for unknown genre returned nil")
	}

	// Should default to fantasy
	fantasyDist := GetGenreBiomeDistribution("fantasy")
	if len(dist.PrimaryBiomes) != len(fantasyDist.PrimaryBiomes) {
		t.Error("Unknown genre should default to fantasy biome distribution")
	}
}

func TestSelectBiomeFromWeights(t *testing.T) {
	dist := &GenreBiomeDistribution{
		PrimaryBiomes: []BiomeType{BiomeForest, BiomeMountain},
		Weights: map[BiomeType]float64{
			BiomeForest:   0.6,
			BiomeMountain: 0.4,
		},
	}

	// Test determinism - same seed should give same result
	seed := int64(12345)
	biome1 := selectBiomeFromWeights(seed, dist)
	biome2 := selectBiomeFromWeights(seed, dist)

	if biome1 != biome2 {
		t.Errorf("selectBiomeFromWeights not deterministic: %d vs %d", biome1, biome2)
	}
}

func TestDetermineBiome(t *testing.T) {
	dist := GetGenreBiomeDistribution("fantasy")

	tests := []struct {
		name        string
		tileType    int
		height      int
		primary     BiomeType
		expectBiome BiomeType
	}{
		{"floor uses primary", 1, 0, BiomeForest, BiomeForest},
		{"high ground mountain", 1, 2, BiomeForest, BiomeMountain},
		{"water becomes lake", 4, -1, BiomeForest, BiomeLake}, // TileWaterShallow = 4
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineBiome(tt.tileType, tt.height, tt.primary, dist)
			if result != tt.expectBiome {
				t.Errorf("determineBiome() = %d, want %d", result, tt.expectBiome)
			}
		})
	}
}

func TestGenreBiomeMapping(t *testing.T) {
	// Verify expected biomes per genre from ROADMAP.md
	genreExpectations := map[string][]BiomeType{
		"fantasy":          {BiomeForest, BiomeMountain, BiomeLake},
		"sci-fi":           {BiomeCrater, BiomeTech},
		"horror":           {BiomeSwamp, BiomeForest},
		"cyberpunk":        {BiomeUrban, BiomeIndustrial},
		"post-apocalyptic": {BiomeWasteland, BiomeRuins},
	}

	for genre, expectedBiomes := range genreExpectations {
		t.Run(genre, func(t *testing.T) {
			dist := GetGenreBiomeDistribution(genre)
			for _, expected := range expectedBiomes {
				found := false
				for _, primary := range dist.PrimaryBiomes {
					if primary == expected {
						found = true
						break
					}
				}
				if !found {
					for _, secondary := range dist.SecondaryBiomes {
						if secondary == expected {
							found = true
							break
						}
					}
				}
				if !found {
					// Check if it's in the weights
					if _, ok := dist.Weights[expected]; ok {
						found = true
					}
				}
				if !found {
					t.Errorf("Genre %s missing expected biome %d", genre, expected)
				}
			}
		})
	}
}
