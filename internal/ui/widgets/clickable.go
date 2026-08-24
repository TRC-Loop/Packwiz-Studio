package widgets

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/tokens"
)

// Clickable wraps arbitrary content in a tappable surface that responds
// to hover, for rows and cards that behave like buttons but hold a
// layout rather than a label.
type Clickable struct {
	widget.BaseWidget

	// OnTapped runs when the surface is clicked. A nil value makes the
	// surface inert but still rendered.
	OnTapped func()

	content  fyne.CanvasObject
	radius   float32
	hovered  bool
	selected bool
}

// NewClickable returns a tappable surface around content.
func NewClickable(content fyne.CanvasObject, onTapped func()) *Clickable {
	c := &Clickable{content: content, radius: tokens.RadiusCard, OnTapped: onTapped}
	c.ExtendBaseWidget(c)
	return c
}

// SetSelected marks the surface as the current selection, which keeps its
// background lit while the pointer is elsewhere.
func (c *Clickable) SetSelected(selected bool) {
	if c.selected == selected {
		return
	}
	c.selected = selected
	c.Refresh()
}

// Tapped implements fyne.Tappable.
func (c *Clickable) Tapped(*fyne.PointEvent) {
	if c.OnTapped != nil {
		c.OnTapped()
	}
}

// MouseIn implements desktop.Hoverable.
func (c *Clickable) MouseIn(*desktop.MouseEvent) {
	c.hovered = true
	c.Refresh()
}

// MouseOut implements desktop.Hoverable.
func (c *Clickable) MouseOut() {
	c.hovered = false
	c.Refresh()
}

// MouseMoved implements desktop.Hoverable.
func (c *Clickable) MouseMoved(*desktop.MouseEvent) {}

// Cursor implements desktop.Cursorable, so the pointer signals that the
// surface can be clicked.
func (c *Clickable) Cursor() desktop.Cursor {
	if c.OnTapped == nil {
		return desktop.DefaultCursor
	}
	return desktop.PointerCursor
}

// CreateRenderer implements fyne.Widget.
func (c *Clickable) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewRectangle(c.fill())
	bg.CornerRadius = c.radius

	padded := container.NewPadded(c.content)
	return &clickableRenderer{
		owner:   c,
		bg:      bg,
		content: padded,
		objects: []fyne.CanvasObject{bg, padded},
	}
}

// fill is the surface colour for the current state. An idle surface is
// transparent so a row blends into whatever pane holds it.
func (c *Clickable) fill() color.Color {
	switch {
	case c.selected:
		return tokens.ColorSelected
	case c.hovered:
		return tokens.ColorHover
	default:
		return color.Transparent
	}
}

type clickableRenderer struct {
	owner   *Clickable
	bg      *canvas.Rectangle
	content *fyne.Container
	objects []fyne.CanvasObject
}

func (r *clickableRenderer) Layout(size fyne.Size) {
	r.bg.Resize(size)
	r.content.Resize(size)
}

func (r *clickableRenderer) MinSize() fyne.Size { return r.content.MinSize() }

func (r *clickableRenderer) Refresh() {
	r.bg.FillColor = r.owner.fill()
	r.bg.CornerRadius = r.owner.radius
	r.bg.Refresh()
	r.content.Refresh()
}

func (r *clickableRenderer) Objects() []fyne.CanvasObject { return r.objects }

func (r *clickableRenderer) Destroy() {}
