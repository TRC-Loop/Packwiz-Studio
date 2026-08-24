package packwin

import (
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"github.com/PalisadeMC/Packwiz-Studio/internal/tomlhl"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
)

// highlight turns a file into styled grid rows.
func highlight(content string) []widget.TextGridRow {
	lines := strings.Split(content, "\n")

	rows := make([]widget.TextGridRow, 0, len(lines))
	for _, line := range lines {
		var cells []widget.TextGridCell

		for _, tok := range tomlhl.Line(line) {
			style := tokenStyle(tok.Kind)
			for _, r := range tok.Text {
				cells = append(cells, widget.TextGridCell{Rune: r, Style: style})
			}
		}
		rows = append(rows, widget.TextGridRow{Cells: cells})
	}
	return rows
}

// tokenStyle maps a token kind onto a shade and weight.
//
// Syntax highlighting stays grayscale: the status colours mean state, and
// spending them on syntax would dilute that. Four shades plus bold and
// italic separate the kinds well enough for a file this size.
func tokenStyle(kind tomlhl.Kind) widget.TextGridStyle {
	switch kind {
	case tomlhl.KindComment:
		return &widget.CustomTextGridStyle{
			FGColor:   tokens.ColorDim,
			TextStyle: fyne.TextStyle{Italic: true, Monospace: true},
		}
	case tomlhl.KindTable:
		return &widget.CustomTextGridStyle{
			FGColor:   tokens.ColorStrong,
			TextStyle: fyne.TextStyle{Bold: true, Monospace: true},
		}
	case tomlhl.KindKey:
		return mono(tokens.ColorText)
	case tomlhl.KindString:
		return mono(tokens.ColorMuted)
	case tomlhl.KindNumber, tomlhl.KindBool:
		return &widget.CustomTextGridStyle{
			FGColor:   tokens.ColorMuted,
			TextStyle: fyne.TextStyle{Bold: true, Monospace: true},
		}
	case tomlhl.KindPunct:
		return mono(tokens.ColorDim)
	default:
		return mono(tokens.ColorText)
	}
}

// mono is a plain monospace style in one colour.
func mono(c color.Color) widget.TextGridStyle {
	return &widget.CustomTextGridStyle{
		FGColor:   c,
		TextStyle: fyne.TextStyle{Monospace: true},
	}
}
