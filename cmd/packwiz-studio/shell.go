package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"

	"github.com/PalisadeMC/Packwiz-Studio/internal/appmeta"
	"github.com/PalisadeMC/Packwiz-Studio/internal/pack"
	"github.com/PalisadeMC/Packwiz-Studio/internal/studio"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/launcher"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/packwin"
)

// shell owns one window and what is currently inside it.
//
// The window is reused: opening a pack replaces the launcher's contents,
// and closing the pack puts the launcher back. A second pack opens in a
// new window with its own shell, which is how two packs sit side by side.
type shell struct {
	sess     *studio.Session
	win      fyne.Window
	launcher *launcher.Window
	current  *packwin.Window
}

// newShell builds a shell showing the launcher.
func newShell(sess *studio.Session, win fyne.Window) *shell {
	s := &shell{sess: sess, win: win}
	s.launcher = launcher.New(sess, win, appmeta.Label())
	s.launcher.SetOnOpenPack(s.openPack)
	return s
}

// showLauncher puts the launcher back in charge of the window.
func (s *shell) showLauncher() {
	s.current = nil
	s.launcher.Install()
	s.applyMenu(s.launcherMenu())
}

// openPack loads a pack and hands the window to a pack view.
func (s *shell) openPack(dir string) {
	p, err := pack.Load(dir)
	if err != nil {
		dialog.ShowError(err, s.win)
		return
	}

	// Reopening a pack in a window that already holds one detaches the
	// old view first, so its log subscription does not outlive it.
	if s.current != nil {
		s.current.Detach()
	}

	w := packwin.New(s.sess, s.win, p, s.showLauncher)
	s.current = w

	// Some menu items depend on what is selected inside the window, and
	// Fyne menus cannot be changed once installed, so the whole menubar is
	// rebuilt whenever the window says its contents moved on.
	w.SetOnMenuChanged(func() { s.applyMenu(s.packMenu(w)) })

	w.Install()
	s.applyMenu(s.packMenu(w))
}

// openPackInNewWindow opens a pack beside the current one.
func (s *shell) openPackInNewWindow(dir string) {
	win := s.sess.App.NewWindow(appName)
	other := newShell(s.sess, win)
	other.openPack(dir)
	win.Show()
}

// newLauncherWindow opens another window showing the launcher, for
// choosing a second pack to work on.
func (s *shell) newLauncherWindow() {
	win := s.sess.App.NewWindow(appName)
	other := newShell(s.sess, win)
	other.showLauncher()
	win.Show()
}

// applyMenu installs a menubar on this shell's window.
func (s *shell) applyMenu(m *fyne.MainMenu) {
	s.win.SetMainMenu(m)
}
