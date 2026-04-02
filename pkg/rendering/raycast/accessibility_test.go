//go:build noebiten

package raycast

import (
	"image/color"
	"testing"
)

func TestDefaultAccessibilityConfig(t *testing.T) {
	cfg := DefaultAccessibilityConfig()
	if cfg == nil {
		t.Fatal("DefaultAccessibilityConfig() returned nil")
	}

	// Default should be normal vision
	if cfg.ColorblindMode != ColorblindNone {
		t.Errorf("Default ColorblindMode = %d, want ColorblindNone", cfg.ColorblindMode)
	}
	if cfg.HighContrastScheme != HighContrastNone {
		t.Errorf("Default HighContrastScheme = %d, want HighContrastNone", cfg.HighContrastScheme)
	}
	if cfg.HighlightBrightness != 1.0 {
		t.Errorf("Default HighlightBrightness = %f, want 1.0", cfg.HighlightBrightness)
	}
}

func TestHighContrastAccessibilityConfig(t *testing.T) {
	cfg := HighContrastAccessibilityConfig()
	if cfg == nil {
		t.Fatal("HighContrastAccessibilityConfig() returned nil")
	}

	if cfg.HighContrastScheme != HighContrastYellowBlack {
		t.Errorf("HighContrast scheme = %d, want YellowBlack", cfg.HighContrastScheme)
	}
	if !cfg.LargerInteractionHighlights {
		t.Error("HighContrast should have larger interaction highlights")
	}
	if cfg.HighlightBrightness <= 1.0 {
		t.Error("HighContrast should have increased brightness")
	}
	if !cfg.FlashingDisabled {
		t.Error("HighContrast should disable flashing")
	}
}

func TestColorblindFriendlyConfig(t *testing.T) {
	modes := []ColorblindMode{
		ColorblindProtanopia,
		ColorblindDeuteranopia,
		ColorblindTritanopia,
		ColorblindAchromatopsia,
	}

	for _, mode := range modes {
		cfg := ColorblindFriendlyConfig(mode)
		if cfg == nil {
			t.Errorf("ColorblindFriendlyConfig(%d) returned nil", mode)
			continue
		}
		if cfg.ColorblindMode != mode {
			t.Errorf("Config mode = %d, want %d", cfg.ColorblindMode, mode)
		}
		// Should have pattern fills for better differentiation
		if !cfg.UsePatternFills {
			t.Errorf("Colorblind config should enable pattern fills")
		}
	}
}

func TestAccessibleColorGetColor(t *testing.T) {
	ac := AccessibleColor{
		Base:         color.RGBA{255, 0, 0, 255},
		Protanopia:   color.RGBA{255, 150, 0, 255},
		Deuteranopia: color.RGBA{255, 180, 0, 255},
		Tritanopia:   color.RGBA{255, 100, 100, 255},
		Achromat:     color.RGBA{100, 100, 100, 255},
	}

	tests := []struct {
		mode ColorblindMode
		want color.RGBA
	}{
		{ColorblindNone, ac.Base},
		{ColorblindProtanopia, ac.Protanopia},
		{ColorblindDeuteranopia, ac.Deuteranopia},
		{ColorblindTritanopia, ac.Tritanopia},
		{ColorblindAchromatopsia, ac.Achromat},
	}

	for _, tt := range tests {
		got := ac.GetColor(tt.mode)
		if got != tt.want {
			t.Errorf("GetColor(%d) = %v, want %v", tt.mode, got, tt.want)
		}
	}
}

func TestApplyBrightness(t *testing.T) {
	base := color.RGBA{100, 100, 100, 255}

	// Double brightness
	bright := ApplyBrightness(base, 2.0)
	if bright.R != 200 || bright.G != 200 || bright.B != 200 {
		t.Errorf("2x brightness = %v, want {200,200,200,255}", bright)
	}

	// Half brightness
	dim := ApplyBrightness(base, 0.5)
	if dim.R != 50 || dim.G != 50 || dim.B != 50 {
		t.Errorf("0.5x brightness = %v, want {50,50,50,255}", dim)
	}

	// Clamping at 255
	bright = ApplyBrightness(color.RGBA{200, 200, 200, 255}, 2.0)
	if bright.R != 255 || bright.G != 255 || bright.B != 255 {
		t.Errorf("Clamped brightness = %v, want {255,255,255,255}", bright)
	}

	// Zero brightness
	zero := ApplyBrightness(base, 0.0)
	if zero.R != 0 || zero.G != 0 || zero.B != 0 {
		t.Errorf("0x brightness = %v, want {0,0,0,255}", zero)
	}
}

