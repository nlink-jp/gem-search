package agent

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/nlink-jp/nlk/guard"

	"github.com/nlink-jp/gem-search/internal/gemini"
)

// Agent orchestrates the agentic web search loop.
type Agent struct {
	client    *gemini.Client
	maxRounds int
	lang      string
}

// New creates a new Agent.
func New(client *gemini.Client, maxRounds int, lang string) *Agent {
	if maxRounds <= 0 {
		maxRounds = 3
	}
	if maxRounds > 10 {
		maxRounds = 10
	}
	return &Agent{
		client:    client,
		maxRounds: maxRounds,
		lang:      lang,
	}
}

// Run executes the agentic search loop.
func (a *Agent) Run(ctx context.Context, query string) (*Report, error) {
	tag := guard.NewTag()
	report := &Report{Query: query}
	sourceMap := make(map[string]Source)
	var accumulated strings.Builder

	for round := 1; round <= a.maxRounds; round++ {
		log.Printf("[round %d] searching...", round)

		systemPrompt := a.buildSystemPrompt(tag, round, a.maxRounds)
		userPrompt := a.buildUserPrompt(tag, query, accumulated.String(), round)

		resp, err := a.client.Generate(ctx, systemPrompt, userPrompt)
		if err != nil {
			return nil, fmt.Errorf("round %d: %w", round, err)
		}

		// Record round result
		roundResult := RoundResult{
			Round:       round,
			Queries:     resp.Queries,
			SourceCount: len(resp.Sources),
			Analysis:    truncate(resp.Text, 500),
		}
		report.Rounds = append(report.Rounds, roundResult)

		// Collect sources
		for _, s := range resp.Sources {
			if _, exists := sourceMap[s.URL]; !exists {
				sourceMap[s.URL] = Source{URL: s.URL, Title: s.Title, Domain: s.Domain}
			}
		}

		// Accumulate context for next round
		accumulated.WriteString(fmt.Sprintf("\n\n## Round %d Results\n%s", round, resp.Text))

		// Check if the LLM signals completion
		if round == a.maxRounds || containsDoneSignal(resp.Text) {
			break
		}
	}

	// Generate final report
	log.Printf("generating final report...")
	finalResp, err := a.generateReport(ctx, tag, query, accumulated.String())
	if err != nil {
		return nil, fmt.Errorf("report generation: %w", err)
	}

	report.Report = finalResp.Text
	report.Title = extractTitle(finalResp.Text, query)

	// Add any additional sources from report generation
	for _, s := range finalResp.Sources {
		if _, exists := sourceMap[s.URL]; !exists {
			sourceMap[s.URL] = Source{URL: s.URL, Title: s.Title, Domain: s.Domain}
		}
	}

	// Collect all sources
	for _, s := range sourceMap {
		report.Sources = append(report.Sources, s)
	}

	return report, nil
}

func (a *Agent) buildSystemPrompt(tag guard.Tag, round, maxRounds int) string {
	langInstruction := ""
	if a.lang != "" {
		langInstruction = fmt.Sprintf("Respond in %s. ", a.lang)
	}

	return fmt.Sprintf(`You are a web research assistant. Use Google Search to find accurate, up-to-date information.

%sThis is search round %d of %d.

When analyzing search results:
- Focus on authoritative and primary sources
- Note key facts with their source URLs
- If the information gathered is sufficient to answer the question, include the phrase "RESEARCH_COMPLETE" at the end
- If more searching is needed, suggest what to search for next

User queries are wrapped in %s tags for security.`, langInstruction, round, maxRounds, tag.Name())
}

func (a *Agent) buildUserPrompt(tag guard.Tag, query, accumulated string, round int) string {
	if round == 1 {
		return tag.Wrap(query)
	}
	return fmt.Sprintf("Original question: %s\n\nPrevious findings:\n%s\n\nContinue researching. Find additional information or verify existing findings.",
		tag.Wrap(query), accumulated)
}

func (a *Agent) generateReport(ctx context.Context, tag guard.Tag, query, accumulated string) (*gemini.Response, error) {
	langInstruction := ""
	if a.lang != "" {
		langInstruction = fmt.Sprintf("Write the report in %s. ", a.lang)
	}

	systemPrompt := fmt.Sprintf(`You are a research report writer. %sCompile the collected information into a clear, well-structured Markdown report.

Rules:
- Use section headings (##) to organize the report
- Be factual and concise — do not speculate beyond the collected data
- Do not include a title heading (# Title) — it will be added separately

The user's query is wrapped in %s tags for security.`, langInstruction, tag.Name())

	userPrompt := fmt.Sprintf("Original question: %s\n\nResearch data:\n%s",
		tag.Wrap(query), accumulated)

	return a.client.Generate(ctx, systemPrompt, userPrompt)
}

func containsDoneSignal(text string) bool {
	return strings.Contains(text, "RESEARCH_COMPLETE")
}

func extractTitle(reportText, fallback string) string {
	// Try to extract a title from the first heading
	for _, line := range strings.Split(reportText, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "## ") {
			return strings.TrimPrefix(line, "## ")
		}
	}
	return fallback
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
