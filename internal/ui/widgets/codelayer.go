package widgets

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	fynetheme "fyne.io/fyne/v2/theme"
)

// tabColumns is how far a tab advances, matching what Fyne's own text
// painter does. The editor's columns have to agree with the entry's
// drawing, or the layer would sit part of a character out on any line
// with a tab in it.
const tabColumns = 4

// repaint rebuilds the highlighted layer from the current text.
//
// One text object per token rather than per character: a file of a few
// hundred lines is then a few thousand objects at worst, where a grid of
// cells would be tens of thousands.
func (c *Code) repaint() {
	cell := codeCell()
	origin := codeOrigin()

	lines := strings.Split(c.entry.Text, "\n")
	objects := make([]fyne.CanvasObject, 0, len(lines)*4)
	widest := 0

	for row, line := range lines {
		col := 0
		for _, span := range c.tokenize(line) {
			objects, col = c.drawSpan(objects, span, origin, cell, row, col)
		}
		widest = max(widest, col)
	}

	c.place.min = fyne.NewSize(
		origin.X*2+cell.Width*float32(widest),
		origin.Y*2+cell.Height*float32(len(lines)),
	)

	c.layer.Objects = objects
	c.layer.Refresh()
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
		objects = append(objects, c.spanText(string(run), span, origin, cell, row, start))
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
func (c *Code) spanText(text string, span Span, origin fyne.Position,
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
// insets by is added straight back on by the provider.
func codeOrigin() fyne.Position {
	pad := fynetheme.InnerPadding()
	return fyne.NewPos(pad, pad)
}

// placedLayout leaves its objects where they were put.
//
// The layer positions every run itself, on a character grid, so a layout
// that arranged them would only undo that work. It exists to report the
// size the runs add up to.
type placedLayout struct {
	min fyne.Size
}

func (l *placedLayout) Layout([]fyne.CanvasObject, fyne.Size) {}

func (l *placedLayout) MinSize([]fyne.CanvasObject) fyne.Size { return l.min }
