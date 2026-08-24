package syntax

import "strings"

// closingQuote finds the index just past a quoted run. An unterminated
// quote runs to the end of the line, which is what a string being typed
// looks like.
func closingQuote(s string, start int) int {
	quote := s[start]

	for i := start + 1; i < len(s); i++ {
		if s[i] == '\\' {
			i++
			continue
		}
		if s[i] == quote {
			return i + 1
		}
	}
	return len(s)
}

// runOf returns the index past a run of any of the given bytes.
func runOf(s string, from int, set string) int {
	i := from
	for i < len(s) && strings.IndexByte(set, s[i]) >= 0 {
		i++
	}
	return i
}

// runOfWord returns the index past a bare word. It always advances, so a
// byte the scanner has no rule for cannot stall it.
func runOfWord(s string, from int) int {
	i := from
	for i < len(s) && !isWordBreak(s[i]) {
		i++
	}
	if i == from {
		return from + 1
	}
	return i
}

// isWordBreak reports a byte that ends a bare word.
func isWordBreak(c byte) bool {
	return c == ' ' || c == '\t' || isPunct(c) || c == '"' || c == '\'' || c == '`'
}

// isPunct reports a structural character.
func isPunct(c byte) bool {
	return strings.IndexByte("=:,[]{}()<>;+*/%&|!?", c) >= 0
}

// looksNumeric reports a word that reads as a number or a date. TOML
// dates are digits, dashes, colons and letters like T and Z, and SNBT
// numbers carry a type suffix, so treating all of it as numeric is close
// enough for colouring.
func looksNumeric(word string) bool {
	if word == "" {
		return false
	}
	if c := word[0]; !(c >= '0' && c <= '9') && c != '-' && c != '+' && c != '.' {
		return false
	}

	for i := range len(word) {
		switch c := word[i]; {
		case c >= '0' && c <= '9':
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z':
		case c == '-' || c == '+' || c == '.' || c == '_':
		default:
			return false
		}
	}
	return true
}
