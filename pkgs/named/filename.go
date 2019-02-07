package named

import (
	"path"
	"strings"
)

// Basename returns the last path element of a slash-delimited name, with the last
// dotted suffix removed.
func Basename(name string) string {
	// First, find the filename and ext in the path
	_, filename := path.Split(name)
	ext := path.Ext(name)

	// Now drop the suffix
	return strings.TrimSuffix(filename, ext)
}

// ToFilename returns snake_case.ext of string s as filename
func ToFilename(s, ext string) string {
	return ToSnakeCase(s) + ext
}
