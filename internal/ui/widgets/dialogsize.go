package widgets

import (
	"fyne.io/fyne/v2"
)

// dialogFill is how much of the window a dialog may occupy. Leaving a
// margin keeps the dialog readable as a layer above the window rather
// than covering it edge to edge.
const dialogFill = 0.9

// FitDialog returns a dialog size that fits inside its parent window.
//
// A dialog sized past its window is clipped: its lower content and its
// own buttons are simply cut off, which looks like content that will not
// scroll. Clamping to the window is what keeps the buttons reachable.
//
// The preferred size is used when there is room for it, so small dialogs
// are not stretched to fill a large window.
func FitDialog(win fyne.Window, prefWidth, prefHeight float32) fyne.Size {
	width, height := prefWidth, prefHeight

	if win != nil && win.Canvas() != nil {
		available := win.Canvas().Size()

		if available.Width > 0 {
			width = min(width, available.Width*dialogFill)
		}
		if available.Height > 0 {
			height = min(height, available.Height*dialogFill)
		}
	}

	return fyne.NewSize(max(width, minDialogWidth), max(height, minDialogHeight))
}

// A dialog below these sizes would be unusable whatever the window does.
const (
	minDialogWidth  float32 = 320
	minDialogHeight float32 = 220
)
