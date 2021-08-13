package fse

import (
	"regexp"
	"strings"
)

var sanitise *regexp.Regexp

func init() {
	sanitise = regexp.MustCompile("[^\\w\\-. ]+")
}

// Sanitise cleans up an input filename to ensure it's
// save for saving to the filesystem
func Sanitise(input string) string {
	var sb strings.Builder
	for _, r := range input {
		sb.WriteString(sanitise.ReplaceAllString(string(r), "_"))
	}
	return sb.String()
}
