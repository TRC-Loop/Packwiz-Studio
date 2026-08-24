package packwin

import (
	"context"

	"fyne.io/fyne/v2"

	"github.com/PalisadeMC/Packwiz-Studio/internal/config"
)

// ToggleSidePanel hides or shows the list pane, widening the detail area.
func (w *Window) ToggleSidePanel() { w.side.toggle() }

// ToggleLog opens or closes the output drawer.
func (w *Window) ToggleLog() { w.toggleLog() }

func (w *Window) toggleLog() {
	open := !w.drawer.open

	// The height the divider was left at is kept, so reopening the drawer
	// restores the size it was dragged to.
	if !open && w.split != nil {
		w.deps.setPrefs(func(p *config.Prefs) { p.DrawerOffset = w.split.Offset })
	}

	w.drawer.setOpen(open)
	w.setDrawerOpen(open)
	w.status.setLogOpen(open)
}

// RefreshStatus re-reads the tool and repository state.
//
// The git probe runs off the main goroutine because it shells out several
// times, which is too slow to do while laying out a window.
func (w *Window) RefreshStatus() {
	loc := w.sess.Packwiz()
	w.status.setPackwiz(loc.Version, loc.Path != "")

	if !w.gitEnabled() {
		return
	}

	go func() {
		st := w.repo.Read(context.Background())
		fyne.Do(func() { w.status.setGit(st) })
	}()
}
