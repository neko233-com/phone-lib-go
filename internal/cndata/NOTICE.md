# Third-party data notice

## phone2region — mainland-China mobile number prefix database

`internal/cndata/phone.dat` is the mainland-China mobile phone number prefix
(号段) database from the **phone2region** project:

- Upstream repository: <https://github.com/iBugCode/phone2region>
- Upstream data file: <https://raw.githubusercontent.com/iBugCode/phone2region/main/phone.dat>
- Upstream license: MIT, Copyright (c) 2017 conzi
- License text: [`LICENSE.phone2region`](LICENSE.phone2region)

### Embedded snapshot

| Field | Value |
| --- | --- |
| Data version | `2502` (2025-02), exposed as `cndata.Version() == "2025-02"` |
| Records (7-digit prefixes) | 517,259 |
| File size | 4,665,431 bytes |
| SHA-256 | `5858836c6a472a706a690a55419d589ccbcd9e72721852642280d9e10e379bc2` |

The snapshot is copied verbatim from upstream and compiled into the binary with
`//go:embed`; `cndata.go` parses it lazily at first use. The MIT license text of
the upstream project is reproduced in `LICENSE.phone2region`, as required by the
license.

### Non-号段 metadata row filtered at lookup time

The upstream snapshot is not purely 号段 data and is embedded unmodified. Its
final row, prefix `9999999` => `作者|pangongzi|99999|999`, is a metadata row and
is not returned by `cndata.Lookup`; rows with an empty province or city would be
filtered for the same reason. The raw table has 517,259 records; 517,258 are
returnable.

The snapshot also contains nine `167xxxx` rows (e.g. `1674931`, `1678306`–
`1678321`) whose carrier is unknown (`CarrierType` 0). These are kept: their
province, city, zip code and area code are real 号段 data. `Lookup` returns them
with `CarrierType` 0, for which `CarrierName(0)` is the empty string; callers
that need a carrier estimate can fall back to libphonenumber in that case.

### Data-quality notes

A full audit of the embedded snapshot (`TestDataQuality` in
`quality_test.go`) reports:

- 517,259 raw rows, of which 517,258 are returnable 号段 rows and 1 is the
  filtered metadata row.
- All prefixes are 7-digit, sorted and unique; returnable prefixes span
  1300000–1999979 (the metadata row is 9999999).
- Every returnable row has a non-empty province and city. Zip codes are 5 or 6
  digits on every row. Area codes are 3–5 characters, start with `0` and
  otherwise contain only digits.
- Known upstream quirk: 509 returnable rows, all `内蒙古/锡林郭勒`, carry the
  five-digit area code `00479` (the real code is `0479`). The snapshot is
  embedded unmodified, so the value is passed through as-is.

The Go code in this directory (`cndata.go`, `cndata_test.go`,
`quality_test.go`) is part of `phone-lib-go`; the MIT notice above applies to the
embedded data.
