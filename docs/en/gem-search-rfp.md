# RFP: gem-search

> Generated: 2026-04-12
> Status: Draft

## 1. Problem Statement

Researching topics on the web involves an inefficient manual loop of searching,
reading results, and refining queries. Existing tools like `product-research`
and `news-collector` use Vertex AI Web Grounding but are purpose-built for
specific use cases and cannot serve general-purpose research.

gem-search is an agentic web search CLI using Vertex AI Web Grounding. It
autonomously executes web search, information gathering, and report generation
from natural language queries, outputting Markdown/JSON. Beyond standalone use,
it functions as an information-gathering component in pipelines via stdin/stdout.

It inherits the UI design of the frozen `agentic-web-search` and replaces the
backend with Vertex AI Web Grounding — an official API feature with no terms
of service concerns. Vertex AI billing infrastructure is already in place.

Primary users are nlink-jp developers.

## 2. Functional Specification

### Commands / API Surface

```bash
# Basic (argument, Markdown to stdout)
gem-search "query in natural language"

# stdin input
echo "query" | gem-search

# Output format
gem-search --format json "..."
gem-search --format markdown "..."
gem-search --format both -o ./result "..."
# → ./result.md + ./result.json

# Agent control
gem-search --max-rounds 5 "..."

# Output language
gem-search --lang ja "..."
```

Key flags:

| Flag | Description | Default |
|------|-------------|---------|
| `--format` | Output format (`json`, `markdown`, `both`) | `markdown` |
| `-o` | Output file prefix (appends `.md`, `.json`) | None (stdout) |
| `--max-rounds` | Maximum autonomous search rounds | TBD |
| `--lang` | Output language code | None (as-is) |

### Input / Output

- **Input**: Natural language query (command argument or stdin)
- **Output**:
  - Markdown: Research report with source URLs
  - JSON: Structured data (queries, source URLs/titles, final answer)
  - File output with `-o`, stdout otherwise

### Configuration

Environment variable convention (`GEMSEARCH_` prefix):

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `GEMSEARCH_PROJECT` | Yes | — | GCP project ID |
| `GEMSEARCH_LOCATION` | No | `us-central1` | Vertex AI region |
| `GEMSEARCH_MODEL` | No | `gemini-2.5-flash` | Model name |
| `GEMSEARCH_LANG` | No | — | Output language (equivalent to `--lang`) |

Authentication: ADC (`gcloud auth application-default login`)

### External Dependencies

- **Vertex AI API**: Gemini + Google Search Grounding
- **nlk**: LLM pre/post-processing library (guard, strip, jsonfix, backoff, validate)
- **cobra**: CLI framework
- **google-genai Go SDK**: `google.golang.org/genai`

## 3. Design Decisions

- **Language: Go** — Single-binary distribution (critical for pipeline use),
  nlk integration, Vertex AI Go SDK (`google.golang.org/genai`) with proven
  Grounding implementation in gem-cli. Advancing the Go adoption of Vertex AI
  features.
- **Vertex AI Web Grounding** — Official API feature. No terms of service
  concerns (avoiding the Brave API ToS risks experienced with
  agentic-web-search). GCP billing infrastructure already established.
- **UI design from `agentic-web-search`** — Flag system (`--format`, `-o`,
  `--max-rounds`, `--lang`), output structure (Markdown+JSON), stdin/stdout
  support carried over.
- **Grounding pattern from gem-cli** — Google Search Grounding enablement
  and grounding_metadata source extraction from
  `internal/client/client.go`.
- **nlk integration**: guard (prompt injection defense), strip (thinking tag
  removal), jsonfix (JSON repair), backoff (retry), validate (output
  validation)
- **Fully independent from product-research/news-collector** — No integration
  with purpose-built tools.
- **Out of scope**:
  - GUI/TUI
  - MCP Tool (future extension candidate)
  - Local LLM support (Vertex AI only)

## 4. Development Plan

Cross-phase principles: **Security first**, **design for testability**,
CONVENTIONS.md compliance.

### Phase 1: Core

- Project scaffold
- Vertex AI Gemini client (Google Search Grounding enabled, nlk integration)
- Agent loop (query → Grounding search → result analysis → continue/done)
- Grounding metadata source extraction (URL + title)
- JSON/Markdown output (`--format`, `-o`)
- stdin/argument input support
- Unit tests (mockable Vertex AI client design)

### Phase 2: Features

- Multi-round autonomous search (`--max-rounds`)
- Output language specification (`--lang` / `GEMSEARCH_LANG`)
- Enhanced error handling (API quota, network errors, rate limits)
- Prompt tuning

### Phase 3: Release

- README.md / README.ja.md
- CHANGELOG.md
- E2E testing (live Vertex AI environment)
- Submodule registration in util-series
- Release procedure (per CONVENTIONS.md)

Phase 1 is independently reviewable.

## 5. Required API Scopes / Permissions

- **Vertex AI API**: Must be enabled in GCP project
- **Authentication**: ADC (Application Default Credentials)
- **IAM role**: `roles/aiplatform.user` (Vertex AI endpoint invocation)
- **Billing**: Gemini API call charges + Google Search Grounding usage fees

## 6. Series Placement

- **Series**: util-series
- **Reason**: Pipe-friendly data processing CLI with JSON/Markdown output and
  stdin/stdout pipeline support. Uses established Vertex AI infrastructure with
  proven Grounding implementation in gem-cli. Not experimental (lab-series) —
  built on known, working patterns.

## 7. External Platform Constraints

### Vertex AI API

- **Grounding + JSON mode incompatibility**: `ResponseSchema` (structured
  output) cannot be used when Grounding is enabled. JSON is constructed
  client-side, following the same pattern as gem-cli.
- **Billing**: Google Search Grounding incurs charges separate from API calls.
- **Region restrictions**: Grounding-capable regions may be limited
  (`us-central1` is reliably supported).
- **Rate limits**: Subject to Vertex AI Gemini API default QPM/TPM limits.
- **Grounding source handling**: URLs and titles extracted from
  `grounding_metadata.grounding_chunks[].web`. Google's terms of service apply
  (no aggressive restrictions like Brave API).

---

## Discussion Log

1. **Predecessor project**: `agentic-web-search` was implemented with Brave
   Search API but frozen due to aggressive ToS restrictions (snippet storage,
   redistribution, and AI training prohibitions) and the psychological barrier
   of paid registration. DuckDuckGo was rejected for robots.txt violation.
   Decision: if paying anyway, Vertex AI Web Grounding is more rational given
   existing GCP billing infrastructure.

2. **Tool naming**: `gem-search` — symmetry with gem-cli, clearly
   Gemini-based.

3. **Language choice**: Go. Policy to advance Go adoption for Vertex AI
   features. Reuses gem-cli's Grounding implementation pattern rather than
   Python patterns from product-research/news-collector. Single-binary
   distribution important for pipeline use.

4. **Model selection**: Confirmed Web Grounding works with Gemini 2.5 Flash
   (not Pro-only). Default is `gemini-2.5-flash` for cost efficiency,
   configurable via `GEMSEARCH_MODEL`.

5. **Pipeline integration**: MCP Tool and library approaches are out of initial
   scope. CLI + stdin/stdout Unix-style composition is the pipeline mechanism.

6. **Series placement**: util-series. Built on established Vertex AI
   infrastructure with gem-cli precedent. Not experimental.

7. **Grounding constraint**: Grounding + JSON mode incompatibility is a known,
   solved pattern from gem-cli. Client-side JSON construction approach carried
   over.
