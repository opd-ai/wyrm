// Package raycast provides first-person raycasting rendering.
// This file contains accessibility features for rendering.
package raycast

import (
	"image/color"
)

// ColorblindMode represents different colorblind vision types.
type ColorblindMode int

const (
	// ColorblindNone is normal color vision.
	ColorblindNone ColorblindMode = iota
	// ColorblindProtanopia simulates red-blindness.
	ColorblindProtanopia
	// ColorblindDeuteranopia simulates green-blindness.
	ColorblindDeuteranopia
	// ColorblindTritanopia simulates blue-blindness.
	ColorblindTritanopia
	// ColorblindAchromatopsia simulates complete color blindness.
	ColorblindAchromatopsia
)

// HighContrastScheme defines high-contrast color schemes.
type HighContrastScheme int

const (
	// HighContrastNone uses normal colors.
	HighContrastNone HighContrastScheme = iota
	// HighContrastBlackWhite uses black and white only.
	HighContrastBlackWhite
	// HighContrastYellowBlack uses yellow on black.
	HighContrastYellowBlack
	// HighContrastWhiteBlack uses white on black.
	HighContrastWhiteBlack
)

// AccessibilityConfig contains settings for accessible rendering.
type AccessibilityConfig struct {
	// ColorblindMode affects color choices for UI elements.
	ColorblindMode ColorblindMode

	// HighContrastScheme for improved visibility.
	HighContrastScheme HighContrastScheme

	// LargerInteractionHighlights increases highlight size.
	LargerInteractionHighlights bool

	// HighlightBrightness multiplier (1.0 = normal, 2.0 = twice as bright).
	HighlightBrightness float64

	// OutlineThickness for highlighted objects (pixels).
	OutlineThickness int

	// FlashingDisabled prevents flashing/pulsing effects.
	FlashingDisabled bool

	// ScreenShakeDisabled prevents screen shake effects.
	ScreenShakeDisabled bool

	// MotionBlurDisabled prevents motion blur effects.
	MotionBlurDisabled bool

	// InteractionIndicatorSize for targeting reticle (pixels).
	InteractionIndicatorSize int

	// UsePatternFills adds patterns to colored areas.
	UsePatternFills bool

	// TextScaling multiplier for text size.
	TextScaling float64
}

// DefaultAccessibilityConfig returns default accessibility settings.
func DefaultAccessibilityConfig() *AccessibilityConfig {
	return &AccessibilityConfig{
		ColorblindMode:              ColorblindNone,
		HighContrastScheme:          HighContrastNone,
		LargerInteractionHighlights: false,
		HighlightBrightness:         1.0,
		OutlineThickness:            1,
		FlashingDisabled:            false,
		ScreenShakeDisabled:         false,
		MotionBlurDisabled:          false,
		InteractionIndicatorSize:    8,
		UsePatternFills:             false,
		TextScaling:                 1.0,
	}
}

// HighContrastAccessibilityConfig returns settings for visually impaired users.
func HighContrastAccessibilityConfig() *AccessibilityConfig {
	return &AccessibilityConfig{
		ColorblindMode:              ColorblindNone,
		HighContrastScheme:          HighContrastYellowBlack,
		LargerInteractionHighlights: true,
		HighlightBrightness:         2.0,
		OutlineThickness:            3,
		FlashingDisabled:            true,
		ScreenShakeDisabled:         true,
		MotionBlurDisabled:          true,
		InteractionIndicatorSize:    16,
		UsePatternFills:             true,
		TextScaling:                 1.5,
	}
}

// ColorblindFriendlyConfig returns settings optimized for colorblind users.
func ColorblindFriendlyConfig(mode ColorblindMode) *AccessibilityConfig {
	return &AccessibilityConfig{
		ColorblindMode:              mode,
		HighContrastScheme:          HighContrastNone,
		LargerInteractionHighlights: true,
		HighlightBrightness:         1.3,
		OutlineThickness:            2,
		FlashingDisabled:            false,
		ScreenShakeDisabled:         false,
		MotionBlurDisabled:          false,
		InteractionIndicatorSize:    12,
		UsePatternFills:             true,
		TextScaling:                 1.0,
	}
}

