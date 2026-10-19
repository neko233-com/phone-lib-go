// Package phone parses and enriches phone numbers from every country and
// region supported by Google's libphonenumber metadata.
//
// Given a number and an optional country hint, Lookup returns a flat Info
// value with the fields commonly used by Chinese phone-number lookup pages:
//
//	号码 (number), 国家/地区 (country), 归属地 (location), 运营商 (carrier),
//	号段 (prefix), 区号 (area code), 邮编 (zip code), 类型 (type) and 时区 (timezone).
//
// Global parsing, validation, formatting and number-type detection are powered
// by github.com/nyaruka/phonenumbers/v2 (a port of Google's libphonenumber).
// Mainland China lookup data (city, carrier, area code, zip code) comes from an
// embedded 500k-row 号段 database, so Chinese numbers resolve to a city rather
// than only to a country.
//
// Basic usage:
//
//	info, err := phone.Lookup("13128500205", phone.WithCountry("CN"))
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(info.Location, info.Carrier, info.Prefix, info.AreaCode)
//	// 广东珠海 中国联通 131 0756
//
// International-format numbers do not need a country hint:
//
//	info, err := phone.Lookup("+8613128500205")
//
// Version, MetadataVersion and ChinaDataVersion report the library and embedded
// data versions, which services can log or expose to show which snapshots they
// are using.
package phone
