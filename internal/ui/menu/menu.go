// Package menu builds the application menubar.
//
// Menus are rebuilt rather than mutated when what they offer changes.
// Fyne has no reliable way to enable or disable an item after the menu is
// installed, so an action that is unavailable is left out of the menu
// entirely, matching how the rail hides absent activities.
package menu

import (
	"fyne.io/fyne/v2"
)

// Actions is everything the menubar can invoke. A nil field means the
// action does not apply right now and its item is omitted.
type Actions struct {
	// File.
	NewPack     func()
	OpenPack    func()
	Recents     []Recent
	CloseWindow func()
	Settings    func()

	// Pack.
	Refresh       func()
	Export        func()
	CheckUpdates  func()
	SetLogo       func()
	RemoveLogo    func()
	RevealInFiles func()

	// Mods.
	AddMod     func()
	RemoveMod  func()
	SideClient func()
	SideServer func()
	SideBoth   func()

	// Git. All nil when the git integration is off.
	GitInit    func()
	GitStage   func()
	GitCommit  func()
	GitPush    func()
	GitPull    func()
	OpenRemote func()

	// Release. All nil when the git integration is off.
	NewRelease        func()
	GenerateChangelog func()
	ForgetToken       func()

	// View.
	ToggleSidePanel func()
	ToggleLog       func()
	GridView        func()
	ListView        func()
	Activities      []ActivityItem

	// Help.
	About          func()
	PackwizVersion func()
}

// Recent is one entry in the File menu's recent packs submenu.
type Recent struct {
	Label string
	Open  func()
}

// ActivityItem is one entry in the View menu's activity submenu, giving
// keyboard access to the icon rail.
type ActivityItem struct {
	Label  string
	Select func()
}

// Build assembles the menubar from whichever actions are present. Menus
// that end up empty are dropped, so a window with git turned off simply
// has no Git menu.
func Build(a Actions) *fyne.MainMenu {
	menus := []*fyne.Menu{
		fileMenu(a),
		packMenu(a),
		modsMenu(a),
		gitMenu(a),
		releaseMenu(a),
		viewMenu(a),
		helpMenu(a),
	}

	kept := make([]*fyne.Menu, 0, len(menus))
	for _, m := range menus {
		if m != nil && len(m.Items) > 0 {
			kept = append(kept, m)
		}
	}
	return fyne.NewMainMenu(kept...)
}

// item builds a menu item, or nil when the action is absent.
func item(label string, fn func()) *fyne.MenuItem {
	if fn == nil {
		return nil
	}
	return fyne.NewMenuItem(label, fn)
}

// withShortcut attaches a keyboard shortcut to an item, tolerating a nil
// item so callers can chain it onto an optional action.
func withShortcut(it *fyne.MenuItem, key fyne.KeyName, mod fyne.KeyModifier) *fyne.MenuItem {
	if it == nil {
		return nil
	}
	it.Shortcut = &desktopShortcut{key: key, mod: mod}
	return it
}

// compact drops nil items and collapses separators that ended up at an
// edge or next to another separator.
func compact(items ...*fyne.MenuItem) []*fyne.MenuItem {
	out := make([]*fyne.MenuItem, 0, len(items))
	for _, it := range items {
		if it == nil {
			continue
		}
		if it.IsSeparator && (len(out) == 0 || out[len(out)-1].IsSeparator) {
			continue
		}
		out = append(out, it)
	}
	for len(out) > 0 && out[len(out)-1].IsSeparator {
		out = out[:len(out)-1]
	}
	return out
}

// separator is a divider between groups of items.
func separator() *fyne.MenuItem { return fyne.NewMenuItemSeparator() }