func TestApplyHighContrast(t *testing.T) {
	light := color.RGBA{200, 200, 200, 255}
	dark := color.RGBA{50, 50, 50, 255}

	// Black & White
	bwLight := ApplyHighContrast(light, HighContrastBlackWhite)
	if bwLight.R != 255 || bwLight.G != 255 || bwLight.B != 255 {
		t.Errorf("B&W light = %v, want white", bwLight)
	}
	bwDark := ApplyHighContrast(dark, HighContrastBlackWhite)
	if bwDark.R != 0 || bwDark.G != 0 || bwDark.B != 0 {
		t.Errorf("B&W dark = %v, want black", bwDark)
	}

	// Yellow on Black
	ybLight := ApplyHighContrast(light, HighContrastYellowBlack)
	if ybLight.R != 255 || ybLight.G != 255 || ybLight.B != 0 {
		t.Errorf("Y/B light = %v, want yellow", ybLight)
	}

	// None should pass through unchanged
	unchanged := ApplyHighContrast(light, HighContrastNone)
	if unchanged != light {
		t.Errorf("None scheme changed color: %v != %v", unchanged, light)
	}
}

func TestGetAccessibleHighlightColor(t *testing.T) {
	palette := DefaultAccessiblePalette()

	tests := []struct {
		entityType string
		mode       ColorblindMode
	}{
		{"weapon", ColorblindNone},
		{"armor", ColorblindProtanopia},
		{"consumable", ColorblindDeuteranopia},
		{"quest", ColorblindTritanopia},
		{"hostile", ColorblindAchromatopsia},
	}

	for _, tt := range tests {
		cfg := DefaultAccessibilityConfig()
		cfg.ColorblindMode = tt.mode
		c := GetAccessibleHighlightColor(tt.entityType, cfg, palette)
		// Should return a non-zero color
		if c.A == 0 {
			t.Errorf("GetAccessibleHighlightColor(%q, %d) returned transparent", tt.entityType, tt.mode)
		}
	}
}

func TestAccessibilityConfigMethods(t *testing.T) {
	cfg := DefaultAccessibilityConfig()

	// Default outline
	if cfg.ShouldShowOutline() != (cfg.OutlineThickness > 0) {
		t.Error("ShouldShowOutline mismatch")
	}

	cfg.LargerInteractionHighlights = true
	if !cfg.ShouldShowOutline() {
		t.Error("LargerInteractionHighlights should enable outline")
	}

	// Larger highlights should increase outline thickness
	thickness := cfg.GetEffectiveOutlineThickness()
	if thickness < 2 {
		t.Errorf("With larger highlights, thickness = %d, want >= 2", thickness)
	}

	// Indicator size
	baseSize := cfg.InteractionIndicatorSize
	effectiveSize := cfg.GetEffectiveIndicatorSize()
	if effectiveSize <= baseSize {
		t.Error("Larger highlights should increase indicator size")
	}

	// Pulse disabled
	cfg.FlashingDisabled = true
	if cfg.ShouldUsePulse() {
		t.Error("Pulse should be disabled when FlashingDisabled is true")
	}
}

func TestColorblindModeName(t *testing.T) {
	tests := []struct {
		mode     ColorblindMode
		expected string
	}{
		{ColorblindNone, "Normal"},
		{ColorblindProtanopia, "Protanopia (Red-blind)"},
		{ColorblindDeuteranopia, "Deuteranopia (Green-blind)"},
		{ColorblindTritanopia, "Tritanopia (Blue-blind)"},
		{ColorblindAchromatopsia, "Achromatopsia (No color)"},
	}

	for _, tt := range tests {
		got := ColorblindModeName(tt.mode)
		if got != tt.expected {
			t.Errorf("ColorblindModeName(%d) = %q, want %q", tt.mode, got, tt.expected)
		}
	}
}

