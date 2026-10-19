package phone_test

import (
	"testing"

	phone "github.com/neko233-com/phone-lib-go"
	pn "github.com/nyaruka/phonenumbers/v2"
)

// numberingPlanOwnerISO records regions whose generated example number is
// indistinguishable from the numbering plan's main region: +212 is shared by
// Western Sahara (EH) and Morocco (MA), +590 by Saint Martin (MF) and Saint
// Barthélemy (BL). libphonenumber therefore reports the main region for those
// E.164 numbers; that is upstream behavior, not a bug in this library.
var numberingPlanOwnerISO = map[string]string{
	"EH": "MA",
	"MF": "BL",
}

// TestWorldExamples checks every supported country/region against
// libphonenumber's own example number: the E.164 form must be a valid number
// whose detected region is the region itself (or its numbering-plan owner for
// the shared calling codes above), and the national format must round-trip back
// to the same E.164 with a country hint.
func TestWorldExamples(t *testing.T) {
	countries := phone.Countries()
	if len(countries) == 0 {
		t.Fatal("Countries() is empty")
	}

	tested, skipped := 0, 0
	for _, country := range countries {
		country := country
		t.Run(country.ISO, func(t *testing.T) {
			example := pn.GetExampleNumber(country.ISO)
			if example == nil {
				skipped++
				t.Skipf("libphonenumber has no example number for %s", country.ISO)
			}

			e164 := pn.Format(example, pn.E164)
			if e164 == "" {
				t.Fatalf("Format(example, E164) for %s is empty", country.ISO)
			}

			info, err := phone.Lookup(e164)
			if err != nil {
				t.Fatalf("Lookup(%q) [%s] error = %v, want nil", e164, country.ISO, err)
			}
			if info == nil {
				t.Fatalf("Lookup(%q) [%s] = nil, want *Info", e164, country.ISO)
			}

			wantISO := country.ISO
			if owner, ok := numberingPlanOwnerISO[country.ISO]; ok {
				wantISO = owner
			}
			if info.ISO != wantISO {
				t.Errorf("Lookup(%q) [%s].ISO = %q, want %q", e164, country.ISO, info.ISO, wantISO)
			}
			if !info.Valid {
				t.Errorf("Lookup(%q) [%s].Valid = false, want true", e164, country.ISO)
			}
			if info.E164 != e164 {
				t.Errorf("Lookup(%q) [%s].E164 = %q, want %q", e164, country.ISO, info.E164, e164)
			}
			if info.National == "" {
				t.Errorf("Lookup(%q) [%s].National is empty", e164, country.ISO)
			}
			if info.International == "" {
				t.Errorf("Lookup(%q) [%s].International is empty", e164, country.ISO)
			}
			if info.Country == "" {
				t.Errorf("Lookup(%q) [%s].Country is empty", e164, country.ISO)
			}
			if info.CountryEn == "" {
				t.Errorf("Lookup(%q) [%s].CountryEn is empty", e164, country.ISO)
			}
			if info.CountryCode == 0 {
				t.Errorf("Lookup(%q) [%s].CountryCode = 0, want the region calling code", e164, country.ISO)
			}

			// The national format must resolve back to the same E.164 number
			// when parsed with the region as its country hint.
			national := pn.Format(example, pn.NATIONAL)
			if national == "" {
				t.Fatalf("Format(example, NATIONAL) for %s is empty", country.ISO)
			}
			nationalInfo, err := phone.Lookup(national, phone.WithCountry(country.ISO))
			if err != nil {
				t.Fatalf("Lookup(%q, WithCountry(%q)) [%s] error = %v, want nil",
					national, country.ISO, country.ISO, err)
			}
			if nationalInfo.E164 != e164 {
				t.Errorf("Lookup(national=%q, WithCountry(%q)).E164 = %q, want %q",
					national, country.ISO, nationalInfo.E164, e164)
			}

			tested++
		})
	}

	if tested+skipped != len(countries) {
		t.Errorf("tested %d + skipped %d regions, want %d", tested, skipped, len(countries))
	}
	if tested < 200 {
		t.Errorf("only %d regions had example numbers, want >= 200", tested)
	}
}

// TestSharedCallingCodeRegions verifies that regions sharing a country calling
// code are distinguished by their numbering plan prefixes, using each
// region's own libphonenumber example number.
func TestSharedCallingCodeRegions(t *testing.T) {
	tests := []struct {
		iso          string
		wantISO      string
		wantPrefix   string
		wantAreaCode string
	}{
		// +1 NANP members.
		{"CA", "CA", "506", "506"},
		{"BS", "BS", "242", "242"},
		{"PR", "PR", "787", "787"},
		{"DO", "DO", "809", "809"},
		// +44 Crown Dependencies.
		{"GG", "GG", "1481", "01481"},
		{"JE", "JE", "1534", "01534"},
		{"IM", "IM", "1624", "01624"},
		// +61 Australia and its external territories.
		{"AU", "AU", "2", "02"},
		{"CC", "CC", "8", "08"},
		{"CX", "CX", "8", "08"},
		// +7 Russia and Kazakhstan.
		{"RU", "RU", "301", "301"},
		{"KZ", "KZ", "71234", "71234"},
		// +39 Italy and the Vatican.
		{"IT", "IT", "02", "02"},
		{"VA", "VA", "06", "06"},
		// +64 New Zealand.
		{"NZ", "NZ", "3", "03"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.iso, func(t *testing.T) {
			example := pn.GetExampleNumber(tt.iso)
			if example == nil {
				t.Fatalf("libphonenumber has no example number for %s", tt.iso)
			}
			e164 := pn.Format(example, pn.E164)

			info, err := phone.Lookup(e164)
			if err != nil {
				t.Fatalf("Lookup(%q) [%s] error = %v, want nil", e164, tt.iso, err)
			}
			if !info.Valid {
				t.Errorf("Lookup(%q) [%s].Valid = false, want true", e164, tt.iso)
			}
			if info.ISO != tt.wantISO {
				t.Errorf("Lookup(%q) [%s].ISO = %q, want %q", e164, tt.iso, info.ISO, tt.wantISO)
			}
			if info.Prefix != tt.wantPrefix {
				t.Errorf("Lookup(%q) [%s].Prefix = %q, want %q", e164, tt.iso, info.Prefix, tt.wantPrefix)
			}
			if info.AreaCode != tt.wantAreaCode {
				t.Errorf("Lookup(%q) [%s].AreaCode = %q, want %q", e164, tt.iso, info.AreaCode, tt.wantAreaCode)
			}

			// The calling code must be the shared one from the catalog.
			country, ok := phone.CountryByISO(tt.iso)
			if !ok {
				t.Fatalf("CountryByISO(%q) not found", tt.iso)
			}
			if info.CountryCode != country.CallingCode {
				t.Errorf("Lookup(%q) [%s].CountryCode = %d, want %d",
					e164, tt.iso, info.CountryCode, country.CallingCode)
			}

			// The national format with the region hint must give the same E.164.
			national := pn.Format(example, pn.NATIONAL)
			nationalInfo, err := phone.Lookup(national, phone.WithCountry(tt.iso))
			if err != nil {
				t.Fatalf("Lookup(%q, WithCountry(%q)) error = %v, want nil", national, tt.iso, err)
			}
			if nationalInfo.E164 != e164 {
				t.Errorf("Lookup(national=%q, WithCountry(%q)).E164 = %q, want %q",
					national, tt.iso, nationalInfo.E164, e164)
			}
		})
	}
}
