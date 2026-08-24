package packwin

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"

	"github.com/TRC-Loop/Packwiz-Studio/internal/pack"
)

// SetLogo copies a chosen PNG into the pack folder as its logo, so it is
// committed with the pack and can be attached to a release.
func (w *Window) SetLogo() {
	open := dialog.NewFileOpen(func(rc fyne.URIReadCloser, err error) {
		if err != nil || rc == nil {
			return
		}
		src := rc.URI().Path()
		rc.Close()

		if err := pack.SetIcon(w.pack.Dir, src); err != nil {
			dialog.ShowError(err, w.win)
			return
		}
		w.rebuildHeader()
	}, w.win)

	open.SetFilter(storage.NewExtensionFileFilter([]string{".png"}))
	if dir, err := storage.ListerForURI(storage.NewFileURI(w.pack.Dir)); err == nil {
		open.SetLocation(dir)
	}
	open.Show()
}

// RemoveLogo drops the pack's logo.
func (w *Window) RemoveLogo() {
	confirm := dialog.NewConfirm("Remove logo",
		"Delete icon.png from the pack folder?",
		func(ok bool) {
			if !ok {
				return
			}
			if err := pack.RemoveIcon(w.pack.Dir); err != nil {
				dialog.ShowError(err, w.win)
				return
			}
			w.rebuildHeader()
		}, w.win)
	confirm.Show()
}

// HasLogo reports whether the pack has a logo, so the menu can offer
// removal only when there is one.
func (w *Window) HasLogo() bool { return pack.IconPath(w.pack.Dir) != "" }

// rebuildHeader redraws the title strip after the logo changed. The
// header holds an image loaded from disk, so it is rebuilt rather than
// refreshed: Fyne caches the decoded file against its path.
func (w *Window) rebuildHeader() {
	w.head.Objects = []fyne.CanvasObject{w.header()}
	w.head.Refresh()
}
