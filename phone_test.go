package phone_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"

	phone "github.com/neko233-com/phone-lib-go"
)

// fieldCheck is one got/want assertion on an Info field.
type fieldCheck struct {
	name string
	got  any
	want any
}

// checkFields reports every mismatched field in checks.
func checkFields(t *testing.T, checks []fieldCheck) {
	t.Helper()
	for _, c := range checks {
		if !reflect.DeepEqual(c.got, c.want) {
			t.Errorf("%s = %#v, want %#v", c.name, c.got, c.want)
		}
	}
}

// mustLookup fails the test when Lookup returns an error.
func mustLookup(t *testing.T, raw string, opts ...phone.Option) *phone.Info {
	t.Helper()
	info, err := phone.Lookup(raw, opts...)
	if err != nil {
		t.Fatalf("Lookup(%q) error = %v, want nil", raw, err)
	}
	if info == nil {
		t.Fatalf("Lookup(%q) = nil, want *Info", raw)
	}
	return info
}

// withoutRaw copies info with Raw cleared so results from different raw inputs
// can be compared.
func withoutRaw(info *phone.Info) phone.Info {
	if info == nil {
		return phone.Info{}
	}
	c := *info
	c.Raw = ""
	return c
}

func TestCNMobile(t *testing.T) {
	info := mustLookup(t, "13128500205", phone.WithCountry("CN"))
	checkFields(t, []fieldCheck{
		{"Raw", info.Raw, "13128500205"},
		{"E164", info.E164, "+8613128500205"},
		{"International", info.International, "+86 131 2850 0205"},
		{"National", info.National, "131 2850 0205"},
		{"RFC3966", info.RFC3966, "tel:+86-131-2850-0205"},
		{"NationalNumber", info.NationalNumber, "13128500205"},
		{"CountryCode", info.CountryCode, 86},
		{"ISO", info.ISO, "CN"},
		{"Country", info.Country, "中国"},
		{"CountryEn", info.CountryEn, "China"},
		{"Flag", info.Flag, "🇨🇳"},
		{"Valid", info.Valid, true},
		{"Possible", info.Possible, true},
		{"Type", info.Type, phone.TypeMobile},
		{"Carrier", info.Carrier, "中国联通"},
		{"CarrierEn", info.CarrierEn, "China Unicom"},
		{"Location", info.Location, "广东珠海"},
		{"LocationEn", info.LocationEn, "Zhuhai, Guangdong"},
		{"Province", info.Province, "广东"},
		{"City", info.City, "珠海"},
		{"Prefix", info.Prefix, "131"},
		{"AreaCode", info.AreaCode, "0756"},
		{"ZipCode", info.ZipCode, "519000"},
		{"Timezones", info.Timezones, []string{"Asia/Shanghai"}},
	})
	if !info.IsMobile() {
		t.Error("IsMobile() = false, want true for a mainland China mobile number")
	}
	wantString := "+8613128500205 中国 广东珠海 中国联通 mobile"
	if got := info.String(); got != wantString {
		t.Errorf("String() = %q, want %q", got, wantString)
	}
}

func TestCNFixedLine(t *testing.T) {
	t.Run("0756", func(t *testing.T) {
		info := mustLookup(t, "07561234567", phone.WithCountry("CN"))
		checkFields(t, []fieldCheck{
			{"E164", info.E164, "+867561234567"},
			{"National", info.National, "0756 123 4567"},
			{"ISO", info.ISO, "CN"},
			{"Country", info.Country, "中国"},
			{"Valid", info.Valid, true},
			{"Type", info.Type, phone.TypeFixedLine},
			{"Location", info.Location, "广东珠海"},
			{"LocationEn", info.LocationEn, "Zhuhai, Guangdong"},
			{"Province", info.Province, "广东"},
			{"City", info.City, "珠海"},
			{"Prefix", info.Prefix, "0756"},
			{"AreaCode", info.AreaCode, "0756"},
			{"Timezones", info.Timezones, []string{"Asia/Shanghai"}},
		})
		if info.IsMobile() {
			t.Error("IsMobile() = true, want false for a fixed-line number")
		}
	})

	t.Run("010", func(t *testing.T) {
		info := mustLookup(t, "01012345678", phone.WithCountry("CN"))
		checkFields(t, []fieldCheck{
			{"E164", info.E164, "+861012345678"},
			{"ISO", info.ISO, "CN"},
			{"Valid", info.Valid, true},
			{"Type", info.Type, phone.TypeFixedLine},
			{"Location", info.Location, "北京"},
			{"LocationEn", info.LocationEn, "Beijing"},
			{"AreaCode", info.AreaCode, "010"},
			{"Prefix", info.Prefix, "010"},
			{"Timezones", info.Timezones, []string{"Asia/Shanghai"}},
		})
	})
}

