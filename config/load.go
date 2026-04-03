// Package config handles configuration loading via Viper.
package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	Window        WindowConfig        `mapstructure:"window"`
	Server        ServerConfig        `mapstructure:"server"`
	World         WorldConfig         `mapstructure:"world"`
	Audio         AudioConfig         `mapstructure:"audio"`
	Federation    FederationConfig    `mapstructure:"federation"`
	Accessibility AccessibilityConfig `mapstructure:"accessibility"`
	Difficulty    DifficultyConfig    `mapstructure:"difficulty"`
	KeyBindings   KeyBindingsConfig   `mapstructure:"keybindings"`
	Mouse         MouseConfig         `mapstructure:"mouse"`
	Debug         DebugConfig         `mapstructure:"debug"`
	RenderQuality RenderQualityConfig `mapstructure:"render_quality"`
	Genre         string              `mapstructure:"genre"`
}

// Validate checks that all configuration values are within valid ranges.
// Returns an error describing the first invalid value found.
func (c *Config) Validate() error {
	if err := c.validateWindow(); err != nil {
		return err
	}
	if err := c.validateServer(); err != nil {
		return err
	}
	if err := c.validateWorld(); err != nil {
		return err
	}
	if err := c.validateAudio(); err != nil {
		return err
	}
	if err := c.validateDifficulty(); err != nil {
		return err
	}
	if err := c.validateDebug(); err != nil {
		return err
	}
	if err := c.validateRenderQuality(); err != nil {
		return err
	}
	return c.validateMouse()
}

// validateWindow checks window configuration values.
func (c *Config) validateWindow() error {
	if c.Window.Width <= 0 {
		return fmt.Errorf("config: window.width must be positive, got %d", c.Window.Width)
	}
	if c.Window.Height <= 0 {
		return fmt.Errorf("config: window.height must be positive, got %d", c.Window.Height)
	}
	return nil
}

// validateServer checks server configuration values.
func (c *Config) validateServer() error {
	if c.Server.TickRate <= 0 {
		return fmt.Errorf("config: server.tick_rate must be positive, got %d", c.Server.TickRate)
	}
	return nil
}

// validateWorld checks world configuration values.
func (c *Config) validateWorld() error {
	if c.World.ChunkSize <= 0 {
		return fmt.Errorf("config: world.chunk_size must be positive, got %d", c.World.ChunkSize)
	}
	return nil
}

// validateAudio checks audio configuration values.
func (c *Config) validateAudio() error {
	if c.Audio.MasterVolume < 0 || c.Audio.MasterVolume > 10 {
		return fmt.Errorf("config: audio.master_volume must be 0-10, got %d", c.Audio.MasterVolume)
	}
	return nil
}

// validateDifficulty checks difficulty multiplier values.
func (c *Config) validateDifficulty() error {
	if c.Difficulty.EnemyDamageMultiplier < 0 {
		return fmt.Errorf("config: difficulty.enemy_damage_multiplier cannot be negative, got %f", c.Difficulty.EnemyDamageMultiplier)
	}
	if c.Difficulty.EnemyHealthMultiplier < 0 {
		return fmt.Errorf("config: difficulty.enemy_health_multiplier cannot be negative, got %f", c.Difficulty.EnemyHealthMultiplier)
	}
	if c.Difficulty.PlayerDamageMultiplier < 0 {
		return fmt.Errorf("config: difficulty.player_damage_multiplier cannot be negative, got %f", c.Difficulty.PlayerDamageMultiplier)
	}
	return nil
}

// validateDebug checks debug configuration values.
func (c *Config) validateDebug() error {
	if c.Debug.ProfilingEnabled && (c.Debug.ProfilingPort < 1 || c.Debug.ProfilingPort > 65535) {
		return fmt.Errorf("config: debug.profiling_port must be 1-65535, got %d", c.Debug.ProfilingPort)
	}
	return nil
}

