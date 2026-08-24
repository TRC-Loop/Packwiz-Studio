package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
)

// completionLimit is how many suggestions are offered at once. A longer
// list is quicker to type past than to read.
const completionLimit = 8

// completionMinPrefix is how much has to be typed before suggestions
// appear, so they do not interrupt ordinary typing.
const completionMinPrefix = 2

// completions is the suggestion list.
//
// It is drawn inside the editor rather than as a popup on purpose. A Fyne
// overlay brings its own focus manager, which would take the keyboard
// away from the entry the moment the list appeared, and an editor whose
// completion list stops you typing is worse than none.
type completions struct {
	host   *fyne.Container
	pick   func(string)
	rows   []*Clickable
	items  []string
	index  int
	prefix string
}

func newCompletions(host *fyne.Container, pick func(string)) *completions {
	return &completions{host: host, pick: pick}
}

// open reports whether the list is showing.
func (c *completions) open() bool { return len(c.items) > 0 }

// show lists items at a position in the editor's content.
func (c *completions) show(items []string, prefix string, at fyne.Position) {
	c.items = items
	c.prefix = prefix
	c.index = 0
	c.rows = nil

	if len(items) == 0 {
		c.hide()
		return
	}

	list := container.NewVBox()
	for i, item := range items {
		text := item
		row := NewClickable(Caption(text), func() { c.pick(text) })
		row.SetSelected(i == 0)

		c.rows = append(c.rows, row)
		list.Add(row)
	}

	bg := canvas.NewRectangle(tokens.ColorElevated)
	bg.StrokeColor = tokens.ColorBorder
	bg.StrokeWidth = tokens.HairlineThickness
	bg.CornerRadius = tokens.RadiusOverlay

	box := container.NewStack(bg, list)
	box.Resize(box.MinSize())
	box.Move(at)

	c.host.Objects = []fyne.CanvasObject{box}
	c.host.Refresh()
}

// hide takes the list away.
func (c *completions) hide() {
	c.items = nil
	c.rows = nil

	if len(c.host.Objects) == 0 {
		return
	}
	c.host.Objects = nil
	c.host.Refresh()
}

// move walks the selection, wrapping at both ends.
func (c *completions) move(delta int) {
	if !c.open() {
		return
	}

	c.index = (c.index + delta + len(c.items)) % len(c.items)
	for i, row := range c.rows {
		row.SetSelected(i == c.index)
	}
}

// selected is the highlighted item.
func (c *completions) selected() string {
	if !c.open() {
		return ""
	}
	return c.items[c.index]
}
