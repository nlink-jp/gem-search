package config

import (
	"os"
	"path/filepath"
	"testing"
)

func clearEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"GEMSEARCH_PROJECT", "GEMSEARCH_LOCATION", "GEMSEARCH_MODEL", "GEMSEARCH_LANG",
		"GOOGLE_CLOUD_PROJECT", "GOOGLE_CLOUD_LOCATION",
	} {
		os.Unsetenv(key)
	}
	// Isolate XDG/HOME so the user's real ~/.config/gem-search/config.toml
	// doesn't leak into tests. Without this, TestLoadMissingProject fails
	// on a developer machine that has a real config installed.
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Cleanup(func() {
		for _, key := range []string{
			"GEMSEARCH_PROJECT", "GEMSEARCH_LOCATION", "GEMSEARCH_MODEL", "GEMSEARCH_LANG",
			"GOOGLE_CLOUD_PROJECT", "GOOGLE_CLOUD_LOCATION",
		} {
			os.Unsetenv(key)
		}
	})
}

func TestLoadDefaults(t *testing.T) {
	clearEnv(t)
	os.Setenv("GEMSEARCH_PROJECT", "test-project")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.GCP.Location != DefaultLocation {
		t.Errorf("Location = %q, want %q", cfg.GCP.Location, DefaultLocation)
	}
	if cfg.Model.Name != DefaultModel {
		t.Errorf("Model = %q, want %q", cfg.Model.Name, DefaultModel)
	}
}

func TestLoadEnvOverrides(t *testing.T) {
	clearEnv(t)
	os.Setenv("GEMSEARCH_PROJECT", "test-project")
	os.Setenv("GEMSEARCH_LOCATION", "asia-northeast1")
	os.Setenv("GEMSEARCH_MODEL", "gemini-2.5-pro")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.GCP.Project != "test-project" {
		t.Errorf("Project = %q", cfg.GCP.Project)
	}
	if cfg.GCP.Location != "asia-northeast1" {
		t.Errorf("Location = %q", cfg.GCP.Location)
	}
	if cfg.Model.Name != "gemini-2.5-pro" {
		t.Errorf("Model = %q", cfg.Model.Name)
	}
}

func TestLoadGenericEnvFallback(t *testing.T) {
	clearEnv(t)
	os.Setenv("GOOGLE_CLOUD_PROJECT", "generic-project")
	os.Setenv("GOOGLE_CLOUD_LOCATION", "europe-west1")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.GCP.Project != "generic-project" {
		t.Errorf("Project = %q, want generic-project", cfg.GCP.Project)
	}
	if cfg.GCP.Location != "europe-west1" {
		t.Errorf("Location = %q, want europe-west1", cfg.GCP.Location)
	}
}

func TestLoadToolSpecificEnvOverridesGeneric(t *testing.T) {
	clearEnv(t)
	os.Setenv("GOOGLE_CLOUD_PROJECT", "generic-project")
	os.Setenv("GEMSEARCH_PROJECT", "specific-project")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.GCP.Project != "specific-project" {
		t.Errorf("Project = %q, want specific-project", cfg.GCP.Project)
	}
}

func TestLoadMissingProject(t *testing.T) {
	clearEnv(t)
	_, err := Load("")
	if err == nil {
		t.Fatal("expected error for missing project")
	}
}

func TestLoadTOMLFile(t *testing.T) {
	clearEnv(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := `[gcp]
project  = "toml-project"
location = "asia-northeast1"

[model]
name = "gemini-2.5-pro"
lang = "ja"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.GCP.Project != "toml-project" {
		t.Errorf("Project = %q", cfg.GCP.Project)
	}
	if cfg.GCP.Location != "asia-northeast1" {
		t.Errorf("Location = %q", cfg.GCP.Location)
	}
	if cfg.Model.Name != "gemini-2.5-pro" {
		t.Errorf("Model = %q", cfg.Model.Name)
	}
	if cfg.Lang != "ja" {
		t.Errorf("Lang = %q, want ja", cfg.Lang)
	}
}

func TestLoadEnvOverridesToml(t *testing.T) {
	clearEnv(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := `[gcp]
project  = "toml-project"
location = "toml-location"

[model]
name = "toml-model"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	os.Setenv("GEMSEARCH_PROJECT", "env-project")
	os.Setenv("GEMSEARCH_MODEL", "env-model")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.GCP.Project != "env-project" {
		t.Errorf("Project = %q, want env-project", cfg.GCP.Project)
	}
	if cfg.Model.Name != "env-model" {
		t.Errorf("Model = %q, want env-model", cfg.Model.Name)
	}
	// Location not overridden by env → keeps TOML value
	if cfg.GCP.Location != "toml-location" {
		t.Errorf("Location = %q, want toml-location", cfg.GCP.Location)
	}
}

func TestApplyFlagsOverridesLang(t *testing.T) {
	clearEnv(t)
	os.Setenv("GEMSEARCH_PROJECT", "test-project")
	os.Setenv("GEMSEARCH_LANG", "en")

	cfg, _ := Load("")
	cfg.ApplyFlags("json", "./out", "ja")

	if cfg.Lang != "ja" {
		t.Errorf("Lang = %q, want ja (flag should override env)", cfg.Lang)
	}
	if cfg.Format != "json" {
		t.Errorf("Format = %q", cfg.Format)
	}
}

func TestApplyFlagsKeepsEnvLang(t *testing.T) {
	clearEnv(t)
	os.Setenv("GEMSEARCH_PROJECT", "test-project")
	os.Setenv("GEMSEARCH_LANG", "en")

	cfg, _ := Load("")
	cfg.ApplyFlags("markdown", "", "")

	if cfg.Lang != "en" {
		t.Errorf("Lang = %q, want en (env should be kept)", cfg.Lang)
	}
}

func TestApplyFlagsEmpty(t *testing.T) {
	clearEnv(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := `[gcp]
project = "test-project"

[model]
lang = "ja"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, _ := Load(path)
	cfg.ApplyFlags("markdown", "", "")

	if cfg.Lang != "ja" {
		t.Errorf("Lang = %q, want ja (TOML should be kept)", cfg.Lang)
	}
}