// AccessibleColor represents a color with accessibility information.
type AccessibleColor struct {
	// Base color for normal vision.
	Base color.RGBA

	// Protanopia-safe alternative.
	Protanopia color.RGBA

	// Deuteranopia-safe alternative.
	Deuteranopia color.RGBA

	// Tritanopia-safe alternative.
	Tritanopia color.RGBA

	// Achromat alternative (high contrast grayscale).
	Achromat color.RGBA

	// Pattern identifier for pattern fills (0 = none).
	PatternID int
}

// AccessibleColorPalette provides colorblind-safe colors for game elements.
type AccessibleColorPalette struct {
	// Item colors
	ItemHighlight     AccessibleColor
	WeaponHighlight   AccessibleColor
	ArmorHighlight    AccessibleColor
	ConsumHighlight   AccessibleColor
	QuestHighlight    AccessibleColor
	KeyHighlight      AccessibleColor
	TreasureHighlight AccessibleColor

	// Interaction colors
	InteractHighlight AccessibleColor
	LockedHighlight   AccessibleColor
	DangerHighlight   AccessibleColor
	FriendlyHighlight AccessibleColor
	NeutralHighlight  AccessibleColor
	HostileHighlight  AccessibleColor

	// UI colors
	SelectionBorder AccessibleColor
	FocusBorder     AccessibleColor
	ErrorColor      AccessibleColor
	SuccessColor    AccessibleColor
}

