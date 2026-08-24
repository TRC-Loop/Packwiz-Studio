package packwin

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/widgets"
)

// activityBar is the icon rail on the far left. It is icon only, matching
// the locked layout, with a hover tooltip naming each activity since Fyne
// provides none and the glyphs alone would not be discoverable.
type activityBar struct {
	buttons []*railButton
	root    *fyne.Container
}

// newActivityBar builds the rail. onSelect receives the chosen activity's
// identifier.
func newActivityBar(items []Activity, current string, onSelect func(id string)) *activityBar {
	bar := &activityBar{}

	slots := make([]fyne.CanvasObject, 0, len(items))
	for _, a := range items {
		b := newRailButton(a, onSelect)
		b.setActive(a.ID() == current)
		bar.buttons = append(bar.buttons, b)
		slots = append(slots, b)
	}

	bg := canvas.NewRectangle(tokens.ColorSurface)
	column := container.NewVBox(slots...)

	bar.root = container.NewStack(
		bg,
		container.NewBorder(nil, nil, nil, widgets.Hairline(), column),
	)
	return bar
}

// object returns the rail for placement.
func (b *activityBar) object() fyne.CanvasObject { return b.root }

// setCurrent lights the button for id and dims the rest.
func (b *activityBar) setCurrent(id string) {
	for _, btn := range b.buttons {
		btn.setActive(btn.activity.ID() == id)
	}
}

// railButton is one glyph in the rail.
type railButton struct {
	widget.BaseWidget

	activity Activity
	onSelect func(id string)

	tip    *widgets.Tooltip
	active bool
	hover  bool
}

func newRailButton(a Activity, onSelect func(id string)) *railButton {
	b := &railButton{
		activity: a,
		onSelect: onSelect,
		tip:      widgets.NewTooltip(a.Title()),
	}
	b.ExtendBaseWidget(b)
	return b
}

func (b *railButton) setActive(active bool) {
	if b.active == active {
		return
	}
	b.active = active
	b.Refresh()
}

// Tapped implements fyne.Tappable.
func (b *railButton) Tapped(*fyne.PointEvent) {
	if b.onSelect != nil {
		b.onSelect(b.activity.ID())
	}
}

// MouseIn implements desktop.Hoverable.
func (b *railButton) MouseIn(*desktop.MouseEvent) {
	b.hover = true
	b.tip.ShowRightOf(b, tokens.SpaceSM)
	b.Refresh()
}

// MouseOut implements desktop.Hoverable.
func (b *railButton) MouseOut() {
	b.hover = false
	b.tip.Hide()
	b.Refresh()
}

// MouseMoved implements desktop.Hoverable.
func (b *railButton) MouseMoved(*desktop.MouseEvent) {}

// Cursor implements desktop.Cursorable.
func (b *railButton) Cursor() desktop.Cursor { return desktop.PointerCursor }

// CreateRenderer implements fyne.Widget.
func (b *railButton) CreateRenderer() fyne.WidgetRenderer {
	icon := canvas.NewImageFromResource(b.activity.Icon())
	icon.FillMode = canvas.ImageFillContain

	marker := canvas.NewRectangle(tokens.ColorText)

	return &railRenderer{
		owner:  b,
		icon:   icon,
		marker: marker,
	}
}

// railRenderer draws the glyph, a hover wash and an active marker down
// the left edge.
type railRenderer struct {
	owner  *railButton
	icon   *canvas.Image
	marker *canvas.Rectangle
}

// markerWidth is the thickness of the active activity's edge marker.
const markerWidth float32 = 2

func (r *railRenderer) Layout(size fyne.Size) {
	inset := (size.Width - tokens.IconActivity) / 2
	r.icon.Resize(fyne.NewSize(tokens.IconActivity, tokens.IconActivity))
	r.icon.Move(fyne.NewPos(inset, (size.Height-tokens.IconActivity)/2))

	r.marker.Resize(fyne.NewSize(markerWidth, size.Height))
	r.marker.Move(fyne.NewPos(0, 0))
}

func (r *railRenderer) MinSize() fyne.Size {
	return fyne.NewSize(tokens.ActivityBarWidth, tokens.ActivityIconSlot)
}

func (r *railRenderer) Refresh() {
	if r.owner.active {
		r.marker.Show()
		r.icon.Translucency = 0
	} else {
		r.marker.Hide()
		// A dimmed glyph reads as inactive without needing a second
		// colour, since Fyne icons are drawn in the foreground colour.
		r.icon.Translucency = inactiveGlyph
		if r.owner.hover {
			r.icon.Translucency = hoverGlyph
		}
	}
	r.marker.Refresh()
	r.icon.Refresh()
}

// Glyph translucency for the rail's three states.
const (
	inactiveGlyph = 0.45
	hoverGlyph    = 0.15
)

func (r *railRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.marker, r.icon}
}

func (r *railRenderer) Destroy() {}
