# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [0.3.0] - 2026-04-13

### Added

- TOML config file support (`~/.config/gem-search/config.toml`)
- `-c, --config` flag for custom config file path
- `GOOGLE_CLOUD_PROJECT` / `GOOGLE_CLOUD_LOCATION` generic env var fallback
- Config precedence: defaults → TOML → env vars → CLI flags (consistent with gem-image, gem-cli)

## [0.2.2] - 2026-04-12

### Fixes

- Update nlk to v0.5.1 — jsonfix: handle zero-width Unicode spaces and parenthesized prose

## [0.2.1] - 2026-04-12

### Security

- Update nlk to v0.5.0 — handle `guard.Wrap()` error return (tag collision defense-in-depth)

## [0.2.0] - 2026-04-12

### Changed

- Replace configurable `--max-rounds` with fixed 3-phase research pipeline
  (Survey → Deep-dive → Verify) for thorough investigation

### Added

- Partial report fallback: if a phase fails, compile report from completed phases
- Architecture documentation (design rationale, predecessor context)

### Removed

- `--max-rounds` flag (always runs 3 phases)

## [0.1.0] - 2026-04-12

### Added

- Initial implementation
- Vertex AI Gemini with Google Search Grounding integration
- Markdown and JSON output formats with file output support
- Output language specification (`--lang`)
- Pipeline support (stdin/stdout)
