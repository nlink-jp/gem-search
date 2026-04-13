# gem-search

Agentic web search CLI using Vertex AI Gemini with Google Search Grounding.

Given a natural language question, the tool autonomously searches the web
via Gemini's built-in Google Search Grounding, analyzes the results, and
produces a Markdown or JSON report. Designed as both a standalone research
tool and a pipeline component via stdin/stdout.

## Prerequisites

- **Google Cloud project** with the Vertex AI API enabled
- **Application Default Credentials** — run `gcloud auth application-default login`

## Installation

```bash
git clone https://github.com/nlink-jp/gem-search.git
cd gem-search
make build
# Binary: dist/gem-search
```

## Configuration

Settings are loaded in this order (higher priority wins):

1. **Defaults** — built-in values
2. **TOML file** — `~/.config/gem-search/config.toml` (or `-c` flag)
3. **Environment variables** — `GEMSEARCH_*` (tool-specific) > `GOOGLE_CLOUD_*` (generic)
4. **CLI flags** — highest priority

### Config file

Copy the example and edit:

```bash
mkdir -p ~/.config/gem-search
cp config.example.toml ~/.config/gem-search/config.toml
```

```toml
[gcp]
project  = "your-project-id"
location = "us-central1"

[model]
name = "gemini-2.5-flash"
lang = ""
```

### Environment variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `GEMSEARCH_PROJECT` | Yes | — | GCP project ID |
| `GEMSEARCH_LOCATION` | No | `us-central1` | Vertex AI region |
| `GEMSEARCH_MODEL` | No | `gemini-2.5-flash` | Gemini model name |
| `GEMSEARCH_LANG` | No | — | Output language |
| `GOOGLE_CLOUD_PROJECT` | — | — | Fallback for `GEMSEARCH_PROJECT` |
| `GOOGLE_CLOUD_LOCATION` | — | — | Fallback for `GEMSEARCH_LOCATION` |

## Usage

```bash
# Basic search (Markdown to stdout)
gem-search "What is web grounding?"

# Stdin input (pipeline)
echo "Go context.WithTimeout best practices" | gem-search

# JSON output
gem-search --format json "Vertex AI pricing"

# Both Markdown and JSON to files
gem-search --format both -o ./report "topic to research"
# → ./report.md + ./report.json

# Output in a specific language
gem-search --lang ja "English topic, Japanese report"
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-c, --config` | `~/.config/gem-search/config.toml` | Config file path |
| `--format` | `markdown` | Output format: `json`, `markdown`, `both` |
| `-o, --output` | (stdout) | Output file prefix (appends `.md`/`.json`) |
| `--lang` | (none) | Output language code (e.g. `ja`, `en`) |

## How It Works

```
Query → Phase 1: Survey (broad overview)
         → Phase 2: Deep-dive (fill gaps, gather details)
           → Phase 3: Verify (cross-check, find updates)
             → Compile final report from all 3 phases
               → Output (Markdown / JSON)
```

Every search runs a fixed 3-phase research pipeline. Each phase uses Gemini
with Google Search Grounding for a different purpose — survey maps the
landscape, deep-dive fills gaps, verify checks for contradictions and currency.
This ensures thorough coverage rather than letting the LLM decide to stop early.

## Background

This tool succeeds [agentic-web-search](https://github.com/nlink-jp/agentic-web-search)
(archived), which used the Brave Search API. That project was frozen due to
Brave's aggressive terms of service restrictions on search result usage and the
requirement for paid API registration. DuckDuckGo was also evaluated and
rejected because their HTML endpoint's robots.txt disallows bot access
(`Disallow: /`), making programmatic use scraping rather than legitimate API
access.

Vertex AI Web Grounding avoids these issues entirely — it is an official
Google Cloud API feature with clear terms of service and existing billing
infrastructure.

## Building

```bash
make build       # Build for current platform → dist/gem-search
make build-all   # Cross-compile for 5 platforms
make test        # Run all tests
make clean       # Remove dist/
```

## Documentation

- [Architecture](docs/en/architecture.md) — Design decisions and their rationale
- [RFP](docs/en/gem-search-rfp.md) — Requirements document

## License

See [LICENSE](LICENSE).
