// Package packwin builds the pack window: an icon rail on the left, a
// list pane beside it, a detail area filling the rest, a collapsible log
// drawer along the bottom and a status bar under that.
package packwin

import (
	"fyne.io/fyne/v2"
)

// Activity is one section of the pack window, reached from the icon rail.
// Each supplies the list pane and the detail area; either may be nil, in
// which case that pane is left empty.
type Activity interface {
	// ID identifies the activity for the config and the menu.
	ID() string
	// Title names it in the side panel header and the View menu.
	Title() string
	// Icon is its glyph in the rail, from Fyne's built-in set.
	Icon() fyne.Resource
	// Side is the list pane content, or nil for an activity that needs
	// the whole window.
	Side() fyne.CanvasObject
	// Main is the detail area content.
	Main() fyne.CanvasObject
}

// Activity identifiers. These are stable strings: they end up in config
// as the last-used activity.
const (
	ActivityMods     = "mods"
	ActivityBrowse   = "browse"
	ActivityFiles    = "files"
	ActivityInstance = "instance"
	ActivityGit      = "git"
	ActivityReleases = "releases"
)

// activities returns the rail's contents. Git and Releases are absent
// when the git integration is off, so nothing in the window offers an
// action it will not perform.
func (w *Window) activities() []Activity {
	list := []Activity{
		newModsActivity(w.deps),
		newBrowseActivity(w.deps),
		newFilesActivity(w.deps),
		newInstanceActivity(w.deps),
	}

	if w.gitEnabled() {
		list = append(list,
			newGitActivity(w.deps, w.repo),
			newReleasesActivity(w.deps, w.repo),
		)
	}

	return list
}

// gitEnabled reports whether git features should appear. Both the setting
// and an actual git binary are required: offering git actions that cannot
// run would be worse than hiding them.
func (w *Window) gitEnabled() bool {
	return w.sess.GitEnabled() && w.gitAvailable
}
