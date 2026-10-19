package phone_test

import (
	"regexp"
	"strings"
	"testing"

	phone "github.com/neko233-com/phone-lib-go"
)

func TestCountries(t *testing.T) {
	countries := phone.Countries()
	if len(countries) < 200 {
		t.Fatalf("len(Countries()) = %d, want >= 200", len(countries))
	}

	byISO := make(map[string]phone.Country, len(countries))
	for i, c := range countries {
		if c.ISO != strings.ToUpper(c.ISO) || len(c.ISO) != 2 {
			t.Errorf("Countries()[%d].ISO = %q, want a 2-letter upper-case code", i, c.ISO)
		}
		if i > 0 && countries[i-1].ISO >= c.ISO {
			t.Errorf("Countries() not sorted by ISO: %q before %q", countries[i-1].ISO, c.ISO)
		}
		if c.CallingCode <= 0 {
			t.Errorf("Countries()[%d] (%s).CallingCode = %d, want > 0", i, c.ISO, c.CallingCode)
		}
		if c.Name == "" {
			t.Errorf("Countries()[%d] (%s).Name is empty", i, c.ISO)
		}
		if c.NameEn == "" {
			t.Errorf("Countries()[%d] (%s).NameEn is empty", i, c.ISO)
		}
		if c.Flag == "" {
			t.Errorf("Countries()[%d] (%s).Flag is empty", i, c.ISO)
		}
		if _, dup := byISO[c.ISO]; dup {
			t.Errorf("Countries() contains duplicate ISO %q", c.ISO)
		}
		byISO[c.ISO] = c
	}

	want := map[string]int{
		"CN": 86,
		"US": 1,
		"HK": 852,
		"TW": 886,
		"GB": 44,
	}
	for iso, callingCode := range want {
		c, ok := byISO[iso]
		if !ok {
			t.Errorf("Countries() does not contain %s", iso)
			continue
		}
		if c.CallingCode != callingCode {
			t.Errorf("Countries()[%s].CallingCode = %d, want %d", iso, c.CallingCode, callingCode)
		}
	}
}

func TestCountryByISO(t *testing.T) {
	tests := []struct {
		input   string
		wantISO string
		wantOK  bool
	}{
		{"CN", "CN", true},
		{"cn", "CN", true},
		{" Cn ", "CN", true},
		{"US", "US", true},
		{"HK", "HK", true},
		{"TW", "TW", true},
		{"GB", "GB", true},
		{"ZZ", "", false},
		{"", "", false},
		{"China", "", false},
	}
	for _, tt := range tests {
		c, ok := phone.CountryByISO(tt.input)
		if ok != tt.wantOK {
			t.Errorf("CountryByISO(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
			continue
		}
		if !tt.wantOK {
			continue
		}
		if c.ISO != tt.wantISO {
			t.Errorf("CountryByISO(%q).ISO = %q, want %q", tt.input, c.ISO, tt.wantISO)
		}
	}

	cn, ok := phone.CountryByISO("CN")
	if !ok {
		t.Fatal(`CountryByISO("CN") not found`)
	}
	if cn.Name != "中国" || cn.NameEn != "China" || cn.Flag != "🇨🇳" || cn.CallingCode != 86 {
		t.Errorf(`CountryByISO("CN") = %+v`, cn)
	}
	hk, ok := phone.CountryByISO("HK")
	if !ok {
		t.Fatal(`CountryByISO("HK") not found`)
	}
	if hk.Name != "中国香港" {
		t.Errorf(`CountryByISO("HK").Name = %q, want "中国香港"`, hk.Name)
	}
	tw, ok := phone.CountryByISO("TW")
	if !ok {
		t.Fatal(`CountryByISO("TW") not found`)
	}
	if tw.Name != "台湾" {
		t.Errorf(`CountryByISO("TW").Name = %q, want "台湾"`, tw.Name)
	}
	us, ok := phone.CountryByISO("US")
	if !ok {
		t.Fatal(`CountryByISO("US") not found`)
	}
	if us.Name != "美国" {
		t.Errorf(`CountryByISO("US").Name = %q, want "美国"`, us.Name)
	}
}

func TestCountryByCallingCode(t *testing.T) {
	tests := []struct {
		code    int
		wantISO string
		wantOK  bool
	}{
		{86, "CN", true},
		{1, "US", true},
		{852, "HK", true},
		{886, "TW", true},
		{44, "GB", true},
		{999999, "", false},
		{0, "", false},
	}
	for _, tt := range tests {
		c, ok := phone.CountryByCallingCode(tt.code)
		if ok != tt.wantOK {
			t.Errorf("CountryByCallingCode(%d) ok = %v, want %v", tt.code, ok, tt.wantOK)
			continue
		}
		if !tt.wantOK {
			continue
		}
		if c.ISO != tt.wantISO {
			t.Errorf("CountryByCallingCode(%d).ISO = %q, want %q", tt.code, c.ISO, tt.wantISO)
		}
		if c.CallingCode != tt.code {
			t.Errorf("CountryByCallingCode(%d).CallingCode = %d, want %d", tt.code, c.CallingCode, tt.code)
		}
	}
}

func TestCountriesReturnsCopy(t *testing.T) {
	first := phone.Countries()
	if len(first) == 0 {
		t.Fatal("Countries() is empty")
	}
	original := first[0]
	first[0] = phone.Country{ISO: "XX", CallingCode: 999, Name: "mutated"}
	second := phone.Countries()
	if second[0] != original {
		t.Errorf("Countries() returned shared backing storage: second[0] = %+v, want %+v", second[0], original)
	}
}

func TestChinaDataVersion(t *testing.T) {
	version := phone.ChinaDataVersion()
	if !regexp.MustCompile(`^\d{4}-\d{2}$`).MatchString(version) {
		t.Errorf("ChinaDataVersion() = %q, want a YYYY-MM version such as %q", version, "2025-02")
	}
}