func TestUSNumber(t *testing.T) {
	info := mustLookup(t, "2025550123", phone.WithCountry("US"))
	checkFields(t, []fieldCheck{
		{"E164", info.E164, "+12025550123"},
		{"International", info.International, "+1 202-555-0123"},
		{"National", info.National, "(202) 555-0123"},
		{"NationalNumber", info.NationalNumber, "2025550123"},
		{"CountryCode", info.CountryCode, 1},
		{"ISO", info.ISO, "US"},
		{"Country", info.Country, "美国"},
		{"CountryEn", info.CountryEn, "United States"},
		{"Flag", info.Flag, "🇺🇸"},
		{"Valid", info.Valid, true},
		{"Type", info.Type, phone.TypeFixedLineOrMobile},
		{"Location", info.Location, "美国"},
		{"LocationEn", info.LocationEn, "Washington D.C."},
		{"Prefix", info.Prefix, "202"},
		{"AreaCode", info.AreaCode, "202"},
		{"Timezones", info.Timezones, []string{"America/New_York"}},
	})
	if !info.IsMobile() {
		t.Error("IsMobile() = false, want true for a US fixed-line-or-mobile number")
	}
}

func TestGBNumber(t *testing.T) {
	info := mustLookup(t, "+442083661177")
	checkFields(t, []fieldCheck{
		{"E164", info.E164, "+442083661177"},
		{"National", info.National, "020 8366 1177"},
		{"ISO", info.ISO, "GB"},
		{"Country", info.Country, "英国"},
		{"CountryEn", info.CountryEn, "United Kingdom"},
		{"Valid", info.Valid, true},
		{"Type", info.Type, phone.TypeFixedLine},
		{"Prefix", info.Prefix, "20"},
		{"AreaCode", info.AreaCode, "020"},
		{"Location", info.Location, "英国"},
		{"LocationEn", info.LocationEn, "London"},
		{"Timezones", info.Timezones, []string{"Europe/London"}},
	})
}

func TestJPNumber(t *testing.T) {
	info := mustLookup(t, "+81312345678")
	checkFields(t, []fieldCheck{
		{"E164", info.E164, "+81312345678"},
		{"ISO", info.ISO, "JP"},
		{"Country", info.Country, "日本"},
		{"Valid", info.Valid, true},
		{"Type", info.Type, phone.TypeFixedLine},
		{"Prefix", info.Prefix, "3"},
		{"AreaCode", info.AreaCode, "03"},
		{"Location", info.Location, "日本"},
		{"LocationEn", info.LocationEn, "Tokyo"},
		{"Timezones", info.Timezones, []string{"Asia/Tokyo"}},
	})
}

func TestDENumber(t *testing.T) {
	info := mustLookup(t, "+4915112345678")
	checkFields(t, []fieldCheck{
		{"E164", info.E164, "+4915112345678"},
		{"ISO", info.ISO, "DE"},
		{"Country", info.Country, "德国"},
		{"Valid", info.Valid, true},
		{"Type", info.Type, phone.TypeMobile},
		{"Prefix", info.Prefix, "1511"},
		{"CarrierEn", info.CarrierEn, "T-Mobile"},
		{"Timezones", info.Timezones, []string{"Europe/Berlin"}},
	})
	if !info.IsMobile() {
		t.Error("IsMobile() = false, want true for a German mobile number")
	}
}

func TestBRNumber(t *testing.T) {
	info := mustLookup(t, "+5511987654321")
	checkFields(t, []fieldCheck{
		{"E164", info.E164, "+5511987654321"},
		{"ISO", info.ISO, "BR"},
		{"Country", info.Country, "巴西"},
		{"Valid", info.Valid, true},
		{"Type", info.Type, phone.TypeMobile},
		{"Prefix", info.Prefix, "11"},
		{"AreaCode", info.AreaCode, "11"},
		{"CarrierEn", info.CarrierEn, "TIM"},
		{"LocationEn", info.LocationEn, "São Paulo"},
		{"Timezones", info.Timezones, []string{"America/Sao_Paulo"}},
	})
}

