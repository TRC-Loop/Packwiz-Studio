package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

// Inset wraps content with a fixed padding on each axis. Fyne's padded
// container uses the theme padding, which is one step of the scale and
// often not the step a given pane wants.
func Inset(x, y float32, content fyne.CanvasObject) *fyne.Container {
	return container.New(&insetLayout{x: x, y: y}, content)
}

type insetLayout struct {
	x, y float32
}

func (l *insetLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	inner := fyne.NewSize(size.Width-2*l.x, size.Height-2*l.y)
	for _, o := range objects {
		o.Move(fyne.NewPos(l.x, l.y))
		o.Resize(inner)
	}
}

func (l *insetLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	var min fyne.Size
	for _, o := range objects {
		min = min.Max(o.MinSize())
	}
	return fyne.NewSize(min.Width+2*l.x, min.Height+2*l.y)
}
