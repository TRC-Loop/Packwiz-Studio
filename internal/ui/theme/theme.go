// Package theme implements the single fyne.Theme used by the app. Every
// value it returns is derived from internal/ui/tokens. The theme adds
// no constants of its own.
package theme

import (
	"fyne.io/fyne/v2"
	fynetheme "fyne.io/fyne/v2/theme"
)

// Studio is the app theme: dark only, pure grayscale, Fyne's built-in
// font and icon sets. It deliberately ignores fyne.ThemeVariant because
// there is no light mode.
type Studio struct{}

// New returns the app theme.
func New() fyne.Theme { return Studio{} }

var _ fyne.Theme = Studio{}

// Font returns Fyne's bundled font for the given style.
func (Studio) Font(style fyne.TextStyle) fyne.Resource {
	return fynetheme.DefaultTheme().Font(style)
}

// Icon returns Fyne's built-in icon for name. The app uses no external
// icon pack.
func (Studio) Icon(name fyne.ThemeIconName) fyne.Resource {
	return fynetheme.DefaultTheme().Icon(name)
}
