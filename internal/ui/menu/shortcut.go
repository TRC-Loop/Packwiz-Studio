package menu

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

// desktopShortcut is a keyboard shortcut for a menu item. Fyne's own
// desktop.CustomShortcut carries the same data; this wraps it so the
// menu package can build one without every call site importing the
// desktop driver.
type desktopShortcut struct {
	key fyne.KeyName
	mod fyne.KeyModifier
}

func (s *desktopShortcut) inner() *desktop.CustomShortcut {
	return &desktop.CustomShortcut{KeyName: s.key, Modifier: s.mod}
}

// ShortcutName implements fyne.Shortcut.
func (s *desktopShortcut) ShortcutName() string { return s.inner().ShortcutName() }

// Key implements fyne.KeyboardShortcut.
func (s *desktopShortcut) Key() fyne.KeyName { return s.key }

// Mod implements fyne.KeyboardShortcut.
func (s *desktopShortcut) Mod() fyne.KeyModifier { return s.mod }

// Modifier keys used by the menubar. The primary modifier is Command on
// macOS and Control elsewhere, which is what fyne.KeyModifierShortcutDefault
// resolves to.
const (
	modPrimary = fyne.KeyModifierShortcutDefault
	modShift   = fyne.KeyModifierShortcutDefault | fyne.KeyModifierShift
)
