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

// activityBar is the icon rail on the far left.
//
// Each entry carries a small caption under its glyph rather than a hover
// tooltip. Fyne has no tooltip, and a hand-rolled one has to be a popup,
// which sits under the pointer and takes the hover for itself: the rail
// entry then sees the pointer leave, hides the popup, receives the
// pointer again, and flickers. A permanent caption cannot do that, and it
// names every section without needing to be discovered.
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

// railButton is one entry in the rail: a glyph over its name.
type railButton struct {
	widget.BaseWidget

	activity Activity
	onSelect func(id string)

	active bool
	hover  bool
}

func newRailButton(a Activity, onSelect func(id string)) *railButton {
	b := &railButton{activity: a, onSelect: onSelect}
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
	b.Refresh()
}

// MouseOut implements desktop.Hoverable.
func (b *railButton) MouseOut() {
	b.hover = false
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

	label := canvas.NewText(b.activity.Title(), tokens.ColorMuted)
	label.TextSize = tokens.TextRailLabel
	label.Alignment = fyne.TextAlignCenter

	return &railRenderer{
		owner:  b,
		bg:     canvas.NewRectangle(tokens.ColorSurface),
		marker: canvas.NewRectangle(tokens.ColorText),
		icon:   icon,
		label:  label,
	}
}

// railRenderer draws the glyph over its caption, with a hover wash and a
// marker down the left edge for the open section.
type railRenderer struct {
	owner  *railButton
	bg     *canvas.Rectangle
	marker *canvas.Rectangle
	icon   *canvas.Image
	label  *canvas.Text
}

// markerWidth is the thickness of the active section's edge marker.
const markerWidth float32 = 2

func (r *railRenderer) Layout(size fyne.Size) {
	r.bg.Resize(size)
	r.bg.Move(fyne.NewPos(0, 0))

	labelHeight := r.label.MinSize().Height
	iconTop := (size.Height - labelHeight - tokens.IconActivity) / 2

	r.icon.Resize(fyne.NewSize(tokens.IconActivity, tokens.IconActivity))
	r.icon.Move(fyne.NewPos((size.Width-tokens.IconActivity)/2, iconTop))

	r.label.Resize(fyne.NewSize(size.Width, labelHeight))
	r.label.Move(fyne.NewPos(0, iconTop+tokens.IconActivity+tokens.SpaceXS))

	r.marker.Resize(fyne.NewSize(markerWidth, size.Height))
	r.marker.Move(fyne.NewPos(0, 0))
}

func (r *railRenderer) MinSize() fyne.Size {
	return fyne.NewSize(tokens.ActivityBarWidth, tokens.ActivityIconSlot)
}

func (r *railRenderer) Refresh() {
	switch {
	case r.owner.active:
		r.bg.FillColor = tokens.ColorSelected
		r.label.Color = tokens.ColorText
		r.icon.Translucency = 0
		r.marker.Show()
	case r.owner.hover:
		r.bg.FillColor = tokens.ColorHover
		r.label.Color = tokens.ColorText
		r.icon.Translucency = 0
		r.marker.Hide()
	default:
		r.bg.FillColor = tokens.ColorSurface
		r.label.Color = tokens.ColorMuted
		// Fyne icons are drawn in one colour, so fading the glyph is how
		// an inactive entry recedes without a second icon set.
		r.icon.Translucency = inactiveGlyph
		r.marker.Hide()
	}

	r.bg.Refresh()
	r.marker.Refresh()
	r.icon.Refresh()
	r.label.Refresh()
}

// inactiveGlyph is how far an unselected glyph fades.
const inactiveGlyph = 0.35

func (r *railRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.bg, r.marker, r.icon, r.label}
}

func (r *railRenderer) Destroy() {}
