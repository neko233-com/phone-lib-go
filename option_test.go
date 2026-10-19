package phone_test

import (
	"errors"
	"reflect"
	"testing"

	phone "github.com/neko233-com/phone-lib-go"
)

func mustParse(t *testing.T, raw, country string) *phone.Info {
	t.Helper()
	info, err := phone.Parse(raw, country)
	if err != nil {
		t.Fatalf("Parse(%q, %q) error = %v, want nil", raw, country, err)
	}
	if info == nil {
		t.Fatalf("Parse(%q, %q) = nil, want *Info", raw, country)
	}
	return info
}

// TestCNHintForms verifies every documented spelling of the China hint
// resolves to the same result through both Lookup and Parse.
func TestCNHintForms(t *testing.T) {
	want := withoutRaw(mustLookup(t, "13128500205", phone.WithCountry("CN")))
	hints := []string{"CN", "cn", "Cn", " cn ", "86", "+86", "0086", "中国", "China", "china"}

	for _, hint := range hints {
		t.Run(hint, func(t *testing.T) {
			got := withoutRaw(mustLookup(t, "13128500205", phone.WithCountry(hint)))
			if !reflect.DeepEqual(got, want) {
				t.Errorf("Lookup with hint %q = %#v, want %#v", hint, got, want)
			}
			if got := withoutRaw(mustParse(t, "13128500205", hint)); !reflect.DeepEqual(got, want) {
				t.Errorf("Parse with hint %q = %#v, want %#v", hint, got, want)
			}
		})
	}
}

func TestUSHintsForms(t *testing.T) {
	want := withoutRaw(mustLookup(t, "2025550123", phone.WithCountry("US")))
	for _, hint := range []string{"US", "us", "1", "+1", "美国", "United States", "usa", "USA"} {
		t.Run(hint, func(t *testing.T) {
			got := withoutRaw(mustLookup(t, "2025550123", phone.WithCountry(hint)))
			if !reflect.DeepEqual(got, want) {
				t.Errorf("Lookup with hint %q = %#v, want %#v", hint, got, want)
			}
		})
	}
}

func TestHKAndTWHints(t *testing.T) {
	hk := mustLookup(t, "21234567", phone.WithCountry("HK"))
	checkFields(t, []fieldCheck{
		{"E164", hk.E164, "+85221234567"},
		{"ISO", hk.ISO, "HK"},
		{"Location", hk.Location, "中国香港"},
	})
	// Alternate spellings of the Hong Kong hint resolve to the same ISO.
	for _, hint := range []string{"香港", "hk", "852", "+852", "中国香港"} {
		got := mustLookup(t, "21234567", phone.WithCountry(hint))
		if got.ISO != "HK" || got.E164 != hk.E164 {
			t.Errorf("Lookup with hint %q: ISO = %q, E164 = %q, want HK/+85221234567", hint, got.ISO, got.E164)
		}
	}

	tw := mustLookup(t, "0912345678", phone.WithCountry("台湾"))
	checkFields(t, []fieldCheck{
		{"E164", tw.E164, "+886912345678"},
		{"ISO", tw.ISO, "TW"},
		{"Country", tw.Country, "台湾"},
	})
}

// TestHintEquivalenceParse checks Parse(raw, country) is equivalent to
// Lookup(raw, WithCountry(country)) for international input as well.
func TestHintEquivalenceParse(t *testing.T) {
	for _, raw := range []string{"+8613128500205", "13128500205", "07561234567"} {
		t.Run(raw, func(t *testing.T) {
			viaLookup := withoutRaw(mustLookup(t, raw, phone.WithCountry("CN")))
			viaParse := withoutRaw(mustParse(t, raw, "CN"))
			if !reflect.DeepEqual(viaLookup, viaParse) {
				t.Errorf("Parse = %#v, Lookup = %#v, want equal", viaParse, viaLookup)
			}
		})
	}
}

func TestBadHint(t *testing.T) {
	for _, hint := range []string{"ZZ", "Atlantis", "999999", "France 2"} {
		info, err := phone.Lookup("13128500205", phone.WithCountry(hint))
		if !errors.Is(err, phone.ErrUnknownCountry) {
			t.Errorf("Lookup with hint %q error = %v, want ErrUnknownCountry", hint, err)
		}
		if info != nil {
			t.Errorf("Lookup with hint %q info = %#v, want nil", hint, info)
		}
	}

	// An empty hint means "no hint", so a national number reports the missing
	// country rather than an unknown one.
	_, err := phone.Lookup("13128500205", phone.WithCountry(""))
	if !errors.Is(err, phone.ErrCountryRequired) {
		t.Errorf("Lookup with empty hint error = %v, want ErrCountryRequired", err)
	}
}

// TestOptionNotMutated checks that reusing an Option value across calls is safe.
func TestOptionNotMutated(t *testing.T) {
	opt := phone.WithCountry("CN")
	first := mustLookup(t, "13128500205", opt)
	second := mustLookup(t, "13128500205", opt)
	if !reflect.DeepEqual(withoutRaw(first), withoutRaw(second)) {
		t.Errorf("reusing WithCountry option changed result: %#v vs %#v", first, second)
	}
}
