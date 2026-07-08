package notekeep

import (
	"regexp"
	"strings"
)

var slugSeparators = regexp.MustCompile(`[^a-z0-9]+`)

// Slugify returns a lowercase ascii slug: "Hello, World!" -> "hello-world".
func Slugify(text string) string {
	var ascii strings.Builder
	for _, char := range strings.ToLower(text) {
		if char < 128 {
			ascii.WriteRune(char)
		}
	}
	return strings.Trim(slugSeparators.ReplaceAllString(ascii.String(), "-"), "-")
}