func TestTWNumber(t *testing.T) {
	info := mustLookup(t, "+886912345678")
	checkFields(t, []fieldCheck{
		{"E164", info.E164, "+886912345678"},
		{"ISO", info.ISO, "TW"},
		{"Country", info.Country, "台湾"},
		{"CountryEn", info.CountryEn, "Taiwan"},
		{"Flag", info.Flag, "🇹🇼"},
		{"Valid", info.Valid, true},
		{"Type", info.Type, phone.TypeMobile},
		{"Prefix", info.Prefix, "912"},
		{"Location", info.Location, "台湾"},
		{"Timezones", info.Timezones, []string{"Asia/Taipei"}},
	})
}

func TestHKNumber(t *testing.T) {
	info := mustLookup(t, "+85221234567")
	checkFields(t, []fieldCheck{
		{"E164", info.E164, "+85221234567"},
		{"ISO", info.ISO, "HK"},
		{"Country", info.Country, "中国香港"},
		{"CountryEn", info.CountryEn, "Hong Kong SAR China"},
		{"Flag", info.Flag, "🇭🇰"},
		{"Valid", info.Valid, true},
		{"Type", info.Type, phone.TypeFixedLine},
		{"Location", info.Location, "中国香港"},
		{"LocationEn", info.LocationEn, "Hong Kong SAR China"},
		{"Timezones", info.Timezones, []string{"Asia/Hong_Kong"}},
	})
}

func TestInternationalAutodetect(t *testing.T) {
	national := withoutRaw(mustLookup(t, "13128500205", phone.WithCountry("CN")))

	for _, raw := range []string{"+8613128500205", "008613128500205"} {
		t.Run(raw, func(t *testing.T) {
			info := mustLookup(t, raw)
			checkFields(t, []fieldCheck{
				{"Raw", info.Raw, raw},
				{"E164", info.E164, "+8613128500205"},
				{"ISO", info.ISO, "CN"},
				{"Location", info.Location, "广东珠海"},
			})
			if got := withoutRaw(info); !reflect.DeepEqual(got, national) {
				t.Errorf("autodetected info = %#v, want %#v", got, national)
			}
		})
	}

	// An explicit country hint must not override the country of an
	// international-format number.
	auto := withoutRaw(mustLookup(t, "+8613128500205"))
	hinted := withoutRaw(mustLookup(t, "+8613128500205", phone.WithCountry("US")))
	if !reflect.DeepEqual(auto, hinted) {
		t.Errorf("hint changed international lookup: auto = %#v, hinted = %#v", auto, hinted)
	}
}

func TestErrors(t *testing.T) {
	t.Run("country required", func(t *testing.T) {
		info, err := phone.Lookup("13128500205")
		if !errors.Is(err, phone.ErrCountryRequired) {
			t.Errorf("Lookup national error = %v, want ErrCountryRequired", err)
		}
		if info != nil {
			t.Errorf("Lookup national info = %#v, want nil", info)
		}
	})

	t.Run("unknown country", func(t *testing.T) {
		_, err := phone.Lookup("13128500205", phone.WithCountry("ZZ"))
		if !errors.Is(err, phone.ErrUnknownCountry) {
			t.Errorf("Lookup with bad hint error = %v, want ErrUnknownCountry", err)
		}
		_, err = phone.Lookup("2025550123", phone.WithCountry("Atlantis"))
		if !errors.Is(err, phone.ErrUnknownCountry) {
			t.Errorf("Lookup with country name error = %v, want ErrUnknownCountry", err)
		}
	})

	t.Run("empty number", func(t *testing.T) {
		for _, raw := range []string{"", "   ", "\t\n"} {
			info, err := phone.Lookup(raw)
			if !errors.Is(err, phone.ErrEmptyNumber) {
				t.Errorf("Lookup(%q) error = %v, want ErrEmptyNumber", raw, err)
			}
			if info != nil {
				t.Errorf("Lookup(%q) info = %#v, want nil", raw, info)
			}
		}
	})

	t.Run("unparseable", func(t *testing.T) {
		info, err := phone.Lookup("not a phone number", phone.WithCountry("US"))
		if err == nil {
			t.Fatalf("Lookup(garbage) error = nil, want non-nil")
		}
		if info != nil {
			t.Errorf("Lookup(garbage) info = %#v, want nil", info)
		}
		for _, sentinel := range []error{phone.ErrCountryRequired, phone.ErrUnknownCountry, phone.ErrEmptyNumber} {
			if errors.Is(err, sentinel) {
				t.Errorf("Lookup(garbage) error = %v, must not match %v", err, sentinel)
			}
		}
	})
}

