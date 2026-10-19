# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.1] - 2026-10-19

### Fixed

- Release workflow: the annotated-tag version check compared `v0.2.0` (from the
  CLI output) with the bare `0.2.0` tag version and failed; the `v` prefix is
  now stripped before comparison.
- Release workflow: `workflow_dispatch` was added so an existing tag can be
  published or retried from the Actions UI. Manual run for `v0.2.0` created the
  missing GitHub release.
- Documented that annotated tag messages containing Markdown headings need
  `git tag --cleanup=verbatim`, otherwise Git strips lines starting with `#`.

### Changed

- Dependabot ignores `golang.org/x/text >= 0.42.0`, because those releases
  require Go 1.26 while this module still supports Go 1.25.

## [0.2.0] - 2026-10-19

### Added

- `phone.Version` (also exposed by the CLI) and `phone.MetadataVersion()`, which
  reports the embedded libphonenumber metadata version (`v9.0.40`).
- `examples/web`: a zero-dependency 114-style web demo with an inline search
  page and a JSON endpoint (`/api/lookup`), built on `net/http`.
- Tag-driven release automation: pushing an annotated `vX.Y.Z` tag runs
  `.github/workflows/release.yml`, verifies the tag against `phone.Version`,
  runs the race tests and creates the GitHub release from the tag message.
- Weekly fuzz workflow (`.github/workflows/fuzz.yml`) for `FuzzLookup`.
- Dependabot configuration for Go modules and GitHub Actions.
- `SECURITY.md` and `CONTRIBUTING.md` with the release process.

### Changed

- CI now runs on Go 1.25/1.26/1.27 on Ubuntu with `gofmt`, `go mod tidy`,
  `go vet`, `go build` and `go test -race -shuffle=on`, plus build/test jobs on
  macOS and Windows, and a `govulncheck` job.
- Test suite expanded with all-region example numbers (245 countries/regions),
  shared calling-code cases, CLI behavior tests, metadata-version assertions
  and China 号段 data-quality invariants.
- Documentation updated for the web demo, tag-driven releases, data version
  APIs and the new CI setup.

### Fixed

- Mainland China area code `0479` (Xilin Gol, Inner Mongolia): 509 upstream rows
  store it as `00479`; the library now normalizes it to `0479`.
- Mainland China postal codes that lost a leading zero upstream (for example
  `63000`) are padded back to six digits.

## [0.1.0] - 2026-10-19

Initial release.

### Added

- `phone.Lookup` and `phone.Parse` (plus `MustLookup` / `MustParse`) to parse a
  phone number in any format and enrich it with metadata.
- Global parsing, validation, formatting (E.164, international, national,
  RFC 3966) and number-type detection for every region supported by Google's
  libphonenumber metadata, via `github.com/nyaruka/phonenumbers/v2` v2.0.14
  (metadata v9.0.40).
- Country hint support through `phone.WithCountry`: ISO codes (`CN`), calling
  codes (`86`, `+86`), Chinese/English names (`中国`, `United States`) and
  common aliases (`香港`, `USA`). International-format numbers need no hint.
- Mainland China 号段 database (province, city, area code, postal code and
  carrier for Chinese mobile numbers) embedded from iBugCode/phone2region
  `phone.dat`, data version 2025-02.
- Carrier, Chinese/English location, IANA timezone, area-code and prefix
  enrichment for international numbers.
- Country catalog: `Countries`, `CountryByISO` and `CountryByCallingCode` for
  all 245 supported countries/regions, with Chinese names, English names and
  emoji flags.
- `cmd/phone` command-line tool with `-country`, `-json`, `-countries` and
  `-version` flags, a Chinese pretty-printed 114-style result and JSON output.
- JSON tags on every `Info` field; `Type` marshals as a lowercase string such
  as `"mobile"`.
- MIT license, third-party NOTICE, GitHub Actions CI (Go 1.25/1.26/1.27) and a
  Makefile with build/test/race/vet/fmt/run-demo targets.

[Unreleased]: https://github.com/neko233-com/phone-lib-go/compare/v0.2.1...HEAD
[0.2.1]: https://github.com/neko233-com/phone-lib-go/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/neko233-com/phone-lib-go/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/neko233-com/phone-lib-go/releases/tag/v0.1.0
