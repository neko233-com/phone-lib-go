package phone

import (
	"fmt"
	"strings"

	pn "github.com/nyaruka/phonenumbers/v2"
	"github.com/nyaruka/phonenumbers/v2/carrier"
	"github.com/nyaruka/phonenumbers/v2/geocoding"
	"github.com/nyaruka/phonenumbers/v2/timezone"

	"github.com/neko233-com/phone-lib-go/internal/cndata"
)

const unknownTimezone = "Etc/Unknown"

// Option configures Lookup.
type Option func(*config)

type config struct {
	country string
}

// WithCountry supplies a country hint for numbers written in national format.
// It accepts an ISO 3166-1 alpha-2 code ("CN", "US"), a calling code ("86",
// "+86"), a Chinese or English country name ("中国", "United States"), or a
// common alias ("香港", "USA"). If the number starts with "+" or "00" the hint
// is not needed and is ignored.
func WithCountry(country string) Option {
	return func(c *config) {
		c.country = country
	}
}

// Lookup parses raw and enriches it with country, location, carrier, prefix,
// area code, type and timezone information.
//
// Numbers in international format ("+" or "00" prefix) are parsed without a
// country hint. Numbers in national format require WithCountry:
//
//	info, err := phone.Lookup("13128500205", phone.WithCountry("CN"))
//	info, err := phone.Lookup("+8613128500205")
//
// A parsed but invalid number is returned with Info.Valid == false and a nil
// error; only unparseable input or a bad country hint returns an error.
func Lookup(raw string, opts ...Option) (*Info, error) {
	var cfg config
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return lookup(raw, cfg)
}

// Parse is a shortcut for Lookup(raw, WithCountry(country)).
func Parse(raw, country string) (*Info, error) {
	return lookup(raw, config{country: country})
}

// MustLookup is like Lookup but panics on error.
func MustLookup(raw string, opts ...Option) *Info {
	info, err := Lookup(raw, opts...)
	if err != nil {
		panic(err)
	}
	return info
}

// MustParse is like Parse but panics on error.
func MustParse(raw, country string) *Info {
	info, err := Parse(raw, country)
	if err != nil {
		panic(err)
	}
	return info
}

// ChinaDataVersion reports the version of the embedded mainland-China 号段
// database, for example "2025-02".
func ChinaDataVersion() string {
	return cndata.Version()
}

func lookup(raw string, cfg config) (*Info, error) {
	input := strings.TrimSpace(raw)
	if input == "" {
		return nil, ErrEmptyNumber
	}

	var hint *Country
	if cfg.country != "" {
		iso, ok := resolveRegion(cfg.country)
		if !ok {
			return nil, fmt.Errorf("%w: %q", ErrUnknownCountry, cfg.country)
		}
		if c, ok := CountryByISO(iso); ok {
			hint = &c
		}
	}

	var (
		num *pn.PhoneNumber
		err error
	)
	switch {
	case strings.HasPrefix(input, "+"):
		num, err = pn.Parse(input, "")
	case strings.HasPrefix(strings.ToLower(input), "tel:"):
		// RFC 3966 / tel: URI input. libphonenumber parses the URI itself.
		num, err = pn.Parse(input, "")
	case strings.HasPrefix(input, "00"):
		num, err = pn.Parse("+"+input[2:], "")
		if err != nil && hint != nil {
			if national, nationalErr := pn.Parse(input, hint.ISO); nationalErr == nil {
				num, err = national, nil
			}
		}
	default:
		if hint == nil {
			return nil, fmt.Errorf("%w: %q (use phone.WithCountry or the international format)", ErrCountryRequired, input)
		}
		num, err = pn.Parse(input, hint.ISO)
	}
	if err != nil {
		return nil, fmt.Errorf("phone: cannot parse %q: %w", input, err)
	}
	return buildInfo(input, num, hint), nil
}