// validateRenderQuality checks render quality configuration values.
func (c *Config) validateRenderQuality() error {
	level := c.RenderQuality.Level
	if level != QualityHigh && level != QualityMedium && level != QualityLow && level != QualityAuto {
		return fmt.Errorf("config: render_quality.level must be high, medium, low, or auto, got %s", level)
	}
	if c.RenderQuality.TargetFrameTime <= 0 {
		return fmt.Errorf("config: render_quality.target_frame_time must be positive, got %f", c.RenderQuality.TargetFrameTime)
	}
	if c.RenderQuality.DegradationThreshold <= 0 {
		return fmt.Errorf("config: render_quality.degradation_threshold must be positive, got %f", c.RenderQuality.DegradationThreshold)
	}
	if c.RenderQuality.RecoveryThreshold < 0 {
		return fmt.Errorf("config: render_quality.recovery_threshold cannot be negative, got %f", c.RenderQuality.RecoveryThreshold)
	}
	if c.RenderQuality.ParticleCount < 0 {
		return fmt.Errorf("config: render_quality.particle_count cannot be negative, got %d", c.RenderQuality.ParticleCount)
	}
	if c.RenderQuality.BarrierDetailLevel < 0 || c.RenderQuality.BarrierDetailLevel > 2 {
		return fmt.Errorf("config: render_quality.barrier_detail_level must be 0-2, got %d", c.RenderQuality.BarrierDetailLevel)
	}
	if c.RenderQuality.DrawDistance <= 0 {
		return fmt.Errorf("config: render_quality.draw_distance must be positive, got %f", c.RenderQuality.DrawDistance)
	}
	if c.RenderQuality.TextureQuality <= 0 {
		return fmt.Errorf("config: render_quality.texture_quality must be positive, got %f", c.RenderQuality.TextureQuality)
	}
	return nil
}

// validateMouse checks mouse configuration values.
func (c *Config) validateMouse() error {
	if c.Mouse.Sensitivity < 0 {
		return fmt.Errorf("config: mouse.sensitivity cannot be negative, got %f", c.Mouse.Sensitivity)
	}
	if c.Mouse.SmoothingFactor < 0 || c.Mouse.SmoothingFactor > 1 {
		return fmt.Errorf("config: mouse.smoothing_factor must be 0-1, got %f", c.Mouse.SmoothingFactor)
	}
	return nil
}

// MouseConfig holds mouse input settings for FPS-style camera control.
type MouseConfig struct {
	Sensitivity     float64 `mapstructure:"sensitivity"`      // Base mouse sensitivity (default 0.5)
	AccelerationOn  bool    `mapstructure:"acceleration_on"`  // Enable mouse acceleration
	Acceleration    float64 `mapstructure:"acceleration"`     // Acceleration multiplier (default 1.0)
	InvertY         bool    `mapstructure:"invert_y"`         // Invert vertical mouse axis
	SmoothingOn     bool    `mapstructure:"smoothing_on"`     // Enable input smoothing
	SmoothingFactor float64 `mapstructure:"smoothing_factor"` // Smoothing interpolation factor (0.1-1.0)
	RawInput        bool    `mapstructure:"raw_input"`        // Use raw mouse input (bypasses OS acceleration)
}

// DebugConfig holds debugging and profiling settings.
type DebugConfig struct {
	ProfilingEnabled bool `mapstructure:"profiling_enabled"`
	ProfilingPort    int  `mapstructure:"profiling_port"`
	ShowFrameTime    bool `mapstructure:"show_frame_time"`
	ShowMemStats     bool `mapstructure:"show_mem_stats"`
	ShowEntityCount  bool `mapstructure:"show_entity_count"`
}

// QualityLevel represents predefined render quality levels.
type QualityLevel string

const (
	QualityHigh   QualityLevel = "high"
	QualityMedium QualityLevel = "medium"
	QualityLow    QualityLevel = "low"
	QualityAuto   QualityLevel = "auto"
)

