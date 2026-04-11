package output

import (
	"encoding/json"

	"github.com/nlink-jp/gem-search/internal/agent"
)

// FormatJSON marshals a Report as indented JSON.
func FormatJSON(r *agent.Report) ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}
