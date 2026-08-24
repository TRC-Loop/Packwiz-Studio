package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/tokens"
)

// Scrollable wraps tall content so it scrolls inside a dialog.
//
// A plain scroll container reports its content's minimum size as its own,
// and a dialog sizes itself to its content's minimum. The scroll area
// therefore grows to fit everything, the dialog grows past the window,
// and nothing ever scrolls. Capping the size here is what makes the
// scrollbar do its job.
//
// The height is a cap, not a fixed size: content shorter than it is left
// alone rather than padded out.
func Scrollable(width, height float32, content fyne.CanvasObject) fyne.CanvasObject {
	scroll := container.NewVScroll(content)

	// Room is left for the dialog's own title bar and button row, so the
	// scroll area does not push them off the bottom.
	inner := height - tokens.DialogChrome

	if content.MinSize().Height < inner {
		return container.New(&capLayout{width: width}, scroll)
	}
	return container.New(&capLayout{width: width, height: inner}, scroll)
}

// capLayout gives its child a bounded minimum size, so a scroll container
// stops advertising the full height of what it holds.
type capLayout struct {
	width  float32
	height float32
}

func (l *capLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objects {
		o.Resize(size)
		o.Move(fyne.NewPos(0, 0))
	}
}

func (l *capLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	min := fyne.NewSize(l.width, 0)

	if l.height > 0 {
		min.Height = l.height
		return min
	}

	for _, o := range objects {
		min.Height = max(min.Height, o.MinSize().Height)
	}
	return min
}
