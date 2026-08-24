package packwin

import (
	"context"

	"fyne.io/fyne/v2"
)

// ToggleSidePanel hides or shows the list pane, widening the detail area.
func (w *Window) ToggleSidePanel() { w.side.toggle() }

// ToggleLog opens or closes the output drawer.
func (w *Window) ToggleLog() { w.toggleLog() }

func (w *Window) toggleLog() {
	w.drawer.toggle()
	w.status.setLogOpen(w.drawer.visible())
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
