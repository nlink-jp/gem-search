# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [0.5.2] - 2026-11-24

### Changed

- `make vuln` — and so `make package` — scans every platform in `PLATFORMS`,
  with the environment `build-all` builds each with. It scanned the host's
  build alone, so an advisory in a Linux- or Windows-only package could not
  stop a release.

### Security

- **OpenTelemetry is updated from v1.44.0 to v1.47.0** (`go.opentelemetry.io/otel`,
  `otel/trace`, `otel/metric`; they come in through the Vertex AI SDK).
  `govulncheck` found GO-2026-6974 (CVE-2026-128085, published 2026-11-23:
  attribute value length limits could be bypassed with U+FFFD runes) on the tracing path of the Google client libraries, which gem-search
  reaches on its Gemini calls and when it resolves a grounding link.
  The same update fixes GO-2026-6975 (memory allocation amplification in W3C
  baggage parsing in `go.opentelemetry.io/otel`), whose package gem-search imports
  but whose affected code it does not call.
  Two advisories remain that gem-search does not call: GO-2026-6278 in
  `github.com/gorilla/websocket` (imported by the SDK; fixed only in an
  untagged commit) and GO-2026-5932 against `x/crypto/openpgp`
  (unmaintained, no fixed release).

## [0.5.1] - 2026-11-14

### Security

- **`golang.org/x/net` is updated from v0.38.0 to v0.62.0.** The connection
  to Vertex AI ran on x/net's own HTTP/2, not the standard library's: the
  SDK's default client configures it that way when a tool does not supply a
  transport of its own. Measured — a request through the same client
  against a local server was handled by `golang.org/x/net/http2`. Three
  advisories apply to an HTTP/2 client there (GO-2026-4918, -6610, -6611);
  each needs the server end of the connection, which is Google's endpoint
  under TLS. With v0.62.0 built by Go 1.27 the same request is handled by
  the standard library, where they are fixed from Go 1.27.2.
- `grpc` moves from v1.66.2 to v1.83.2, `x/text` to v0.43.0 and `x/crypto`
  to v0.58.0, for advisories `govulncheck` traced through code this build
  does not run (the gRPC path is taken only when S2A is switched on).
  `grpc` v1.83.2 requires `cloud.google.com/go/auth` v0.18.2, so that
  moves from v0.9.3 with it. The Vertex AI SDK is unchanged (v1.53.0).
  The new set was checked against a local server, not against the live
  endpoint.

### Changed

- **Building from source needs Go 1.27.4 or later** (`go.mod`; it was
  1.26.2), so that no build uses a standard library with the HTTP/2
  advisories above.
- **`make package` runs the vulnerability check first** (`make vuln`), so a
  release cannot be built past a finding.

## [0.5.0] - 2026-11-04

### Changed

- **An unknown key in `config.toml` is now an error** that names the key,
  where it used to be skipped without a word. A typo such as `lnag`
  no longer leaves a setting that seems to do nothing. Remove or correct
  the key the message names.
- **The config file is read by nlink-jp/toml** (v0.1.1) instead of the
  third-party `BurntSushi/toml`, so the tool carries no TOML code from
  outside the organization.

### Fixed

- **`make verify-release` now fails closed.** Its last block chained unzip, the
  packaged binary's `--version` and `spctl` with `&&` and ended the whole chain
  in `|| true`, so a zip that did not unpack or a binary that did not run exited
  0 and the upload proceeded. Each step is now judged on its own, the packaged
  binary's `--version` must contain the tag being released, and only the
  informational `spctl` line may be ignored. Matches the org template
  (CONVENTIONS.md §Code Signing → Verifying a release).
- **The Linux archives no longer carry macOS file metadata.** macOS `tar` wrote
  each bundled file's extended attributes (`com.apple.provenance`, and a Dropbox
  attribute where the tree is synced) into the `.tar.gz` twice: as AppleDouble
  `._` members, which GNU tar extracts as stray `._<name>` files beside the real
  ones, and as `LIBARCHIVE.xattr.*` / `SCHILY.xattr.*` pax headers, which it
  reports as unknown keywords. `make package` now archives with
  `COPYFILE_DISABLE=1 tar --no-xattrs`; each setting stops one of the two.
  Archives already published still carry them; the files themselves are
  unaffected.

### Internal

- `make verify-release` also judges each Linux archive: no AppleDouble or other
  macOS metadata members — listed with `--options 'tar:!mac-ext'`, because a
  plain macOS listing folds `._` members away — no extended attributes as pax
  headers, and exactly the canonical binary, `README.md` and `LICENSE`, compared
  in the C locale.
- The Linux-archive check in `make verify-release` reads each archive's pax
  headers with Python's `tarfile` instead of grepping the decompressed stream,
  which also matched file text that names the keywords (a bundled CHANGELOG,
  for one).

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
