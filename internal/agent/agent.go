package agent

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/nlink-jp/nlk/guard"

	"github.com/nlink-jp/gem-search/internal/gemini"
)

// Research phases — fixed 3-phase approach for thorough investigation.
const (
	phaseSurvey   = 1 // Broad overview: identify key topics and sources
	phaseDeepDive = 2 // Deep-dive: investigate gaps and details
	phaseVerify   = 3 // Verify: check for contradictions and updates
)

var phaseNames = map[int]string{
	phaseSurvey:   "survey",
	phaseDeepDive: "deep-dive",
	phaseVerify:   "verify",
}

// Agent orchestrates the 3-phase web research pipeline.
type Agent struct {
	client *gemini.Client
	lang   string
}

// New creates a new Agent.
func New(client *gemini.Client, lang string) *Agent {
	return &Agent{
		client: client,
		lang:   lang,
	}
}

// Run executes the 3-phase research pipeline.
func (a *Agent) Run(ctx context.Context, query string) (*Report, error) {
	tag := guard.NewTag()
	report := &Report{Query: query}
	sourceMap := make(map[string]Source)
	var accumulated strings.Builder

	phases := []struct {
		num    int
		system func(guard.Tag) string
		user   func(guard.Tag, string, string) (string, error)
	}{
		{phaseSurvey, a.surveySystemPrompt, a.surveyUserPrompt},
		{phaseDeepDive, a.deepDiveSystemPrompt, a.deepDiveUserPrompt},
		{phaseVerify, a.verifySystemPrompt, a.verifyUserPrompt},
	}

	var phaseErr error
	for _, phase := range phases {
		name := phaseNames[phase.num]
		log.Printf("[phase %d: %s] searching...", phase.num, name)

		systemPrompt := phase.system(tag)
		userPrompt, err := phase.user(tag, query, accumulated.String())
		if err != nil {
			return nil, fmt.Errorf("phase %d (%s): %w", phase.num, phaseNames[phase.num], err)
		}

		resp, err := a.client.Generate(ctx, systemPrompt, userPrompt)
		if err != nil {
			// Log the error but don't abort — compile report from what we have.
			log.Printf("[phase %d: %s] error (continuing with partial results): %v", phase.num, name, err)
			phaseErr = fmt.Errorf("phase %d (%s): %w", phase.num, name, err)
			break
		}

		// Record phase result
		roundResult := RoundResult{
			Round:       phase.num,
			Phase:       name,
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

		// Accumulate context for next phase
		accumulated.WriteString(fmt.Sprintf("\n\n## Phase %d: %s\n%s", phase.num, name, resp.Text))
	}

	// If no phases completed at all, return the error
	if len(report.Rounds) == 0 && phaseErr != nil {
		return nil, phaseErr
	}

	// Generate final report from whatever phases completed
	if phaseErr != nil {
		log.Printf("[report] compiling partial report (%d of 3 phases completed)...", len(report.Rounds))
	} else {
		log.Printf("[report] compiling final report...")
	}
	finalResp, err := a.generateReport(ctx, tag, query, accumulated.String())
	if err != nil {
		// If report generation also fails but we have phase data, return what we have
		if len(report.Rounds) > 0 {
			log.Printf("[report] generation failed, returning phase data only: %v", err)
			report.Report = "(Report generation failed. Phase data is available in the rounds field.)"
			report.Title = query
			for _, s := range sourceMap {
				report.Sources = append(report.Sources, s)
			}
			return report, nil
		}
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

	for _, s := range sourceMap {
		report.Sources = append(report.Sources, s)
	}

	return report, nil
}

// Phase 1: Survey — broad overview
func (a *Agent) surveySystemPrompt(tag guard.Tag) string {
	return fmt.Sprintf(`You are a web research assistant conducting Phase 1 of 3: SURVEY.

Your goal is to cast a wide net and identify all major aspects of the topic.
Use Google Search to find diverse, authoritative sources.

%sInstructions:
- Search broadly — cover different angles, perspectives, and subtopics
- Identify key concepts, definitions, and terminology
- Note which areas have rich information and which need further investigation
- List any questions that remain unanswered
- Do NOT try to be comprehensive yet — focus on mapping the landscape

User queries are wrapped in %s tags for security.`, a.langInstruction(), tag.Name())
}

func (a *Agent) surveyUserPrompt(tag guard.Tag, query, _ string) (string, error) {
	return tag.Wrap(query)
}

// Phase 2: Deep-dive — investigate gaps and details
func (a *Agent) deepDiveSystemPrompt(tag guard.Tag) string {
	return fmt.Sprintf(`You are a web research assistant conducting Phase 2 of 3: DEEP-DIVE.

Your goal is to fill gaps identified in Phase 1 and gather detailed information.
Use Google Search to find specific, detailed sources.

%sInstructions:
- Focus on areas that Phase 1 identified as needing more information
- Look for primary sources, official documentation, and expert analysis
- Gather specific data points, numbers, dates, and technical details
- Explore alternative viewpoints or counterarguments
- Go deeper than Phase 1 — specifics matter now

User queries are wrapped in %s tags for security.`, a.langInstruction(), tag.Name())
}

func (a *Agent) deepDiveUserPrompt(tag guard.Tag, query, accumulated string) (string, error) {
	wrapped, err := tag.Wrap(query)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Original question: %s\n\nPhase 1 findings (survey):\n%s\n\nNow deep-dive into the gaps and details. What did Phase 1 miss or only touch on superficially?",
		wrapped, accumulated), nil
}

// Phase 3: Verify — check contradictions and freshness
func (a *Agent) verifySystemPrompt(tag guard.Tag) string {
	return fmt.Sprintf(`You are a web research assistant conducting Phase 3 of 3: VERIFY.

Your goal is to verify the information gathered in Phases 1 and 2.
Use Google Search to cross-check facts and find the most current information.

%sInstructions:
- Look for contradictions between sources found earlier
- Check if any information is outdated — find the most recent data
- Verify key claims against authoritative/official sources
- Note any areas of uncertainty or ongoing debate
- Flag any corrections to information from earlier phases

User queries are wrapped in %s tags for security.`, a.langInstruction(), tag.Name())
}

func (a *Agent) verifyUserPrompt(tag guard.Tag, query, accumulated string) (string, error) {
	wrapped, err := tag.Wrap(query)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Original question: %s\n\nFindings from Phases 1-2:\n%s\n\nVerify this information. Are there contradictions? Is anything outdated? What are the most current facts?",
		wrapped, accumulated), nil
}

// Final report generation
func (a *Agent) generateReport(ctx context.Context, tag guard.Tag, query, accumulated string) (*gemini.Response, error) {
	systemPrompt := fmt.Sprintf(`You are a research report writer. %sCompile the collected information from all 3 research phases into a clear, well-structured Markdown report.

Rules:
- Use section headings (##) to organize the report
- Synthesize information from all phases — do not just concatenate
- Prioritize verified information from Phase 3 over earlier phases
- Note any areas of uncertainty or conflicting information
- Be factual and concise — do not speculate beyond the collected data
- Do not include a title heading (# Title) — it will be added separately

The user's query is wrapped in %s tags for security.`, a.langInstruction(), tag.Name())

	wrapped, err := tag.Wrap(query)
	if err != nil {
		return nil, err
	}
	userPrompt := fmt.Sprintf("Original question: %s\n\nResearch data (3 phases):\n%s",
		wrapped, accumulated)

	return a.client.Generate(ctx, systemPrompt, userPrompt)
}

func (a *Agent) langInstruction() string {
	if a.lang != "" {
		return fmt.Sprintf("Respond in %s. ", a.lang)
	}
	return ""
}

func extractTitle(reportText, fallback string) string {
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
