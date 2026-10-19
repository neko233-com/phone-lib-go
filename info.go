package phone

import (
	"encoding/json"
	"fmt"
)

// Type is the type of a phone number, mirroring libphonenumber's
// PhoneNumberType values.
type Type int

// Phone number types.
const (
	TypeUnknown Type = iota
	TypeFixedLine
	TypeMobile
	TypeFixedLineOrMobile
	TypeTollFree
	TypePremiumRate
	TypeSharedCost
	TypeVoIP
	TypePersonalNumber
	TypePager
	TypeUAN
	TypeVoicemail
)

var typeNames = [...]string{
	"unknown",
	"fixed_line",
	"mobile",
	"fixed_line_or_mobile",
	"toll_free",
	"premium_rate",
	"shared_cost",
	"voip",
	"personal_number",
	"pager",
	"uan",
	"voicemail",
}

var typeNamesZH = [...]string{
	"未知",
	"固定电话",
	"手机",
	"固定电话或手机",
	"免费电话",
	"付费电话",
	"共享成本号码",
	"网络电话",
	"个人号码",
	"寻呼机",
	"通用接入号码",
	"语音信箱",
}

// String returns the English name of the type, for example "mobile".
func (t Type) String() string {
	if int(t) < 0 || int(t) >= len(typeNames) {
		return typeNames[0]
	}
	return typeNames[t]
}

// Chinese returns the Chinese name of the type, for example "手机".
func (t Type) Chinese() string {
	if int(t) < 0 || int(t) >= len(typeNamesZH) {
		return typeNamesZH[0]
	}
	return typeNamesZH[t]
}

// MarshalJSON encodes a Type as its English name.
func (t Type) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// UnmarshalJSON accepts either the English name produced by MarshalJSON (for
// example "mobile") or the underlying integer value.
func (t *Type) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("phone: empty JSON value for Type")
	}
	if data[0] == '"' {
		var name string
		if err := json.Unmarshal(data, &name); err != nil {
			return err
		}
		for i, candidate := range typeNames {
			if candidate == name {
				*t = Type(i)
				return nil
			}
		}
		return fmt.Errorf("phone: unknown phone number type %q", name)
	}
	var value int
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*t = Type(value)
	return nil
}

// Info is the result of Lookup. Fields that a particular country does not
// provide are empty strings; Timezones is omitted when unknown.
//
// Carrier information is inferred from number prefixes and may be wrong for
// ported numbers (number portability), which is the same caveat libphonenumber
// documents for its carrier mapper.
type Info struct {
	// Raw is the original input.
	Raw string `json:"raw"`
	// E164 is the number in E.164 format, for example "+8613128500205".
	E164 string `json:"e164"`
	// International is the number in international format, for example
	// "+86 131 2850 0205".
	International string `json:"international"`
	// National is the number in national format, for example "131 2850 0205".
	National string `json:"national"`
	// RFC3966 is the number in RFC 3966 / tel: URI format.
	RFC3966 string `json:"rfc3966,omitempty"`

	// NationalNumber is the national significant number without the country
	// calling code or national prefix, for example "13128500205".
	NationalNumber string `json:"national_number"`
	// CountryCode is the country calling code, for example 86.
	CountryCode int `json:"country_code"`
	// ISO is the ISO 3166-1 alpha-2 region code, for example "CN".
	ISO string `json:"iso"`
	// Country is the country/region name in Chinese, for example "中国".
	Country string `json:"country"`
	// CountryEn is the country/region name in English, for example "China".
	CountryEn string `json:"country_en"`
	// Flag is the emoji flag of the region, for example "🇨🇳".
	Flag string `json:"flag,omitempty"`

	// Valid reports whether the number matches the numbering plan of its
	// region.
	Valid bool `json:"valid"`
	// Possible reports whether the number has a length that is plausible for
	// its region, without checking the full numbering plan.
	Possible bool `json:"possible"`
	// Type is the detected number type.
	Type Type `json:"type"`

	// Carrier is the operator name in Chinese when available, for example
	// "中国联通". For countries without Chinese carrier data it falls back to
	// the English carrier name.
	Carrier string `json:"carrier,omitempty"`
	// CarrierEn is the operator name in English, for example "China Unicom".
	CarrierEn string `json:"carrier_en,omitempty"`

	// Location is the 归属地 in Chinese. For mainland China it is the
	// province plus city (for example "广东珠海"); for most other countries it
	// is the country name.
	Location string `json:"location,omitempty"`
	// LocationEn is the 归属地 in English, for example "Zhuhai, Guangdong".
	LocationEn string `json:"location_en,omitempty"`
	// Province is the province for mainland China mobile numbers, for example
	// "广东".
	Province string `json:"province,omitempty"`
	// City is the city for mainland China mobile numbers, for example "珠海".
	City string `json:"city,omitempty"`

	// Prefix is the 号段, the leading digits of the national significant
	// number, for example "131". For mainland China fixed lines it is the
	// area code, for example "0756".
	Prefix string `json:"prefix,omitempty"`
	// AreaCode is the 区号. Mainland China area codes keep their leading zero
	// ("0756"); other countries use the digits of the national destination
	// code as shown by their national format ("020", "03", "202").
	AreaCode string `json:"area_code,omitempty"`
	// ZipCode is the postal code of the area, currently only available for
	// mainland China mobile numbers.
	ZipCode string `json:"zip_code,omitempty"`

	// Timezones lists the IANA time zones of the number, for example
	// ["Asia/Shanghai"].
	Timezones []string `json:"timezones,omitempty"`
	// Extension is the extension if the input contained one.
	Extension string `json:"extension,omitempty"`
}

// IsMobile reports whether the number is a mobile number (or a number that can
// be either fixed line or mobile, as in the NANP).
func (i *Info) IsMobile() bool {
	return i.Type == TypeMobile || i.Type == TypeFixedLineOrMobile
}

// String returns a compact one-line summary of the number.
func (i *Info) String() string {
	return fmt.Sprintf("%s %s %s %s %s", i.E164, i.Country, i.Location, i.Carrier, i.Type)
}
