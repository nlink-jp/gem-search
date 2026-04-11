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

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `GEMSEARCH_PROJECT` | Yes | — | GCP project ID |
| `GEMSEARCH_LOCATION` | No | `us-central1` | Vertex AI region |
| `GEMSEARCH_MODEL` | No | `gemini-2.5-flash` | Gemini model name |
| `GEMSEARCH_LANG` | No | — | Output language |

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

# Control search depth
gem-search --max-rounds 5 "complex research topic"

# Output in a specific language
gem-search --lang ja "English topic, Japanese report"
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--format` | `markdown` | Output format: `json`, `markdown`, `both` |
| `-o, --output` | (stdout) | Output file prefix (appends `.md`/`.json`) |
| `--max-rounds` | `3` | Maximum autonomous search rounds (cap: 10) |
| `--lang` | (none) | Output language code (e.g. `ja`, `en`) |

## How It Works

```
Query → Gemini + Google Search Grounding
         → LLM analyzes grounding results
           → Decides: search more / done
             → Compiles final report
               → Output (Markdown / JSON)
```

The agent loop runs up to `--max-rounds` iterations. Gemini's Google Search
Grounding provides web search results and pre-extracted content in a single
API call — no separate search API or web scraping required.

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

- [RFP](docs/en/gem-search-rfp.md) — Requirements document

## License

See [LICENSE](LICENSE).
