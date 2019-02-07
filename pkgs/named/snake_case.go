package named

import (
	"strings"
	"unicode/utf8"
)

// ToSnakeCase converts string s to snake_case format.
func ToSnakeCase(s string) string {
	var names []string
	for i, name := range parse(s) {
		name = strings.TrimSpace(name)

		// ignore empty
		if len(name) == 0 {
			continue
		}

		// ignore underline
		if name == "_" {
			if i == 0 {
				names = append(names, "")
			}

			continue
		}

		names = append(names, name)
	}

	for i, name := range names {
		if !utf8.ValidString(name) {
			names[i] = name
		} else {
			names[i] = strings.ToLower(name)
		}
	}

	return strings.Join(names, "_")
}
