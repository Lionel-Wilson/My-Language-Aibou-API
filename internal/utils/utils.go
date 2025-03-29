package utils

import (
	"unicode"
)

// containsNumber checks if a given string contains a number.
func ContainsNumber(s string) bool {
	for _, ch := range s {
		if unicode.IsDigit(ch) {
			return true
		}
	}

	return false
}
