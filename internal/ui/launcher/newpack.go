package launcher

import (
	"context"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	fynetheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/TRC-Loop/Packwiz-Studio/internal/pack"
	"github.com/TRC-Loop/Packwiz-Studio/internal/packwiz"
	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/tokens"
	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/widgets"
)

// newPackForm holds the widgets of the new pack dialog, so validation and
// submission can read them without a closure per field.
type newPackForm struct {
	dir           *widget.Entry
	name          *widget.Entry
	author        *widget.Entry
	version       *widget.Entry
	mcVersion     *widget.Entry
	loader        *widget.Select
	loaderVersion *widget.Entry
	problem       *widget.Label
}

// showNewPack collects everything `packwiz init` asks for and runs it.
func (w *Window) ShowNewPack() {
	f := &newPackForm{
		dir:           widget.NewEntry(),
		name:          widget.NewEntry(),
		author:        widget.NewEntry(),
		version:       widget.NewEntry(),
		mcVersion:     widget.NewEntry(),
		loaderVersion: widget.NewEntry(),
		problem:       widget.NewLabel(""),
	}

	f.dir.SetPlaceHolder("an empty folder for the pack")
	f.version.SetText("1.0.0")
	f.mcVersion.SetPlaceHolder("1.20.1")
	f.loaderVersion.SetPlaceHolder("leave empty for the latest")

	labels := make([]string, 0, len(packwiz.Loaders))
	for _, l := range packwiz.Loaders {
		labels = append(labels, l.Label())
	}
	f.loader = widget.NewSelect(labels, nil)
	f.loader.SetSelectedIndex(0)

	f.problem.Importance = widget.DangerImportance
	f.problem.Wrapping = fyne.TextWrapWord
	f.problem.Hide()

	browse := widget.NewButtonWithIcon("", fynetheme.FolderOpenIcon(), func() {
		open := dialog.NewFolderOpen(func(list fyne.ListableURI, err error) {
			if err != nil || list == nil {
				return
			}
			f.dir.SetText(list.Path())
			if f.name.Text == "" {
				f.name.SetText(filepath.Base(list.Path()))
			}
		}, w.win)
		if dir, err := storage.ListerForURI(storage.NewFileURI(defaultBrowseDir())); err == nil {
			open.SetLocation(dir)
		}
		open.Show()
	})

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

		widgets.Muted("Minecraft version"),
		f.mcVersion,
		widgets.Muted("Mod loader"),
		f.loader,
		widgets.Muted("Loader version"),
		f.loaderVersion,

		widgets.VSpace(tokens.SpaceSM),
		f.problem,
	)

	d := dialog.NewCustomConfirm("New pack", "Create", "Cancel",
		container.NewVScroll(widgets.Inset(tokens.SpaceMD, tokens.SpaceMD, body)),
		func(create bool) {
			if create {
				w.createPack(f)
			}
		}, w.win)

	d.Resize(fyne.NewSize(tokens.FormWidth, tokens.FormHeight))
	d.Show()
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
		MCVersion:     f.mcVersion.Text,
		Loader:        loader,
		LoaderVersion: f.loaderVersion.Text,
	}
}

// createPack validates the form, then runs packwiz init and opens the
// result. Validation happens before anything is run, so a bad form never
// reaches the command line.
func (w *Window) createPack(f *newPackForm) {
	opts := f.options()
	if err := opts.Validate(); err != nil {
		dialog.ShowError(err, w.win)
		return
	}
	if pack.IsPack(opts.Dir) {
		dialog.ShowError(errAlreadyAPack, w.win)
		return
	}

	go func() {
		res, err := w.sess.Client(opts.Dir).Init(context.Background(), opts)

		fyne.Do(func() {
			if err != nil {
				dialog.ShowError(err, w.win)
				return
			}
			if !res.OK() {
				dialog.ShowError(initFailed(res.Output()), w.win)
				return
			}
			w.finishNewPack(opts.Dir)
		})
	}()
}

// finishNewPack records the new pack and opens it.
func (w *Window) finishNewPack(dir string) {
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
}
