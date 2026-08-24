package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

// ClampWidth stops a child's width requirement from reaching the window.
//
// A scroll container only absorbs the axis it scrolls: a vertical scroll
// still reports its content's full width as a minimum, and Fyne grows the
// window to satisfy a minimum. So one long line of unwrapped text
// anywhere inside a scrolling list makes the whole window wider, and the
// window then jumps about as the list's contents change. Capping the
// width here keeps that from escaping.
//
// The child is still laid out at the container's real width, so nothing
// is squashed: only the minimum is capped.
func ClampWidth(width float32, child fyne.CanvasObject) fyne.CanvasObject {
	return container.New(&clampLayout{width: width}, child)
}

// clampLayout fills its container, reporting a fixed width as its own
// minimum and passing the child's height requirement through.
type clampLayout struct {
	width float32
}

func (l *clampLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objects {
		o.Resize(size)
		o.Move(fyne.NewPos(0, 0))
	}
}

func (l *clampLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	height := float32(0)
	for _, o := range objects {
		height = fyne.Max(height, o.MinSize().Height)
	}
	return fyne.NewSize(l.width, height)
}
