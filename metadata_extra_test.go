package phone_test

import (
	"regexp"
	"strconv"
	"testing"

	phone "github.com/neko233-com/phone-lib-go"
	"github.com/nyaruka/phonenumbers/v2/metadata"
)

// metadataVersionPattern matches the upstream libphonenumber release tag, for
// example "v9.0.40".
var metadataVersionPattern = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)

// chinaDataVersionPattern matches the version of the embedded China 号段
// database, for example "2025-02".
var chinaDataVersionPattern = regexp.MustCompile(`^\d{4}-\d{2}$`)

func TestMetadataVersion(t *testing.T) {
	got := phone.MetadataVersion()
	if got != metadata.Version {
		t.Errorf("MetadataVersion() = %q, want upstream nyaruka/phonenumbers metadata.Version %q", got, metadata.Version)
	}
	if !metadataVersionPattern.MatchString(got) {
		t.Errorf("MetadataVersion() = %q, want a libphonenumber release tag such as \"v9.0.40\"", got)
	}
}

func TestChinaDataVersionFormat(t *testing.T) {
	got := phone.ChinaDataVersion()
	if !chinaDataVersionPattern.MatchString(got) {
		t.Fatalf("ChinaDataVersion() = %q, want a YYYY-MM version such as \"2025-02\"", got)
	}

	// The month component must be a real month.
	month, err := strconv.Atoi(got[5:7])
	if err != nil {
		t.Fatalf("ChinaDataVersion() = %q: month is not numeric: %v", got, err)
	}
	if month < 1 || month > 12 {
		t.Errorf("ChinaDataVersion() = %q has month %d, want 01-12", got, month)
	}
}
