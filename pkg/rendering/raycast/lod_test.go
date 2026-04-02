//go:build noebiten

package raycast

import (
	"testing"
)

func TestLODLevelThresholds(t *testing.T) {
	cfg := DefaultLODConfig()

	tests := []struct {
		name     string
		distance float64
		want     LODLevel
	}{
		{"very close", 1.0, LODFull},
		{"close", 4.9, LODFull},
		{"at high threshold", 5.0, LODHigh},
		{"medium distance", 7.0, LODHigh},
		{"at medium threshold", 10.0, LODMedium},
		{"far", 15.0, LODMedium},
		{"at low threshold", 20.0, LODLow},
		{"very far", 50.0, LODLow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cfg.GetLODLevel(tt.distance)
			if got != tt.want {
				t.Errorf("GetLODLevel(%f) = %d, want %d", tt.distance, got, tt.want)
			}
		})
	}
}

func TestLODDisabled(t *testing.T) {
	cfg := DefaultLODConfig()
	cfg.Enabled = false

	// When disabled, always return LODFull regardless of distance
	tests := []float64{1.0, 10.0, 50.0, 100.0}
	for _, dist := range tests {
		got := cfg.GetLODLevel(dist)
		if got != LODFull {
			t.Errorf("GetLODLevel(%f) with disabled LOD = %d, want %d", dist, got, LODFull)
		}
	}
}

func TestLODRenderFlags(t *testing.T) {
	cfg := DefaultLODConfig()

	// At close distance, all features enabled
	closeFlags := cfg.GetRenderFlags(2.0)
	if !closeFlags.RenderHighlights {
		t.Error("Close distance should render highlights")
	}
	if !closeFlags.RenderNormalMaps {
		t.Error("Close distance should render normal maps")
	}
	if closeFlags.SimplifiedSprites {
		t.Error("Close distance should not simplify sprites")
	}
	if closeFlags.Level != LODFull {
		t.Errorf("Close distance level = %d, want %d", closeFlags.Level, LODFull)
	}

	// At medium distance, some features disabled
	mediumFlags := cfg.GetRenderFlags(12.0)
	if mediumFlags.RenderNormalMaps {
		t.Error("Medium distance should skip normal maps")
	}
	if !mediumFlags.SimplifiedSprites {
		t.Error("Medium distance should simplify sprites")
	}

	// At far distance, minimal features
	farFlags := cfg.GetRenderFlags(25.0)
	if farFlags.RenderHighlights {
		t.Error("Far distance should skip highlights")
	}
	if farFlags.RenderNormalMaps {
		t.Error("Far distance should skip normal maps")
	}
}

func TestColumnSkipForLOD(t *testing.T) {
	tests := []struct {
		lod  LODLevel
		want int
	}{
		{LODFull, 1},
		{LODHigh, 1},
		{LODMedium, 2},
		{LODLow, 4},
	}

	for _, tt := range tests {
		got := ColumnSkipForLOD(tt.lod)
		if got != tt.want {
			t.Errorf("ColumnSkipForLOD(%d) = %d, want %d", tt.lod, got, tt.want)
		}
	}
}

func TestRowSkipForLOD(t *testing.T) {
	tests := []struct {
		lod  LODLevel
		want int
	}{
		{LODFull, 1},
		{LODHigh, 1},
		{LODMedium, 2},
		{LODLow, 3},
	}

	for _, tt := range tests {
		got := RowSkipForLOD(tt.lod)
		if got != tt.want {
			t.Errorf("RowSkipForLOD(%d) = %d, want %d", tt.lod, got, tt.want)
		}
	}
}

func TestRendererLODConfig(t *testing.T) {
	r := NewRenderer(320, 240)

	// Default config should be used
	cfg := r.GetLODConfig()
	if cfg == nil {
		t.Fatal("GetLODConfig should not return nil")
	}
	if !cfg.Enabled {
		t.Error("Default LOD config should be enabled")
	}

	// Custom config
	custom := &LODConfig{
		Enabled:         false,
		HighThreshold:   10.0,
		MediumThreshold: 20.0,
		LowThreshold:    30.0,
	}
	r.SetLODConfig(custom)
	if r.GetLODConfig() != custom {
		t.Error("SetLODConfig should update the config")
	}
}

func TestRendererGetLODLevel(t *testing.T) {
	r := NewRenderer(320, 240)

	// Test through renderer
	lod := r.GetLODLevel(1.0)
	if lod != LODFull {
		t.Errorf("GetLODLevel(1.0) = %d, want %d", lod, LODFull)
	}

	lod = r.GetLODLevel(25.0)
	if lod != LODLow {
		t.Errorf("GetLODLevel(25.0) = %d, want %d", lod, LODLow)
	}
}

func TestLODHighlightControl(t *testing.T) {
	cfg := DefaultLODConfig()

	// By default, skip highlights at LODLow
	if cfg.ShouldRenderHighlight(LODFull) != true {
		t.Error("Should render highlights at LODFull")
	}
	if cfg.ShouldRenderHighlight(LODHigh) != true {
		t.Error("Should render highlights at LODHigh")
	}
	if cfg.ShouldRenderHighlight(LODMedium) != true {
		t.Error("Should render highlights at LODMedium")
	}
	if cfg.ShouldRenderHighlight(LODLow) != false {
		t.Error("Should skip highlights at LODLow")
	}
}

func TestLODNormalMapControl(t *testing.T) {
	cfg := DefaultLODConfig()

	// By default, skip normal maps at LODMedium and below
	if cfg.ShouldSkipNormalMaps(LODFull) != false {
		t.Error("Should render normal maps at LODFull")
	}
	if cfg.ShouldSkipNormalMaps(LODHigh) != false {
		t.Error("Should render normal maps at LODHigh")
	}
	if cfg.ShouldSkipNormalMaps(LODMedium) != true {
		t.Error("Should skip normal maps at LODMedium")
	}
	if cfg.ShouldSkipNormalMaps(LODLow) != true {
		t.Error("Should skip normal maps at LODLow")
	}
}

func TestLODSpriteSimplification(t *testing.T) {
	cfg := DefaultLODConfig()

	// By default, simplify sprites at LODMedium and below
	if cfg.ShouldSimplifySprite(LODFull) != false {
		t.Error("Should not simplify sprites at LODFull")
	}
	if cfg.ShouldSimplifySprite(LODHigh) != false {
		t.Error("Should not simplify sprites at LODHigh")
	}
	if cfg.ShouldSimplifySprite(LODMedium) != true {
		t.Error("Should simplify sprites at LODMedium")
	}
	if cfg.ShouldSimplifySprite(LODLow) != true {
		t.Error("Should simplify sprites at LODLow")
	}
}

func BenchmarkGetLODLevel(b *testing.B) {
	cfg := DefaultLODConfig()
	for i := 0; i < b.N; i++ {
		cfg.GetLODLevel(15.0)
	}
}

func BenchmarkGetRenderFlags(b *testing.B) {
	cfg := DefaultLODConfig()
	for i := 0; i < b.N; i++ {
		cfg.GetRenderFlags(15.0)
	}
}
