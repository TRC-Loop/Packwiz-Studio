// Package tokens holds every design constant used by the UI: colours,
// spacing, radii and the type scale. Nothing outside this package may
// declare a colour literal or a raw spacing value.
package tokens

import "image/color"

// The chrome palette is pure grayscale, with no accent colour in any
// structural surface, control or text tier. Levels are ordered from the
// window ground upward. Semantic status colours are the one exception
// and live in the block below.
var (
	// ColorBG is the window ground and the main content area.
	ColorBG = gray(0x1E)
	// ColorSurface backs the side panel, activity bar and status bar.
	ColorSurface = gray(0x25, 0x25, 0x26)
	// ColorElevated backs cards, inputs, menus and the log drawer.
	ColorElevated = gray(0x2D, 0x2D, 0x30)

	// ColorHover is a row or control under the pointer.
	ColorHover = gray(0x37, 0x37, 0x3A)
	// ColorSelected is the current row in a list.
	ColorSelected = gray(0x3F, 0x3F, 0x45)
	// ColorPressed is a control held down.
	ColorPressed = gray(0x45, 0x45, 0x4B)

	// ColorBorder draws hairlines, dividers and input outlines.
	ColorBorder = gray(0x3E, 0x3E, 0x42)

	// ColorText is primary body and label text.
	ColorText = gray(0xD4)
	// ColorMuted is secondary text: versions, paths, categories.
	ColorMuted = gray(0x9A)
	// ColorDim is placeholder and disabled text.
	ColorDim = gray(0x6E)
	// ColorStrong is the emphasis tier: headings and error text.
	ColorStrong = gray(0xF0)

	// ColorFocus is the keyboard focus ring.
	ColorFocus = gray(0xC8)
	// ColorScrim dims the window behind a modal dialog.
	ColorScrim = color.NRGBA{R: 0, G: 0, B: 0, A: 0x99}
	// ColorScrollBar is the scrollbar thumb.
	ColorScrollBar = color.NRGBA{R: 0x9A, G: 0x9A, B: 0x9A, A: 0x66}
)

// Semantic status colours. These are the only hues in the app: they mark
// state, never decoration. Use them for error and warning text, git
// clean/dirty state, and added/removed changelog entries, never as an
// accent on a button, a selection or a heading. Each is chosen to stay
// legible on ColorBG without glowing.
var (
	// ColorError marks a failed command or an invalid field.
	ColorError = rgb(0xF1, 0x4C, 0x4C)
	// ColorWarning marks a recoverable problem or an unsaved change.
	ColorWarning = rgb(0xCC, 0xA7, 0x00)
	// ColorSuccess marks a completed command or a clean worktree.
	ColorSuccess = rgb(0x89, 0xD1, 0x85)
	// ColorLink marks an external link, such as a mod's Modrinth page.
	ColorLink = rgb(0x37, 0x94, 0xFF)

	// ColorOnError, ColorOnWarning and ColorOnSuccess are text drawn on
	// top of a filled status colour, picked for contrast against it.
	ColorOnError   = ColorStrong
	ColorOnWarning = ColorBG
	ColorOnSuccess = ColorBG
)

// rgb builds an opaque colour from explicit channels.
func rgb(r, g, b uint8) color.NRGBA {
	return color.NRGBA{R: r, G: g, B: b, A: 0xFF}
}

// gray builds an opaque grey. Called with one argument it produces a
// neutral grey; with three it allows the slight tint the charcoal ramp
// uses to separate surface levels.
func gray(v ...uint8) color.NRGBA {
	if len(v) == 1 {
		return color.NRGBA{R: v[0], G: v[0], B: v[0], A: 0xFF}
	}
	return color.NRGBA{R: v[0], G: v[1], B: v[2], A: 0xFF}
}