func buildInfo(raw string, num *pn.PhoneNumber, hint *Country) *Info {
	info := &Info{
		Raw:            raw,
		CountryCode:    int(num.GetCountryCode()),
		NationalNumber: pn.GetNationalSignificantNumber(num),
		Extension:      num.GetExtension(),
	}

	region := pn.GetRegionCodeForNumber(num)
	if region == "" {
		region = pn.GetRegionCodeForCountryCode(int(num.GetCountryCode()))
	}
	if region == "" && hint != nil {
		region = hint.ISO
	}
	info.ISO = region
	if c, ok := CountryByISO(region); ok {
		info.Country = c.Name
		info.CountryEn = c.NameEn
		info.Flag = c.Flag
	} else if hint != nil {
		info.Country = hint.Name
		info.CountryEn = hint.NameEn
		info.Flag = hint.Flag
	}
	if info.ISO == "001" {
		// Non-geographical entities such as +800 (international freephone).
		info.Country = "国际"
		info.CountryEn = "International"
		info.Flag = ""
	}

	info.Valid = pn.IsValidNumber(num)
	info.Possible = pn.IsPossibleNumber(num)
	info.Type = typeFromPN(pn.GetNumberType(num))

	info.E164 = pn.Format(num, pn.E164)
	info.International = pn.Format(num, pn.INTERNATIONAL)
	info.National = pn.Format(num, pn.NATIONAL)
	info.RFC3966 = pn.Format(num, pn.RFC3966)

	enrich(info, num)
	return info
}

func typeFromPN(t pn.PhoneNumberType) Type {
	switch t {
	case pn.FIXED_LINE:
		return TypeFixedLine
	case pn.MOBILE:
		return TypeMobile
	case pn.FIXED_LINE_OR_MOBILE:
		return TypeFixedLineOrMobile
	case pn.TOLL_FREE:
		return TypeTollFree
	case pn.PREMIUM_RATE:
		return TypePremiumRate
	case pn.SHARED_COST:
		return TypeSharedCost
	case pn.VOIP:
		return TypeVoIP
	case pn.PERSONAL_NUMBER:
		return TypePersonalNumber
	case pn.PAGER:
		return TypePager
	case pn.UAN:
		return TypeUAN
	case pn.VOICEMAIL:
		return TypeVoicemail
	default:
		return TypeUnknown
	}
}

func enrich(info *Info, num *pn.PhoneNumber) {
	nsn := info.NationalNumber
	locationZH, _ := geocoding.GetDescriptionForNumber(num, "zh")
	locationEN, _ := geocoding.GetDescriptionForNumber(num, "en")
	carrierZH, _ := carrier.GetNameForNumber(num, "zh")
	carrierEN, _ := carrier.GetNameForNumber(num, "en")

	if info.ISO == "CN" {
		if rec, ok := cndata.Lookup(nsn); ok {
			info.Province = rec.Province
			info.City = rec.City
			info.ZipCode = normalizeChinaZip(rec.ZipCode)
			if name := cndata.CarrierName(rec.CarrierType); name != "" {
				info.Carrier = name
			}
			if rec.AreaCode != "" {
				info.AreaCode = rec.AreaCode
			}
			info.Location = joinChinaLocation(rec.Province, rec.City)
		}
		if info.Location == "" && locationZH != "" {
			info.Location = normalizeChinaLocation(locationZH)
			if info.Province == "" {
				info.Province, info.City = splitChinaLocation(info.Location)
			}
		}
		if info.Carrier == "" {
			info.Carrier = carrierZH
		}
	} else {
		info.Location = strings.ReplaceAll(locationZH, "特别行政区", "")
		info.Carrier = carrierZH
	}
	if info.Carrier == "" {
		info.Carrier = carrierEN
	}
	info.CarrierEn = carrierEN
	info.LocationEn = locationEN
	if info.Location == "" {
		info.Location = info.Country
	}
	if info.LocationEn == "" {
		info.LocationEn = info.CountryEn
	}

	enrichNumberParts(info, num)

	if zones, err := timezone.GetTimeZonesForNumber(num); err == nil {
		for _, zone := range zones {
			if zone != "" && zone != unknownTimezone {
				info.Timezones = append(info.Timezones, zone)
			}
		}
	}
}

