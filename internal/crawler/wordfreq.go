package crawler

import (
	"strings"
	"unicode"
)

// wordFrequency counts word occurrences in text, lowercased.
func wordFrequency(text string) map[string]int {
	freq := make(map[string]int)
	words := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	for _, w := range words {
		if len(w) > 1 { // skip single-char tokens
			freq[w]++
		}
	}
	return freq
}
