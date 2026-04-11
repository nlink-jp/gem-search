// Package output provides Markdown and JSON output formatters.
package output

import (
	"fmt"
	"strings"

	"github.com/nlink-jp/gem-search/internal/agent"
)

// FormatMarkdown formats a Report as a Markdown document.
func FormatMarkdown(r *agent.Report) string {
	var sb strings.Builder

	title := r.Title
	if title == "" {
		title = r.Query
	}
	fmt.Fprintf(&sb, "# %s\n\n", title)

	if r.Report != "" {
		sb.WriteString(r.Report)
		sb.WriteString("\n\n")
	}

	// Search rounds detail
	if len(r.Rounds) > 0 {
		sb.WriteString("---\n\n## Search Process\n\n")
		for _, round := range r.Rounds {
			fmt.Fprintf(&sb, "### Round %d\n\n", round.Round)

			if len(round.Queries) > 0 {
				fmt.Fprintf(&sb, "**Search queries:** %s\n\n", strings.Join(round.Queries, ", "))
			}

			if round.SourceCount > 0 {
				fmt.Fprintf(&sb, "**Sources found:** %d\n\n", round.SourceCount)
			}

			if round.Analysis != "" {
				fmt.Fprintf(&sb, "%s\n\n", round.Analysis)
			}
		}
	}

	// Sources
	if len(r.Sources) > 0 {
		sb.WriteString("## Sources\n\n")
		for i, s := range r.Sources {
			title := s.Title
			if title == "" {
				title = s.URL
			}
			fmt.Fprintf(&sb, "%d. [%s](%s)\n", i+1, title, s.URL)
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
