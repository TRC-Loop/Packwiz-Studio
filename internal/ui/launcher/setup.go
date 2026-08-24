package launcher

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	fynetheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/TRC-Loop/Packwiz-Studio/internal/packwiz"
	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/tokens"
	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/widgets"
)

// buildSetup is the screen shown until packwiz resolves. It offers the
// two ways out of the situation and can be dismissed.
func (w *Window) buildSetup() fyne.CanvasObject {
	blurb := widget.NewLabel(
		"Packwiz Studio drives the packwiz command line tool,\n" +
			"which is not on your PATH yet.")
	blurb.Wrapping = fyne.TextWrapWord

	progress := widget.NewProgressBar()
	progress.Hide()

	note := widgets.Caption("")
	note.Hide()

	var install, browse, skip *widget.Button

	setBusy := func(busy bool) {
		for _, b := range []*widget.Button{install, browse, skip} {
			if busy {
				b.Disable()
			} else {
				b.Enable()
			}
		}
	}

	install = widget.NewButtonWithIcon("Install packwiz", fynetheme.DownloadIcon(), func() {
		setBusy(true)
		progress.SetValue(0)
		progress.Show()
		note.Text = "Contacting GitHub"
		note.Color = tokens.ColorMuted
		note.Show()
		note.Refresh()

		go w.runInstall(progress, note, setBusy)
	})

	browse = widget.NewButtonWithIcon("Choose an existing binary", fynetheme.FolderOpenIcon(), func() {
		w.browseForBinary(note)
	})

	skip = widget.NewButton("Continue without packwiz", func() {
		w.setupDismissed = true
		w.Refresh()
	})
	skip.Importance = widget.LowImportance

	buttons := container.NewHBox(install, browse)

	body := container.NewVBox(
		widgets.Heading("Packwiz Studio"),
		blurb,
		widgets.VSpace(tokens.SpaceMD),
		buttons,
		progress,
		note,
		widgets.VSpace(tokens.SpaceLG),
		container.NewHBox(skip),
	)

	return widgets.Inset(tokens.SpaceXXL, tokens.SpaceXL, body)
}

// runInstall downloads packwiz off the main goroutine, reporting progress
// back onto it.
func (w *Window) runInstall(progress *widget.ProgressBar, note *canvas.Text, setBusy func(bool)) {
	installer := &packwiz.Installer{}

	loc, err := installer.Install(context.Background(), func(done, total int64) {
		fyne.Do(func() {
			if total > 0 {
				progress.SetValue(float64(done) / float64(total))
				note.Text = fmt.Sprintf("Downloading %s of %s", megabytes(done), megabytes(total))
			} else {
				note.Text = "Downloading " + megabytes(done)
			}
			note.Refresh()
		})
	})

	fyne.Do(func() {
		setBusy(false)
		progress.Hide()

		if err != nil {
			note.Text = err.Error()
			note.Color = tokens.ColorError
			note.Refresh()
			return
		}

		if err := w.sess.SetPackwiz(loc); err != nil {
			dialog.ShowError(err, w.win)
			return
		}
		w.Refresh()
	})
}

// browseForBinary lets the user point at a packwiz they already have. The
// binary is verified before the path is stored, so a wrong pick fails
// here rather than at the first command.
func (w *Window) browseForBinary(note *canvas.Text) {
	open := dialog.NewFileOpen(func(rc fyne.URIReadCloser, err error) {
		if err != nil || rc == nil {
			return
		}
		path := rc.URI().Path()
		rc.Close()

		go func() {
			loc, err := packwiz.Verify(context.Background(), path)
			fyne.Do(func() {
				if err != nil {
					note.Text = err.Error()
					note.Color = tokens.ColorError
					note.Show()
					note.Refresh()
					return
				}
				if err := w.sess.SetPackwiz(loc); err != nil {
					dialog.ShowError(err, w.win)
					return
				}
				w.Refresh()
			})
		}()
	}, w.win)

	if dir, err := storage.ListerForURI(storage.NewFileURI(defaultBrowseDir())); err == nil {
		open.SetLocation(dir)
	}
	open.Show()
}

// megabytes renders a byte count for a progress line.
func megabytes(n int64) string {
	return fmt.Sprintf("%.1f MB", float64(n)/(1<<20))
}
