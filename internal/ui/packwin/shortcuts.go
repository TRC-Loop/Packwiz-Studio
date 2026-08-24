package packwin

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

// installShortcuts registers the window's keyboard shortcuts.
//
// Save is here rather than in the menubar because it acts on whatever the
// file editor currently holds, and a menu item would have to be rebuilt
// every time the selection changed.
func (w *Window) installShortcuts() {
	w.win.Canvas().AddShortcut(
		&desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: fyne.KeyModifierShortcutDefault},
		func(fyne.Shortcut) { w.SaveCurrentFile() },
	)
}

// SaveCurrentFile writes the file open in the editor, if there is one.
func (w *Window) SaveCurrentFile() {
	if files := w.filesActivity(); files != nil {
		files.editor.save()
	}
}

// filesActivity finds the file editing activity.
func (w *Window) filesActivity() *filesActivity {
	for _, a := range w.items {
		if f, ok := a.(*filesActivity); ok {
			return f
		}
	}
	return nil
}
