package packwin

import (
	"fyne.io/fyne/v2"
)

// watchConfig rebuilds the window when settings change.
//
// Turning the git integration off has to remove the Git and Releases
// sections from an already open window, and turning it on has to add
// them, so the rail and the activities are rebuilt rather than left as
// they were when the window opened.
func (w *Window) watchConfig() {
	w.sess.OnConfigChange(func() {
		fyne.Do(func() {
			if w.gitEnabled() == w.hadGit {
				// Nothing that shapes this window changed, so restating
				// the status bar is enough.
				w.RefreshStatus()
				return
			}
			w.rebuildShell()
		})
	})
}

// rebuildShell recreates the activities, the rail and the status bar for
// the current settings, keeping the open section where it still exists.
func (w *Window) rebuildShell() {
	w.hadGit = w.gitEnabled()
	previous := w.current

	w.items = w.activities()
	w.status = newStatusBar(w.gitEnabled(), w.toggleLog)
	w.rail = newActivityBar(w.items, previous, w.selectActivity)

	w.root = w.build()
	w.win.SetContent(w.root)

	// The section that was open may have just been removed, in which case
	// the window falls back to the mods list rather than showing nothing.
	if !w.has(previous) {
		previous = ActivityMods
	}
	w.selectActivity(previous)

	w.RefreshStatus()
	w.rebuildMenu()
}

// has reports whether an activity with this identifier is present.
func (w *Window) has(id string) bool {
	for _, a := range w.items {
		if a.ID() == id {
			return true
		}
	}
	return false
}
