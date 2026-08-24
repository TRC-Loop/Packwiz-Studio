package packwin

import (
	"context"
	"errors"
	"strconv"

	"fyne.io/fyne/v2/dialog"

	"github.com/PalisadeMC/Packwiz-Studio/internal/cmdrun"
	"github.com/PalisadeMC/Packwiz-Studio/internal/instance"
)

// confirmImport copies whole folders in, which is how a pack that was
// built by playing gets its configs in the first place.
//
// It replaces the pack's copy of anything in both, so it is the blunt
// version of the file by file sync and is confirmed accordingly.
func (a *instanceActivity) confirmImport() {
	dir := a.path.Text
	folders := a.chosen()

	if !instance.Exists(dir) {
		dialog.ShowError(errors.New("that folder is not there"), a.deps.win)
		return
	}
	if len(folders) == 0 {
		dialog.ShowError(errors.New("tick at least one folder to import"), a.deps.win)
		return
	}

	dialog.NewConfirm("Import from the instance",
		"Copy everything in "+joinFolders(folders)+" from the instance into "+
			"the pack, replacing the pack's copy of any file that is in both?",
		func(ok bool) {
			if !ok {
				return
			}
			a.runImport(dir, folders)
		}, a.deps.win).Show()
}

// runImport does the copying and refreshes the index, since the pack has
// gained files packwiz does not know about yet.
func (a *instanceActivity) runImport(dir string, folders []string) {
	a.deps.run("import from the instance", func(ctx context.Context) error {
		count, err := instance.Import(dir, a.deps.pack.Dir, folders)
		if err != nil {
			return err
		}
		a.deps.notice("imported " + strconv.Itoa(count) + " files")

		return exec(func() (cmdrun.Result, error) {
			return a.deps.client().Refresh(ctx)
		})
	})
}
