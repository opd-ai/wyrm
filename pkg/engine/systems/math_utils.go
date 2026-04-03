package systems

// clampFloat64 limits a value to the [minV, maxV] range.
// This is a package-level utility to consolidate duplicated clamp functions.
func clampFloat64(v, minV, maxV float64) float64 {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}
