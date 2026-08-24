package packwin

import (
	"github.com/PalisadeMC/Packwiz-Studio/internal/config"
	"github.com/PalisadeMC/Packwiz-Studio/internal/pack"
)

// browse finds the browser activity, or nil when it is not present.
func (w *Window) browse() *browseActivity {
	for _, a := range w.items {
		if b, ok := a.(*browseActivity); ok {
			return b
		}
	}
	return nil
}

// mods finds the installed mods activity, or nil when it is not present.
func (w *Window) mods() *modsActivity {
	for _, a := range w.items {
		if m, ok := a.(*modsActivity); ok {
			return m
		}
	}
	return nil
}

// FocusBrowse opens the browser and puts the cursor in its search field,
// which is what the Mods menu's add item does.
func (w *Window) FocusBrowse() {
	b := w.browse()
	if b == nil {
		return
	}

	w.Select(ActivityBrowse)
	w.win.Canvas().Focus(b.search)
}

// SetViewMode switches the browser between grid and list.
func (w *Window) SetViewMode(mode config.ViewMode) {
	b := w.browse()
	if b == nil || b.mode == mode {
		return
	}
	b.toggleMode()
}

// SelectedMod reports the mod currently selected in the mods list, and
// whether there is one. The menu uses this to act on the selection.
func (w *Window) SelectedMod() (string, bool) {
	m := w.mods()
	if m == nil || m.selected == "" {
		return "", false
	}
	return m.selected, true
}

// RemoveSelectedMod removes whatever the mods list has selected.
func (w *Window) RemoveSelectedMod() {
	m := w.mods()
	if m == nil {
		return
	}
	if mod := m.find(m.selected); mod != nil {
		m.confirmRemove(*mod)
	}
}

// SetSelectedSide changes the selected mod's side flag.
func (w *Window) SetSelectedSide(side pack.Side) {
	m := w.mods()
	if m == nil {
		return
	}
	mod := m.find(m.selected)
	if mod == nil || mod.SideFlag == side {
		return
	}
	m.applySide(*mod, side)
}
