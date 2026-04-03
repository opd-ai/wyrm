package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Clear any environment variables that could affect the test
	for _, key := range []string{"WYRM_WINDOW_WIDTH", "WYRM_WINDOW_HEIGHT", "WYRM_GENRE"} {
		os.Unsetenv(key)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify defaults are set
	if cfg.Window.Width != 1280 {
		t.Errorf("expected Window.Width=1280, got %d", cfg.Window.Width)
	}
	if cfg.Window.Height != 720 {
		t.Errorf("expected Window.Height=720, got %d", cfg.Window.Height)
	}
	if cfg.Window.Title != "Wyrm" {
		t.Errorf("expected Window.Title='Wyrm', got %q", cfg.Window.Title)
	}
	if cfg.Server.Address != "localhost:7777" {
		t.Errorf("expected Server.Address='localhost:7777', got %q", cfg.Server.Address)
	}
	if cfg.Server.Protocol != "tcp" {
		t.Errorf("expected Server.Protocol='tcp', got %q", cfg.Server.Protocol)
	}
	if cfg.Server.TickRate != 20 {
		t.Errorf("expected Server.TickRate=20, got %d", cfg.Server.TickRate)
	}
	if cfg.World.ChunkSize != 512 {
		t.Errorf("expected World.ChunkSize=512, got %d", cfg.World.ChunkSize)
	}
	if cfg.Genre != "fantasy" {
		t.Errorf("expected Genre='fantasy', got %q", cfg.Genre)
	}
}

func TestLoadWithEnvOverride(t *testing.T) {
	// Set environment variable override
	os.Setenv("WYRM_GENRE", "sci-fi")
	defer os.Unsetenv("WYRM_GENRE")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Genre != "sci-fi" {
		t.Errorf("expected Genre='sci-fi' from env override, got %q", cfg.Genre)
	}
}

