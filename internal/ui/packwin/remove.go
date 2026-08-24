package packwin

import (
	"context"

	"fyne.io/fyne/v2/dialog"

	"github.com/PalisadeMC/Packwiz-Studio/internal/cmdrun"
	"github.com/PalisadeMC/Packwiz-Studio/internal/pack"
)

// confirmRemoveMod asks before deleting a mod, then deletes it.
//
// Removal is offered from the mod list, the mod's detail pane and the
// browser, so the confirmation and the command live here rather than
// three times over. It is always confirmed: removal rewrites the pack and
// the app has no undo for it.
func confirmRemoveMod(deps *activityDeps, m pack.Mod) {
	dialog.NewConfirm("Remove mod",
		"Remove "+m.Name+" from this pack?",
		func(ok bool) {
			if !ok {
				return
			}
			deps.run("remove "+m.Slug(), func(ctx context.Context) error {
				return exec(func() (cmdrun.Result, error) {
					return deps.client().Remove(ctx, m.Slug())
				})
			})
		}, deps.win).Show()
}
