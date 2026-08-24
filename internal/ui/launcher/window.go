// Package launcher builds the window the app opens with: a list of known
// packs on the left, actions on the right, and a packwiz status line at
// the foot.
//
// When packwiz has not been resolved the same window shows a setup screen
// instead. The setup screen is dismissible: the launcher then works with
// its pack actions disabled, so a user can still reach Settings and fix
// the path there.
package launcher

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	"github.com/TRC-Loop/Packwiz-Studio/internal/packwiz"
	"github.com/TRC-Loop/Packwiz-Studio/internal/studio"
	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/tokens"
)

// Window is the launcher.
type Window struct {
	sess  *studio.Session
	win   fyne.Window
	title string

	// onOpenPack is called with a validated pack folder. Wiring it is the
	// caller's job: the launcher does not know how a pack window is made.
	onOpenPack func(dir string)

	root   *fyne.Container
	body   *fyne.Container
	status *statusBar

	// setupDismissed records that the user closed the setup screen, so
	// a later refresh does not put it back in front of them.
	setupDismissed bool
}

// New returns a launcher rendered into win. The title is restored
// whenever the launcher is reinstalled, since a pack window renames the
// window while it holds it.
func New(sess *studio.Session, win fyne.Window, title string) *Window {
	w := &Window{
		sess:   sess,
		win:    win,
		title:  title,
		body:   container.NewStack(),
		status: newStatusBar(sess),
	}

	w.root = container.NewBorder(
		nil, w.status.object(), nil, nil,
		w.body,
	)
	w.Install()
	w.win.CenterOnScreen()

	// The status line and the action buttons both depend on whether a
	// binary is available, so a change from either the setup screen or
	// settings rebuilds the view.
	sess.OnPackwizChange(func(packwiz.Location) {
		fyne.Do(func() {
			w.status.refresh()
			w.Refresh()
		})
	})

	return w
}

// Install puts the launcher into its window and sizes it. It is also how
// a pack window hands the window back when it closes.
func (w *Window) Install() {
	w.win.SetTitle(w.title)
	w.win.SetContent(w.root)
	w.win.Resize(fyne.NewSize(tokens.LauncherWidth, tokens.LauncherHeight))
	w.Refresh()
}

// SetOnOpenPack registers what happens when a pack is chosen.
func (w *Window) SetOnOpenPack(fn func(dir string)) { w.onOpenPack = fn }

// Window exposes the underlying window, for callers that need to show it
// or reuse it for a pack.
func (w *Window) Window() fyne.Window { return w.win }

// Refresh rebuilds the launcher body for the current state.
func (w *Window) Refresh() {
	w.body.Objects = []fyne.CanvasObject{w.content()}
	w.body.Refresh()
	w.status.refresh()
}

// content picks the screen the launcher should be showing.
func (w *Window) content() fyne.CanvasObject {
	if !w.sess.HasPackwiz() && !w.setupDismissed {
		return w.buildSetup()
	}
	return w.buildRecents()
}

// openPack hands a folder to the caller's open handler.
func (w *Window) openPack(dir string) {
	if w.onOpenPack != nil {
		w.onOpenPack(dir)
	}
}
