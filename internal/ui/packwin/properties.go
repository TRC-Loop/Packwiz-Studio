package packwin

import (
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	fynetheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/PalisadeMC/Packwiz-Studio/internal/pack"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/widgets"
)

// EditProperties edits the pack's own metadata: its name, author, version
// and image.
//
// The Minecraft and loader versions are deliberately absent. Changing
// those is a migration, not an edit: packwiz has a migrate command for it
// and doing it by rewriting pack.toml would leave every mod pinned to the
// old version.
func (w *Window) EditProperties() {
	name := widget.NewEntry()
	name.SetText(w.pack.Name)

	author := widget.NewEntry()
	author.SetText(w.pack.Author)

	version := widget.NewEntry()
	version.SetText(w.pack.Version)

	logo := &logoField{win: w}
	logo.build()

	body := container.NewVBox(
		widgets.Muted("Pack name"),
		name,
		widgets.Muted("Author"),
		author,
		widgets.Muted("Pack version"),
		version,

		widgets.VSpace(tokens.SpaceSM),
		widgets.Muted("Pack image"),
		logo.row,
		widgets.Note("Stored in the pack folder as icon.png. Any common image "+
			"format is accepted and converted."),

		widgets.VSpace(tokens.SpaceSM),
		widgets.Note("Minecraft and loader versions are not edited here. "+
			"Changing those is a migration, which packwiz handles separately."),
	)

	size := widgets.FitDialog(w.win, tokens.FormWidth, tokens.PropertiesHeight)

	d := dialog.NewCustomConfirm("Pack properties", "Save", "Cancel",
		widgets.Scrollable(size.Width, size.Height,
			widgets.Inset(tokens.SpaceMD, tokens.SpaceMD, body)),
		func(save bool) {
			if !save {
				return
			}
			w.saveProperties(pack.Properties{
				Name:    name.Text,
				Author:  author.Text,
				Version: version.Text,
			})
		}, w.win)

	d.Resize(size)
	d.Show()
}

// saveProperties writes the metadata and reloads the window, so the title
// strip and the recents entry both catch up.
func (w *Window) saveProperties(p pack.Properties) {
	if err := pack.SetProperties(w.pack.Dir, p); err != nil {
		dialog.ShowError(err, w.win)
		return
	}

	reloaded, err := pack.Load(w.pack.Dir)
	if err != nil {
		dialog.ShowError(err, w.win)
		return
	}

	w.pack = reloaded
	w.deps.pack = reloaded

	if err := w.sess.Cfg.Touch(reloaded.Dir, reloaded.Name,
		reloaded.MCVersion, reloaded.Loader); err != nil {
		dialog.ShowError(err, w.win)
	}

	w.win.SetTitle(reloaded.Name)
	w.rebuildHeader()
	w.Reload()
}

// logoField is the image control of the properties dialog: a preview, the
// current filename, and buttons to change or remove it.
type logoField struct {
	win *Window
	row *fyne.Container
}

// build assembles the control for the pack's current state.
func (l *logoField) build() {
	if l.row == nil {
		l.row = container.NewStack()
	}

	path := pack.IconPath(l.win.pack.Dir)

	label := widgets.Note("No image set")
	if path != "" {
		label = widgets.Note(filepath.Base(path))
	}

	choose := widget.NewButtonWithIcon("Choose", fynetheme.FileImageIcon(), l.choose)

	buttons := container.NewHBox(choose)
	if path != "" {
		remove := widget.NewButton("Remove", l.remove)
		remove.Importance = widget.LowImportance
		buttons.Add(remove)
	}

	l.row.Objects = []fyne.CanvasObject{
		container.NewBorder(nil, nil,
			widgets.PackLogo(path, tokens.IconPackLogo), buttons, label),
	}
	l.row.Refresh()
}

// choose picks a new image and applies it at once, so the preview shows
// what was chosen rather than waiting for the dialog to be saved.
func (l *logoField) choose() {
	open := dialog.NewFileOpen(func(rc fyne.URIReadCloser, err error) {
		if err != nil || rc == nil {
			return
		}
		src := rc.URI().Path()
		rc.Close()

		if err := pack.SetIcon(l.win.pack.Dir, src); err != nil {
			dialog.ShowError(err, l.win.win)
			return
		}
		l.build()
		l.win.rebuildHeader()
	}, l.win.win)

	open.SetFilter(storage.NewExtensionFileFilter(pack.ImageExtensions))
	if dir, err := storage.ListerForURI(storage.NewFileURI(l.win.pack.Dir)); err == nil {
		open.SetLocation(dir)
	}
	open.Show()
}

// remove drops the pack's image.
func (l *logoField) remove() {
	if err := pack.RemoveIcon(l.win.pack.Dir); err != nil {
		dialog.ShowError(err, l.win.win)
		return
	}
	l.build()
	l.win.rebuildHeader()
}
