package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

// VSpace is a fixed vertical gap, for separating groups inside a box
// layout where the layout's own spacing is the wrong step.
func VSpace(height float32) fyne.CanvasObject {
	r := canvas.NewRectangle(nil)
	r.SetMinSize(fyne.NewSize(0, height))
	return r
}

// HSpace is a fixed horizontal gap.
func HSpace(width float32) fyne.CanvasObject {
	r := canvas.NewRectangle(nil)
	r.SetMinSize(fyne.NewSize(width, 0))
	return r
}
