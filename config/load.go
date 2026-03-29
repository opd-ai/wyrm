// Package config handles configuration loading via Viper.
package config

import (
	"strings"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	Window     WindowConfig     `mapstructure:"window"`
	Server     ServerConfig     `mapstructure:"server"`
	World      WorldConfig      `mapstructure:"world"`
	Federation FederationConfig `mapstructure:"federation"`
	Genre      string           `mapstructure:"genre"`
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

	viper.SetDefault("genre", "fantasy")
}
