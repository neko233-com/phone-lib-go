package phone

import (
	"sort"
	"strconv"
	"strings"
	"sync"

	pn "github.com/nyaruka/phonenumbers/v2"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

// Country describes a country or region supported by the library.
type Country struct {
	// ISO is the ISO 3166-1 alpha-2 region code, for example "CN".
	ISO string `json:"iso"`
	// CallingCode is the country calling code, for example 86.
	CallingCode int `json:"calling_code"`
	// Name is the Chinese name, for example "中国".
	Name string `json:"name"`
	// NameEn is the English name, for example "China".
	NameEn string `json:"name_en"`
	// Flag is the emoji flag, for example "🇨🇳".
	Flag string `json:"flag"`
}

var (
	catalogOnce sync.Once
	countries   []Country
	byISO       map[string]Country
	byCalling   map[int]Country
	aliases     map[string]string
)

// zhNameOverrides shortens a few CLDR names that are inconvenient in a lookup
// result.
var zhNameOverrides = map[string]string{
	"HK": "中国香港",
	"MO": "中国澳门",
}

var extraAliases = map[string]string{
	"大陆":             "CN",
	"中国大陆":           "CN",
	"内地":             "CN",
	"香港":             "HK",
	"澳门":             "MO",
	"台湾":             "TW",
	"中国香港":           "HK",
	"中国澳门":           "MO",
	"usa":            "US",
	"u.s.a.":         "US",
	"uk":             "GB",
	"u.k.":           "GB",
	"britain":        "GB",
	"great britain":  "GB",
	"south korea":    "KR",
	"korea":          "KR",
	"north korea":    "KP",
	"russia":         "RU",
	"vietnam":        "VN",
	"viet nam":       "VN",
	"laos":           "LA",
	"syria":          "SY",
	"bolivia":        "BO",
	"tanzania":       "TZ",
	"venezuela":      "VE",
	"iran":           "IR",
	"moldova":        "MD",
	"macedonia":      "MK",
	"czech republic": "CZ",
	"ivory coast":    "CI",
	"cape verde":     "CV",
	"swaziland":      "SZ",
	"east timor":     "TL",
	"brunei":         "BN",
}

func loadCatalog() {
	zh := display.Regions(language.Chinese)
	en := display.Regions(language.English)

	byISO = make(map[string]Country, 256)
	byCalling = make(map[int]Country, 256)
	aliases = make(map[string]string, 1024)

	for iso := range pn.GetSupportedRegions() {
		callingCode := pn.GetCountryCodeForRegion(iso)
		if callingCode == 0 {
			continue
		}
		nameZH := zh.Name(mustRegion(iso))
		if override, ok := zhNameOverrides[iso]; ok {
			nameZH = override
		}
		nameEN := en.Name(mustRegion(iso))
		if nameZH == "" {
			nameZH = nameEN
		}
		if nameEN == "" {
			nameEN = nameZH
		}
		c := Country{
			ISO:         iso,
			CallingCode: callingCode,
			Name:        nameZH,
			NameEn:      nameEN,
			Flag:        flagOf(iso),
		}
		byISO[iso] = c
	}

	for iso, c := range byISO {
		aliases[strings.ToUpper(iso)] = iso
		if c.Name != "" {
			aliases[strings.ToLower(c.Name)] = iso
		}
		if c.NameEn != "" {
			aliases[strings.ToLower(c.NameEn)] = iso
		}
	}

	// The country returned for a calling code should be the numbering plan's
	// main country (for example +1 -> US, +7 -> RU, +44 -> GB).
	for code := range pn.GetSupportedCallingCodes() {
		main := pn.GetRegionCodeForCountryCode(code)
		if c, ok := byISO[main]; ok {
			byCalling[code] = c
		}
	}

	for alias, iso := range extraAliases {
		if _, exists := aliases[alias]; !exists {
			if _, ok := byISO[iso]; ok {
				aliases[alias] = iso
			}
		}
	}

	countries = make([]Country, 0, len(byISO))
	for _, c := range byISO {
		countries = append(countries, c)
	}
	sort.Slice(countries, func(i, j int) bool { return countries[i].ISO < countries[j].ISO })
}

func mustRegion(iso string) language.Region {
	r, err := language.ParseRegion(iso)
	if err != nil {
		return language.Region{}
	}
	return r
}

func flagOf(iso string) string {
	if len(iso) != 2 {
		return ""
	}
	a := iso[0]
	b := iso[1]
	if a < 'A' || a > 'Z' || b < 'A' || b > 'Z' {
		return ""
	}
	const base = 0x1F1E6 // REGIONAL INDICATOR SYMBOL LETTER A
	return string([]rune{rune(base + int(a-'A')), rune(base + int(b-'A'))})
}

// Countries returns all supported countries and regions, sorted by ISO code.
func Countries() []Country {
	catalogOnce.Do(loadCatalog)
	out := make([]Country, len(countries))
	copy(out, countries)
	return out
}

// CountryByISO returns the country with the given ISO 3166-1 alpha-2 code.
func CountryByISO(iso string) (Country, bool) {
	catalogOnce.Do(loadCatalog)
	c, ok := byISO[strings.ToUpper(strings.TrimSpace(iso))]
	return c, ok
}

// CountryByCallingCode returns the main country for a calling code, for
// example 86 -> China and 1 -> United States.
func CountryByCallingCode(code int) (Country, bool) {
	catalogOnce.Do(loadCatalog)
	c, ok := byCalling[code]
	return c, ok
}

// resolveRegion resolves a country hint from an ISO code, a calling code
// (with or without "+"), a Chinese/English country name, or a common alias.
func resolveRegion(input string) (string, bool) {
	catalogOnce.Do(loadCatalog)
	s := strings.TrimSpace(input)
	if s == "" {
		return "", false
	}
	upper := strings.ToUpper(s)
	if c, ok := byISO[upper]; ok {
		return c.ISO, true
	}
	if iso, ok := aliases[strings.ToLower(s)]; ok {
		return iso, true
	}
	digits := s
	if strings.HasPrefix(digits, "+") {
		digits = digits[1:]
	}
	digits = strings.TrimPrefix(digits, "00")
	if digits != "" {
		if code, err := strconv.Atoi(digits); err == nil {
			if c, ok := byCalling[code]; ok {
				return c.ISO, true
			}
		}
	}
	return "", false
}
