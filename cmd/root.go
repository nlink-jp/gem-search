package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/nlink-jp/gem-search/internal/agent"
	"github.com/nlink-jp/gem-search/internal/config"
	"github.com/nlink-jp/gem-search/internal/gemini"
	"github.com/nlink-jp/gem-search/internal/output"
)

const maxQueryLength = 1000

var (
	flagFormat string
	flagOutput string
	flagLang   string
)

var rootCmd = &cobra.Command{
	Use:   "gem-search [query]",
	Short: "Agentic web search using Vertex AI Gemini with Google Search Grounding",
	Long: `gem-search accepts a natural language query, uses Vertex AI Gemini with
Google Search Grounding to autonomously search the web in 3 phases
(survey → deep-dive → verify), and produces a structured Markdown or JSON report.`,
	Args: cobra.MaximumNArgs(1),
	RunE: run,
}

// Execute runs the root command.
func Execute(version string) {
	rootCmd.Version = version
	rootCmd.Flags().StringVar(&flagFormat, "format", "markdown", "Output format: json, markdown, both")
	rootCmd.Flags().StringVarP(&flagOutput, "output", "o", "", "Output file prefix (appends .md/.json)")
	rootCmd.Flags().StringVar(&flagLang, "lang", "", "Output language code (e.g. ja, en)")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	query, err := readQuery(args)
	if err != nil {
		return err
	}
	if len(query) > maxQueryLength {
		return fmt.Errorf("query too long: %d characters (max %d)", len(query), maxQueryLength)
	}

	switch flagFormat {
	case "json", "markdown", "both":
	default:
		return fmt.Errorf("invalid format: %s (must be json, markdown, or both)", flagFormat)
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	cfg.ApplyFlags(flagFormat, flagOutput, flagLang)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	client, err := gemini.NewClient(ctx, cfg.Project, cfg.Location, cfg.Model)
	if err != nil {
		return err
	}

	ag := agent.New(client, cfg.Lang)
	report, err := ag.Run(ctx, query)
	if err != nil {
		return fmt.Errorf("agent error: %w", err)
	}

	return output.Write(report, cfg.Format, cfg.Output)
}

func readQuery(args []string) (string, error) {
	if len(args) > 0 {
		return strings.TrimSpace(args[0]), nil
	}

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return "", fmt.Errorf("no query provided. Usage: gem-search \"your question\" or echo \"question\" | gem-search")
	}

	scanner := bufio.NewScanner(os.Stdin)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("reading stdin: %w", err)
	}

	query := strings.TrimSpace(strings.Join(lines, "\n"))
	if query == "" {
		return "", fmt.Errorf("empty query from stdin")
	}
	return query, nil
}
