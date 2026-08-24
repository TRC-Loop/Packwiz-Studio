package widgets

import (
	"image/color"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
)

// Span is a run of text in one colour, as a tokenizer returns it.
type Span struct {
	Text   string
	Color  color.Color
	Bold   bool
	Italic bool
}

// Tokenizer colours whole text, one token list per line. It takes the
// whole text rather than a line because a block comment or an unclosed
// string carries from one line into the next.
type Tokenizer func(text string) [][]Span

// Code is a text editor with line numbers, syntax colouring as you type,
// automatic indentation and word completion.
//
// Fyne has no editable widget that takes a colour per token: an Entry can
// be typed into but draws its text in a single colour, and a TextGrid
// takes a colour per cell but cannot be typed into. So an Entry does the
// editing with its own text made transparent, and a layer of positioned
// text on top supplies the colour. Both draw the same monospace font at
// the same size, so the layer lines up character for character with what
// the Entry believes it is showing.
//
// The layer holds nothing Fyne can send an event to, so every click, drag
// and keystroke reaches the Entry beneath it. The caret, the selection and
// the current line sit under the text, drawn by the Entry and by the
// underlay, so all three stay visible through the layer.
type Code struct {
	// OnChanged reports every edit, after the colouring is redrawn.
	OnChanged func(string)

	entry  *codeEntry
	layer  *fyne.Container
	place  *placedLayout
	gutter *gutterLayout
	line   *canvas.Rectangle
	scroll *container.Scroll
	root   fyne.CanvasObject

	tokenize Tokenizer
	numbers  []*canvas.Text
	rows     int
}

// NewCode builds an editor with nothing open in it.
func NewCode() *Code {
	c := &Code{
		place:  &placedLayout{},
		gutter: &gutterLayout{},
		line:   canvas.NewRectangle(tokens.SyntaxCurrentLine),
	}

	c.entry = newCodeEntry(c)
	c.layer = container.New(c.place)

	hidden := container.NewThemeOverride(c.entry, newInvisibleText())
	held := container.New(c.gutter, hidden)

	// The current line sits under everything, including the entry, so the
	// caret and any selection are drawn on top of it rather than hidden
	// behind it.
	underlay := container.NewWithoutLayout(c.line)

	// The suggestion list is a positioned child of the editor rather than
	// an overlay, so that showing it does not move the keyboard focus.
	menu := container.NewWithoutLayout()
	c.entry.menu = newCompletions(menu, c.entry.take)

	c.scroll = container.NewScroll(
		container.NewStack(underlay, held, c.layer, menu))
	c.root = c.scroll

	return c
}

// Object returns the editor for placement.
func (c *Code) Object() fyne.CanvasObject { return c.root }

// SetTokenizer chooses how the text is coloured.
func (c *Code) SetTokenizer(t Tokenizer) {
	c.tokenize = t
	c.repaint()
}

// SetWords gives the completion popup the language's own words, on top of
// whatever the file itself contains.
func (c *Code) SetWords(list []string) { c.entry.words = list }

// Text is the current content.
func (c *Code) Text() string { return c.entry.Text }

// SetText replaces the content and redraws.
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

// changed is the entry's edit handler.
func (c *Code) changed(text string) {
	c.repaint()
	if c.OnChanged != nil {
		c.OnChanged(text)
	}
}

// gutterWidth is how much room the line numbers need for this file. It is
// measured from the line count so that the text does not shift sideways
// when a file grows past ten or a hundred lines while being edited.
func (c *Code) gutterWidth() float32 {
	digits := len(strconv.Itoa(max(c.rows, 1)))
	return codeCell().Width*float32(digits+1) + gutterPad
}

// markLine puts the current line tint behind the caret's row and follows
// the caret with the scroll.
//
// The entry reports a cursor move while it is still being built, before
// this widget has a scroll to move, so the guard is not optional.
func (c *Code) markLine() {
	if c.scroll == nil {
		return
	}

	cell := codeCell()
	origin := codeOrigin(c.gutter.width)

	width := max(c.scroll.Content.Size().Width, c.scroll.Size().Width)
	c.line.Resize(fyne.NewSize(width, cell.Height))
	c.line.Move(fyne.NewPos(0, origin.Y+cell.Height*float32(c.entry.CursorRow)))
	c.line.Refresh()

	c.repaintGutter()
	c.followCursor()
}

// followCursor keeps the caret in view. The entry does no scrolling of
// its own here, so moving off the visible area is this widget's job.
func (c *Code) followCursor() {
	if c.scroll == nil {
		return
	}

	cell := codeCell()
	origin := codeOrigin(c.gutter.width)

	top := origin.Y + cell.Height*float32(c.entry.CursorRow)
	left := origin.X + cell.Width*float32(c.entry.CursorColumn)

	view := c.scroll.Size()
	off := c.scroll.Offset

	switch {
	case top < off.Y:
		off.Y = top
	case top+cell.Height > off.Y+view.Height:
		off.Y = top + cell.Height - view.Height
	}
	switch {
	case left-origin.X < off.X:
		off.X = max(left-origin.X, 0)
	case left+cell.Width > off.X+view.Width:
		off.X = left + cell.Width - view.Width
	}

	if off != c.scroll.Offset {
		c.scroll.ScrollToOffset(off)
	}
}
