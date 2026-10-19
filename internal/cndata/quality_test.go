package cndata

import (
	"encoding/binary"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// area code / zip shapes used by the data-quality audit.
//
// areaCodeStandard is the normal mainland area-code shape (0 + 2..3 digits).
// The embedded 2502 snapshot has 509 rows from 内蒙古/锡林郭勒 whose upstream
// area code is the 5-digit "00479" (the real code is 0479), so the *asserted*
// invariant is the more lenient areaCodeLenient shape; the strict violations are
// reported and pinned instead.
var (
	areaCodeStandard = regexp.MustCompile(`^0[0-9]{2,3}$`)
	areaCodeLenient  = regexp.MustCompile(`^0[0-9]{2,4}$`)
	zipPattern       = regexp.MustCompile(`^[0-9]{5,6}$`)
)

// dataQualityStats is the result of walking the whole embedded record table.
type dataQualityStats struct {
	rows       int // raw record count
	returnable int // rows Lookup returns (isRealRecord)
	filtered   int // rows Lookup rejects

	filteredPrefixes []int32

	carriers       [9]int // raw rows indexed by CarrierType 0..8
	carriersOther  int
	returnCarriers [9]int // returnable rows indexed by CarrierType 0..8

	provinces map[string]struct{}
	cities    map[string]struct{}
	areaCodes map[string]struct{}
	zips      map[string]struct{}

	minPrefix, maxPrefix         int32
	minReturnable, maxReturnable int32   // maxReturnable stays -1 while none seen
	decades                      [20]int // returnable rows indexed by prefix/100000 (13..19 used)
	decadesOther                 int

	terminators [3]int // 0: NUL (0x00), 1: newline (0x0A), 2: EOF without terminator

	emptyProvince int
	emptyCity     int
	badZip        int
	badFieldCount int
	badPrefix     int // raw prefix outside 1000000..9999999
	duplicates    int // prefix equal to previous
	unsorted      int // prefix lower than previous

	// Area-code shape audit. Raw counters cover every row; returnable counters
	// cover only rows Lookup returns.
	areaCodeLengths    map[int]int    // raw
	strictBadAreaCodes map[string]int // ^0\d{2,3}$ violations, raw
	zipLengths         map[int]int    // raw

	returnAreaCodeLengths      map[int]int
	returnAreaCodeEmpty        int
	returnAreaCodeNotZeroStart int
	returnAreaCodeNonDigit     int
	returnLenientBad           int // first zero, then 2..4 digits: expected 0

	areaCodeExamples []string // first few strict violations, "prefix:value"
}

// collectDataQuality walks every record of the embedded snapshot once and
// gathers the data-quality report shared by TestDataQuality.
func collectDataQuality() *dataQualityStats {
	ensureLoaded()

	s := &dataQualityStats{
		rows:                  db.records,
		provinces:             make(map[string]struct{}),
		cities:                make(map[string]struct{}),
		areaCodes:             make(map[string]struct{}),
		zips:                  make(map[string]struct{}),
		minPrefix:             1<<31 - 1,
		minReturnable:         1<<31 - 1,
		maxReturnable:         -1,
		areaCodeLengths:       make(map[int]int),
		strictBadAreaCodes:    make(map[string]int),
		zipLengths:            make(map[int]int),
		returnAreaCodeLengths: make(map[int]int),
	}

	prev := int32(-1)
	for i := 0; i < db.records; i++ {
		pos := db.indexStart + i*recordSize
		prefix := int32(binary.LittleEndian.Uint32(phoneData[pos : pos+4]))
		contentOffset := int(binary.LittleEndian.Uint32(phoneData[pos+4 : pos+8]))
		carrierType := int(phoneData[pos+carrierOffset])

		switch {
		case prefix < prev:
			s.unsorted++
		case prefix == prev:
			s.duplicates++
		}
		prev = prefix

		if prefix < 1000000 || prefix > 9999999 {
			s.badPrefix++
		}
		if prefix < s.minPrefix {
			s.minPrefix = prefix
		}
		if prefix > s.maxPrefix {
			s.maxPrefix = prefix
		}

		if carrierType >= 0 && carrierType < len(s.carriers) {
			s.carriers[carrierType]++
		} else {
			s.carriersOther++
		}

		// Content-string terminator.
		end := contentOffset
		for end < len(phoneData) && phoneData[end] != 0 && phoneData[end] != '\n' {
			end++
		}
		switch {
		case end >= len(phoneData):
			s.terminators[2]++
		case phoneData[end] == '\n':
			s.terminators[1]++
		default:
			s.terminators[0]++
		}

		content := string(phoneData[contentOffset:end])
		rec := parseRecord(content, carrierType)
		if strings.Count(content, "|")+1 != 4 {
			s.badFieldCount++
		}

		if rec.Province == "" {
			s.emptyProvince++
		}
		if rec.City == "" {
			s.emptyCity++
		}

		s.areaCodeLengths[len(rec.AreaCode)]++
		s.zipLengths[len(rec.ZipCode)]++
		if !areaCodeStandard.MatchString(rec.AreaCode) {
			s.strictBadAreaCodes[rec.AreaCode]++
			if len(s.areaCodeExamples) < 5 {
				s.areaCodeExamples = append(s.areaCodeExamples, fmt.Sprintf("%d:%q", prefix, rec.AreaCode))
			}
		}
		if !zipPattern.MatchString(rec.ZipCode) {
			s.badZip++
		}

		if !isRealRecord(rec) {
			s.filtered++
			if len(s.filteredPrefixes) < 10 {
				s.filteredPrefixes = append(s.filteredPrefixes, prefix)
			}
			continue
		}

		s.returnable++
		if carrierType >= 0 && carrierType < len(s.returnCarriers) {
			s.returnCarriers[carrierType]++
		}
		if prefix < s.minReturnable {
			s.minReturnable = prefix
		}
		if prefix > s.maxReturnable {
			s.maxReturnable = prefix
		}
		decade := int(prefix) / 100000
		if decade >= 0 && decade < len(s.decades) {
			s.decades[decade]++
		} else {
			s.decadesOther++
		}
		s.provinces[rec.Province] = struct{}{}
		s.cities[rec.City] = struct{}{}
		s.areaCodes[rec.AreaCode] = struct{}{}
		s.zips[rec.ZipCode] = struct{}{}

		s.returnAreaCodeLengths[len(rec.AreaCode)]++
		switch {
		case rec.AreaCode == "":
			s.returnAreaCodeEmpty++
		default:
			if rec.AreaCode[0] != '0' {
				s.returnAreaCodeNotZeroStart++
			}
			digits := true
			for j := 0; j < len(rec.AreaCode); j++ {
				if rec.AreaCode[j] < '0' || rec.AreaCode[j] > '9' {
					digits = false
					break
				}
			}
			if !digits {
				s.returnAreaCodeNonDigit++
			}
			if !areaCodeLenient.MatchString(rec.AreaCode) {
				s.returnLenientBad++
			}
		}
	}
	return s
}

// TestDataQuality walks the embedded snapshot and asserts the strongest
// invariants that hold for the pinned 2502 data version. Its log output is the
// source of the README/CHANGELOG coverage figures.
//
// Known upstream quirk documented rather than "fixed" in the data: 509 rows from
// 内蒙古/锡林郭勒 carry the 5-digit area code "00479" (the real code is 0479).
// The raw table is embedded unmodified, so these rows are reported and pinned;
// the asserted area-code invariant is the lenient 0+2..4 digits shape.
func TestDataQuality(t *testing.T) {
	s := collectDataQuality()

	// ---- Compact report -------------------------------------------------
	t.Logf("rows=%d returnable=%d filtered=%d (filtered prefixes=%v)",
		s.rows, s.returnable, s.filtered, s.filteredPrefixes)
	t.Logf("prefix range raw=[%d,%d] returnable=[%d,%d]",
		s.minPrefix, s.maxPrefix, s.minReturnable, s.maxReturnable)
	t.Logf("carriers raw: 0=%d 1=%d 2=%d 3=%d 4=%d 5=%d 6=%d 7=%d 8=%d other=%d",
		s.carriers[0], s.carriers[1], s.carriers[2], s.carriers[3], s.carriers[4],
		s.carriers[5], s.carriers[6], s.carriers[7], s.carriers[8], s.carriersOther)
	t.Logf("distinct provinces=%d cities=%d area codes=%d zip codes=%d",
		len(s.provinces), len(s.cities), len(s.areaCodes), len(s.zips))
	t.Logf("coverage 13x=%d 14x=%d 15x=%d 16x=%d 17x=%d 18x=%d 19x=%d other=%d",
		s.decades[13], s.decades[14], s.decades[15], s.decades[16], s.decades[17],
		s.decades[18], s.decades[19], s.decadesOther)
	t.Logf("terminators NUL=%d newline=%d EOF=%d; areaCode lengths=%v zip lengths=%v",
		s.terminators[0], s.terminators[1], s.terminators[2],
		sortedIntCounts(s.areaCodeLengths), sortedIntCounts(s.zipLengths))
	t.Logf("anomalies: emptyProvince=%d emptyCity=%d badZip=%d badFieldCount=%d badPrefix=%d duplicates=%d unsorted=%d",
		s.emptyProvince, s.emptyCity, s.badZip, s.badFieldCount,
		s.badPrefix, s.duplicates, s.unsorted)
	t.Logf("known area-code quirk: strict ^0[0-9]{2,3}$ violations=%v (examples %v)",
		sortedStringCounts(s.strictBadAreaCodes), s.areaCodeExamples)

	// ---- Invariants -----------------------------------------------------
	// Table size and filtering: exactly the upstream metadata row is filtered.
	if s.rows != 517259 {
		t.Fatalf("raw rows = %d, want 517259", s.rows)
	}
	if s.returnable != 517258 {
		t.Fatalf("returnable rows = %d, want 517258", s.returnable)
	}
	if s.filtered != 1 || len(s.filteredPrefixes) != 1 || s.filteredPrefixes[0] != 9999999 {
		t.Fatalf("filtered = %d %v, want exactly prefix 9999999", s.filtered, s.filteredPrefixes)
	}

	// Prefixes: 7 digits, strictly increasing (hence unique), in the mobile
	// 13x-19x ranges. 9999999 is the documented metadata row.
	if s.duplicates != 0 || s.unsorted != 0 {
		t.Fatalf("prefix ordering broken: duplicates=%d unsorted=%d", s.duplicates, s.unsorted)
	}
	if s.badPrefix != 0 {
		t.Fatalf("%d raw prefixes are not 7-digit numbers", s.badPrefix)
	}
	if s.minPrefix != 1300000 || s.maxPrefix != 9999999 {
		t.Fatalf("raw prefix range = [%d,%d], want [1300000,9999999]", s.minPrefix, s.maxPrefix)
	}
	if s.minReturnable != 1300000 {
		t.Fatalf("min returnable prefix = %d, want 1300000", s.minReturnable)
	}
	if s.maxReturnable < 1900000 || s.maxReturnable > 1999999 {
		t.Fatalf("max returnable prefix = %d, want within [1900000,1999999]", s.maxReturnable)
	}
	var decadeSum int
	for decade := 13; decade <= 19; decade++ {
		if s.decades[decade] == 0 {
			t.Fatalf("no returnable prefixes in the %dx range", decade)
		}
		decadeSum += s.decades[decade]
	}
	if s.decadesOther != 0 {
		t.Fatalf("%d returnable prefixes outside the 13x-19x ranges", s.decadesOther)
	}
	if decadeSum != s.returnable {
		t.Fatalf("13x-19x coverage %d != returnable rows %d", decadeSum, s.returnable)
	}

	// Carriers: every raw row has a carrier code 0..8. Code 0 appears on the
	// nine 167xxxx rows with real locations (kept) and on the filtered metadata
	// row, so raw 0 = 10 and returnable 0 = 9.
	if s.carriersOther != 0 {
		t.Fatalf("%d rows carry a carrier type outside 0..8", s.carriersOther)
	}
	var carrierSum int
	for _, n := range s.carriers {
		carrierSum += n
	}
	if carrierSum != s.rows {
		t.Fatalf("carrier counts %d != raw rows %d", carrierSum, s.rows)
	}
	if s.carriers[0] != 10 || s.returnCarriers[0] != 9 {
		t.Fatalf("CarrierType-0 rows: raw=%d returnable=%d, want 10/9",
			s.carriers[0], s.returnCarriers[0])
	}

	// Returnable rows always have a province and a city.
	if s.emptyProvince != 0 || s.emptyCity != 0 {
		t.Fatalf("empty province/city on raw rows: %d/%d", s.emptyProvince, s.emptyCity)
	}

	// Zip codes are 5 or 6 ASCII digits on every row.
	if s.badZip != 0 {
		t.Fatalf("%d rows have a zip code that is not 5-6 digits", s.badZip)
	}
	var zipSum int
	for length, n := range s.zipLengths {
		if length != 5 && length != 6 {
			t.Fatalf("zip code length %d present, want only 5 or 6", length)
		}
		zipSum += n
	}
	if zipSum != s.rows {
		t.Fatalf("zip length counts %d != raw rows %d", zipSum, s.rows)
	}

	// Every returnable row's area code is non-empty, starts with 0 and is
	// 0+2..4 ASCII digits. (The 509 "00479" rows satisfy the lenient shape.)
	if s.returnAreaCodeEmpty != 0 {
		t.Fatalf("%d returnable rows have an empty area code", s.returnAreaCodeEmpty)
	}
	if s.returnAreaCodeNotZeroStart != 0 {
		t.Fatalf("%d returnable area codes do not start with 0", s.returnAreaCodeNotZeroStart)
	}
	if s.returnAreaCodeNonDigit != 0 {
		t.Fatalf("%d returnable area codes contain non-digits", s.returnAreaCodeNonDigit)
	}
	if s.returnLenientBad != 0 {
		t.Fatalf("%d returnable area codes violate ^0[0-9]{2,4}$", s.returnLenientBad)
	}
	var areaLengthSum int
	for length, n := range s.returnAreaCodeLengths {
		if length < 3 || length > 5 {
			t.Fatalf("%d returnable rows have an out-of-range area-code length of %d", n, length)
		}
		areaLengthSum += n
	}
	if areaLengthSum != s.returnable {
		t.Fatalf("area-code length counts %d != returnable rows %d", areaLengthSum, s.returnable)
	}

	// Pinned upstream quirk: the only strict-shape violations are "00479"
	// (509 rows, 内蒙古/锡林郭勒) and the metadata row's "999".
	if len(s.strictBadAreaCodes) != 2 ||
		s.strictBadAreaCodes["00479"] != 509 || s.strictBadAreaCodes["999"] != 1 {
		t.Fatalf("strict area-code violations = %v, want exactly 00479=509 and 999=1",
			sortedStringCounts(s.strictBadAreaCodes))
	}

	// Every content string is 4 pipe-separated fields and is NUL-terminated.
	if s.badFieldCount != 0 {
		t.Fatalf("%d rows have a content string that is not 4 fields", s.badFieldCount)
	}
	if s.terminators[0] != s.rows || s.terminators[1] != 0 || s.terminators[2] != 0 {
		t.Fatalf("terminators NUL/newline/EOF = %d/%d/%d, want all NUL",
			s.terminators[0], s.terminators[1], s.terminators[2])
	}

	// Coverage sanity for the README figures: all 31 mainland provincial-level
	// divisions are represented, with hundreds of cities and area codes.
	if len(s.provinces) != 31 {
		t.Fatalf("distinct provinces = %d, want 31", len(s.provinces))
	}
	if len(s.cities) < 300 {
		t.Fatalf("distinct cities = %d, want at least 300", len(s.cities))
	}
	if len(s.areaCodes) < 300 {
		t.Fatalf("distinct area codes = %d, want at least 300", len(s.areaCodes))
	}
}

// sortedIntCounts renders a map[int]int as "key:count" pairs in key order.
func sortedIntCounts(m map[int]int) []string {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	out := make([]string, 0, len(keys))
	for _, k := range keys {
		out = append(out, fmt.Sprintf("%d:%d", k, m[k]))
	}
	return out
}

// sortedStringCounts renders a map[string]int as `"value"=count` pairs sorted by
// value.
func sortedStringCounts(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]string, 0, len(keys))
	for _, k := range keys {
		out = append(out, fmt.Sprintf("%q=%d", k, m[k]))
	}
	return out
}
