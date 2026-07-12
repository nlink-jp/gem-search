# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [0.4.0] - 2026-07-12

### Removed

- **darwin/amd64 (Intel) pre-built binary.** macOS releases now ship
  **arm64 only**, following the org-wide policy (darwin is Apple-Silicon
  only; universal binaries are not produced). Intel Mac users can still
  build from source via `go install` / `make build`.

### Changed

- **Release archive names unified** to
  `gem-search-v<version>-<os>-<arch>.<ext>` across all platforms, per
  `nlink-jp/.github` CONVENTIONS.md §Release Archive Standard (previously
  `gem-search-<os>-<arch>-v<version>.zip`).
- **Linux archives are now `.tar.gz`** (darwin/windows remain `.zip`).
- **`LICENSE` is now bundled** in every release archive alongside `README.md`.
- **darwin code-signature identifier** is now the canonical `gem-search`
  (previously the build-time `gem-search-darwin-arm64`).

### Fixed

- Store the binary under its canonical name (`gem-search`) inside release
  archives, so unzip yields `gem-search` directly.
- Notarization surfaces the real `notarytool` error (e.g. an expired Apple
  Developer agreement / HTTP 403) instead of a misleading "profile not
  found" message.

No change to the binary's behaviour — this is a packaging / build-config release.

## [0.3.2] - 2026-05-22

### Added

- **Pre-built binary releases for the first time.** A new `package`
  target produces zipped binaries for darwin/amd64, darwin/arm64,
  linux/amd64, linux/arm64, and windows/amd64. Previously gem-search
  was installed via `go install` only.
- **Darwin builds are Developer ID signed and Apple-notarized.**
  `make package` runs `scripts/codesign-darwin.sh` per darwin
  binary and `scripts/notarize-darwin.sh` per darwin zip, following
  the org-wide convention in `nlink-jp/.github` CONVENTIONS.md
  §Code Signing. End users no longer need to bypass Gatekeeper
  with right-click → Open or `xattr -d com.apple.quarantine`;
  local Dropbox-synced (FileProvider-managed) install paths no
  longer SIGKILL the binary on launch.

No behaviour change to the binary itself — feature-wise this is
identical to v0.3.1.

## [0.3.1] - 2026-05-03

### Fixed

- Bump nlk to v0.5.2 to pick up the strip fix: think-tag handling
  no longer truncates LLM responses that explain the literal
  `<think>` tag inside a markdown inline-code span.
- Test isolation: `clearEnv` now resets `HOME` / `XDG_CONFIG_HOME`
  to a temp dir so the developer's real
  `~/.config/gem-search/config.toml` doesn't leak into tests.

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
