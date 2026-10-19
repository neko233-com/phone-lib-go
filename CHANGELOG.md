# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/neko233-com/phone-lib-go/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/neko233-com/phone-lib-go/releases/tag/v0.1.0
