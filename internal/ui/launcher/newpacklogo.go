package launcher

import (
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"

	"github.com/PalisadeMC/Packwiz-Studio/internal/pack"
)

// noLogoLabel is what the image field says with nothing chosen.
const noLogoLabel = "No image chosen"

// pickLogo chooses the pack's image.
//
// The file is only remembered here. It is copied into the pack once
// packwiz has created the folder, because there is nowhere to put it
// before that.
func (f *newPackForm) pickLogo() {
	open := dialog.NewFileOpen(func(rc fyne.URIReadCloser, err error) {
		if err != nil || rc == nil {
			return
		}
		path := rc.URI().Path()
		rc.Close()

		f.logo = path
		f.logoLabel.SetText(filepath.Base(path))
	}, f.win.win)

	open.SetFilter(storage.NewExtensionFileFilter(pack.ImageExtensions))
	open.Show()
}

// clearLogo forgets the chosen image.
func (f *newPackForm) clearLogo() {
	f.logo = ""
	f.logoLabel.SetText(noLogoLabel)
}
