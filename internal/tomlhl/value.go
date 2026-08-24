package tomlhl

import "strings"

// valueTokens splits the right hand side of an assignment.
func valueTokens(value string) []Token {
	body, comment := splitComment(value)

	out := scanValue(body)
	if comment != "" {
		out = append(out, Token{Text: comment, Kind: KindComment})
	}
	return out
}

// scanValue walks a value, picking out strings, numbers, booleans and
// structure. Anything it does not recognise stays plain text, so a
// half-typed value is displayed rather than mangled.
func scanValue(s string) []Token {
	var out []Token
	i := 0

	for i < len(s) {
		c := s[i]

		switch {
		case c == ' ' || c == '\t':
			j := i
			for j < len(s) && (s[j] == ' ' || s[j] == '\t') {
				j++
			}
			out = append(out, Token{Text: s[i:j], Kind: KindText})
			i = j

		case c == '"' || c == '\'':
			end := closingQuote(s, i)
			out = append(out, Token{Text: s[i:end], Kind: KindString})
			i = end

		case strings.ContainsRune("[]{},", rune(c)):
			out = append(out, Token{Text: string(c), Kind: KindPunct})
			i++

		default:
			j := i
			for j < len(s) && !isValueBreak(s[j]) {
				j++
			}
			word := s[i:j]
			out = append(out, Token{Text: word, Kind: wordKind(word)})
			i = j
		}
	}
	return out
}

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

// isValueBreak reports a byte that ends a bare word.
func isValueBreak(c byte) bool {
	return c == ' ' || c == '\t' || c == ',' ||
		c == '[' || c == ']' || c == '{' || c == '}' ||
		c == '"' || c == '\''
}

// wordKind classifies a bare word.
func wordKind(word string) Kind {
	switch word {
	case "true", "false":
		return KindBool
	}
	if looksNumeric(word) {
		return KindNumber
	}
	return KindText
}

// looksNumeric reports a word that reads as a number or a date. TOML
// dates are made of digits, dashes, colons and letters like T and Z, and
// treating them as numeric is close enough for highlighting.
func looksNumeric(word string) bool {
	if word == "" {
		return false
	}

	hasDigit := false
	for i := range len(word) {
		c := word[i]
		switch {
		case c >= '0' && c <= '9':
			hasDigit = true
		case c == '-' || c == '+' || c == '.' || c == ':' || c == '_':
		case c == 'e' || c == 'E' || c == 'T' || c == 'Z' || c == 'x' || c == 'o' || c == 'b':
		case c >= 'a' && c <= 'f', c >= 'A' && c <= 'F':
		default:
			return false
		}
	}
	return hasDigit
}
