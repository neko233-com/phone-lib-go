package phone_test

import (
	"fmt"

	phone "github.com/neko233-com/phone-lib-go"
)

// ExampleLookup shows a lookup for a mainland China mobile number. National
// format needs a country hint.
func ExampleLookup() {
	info, err := phone.Lookup("13128500205", phone.WithCountry("CN"))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(info.E164)
	fmt.Println(info.National)
	fmt.Println(info.Country, info.Location, info.Carrier)
	fmt.Println(info.Prefix, info.AreaCode, info.ZipCode)
	fmt.Println(info.Type)
	// Output:
	// +8613128500205
	// 131 2850 0205
	// 中国 广东珠海 中国联通
	// 131 0756 519000
	// mobile
}

// ExampleParse shows Parse as a shortcut for Lookup with a country hint.
func ExampleParse() {
	info, err := phone.Parse("2025550123", "US")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(info.E164)
	fmt.Println(info.National)
	fmt.Println(info.Country, info.LocationEn)
	fmt.Println(info.Type, info.AreaCode, info.Timezones)
	// Output:
	// +12025550123
	// (202) 555-0123
	// 美国 Washington D.C.
	// fixed_line_or_mobile 202 [America/New_York]
}

// ExampleLookup_international shows that "+" and "00" numbers need no hint.
func ExampleLookup_international() {
	for _, raw := range []string{"+8613128500205", "008613128500205"} {
		info, err := phone.Lookup(raw)
		if err != nil {
			fmt.Println("error:", err)
			return
		}
		fmt.Println(info.E164, info.ISO)
	}
	// Output:
	// +8613128500205 CN
	// +8613128500205 CN
}
