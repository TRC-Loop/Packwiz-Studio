package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

// Fixed pins content to exactly one size, whatever the layout around it
// would rather do.
//
// An image's SetMinSize only sets a floor, so a border or box layout is
// free to stretch it to fill the row: a square icon then comes out
// rectangular. This holds the size on both axes and centres the result in
// whatever space it is given.
func Fixed(size fyne.Size, content fyne.CanvasObject) fyne.CanvasObject {
	return container.NewCenter(container.New(&fixedLayout{size: size}, content))
}

// FixedSquare is Fixed for the common case of a square.
func FixedSquare(side float32, content fyne.CanvasObject) fyne.CanvasObject {
	return Fixed(fyne.NewSize(side, side), content)
}

type fixedLayout struct {
	size fyne.Size
}

func (l *fixedLayout) Layout(objects []fyne.CanvasObject, _ fyne.Size) {
	for _, o := range objects {
		o.Resize(l.size)
		o.Move(fyne.NewPos(0, 0))
	}
}

func (l *fixedLayout) MinSize([]fyne.CanvasObject) fyne.Size { return l.size }
