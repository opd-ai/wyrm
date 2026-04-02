// Package geom provides geometry utilities for 2D calculations.
package geom

// PointInPolygon tests if point (x, y) is inside a polygon defined by vertices.
// The vertices slice contains alternating x,y coordinates: [x0, y0, x1, y1, ...].
// Uses the ray casting algorithm for determining point-in-polygon membership.
func PointInPolygon(x, y float64, vertices []float64) bool {
	if len(vertices) < 6 {
		return false
	}

	inside := false
	n := len(vertices) / 2
	j := n - 1

	for i := 0; i < n; i++ {
		xi, yi := vertices[i*2], vertices[i*2+1]
		xj, yj := vertices[j*2], vertices[j*2+1]

		if ((yi > y) != (yj > y)) && (x < (xj-xi)*(y-yi)/(yj-yi)+xi) {
			inside = !inside
		}
		j = i
	}

	return inside
}
