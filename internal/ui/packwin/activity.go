// Package packwin builds the pack window: an icon rail on the left, a
// list pane beside it, a detail area filling the rest, a collapsible log
// drawer along the bottom and a status bar under that.
package packwin

import (
	"fyne.io/fyne/v2"
	fynetheme "fyne.io/fyne/v2/theme"

	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/tokens"
	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/widgets"
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
	ActivityGit      = "git"
	ActivityReleases = "releases"
)

// placeholder is an activity whose screen has not been built yet. It
// keeps the rail complete while the sections land one at a time.
type placeholder struct {
	id    string
	title string
	icon  fyne.Resource
	note  string
}

func (p placeholder) ID() string          { return p.id }
func (p placeholder) Title() string       { return p.title }
func (p placeholder) Icon() fyne.Resource { return p.icon }

func (p placeholder) Side() fyne.CanvasObject { return nil }

func (p placeholder) Main() fyne.CanvasObject {
	return widgets.Inset(tokens.SpaceXL, tokens.SpaceLG, widgets.Muted(p.note))
}

// activities returns the rail's contents. Git and Releases are absent
// when the git integration is off, so nothing in the window offers an
// action it will not perform.
func (w *Window) activities() []Activity {
	list := []Activity{
		newModsActivity(w.deps),
		newBrowseActivity(w.deps),
		newFilesActivity(w.deps),
	}

	if w.gitEnabled() {
		list = append(list,
			placeholder{ActivityGit, "Git", fynetheme.StorageIcon(),
				"Staging, committing and pushing are not built yet."},
			placeholder{ActivityReleases, "Releases", fynetheme.UploadIcon(),
				"Release publishing is not built yet."},
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
