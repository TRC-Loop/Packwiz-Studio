package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/tokens"
)

// NewPanel wraps content in a raised surface with a hairline border, for
// a card in a grid.
//
// Fyne's own Card widget adds a title row and its own padding scale, so
// this is a plain surface instead: the card's own layout decides what
// goes in it.
func NewPanel(content fyne.CanvasObject) fyne.CanvasObject {
	bg := canvas.NewRectangle(tokens.ColorElevated)
	bg.StrokeColor = tokens.ColorBorder
	bg.StrokeWidth = tokens.HairlineThickness
	bg.CornerRadius = tokens.RadiusCard

	return container.NewStack(bg, Inset(tokens.SpaceMD, tokens.SpaceMD, content))
}