func TestHighContrastSchemeName(t *testing.T) {
	tests := []struct {
		scheme   HighContrastScheme
		expected string
	}{
		{HighContrastNone, "Normal"},
		{HighContrastBlackWhite, "Black & White"},
		{HighContrastYellowBlack, "Yellow on Black"},
		{HighContrastWhiteBlack, "White on Black"},
	}

	for _, tt := range tests {
		got := HighContrastSchemeName(tt.scheme)
		if got != tt.expected {
			t.Errorf("HighContrastSchemeName(%d) = %q, want %q", tt.scheme, got, tt.expected)
		}
	}
}

func TestDefaultAccessiblePalette(t *testing.T) {
	palette := DefaultAccessiblePalette()
	if palette == nil {
		t.Fatal("DefaultAccessiblePalette() returned nil")
	}

	// Check that all colors have non-zero alpha
	colors := []AccessibleColor{
		palette.ItemHighlight,
		palette.WeaponHighlight,
		palette.ArmorHighlight,
		palette.ConsumHighlight,
		palette.QuestHighlight,
		palette.KeyHighlight,
		palette.TreasureHighlight,
		palette.InteractHighlight,
		palette.LockedHighlight,
		palette.DangerHighlight,
		palette.FriendlyHighlight,
		palette.NeutralHighlight,
		palette.HostileHighlight,
	}

	for i, ac := range colors {
		if ac.Base.A == 0 {
			t.Errorf("Palette color %d has zero alpha", i)
		}
		// Verify colorblind variants exist
		if ac.Protanopia.A == 0 {
			t.Errorf("Palette color %d has zero Protanopia alpha", i)
		}
		if ac.Deuteranopia.A == 0 {
			t.Errorf("Palette color %d has zero Deuteranopia alpha", i)
		}
	}
}

func TestRendererAccessibilityConfig(t *testing.T) {
	r := NewRenderer(320, 240)

	// Default config
	cfg := r.GetAccessibilityConfig()
	if cfg == nil {
		t.Fatal("GetAccessibilityConfig should not return nil")
	}

	// Set custom config
	custom := HighContrastAccessibilityConfig()
	r.SetAccessibilityConfig(custom)
	if r.GetAccessibilityConfig() != custom {
		t.Error("SetAccessibilityConfig should update the config")
	}

	// Set colorblind mode
	r.SetColorblindMode(ColorblindProtanopia)
	if r.GetAccessibilityConfig().ColorblindMode != ColorblindProtanopia {
		t.Error("SetColorblindMode should update the mode")
	}

	// Set high contrast scheme
	r.SetHighContrastScheme(HighContrastYellowBlack)
	if r.GetAccessibilityConfig().HighContrastScheme != HighContrastYellowBlack {
		t.Error("SetHighContrastScheme should update the scheme")
	}
}

func TestEntityTypeColorMapping(t *testing.T) {
	cfg := DefaultAccessibilityConfig()
	palette := DefaultAccessiblePalette()

	// Entity types should map to different colors
	entityTypes := []string{
		"weapon", "armor", "consumable", "quest", "key",
		"treasure", "locked", "danger", "friendly", "hostile", "neutral",
	}

	colors := make(map[color.RGBA]string)
	for _, et := range entityTypes {
		c := GetAccessibleHighlightColor(et, cfg, palette)
		if existing, ok := colors[c]; ok {
			// Some overlap is acceptable (e.g., similar items)
			// But danger and friendly should not overlap
			if (et == "danger" && existing == "friendly") ||
				(et == "friendly" && existing == "danger") {
				t.Errorf("Danger and friendly have same color: %v", c)
			}
		}
		colors[c] = et
	}
}

func BenchmarkGetAccessibleHighlightColor(b *testing.B) {
	cfg := DefaultAccessibilityConfig()
	palette := DefaultAccessiblePalette()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetAccessibleHighlightColor("weapon", cfg, palette)
	}
}

func BenchmarkApplyBrightness(b *testing.B) {
	c := color.RGBA{100, 150, 200, 255}
	for i := 0; i < b.N; i++ {
		ApplyBrightness(c, 1.5)
	}
}

func BenchmarkApplyHighContrast(b *testing.B) {
	c := color.RGBA{100, 150, 200, 255}
	for i := 0; i < b.N; i++ {
		ApplyHighContrast(c, HighContrastYellowBlack)
	}
}
