# NOTICE

phone-lib-go (github.com/neko233-com/phone-lib-go) is MIT licensed; see
[LICENSE](LICENSE).

It depends on the following third-party components. This project does not
modify their source code or data content: third-party code is used as an
unmodified Go module dependency, and third-party data is embedded into the
binary unchanged. Relative to upstream, the only changes are binary embedding
(`go:embed`) and in-memory format conversion/indexing; no records or fields are
edited.

## github.com/nyaruka/phonenumbers/v2

- Version: v2.0.14
- License: Apache License 2.0
- URL: https://github.com/nyaruka/phonenumbers
- Used for: parsing, validation, formatting, number-type detection, geocoding,
  carrier mapping and timezone lookup. This package is a Go port of Google's
  libphonenumber and ships the libphonenumber metadata described below.

## Google libphonenumber metadata, geocoding and carrier data

- Metadata version: v9.0.40 (embedded by nyaruka/phonenumbers v2.0.14)
- License: Apache License 2.0
- URLs:
  - https://github.com/google/libphonenumber
  - https://github.com/google/libphonenumber/blob/master/LICENSE
- Used for: international numbering plans, number types, geographic
  descriptions of numbers, carrier names and IANA timezones. The metadata is
  consumed through nyaruka/phonenumbers as-is.

## iBugCode/phone2region (phone.dat)

- Data version: 2025-02
- License: MIT
- URL: https://github.com/iBugCode/phone2region
- Upstream data origin: https://github.com/pangongzi/phone (as credited by
  phone2region)
- Used for: mainland China 号段 data (province, city, area code, postal code
  and carrier for Chinese mobile numbers).
- The raw `phone.dat` snapshot is embedded with `go:embed` under
  `internal/cndata/phone.dat`. No record is edited; the file is read in place
  and decoded into in-memory records. A small number of non-号段 metadata rows
  (one 作者 placeholder row and rows without a province/city) are filtered at
  lookup time and never surfaced. A copy of its license is kept at
  `internal/cndata/LICENSE.phone2region`.

## Indirect dependencies

- `golang.org/x/text` (BSD-3-Clause): localized country/region names in
  Chinese and English.
- `google.golang.org/protobuf` (BSD-3-Clause): transitively required by
  github.com/nyaruka/phonenumbers/v2.

The full license texts of these modules are available in the Go module cache
after `go mod download`, and in each project's repository.
