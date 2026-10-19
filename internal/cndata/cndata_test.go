package cndata

import (
	"encoding/binary"
	"strconv"
	"sync"
	"testing"
)

// realPrefixes are verified rows of the embedded 2502 snapshot. The last two
// keep their real location data despite CarrierType 0 (unknown carrier), which
// Lookup must still return.
var realPrefixes = []struct {
	prefix  string
	want    Record
	carrier string
}{
	{"1312850", Record{Province: "广东", City: "珠海", ZipCode: "519000", AreaCode: "0756", CarrierType: 2}, "中国联通"},
	{"1380013", Record{Province: "北京", City: "北京", ZipCode: "100000", AreaCode: "010", CarrierType: 1}, "中国移动"},
	{"1529900", Record{Province: "新疆", City: "伊犁", ZipCode: "835000", AreaCode: "0999", CarrierType: 1}, "中国移动"},
	{"1300000", Record{Province: "山东", City: "济南", ZipCode: "250000", AreaCode: "0531", CarrierType: 2}, "中国联通"},
	{"1590000", Record{Province: "广东", City: "中山", ZipCode: "528400", AreaCode: "0760", CarrierType: 1}, "中国移动"},
	{"1440000", Record{Province: "福建", City: "福州", ZipCode: "350000", AreaCode: "0591", CarrierType: 1}, "中国移动"},
	{"1674931", Record{Province: "重庆", City: "重庆", ZipCode: "400000", AreaCode: "023", CarrierType: 0}, ""},
	{"1678321", Record{Province: "上海", City: "上海", ZipCode: "200000", AreaCode: "021", CarrierType: 0}, ""},
}

func TestLookupKnownPrefixes(t *testing.T) {
	t.Parallel()

	for _, tt := range realPrefixes {
		tt := tt
		t.Run(tt.prefix, func(t *testing.T) {
			t.Parallel()

			got, ok := Lookup(tt.prefix)
			if !ok {
				t.Fatalf("Lookup(%q) = false, want record %+v", tt.prefix, tt.want)
			}
			if got != tt.want {
				t.Fatalf("Lookup(%q) = %+v, want %+v", tt.prefix, got, tt.want)
			}
			if name := CarrierName(got.CarrierType); name != tt.carrier {
				t.Fatalf("CarrierName(%d) = %q, want %q", got.CarrierType, name, tt.carrier)
			}
		})
	}
}

// TestLookupFullNationalNumber checks that only the first seven digits matter:
// a full 11-digit national number resolves to the same record as its prefix.
func TestLookupFullNationalNumber(t *testing.T) {
	t.Parallel()

	for _, tt := range realPrefixes {
		want, ok := Lookup(tt.prefix)
		if !ok {
			t.Fatalf("Lookup(%q) = false, want true", tt.prefix)
		}

		for _, full := range []string{
			tt.prefix + "0205",
			tt.prefix + "0000",
			tt.prefix + "020", // 10 digits
			tt.prefix,         // 7 digits
		} {
			got, ok := Lookup(full)
			if !ok {
				t.Fatalf("Lookup(%q) = false, want %+v", full, want)
			}
			if got != want {
				t.Fatalf("Lookup(%q) = %+v, want %+v", full, got, want)
			}
		}
	}

	// The canonical example from the task: 13128500205 -> 广东/珠海.
	full, ok := Lookup("13128500205")
	if !ok {
		t.Fatal(`Lookup("13128500205") = false, want true`)
	}
	short, _ := Lookup("1312850")
	if full != short {
		t.Fatalf(`Lookup("13128500205") = %+v, Lookup("1312850") = %+v; want equal`, full, short)
	}
}

func TestLookupInvalid(t *testing.T) {
	t.Parallel()

	invalid := []string{
		"",             // empty
		"1",            // too short
		"123",          // too short
		"123456",       // six digits, one short
		"12ab",         // too short and non-digit
		"12ab567",      // non-digit in first seven
		"131285x",      // non-digit inside prefix
		"1312850 ",     // trailing space is not stripped here
		" 1312850",     // leading space
		"+861312850",   // punctuation is the caller's responsibility
		"131285000000", // 12 digits, too long
		// Definitely missing prefixes (verified against the embedded snapshot).
		"1000000",
		"1234567",
		"1999999",
		"9999998",
		// Non-real rows of the embedded snapshot: only the trailing metadata row
		// is filtered out. CarrierType-0 rows with real locations are kept.
		"9999999", // "作者|pangongzi|99999|999"
	}
	for _, in := range invalid {
		in := in
		t.Run(strconv.Quote(in), func(t *testing.T) {
			t.Parallel()
			if rec, ok := Lookup(in); ok {
				t.Fatalf("Lookup(%q) = %+v, true; want false", in, rec)
			}
		})
	}
}

