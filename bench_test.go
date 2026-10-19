package phone_test

import (
	"testing"

	phone "github.com/neko233-com/phone-lib-go"
)

func BenchmarkLookupCNMobile(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		info, err := phone.Lookup("13128500205", phone.WithCountry("CN"))
		if err != nil {
			b.Fatalf("Lookup error = %v", err)
		}
		if !info.Valid || info.ISO != "CN" {
			b.Fatalf("Lookup = %#v, want a valid CN number", info)
		}
	}
}

func BenchmarkLookupUS(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		info, err := phone.Lookup("2025550123", phone.WithCountry("US"))
		if err != nil {
			b.Fatalf("Lookup error = %v", err)
		}
		if !info.Valid || info.ISO != "US" {
			b.Fatalf("Lookup = %#v, want a valid US number", info)
		}
	}
}
