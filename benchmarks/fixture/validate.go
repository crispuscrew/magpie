package notekeep

import "regexp"

var emailPattern = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

// ValidEmail reports whether value looks like an email address.
func ValidEmail(value string) bool {
	return emailPattern.MatchString(value)
}
