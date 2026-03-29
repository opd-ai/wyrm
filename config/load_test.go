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
