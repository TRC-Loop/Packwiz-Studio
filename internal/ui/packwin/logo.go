package packwin

import (
	"fyne.io/fyne/v2"
)

// rebuildHeader redraws the title strip after the pack's name or image
// changed.
//
// The header holds an image loaded from a file, and Fyne caches a decoded
// file against its path, so the header is rebuilt rather than refreshed.
func (w *Window) rebuildHeader() {
	w.head.Objects = []fyne.CanvasObject{w.header()}
	w.head.Refresh()
}