func enrichNumberParts(info *Info, num *pn.PhoneNumber) {
	nsn := info.NationalNumber
	if nsn == "" {
		return
	}

	if ndcLen := pn.GetLengthOfNationalDestinationCode(num); ndcLen > 0 && ndcLen < len(nsn) {
		info.Prefix = nsn[:ndcLen]
	}

	if info.AreaCode == "" {
		if geoLen := pn.GetLengthOfGeographicalAreaCode(num); geoLen > 0 && geoLen <= len(nsn) {
			area := nsn[:geoLen]
			if group := firstGroupDigits(info.National); len(group) >= geoLen &&
				len(group) <= geoLen+2 && len(group) <= len(nsn)+2 {
				area = group
			}
			if info.ISO == "CN" && !strings.HasPrefix(area, "0") {
				area = "0" + area
			}
			info.AreaCode = area
		}
	}

	if info.ISO == "CN" && info.Type == TypeFixedLine && info.AreaCode != "" {
		info.Prefix = info.AreaCode
	}
}

func firstGroupDigits(formatted string) string {
	fields := strings.FieldsFunc(formatted, func(r rune) bool {
		return r == ' ' || r == '-' || r == '\t'
	})
	if len(fields) == 0 {
		return ""
	}
	var b strings.Builder
	for _, r := range fields[0] {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func joinChinaLocation(province, city string) string {
	switch {
	case province == "":
		return city
	case city == "" || city == province:
		return province
	default:
		return province + city
	}
}

// normalizeChinaZip restores leading zeros that some upstream rows lost
// (for example "63000" for 063000). Chinese postal codes are six digits.
func normalizeChinaZip(zip string) string {
	if len(zip) == 0 || len(zip) >= 6 {
		return zip
	}
	for _, r := range zip {
		if r < '0' || r > '9' {
			return zip
		}
	}
	for len(zip) < 6 {
		zip = "0" + zip
	}
	return zip
}

// normalizeChinaLocation turns libphonenumber's administratively complete
// descriptions ("广东省珠海市", "新疆维吾尔自治区乌鲁木齐市") into the short
// "广东珠海" style used by Chinese phone lookup pages.
func normalizeChinaLocation(location string) string {
	location = strings.TrimSpace(location)
	for _, suffix := range []string{
		"壮族自治区", "回族自治区", "维吾尔自治区", "特别行政区", "自治区",
	} {
		if strings.HasSuffix(location, suffix) {
			return strings.TrimSuffix(location, suffix)
		}
	}
	location = strings.ReplaceAll(location, "省", "")
	for _, suffix := range []string{"市", "地区", "自治州", "盟"} {
		if strings.HasSuffix(location, suffix) {
			return strings.TrimSuffix(location, suffix)
		}
	}
	return location
}

var chinaProvinceNames = []string{
	"内蒙古", "黑龙江",
	"北京", "天津", "河北", "山西", "辽宁", "吉林", "上海", "江苏", "浙江",
	"安徽", "福建", "江西", "山东", "河南", "湖北", "湖南", "广东", "广西",
	"海南", "重庆", "四川", "贵州", "云南", "西藏", "陕西", "甘肃", "青海",
	"宁夏", "新疆", "台湾", "香港", "澳门",
}

var chinaMunicipalities = map[string]bool{
	"北京": true, "天津": true, "上海": true, "重庆": true,
}

func splitChinaLocation(location string) (province, city string) {
	for _, name := range chinaProvinceNames {
		if strings.HasPrefix(location, name) {
			if chinaMunicipalities[name] {
				return name, name
			}
			return name, strings.TrimPrefix(location, name)
		}
	}
	return "", ""
}
