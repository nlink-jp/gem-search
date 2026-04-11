// Package config provides configuration for gem-search.
package config

import (
	"fmt"
	"os"
)

const (
	DefaultLocation = "us-central1"
	DefaultModel    = "gemini-2.5-flash"
)

// Config holds runtime configuration.
type Config struct {
	Project   string // GCP project ID
	Location  string // Vertex AI region
	Model     string // Gemini model name
	Format    string // Output format: json, markdown, both
	Output    string // Output file prefix
	MaxRounds int    // Maximum autonomous search rounds
	Lang      string // Output language code
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	c := &Config{
		Project:  os.Getenv("GEMSEARCH_PROJECT"),
		Location: envOrDefault("GEMSEARCH_LOCATION", DefaultLocation),
		Model:    envOrDefault("GEMSEARCH_MODEL", DefaultModel),
		Lang:     os.Getenv("GEMSEARCH_LANG"),
	}

	if c.Project == "" {
		return nil, fmt.Errorf("GEMSEARCH_PROJECT is required")
	}

	return c, nil
}

// ApplyFlags merges CLI flag values into the config.
func (c *Config) ApplyFlags(format, output, lang string, maxRounds int) {
	c.Format = format
	c.Output = output
	c.MaxRounds = maxRounds
	// CLI flag overrides env var
	if lang != "" {
		c.Lang = lang
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