// DefaultAccessiblePalette returns a colorblind-safe color palette.
func DefaultAccessiblePalette() *AccessibleColorPalette {
	return &AccessibleColorPalette{
		// Items use distinct hues that work across colorblind types
		ItemHighlight: AccessibleColor{
			Base:         color.RGBA{255, 255, 100, 255}, // Yellow
			Protanopia:   color.RGBA{255, 255, 100, 255}, // Yellow works for all
			Deuteranopia: color.RGBA{255, 255, 100, 255},
			Tritanopia:   color.RGBA{255, 200, 100, 255}, // Shift to orange for tritanopia
			Achromat:     color.RGBA{255, 255, 255, 255}, // White
			PatternID:    1,
		},
		WeaponHighlight: AccessibleColor{
			Base:         color.RGBA{255, 100, 100, 255}, // Red
			Protanopia:   color.RGBA{255, 180, 100, 255}, // Orange for protanopia
			Deuteranopia: color.RGBA{255, 180, 100, 255}, // Orange for deuteranopia
			Tritanopia:   color.RGBA{255, 100, 100, 255}, // Red works
			Achromat:     color.RGBA{200, 200, 200, 255}, // Light gray
			PatternID:    2,
		},
		ArmorHighlight: AccessibleColor{
			Base:         color.RGBA{100, 150, 255, 255}, // Blue
			Protanopia:   color.RGBA{100, 150, 255, 255}, // Blue works
			Deuteranopia: color.RGBA{100, 150, 255, 255}, // Blue works
			Tritanopia:   color.RGBA{150, 255, 150, 255}, // Green for tritanopia
			Achromat:     color.RGBA{150, 150, 150, 255}, // Medium gray
			PatternID:    3,
		},
		ConsumHighlight: AccessibleColor{
			Base:         color.RGBA{100, 255, 100, 255}, // Green
			Protanopia:   color.RGBA{100, 255, 200, 255}, // Cyan for protanopia
			Deuteranopia: color.RGBA{100, 255, 200, 255}, // Cyan for deuteranopia
			Tritanopia:   color.RGBA{255, 150, 255, 255}, // Pink for tritanopia
			Achromat:     color.RGBA{100, 100, 100, 255}, // Dark gray
			PatternID:    4,
		},
		QuestHighlight: AccessibleColor{
			Base:         color.RGBA{255, 200, 100, 255}, // Gold/Orange
			Protanopia:   color.RGBA{255, 255, 150, 255}, // Light yellow
			Deuteranopia: color.RGBA{255, 255, 150, 255}, // Light yellow
			Tritanopia:   color.RGBA{255, 200, 100, 255}, // Orange works
			Achromat:     color.RGBA{230, 230, 230, 255}, // Very light gray
			PatternID:    5,
		},
		KeyHighlight: AccessibleColor{
			Base:         color.RGBA{255, 150, 255, 255}, // Pink/Magenta
			Protanopia:   color.RGBA{150, 150, 255, 255}, // Light blue
			Deuteranopia: color.RGBA{150, 150, 255, 255}, // Light blue
			Tritanopia:   color.RGBA{255, 150, 255, 255}, // Pink works
			Achromat:     color.RGBA{180, 180, 180, 255}, // Gray
			PatternID:    6,
		},
		TreasureHighlight: AccessibleColor{
			Base:         color.RGBA{255, 215, 0, 255},   // Gold
			Protanopia:   color.RGBA{255, 215, 0, 255},   // Gold works
			Deuteranopia: color.RGBA{255, 215, 0, 255},   // Gold works
			Tritanopia:   color.RGBA{255, 180, 100, 255}, // Orange-gold
			Achromat:     color.RGBA{250, 250, 250, 255}, // Near white
			PatternID:    7,
		},

		// Interaction states
		InteractHighlight: AccessibleColor{
			Base:         color.RGBA{255, 255, 255, 255}, // White
			Protanopia:   color.RGBA{255, 255, 255, 255},
			Deuteranopia: color.RGBA{255, 255, 255, 255},
			Tritanopia:   color.RGBA{255, 255, 255, 255},
			Achromat:     color.RGBA{255, 255, 255, 255},
			PatternID:    0,
		},
		LockedHighlight: AccessibleColor{
			Base:         color.RGBA{150, 100, 50, 255},  // Brown
			Protanopia:   color.RGBA{150, 150, 100, 255}, // Olive
			Deuteranopia: color.RGBA{150, 150, 100, 255}, // Olive
			Tritanopia:   color.RGBA{150, 100, 50, 255},  // Brown works
			Achromat:     color.RGBA{100, 100, 100, 255}, // Dark gray
			PatternID:    8,
		},
		DangerHighlight: AccessibleColor{
			Base:         color.RGBA{255, 50, 50, 255},  // Bright red
			Protanopia:   color.RGBA{255, 200, 50, 255}, // Orange-yellow
			Deuteranopia: color.RGBA{255, 200, 50, 255}, // Orange-yellow
			Tritanopia:   color.RGBA{255, 50, 50, 255},  // Red works
			Achromat:     color.RGBA{50, 50, 50, 255},   // Very dark
			PatternID:    9,
		},
		FriendlyHighlight: AccessibleColor{
			Base:         color.RGBA{50, 200, 50, 255},   // Green
			Protanopia:   color.RGBA{50, 200, 255, 255},  // Cyan
			Deuteranopia: color.RGBA{50, 200, 255, 255},  // Cyan
			Tritanopia:   color.RGBA{200, 150, 255, 255}, // Light purple
			Achromat:     color.RGBA{200, 200, 200, 255}, // Light gray
			PatternID:    10,
		},
		NeutralHighlight: AccessibleColor{
			Base:         color.RGBA{200, 200, 200, 255}, // Gray
			Protanopia:   color.RGBA{200, 200, 200, 255},
			Deuteranopia: color.RGBA{200, 200, 200, 255},
			Tritanopia:   color.RGBA{200, 200, 200, 255},
			Achromat:     color.RGBA{200, 200, 200, 255},
			PatternID:    0,
		},
		HostileHighlight: AccessibleColor{
			Base:         color.RGBA{255, 0, 0, 255},   // Red
			Protanopia:   color.RGBA{255, 150, 0, 255}, // Orange
			Deuteranopia: color.RGBA{255, 150, 0, 255}, // Orange
			Tritanopia:   color.RGBA{255, 0, 0, 255},   // Red works
			Achromat:     color.RGBA{0, 0, 0, 255},     // Black
			PatternID:    11,
		},

		// UI elements
		SelectionBorder: AccessibleColor{
			Base:         color.RGBA{0, 200, 255, 255}, // Cyan
			Protanopia:   color.RGBA{0, 200, 255, 255},
			Deuteranopia: color.RGBA{0, 200, 255, 255},
			Tritanopia:   color.RGBA{255, 200, 0, 255}, // Yellow
			Achromat:     color.RGBA{255, 255, 255, 255},
			PatternID:    0,
		},
		FocusBorder: AccessibleColor{
			Base:         color.RGBA{255, 255, 0, 255}, // Yellow
			Protanopia:   color.RGBA{255, 255, 0, 255},
			Deuteranopia: color.RGBA{255, 255, 0, 255},
			Tritanopia:   color.RGBA{255, 200, 150, 255}, // Peach
			Achromat:     color.RGBA{255, 255, 255, 255},
			PatternID:    0,
		},
		ErrorColor: AccessibleColor{
			Base:         color.RGBA{255, 80, 80, 255}, // Red
			Protanopia:   color.RGBA{255, 180, 0, 255}, // Orange
			Deuteranopia: color.RGBA{255, 180, 0, 255}, // Orange
			Tritanopia:   color.RGBA{255, 80, 80, 255}, // Red works
			Achromat:     color.RGBA{80, 80, 80, 255},  // Dark gray
			PatternID:    12,
		},
		SuccessColor: AccessibleColor{
			Base:         color.RGBA{80, 255, 80, 255},   // Green
			Protanopia:   color.RGBA{80, 200, 255, 255},  // Cyan
			Deuteranopia: color.RGBA{80, 200, 255, 255},  // Cyan
			Tritanopia:   color.RGBA{200, 100, 255, 255}, // Purple
			Achromat:     color.RGBA{220, 220, 220, 255}, // Light gray
			PatternID:    13,
		},
	}
}

