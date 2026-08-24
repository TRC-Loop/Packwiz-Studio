package theme

import (
	"image/color"

	"fyne.io/fyne/v2"
	fynetheme "fyne.io/fyne/v2/theme"

	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/tokens"
)

// Color maps Fyne's colour roles onto the grayscale palette. The variant
// is ignored: the app is dark only.
//
// Fyne's semantic roles (primary, success, warning, error, hyperlink)
// have no accent colour to map to, so they resolve to text tiers —
// emphasis is carried by weight and iconography, never by hue.
func (Studio) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	if c, ok := colors[name]; ok {
		return c
	}
	return fynetheme.DefaultTheme().Color(name, fynetheme.VariantDark)
}

var colors = map[fyne.ThemeColorName]color.Color{
	// Grounds and surfaces.
	fynetheme.ColorNameBackground:        tokens.ColorBG,
	fynetheme.ColorNameOverlayBackground: tokens.ColorElevated,
	fynetheme.ColorNameMenuBackground:    tokens.ColorElevated,
	fynetheme.ColorNameHeaderBackground:  tokens.ColorSurface,
	fynetheme.ColorNameInputBackground:   tokens.ColorElevated,
	fynetheme.ColorNameButton:            tokens.ColorElevated,

	// Interaction states.
	fynetheme.ColorNameHover:     tokens.ColorHover,
	fynetheme.ColorNamePressed:   tokens.ColorPressed,
	fynetheme.ColorNameSelection: tokens.ColorSelected,
	fynetheme.ColorNameFocus:     tokens.ColorFocus,

	// Lines.
	fynetheme.ColorNameSeparator:                 tokens.ColorBorder,
	fynetheme.ColorNameInputBorder:               tokens.ColorBorder,
	fynetheme.ColorNameInnerWindowBorder:         tokens.ColorBorder,
	fynetheme.ColorNameInnerWindowBorderInactive: tokens.ColorBorder,

	// Text.
	fynetheme.ColorNameForeground:  tokens.ColorText,
	fynetheme.ColorNamePlaceHolder: tokens.ColorDim,
	fynetheme.ColorNameDisabled:    tokens.ColorDim,
	fynetheme.ColorNameHyperlink:   tokens.ColorText,

	// Semantic roles with no hue available: emphasis tier.
	fynetheme.ColorNamePrimary: tokens.ColorText,
	fynetheme.ColorNameError:   tokens.ColorStrong,
	fynetheme.ColorNameWarning: tokens.ColorStrong,
	fynetheme.ColorNameSuccess: tokens.ColorText,

	// Text drawn on top of those roles.
	fynetheme.ColorNameForegroundOnPrimary: tokens.ColorBG,
	fynetheme.ColorNameForegroundOnError:   tokens.ColorBG,
	fynetheme.ColorNameForegroundOnWarning: tokens.ColorBG,
	fynetheme.ColorNameForegroundOnSuccess: tokens.ColorBG,

	// Chrome.
	fynetheme.ColorNameDisabledButton:      tokens.ColorSurface,
	fynetheme.ColorNameScrollBar:           tokens.ColorScrollBar,
	fynetheme.ColorNameScrollBarBackground: tokens.ColorBG,
	fynetheme.ColorNameShadow:              tokens.ColorScrim,
}
