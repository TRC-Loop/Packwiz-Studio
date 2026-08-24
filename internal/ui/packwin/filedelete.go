package packwin

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2/dialog"

	"github.com/PalisadeMC/Packwiz-Studio/internal/cmdrun"
	"github.com/PalisadeMC/Packwiz-Studio/internal/pack"
)

// promptDelete removes the selection, after asking.
//
// A mod's metafile is removed through packwiz rather than by deleting the
// file, so that the index and the pack agree afterwards.
func (a *filesActivity) promptDelete() {
	if a.selected == "" {
		dialog.ShowInformation("Delete", "Select a file or folder first.", a.deps.win)
		return
	}

	target := a.selected
	if mod := a.modAt(target); mod != nil {
		confirmRemoveMod(a.deps, *mod)
		return
	}

	dialog.NewConfirm("Delete", "Delete "+target+" from the pack folder?",
		func(ok bool) {
			if !ok {
				return
			}
			a.remove(target)
		}, a.deps.win).Show()
}

// remove deletes a path and refreshes the index.
func (a *filesActivity) remove(rel string) {
	a.selected = ""

	a.deps.run("delete "+rel, func(ctx context.Context) error {
		full := filepath.Join(a.deps.pack.Dir, filepath.FromSlash(rel))
		if err := os.RemoveAll(full); err != nil {
			return err
		}

		return exec(func() (cmdrun.Result, error) {
			return a.deps.client().Refresh(ctx)
		})
	})
}

// modAt reports the mod a metafile path belongs to, so that deleting one
// goes through packwiz rather than removing the file behind its back.
func (a *filesActivity) modAt(rel string) *pack.Mod {
	if !strings.HasSuffix(rel, ".pw.toml") {
		return nil
	}

	mods, err := a.deps.pack.Mods()
	if err != nil {
		return nil
	}
	for i := range mods {
		if mods[i].Path == rel {
			return &mods[i]
		}
	}
	return nil
}

// base is the folder a new file goes in: the selected folder, or the
// folder holding the selected file.
func (a *filesActivity) base() string {
	if a.selected == "" {
		return ""
	}
	if a.scan.IsDir(a.selected) {
		return a.selected
	}

	parent := path.Dir(a.selected)
	if parent == "." {
		return ""
	}
	return parent
}

// baseLabel names the base folder for the prompt.
func (a *filesActivity) baseLabel() string {
	if base := a.base(); base != "" {
		return base
	}
	return "the pack folder"
}