// GetColor returns the appropriate color for the given colorblind mode.
func (ac *AccessibleColor) GetColor(mode ColorblindMode) color.RGBA {
	switch mode {
	case ColorblindProtanopia:
		return ac.Protanopia
	case ColorblindDeuteranopia:
		return ac.Deuteranopia
	case ColorblindTritanopia:
		return ac.Tritanopia
	case ColorblindAchromatopsia:
		return ac.Achromat
	default:
		return ac.Base
	}
}

// ApplyBrightness returns the color with brightness adjustment.
func ApplyBrightness(c color.RGBA, brightness float64) color.RGBA {
	if brightness <= 0 {
		return color.RGBA{0, 0, 0, c.A}
	}

	r := float64(c.R) * brightness
	g := float64(c.G) * brightness
	b := float64(c.B) * brightness

	// Clamp to 255
	if r > 255 {
		r = 255
	}
	if g > 255 {
		g = 255
	}
	if b > 255 {
		b = 255
	}

	return color.RGBA{uint8(r), uint8(g), uint8(b), c.A}
}

// ApplyHighContrast modifies a color based on high contrast scheme.
func ApplyHighContrast(c color.RGBA, scheme HighContrastScheme) color.RGBA {
	switch scheme {
	case HighContrastBlackWhite:
		// Convert to grayscale and threshold
		gray := (int(c.R) + int(c.G) + int(c.B)) / 3
		if gray > 127 {
			return color.RGBA{255, 255, 255, c.A}
		}
		return color.RGBA{0, 0, 0, c.A}

	case HighContrastYellowBlack:
		gray := (int(c.R) + int(c.G) + int(c.B)) / 3
		if gray > 100 {
			return color.RGBA{255, 255, 0, c.A} // Yellow
		}
		return color.RGBA{0, 0, 0, c.A} // Black

	case HighContrastWhiteBlack:
		gray := (int(c.R) + int(c.G) + int(c.B)) / 3
		if gray > 100 {
			return color.RGBA{255, 255, 255, c.A}
		}
		return color.RGBA{0, 0, 0, c.A}

	default:
		return c
	}
}

