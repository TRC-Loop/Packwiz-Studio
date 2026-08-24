package packwin

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	"github.com/PalisadeMC/Packwiz-Studio/internal/config"
	"github.com/PalisadeMC/Packwiz-Studio/internal/git"
	"github.com/PalisadeMC/Packwiz-Studio/internal/pack"
	"github.com/PalisadeMC/Packwiz-Studio/internal/studio"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
)

// Window is a pack open for editing.
type Window struct {
	sess *studio.Session
	win  fyne.Window
	pack pack.Pack
	repo *git.Repo

	// gitAvailable records whether a git binary exists. Combined with the
	// setting it decides whether git features appear at all.
	gitAvailable bool

	// hadGit is what gitEnabled reported when the window was last built,
	// so a settings change can be recognised as one that reshapes it.
	hadGit bool

	// onClose hands the window back to whatever opened this pack.
	onClose func()

	// onMenuChanged asks the shell to reinstall the menubar.
	onMenuChanged func()

	// deps is what the activities share.
	deps *activityDeps

	head   *fyne.Container
	rail   *activityBar
	side   *sidePanel
	main   *fyne.Container
	drawer *logDrawer
	status *statusBar
	root   *fyne.Container

	current string
	items   []Activity
}

// New builds a pack window for p inside win. Closing it calls onClose,
// which is how the launcher gets its window back.
func New(sess *studio.Session, win fyne.Window, p pack.Pack, onClose func()) *Window {
	w := &Window{
		sess:         sess,
		win:          win,
		pack:         p,
		repo:         git.New(p.Dir, sess.Runner),
		gitAvailable: git.Available(),
		onClose:      onClose,
		current:      ActivityMods,
	}

	// Reopening a pack lands on the section it was left on, unless that
	// section is not available any more.
	if last := sess.Cfg.Prefs(p.Dir).Activity; last != "" {
		w.current = last
	}

	w.deps = &activityDeps{
		sess:          sess,
		win:           win,
		pack:          p,
		onPackChanged: w.Reload,
		onMenuChanged: w.rebuildMenu,
		addActions:    w.addMenu,
	}

	w.items = w.activities()
	w.head = container.NewStack(w.header())
	w.drawer = newLogDrawer(sess.Bus)
	w.status = newStatusBar(w.gitEnabled(), w.toggleLog)
	w.side = newSidePanel()
	w.main = container.NewStack()
	w.rail = newActivityBar(w.items, w.current, w.selectActivity)

	w.root = w.build()
	w.hadGit = w.gitEnabled()

	// A remembered section may no longer exist, for instance a git one
	// after the integration was turned off.
	if !w.has(w.current) {
		w.current = ActivityMods
	}

	w.selectActivity(w.current)
	w.watchConfig()

	return w
}

// Install puts the pack window into its window and starts its refreshes.
func (w *Window) Install() {
	w.win.SetTitle(w.pack.Name)
	w.win.SetContent(w.root)
	w.win.Resize(fyne.NewSize(tokens.PackWindowWidth, tokens.PackWindowHeight))
	w.RefreshStatus()
}

// Detach stops the window's listeners. It is separate from Close so a
// shell can retire a pack view without triggering the close handler that
// would put the launcher back.
func (w *Window) Detach() {
	w.drawer.detach()
}

// Close detaches the window's listeners and hands control back to
// whatever opened this pack.
func (w *Window) Close() {
	w.Detach()
	if w.onClose != nil {
		w.onClose()
	}
}

// Pack reports the pack this window edits.
func (w *Window) Pack() pack.Pack { return w.pack }

// Window exposes the underlying window.
func (w *Window) Window() fyne.Window { return w.win }

// Activities reports the rail's contents, for the View menu.
func (w *Window) Activities() []Activity { return w.items }

// Current reports the open activity's identifier.
func (w *Window) Current() string { return w.current }

// Select opens the activity with the given identifier, which is how the
// View menu reaches the icon rail's sections by name.
func (w *Window) Select(id string) { w.selectActivity(id) }

// SetOnMenuChanged registers a callback for when the menubar's contents
// need rebuilding, such as after the mod selection changed.
func (w *Window) SetOnMenuChanged(fn func()) { w.onMenuChanged = fn }

// rebuildMenu asks the shell to reinstall the menubar.
func (w *Window) rebuildMenu() {
	if w.onMenuChanged != nil {
		w.onMenuChanged()
	}
}

// selectActivity swaps the side panel and detail area.
func (w *Window) selectActivity(id string) {
	for _, a := range w.items {
		if a.ID() != id {
			continue
		}

		if w.current != id {
			w.deps.setPrefs(func(p *config.Prefs) { p.Activity = id })
		}

		w.current = id
		w.rail.setCurrent(id)
		w.side.set(a.Title(), a.Side())

		w.main.Objects = nil
		if main := a.Main(); main != nil {
			w.main.Objects = []fyne.CanvasObject{main}
		}
		w.main.Refresh()
		return
	}
}
