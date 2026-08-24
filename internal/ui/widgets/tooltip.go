package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/tokens"
)

// Tooltip is a small label shown beside a widget on hover.
//
// Fyne has no tooltip of its own, and an icon-only control is
// undiscoverable without one, so this is a minimal stand-in: a popup
// holding one line of text, positioned relative to the control that owns
// it. It is not a general purpose tooltip and has no show delay.
type Tooltip struct {
	text  string
	popup *widget.PopUp
}

// NewTooltip returns a tooltip carrying text.
func NewTooltip(text string) *Tooltip {
	return &Tooltip{text: text}
}

// ShowRightOf displays the tooltip beside owner, vertically centred on
// it. A tooltip with no text, or an owner not yet on a canvas, does
// nothing.
func (t *Tooltip) ShowRightOf(owner fyne.CanvasObject, gap float32) {
	if t.text == "" {
		return
	}
	canvasFor := fyne.CurrentApp().Driver().CanvasForObject(owner)
	if canvasFor == nil {
		return
	}

	t.Hide()

	label := Caption(t.text)
	label.Color = tokens.ColorText

	bg := canvas.NewRectangle(tokens.ColorElevated)
	bg.StrokeColor = tokens.ColorBorder
	bg.StrokeWidth = tokens.HairlineThickness
	bg.CornerRadius = tokens.RadiusControl

	content := container.NewStack(bg, Inset(tokens.SpaceMD, tokens.SpaceSM, label))

	t.popup = widget.NewPopUp(content, canvasFor)

	pos := fyne.CurrentApp().Driver().AbsolutePositionForObject(owner)
	size := t.popup.Content.MinSize()
	t.popup.ShowAtPosition(fyne.NewPos(
		pos.X+owner.Size().Width+gap,
		pos.Y+(owner.Size().Height-size.Height)/2,
	))
}

// Hide removes the tooltip if it is showing.
func (t *Tooltip) Hide() {
	if t.popup != nil {
		t.popup.Hide()
		t.popup = nil
	}
}
