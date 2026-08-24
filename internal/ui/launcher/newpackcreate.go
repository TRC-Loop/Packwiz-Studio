package launcher

import (
	"context"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"

	"github.com/PalisadeMC/Packwiz-Studio/internal/pack"
	"github.com/PalisadeMC/Packwiz-Studio/internal/packwiz"
)

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

			// The image and the ignore files are written after init, since
			// the pack folder only exists once packwiz has made it. A
			// failure in either is reported without discarding the pack
			// that was created.
			if f.logo != "" {
				if err := pack.SetIcon(opts.Dir, f.logo); err != nil {
					dialog.ShowError(err, w.win)
				}
			}
			if f.ignore.Checked {
				if err := writeIgnoreRules(opts.Dir); err != nil {
					dialog.ShowError(err, w.win)
				}
			}

			w.finishNewPack(opts.Dir)
		})
	}()
}

// writeIgnoreRules gives a new pack its two ignore files.
//
// Both are written from scratch here rather than merged: the folder was
// empty a moment ago, and anything already in them came from packwiz
// itself, which writes neither.
func writeIgnoreRules(dir string) error {
	if err := pack.WriteIgnore(dir, pack.GitIgnoreFile,
		pack.AddRules("", pack.DefaultGitIgnore())); err != nil {
		return err
	}
	return pack.WriteIgnore(dir, pack.PackwizIgnoreFile,
		pack.AddRules("", pack.DefaultPackwizIgnore()))
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

	// The form is discarded once its pack exists, so opening the screen
	// again starts blank rather than showing the pack just created.
	w.newPack = nil
	w.screen = screenRecents

	w.openPack(p.Dir)
}

// plural renders a count with its noun.
func plural(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return strconv.Itoa(n) + " " + noun + "s"
}
