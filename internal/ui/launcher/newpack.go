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
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/widgets"
)

// newPackForm holds the widgets of the new pack screen.
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

	// ignore writes the two ignore files once the pack exists. It is on by
	// default: a pack repository without them commits editor clutter and
	// ships its own README into everybody's game folder.
	ignore *widget.Check

	status *widget.Label

	// logo is the image chosen for the pack, copied in after packwiz has
	// created the folder. Empty means the pack starts without one.
	logo      string
	logoLabel *widget.Label

	// games holds every Minecraft version once fetched, so toggling
	// snapshots refilters without another request.
	games []mcmeta.Version
	// loaded records that the version lists have been requested, so
	// returning to the screen does not refetch them.
	loaded bool
}

// ShowNewPack opens the pack creation screen.
func (w *Window) ShowNewPack() {
	if w.newPack == nil {
		w.newPack = newForm(w)
	}
	w.show(screenNewPack)

	// Fetching starts after the screen is up, so it appears at once and
	// fills itself in.
	if !w.newPack.loaded {
		w.newPack.loaded = true
		w.newPack.loadGameVersions()
	}
}

// newForm builds the form's widgets.
func newForm(w *Window) *newPackForm {
	f := &newPackForm{
		win:       w,
		dir:       widget.NewEntry(),
		name:      widget.NewEntry(),
		author:    widget.NewEntry(),
		version:   widget.NewEntry(),
		status:    widgets.Note(""),
		logoLabel: widgets.Note(noLogoLabel),
	}

	f.dir.SetPlaceHolder("an empty folder for the pack")
	f.version.SetText("1.0.0")

	f.ignore = widget.NewCheck("Add recommended ignore rules", nil)
	f.ignore.SetChecked(true)

	f.buildVersionControls()

	return f
}

// buildNewPack lays the form out as a launcher screen: a heading, the
// fields in a scrolling middle, and the actions pinned to the bottom so
// they stay reachable however tall the form gets.
func (w *Window) buildNewPack() fyne.CanvasObject {
	f := w.newPack

	browse := widget.NewButtonWithIcon("", fynetheme.FolderOpenIcon(), f.pickFolder)

	fields := container.NewVBox(
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
		widgets.Muted("Pack image"),
		container.NewBorder(nil, nil, nil,
			container.NewHBox(
				widget.NewButtonWithIcon("Choose", fynetheme.FileImageIcon(), f.pickLogo),
				widget.NewButton("Clear", f.clearLogo),
			),
			f.logoLabel,
		),
		widgets.Note("Optional. Saved into the pack folder as icon.png, so it "+
			"travels with the pack and can be attached to a release."),

		widgets.VSpace(tokens.SpaceSM),
		widgets.Muted("Ignore rules"),
		f.ignore,
		widgets.Note("Writes .gitignore and .packwizignore: OS and editor "+
			"clutter such as .DS_Store kept out of the repository, and the "+
			"repository's own files kept out of the exported pack. Both can "+
			"be edited later from the Pack menu."),
	)

	create := widget.NewButtonWithIcon("Create pack", fynetheme.ConfirmIcon(),
		func() { w.createPack(f) })
	create.Importance = widget.HighImportance

	cancel := widget.NewButton("Cancel", func() { w.show(screenRecents) })
	cancel.Importance = widget.LowImportance

	actions := container.NewVBox(
		f.status,
		container.NewHBox(cancel, create),
	)

	return container.NewBorder(
		widgets.Inset(tokens.SpaceXL, tokens.SpaceMD, widgets.Heading("New pack")),
		widgets.Inset(tokens.SpaceXL, tokens.SpaceMD, actions),
		nil, nil,
		container.NewVScroll(widgets.Inset(tokens.SpaceXL, tokens.SpaceSM, fields)),
	)
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
