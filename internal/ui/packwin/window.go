package packwin

import (
	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	"github.com/TRC-Loop/Packwiz-Studio/internal/git"
	"github.com/TRC-Loop/Packwiz-Studio/internal/pack"
	"github.com/TRC-Loop/Packwiz-Studio/internal/studio"
	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/tokens"
	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/widgets"
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

	// onClose hands the window back to whatever opened this pack.
	onClose func()

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

	w.items = w.activities()
	w.head = container.NewStack(w.header())
	w.drawer = newLogDrawer(sess.Bus)
	w.status = newStatusBar(w.gitEnabled(), w.toggleLog)
	w.side = newSidePanel()
	w.main = container.NewStack()
	w.rail = newActivityBar(w.items, w.current, w.selectActivity)

	w.root = w.build()
	w.selectActivity(w.current)

	return w
}

// build assembles the window's panes.
func (w *Window) build() *fyne.Container {
	bg := canvas.NewRectangle(tokens.ColorBG)
	body := container.NewStack(bg, w.main)

	// The side panel keeps its width while the detail area takes the
	// slack, so widening the window widens the content rather than the
	// list.
	panes := container.NewBorder(nil, nil, w.side.object(), nil, body)

	foot := container.NewVBox(w.drawer.object(), w.status.object())

	return container.NewBorder(
		w.head, foot, w.rail.object(), nil,
		panes,
	)
}

// header names the pack and shows its logo.
func (w *Window) header() fyne.CanvasObject {
	bg := canvas.NewRectangle(tokens.ColorSurface)

	title := container.NewHBox(
		widgets.PackLogo(pack.IconPath(w.pack.Dir), tokens.IconPackLogo),
		container.NewCenter(widgets.SubHeading(w.pack.Name)),
	)

	line := container.NewBorder(nil, nil,
		widgets.Inset(tokens.SpaceMD, tokens.SpaceSM, title), nil, nil)

	return container.NewStack(bg,
		container.NewBorder(nil, widgets.Hairline(), nil, nil, line))
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

// selectActivity swaps the side panel and detail area.
func (w *Window) selectActivity(id string) {
	for _, a := range w.items {
		if a.ID() != id {
			continue
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

// ToggleSidePanel hides or shows the list pane.
func (w *Window) ToggleSidePanel() { w.side.toggle() }

// ToggleLog opens or closes the output drawer.
func (w *Window) ToggleLog() { w.toggleLog() }

func (w *Window) toggleLog() {
	w.drawer.toggle()
	w.status.setLogOpen(w.drawer.visible())
}

// RefreshStatus re-reads the tool and repository state. The git probe runs
// off the main goroutine because it shells out several times.
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
