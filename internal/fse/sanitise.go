package fse

import "strings"

var illegalChars = []rune{'<', '>', ':', '"', '\\', '/', '|', '?', '*'}

// Sanitise cleans up an input filename to ensure it's
// save for saving to the filesystem
func Sanitise(input string) string {
	var sb strings.Builder
	for _, r := range input {
		for _, c := range illegalChars {
			if r == c {
				sb.WriteRune('_')
				continue
			}
		}
		sb.WriteRune(r)
	}

	return strings.TrimRight(sb.String(), ".")
}