func TestCarrierName(t *testing.T) {
	t.Parallel()

	cases := map[int]string{
		0:  "",
		1:  "中国移动",
		2:  "中国联通",
		3:  "中国电信",
		4:  "电信虚拟运营商",
		5:  "联通虚拟运营商",
		6:  "移动虚拟运营商",
		7:  "中国广电",
		8:  "广电虚拟运营商",
		9:  "",
		-1: "",
		99: "",
	}
	for code, want := range cases {
		if got := CarrierName(code); got != want {
			t.Errorf("CarrierName(%d) = %q, want %q", code, got, want)
		}
	}
}

func TestVersion(t *testing.T) {
	t.Parallel()

	if got, want := Version(), "2025-02"; got != want {
		t.Fatalf("Version() = %q, want %q", got, want)
	}

	formats := []struct {
		raw  string
		want string
	}{
		{"2502", "2025-02"},
		{"2310", "2023-10"},
		{"1201", "2012-01"},
		{"", ""},
		{"250", "250"},
		{"25022", "25022"},
		{"25x2", "25x2"},
		{"abcd", "abcd"},
	}
	for _, tt := range formats {
		if got := formatVersion(tt.raw); got != tt.want {
			t.Errorf("formatVersion(%q) = %q, want %q", tt.raw, got, tt.want)
		}
	}
}

// TestDatabaseCoverage asserts the embedded table is complete, ordered and
// structurally sound, and that binary search resolves real rows. It reads the
// embedded snapshot through the package internals rather than os.
func TestDatabaseCoverage(t *testing.T) {
	ensureLoaded()

	if got, want := db.version, "2025-02"; got != want {
		t.Fatalf("parsed version = %q, want %q", got, want)
	}
	if got, want := db.records, 517259; got != want {
		t.Fatalf("record count = %d, want %d", got, want)
	}
	if db.records < 500000 {
		t.Fatalf("record count = %d, want at least 500000", db.records)
	}
	if db.indexStart <= headerSize || db.indexStart >= len(phoneData) {
		t.Fatalf("index start = %d, want within (%d, %d)", db.indexStart, headerSize, len(phoneData))
	}

	prev := int32(-1)
	for i := 0; i < db.records; i++ {
		pos := db.indexStart + i*recordSize
		prefix := int32(binary.LittleEndian.Uint32(phoneData[pos : pos+4]))
		if prefix <= prev {
			t.Fatalf("record %d: prefix %d is not strictly greater than previous %d", i, prefix, prev)
		}
		if prefix < 1300000 || prefix > 9999999 {
			t.Fatalf("record %d: prefix %d outside the expected 7-digit range", i, prefix)
		}
		prev = prefix

		contentOffset := int(binary.LittleEndian.Uint32(phoneData[pos+4 : pos+8]))
		if contentOffset < headerSize || contentOffset >= db.indexStart {
			t.Fatalf("record %d (prefix %d): content offset %d outside the content section [%d, %d)",
				i, prefix, contentOffset, headerSize, db.indexStart)
		}
		carrierType := int(phoneData[pos+carrierOffset])
		if carrierType < 0 || carrierType > 8 {
			t.Fatalf("record %d (prefix %d): carrier type %d out of range", i, prefix, carrierType)
		}
	}

	// Spot-check the binary search against raw table rows across the snapshot.
	checked := 0
	for i := 0; i < db.records; i += 1000 {
		pos := db.indexStart + i*recordSize
		prefix := int(binary.LittleEndian.Uint32(phoneData[pos : pos+4]))
		contentOffset := int(binary.LittleEndian.Uint32(phoneData[pos+4 : pos+8]))
		carrierType := int(phoneData[pos+carrierOffset])
		raw := parseRecord(contentString(contentOffset), carrierType)

		got, ok := Lookup(strconv.Itoa(prefix))
		if !isRealRecord(raw) {
			if ok {
				t.Fatalf("Lookup(%d) = %+v, true; non-real row %d must be filtered", prefix, got, i)
			}
			continue
		}
		if !ok {
			t.Fatalf("Lookup(%d) = false for record %d, want %+v", prefix, i, raw)
		}
		if got != raw {
			t.Fatalf("Lookup(%d) = %+v, want %+v (record %d)", prefix, got, raw, i)
		}
		checked++
	}
	if checked < 500 {
		t.Fatalf("only spot-checked %d rows, want at least 500", checked)
	}
}

