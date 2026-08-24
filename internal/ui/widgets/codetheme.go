package widgets

import (
	"image/color"

	"fyne.io/fyne/v2"
	fynetheme "fyne.io/fyne/v2/theme"
)

// invisibleText is the app theme with the foreground text colour removed.
//
// It is applied to the editing entry only. The entry keeps its own
// background, border, cursor and selection colours, but draws its text in
// nothing at all, which is what lets the highlighted layer above it be
// the visible text.
type invisibleText struct {
	base fyne.Theme
}

func newInvisibleText() invisibleText {
	return invisibleText{base: fyne.CurrentApp().Settings().Theme()}
}

func (t invisibleText) Color(name fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	if name == fynetheme.ColorNameForeground {
		return color.Transparent
	}
	return t.base.Color(name, v)
}

func (t invisibleText) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t invisibleText) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t invisibleText) Size(name fyne.ThemeSizeName) float32 {
	return t.base.Size(name)
}
