package widgets

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Span is a run of text in one colour, as a highlighter returns it.
type Span struct {
	Text   string
	Color  color.Color
	Bold   bool
	Italic bool
}

// Code is a plain text editor that paints syntax highlighting over the
// text as it is typed.
//
// Fyne has no editable widget that takes a colour per token: an Entry can
// be typed into but draws its text in a single colour, and a TextGrid
// takes a colour per cell but cannot be typed into. So the Entry does the
// editing with its own text made transparent, and a layer of positioned
// text on top supplies the colour. Both draw the same monospace font at
// the same size, so the layer lines up character for character with what
// the Entry believes it is showing.
//
// The layer is not interactive: it holds nothing that Fyne can send an
// event to, so every click, drag and keystroke reaches the Entry beneath
// it. Selection and the cursor are drawn under the text by the Entry, so
// they stay visible through the layer.
type Code struct {
	// OnChanged reports every edit, after the highlighting is redrawn.
	OnChanged func(string)

	entry    *widget.Entry
	layer    *fyne.Container
	place    *placedLayout
	scroll   *container.Scroll
	root     fyne.CanvasObject
	tokenize func(string) []Span
}

// NewCode builds an editor that colours each line with tokenize.
func NewCode(tokenize func(string) []Span) *Code {
	c := &Code{
		tokenize: tokenize,
		place:    &placedLayout{},
		entry:    widget.NewMultiLineEntry(),
	}

	c.entry.TextStyle = fyne.TextStyle{Monospace: true}

	// Wrapping off and scrolling off together are what make the entry size
	// itself to its content. It then sits inside this widget's own scroll,
	// which the highlighted layer shares, so the two cannot drift apart.
	c.entry.Wrapping = fyne.TextWrapOff
	c.entry.Scroll = fyne.ScrollNone

	c.entry.OnChanged = func(text string) {
		c.repaint()
		if c.OnChanged != nil {
			c.OnChanged(text)
		}
	}
	c.entry.OnCursorChanged = c.followCursor

	c.layer = container.New(c.place)

	hidden := container.NewThemeOverride(c.entry, newInvisibleText())
	c.scroll = container.NewScroll(container.NewStack(hidden, c.layer))
	c.root = c.scroll

	return c
}

// Object returns the editor for placement.
func (c *Code) Object() fyne.CanvasObject { return c.root }

// Text is the current content.
func (c *Code) Text() string { return c.entry.Text }

// SetText replaces the content and redraws the highlighting.
func (c *Code) SetText(text string) {
	c.entry.SetText(text)
	c.repaint()
	c.scroll.ScrollToTop()
}

// Focus puts the caret in the editor, so a file opens ready to type in.
func (c *Code) Focus(canvas fyne.Canvas) {
	if canvas != nil {
		canvas.Focus(c.entry)
	}
}

// followCursor keeps the caret in view. The entry does no scrolling of
// its own here, so moving off the visible area is this widget's job.
func (c *Code) followCursor() {
	cell := codeCell()
	origin := codeOrigin()

	top := origin.Y + float32(c.entry.CursorRow)*cell.Height
	left := origin.X + float32(c.entry.CursorColumn)*cell.Width

	view := c.scroll.Size()
	off := c.scroll.Offset

	switch {
	case top < off.Y:
		off.Y = top
	case top+cell.Height > off.Y+view.Height:
		off.Y = top + cell.Height - view.Height
	}
	switch {
	case left < off.X:
		off.X = left
	case left+cell.Width > off.X+view.Width:
		off.X = left + cell.Width - view.Width
	}

	if off != c.scroll.Offset {
		c.scroll.ScrollToOffset(off)
	}
}
