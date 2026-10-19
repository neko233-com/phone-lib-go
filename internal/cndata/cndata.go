// Package cndata exposes the embedded mainland-China mobile number prefix
// database (号段库).
//
// The database maps the first seven digits of a mainland Chinese mobile number
// to its province, city, area code, postal code and carrier. The raw phone.dat
// snapshot comes from the iBugCode/phone2region project (MIT licensed, see
// NOTICE.md and LICENSE.phone2region in this directory) and is compiled into
// the binary with go:embed, so lookups are in-memory binary searches.
//
// Parsing is lazy: nothing is decoded at init time. The first call to Lookup or
// Version parses the fixed-size header once, validates it and caches the record
// table offsets.
package cndata

import (
	_ "embed"
	"encoding/binary"
	"fmt"
	"sync"
)

//go:embed phone.dat
var phoneData []byte

// Container layout of phone.dat. All multi-byte integers are little-endian.
//
//	[0:4)            ASCII data version, "YYMM" (e.g. "2502" = 2025-02)
//	[4:8)            uint32 offset of the first fixed-width record
//	[8:indexStart)   content section: UTF-8 "province|city|zip|area_code"
//	                 strings terminated by NUL (0x00) or newline (0x0A)
//	[indexStart:...) recordNum records of recordSize bytes each, sorted by prefix
//
// Each record is:
//
//	[0:4) uint32 seven-digit prefix, e.g. 1312850
//	[4:8) uint32 byte offset of the content string
//	[8]   uint8  carrier type, see CarrierName
const (
	versionSize   = 4
	headerSize    = 8
	recordSize    = 9
	prefixLen     = 7
	carrierOffset = 8
	minNational   = prefixLen
	maxNational   = 11
)

// Record is one row of the China mobile prefix database.
type Record struct {
	// Province is the province name without a suffix, for example "广东".
	Province string
	// City is the city name without a suffix, for example "珠海".
	City string
	// ZipCode is the postal code, for example "519000".
	ZipCode string
	// AreaCode is the city's telephone area code with its leading zero, for
	// example "0756".
	AreaCode string
	// CarrierType is the raw carrier code. See CarrierName.
	CarrierType int
}

// CarrierName returns the Chinese carrier name for a raw carrier code. It
// returns the empty string for 0 and for unknown codes.
//
// A switch is used instead of a package-level map so that the package performs
// no heap work at init time.
func CarrierName(carrierType int) string {
	switch carrierType {
	case 1:
		return "中国移动"
	case 2:
		return "中国联通"
	case 3:
		return "中国电信"
	case 4:
		return "电信虚拟运营商"
	case 5:
		return "联通虚拟运营商"
	case 6:
		return "移动虚拟运营商"
	case 7:
		return "中国广电"
	case 8:
		return "广电虚拟运营商"
	default:
		return ""
	}
}

// database is the parsed, validated header of the embedded snapshot.
type database struct {
	version    string // display version, for example "2025-02"
	indexStart int    // byte offset of the first record
	records    int    // number of records in the table
}

var (
	dbOnce sync.Once
	db     database
)

// ensureLoaded parses and validates the embedded header exactly once.
func ensureLoaded() {
	dbOnce.Do(loadDatabase)
}

// loadDatabase parses the header of phone.dat. The data is compiled into the
// binary, so a corrupt header is a build bug rather than a runtime condition;
// panic with a clear message instead of silently failing every lookup.
func loadDatabase() {
	if len(phoneData) < headerSize {
		panic(fmt.Sprintf("cndata: corrupt embedded phone.dat: %d bytes, need at least %d", len(phoneData), headerSize))
	}

	version := string(phoneData[:versionSize])
	indexStart := int(binary.LittleEndian.Uint32(phoneData[versionSize:headerSize]))
	if indexStart < headerSize || indexStart > len(phoneData) || (len(phoneData)-indexStart)%recordSize != 0 {
		panic(fmt.Sprintf("cndata: corrupt embedded phone.dat header: index offset %d, data size %d", indexStart, len(phoneData)))
	}

	db = database{
		version:    formatVersion(version),
		indexStart: indexStart,
		records:    (len(phoneData) - indexStart) / recordSize,
	}
}

