package phone_test

import (
	"encoding/json"
	"errors"
	"testing"

	phone "github.com/neko233-com/phone-lib-go"
)

func TestTelURIInput(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{
		"tel:+8613128500205",
		"tel:+86-131-2850-0205",
		"TEL:+8613128500205",
	} {
		info, err := phone.Lookup(raw)
		if err != nil {
			t.Fatalf("Lookup(%q): %v", raw, err)
		}
		if info.E164 != "+8613128500205" {
			t.Errorf("Lookup(%q).E164 = %q, want %q", raw, info.E164, "+8613128500205")
		}
		if info.Location != "广东珠海" {
			t.Errorf("Lookup(%q).Location = %q, want %q", raw, info.Location, "广东珠海")
		}
	}
}

func TestTypeJSONRoundTrip(t *testing.T) {
	t.Parallel()

	for _, typ := range []phone.Type{
		phone.TypeUnknown,
		phone.TypeFixedLine,
		phone.TypeMobile,
		phone.TypeFixedLineOrMobile,
		phone.TypeTollFree,
		phone.TypePremiumRate,
		phone.TypeSharedCost,
		phone.TypeVoIP,
		phone.TypePersonalNumber,
		phone.TypePager,
		phone.TypeUAN,
		phone.TypeVoicemail,
	} {
		encoded, err := json.Marshal(typ)
		if err != nil {
			t.Fatalf("Marshal(%v): %v", typ, err)
		}
		var decoded phone.Type
		if err := json.Unmarshal(encoded, &decoded); err != nil {
			t.Fatalf("Unmarshal(%s): %v", encoded, err)
		}
		if decoded != typ {
			t.Errorf("round trip of %v through %s produced %v", typ, encoded, decoded)
		}
	}

	var numeric phone.Type
	if err := json.Unmarshal([]byte("4"), &numeric); err != nil {
		t.Fatalf("Unmarshal numeric: %v", err)
	}
	if numeric != phone.TypeTollFree {
		t.Errorf("numeric type 4 = %v, want %v", numeric, phone.TypeTollFree)
	}

	if err := json.Unmarshal([]byte(`"not_a_type"`), &numeric); err == nil {
		t.Error("Unmarshal of an unknown type name did not return an error")
	}
}

func TestInfoJSONUnmarshal(t *testing.T) {
	t.Parallel()

	original := phone.MustLookup("13128500205", phone.WithCountry("CN"))
	encoded, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal Info: %v", err)
	}

	var decoded phone.Info
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal Info: %v", err)
	}
	if decoded.Type != phone.TypeMobile {
		t.Errorf("decoded Type = %v, want %v", decoded.Type, phone.TypeMobile)
	}
	if decoded.E164 != original.E164 || decoded.Location != original.Location ||
		decoded.Carrier != original.Carrier || decoded.AreaCode != original.AreaCode ||
		decoded.ZipCode != original.ZipCode {
		t.Errorf("decoded Info differs from original:\ngot  %+v\nwant %+v", decoded, *original)
	}
}

func TestLookupTelURIErrorIsNotCountryRequired(t *testing.T) {
	t.Parallel()

	_, err := phone.Lookup("tel:+")
	if err == nil {
		t.Fatal("Lookup(\"tel:+\") returned no error")
	}
	if errors.Is(err, phone.ErrCountryRequired) {
		t.Fatalf("tel: URI should not be treated as national format: %v", err)
	}
}
