package widgets

import (
	"strings"

	"fyne.io/fyne/v2"
)

// indentOf returns the leading whitespace of a line, which is what a new
// line below it starts with.
func indentOf(line string) string {
	return line[:len(line)-len(strings.TrimLeft(line, " \t"))]
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

// gutterLayout holds the entry to the right of the line numbers.
type gutterLayout struct {
	width float32
}

func (l *gutterLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objects {
		o.Move(fyne.NewPos(l.width, 0))
		o.Resize(fyne.NewSize(size.Width-l.width, size.Height))
	}
}

func (l *gutterLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	min := fyne.NewSize(0, 0)
	for _, o := range objects {
		min = min.Max(o.MinSize())
	}
	return min.AddWidthHeight(l.width, 0)
}
