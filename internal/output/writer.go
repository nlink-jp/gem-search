package output

import (
	"fmt"
	"os"

	"github.com/nlink-jp/gem-search/internal/agent"
)

// Write outputs the report according to format and output prefix.
func Write(r *agent.Report, format, outputPrefix string) error {
	switch format {
	case "markdown":
		return writeMarkdown(r, outputPrefix)
	case "json":
		return writeJSON(r, outputPrefix)
	case "both":
		if outputPrefix == "" {
			return fmt.Errorf("--output (-o) is required when --format=both")
		}
		if err := writeMarkdown(r, outputPrefix); err != nil {
			return err
		}
		return writeJSON(r, outputPrefix)
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

func writeMarkdown(r *agent.Report, prefix string) error {
	md := FormatMarkdown(r)
	if prefix == "" {
		_, err := fmt.Print(md)
		return err
	}
	return os.WriteFile(prefix+".md", []byte(md), 0644)
}

func writeJSON(r *agent.Report, prefix string) error {
	data, err := FormatJSON(r)
	if err != nil {
		return fmt.Errorf("formatting JSON: %w", err)
	}
	if prefix == "" {
		_, err := fmt.Println(string(data))
		return err
	}
	return os.WriteFile(prefix+".json", data, 0644)
}