// GetAccessibleHighlightColor returns an accessible highlight color
// for the given entity type and accessibility settings.
func GetAccessibleHighlightColor(
	entityType string,
	cfg *AccessibilityConfig,
	palette *AccessibleColorPalette,
) color.RGBA {
	if cfg == nil {
		cfg = DefaultAccessibilityConfig()
	}
	if palette == nil {
		palette = DefaultAccessiblePalette()
	}

	var ac *AccessibleColor

	// Map entity type to color
	switch entityType {
	case "weapon":
		ac = &palette.WeaponHighlight
	case "armor":
		ac = &palette.ArmorHighlight
	case "consumable", "potion", "food":
		ac = &palette.ConsumHighlight
	case "quest", "quest_item":
		ac = &palette.QuestHighlight
	case "key":
		ac = &palette.KeyHighlight
	case "treasure", "gold", "loot":
		ac = &palette.TreasureHighlight
	case "locked", "door_locked":
		ac = &palette.LockedHighlight
	case "danger", "trap", "hazard":
		ac = &palette.DangerHighlight
	case "friendly", "ally", "merchant":
		ac = &palette.FriendlyHighlight
	case "hostile", "enemy":
		ac = &palette.HostileHighlight
	case "neutral", "npc":
		ac = &palette.NeutralHighlight
	default:
		ac = &palette.ItemHighlight
	}

	// Get base color for colorblind mode
	c := ac.GetColor(cfg.ColorblindMode)

	// Apply brightness adjustment
	if cfg.HighlightBrightness != 1.0 {
		c = ApplyBrightness(c, cfg.HighlightBrightness)
	}

	// Apply high contrast if enabled
	if cfg.HighContrastScheme != HighContrastNone {
		c = ApplyHighContrast(c, cfg.HighContrastScheme)
	}

	return c
}

// ShouldShowOutline returns whether to show outline for entity.
func (cfg *AccessibilityConfig) ShouldShowOutline() bool {
	return cfg.OutlineThickness > 0 || cfg.LargerInteractionHighlights
}

// GetEffectiveOutlineThickness returns outline thickness to use.
func (cfg *AccessibilityConfig) GetEffectiveOutlineThickness() int {
	t := cfg.OutlineThickness
	if cfg.LargerInteractionHighlights && t < 2 {
		t = 2
	}
	return t
}

// GetEffectiveIndicatorSize returns interaction indicator size.
func (cfg *AccessibilityConfig) GetEffectiveIndicatorSize() int {
	size := cfg.InteractionIndicatorSize
	if cfg.LargerInteractionHighlights {
		size = int(float64(size) * 1.5)
	}
	return size
}

// ShouldUsePulse returns whether pulsing effects are allowed.
func (cfg *AccessibilityConfig) ShouldUsePulse() bool {
	return !cfg.FlashingDisabled
}

// ColorblindModeName returns the display name for a colorblind mode.
func ColorblindModeName(mode ColorblindMode) string {
	switch mode {
	case ColorblindProtanopia:
		return "Protanopia (Red-blind)"
	case ColorblindDeuteranopia:
		return "Deuteranopia (Green-blind)"
	case ColorblindTritanopia:
		return "Tritanopia (Blue-blind)"
	case ColorblindAchromatopsia:
		return "Achromatopsia (No color)"
	default:
		return "Normal"
	}
}

// HighContrastSchemeName returns the display name for a contrast scheme.
func HighContrastSchemeName(scheme HighContrastScheme) string {
	switch scheme {
	case HighContrastBlackWhite:
		return "Black & White"
	case HighContrastYellowBlack:
		return "Yellow on Black"
	case HighContrastWhiteBlack:
		return "White on Black"
	default:
		return "Normal"
	}
}

// GetAccessibilityConfig returns the renderer's accessibility configuration.
func (r *Renderer) GetAccessibilityConfig() *AccessibilityConfig {
	if r.accessibilityConfig == nil {
		r.accessibilityConfig = DefaultAccessibilityConfig()
	}
	return r.accessibilityConfig
}

// SetAccessibilityConfig sets the renderer's accessibility configuration.
func (r *Renderer) SetAccessibilityConfig(cfg *AccessibilityConfig) {
	r.accessibilityConfig = cfg
}

// SetColorblindMode sets the colorblind mode for rendering.
func (r *Renderer) SetColorblindMode(mode ColorblindMode) {
	cfg := r.GetAccessibilityConfig()
	cfg.ColorblindMode = mode
}

// SetHighContrastScheme sets the high contrast scheme for rendering.
func (r *Renderer) SetHighContrastScheme(scheme HighContrastScheme) {
	cfg := r.GetAccessibilityConfig()
	cfg.HighContrastScheme = scheme
}
