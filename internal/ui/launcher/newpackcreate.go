package launcher

import (
	"context"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"

	"github.com/PalisadeMC/Packwiz-Studio/internal/pack"
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

			// The image is copied in after init, since the pack folder
			// only exists once packwiz has written it. A failure here is
			// reported without discarding the pack that was created.
			if f.logo != "" {
				if err := pack.SetIcon(opts.Dir, f.logo); err != nil {
					dialog.ShowError(err, w.win)
				}
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
