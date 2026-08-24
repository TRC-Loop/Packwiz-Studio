package launcher

import (
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	fynetheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/PalisadeMC/Packwiz-Studio/internal/mcmeta"
	"github.com/PalisadeMC/Packwiz-Studio/internal/packwiz"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/widgets"
)

// newPackForm holds the widgets of the new pack dialog.
//
// The three version fields are dependent: a loader is chosen first, then a
// Minecraft version, then a build of that loader for that version. Each
// list is fetched from the loader's own metadata service, so the form
// offers what actually exists rather than trusting a typed version.
type newPackForm struct {
	win *Window

	dir     *widget.Entry
	name    *widget.Entry
	author  *widget.Entry
	version *widget.Entry

	loader        *widget.Select
	mcVersion     *widget.Select
	snapshots     *widget.Check
	loaderVersion *widget.Select
	loaderManual  *widget.Entry

	status *widget.Label

	// games holds every Minecraft version once fetched, so toggling
	// snapshots refilters without another request.
	games []mcmeta.Version
}

// ShowNewPack collects everything packwiz init asks for and runs it.
func (w *Window) ShowNewPack() {
	f := &newPackForm{
		win:     w,
		dir:     widget.NewEntry(),
		name:    widget.NewEntry(),
		author:  widget.NewEntry(),
		version: widget.NewEntry(),
		status:  widgets.Note(""),
	}

	f.dir.SetPlaceHolder("an empty folder for the pack")
	f.version.SetText("1.0.0")

	f.buildVersionControls()

	browse := widget.NewButtonWithIcon("", fynetheme.FolderOpenIcon(), f.pickFolder)

	body := container.NewVBox(
		widgets.Muted("Folder"),
		container.NewBorder(nil, nil, nil, browse, f.dir),

		widgets.VSpace(tokens.SpaceSM),
		widgets.Muted("Pack name"),
		f.name,
		widgets.Muted("Author"),
		f.author,
		widgets.Muted("Pack version"),
		f.version,

		widgets.VSpace(tokens.SpaceSM),
		widgets.Muted("Mod loader"),
		f.loader,

		widgets.Muted("Minecraft version"),
		f.mcVersion,
		f.snapshots,

		widgets.Muted("Loader version"),
		f.loaderVersion,
		f.loaderManual,

		widgets.VSpace(tokens.SpaceSM),
		f.status,
	)

	size := widgets.FitDialog(w.win, tokens.FormWidth, tokens.FormHeight)

	d := dialog.NewCustomConfirm("New pack", "Create", "Cancel",
		widgets.Scrollable(size.Width, size.Height,
			widgets.Inset(tokens.SpaceMD, tokens.SpaceMD, body)),
		func(create bool) {
			if create {
				w.createPack(f)
			}
		}, w.win)

	d.Resize(size)
	d.Show()

	// Fetching starts after the dialog is up, so the form appears at once
	// and fills itself in.
	f.loadGameVersions()
}

// pickFolder chooses the pack folder, defaulting the pack name to it.
func (f *newPackForm) pickFolder() {
	open := dialog.NewFolderOpen(func(list fyne.ListableURI, err error) {
		if err != nil || list == nil {
			return
		}
		f.dir.SetText(list.Path())
		if f.name.Text == "" {
			f.name.SetText(filepath.Base(list.Path()))
		}
	}, f.win.win)

	if dir, err := storage.ListerForURI(storage.NewFileURI(defaultBrowseDir())); err == nil {
		open.SetLocation(dir)
	}
	open.Show()
}

// options turns the form into init options.
func (f *newPackForm) options() packwiz.InitOptions {
	loader := packwiz.LoaderFabric
	if i := f.loader.SelectedIndex(); i >= 0 && i < len(packwiz.Loaders) {
		loader = packwiz.Loaders[i]
	}

	return packwiz.InitOptions{
		Dir:           f.dir.Text,
		Name:          f.name.Text,
		Author:        f.author.Text,
		Version:       f.version.Text,
		MCVersion:     f.mcVersion.Selected,
		Loader:        loader,
		LoaderVersion: f.chosenLoaderVersion(),
	}
}

// chosenLoaderVersion reads whichever loader version control is in use.
// An empty result means latest, which is what packwiz defaults to.
func (f *newPackForm) chosenLoaderVersion() string {
	if f.loaderManual.Visible() {
		return f.loaderManual.Text
	}
	if f.loaderVersion.Selected == latestLabel {
		return ""
	}
	return f.loaderVersion.Selected
}
