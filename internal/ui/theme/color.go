package theme

import (
	"image/color"

	"fyne.io/fyne/v2"
	fynetheme "fyne.io/fyne/v2/theme"

	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/tokens"
)

// Color maps Fyne's colour roles onto the palette. The variant is
// ignored: the app is dark only.
//
// Structural roles resolve to the grayscale ramp. Fyne's semantic roles
// (error, warning, success, hyperlink) resolve to the status colours, so
// a failure reads as a failure at a glance. ColorNamePrimary stays grey:
// it tints buttons, selection and progress bars across the whole UI, and
// colouring it would make the chrome accented.
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

	// Semantic roles carry real hues.
	fynetheme.ColorNameError:   tokens.ColorError,
	fynetheme.ColorNameWarning: tokens.ColorWarning,
	fynetheme.ColorNameSuccess: tokens.ColorSuccess,

	// Text drawn on top of a filled status colour.
	fynetheme.ColorNameForegroundOnError:   tokens.ColorOnError,
	fynetheme.ColorNameForegroundOnWarning: tokens.ColorOnWarning,
	fynetheme.ColorNameForegroundOnSuccess: tokens.ColorOnSuccess,

	// Chrome accent stays grey — see the note above.
	fynetheme.ColorNamePrimary:             tokens.ColorText,
	fynetheme.ColorNameForegroundOnPrimary: tokens.ColorBG,

	// Chrome.
	fynetheme.ColorNameDisabledButton:      tokens.ColorSurface,
	fynetheme.ColorNameScrollBar:           tokens.ColorScrollBar,
	fynetheme.ColorNameScrollBarBackground: tokens.ColorBG,
	fynetheme.ColorNameShadow:              tokens.ColorScrim,
}
