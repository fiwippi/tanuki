package fse

import "strings"

var illegalChars = []rune{'<', '>', ':', '"', '\\', '/', '|', '?', '*'}

// Sanitise cleans up an input filename to ensure it's
// save for saving to the filesystem
func Sanitise(input string) string {
	var sb strings.Builder
	for _, r := range input {
		invalid := false
		for _, c := range illegalChars {
			if r == c {
				invalid = true
				sb.WriteRune('_')
				break
			}
		}
		if !invalid {
			sb.WriteRune(r)
		}
	}

	return strings.TrimRight(sb.String(), ".")
}
