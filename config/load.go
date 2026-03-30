// Package config handles configuration loading via Viper.
package config

import (
	"strings"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	Window        WindowConfig        `mapstructure:"window"`
	Server        ServerConfig        `mapstructure:"server"`
	World         WorldConfig         `mapstructure:"world"`
	Federation    FederationConfig    `mapstructure:"federation"`
	Accessibility AccessibilityConfig `mapstructure:"accessibility"`
	Difficulty    DifficultyConfig    `mapstructure:"difficulty"`
	Genre         string              `mapstructure:"genre"`
}

// WindowConfig holds display settings.
type WindowConfig struct {
	Width  int    `mapstructure:"width"`
	Height int    `mapstructure:"height"`
	Title  string `mapstructure:"title"`
}

// ServerConfig holds server settings.
type ServerConfig struct {
	Address  string `mapstructure:"address"`
	Protocol string `mapstructure:"protocol"`
	TickRate int    `mapstructure:"tick_rate"`
}

// WorldConfig holds world generation settings.
type WorldConfig struct {
	Seed      int64 `mapstructure:"seed"`
	ChunkSize int   `mapstructure:"chunk_size"`
}

// FederationConfig holds cross-server federation settings.
type FederationConfig struct {
	Enabled        bool     `mapstructure:"enabled"`
	NodeID         string   `mapstructure:"node_id"`
	Peers          []string `mapstructure:"peers"`
	GossipInterval int      `mapstructure:"gossip_interval"` // seconds
}

// ColorblindMode represents a colorblind accessibility mode.
type ColorblindMode string

const (
	ColorblindNone          ColorblindMode = "none"
	ColorblindProtanopia    ColorblindMode = "protanopia"    // Red-blind
	ColorblindDeuteranopia  ColorblindMode = "deuteranopia"  // Green-blind
	ColorblindTritanopia    ColorblindMode = "tritanopia"    // Blue-blind
	ColorblindAchromatopsia ColorblindMode = "achromatopsia" // Total color blindness
)

// AccessibilityConfig holds accessibility settings.
type AccessibilityConfig struct {
	ColorblindMode   ColorblindMode `mapstructure:"colorblind_mode"`
	HighContrast     bool           `mapstructure:"high_contrast"`
	LargeText        bool           `mapstructure:"large_text"`
	ReducedMotion    bool           `mapstructure:"reduced_motion"`
	ScreenReaderMode bool           `mapstructure:"screen_reader_mode"`
	SubtitlesEnabled bool           `mapstructure:"subtitles_enabled"`
	SubtitleSize     float64        `mapstructure:"subtitle_size"` // Multiplier (1.0 = normal)
}

// DifficultyLevel represents predefined difficulty levels.
type DifficultyLevel string

const (
	DifficultyEasy   DifficultyLevel = "easy"
	DifficultyNormal DifficultyLevel = "normal"
	DifficultyHard   DifficultyLevel = "hard"
	DifficultyCustom DifficultyLevel = "custom"
)

// DifficultyConfig holds difficulty settings.
type DifficultyConfig struct {
	Level                  DifficultyLevel `mapstructure:"level"`
	EnemyDamageMultiplier  float64         `mapstructure:"enemy_damage_multiplier"`
	EnemyHealthMultiplier  float64         `mapstructure:"enemy_health_multiplier"`
	PlayerDamageMultiplier float64         `mapstructure:"player_damage_multiplier"`
	ResourceScarcity       float64         `mapstructure:"resource_scarcity"` // 1.0 = normal, 2.0 = scarce
	PermaDeath             bool            `mapstructure:"perma_death"`
	FriendlyFire           bool            `mapstructure:"friendly_fire"`
	AutoAim                bool            `mapstructure:"auto_aim"`
}

