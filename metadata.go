package phone

import "github.com/nyaruka/phonenumbers/v2/metadata"

// MetadataVersion returns the upstream Google libphonenumber metadata version
// embedded by the parsing engine, for example "v9.0.40".
//
// The value is compiled into the library, so it can be logged or exposed by
// services that need to prove which numbering-plan data they are using.
func MetadataVersion() string {
	return metadata.Version
}
