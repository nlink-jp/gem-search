# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/).

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
