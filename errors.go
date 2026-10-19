package phone

import "errors"

// ErrCountryRequired is returned when a number is given in national format but
// no country hint was supplied. Use WithCountry or the international format
// (leading "+" or "00") instead.
var ErrCountryRequired = errors.New("phone: a country hint is required for national-format numbers")

// ErrEmptyNumber is returned when the input is empty.
var ErrEmptyNumber = errors.New("phone: empty phone number")

// ErrUnknownCountry is returned when the country hint cannot be resolved.
var ErrUnknownCountry = errors.New("phone: unknown country")
