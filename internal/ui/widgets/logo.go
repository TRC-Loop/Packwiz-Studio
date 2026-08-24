package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/tokens"
)

// PackLogo renders a pack's icon.png at the given side length. A pack
// without a logo gets a plain placeholder square rather than a stock
// icon, so a list of packs stays visually quiet.
func PackLogo(path string, size float32) fyne.CanvasObject {
	box := fyne.NewSize(size, size)

	if path == "" {
		return placeholder(box)
	}

	img := canvas.NewImageFromFile(path)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(box)

	// A file that turns out not to be a decodable image renders as
	// nothing, which would leave a hole in the row. The placeholder sits
	// behind it so the slot is filled either way.
	return container.NewStack(placeholder(box), img)
}

// placeholder is the empty logo slot: a filled square with a hairline
// border, matching the card radius.
func placeholder(size fyne.Size) fyne.CanvasObject {
	r := canvas.NewRectangle(tokens.ColorElevated)
	r.StrokeColor = tokens.ColorBorder
	r.StrokeWidth = tokens.HairlineThickness
	r.CornerRadius = tokens.RadiusCard
	r.SetMinSize(size)
	return r
}
