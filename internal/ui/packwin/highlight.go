package packwin

import (
	"github.com/PalisadeMC/Packwiz-Studio/internal/syntax"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/widgets"
)

// tokenizerFor colours whole text as the given language.
func tokenizerFor(lang syntax.Lang) widgets.Tokenizer {
	return func(text string) [][]widgets.Span {
		lines := lang.Lines(text)

		out := make([][]widgets.Span, 0, len(lines))
		for _, line := range lines {
			spans := make([]widgets.Span, 0, len(line))
			for _, tok := range line {
				spans = append(spans, span(tok))
			}
			out = append(out, spans)
		}
		return out
	}
}

// span maps a token onto its colour and weight.
//
// The chrome is grayscale, but code is content: telling a key from a
// string from a comment at a glance is the whole point, and shades of
// grey do not do it. The hues are kept clear of the status colours, so
// red still only ever means a failure.
func span(tok syntax.Token) widgets.Span {
	s := widgets.Span{Text: tok.Text, Color: tokens.SyntaxText}

	switch tok.Kind {
	case syntax.KindComment:
		s.Color = tokens.SyntaxComment
		s.Italic = true
	case syntax.KindTable:
		s.Color = tokens.SyntaxTable
		s.Bold = true
	case syntax.KindKey:
		s.Color = tokens.SyntaxKey
	case syntax.KindString:
		s.Color = tokens.SyntaxString
	case syntax.KindNumber:
		s.Color = tokens.SyntaxNumber
	case syntax.KindBool:
		s.Color = tokens.SyntaxBool
	case syntax.KindKeyword:
		s.Color = tokens.SyntaxKeyword
	case syntax.KindPunct:
		s.Color = tokens.SyntaxPunct
	}
	return s
}
