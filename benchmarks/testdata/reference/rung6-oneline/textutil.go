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

// TitleInitials returns the uppercased first letter of each word: "hello brave world" -> "HBW".
func TitleInitials(title string) string {
	initials := ""
	for _, word := range strings.Fields(title) {
		initials += strings.ToUpper(word[:1])
	}
	return initials
}