// RenderQualityConfig holds rendering quality settings.
type RenderQualityConfig struct {
	// Level is the current quality tier (high, medium, low, auto).
	Level QualityLevel `mapstructure:"level"`
	// AutoDetect enables automatic quality detection at startup.
	AutoDetect bool `mapstructure:"auto_detect"`
	// AdaptiveQuality enables runtime quality adjustments based on frame time.
	AdaptiveQuality bool `mapstructure:"adaptive_quality"`
	// TargetFrameTime is the target frame time in milliseconds (e.g., 16.67 for 60 FPS).
	TargetFrameTime float64 `mapstructure:"target_frame_time"`
	// DegradationThreshold is the frame time (ms) above which quality is reduced.
	DegradationThreshold float64 `mapstructure:"degradation_threshold"`
	// RecoveryThreshold is the frame time (ms) below which quality can be restored.
	RecoveryThreshold float64 `mapstructure:"recovery_threshold"`
	// ParticleCount is the maximum number of particles (scaled by quality tier).
	ParticleCount int `mapstructure:"particle_count"`
	// NormalMapsEnabled enables normal map rendering.
	NormalMapsEnabled bool `mapstructure:"normal_maps_enabled"`
	// BarrierDetailLevel is the detail level for barrier sprites (0-2).
	BarrierDetailLevel int `mapstructure:"barrier_detail_level"`
	// ShadowsEnabled enables shadow rendering.
	ShadowsEnabled bool `mapstructure:"shadows_enabled"`
	// ReflectionsEnabled enables reflection rendering.
	ReflectionsEnabled bool `mapstructure:"reflections_enabled"`
	// DrawDistance is the maximum render distance in world units.
	DrawDistance float64 `mapstructure:"draw_distance"`
	// TextureQuality is the texture resolution multiplier (0.5, 1.0, 2.0).
	TextureQuality float64 `mapstructure:"texture_quality"`
}

// GetEffectiveQuality returns the quality settings for a given level.
func (c *RenderQualityConfig) GetEffectiveQuality(level QualityLevel) RenderQualityConfig {
	switch level {
	case QualityHigh:
		return RenderQualityConfig{
			Level:              QualityHigh,
			ParticleCount:      1000,
			NormalMapsEnabled:  true,
			BarrierDetailLevel: 2,
			ShadowsEnabled:     true,
			ReflectionsEnabled: true,
			DrawDistance:       100.0,
			TextureQuality:     1.0,
		}
	case QualityMedium:
		return RenderQualityConfig{
			Level:              QualityMedium,
			ParticleCount:      500,
			NormalMapsEnabled:  true,
			BarrierDetailLevel: 1,
			ShadowsEnabled:     true,
			ReflectionsEnabled: false,
			DrawDistance:       75.0,
			TextureQuality:     1.0,
		}
	case QualityLow:
		return RenderQualityConfig{
			Level:              QualityLow,
			ParticleCount:      200,
			NormalMapsEnabled:  false,
			BarrierDetailLevel: 0,
			ShadowsEnabled:     false,
			ReflectionsEnabled: false,
			DrawDistance:       50.0,
			TextureQuality:     0.5,
		}
	default:
		return *c
	}
}

// ShouldDegrade returns true if quality should be reduced based on frame time.
func (c *RenderQualityConfig) ShouldDegrade(frameTimeMs float64) bool {
	return c.AdaptiveQuality && frameTimeMs > c.DegradationThreshold
}

// ShouldRecover returns true if quality can be restored based on frame time.
func (c *RenderQualityConfig) ShouldRecover(frameTimeMs float64) bool {
	return c.AdaptiveQuality && frameTimeMs < c.RecoveryThreshold
}

// WindowConfig holds display settings.
type WindowConfig struct {
	Width      int    `mapstructure:"width"`
	Height     int    `mapstructure:"height"`
	Title      string `mapstructure:"title"`
	Fullscreen bool   `mapstructure:"fullscreen"`
	ShowFPS    bool   `mapstructure:"show_fps"`
}

// AudioConfig holds audio settings.
type AudioConfig struct {
	MasterVolume int  `mapstructure:"master_volume"` // 0-10
	MusicEnabled bool `mapstructure:"music_enabled"`
	SFXEnabled   bool `mapstructure:"sfx_enabled"`
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
	// Death penalty settings
	DeathXPLossPercent     float64 `mapstructure:"death_xp_loss_percent"`    // 0.0-1.0, portion of XP lost on death
	DeathGoldLossPercent   float64 `mapstructure:"death_gold_loss_percent"`  // 0.0-1.0, portion of gold lost on death
	DeathDropItems         bool    `mapstructure:"death_drop_items"`         // Whether items drop on death
	DeathRespawnAtGrave    bool    `mapstructure:"death_respawn_at_grave"`   // Respawn at death location or checkpoint
	DeathDurabilityLoss    float64 `mapstructure:"death_durability_loss"`    // 0.0-1.0, equipment durability lost
	DeathCorpseRetrievable bool    `mapstructure:"death_corpse_retrievable"` // Can retrieve items from corpse
}

