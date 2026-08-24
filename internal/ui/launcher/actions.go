package launcher

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	fynetheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/TRC-Loop/Packwiz-Studio/internal/config"
	"github.com/TRC-Loop/Packwiz-Studio/internal/pack"
	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/tokens"
	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/widgets"
)

// actionColumn is the launcher's right hand side. Pack actions need
// packwiz, so they are disabled until a binary is available; Settings
// stays live, because it is where the path gets fixed.
func (w *Window) actionColumn() []fyne.CanvasObject {
	newPack := widget.NewButtonWithIcon("New pack", fynetheme.ContentAddIcon(), w.showNewPack)
	openPack := widget.NewButtonWithIcon("Open pack", fynetheme.FolderOpenIcon(), w.showOpenPack)
	settings := widget.NewButtonWithIcon("Settings", fynetheme.SettingsIcon(), w.showSettings)

	if !w.sess.HasPackwiz() {
		newPack.Disable()
		openPack.Disable()
	}

	column := []fyne.CanvasObject{newPack, openPack, settings}
	if !w.sess.HasPackwiz() && w.setupDismissed {
		column = append(column,
			widgets.VSpace(tokens.SpaceMD),
			widget.NewButton("Set up packwiz", func() {
				w.setupDismissed = false
				w.Refresh()
			}),
		)
	}

	return []fyne.CanvasObject{container.NewVBox(column...)}
}

// forgetButton removes a pack from the recents list. The pack itself is
// left alone, so the icon is a list removal rather than a delete.
func (w *Window) forgetButton(p config.Pack) fyne.CanvasObject {
	btn := widget.NewButtonWithIcon("", fynetheme.ContentClearIcon(), func() {
		if err := w.sess.Cfg.Forget(p.Path); err != nil {
			dialog.ShowError(err, w.win)
			return
		}
		w.Refresh()
	})
	btn.Importance = widget.LowImportance
	return container.NewCenter(btn)
}

// showOpenPack asks for a folder and checks it holds a pack before
// opening it, so a wrong pick is refused here rather than failing later.
func (w *Window) showOpenPack() {
	open := dialog.NewFolderOpen(func(list fyne.ListableURI, err error) {
		if err != nil || list == nil {
			return
		}
		dir := list.Path()

		p, err := pack.Load(dir)
		if err != nil {
			dialog.ShowError(err, w.win)
			return
		}
		if err := w.sess.Cfg.Touch(p.Dir, p.Name, p.MCVersion, p.Loader); err != nil {
			dialog.ShowError(err, w.win)
			return
		}
		w.openPack(p.Dir)
	}, w.win)

	if dir, err := storage.ListerForURI(storage.NewFileURI(defaultBrowseDir())); err == nil {
		open.SetLocation(dir)
	}
	open.Show()
}
