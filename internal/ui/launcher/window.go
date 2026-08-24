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

	"github.com/PalisadeMC/Packwiz-Studio/internal/packwiz"
	"github.com/PalisadeMC/Packwiz-Studio/internal/studio"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
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

	// screen is which launcher screen is showing.
	screen string
	// newPack is the new pack form, kept across refreshes so a settings
	// change or a finished download does not wipe what has been typed.
	newPack *newPackForm
}

// Launcher screens.
const (
	// screenRecents is the default: known packs and the action column.
	screenRecents = ""
	// screenNewPack is the pack creation form.
	screenNewPack = "newpack"
)

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
//
// The new pack form is a screen rather than a dialog. It has nine fields,
// which is more than a popup should hold, and a dialog cannot be smaller
// than its content's minimum size, so a tall form either clips against
// the window or has to be capped and scrolled. A screen just scrolls.
func (w *Window) content() fyne.CanvasObject {
	if !w.sess.HasPackwiz() && !w.setupDismissed {
		return w.buildSetup()
	}
	if w.screen == screenNewPack {
		return w.buildNewPack()
	}
	return w.buildRecents()
}

// show switches the launcher to a screen.
func (w *Window) show(screen string) {
	w.screen = screen
	w.Refresh()
}

// openPack hands a folder to the caller's open handler.
func (w *Window) openPack(dir string) {
	if w.onOpenPack != nil {
		w.onOpenPack(dir)
	}
}
