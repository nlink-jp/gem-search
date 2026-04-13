// Package config manages gem-search configuration.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const (
	DefaultLocation = "us-central1"
	DefaultModel    = "gemini-2.5-flash"
)

// Config holds all gem-search configuration.
type Config struct {
	GCP   GCPConfig   `toml:"gcp"`
	Model ModelConfig `toml:"model"`

	// Runtime-only fields (not in TOML).
	Format string `toml:"-"`
	Output string `toml:"-"`
	Lang   string `toml:"-"`
}

// GCPConfig holds Google Cloud settings.
type GCPConfig struct {
	Project  string `toml:"project"`
	Location string `toml:"location"`
}

// ModelConfig holds model settings.
type ModelConfig struct {
	Name string `toml:"name"`
	Lang string `toml:"lang"`
}

// Load reads config from the given path, with env var overrides.
// If path is empty, tries the default location (~/.config/gem-search/config.toml).
func Load(path string) (*Config, error) {
	cfg := &Config{
		GCP: GCPConfig{
			Location: DefaultLocation,
		},
		Model: ModelConfig{
			Name: DefaultModel,
		},
	}

	if path == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(home, ".config", "gem-search", "config.toml")
		}
	}

	if path != "" {
		if _, err := os.Stat(path); err == nil {
			if _, err := toml.DecodeFile(path, cfg); err != nil {
				return nil, fmt.Errorf("parse config %s: %w", path, err)
			}
		}
	}

	// TOML model.lang → runtime Lang
	if cfg.Model.Lang != "" {
		cfg.Lang = cfg.Model.Lang
	}

	// Env overrides (tool-specific > generic)
	if v := os.Getenv("GEMSEARCH_PROJECT"); v != "" {
		cfg.GCP.Project = v
	} else if v := os.Getenv("GOOGLE_CLOUD_PROJECT"); v != "" {
		cfg.GCP.Project = v
	}
	if v := os.Getenv("GEMSEARCH_LOCATION"); v != "" {
		cfg.GCP.Location = v
	} else if v := os.Getenv("GOOGLE_CLOUD_LOCATION"); v != "" {
		cfg.GCP.Location = v
	}
	if v := os.Getenv("GEMSEARCH_MODEL"); v != "" {
		cfg.Model.Name = v
	}
	if v := os.Getenv("GEMSEARCH_LANG"); v != "" {
		cfg.Lang = v
	}

	if cfg.GCP.Project == "" {
		return nil, fmt.Errorf("GCP project is required: set gcp.project in config or GOOGLE_CLOUD_PROJECT env var")
	}

	return cfg, nil
}

// ApplyFlags merges CLI flag values into the config.
func (c *Config) ApplyFlags(format, output, lang string) {
	c.Format = format
	c.Output = output
	if lang != "" {
		c.Lang = lang
	}
}
