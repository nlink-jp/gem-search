package config

import (
	"os"
	"testing"
)

func setRequiredEnv(t *testing.T) {
	t.Helper()
	os.Setenv("GEMSEARCH_PROJECT", "test-project")
	t.Cleanup(func() {
		os.Unsetenv("GEMSEARCH_PROJECT")
		os.Unsetenv("GEMSEARCH_LOCATION")
		os.Unsetenv("GEMSEARCH_MODEL")
		os.Unsetenv("GEMSEARCH_LANG")
	})
}

func TestLoadSuccess(t *testing.T) {
	setRequiredEnv(t)
	os.Setenv("GEMSEARCH_LOCATION", "asia-northeast1")
	os.Setenv("GEMSEARCH_MODEL", "gemini-2.5-pro")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Project != "test-project" {
		t.Errorf("Project = %q", cfg.Project)
	}
	if cfg.Location != "asia-northeast1" {
		t.Errorf("Location = %q", cfg.Location)
	}
	if cfg.Model != "gemini-2.5-pro" {
		t.Errorf("Model = %q", cfg.Model)
	}
}

func TestLoadDefaults(t *testing.T) {
	setRequiredEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Location != DefaultLocation {
		t.Errorf("Location = %q, want %q", cfg.Location, DefaultLocation)
	}
	if cfg.Model != DefaultModel {
		t.Errorf("Model = %q, want %q", cfg.Model, DefaultModel)
	}
}

func TestLoadMissingProject(t *testing.T) {
	os.Unsetenv("GEMSEARCH_PROJECT")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing PROJECT")
	}
}

func TestApplyFlagsOverridesLang(t *testing.T) {
	setRequiredEnv(t)
	os.Setenv("GEMSEARCH_LANG", "en")

	cfg, _ := Load()
	cfg.ApplyFlags("json", "./out", "ja", 5)

	if cfg.Lang != "ja" {
		t.Errorf("Lang = %q, want %q (flag should override env)", cfg.Lang, "ja")
	}
	if cfg.Format != "json" {
		t.Errorf("Format = %q", cfg.Format)
	}
	if cfg.MaxRounds != 5 {
		t.Errorf("MaxRounds = %d", cfg.MaxRounds)
	}
}

func TestApplyFlagsKeepsEnvLang(t *testing.T) {
	setRequiredEnv(t)
	os.Setenv("GEMSEARCH_LANG", "en")

	cfg, _ := Load()
	cfg.ApplyFlags("markdown", "", "", 3) // empty lang flag

	if cfg.Lang != "en" {
		t.Errorf("Lang = %q, want %q (env should be kept)", cfg.Lang, "en")
	}
}
