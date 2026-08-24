package syntax

import "strings"

// Lines splits whole text into one token list per line.
//
// The whole text at once rather than a line at a time, because a block
// comment or an unclosed triple-quoted string carries from one line to
// the next, and a lexer that saw only one line would colour the rest of
// the file wrongly the moment one was opened.
func (l Lang) Lines(text string) [][]Token {
	lines := strings.Split(text, "\n")
	out := make([][]Token, 0, len(lines))

	inBlock := false
	for _, line := range lines {
		var toks []Token
		toks, inBlock = l.line(line, inBlock)
		out = append(out, toks)
	}
	return out
}

// line scans one line, given whether a block comment is still open, and
// reports whether it still is at the end.
func (l Lang) line(line string, inBlock bool) ([]Token, bool) {
	var out []Token
	i := 0

	if inBlock {
		end, closed := l.blockEnd(line, 0)
		out = append(out, Token{Text: line[:end], Kind: KindComment})
		if !closed {
			return out, true
		}
		i = end
	}

	if l.tableHeaders && i == 0 && strings.HasPrefix(strings.TrimLeft(line, " \t"), "[") {
		return l.tableLine(line), false
	}

	for i < len(line) {
		start := i
		switch {
		case line[i] == ' ' || line[i] == '\t':
			i = runOf(line, i, " \t")
			out = append(out, Token{Text: line[start:i], Kind: KindText})

		case l.startsLineComment(line, i):
			out = append(out, Token{Text: line[i:], Kind: KindComment})
			return out, false

		case l.blockOpen != "" && strings.HasPrefix(line[i:], l.blockOpen):
			end, closed := l.blockEnd(line, i+len(l.blockOpen))
			out = append(out, Token{Text: line[i:end], Kind: KindComment})
			if !closed {
				return out, true
			}
			i = end

		case strings.IndexByte(l.quotes, line[i]) >= 0:
			i = closingQuote(line, i)
			out = append(out, Token{
				Text: line[start:i],
				Kind: l.stringOrKey(line, i),
			})

		case isPunct(line[i]):
			i++
			out = append(out, Token{Text: line[start:i], Kind: KindPunct})

		default:
			i = runOfWord(line, i)
			word := line[start:i]
			out = append(out, Token{Text: word, Kind: l.wordKind(word, line, i)})
		}
	}
	return out, false
}

// tableLine colours a section header, keeping any trailing comment.
func (l Lang) tableLine(line string) []Token {
	for _, marker := range l.lineComment {
		if at := strings.Index(line, marker); at >= 0 {
			return []Token{
				{Text: line[:at], Kind: KindTable},
				{Text: line[at:], Kind: KindComment},
			}
		}
	}
	return []Token{{Text: line, Kind: KindTable}}
}

// startsLineComment reports a comment marker at this position.
func (l Lang) startsLineComment(line string, i int) bool {
	for _, marker := range l.lineComment {
		if strings.HasPrefix(line[i:], marker) {
			return true
		}
	}
	return false
}

// blockEnd finds where an open block comment ends on this line, and
// whether it ended at all.
func (l Lang) blockEnd(line string, from int) (int, bool) {
	if l.blockClose == "" {
		return len(line), true
	}
	if at := strings.Index(line[from:], l.blockClose); at >= 0 {
		return from + at + len(l.blockClose), true
	}
	return len(line), false
}

// stringOrKey decides whether a quoted run is a value or the name of one,
// which is what separates a JSON key from a JSON string.
func (l Lang) stringOrKey(line string, after int) Kind {
	if l.followedByAssign(line, after) {
		return KindKey
	}
	return KindString
}

// wordKind classifies a bare word.
func (l Lang) wordKind(word, line string, after int) Kind {
	switch {
	case l.constants[word]:
		return KindBool
	case l.keywords[word]:
		return KindKeyword
	case l.followedByAssign(line, after):
		return KindKey
	case looksNumeric(word):
		return KindNumber
	default:
		return KindText
	}
}

// followedByAssign reports an assignment character next along the line,
// ignoring the space in between.
func (l Lang) followedByAssign(line string, from int) bool {
	if l.assign == "" {
		return false
	}
	rest := strings.TrimLeft(line[from:], " \t")
	return rest != "" && strings.IndexByte(l.assign, rest[0]) >= 0
}
