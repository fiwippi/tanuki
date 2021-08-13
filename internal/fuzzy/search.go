// Package fuzzy implements fuzzy searching
package fuzzy

import (
	"strings"
)

func indexOf(search string, letter uint8, start int) int {
	for i := start; i < len(search); i++ {
		if search[i] == letter {
			return i
		}
	}
	return -1
}

// Search attempts to fuzzy find a searchTerm within the given
// piece of text. It returns true on success
func Search(text, searchTerm string) bool {
	searchTerm = strings.ToUpper(searchTerm)
	text = strings.ToUpper(text)

	j := -1
	for i := 0; i < len(searchTerm); i++ {
		l := searchTerm[i]
		if l == ' ' { // Ignore spaces
			continue
		}

		j = indexOf(text, l, j+1) // Search for character & update position
		if j == -1 {
			return false
		}
	}
	return true
}
