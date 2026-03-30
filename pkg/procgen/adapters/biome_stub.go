//go:build noebiten

// Package adapters provides V-Series integration for Wyrm.
// This stub file provides biome distribution for headless testing.
package adapters

// BiomeType represents different biome categories.
type BiomeType int

const (
	// BiomeForest represents forested areas.
	BiomeForest BiomeType = iota
	// BiomeMountain represents mountainous terrain.
	BiomeMountain
	// BiomeLake represents water bodies.
	BiomeLake
	// BiomeSwamp represents swampy wetlands.
	BiomeSwamp
	// BiomeWasteland represents barren wasteland.
	BiomeWasteland
	// BiomeUrban represents urban/city areas.
	BiomeUrban
	// BiomeIndustrial represents industrial zones.
	BiomeIndustrial
	// BiomeRuins represents ruined structures.
	BiomeRuins
	// BiomeCrater represents impact craters.
	BiomeCrater
	// BiomeTech represents high-tech structures.
	BiomeTech
)

// GenreBiomeDistribution defines biome weights for each genre.
type GenreBiomeDistribution struct {
	PrimaryBiomes   []BiomeType
	SecondaryBiomes []BiomeType
	Weights         map[BiomeType]float64
}

// genreBiomeDistributions maps genre to biome distribution.
var genreBiomeDistributions = map[string]*GenreBiomeDistribution{
	"fantasy": {
		PrimaryBiomes:   []BiomeType{BiomeForest, BiomeMountain, BiomeLake},
		SecondaryBiomes: []BiomeType{BiomeRuins},
		Weights: map[BiomeType]float64{
			BiomeForest:   0.5,
			BiomeMountain: 0.3,
			BiomeLake:     0.2,
		},
	},
	"sci-fi": {
		PrimaryBiomes:   []BiomeType{BiomeCrater, BiomeTech},
		SecondaryBiomes: []BiomeType{BiomeWasteland},
		Weights: map[BiomeType]float64{
			BiomeCrater: 0.4,
			BiomeTech:   0.6,
		},
	},
	"horror": {
		PrimaryBiomes:   []BiomeType{BiomeSwamp, BiomeForest},
		SecondaryBiomes: []BiomeType{BiomeRuins},
		Weights: map[BiomeType]float64{
			BiomeSwamp:  0.5,
			BiomeForest: 0.5,
		},
	},
	"cyberpunk": {
		PrimaryBiomes:   []BiomeType{BiomeUrban, BiomeIndustrial},
		SecondaryBiomes: []BiomeType{BiomeTech},
		Weights: map[BiomeType]float64{
			BiomeUrban:      0.6,
			BiomeIndustrial: 0.4,
		},
	},
	"post-apocalyptic": {
		PrimaryBiomes:   []BiomeType{BiomeWasteland, BiomeRuins},
		SecondaryBiomes: []BiomeType{BiomeUrban},
		Weights: map[BiomeType]float64{
			BiomeWasteland: 0.6,
			BiomeRuins:     0.4,
		},
	},
}

// GetGenreBiomeDistribution returns biome distribution for a genre.
func GetGenreBiomeDistribution(genre string) *GenreBiomeDistribution {
	if dist, ok := genreBiomeDistributions[genre]; ok {
		return dist
	}
	return genreBiomeDistributions["fantasy"]
}
