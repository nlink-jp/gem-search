# CLAUDE.md — gem-search

**Organization rules (mandatory): https://github.com/nlink-jp/.github/blob/main/CONVENTIONS.md**

## Project overview

Agentic web search CLI using Vertex AI Gemini with Google Search Grounding.
Accepts natural language queries, autonomously searches the web, and outputs
Markdown/JSON reports. Designed as both a standalone tool and a pipeline
component (stdin/stdout).

## Non-negotiable rules

- **Tests are mandatory** — write them with the implementation.
- **Never `go build` directly** — always `make build` (outputs to `dist/`).
- **Docs in sync** — update `README.md` and `README.ja.md` together.
- **Small, typed commits** — `feat:`, `fix:`, `test:`, `chore:`, etc.
- **Security first** — prompt injection defense (nlk/guard), no secrets in code.

## Build & test

```bash
make build          # → dist/gem-search
make test           # or: go test ./...
make build-all      # cross-compile 5 platforms
```

## Configuration

Settings are loaded: defaults → TOML file → env vars → CLI flags.

- **Config file**: `~/.config/gem-search/config.toml` (or `-c` flag)
- **Env vars**: `GEMSEARCH_*` (tool-specific) > `GOOGLE_CLOUD_*` (generic fallback)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `GEMSEARCH_PROJECT` | Yes | — | GCP project ID |
| `GEMSEARCH_LOCATION` | No | `us-central1` | Vertex AI region |
| `GEMSEARCH_MODEL` | No | `gemini-2.5-flash` | Model name |
| `GEMSEARCH_LANG` | No | — | Output language |

## Key dependencies

- `google.golang.org/genai` — Google Gemini SDK (Vertex AI backend)
- `github.com/nlink-jp/nlk` — LLM pre/post-processing
- `github.com/spf13/cobra` — CLI framework
- `github.com/BurntSushi/toml` — config file parsing

## Architecture

- `internal/config/` — TOML + environment variable configuration
- `internal/gemini/` — Vertex AI Gemini client with Google Search Grounding
- `internal/agent/` — agentic loop orchestrator
- `internal/output/` — Markdown/JSON formatters