func TestLoadWindowEnvOverrides(t *testing.T) {
	os.Setenv("WYRM_WINDOW_WIDTH", "1920")
	os.Setenv("WYRM_WINDOW_HEIGHT", "1080")
	defer func() {
		os.Unsetenv("WYRM_WINDOW_WIDTH")
		os.Unsetenv("WYRM_WINDOW_HEIGHT")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Window.Width != 1920 {
		t.Errorf("expected Window.Width=1920 from env, got %d", cfg.Window.Width)
	}
	if cfg.Window.Height != 1080 {
		t.Errorf("expected Window.Height=1080 from env, got %d", cfg.Window.Height)
	}
}

func TestConfigStruct(t *testing.T) {
	// Verify struct field types
	cfg := &Config{}
	cfg.Window.Width = 800
	cfg.Window.Height = 600
	cfg.Window.Title = "Test"
	cfg.Server.Address = "test:1234"
	cfg.Server.Protocol = "udp"
	cfg.Server.TickRate = 30
	cfg.World.Seed = 12345
	cfg.World.ChunkSize = 256
	cfg.Genre = "horror"

	if cfg.Window.Width != 800 {
		t.Error("Window.Width field assignment failed")
	}
	if cfg.World.Seed != 12345 {
		t.Error("World.Seed field assignment failed")
	}
}

func TestRenderQualityConfigDefaults(t *testing.T) {
	// Clear any environment variables that could affect the test
	for _, key := range []string{"WYRM_RENDER_QUALITY_LEVEL", "WYRM_RENDER_QUALITY_PARTICLE_COUNT"} {
		os.Unsetenv(key)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify render quality defaults
	if cfg.RenderQuality.Level != QualityAuto {
		t.Errorf("expected RenderQuality.Level='auto', got %q", cfg.RenderQuality.Level)
	}
	if !cfg.RenderQuality.AutoDetect {
		t.Error("expected RenderQuality.AutoDetect=true")
	}
	if !cfg.RenderQuality.AdaptiveQuality {
		t.Error("expected RenderQuality.AdaptiveQuality=true")
	}
	if cfg.RenderQuality.TargetFrameTime != 16.67 {
		t.Errorf("expected RenderQuality.TargetFrameTime=16.67, got %f", cfg.RenderQuality.TargetFrameTime)
	}
	if cfg.RenderQuality.ParticleCount != 500 {
		t.Errorf("expected RenderQuality.ParticleCount=500, got %d", cfg.RenderQuality.ParticleCount)
	}
	if !cfg.RenderQuality.NormalMapsEnabled {
		t.Error("expected RenderQuality.NormalMapsEnabled=true")
	}
	if cfg.RenderQuality.BarrierDetailLevel != 1 {
		t.Errorf("expected RenderQuality.BarrierDetailLevel=1, got %d", cfg.RenderQuality.BarrierDetailLevel)
	}
}

func TestRenderQualityGetEffectiveQuality(t *testing.T) {
	cfg := &RenderQualityConfig{}

	// Test high quality settings
	high := cfg.GetEffectiveQuality(QualityHigh)
	if high.ParticleCount != 1000 {
		t.Errorf("expected high quality particle_count=1000, got %d", high.ParticleCount)
	}
	if !high.NormalMapsEnabled {
		t.Error("expected high quality normal_maps_enabled=true")
	}
	if high.BarrierDetailLevel != 2 {
		t.Errorf("expected high quality barrier_detail_level=2, got %d", high.BarrierDetailLevel)
	}

	// Test medium quality settings
	medium := cfg.GetEffectiveQuality(QualityMedium)
	if medium.ParticleCount != 500 {
		t.Errorf("expected medium quality particle_count=500, got %d", medium.ParticleCount)
	}
	if medium.BarrierDetailLevel != 1 {
		t.Errorf("expected medium quality barrier_detail_level=1, got %d", medium.BarrierDetailLevel)
	}

	// Test low quality settings
	low := cfg.GetEffectiveQuality(QualityLow)
	if low.ParticleCount != 200 {
		t.Errorf("expected low quality particle_count=200, got %d", low.ParticleCount)
	}
	if low.NormalMapsEnabled {
		t.Error("expected low quality normal_maps_enabled=false")
	}
	if low.BarrierDetailLevel != 0 {
		t.Errorf("expected low quality barrier_detail_level=0, got %d", low.BarrierDetailLevel)
	}
}

func TestRenderQualityDegradation(t *testing.T) {
	cfg := &RenderQualityConfig{
		AdaptiveQuality:      true,
		DegradationThreshold: 25.0,
		RecoveryThreshold:    12.0,
	}

	// Test degradation detection
	if !cfg.ShouldDegrade(30.0) {
		t.Error("expected ShouldDegrade(30.0)=true")
	}
	if cfg.ShouldDegrade(20.0) {
		t.Error("expected ShouldDegrade(20.0)=false")
	}

	// Test recovery detection
	if !cfg.ShouldRecover(10.0) {
		t.Error("expected ShouldRecover(10.0)=true")
	}
	if cfg.ShouldRecover(15.0) {
		t.Error("expected ShouldRecover(15.0)=false")
	}

	// Test with adaptive quality disabled
	cfg.AdaptiveQuality = false
	if cfg.ShouldDegrade(30.0) {
		t.Error("expected ShouldDegrade=false when AdaptiveQuality disabled")
	}
	if cfg.ShouldRecover(10.0) {
		t.Error("expected ShouldRecover=false when AdaptiveQuality disabled")
	}
}

func TestRenderQualityValidation(t *testing.T) {
	testCases := []struct {
		name        string
		config      RenderQualityConfig
		expectError bool
	}{
		{
			name: "valid_high_quality",
			config: RenderQualityConfig{
				Level:                QualityHigh,
				TargetFrameTime:      16.67,
				DegradationThreshold: 25.0,
				RecoveryThreshold:    12.0,
				ParticleCount:        1000,
				BarrierDetailLevel:   2,
				DrawDistance:         100.0,
				TextureQuality:       1.0,
			},
			expectError: false,
		},
		{
			name: "invalid_level",
			config: RenderQualityConfig{
				Level:                "invalid",
				TargetFrameTime:      16.67,
				DegradationThreshold: 25.0,
				RecoveryThreshold:    12.0,
				ParticleCount:        1000,
				BarrierDetailLevel:   2,
				DrawDistance:         100.0,
				TextureQuality:       1.0,
			},
			expectError: true,
		},
		{
			name: "negative_particle_count",
			config: RenderQualityConfig{
				Level:                QualityMedium,
				TargetFrameTime:      16.67,
				DegradationThreshold: 25.0,
				RecoveryThreshold:    12.0,
				ParticleCount:        -1,
				BarrierDetailLevel:   1,
				DrawDistance:         75.0,
				TextureQuality:       1.0,
			},
			expectError: true,
		},
		{
			name: "invalid_barrier_detail_level",
			config: RenderQualityConfig{
				Level:                QualityMedium,
				TargetFrameTime:      16.67,
				DegradationThreshold: 25.0,
				RecoveryThreshold:    12.0,
				ParticleCount:        500,
				BarrierDetailLevel:   5,
				DrawDistance:         75.0,
				TextureQuality:       1.0,
			},
			expectError: true,
		},
		{
			name: "zero_draw_distance",
			config: RenderQualityConfig{
				Level:                QualityLow,
				TargetFrameTime:      16.67,
				DegradationThreshold: 25.0,
				RecoveryThreshold:    12.0,
				ParticleCount:        200,
				BarrierDetailLevel:   0,
				DrawDistance:         0,
				TextureQuality:       0.5,
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &Config{
				Window:        WindowConfig{Width: 1280, Height: 720, Title: "Test"},
				Server:        ServerConfig{Address: "localhost:7777", Protocol: "tcp", TickRate: 20},
				World:         WorldConfig{ChunkSize: 512},
				Audio:         AudioConfig{MasterVolume: 1.0},
				Difficulty:    DifficultyConfig{Level: DifficultyNormal},
				Debug:         DebugConfig{ProfilingPort: 6060},
				Mouse:         MouseConfig{SmoothingFactor: 0.5},
				RenderQuality: tc.config,
				Genre:         "fantasy",
			}

			err := cfg.Validate()
			if tc.expectError && err == nil {
				t.Errorf("expected validation error for %s, but got none", tc.name)
			}
			if !tc.expectError && err != nil {
				t.Errorf("unexpected validation error for %s: %v", tc.name, err)
			}
		})
	}
}
