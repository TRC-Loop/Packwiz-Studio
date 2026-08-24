package packwin

import (
	"context"
	"errors"
	"os"
	"path"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/PalisadeMC/Packwiz-Studio/internal/cmdrun"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
)

// promptNewFile asks for a name and creates an empty file.
//
// A new file goes in the selected folder, or beside the selected file,
// which is what a file tree is expected to do.
func (a *filesActivity) promptNewFile() {
	a.promptName("New file", "config/example.json", func(name string) {
		a.create(name, false)
	})
}

// promptNewFolder asks for a name and creates a folder.
func (a *filesActivity) promptNewFolder() {
	a.promptName("New folder", "kubejs/server_scripts", func(name string) {
		a.create(name, true)
	})
}

// promptName is the shared one field prompt. A name may carry slashes, so
// a whole path can be made in one go rather than a folder at a time.
func (a *filesActivity) promptName(title, hint string, done func(string)) {
	entry := widget.NewEntry()
	entry.SetPlaceHolder(hint)

	d := dialog.NewForm(title, "Create", "Cancel",
		[]*widget.FormItem{
			{Text: "In", Widget: widget.NewLabel(a.baseLabel())},
			{Text: "Name", Widget: entry},
		},
		func(ok bool) {
			if !ok {
				return
			}
			if name := strings.TrimSpace(entry.Text); name != "" {
				done(name)
			}
		}, a.deps.win)

	d.Resize(fyne.NewSize(tokens.FormWidth, d.MinSize().Height))
	d.Show()
}

// create makes a file or a folder under the current base folder.
func (a *filesActivity) create(name string, dir bool) {
	rel := path.Join(a.base(), filepath.ToSlash(name))
	full := filepath.Join(a.deps.pack.Dir, filepath.FromSlash(rel))

	if _, err := os.Stat(full); err == nil {
		dialog.ShowError(errors.New(rel+" already exists"), a.deps.win)
		return
	}

	if !dir {
		a.selected = rel
	}

	a.deps.run("create "+rel, func(ctx context.Context) error {
		if dir {
			return os.MkdirAll(full, 0o755)
		}
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(full, nil, 0o644); err != nil {
			return err
		}

		return exec(func() (cmdrun.Result, error) {
			return a.deps.client().Refresh(ctx)
		})
	})
}

// promptRename renames whatever is selected.
func (a *filesActivity) promptRename() {
	if a.selected == "" {
		dialog.ShowInformation("Rename", "Select a file or folder first.", a.deps.win)
		return
	}

	entry := widget.NewEntry()
	entry.SetText(path.Base(a.selected))

	d := dialog.NewForm("Rename", "Rename", "Cancel",
		[]*widget.FormItem{{Text: "Name", Widget: entry}},
		func(ok bool) {
			if !ok {
				return
			}
			name := strings.TrimSpace(entry.Text)
			if name == "" || name == path.Base(a.selected) {
				return
			}
			a.rename(name)
		}, a.deps.win)

	d.Resize(fyne.NewSize(tokens.FormWidth, d.MinSize().Height))
	d.Show()
}

// rename moves the selection within its own folder.
func (a *filesActivity) rename(name string) {
	from := a.selected
	parent := path.Dir(from)
	if parent == "." {
		parent = ""
	}
	to := path.Join(parent, filepath.ToSlash(name))

	a.selected = to

	a.deps.run("rename "+from, func(ctx context.Context) error {
		err := os.Rename(
			filepath.Join(a.deps.pack.Dir, filepath.FromSlash(from)),
			filepath.Join(a.deps.pack.Dir, filepath.FromSlash(to)))
		if err != nil {
			return err
		}

		return exec(func() (cmdrun.Result, error) {
			return a.deps.client().Refresh(ctx)
		})
	})
}
