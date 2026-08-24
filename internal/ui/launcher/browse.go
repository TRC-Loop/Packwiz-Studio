package launcher

import (
	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"

	"github.com/PalisadeMC/Packwiz-Studio/internal/packwiz"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/widgets"
)

// report is how a caller receives a message about what happened, so the
// same flow can write into whichever status line the calling screen owns.
type report func(text string, state widgets.State)

// browseForBinary lets the user point at a packwiz they already have.
//
// The binary is checked before its path is stored, so a wrong pick fails
// here with an explanation rather than at the first pack operation.
func (w *Window) browseForBinary(say report) {
	open := dialog.NewFileOpen(func(rc fyne.URIReadCloser, err error) {
		if err != nil || rc == nil {
			return
		}
		path := rc.URI().Path()
		rc.Close()

		say("Checking "+path, widgets.StateNeutral)
		go w.verifyBinary(path, say)
	}, w.win)

	if dir, err := storage.ListerForURI(storage.NewFileURI(defaultBrowseDir())); err == nil {
		open.SetLocation(dir)
	}
	open.Show()
}

// verifyBinary checks a chosen path off the main goroutine and records it
// when it works.
func (w *Window) verifyBinary(path string, say report) {
	loc, err := packwiz.Verify(context.Background(), path)

	fyne.Do(func() {
		if err != nil {
			say(err.Error(), widgets.StateError)
			return
		}
		if err := w.sess.SetPackwiz(loc); err != nil {
			dialog.ShowError(err, w.win)
			return
		}
		w.Refresh()
	})
}
