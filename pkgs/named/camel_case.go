package named

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// ToCamelCase converts string s to CamelCase format.
// For more info please check: http://en.wikipedia.org/wiki/CamelCase
func ToCamelCase(s string) string {
	var names []string
	for _, name := range parse(s) {
		name = strings.TrimSpace(name)

		// ignore empty
		if len(name) == 0 {
			continue
		}

		// ignore _
		if name == "_" {
			continue
		}

		names = append(names, name)
	}

	for i, name := range names {
		if !utf8.ValidString(name) {
			names[i] = name
		} else {
			r := rune(name[0])
			if unicode.IsUpper(r) {
				continue
			}

			names[i] = string(unicode.ToUpper(r)) + name[1:]
		}
	}

	return strings.Join(names, "")
}
