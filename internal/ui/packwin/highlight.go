package packwin

import (
	"github.com/PalisadeMC/Packwiz-Studio/internal/tomlhl"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/widgets"
)

// tomlSpans colours one line of TOML for the editor.
func tomlSpans(line string) []widgets.Span {
	toks := tomlhl.Line(line)

	out := make([]widgets.Span, 0, len(toks))
	for _, tok := range toks {
		out = append(out, span(tok))
	}
	return out
}

// span maps a token onto its colour and weight.
func span(tok tomlhl.Token) widgets.Span {
	s := widgets.Span{Text: tok.Text, Color: tokens.SyntaxText}

	switch tok.Kind {
	case tomlhl.KindComment:
		s.Color = tokens.SyntaxComment
		s.Italic = true
	case tomlhl.KindTable:
		s.Color = tokens.SyntaxTable
		s.Bold = true
	case tomlhl.KindKey:
		s.Color = tokens.SyntaxKey
	case tomlhl.KindString:
		s.Color = tokens.SyntaxString
	case tomlhl.KindNumber:
		s.Color = tokens.SyntaxNumber
	case tomlhl.KindBool:
		s.Color = tokens.SyntaxBool
	case tomlhl.KindPunct:
		s.Color = tokens.SyntaxPunct
	}
	return s
}