// TestLookupFiltersNonRealRows walks every row of the embedded table and checks
// that Lookup either returns that exact row (real 号段 data, including
// CarrierType 0 rows with a real location) or filters it out. In particular, it
// proves no lookup can ever surface the upstream "作者|pangongzi" metadata row or
// a row with an empty province/city as a 归属地.
func TestLookupFiltersNonRealRows(t *testing.T) {
	ensureLoaded()

	// The known sentinel row must be rejected.
	if rec, ok := Lookup("9999999"); ok {
		t.Fatalf(`Lookup("9999999") = %+v, true; want false for the upstream metadata row`, rec)
	}
	// A carrier-less row with a real location must be kept.
	if rec, ok := Lookup("1674931"); !ok {
		t.Fatal(`Lookup("1674931") = false, want the carrier-less 重庆 row`)
	} else if want := (Record{Province: "重庆", City: "重庆", ZipCode: "400000", AreaCode: "023", CarrierType: 0}); rec != want {
		t.Fatalf(`Lookup("1674931") = %+v, want %+v`, rec, want)
	}

	var real, filtered, carrierless int
	for i := 0; i < db.records; i++ {
		pos := db.indexStart + i*recordSize
		prefix := int(binary.LittleEndian.Uint32(phoneData[pos : pos+4]))
		contentOffset := int(binary.LittleEndian.Uint32(phoneData[pos+4 : pos+8]))
		carrierType := int(phoneData[pos+carrierOffset])
		raw := parseRecord(contentString(contentOffset), carrierType)

		got, ok := Lookup(strconv.Itoa(prefix))
		if !isRealRecord(raw) {
			filtered++
			if ok {
				t.Fatalf("Lookup(%d) = %+v, true; non-real row %d must be filtered", prefix, got, i)
			}
			continue
		}

		real++
		if !ok {
			t.Fatalf("Lookup(%d) = false for real record %d (%+v)", prefix, i, raw)
		}
		if got != raw {
			t.Fatalf("Lookup(%d) = %+v, want %+v (record %d)", prefix, got, raw, i)
		}
		if got.Province == sentinelProvince || got.City == sentinelCity ||
			got.Province == "" || got.City == "" || got.CarrierType < 0 || got.CarrierType > 8 {
			t.Fatalf("Lookup(%d) returned a non-real record: %+v", prefix, got)
		}
		if got.CarrierType == 0 {
			carrierless++
		}
	}

	// Only the trailing metadata row is non-real; the nine CarrierType-0 rows
	// with real locations are returned, leaving 517258 returnable rows.
	if filtered != 1 {
		t.Fatalf("filtered %d non-real rows, want 1 (the metadata row)", filtered)
	}
	if real != 517258 {
		t.Fatalf("returned %d real rows, want 517258", real)
	}
	if carrierless != 9 {
		t.Fatalf("returned %d CarrierType-0 rows, want 9", carrierless)
	}
	if real+filtered != db.records {
		t.Fatalf("real (%d) + filtered (%d) != table size (%d)", real, filtered, db.records)
	}
}

// TestLookupConcurrent exercises the lazy parser and the read-only table from
// many goroutines; run with -race to catch data races.
func TestLookupConcurrent(t *testing.T) {
	t.Parallel()

	const (
		goroutines = 16
		iterations = 500
	)

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func(g int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				tt := realPrefixes[(g+i)%len(realPrefixes)]
				rec, ok := Lookup(tt.prefix + "0205")
				if !ok {
					t.Errorf("goroutine %d: Lookup(%q) = false, want record", g, tt.prefix+"0205")
					return
				}
				if rec != tt.want {
					t.Errorf("goroutine %d: Lookup(%q) = %+v, want %+v", g, tt.prefix+"0205", rec, tt.want)
					return
				}
			}
		}(g)
	}
	wg.Wait()
}

// TestLookupAllocations documents that a successful lookup allocates at most the
// single content string backing the four Record fields.
func TestLookupAllocations(t *testing.T) {
	ensureLoaded()

	allocs := testing.AllocsPerRun(100, func() {
		rec, ok := Lookup("1312850")
		if !ok || rec.Province != "广东" {
			t.Fatalf(`Lookup("1312850") = %+v, %v; want 广东 record`, rec, ok)
		}
	})
	if allocs > 1 {
		t.Fatalf("Lookup allocated %.1f objects per call, want at most 1", allocs)
	}
}

var benchmarkRecord Record

func BenchmarkLookup(b *testing.B) {
	b.ReportAllocs()
	if _, ok := Lookup("1312850"); !ok {
		b.Fatal(`Lookup("1312850") = false, want true`)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchmarkRecord, _ = Lookup("1312850")
	}
}
