package fse

import (
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

func isDigit(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func SortNatural(a, b string) bool {
	if len(a) == 0 {
		return true
	}
	if len(b) == 0 {
		return false
	}

	aFirst := rune(a[0])
	bFirst := rune(b[0])

	// First case if both aren't digits
	if !(isDigit(string(aFirst)) && isDigit(string(bFirst))) {
		// Only sorts them if one is lowercase and one is uppercase
		// and they're alphanumeric, i.e. not '_' or '.' etc.
		if unicode.IsLetter(aFirst) && unicode.IsLetter(bFirst) {
			aLowercase := string(aFirst) == strings.ToLower(string(aFirst))
			bUppercase := string(bFirst) == strings.ToUpper(string(bFirst))

			if aLowercase && bUppercase {
				return true
			} else if !aLowercase && !bUppercase {
				return false
			}
		}
	} else {
		// Second case if both are digits
		aBase := filepath.Base(a)
		bBase := filepath.Base(b)
		if isDigit(aBase) && isDigit(bBase) {
			aNum, err := strconv.Atoi(aBase)
			if err != nil {
				panic(err)
			}
			bNum, err := strconv.Atoi(bBase)
			if err != nil {
				panic(err)
			}
			return aNum < bNum
		}
	}

	// Third case is an underscore and a letter/digit
	if (aFirst == '_' || bFirst == '_') && (aFirst != bFirst) {
		return aFirst == '_' && bFirst != '_'
	}

	// Fourth case is if they're all the same up to the
	// base then we sort them based on the base value
	if filepath.Dir(a) == filepath.Dir(b) {
		aBase := filepath.Base(a)
		bBase := filepath.Base(b)
		return aBase < bBase
	}

	// Fifth case is general
	return a < b
}
