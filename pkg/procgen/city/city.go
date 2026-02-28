// Package city provides procedural city generation.
package city

// City represents a procedurally generated city.
type City struct {
	Name      string
	Districts []District
	Seed      int64
}

// District represents a city district.
type District struct {
	Name     string
	CenterX  float64
	CenterY  float64
	Buildings int
}

// Generate creates a new procedural city from the given seed and genre.
func Generate(seed int64, genre string) *City {
	return &City{
		Name: "Unnamed City",
		Seed: seed,
	}
}
