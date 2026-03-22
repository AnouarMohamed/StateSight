package fieldpath

import "strings"

// Normalize returns a predictable field path form for matching and indexing.
func Normalize(path string) string {
	return strings.TrimSpace(strings.ReplaceAll(path, " ", ""))
}
