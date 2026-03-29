// Package city provides procedural city generation.
package city

import (
	"fmt"
	"math/rand"
)

// City represents a procedurally generated city.
type City struct {
	Name      string
	Districts []District
	Seed      int64
	Genre     string
}

// District represents a city district.
type District struct {
	Name      string
	CenterX   float64
	CenterY   float64
	Buildings int
	Type      string
}

// genreNames holds city name prefixes and suffixes by genre.
var genreNames = map[string]struct {
	prefixes []string
	suffixes []string
}{
	"fantasy": {
		prefixes: []string{"Elder", "Golden", "Silver", "Iron", "Dragon", "Storm", "Shadow", "Dawn"},
		suffixes: []string{"hold", "haven", "reach", "fall", "gate", "forge", "keep", "spire"},
	},
	"sci-fi": {
		prefixes: []string{"Neo", "Nova", "Alpha", "Omega", "Stellar", "Quantum", "Helix", "Nexus"},
		suffixes: []string{"prime", "station", "port", "dome", "arc", "core", "hub", "node"},
	},
	"horror": {
		prefixes: []string{"Hollow", "Bleak", "Silent", "Ashen", "Cursed", "Withered", "Pale", "Dread"},
		suffixes: []string{"moor", "vale", "marsh", "hollow", "crypt", "shade", "glen", "barrow"},
	},
	"cyberpunk": {
		prefixes: []string{"Neon", "Chrome", "Data", "Grid", "Cyber", "Synth", "Wire", "Volt"},
		suffixes: []string{"sprawl", "sector", "block", "zone", "grid", "strip", "stack", "layer"},
	},
	"post-apocalyptic": {
		prefixes: []string{"Rust", "Scrap", "Bone", "Dust", "Rad", "Ash", "Iron", "Salt"},
		suffixes: []string{"town", "camp", "fort", "haven", "pit", "wastes", "hold", "ruins"},
	},
}

// genreDistricts holds district types by genre.
var genreDistricts = map[string][]string{
	"fantasy":          {"Market", "Temple", "Castle", "Guild", "Harbor", "Slums", "Noble", "Mage"},
	"sci-fi":           {"Residential", "Industrial", "Commercial", "Military", "Research", "Transit", "Medical", "Admin"},
	"horror":           {"Old Town", "Cemetery", "Manor", "Asylum", "Church", "Docks", "Ruins", "Catacombs"},
	"cyberpunk":        {"Corporate", "Residential", "Industrial", "Entertainment", "Black Market", "Slums", "Tech", "Enforcement"},
	"post-apocalyptic": {"Salvage", "Farm", "Trade", "Shelter", "Water", "Armory", "Med Bay", "Wasteland"},
}

// Generate creates a new procedural city from the given seed and genre.
func Generate(seed int64, genre string) *City {
	rng := rand.New(rand.NewSource(seed))

	// Default to fantasy if unknown genre
	names, ok := genreNames[genre]
	if !ok {
		names = genreNames["fantasy"]
	}
	districts, ok := genreDistricts[genre]
	if !ok {
		districts = genreDistricts["fantasy"]
	}

	// Generate city name
	prefix := names.prefixes[rng.Intn(len(names.prefixes))]
	suffix := names.suffixes[rng.Intn(len(names.suffixes))]
	cityName := prefix + suffix

	// Generate 3-6 districts
	numDistricts := 3 + rng.Intn(4)
	cityDistricts := make([]District, numDistricts)

	for i := 0; i < numDistricts; i++ {
		districtType := districts[rng.Intn(len(districts))]
		cityDistricts[i] = District{
			Name:      fmt.Sprintf("%s %s", cityName, districtType),
			CenterX:   rng.Float64()*1000 - 500,
			CenterY:   rng.Float64()*1000 - 500,
			Buildings: 10 + rng.Intn(41),
			Type:      districtType,
		}
	}

	return &City{
		Name:      cityName,
		Districts: cityDistricts,
		Seed:      seed,
		Genre:     genre,
	}
}
