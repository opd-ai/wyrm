package geom

import "testing"

func TestPointInPolygon(t *testing.T) {
	// Simple square: (0,0), (4,0), (4,4), (0,4)
	square := []float64{0, 0, 4, 0, 4, 4, 0, 4}

	tests := []struct {
		name     string
		x, y     float64
		vertices []float64
		want     bool
	}{
		{"center of square", 2, 2, square, true},
		{"inside near edge", 0.1, 0.1, square, true},
		{"outside left", -1, 2, square, false},
		{"outside right", 5, 2, square, false},
		{"outside top", 2, 5, square, false},
		{"outside bottom", 2, -1, square, false},
		{"on edge", 0, 2, square, true}, // ray-cast can include some edge points
		{"too few vertices", 2, 2, []float64{0, 0, 1, 1}, false},
		{"empty vertices", 2, 2, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PointInPolygon(tt.x, tt.y, tt.vertices)
			if got != tt.want {
				t.Errorf("PointInPolygon(%v, %v) = %v, want %v", tt.x, tt.y, got, tt.want)
			}
		})
	}
}

func TestPointInPolygon_Triangle(t *testing.T) {
	// Triangle: (0,0), (4,0), (2,4)
	triangle := []float64{0, 0, 4, 0, 2, 4}

	if !PointInPolygon(2, 1, triangle) {
		t.Error("Point (2,1) should be inside triangle")
	}
	if PointInPolygon(0, 4, triangle) {
		t.Error("Point (0,4) should be outside triangle")
	}
}
