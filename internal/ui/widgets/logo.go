package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
)

// PackLogo renders a pack's image at the given side length, pinned square
// so a row layout cannot stretch it.
//
// A pack with no image gets nothing at all. An empty placeholder square
// in every row is just noise, and most packs have no image.
func PackLogo(path string, size float32) fyne.CanvasObject {
	if path == "" {
		return nil
	}

	box := fyne.NewSize(size, size)

	img := canvas.NewImageFromFile(path)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(box)

	// A file that turns out not to be a decodable image renders as
	// nothing, so the placeholder sits behind it to keep the slot filled.
	return FixedSquare(size, container.NewStack(placeholder(box), img))
}

// PackLogoOrPlaceholder always renders a slot, for a screen that shows
// what the image would be rather than a list of packs.
func PackLogoOrPlaceholder(path string, size float32) fyne.CanvasObject {
	if logo := PackLogo(path, size); logo != nil {
		return logo
	}
	return FixedSquare(size, placeholder(fyne.NewSize(size, size)))
}

// placeholder is the empty image slot: a filled square with a hairline
// border, matching the card radius.
func placeholder(size fyne.Size) fyne.CanvasObject {
	r := canvas.NewRectangle(tokens.ColorElevated)
	r.StrokeColor = tokens.ColorBorder
	r.StrokeWidth = tokens.HairlineThickness
	r.CornerRadius = tokens.RadiusCard
	r.SetMinSize(size)
	return r
}
