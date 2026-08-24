package widgets

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	fynetheme "fyne.io/fyne/v2/theme"

	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
)

// tabColumns is how far a tab advances, matching what Fyne's own text
// painter does. The editor's columns have to agree with the entry's
// drawing, or the layer would sit part of a character out on any line
// with a tab in it.
const tabColumns = 4

// gutterPad is the gap between the line numbers and the text.
const gutterPad float32 = 8

// repaint rebuilds the layer: the line numbers, then the coloured text.
//
// One text object per token rather than per character: a file of a few
// hundred lines is then a few thousand objects at worst, where a grid of
// cells would be tens of thousands.
func (c *Code) repaint() {
	if c.tokenize == nil {
		return
	}

	lines := c.tokenize(c.entry.Text)
	c.rows = len(lines)
	c.gutter.width = c.gutterWidth()

	cell := codeCell()
	origin := codeOrigin(c.gutter.width)

	objects := make([]fyne.CanvasObject, 0, len(lines)*4)
	c.numbers = make([]*canvas.Text, 0, len(lines))
	widest := 0

	for row, spans := range lines {
		objects = append(objects, c.number(row, cell))

		col := 0
		for _, span := range spans {
			objects, col = c.drawSpan(objects, span, origin, cell, row, col)
		}
		widest = max(widest, col)
	}

	c.place.min = fyne.NewSize(
		origin.X+codeOriginPad()+cell.Width*float32(widest),
		codeOriginPad()*2+cell.Height*float32(len(lines)),
	)

	c.layer.Objects = objects
	c.layer.Refresh()
	c.markLine()
}

// number is one line number, right aligned against the text.
func (c *Code) number(row int, cell fyne.Size) fyne.CanvasObject {
	label := strconv.Itoa(row + 1)

	t := canvas.NewText(label, tokens.SyntaxGutter)
	t.TextSize = fynetheme.TextSize()
	t.TextStyle = fyne.TextStyle{Monospace: true}

	width := cell.Width * float32(len(label))
	t.Move(fyne.NewPos(c.gutter.width-gutterPad-width, codeOriginPad()+cell.Height*float32(row)))
	t.Resize(fyne.NewSize(width, cell.Height))

	c.numbers = append(c.numbers, t)
	return t
}

// repaintGutter brightens the caret's line number and dims the rest, so
// the gutter says where you are as well as how far down you have got.
func (c *Code) repaintGutter() {
	for i, t := range c.numbers {
		want := tokens.SyntaxGutter
		if i == c.entry.CursorRow {
			want = tokens.SyntaxGutterCurrent
		}
		if t.Color != want {
			t.Color = want
			t.Refresh()
		}
	}
}

// drawSpan appends the text objects for one token and reports the column
// the next token starts at.
//
// A token is split at every tab, because a tab moves to the next tab stop
// rather than advancing a fixed number of columns.
func (c *Code) drawSpan(objects []fyne.CanvasObject, span Span,
	origin fyne.Position, cell fyne.Size, row, col int) ([]fyne.CanvasObject, int) {

	var run []rune
	start := col

	flush := func() {
		if len(run) == 0 {
			return
		}
		objects = append(objects, spanText(string(run), span, origin, cell, row, start))
		run = nil
	}

	for _, r := range span.Text {
		if r == '\t' {
			flush()
			col = (col/tabColumns + 1) * tabColumns
			start = col
			continue
		}
		if len(run) == 0 {
			start = col
		}
		run = append(run, r)
		col++
	}
	flush()

	return objects, col
}

// spanText builds one positioned run of text.
func spanText(text string, span Span, origin fyne.Position,
	cell fyne.Size, row, col int) fyne.CanvasObject {

	t := canvas.NewText(text, span.Color)
	t.TextSize = fynetheme.TextSize()
	t.TextStyle = fyne.TextStyle{Monospace: true, Bold: span.Bold, Italic: span.Italic}

	t.Move(fyne.NewPos(
		origin.X+cell.Width*float32(col),
		origin.Y+cell.Height*float32(row),
	))
	t.Resize(fyne.NewSize(cell.Width*float32(len([]rune(text))), cell.Height))

	return t
}

// codeCell is the size of one character in the editor's font. Fyne's own
// text widgets measure their grid the same way, so this matches what the
// entry underneath is doing.
func codeCell() fyne.Size {
	return fyne.MeasureText("M", fynetheme.TextSize(), fyne.TextStyle{Monospace: true})
}

// codeOrigin is where the entry starts drawing its first character.
//
// The entry places its text provider one inner padding in on both axes:
// horizontally that is the padding alone, and vertically the border it
// insets by is added straight back on by the provider. The gutter shifts
// the entry itself, so its width comes first.
func codeOrigin(gutter float32) fyne.Position {
	pad := codeOriginPad()
	return fyne.NewPos(gutter+pad, pad)
}

func codeOriginPad() float32 { return fynetheme.InnerPadding() }
