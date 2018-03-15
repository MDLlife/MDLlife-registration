package validation_rules

import "regexp"

var(
	NameRegex = regexp.MustCompile("^(?:(\\pL)+(?:(?:\\pL|[-\\s])+)?)$")
	// YYYY-MM-DD // YYYY >= 1000 matches correct dates in months
	DateRegex = regexp.MustCompile("^(?:[1-9]\\d{3}-(?:(?:0[1-9]|1[0-2])-(?:0[1-9]|1\\d|2[0-9])|(?:0[13-9]|1[0-2])-(?:29|30)|(?:0[13578]|1[02])-31))$")
	// phone number, simple
	PhoneNumberRegex = regexp.MustCompile("[0-9()\\pL\\s-+#]+")

	TokenRegex = regexp.MustCompile("[0-9a-zA-Z]+")
)
