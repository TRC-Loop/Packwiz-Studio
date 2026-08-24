package launcher

import (
	"fyne.io/fyne/v2"

	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
)

// stackedLines lays text lines out top to bottom at their own heights,
// with one gap between them and no stretching.
//
// A vertical box would work, except it distributes leftover space and
// pads with the theme's own step, which leaves the three lines of a
// recents row unevenly spaced against the image beside them. This gives
// every line exactly its height and the same gap.
type stackedLines struct{}

func (l *stackedLines) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	y := float32(0)

	for _, o := range objects {
		height := o.MinSize().Height
		o.Resize(fyne.NewSize(size.Width, height))
		o.Move(fyne.NewPos(0, y))
		y += height + tokens.SpaceXS
	}
}

func (l *stackedLines) MinSize(objects []fyne.CanvasObject) fyne.Size {
	var min fyne.Size

	for i, o := range objects {
		child := o.MinSize()

		min.Width = max(min.Width, child.Width)
		min.Height += child.Height
		if i > 0 {
			min.Height += tokens.SpaceXS
		}
	}
	return min
}
