// Package agent provides the agentic search loop orchestrator.
package agent

import "github.com/nlink-jp/gem-search/internal/gemini"

// Report is the final output of an agent run.
type Report struct {
	Query   string        `json:"query"`
	Rounds  []RoundResult `json:"rounds"`
	Report  string        `json:"report"`
	Title   string        `json:"title"`
	Sources []Source      `json:"sources"`
}

// RoundResult captures one round of the agent loop.
type RoundResult struct {
	Round       int      `json:"round"`
	Queries     []string `json:"queries,omitempty"`
	SourceCount int      `json:"source_count"`
	Analysis    string   `json:"analysis"`
}

// Source records a referenced source.
type Source struct {
	URL    string `json:"url"`
	Title  string `json:"title"`
	Domain string `json:"domain,omitempty"`
}

// sourcesFromGemini converts Gemini grounding sources to report sources.
func sourcesFromGemini(gs []gemini.Source) []Source {
	sources := make([]Source, len(gs))
	for i, s := range gs {
		sources[i] = Source{URL: s.URL, Title: s.Title, Domain: s.Domain}
	}
	return sources
}
