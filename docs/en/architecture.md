# Architecture: gem-search

> Last updated: 2026-04-12

## Why This Tool Exists

This tool succeeds [agentic-web-search](https://github.com/nlink-jp/agentic-web-search)
(archived). That project explored agentic web search using third-party search
APIs (DuckDuckGo, then Brave). Both were rejected:

- **DuckDuckGo** — No Web Search API exists. The HTML endpoint's robots.txt
  is `Disallow: /`. Using it is scraping.
- **Brave Search API** — Functional, but aggressive ToS restrictions on data
  reuse (storage, redistribution, AI training prohibitions) and the
  psychological barrier of paid registration for a personal tool.

The conclusion: if paying for search anyway, Vertex AI Web Grounding is more
rational — it is an official Google Cloud API feature, billing infrastructure
is already in place, and there are no ToS ambiguities about LLM consumption of
search results (that is its designed purpose).

## Why a Fixed 3-Phase Pipeline (not configurable rounds)

The original `agentic-web-search` used `--max-rounds` to let the LLM decide
when to stop searching. With Vertex AI Grounding, this failed: Grounding
returns a complete answer with sources in a single API call, so the LLM
always concluded "sufficient information" after round 1.

The fix is not to force more rounds of the same thing. Instead, each phase
asks a fundamentally different question:

| Phase | Purpose | Why it can't be skipped |
|-------|---------|------------------------|
| **Survey** | Map the landscape — identify key topics, terminology, major sources | Without breadth, you miss entire aspects of the topic |
| **Deep-dive** | Fill gaps — specific data, primary sources, expert analysis | Without depth, the report stays superficial |
| **Verify** | Cross-check — contradictions, outdated info, most current data | Without verification, errors from phases 1-2 propagate |

This is not configurable because skipping a phase degrades quality in
predictable, specific ways. Three phases is the minimum for thorough research.

## Why Vertex AI Web Grounding (not a separate search API)

Grounding is a single API call that performs search + content extraction +
answer generation. This eliminates the need for:

- A separate search engine package (`internal/search/`)
- A web fetcher with robots.txt and SSRF protection (`internal/fetch/`)
- Snippet storage/redistribution compliance concerns
- Rate limiting against search engines

The entire search infrastructure is Google's responsibility. Our code focuses
on what only we can do: the 3-phase research strategy and report compilation.

## Why Go

- **Single-binary distribution** — critical for pipeline use (`echo "query" | gem-search --format json | jq`)
- **nlk integration** — guard (prompt injection defense), strip (thinking tag removal)
- **gem-cli precedent** — Grounding implementation pattern already proven in Go
- **Advancing Go adoption for Vertex AI** — product-research and news-collector
  are Python; building gem-search in Go validates the Go SDK for Grounding
  workloads

## Security

| Concern | Mitigation |
|---------|-----------|
| Prompt injection | User queries wrapped with nlk/guard nonce-tagged XML in all 4 LLM calls |
| No SSRF risk | No URL fetching — Grounding handles web access internally |
| No robots.txt concern | API access, not scraping |
| No ToS ambiguity | Grounding is designed for LLM consumption of search results |
| Credentials | ADC only, no secrets in code or config files |
