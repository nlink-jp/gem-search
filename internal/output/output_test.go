package output

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nlink-jp/gem-search/internal/agent"
)

func sampleReport() *agent.Report {
	return &agent.Report{
		Query: "What is web grounding?",
		Title: "Web Grounding Overview",
		Rounds: []agent.RoundResult{
			{
				Round:       1,
				Queries:     []string{"web grounding AI"},
				SourceCount: 3,
				Analysis:    "Found comprehensive information.",
			},
		},
		Report: "## Overview\n\nWeb grounding connects LLMs to real-time web data.",
		Sources: []agent.Source{
			{URL: "https://example.com/grounding", Title: "Grounding Guide", Domain: "example.com"},
		},
	}
}

func TestFormatMarkdown(t *testing.T) {
	md := FormatMarkdown(sampleReport())

	checks := []string{
		"# Web Grounding Overview",
		"## Overview",
		"## Search Process",
		"### Round 1",
		"web grounding AI",
		"## Sources",
		"https://example.com/grounding",
	}
	for _, check := range checks {
		if !strings.Contains(md, check) {
			t.Errorf("markdown should contain %q", check)
		}
	}
}

func TestFormatJSON(t *testing.T) {
	data, err := FormatJSON(sampleReport())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed agent.Report
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if parsed.Query != "What is web grounding?" {
		t.Errorf("Query = %q", parsed.Query)
	}
}

func TestWriteBothFiles(t *testing.T) {
	dir := t.TempDir()
	prefix := filepath.Join(dir, "output")

	err := Write(sampleReport(), "both", prefix)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(prefix + ".md"); err != nil {
		t.Error(".md file should exist")
	}
	if _, err := os.Stat(prefix + ".json"); err != nil {
		t.Error(".json file should exist")
	}
}

func TestWriteBothRequiresOutput(t *testing.T) {
	err := Write(sampleReport(), "both", "")
	if err == nil {
		t.Fatal("expected error when format=both and no output prefix")
	}
}