func TestInvalidParsedNumber(t *testing.T) {
	// 200-555-0123 is ten digits, so libphonenumber parses it as a US number
	// and marks it possible, but 200 is not an assigned US area code.
	info, err := phone.Lookup("2005550123", phone.WithCountry("US"))
	if err != nil {
		t.Fatalf("Lookup(invalid US number) error = %v, want nil", err)
	}
	if info == nil {
		t.Fatal("Lookup(invalid US number) = nil, want parsed Info")
	}
	checkFields(t, []fieldCheck{
		{"E164", info.E164, "+12005550123"},
		{"ISO", info.ISO, "US"},
		{"Valid", info.Valid, false},
		{"Possible", info.Possible, true},
	})

	// The same behavior through the international format.
	info, err = phone.Lookup("+12005550123")
	if err != nil {
		t.Fatalf("Lookup(+12005550123) error = %v, want nil", err)
	}
	if info.Valid {
		t.Errorf("Lookup(+12005550123).Valid = true, want false")
	}
}

func TestJSON(t *testing.T) {
	info := mustLookup(t, "13128500205", phone.WithCountry("CN"))
	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("json.Marshal(info) error = %v", err)
	}
	for _, want := range []string{
		`"type":"mobile"`,
		`"area_code":"0756"`,
		`"iso":"CN"`,
		`"e164":"+8613128500205"`,
		`"timezones":["Asia/Shanghai"]`,
	} {
		if !strings.Contains(string(data), want) {
			t.Errorf("JSON = %s\nmissing %s", data, want)
		}
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal error = %v", err)
	}
	if got, ok := m["type"].(string); !ok || got != "mobile" {
		t.Errorf(`m["type"] = %#v, want string "mobile"`, m["type"])
	}
	if got := m["valid"]; got != true {
		t.Errorf(`m["valid"] = %#v, want true`, got)
	}
	if got := m["country_code"]; got != float64(86) {
		t.Errorf(`m["country_code"] = %#v, want 86`, got)
	}
	for _, key := range []string{
		"raw", "e164", "international", "national", "rfc3966", "national_number",
		"country_code", "iso", "country", "country_en", "flag", "valid",
		"possible", "type", "carrier", "carrier_en", "location", "location_en",
		"province", "city", "prefix", "area_code", "zip_code", "timezones",
	} {
		if _, ok := m[key]; !ok {
			t.Errorf("JSON is missing key %q", key)
		}
	}
	if _, ok := m["extension"]; ok {
		t.Errorf("JSON contains extension for a number without one: %v", m["extension"])
	}

	// Extension serializes as a string when present.
	withExt := mustLookup(t, "+12025550123 ext. 42")
	if withExt.Extension != "42" {
		t.Errorf("Extension = %q, want %q", withExt.Extension, "42")
	}
	extData, err := json.Marshal(withExt)
	if err != nil {
		t.Fatalf("json.Marshal(extension number) error = %v", err)
	}
	if !strings.Contains(string(extData), `"extension":"42"`) {
		t.Errorf("JSON = %s\nmissing extension", extData)
	}
}

func TestType(t *testing.T) {
	tests := []struct {
		typ  phone.Type
		name string
		zh   string
		json string
	}{
		{phone.TypeUnknown, "unknown", "未知", `"unknown"`},
		{phone.TypeFixedLine, "fixed_line", "固定电话", `"fixed_line"`},
		{phone.TypeMobile, "mobile", "手机", `"mobile"`},
		{phone.TypeFixedLineOrMobile, "fixed_line_or_mobile", "固定电话或手机", `"fixed_line_or_mobile"`},
		{phone.TypeTollFree, "toll_free", "免费电话", `"toll_free"`},
		{phone.TypePremiumRate, "premium_rate", "付费电话", `"premium_rate"`},
		{phone.TypeSharedCost, "shared_cost", "共享成本号码", `"shared_cost"`},
		{phone.TypeVoIP, "voip", "网络电话", `"voip"`},
		{phone.TypePersonalNumber, "personal_number", "个人号码", `"personal_number"`},
		{phone.TypePager, "pager", "寻呼机", `"pager"`},
		{phone.TypeUAN, "uan", "通用接入号码", `"uan"`},
		{phone.TypeVoicemail, "voicemail", "语音信箱", `"voicemail"`},
	}
	for _, tt := range tests {
		if got := tt.typ.String(); got != tt.name {
			t.Errorf("Type(%d).String() = %q, want %q", int(tt.typ), got, tt.name)
		}
		if got := tt.typ.Chinese(); got != tt.zh {
			t.Errorf("Type(%d).Chinese() = %q, want %q", int(tt.typ), got, tt.zh)
		}
		data, err := json.Marshal(tt.typ)
		if err != nil {
			t.Fatalf("json.Marshal(Type(%d)) error = %v", int(tt.typ), err)
		}
		if string(data) != tt.json {
			t.Errorf("json.Marshal(Type(%d)) = %s, want %s", int(tt.typ), data, tt.json)
		}
	}

	// Out-of-range types fall back to "unknown".
	for _, typ := range []phone.Type{-1, phone.Type(99)} {
		if got, want := typ.String(), "unknown"; got != want {
			t.Errorf("Type(%d).String() = %q, want %q", int(typ), got, want)
		}
		if got, want := typ.Chinese(), "未知"; got != want {
			t.Errorf("Type(%d).Chinese() = %q, want %q", int(typ), got, want)
		}
	}
}

