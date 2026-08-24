package launcher

import (
	"context"
	"errors"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	fynetheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/PalisadeMC/Packwiz-Studio/internal/git"
	"github.com/PalisadeMC/Packwiz-Studio/internal/pack"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/widgets"
)

// ShowClone clones a pack's repository and opens it.
//
// This is how someone else's pack gets onto the machine: most packs live
// in a git repository, and cloning then opening is otherwise two tools and
// a file manager.
func (w *Window) ShowClone() {
	url := widget.NewEntry()
	url.SetPlaceHolder("https://github.com/owner/pack.git")

	parent := widget.NewEntry()
	parent.SetText(defaultBrowseDir())

	folder := widget.NewEntry()
	folder.SetPlaceHolder("taken from the repository name")

	// Typing a URL fills in the folder name, which is almost always what
	// is wanted, while leaving it editable.
	url.OnChanged = func(text string) {
		if folder.Text == "" || folder.Text == folderFromURL(url.Text) {
			folder.SetText(folderFromURL(text))
		}
	}

	browse := widget.NewButtonWithIcon("", fynetheme.FolderOpenIcon(), func() {
		open := dialog.NewFolderOpen(func(list fyne.ListableURI, err error) {
			if err != nil || list == nil {
				return
			}
			parent.SetText(list.Path())
		}, w.win)

		if dir, err := storage.ListerForURI(storage.NewFileURI(parent.Text)); err == nil {
			open.SetLocation(dir)
		}
		open.Show()
	})

	body := container.NewVBox(
		widgets.Muted("Repository URL"),
		url,
		widgets.Note("An https or SSH URL. The app uses your existing git "+
			"credentials, so a private repository works if your key or "+
			"credential helper is already set up."),

		widgets.VSpace(tokens.SpaceSM),
		widgets.Muted("Clone into"),
		container.NewBorder(nil, nil, nil, browse, parent),
		widgets.Muted("Folder name"),
		folder,
	)

	size := widgets.FitDialog(w.win, tokens.FormWidth, tokens.CloneHeight)

	d := dialog.NewCustomConfirm("Clone a pack", "Clone", "Cancel",
		widgets.Scrollable(size.Width, size.Height,
			widgets.Inset(tokens.SpaceMD, tokens.SpaceMD, body)),
		func(clone bool) {
			if clone {
				w.startClone(url.Text, parent.Text, folder.Text)
			}
		}, w.win)

	d.Resize(size)
	d.Show()
}

// folderFromURL derives a folder name from a repository URL, which is
// what git itself would have picked.
func folderFromURL(raw string) string {
	trimmed := strings.TrimSuffix(strings.TrimRight(strings.TrimSpace(raw), "/"), ".git")
	if trimmed == "" {
		return ""
	}

	// Both URL and scp-style forms end with the repository name after the
	// last separator.
	if i := strings.LastIndexAny(trimmed, "/:"); i >= 0 {
		return trimmed[i+1:]
	}
	return trimmed
}

// errNoURL reports a clone with nothing to clone from.
var errNoURL = errors.New("enter a repository URL")

// startClone runs the clone, then opens the result if it holds a pack.
func (w *Window) startClone(rawURL, parent, name string) {
	if strings.TrimSpace(rawURL) == "" {
		dialog.ShowError(errNoURL, w.win)
		return
	}
	if strings.TrimSpace(name) == "" {
		name = folderFromURL(rawURL)
	}
	if strings.TrimSpace(name) == "" {
		dialog.ShowError(errors.New("enter a folder name for the clone"), w.win)
		return
	}

	dest := filepath.Join(parent, name)

	progress := dialog.NewCustomWithoutButtons("Cloning",
		widgets.Inset(tokens.SpaceMD, tokens.SpaceMD, container.NewVBox(
			widgets.Muted("Cloning into "+dest),
			widget.NewProgressBarInfinite(),
			widgets.Note("The output panel of a pack window shows git's own progress."),
		)), w.win)
	progress.Show()

	go func() {
		res, err := git.Clone(context.Background(), w.sess.Runner, rawURL, dest)

		fyne.Do(func() {
			progress.Hide()

			switch {
			case err != nil:
				dialog.ShowError(err, w.win)
			case !res.OK():
				dialog.ShowError(cloneFailed(res.Output()), w.win)
			default:
				w.finishClone(dest)
			}
		})
	}()
}

// finishClone records the cloned pack and opens it.
func (w *Window) finishClone(dir string) {
	p, err := pack.Load(dir)
	if err != nil {
		// The clone worked but there is no pack in it. That is worth
		// saying plainly rather than opening an empty window.
		dialog.ShowError(errors.New("cloned into "+dir+
			", but it holds no pack.toml, so it is not a packwiz pack"), w.win)
		w.Refresh()
		return
	}

	if err := w.sess.Cfg.Touch(p.Dir, p.Name, p.MCVersion, p.Loader); err != nil {
		dialog.ShowError(err, w.win)
		return
	}
	w.openPack(p.Dir)
}

// cloneFailed turns git's complaint into an error for a dialog.
func cloneFailed(output string) error {
	if msg := strings.TrimSpace(output); msg != "" {
		return errors.New("git clone failed: " + msg)
	}
	return errors.New("git clone failed, see the output panel")
}