// formatVersion turns the "YYMM" header into "YYYY-MM". Any header that is not
// exactly four ASCII digits is returned unchanged.
func formatVersion(raw string) string {
	if len(raw) != versionSize {
		return raw
	}
	for i := 0; i < len(raw); i++ {
		if raw[i] < '0' || raw[i] > '9' {
			return raw
		}
	}
	// raw is "2502" -> "2025-02".
	return "20" + raw[:2] + "-" + raw[2:]
}

// Version reports the data version encoded in the embedded database, for
// example "2025-02".
func Version() string {
	ensureLoaded()
	return db.version
}

// Upstream metadata row content markers. The 2502 snapshot ends with a
// non-号段 row, prefix 9999999 => "作者|pangongzi|99999|999".
const (
	sentinelProvince = "作者"
	sentinelCity     = "pangongzi"
)

// isRealRecord reports whether rec is a genuine 号段 row that may be returned to
// callers.
//
// The embedded snapshot is not purely 号段 data: it ends with a metadata row
// (prefix 9999999, content "作者|pangongzi|99999|999") that must never surface as
// a 归属地. Rows with an unknown carrier (CarrierType 0) are kept: their
// province, city, zip and area code are real 号段 data, and callers can fall
// back to their own carrier estimate. Only the metadata row and rows whose
// province or city is empty are rejected.
func isRealRecord(rec Record) bool {
	if rec.Province == "" || rec.City == "" {
		return false
	}
	// Defensive: identify the upstream placeholder row by content as well as by
	// prefix, since the record table does not expose the prefix here.
	return !(rec.Province == sentinelProvince && rec.City == sentinelCity)
}

// Lookup resolves a mainland Chinese mobile number (or its first seven digits)
// to its prefix record. It returns false when the input is not 7..11 ASCII
// digits, when the prefix is not in the database, or when the matching row is
// not real 号段 data (see isRealRecord: the upstream metadata row and rows with
// an empty province or city are filtered out).
//
// Rows whose carrier is unknown are returned with CarrierType 0;
// CarrierName(0) is the empty string. Callers that need a carrier estimate can
// fall back to libphonenumber in that case.
//
// The full national significant number may be passed: only the first seven
// digits, the 号段 prefix, select the record. Punctuation is not accepted here;
// callers must strip it beforehand.
func Lookup(national string) (Record, bool) {
	if len(national) < minNational || len(national) > maxNational {
		return Record{}, false
	}

	var prefix uint32
	for i := 0; i < prefixLen; i++ {
		c := national[i]
		if c < '0' || c > '9' {
			return Record{}, false
		}
		prefix = prefix*10 + uint32(c-'0')
	}
	for i := prefixLen; i < len(national); i++ {
		if c := national[i]; c < '0' || c > '9' {
			return Record{}, false
		}
	}

	ensureLoaded()

	lo, hi := 0, db.records-1
	for lo <= hi {
		mid := (lo + hi) / 2
		pos := db.indexStart + mid*recordSize
		switch got := binary.LittleEndian.Uint32(phoneData[pos : pos+4]); {
		case got == prefix:
			contentOffset := int(binary.LittleEndian.Uint32(phoneData[pos+4 : pos+8]))
			if contentOffset < 0 || contentOffset >= len(phoneData) {
				return Record{}, false
			}
			carrierType := int(phoneData[pos+carrierOffset])
			rec := parseRecord(contentString(contentOffset), carrierType)
			if !isRealRecord(rec) {
				return Record{}, false
			}
			return rec, true
		case got < prefix:
			lo = mid + 1
		default:
			hi = mid - 1
		}
	}
	return Record{}, false
}

// contentString returns the content string starting at offset, stopping at its
// NUL or newline terminator. The one copy it makes backs all four string fields
// of the resulting Record, so a successful lookup allocates at most once.
func contentString(offset int) string {
	end := offset
	for end < len(phoneData) {
		b := phoneData[end]
		if b == 0 || b == '\n' {
			break
		}
		end++
	}
	return string(phoneData[offset:end])
}

// parseRecord splits "province|city|zip|area_code" without allocating: every
// field is a substring of content. Missing trailing fields are left empty.
func parseRecord(content string, carrierType int) Record {
	var fields [4]string
	start, n := 0, 0
	for i := 0; i < len(content) && n < len(fields)-1; i++ {
		if content[i] == '|' {
			fields[n] = content[start:i]
			n++
			start = i + 1
		}
	}
	fields[n] = content[start:]

	return Record{
		Province:    fields[0],
		City:        fields[1],
		ZipCode:     fields[2],
		AreaCode:    fields[3],
		CarrierType: carrierType,
	}
}
