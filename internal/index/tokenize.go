package index

import (
	"regexp"
	"strings"
)

var tokenRegexp = regexp.MustCompile(`[a-zA-Z0-9]+`)

// Tokenize lowercases the input and returns word tokens suitable for indexing.
func Tokenize(text string) []string {
	matches := tokenRegexp.FindAllString(text, -1)
	if len(matches) == 0 {
		return nil
	}
	tokens := make([]string, 0, len(matches))
	for _, m := range matches {
		tokens = append(tokens, strings.ToLower(m))
	}
	return tokens
}
