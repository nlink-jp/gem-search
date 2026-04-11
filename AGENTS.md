# AGENTS.md — gem-search

## Project summary

Agentic web search CLI using Vertex AI Gemini with Google Search Grounding.
Part of util-series. Predecessor: agentic-web-search (frozen, Brave API ToS).

## Build commands

```bash
make build          # Build → dist/gem-search
make test           # Run all tests
make build-all      # Cross-compile for 5 platforms
make clean          # Remove dist/
```

## Module path

`github.com/nlink-jp/gem-search`

## Key structure

```
gem-search/
├── main.go                 ← entry point
├── cmd/root.go             ← cobra command, flag definitions
├── internal/
│   ├── config/             ← GEMSEARCH_* env var loading
│   ├── gemini/             ← Vertex AI client + Grounding
│   ├── agent/              ← agentic loop orchestrator
│   └── output/             ← Markdown/JSON output formatters
├── Makefile
└── docs/                   ← RFP and design documents
```

## Environment variables

- `GEMSEARCH_PROJECT` (required) — GCP project ID
- `GEMSEARCH_LOCATION` (optional, default: us-central1) — Vertex AI region
- `GEMSEARCH_MODEL` (optional, default: gemini-2.5-flash) — model name
- `GEMSEARCH_LANG` (optional) — output language

## Research pipeline

Fixed 3-phase pipeline (not configurable):
1. **Survey** — broad overview, identify key topics and sources
2. **Deep-dive** — fill gaps, gather detailed/specific information
3. **Verify** — cross-check facts, check for contradictions and currency

## Gotchas

- Google Search Grounding and ResponseSchema (JSON mode) are incompatible.
  JSON output is constructed client-side (same pattern as gem-cli).
- Grounding returns redirect URIs that need HTTP HEAD resolution to get
  actual destination URLs.
- Authentication via ADC (`gcloud auth application-default login`).
- No `--max-rounds` flag — always runs 3 phases. This is by design:
  Grounding produces complete answers in 1 call, so the LLM always signals
  "done" after round 1 if given the choice.