// ColorblindPalette returns adjusted color values for a colorblind mode.
func (m ColorblindMode) ColorblindPalette() map[string][3]uint8 {
	switch m {
	case ColorblindProtanopia:
		return map[string][3]uint8{
			"red":     {122, 122, 0}, // Red appears as yellow-brown
			"green":   {163, 163, 0}, // Green appears as yellow
			"blue":    {0, 0, 255},   // Blue unchanged
			"warning": {255, 200, 0}, // Use yellow for warnings
			"danger":  {0, 0, 200},   // Use blue for danger
		}
	case ColorblindDeuteranopia:
		return map[string][3]uint8{
			"red":     {166, 166, 0}, // Red appears as yellow-brown
			"green":   {194, 194, 0}, // Green appears as yellow
			"blue":    {0, 0, 255},   // Blue unchanged
			"warning": {255, 200, 0}, // Use yellow for warnings
			"danger":  {0, 0, 200},   // Use blue for danger
		}
	case ColorblindTritanopia:
		return map[string][3]uint8{
			"red":     {255, 0, 0},   // Red unchanged
			"green":   {0, 200, 200}, // Green appears as cyan
			"blue":    {0, 150, 150}, // Blue appears as cyan
			"warning": {255, 150, 0}, // Use orange for warnings
			"danger":  {255, 0, 100}, // Use pink for danger
		}
	case ColorblindAchromatopsia:
		return map[string][3]uint8{
			"red":     {128, 128, 128}, // All as grayscale
			"green":   {170, 170, 170},
			"blue":    {85, 85, 85},
			"warning": {200, 200, 200},
			"danger":  {50, 50, 50},
		}
	default:
		return map[string][3]uint8{
			"red":     {255, 0, 0},
			"green":   {0, 255, 0},
			"blue":    {0, 0, 255},
			"warning": {255, 200, 0},
			"danger":  {255, 0, 0},
		}
	}
}

// GetDifficultyMultipliers returns the combat multipliers for a difficulty level.
func (d DifficultyLevel) GetDifficultyMultipliers() (enemyDmg, enemyHP, playerDmg float64) {
	switch d {
	case DifficultyEasy:
		return 0.5, 0.75, 1.5
	case DifficultyNormal:
		return 1.0, 1.0, 1.0
	case DifficultyHard:
		return 1.5, 1.5, 0.75
	default:
		return 1.0, 1.0, 1.0
	}
}

// Load reads configuration from file and environment, returning the populated Config.
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	setDefaults()

	viper.SetEnvPrefix("WYRM")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
		// Config file not found; use defaults
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func setDefaults() {
	viper.SetDefault("window.width", 1280)
	viper.SetDefault("window.height", 720)
	viper.SetDefault("window.title", "Wyrm")

	viper.SetDefault("server.address", "localhost:7777")
	viper.SetDefault("server.protocol", "tcp")
	viper.SetDefault("server.tick_rate", 20)

	viper.SetDefault("world.seed", 0)
	viper.SetDefault("world.chunk_size", 512)

	viper.SetDefault("federation.enabled", false)
	viper.SetDefault("federation.node_id", "")
	viper.SetDefault("federation.peers", []string{})
	viper.SetDefault("federation.gossip_interval", 5)

	viper.SetDefault("accessibility.colorblind_mode", "none")
	viper.SetDefault("accessibility.high_contrast", false)
	viper.SetDefault("accessibility.large_text", false)
	viper.SetDefault("accessibility.reduced_motion", false)
	viper.SetDefault("accessibility.screen_reader_mode", false)
	viper.SetDefault("accessibility.subtitles_enabled", false)
	viper.SetDefault("accessibility.subtitle_size", 1.0)

	viper.SetDefault("difficulty.level", "normal")
	viper.SetDefault("difficulty.enemy_damage_multiplier", 1.0)
	viper.SetDefault("difficulty.enemy_health_multiplier", 1.0)
	viper.SetDefault("difficulty.player_damage_multiplier", 1.0)
	viper.SetDefault("difficulty.resource_scarcity", 1.0)
	viper.SetDefault("difficulty.perma_death", false)
	viper.SetDefault("difficulty.friendly_fire", false)
	viper.SetDefault("difficulty.auto_aim", false)

	viper.SetDefault("genre", "fantasy")
}
