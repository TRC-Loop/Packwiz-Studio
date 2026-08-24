// Package tomlhl splits a line of TOML into coloured tokens.
//
// This is deliberately a lexer and not a parser. The editor is for the
// rare manual fix, so highlighting only has to be right line by line and
// must never fail on input that is mid-edit and temporarily invalid.
package tomlhl

import "strings"

// Kind is what a token is, which decides how it is drawn.
type Kind int

// Token kinds.
const (
	// KindText is anything unclassified, including whitespace.
	KindText Kind = iota
	// KindComment is a hash comment through to the end of the line.
	KindComment
	// KindTable is a table header such as [update.modrinth].
	KindTable
	// KindKey is a bare key on the left of an assignment.
	KindKey
	// KindString is a quoted string, single or double.
	KindString
	// KindNumber is a numeric or date literal.
	KindNumber
	// KindBool is true or false.
	KindBool
	// KindPunct is structural: equals, brackets, braces, commas.
	KindPunct
)

// Token is a run of text of one kind.
type Token struct {
	Text string
	Kind Kind
}

// Line splits one line of TOML. The result always joins back into the
// input exactly, so the editor can rely on it for display.
func Line(line string) []Token {
	trimmed := strings.TrimLeft(line, " \t")
	indent := line[:len(line)-len(trimmed)]

	switch {
	case trimmed == "":
		return []Token{{Text: line, Kind: KindText}}
	case strings.HasPrefix(trimmed, "#"):
		return emit(indent, Token{Text: trimmed, Kind: KindComment})
	case strings.HasPrefix(trimmed, "["):
		return tableHeader(indent, trimmed)
	default:
		return assignment(indent, trimmed)
	}
}

// tableHeader handles a [table] or [[array]] line, keeping any trailing
// comment separate.
func tableHeader(indent, body string) []Token {
	head, comment := splitComment(body)

	out := emit(indent, Token{Text: head, Kind: KindTable})
	if comment != "" {
		out = append(out, Token{Text: comment, Kind: KindComment})
	}
	return out
}

// assignment handles a key and value line. A line with no equals sign is
// returned as plain text rather than guessed at: it is most likely a
// partly typed line.
func assignment(indent, body string) []Token {
	key, value, found := strings.Cut(body, "=")
	if !found {
		return emit(indent, Token{Text: body, Kind: KindText})
	}

	out := emit(indent)
	out = append(out, keyTokens(key)...)
	out = append(out, Token{Text: "=", Kind: KindPunct})
	out = append(out, valueTokens(value)...)
	return out
}

// keyTokens splits a key from the trailing space before the equals sign,
// so the space is not coloured as part of the key.
func keyTokens(key string) []Token {
	name := strings.TrimRight(key, " \t")
	space := key[len(name):]

	out := []Token{{Text: name, Kind: KindKey}}
	if space != "" {
		out = append(out, Token{Text: space, Kind: KindText})
	}
	return out
}

// emit builds a token list that starts with the line's indentation.
func emit(indent string, rest ...Token) []Token {
	out := make([]Token, 0, len(rest)+1)
	if indent != "" {
		out = append(out, Token{Text: indent, Kind: KindText})
	}
	return append(out, rest...)
}

// splitComment separates a trailing comment from the rest of a line,
// ignoring hashes that sit inside a quoted string.
func splitComment(s string) (body, comment string) {
	var quote rune

	for i, r := range s {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
			}
		case r == '"' || r == '\'':
			quote = r
		case r == '#':
			return s[:i], s[i:]
		}
	}
	return s, ""
}