// KeyBindingsConfig holds configurable key bindings.
// Keys are specified as Ebitengine key names (e.g., "W", "Space", "Escape").
type KeyBindingsConfig struct {
	// Movement
	MoveForward  string `mapstructure:"move_forward"`
	MoveBackward string `mapstructure:"move_backward"`
	MoveLeft     string `mapstructure:"move_left"`
	MoveRight    string `mapstructure:"move_right"`
	StrafeLeft   string `mapstructure:"strafe_left"`
	StrafeRight  string `mapstructure:"strafe_right"`
	Jump         string `mapstructure:"jump"`
	Crouch       string `mapstructure:"crouch"`
	Sprint       string `mapstructure:"sprint"`

	// Combat
	Attack       string `mapstructure:"attack"`
	Block        string `mapstructure:"block"`
	UseAbility1  string `mapstructure:"ability_1"`
	UseAbility2  string `mapstructure:"ability_2"`
	UseAbility3  string `mapstructure:"ability_3"`
	UseAbility4  string `mapstructure:"ability_4"`
	QuickHeal    string `mapstructure:"quick_heal"`
	ToggleWeapon string `mapstructure:"toggle_weapon"`

	// Interaction
	Interact     string `mapstructure:"interact"`
	PickUp       string `mapstructure:"pick_up"`
	DropItem     string `mapstructure:"drop_item"`
	UseItem      string `mapstructure:"use_item"`
	Talk         string `mapstructure:"talk"`
	ReadSign     string `mapstructure:"read_sign"`
	Mount        string `mapstructure:"mount"`
	EnterVehicle string `mapstructure:"enter_vehicle"`

	// UI
	Inventory   string `mapstructure:"inventory"`
	Map         string `mapstructure:"map"`
	QuestLog    string `mapstructure:"quest_log"`
	CharSheet   string `mapstructure:"character_sheet"`
	SkillTree   string `mapstructure:"skill_tree"`
	Crafting    string `mapstructure:"crafting"`
	Pause       string `mapstructure:"pause"`
	QuickSave   string `mapstructure:"quick_save"`
	QuickLoad   string `mapstructure:"quick_load"`
	Screenshot  string `mapstructure:"screenshot"`
	ToggleHUD   string `mapstructure:"toggle_hud"`
	Console     string `mapstructure:"console"`
	ChatWindow  string `mapstructure:"chat_window"`
	SocialMenu  string `mapstructure:"social_menu"`
	TradeWindow string `mapstructure:"trade_window"`
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

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func setDefaults() {
	viper.SetDefault("window.width", 1280)
	viper.SetDefault("window.height", 720)
	viper.SetDefault("window.title", "Wyrm")
	viper.SetDefault("window.fullscreen", false)
	viper.SetDefault("window.show_fps", true)

	viper.SetDefault("audio.master_volume", 7)
	viper.SetDefault("audio.music_enabled", true)
	viper.SetDefault("audio.sfx_enabled", true)

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
	viper.SetDefault("difficulty.death_xp_loss_percent", 0.1)   // 10% XP loss
	viper.SetDefault("difficulty.death_gold_loss_percent", 0.1) // 10% gold loss
	viper.SetDefault("difficulty.death_drop_items", false)
	viper.SetDefault("difficulty.death_respawn_at_grave", false)
	viper.SetDefault("difficulty.death_durability_loss", 0.1) // 10% durability
	viper.SetDefault("difficulty.death_corpse_retrievable", true)

	// Default key bindings
	viper.SetDefault("keybindings.move_forward", "W")
	viper.SetDefault("keybindings.move_backward", "S")
	viper.SetDefault("keybindings.move_left", "A")
	viper.SetDefault("keybindings.move_right", "D")
	viper.SetDefault("keybindings.jump", "Space")
	viper.SetDefault("keybindings.crouch", "ControlLeft")
	viper.SetDefault("keybindings.sprint", "ShiftLeft")

	viper.SetDefault("keybindings.attack", "MouseButtonLeft")
	viper.SetDefault("keybindings.block", "MouseButtonRight")
	viper.SetDefault("keybindings.ability_1", "1")
	viper.SetDefault("keybindings.ability_2", "2")
	viper.SetDefault("keybindings.ability_3", "3")
	viper.SetDefault("keybindings.ability_4", "4")
	viper.SetDefault("keybindings.quick_heal", "H")
	viper.SetDefault("keybindings.toggle_weapon", "Tab")

	viper.SetDefault("keybindings.interact", "E")
	viper.SetDefault("keybindings.pick_up", "F")
	viper.SetDefault("keybindings.drop_item", "G")
	viper.SetDefault("keybindings.use_item", "R")
	viper.SetDefault("keybindings.talk", "T")
	viper.SetDefault("keybindings.read_sign", "V")
	viper.SetDefault("keybindings.mount", "X")
	viper.SetDefault("keybindings.enter_vehicle", "C")

	viper.SetDefault("keybindings.inventory", "I")
	viper.SetDefault("keybindings.map", "M")
	viper.SetDefault("keybindings.quest_log", "J")
	viper.SetDefault("keybindings.character_sheet", "K")
	viper.SetDefault("keybindings.skill_tree", "P")
	viper.SetDefault("keybindings.crafting", "B")
	viper.SetDefault("keybindings.pause", "Escape")
	viper.SetDefault("keybindings.quick_save", "F5")
	viper.SetDefault("keybindings.quick_load", "F9")
	viper.SetDefault("keybindings.screenshot", "F12")
	viper.SetDefault("keybindings.toggle_hud", "F1")
	viper.SetDefault("keybindings.console", "Backquote")
	viper.SetDefault("keybindings.chat_window", "Enter")
	viper.SetDefault("keybindings.social_menu", "O")
	viper.SetDefault("keybindings.trade_window", "Y")

	// Mouse settings defaults
	viper.SetDefault("mouse.sensitivity", 0.5)
	viper.SetDefault("mouse.acceleration_on", false)
	viper.SetDefault("mouse.acceleration", 1.0)
	viper.SetDefault("mouse.invert_y", false)
	viper.SetDefault("mouse.smoothing_on", false)
	viper.SetDefault("mouse.smoothing_factor", 0.5)
	viper.SetDefault("mouse.raw_input", true)

	// Debug/profiling defaults
	viper.SetDefault("debug.profiling_enabled", false)
	viper.SetDefault("debug.profiling_port", 6060)
	viper.SetDefault("debug.show_frame_time", false)
	viper.SetDefault("debug.show_mem_stats", false)
	viper.SetDefault("debug.show_entity_count", false)

	// Render quality defaults
	viper.SetDefault("render_quality.level", "auto")
	viper.SetDefault("render_quality.auto_detect", true)
	viper.SetDefault("render_quality.adaptive_quality", true)
	viper.SetDefault("render_quality.target_frame_time", 16.67) // 60 FPS
	viper.SetDefault("render_quality.degradation_threshold", 25.0)
	viper.SetDefault("render_quality.recovery_threshold", 12.0)
	viper.SetDefault("render_quality.particle_count", 500)
	viper.SetDefault("render_quality.normal_maps_enabled", true)
	viper.SetDefault("render_quality.barrier_detail_level", 1)
	viper.SetDefault("render_quality.shadows_enabled", true)
	viper.SetDefault("render_quality.reflections_enabled", false)
	viper.SetDefault("render_quality.draw_distance", 75.0)
	viper.SetDefault("render_quality.texture_quality", 1.0)

	viper.SetDefault("genre", "fantasy")
}

// Save writes the current configuration to a file.
func (c *Config) Save(path string) error {
	// Update viper with current config values
	viper.Set("window.width", c.Window.Width)
	viper.Set("window.height", c.Window.Height)
	viper.Set("window.title", c.Window.Title)
	viper.Set("window.fullscreen", c.Window.Fullscreen)
	viper.Set("window.show_fps", c.Window.ShowFPS)

	viper.Set("audio.master_volume", c.Audio.MasterVolume)
	viper.Set("audio.music_enabled", c.Audio.MusicEnabled)
	viper.Set("audio.sfx_enabled", c.Audio.SFXEnabled)

	viper.Set("genre", c.Genre)

	// Write to file
	if path == "" {
		return viper.WriteConfig()
	}
	return viper.WriteConfigAs(path)
}
