package named

import (
	"strings"
	"unicode"
)

// ToIdentifier makes sure s is a valid 'identifier' string. That means
// it contains only letters, numbers, and underscore.
func ToIdentifier(s string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r):
			return r
		}

		return '_'
	}, s)
}