func TestMustLookupAndMustParse(t *testing.T) {
	got := phone.MustLookup("2025550123", phone.WithCountry("US"))
	want := mustLookup(t, "2025550123", phone.WithCountry("US"))
	if !reflect.DeepEqual(got, want) {
		t.Errorf("MustLookup = %#v, want %#v", got, want)
	}

	got = phone.MustParse("+8613128500205", "")
	want = mustLookup(t, "+8613128500205")
	if !reflect.DeepEqual(got, want) {
		t.Errorf("MustParse = %#v, want %#v", got, want)
	}

	assertPanics(t, func() { phone.MustLookup("13128500205") })
	assertPanics(t, func() { phone.MustParse("13128500205", "ZZ") })
}

func assertPanics(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Error("function did not panic, want panic")
		}
	}()
	f()
}

func TestConcurrentLookups(t *testing.T) {
	const (
		goroutines   = 50
		perGoroutine = 20
	)

	inputs := []struct {
		raw     string
		opts    []phone.Option
		wantISO string
	}{
		{"13128500205", []phone.Option{phone.WithCountry("CN")}, "CN"},
		{"07561234567", []phone.Option{phone.WithCountry("CN")}, "CN"},
		{"2025550123", []phone.Option{phone.WithCountry("US")}, "US"},
		{"+442083661177", nil, "GB"},
		{"+81312345678", nil, "JP"},
		{"+4915112345678", nil, "DE"},
	}

	errCh := make(chan error, goroutines)
	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for i := 0; i < perGoroutine; i++ {
				in := inputs[(worker+i)%len(inputs)]
				info, err := phone.Lookup(in.raw, in.opts...)
				if err != nil {
					errCh <- fmt.Errorf("worker %d: Lookup(%q) error = %v", worker, in.raw, err)
					return
				}
				if info.ISO != in.wantISO {
					errCh <- fmt.Errorf("worker %d: Lookup(%q).ISO = %q, want %q", worker, in.raw, info.ISO, in.wantISO)
					return
				}
				if !info.Valid || info.E164 == "" || len(info.Timezones) == 0 {
					errCh <- fmt.Errorf("worker %d: Lookup(%q) = %#v, want valid info with E164 and timezone", worker, in.raw, info)
					return
				}
			}
		}(g)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Error(err)
	}
}

func FuzzLookup(f *testing.F) {
	seeds := []string{
		"13128500205",
		"+8613128500205",
		"008613128500205",
		"+12025550123",
		"07561234567",
		"01012345678",
		"",
		"   ",
		"not a phone number",
		"+",
		"00",
		"abc123",
		"+999999999999999",
		"+1",
		"1312850020",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, raw string) {
		for _, hint := range []string{"", "CN", "US"} {
			var opts []phone.Option
			if hint != "" {
				opts = append(opts, phone.WithCountry(hint))
			}
			info, err := phone.Lookup(raw, opts...)
			if err != nil {
				if info != nil {
					t.Fatalf("Lookup(%q, hint=%q) returned info %#v with error %v", raw, hint, info, err)
				}
				continue
			}
			if info == nil {
				t.Fatalf("Lookup(%q, hint=%q) = nil, nil", raw, hint)
			}
			if !strings.HasPrefix(info.E164, "+") {
				t.Fatalf("Lookup(%q, hint=%q).E164 = %q, want leading +", raw, hint, info.E164)
			}
		}
	})
}
