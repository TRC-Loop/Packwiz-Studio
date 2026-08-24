// Package widgets holds the small reusable pieces shared by the launcher
// and the pack window. Everything here reads its colours and metrics from
// internal/ui/tokens.
package widgets

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	fynetheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
)

// Heading is the largest text step, for a window's subject: the app name
// on the launcher, the pack name in a pack window.
func Heading(text string) *widget.Label {
	l := widget.NewLabel(text)
	l.SizeName = fynetheme.SizeNameHeadingText
	l.TextStyle = fyne.TextStyle{Bold: true}
	return l
}

// SubHeading titles a pane or a form section.
func SubHeading(text string) *widget.Label {
	l := widget.NewLabel(text)
	l.SizeName = fynetheme.SizeNameSubHeadingText
	l.TextStyle = fyne.TextStyle{Bold: true}
	return l
}

// Body is ordinary label text.
func Body(text string) *widget.Label {
	return widget.NewLabel(text)
}

// Strong is primary single-line text drawn as canvas text rather than as
// a Label.
//
// A Label carries the theme's inner padding, so a Label stacked above
// canvas text sits indented from it. Rows that mix tiers use this for
// their first line to keep every line on the same left edge.
func Strong(text string) *canvas.Text {
	return sized(text, tokens.ColorText, tokens.TextBody)
}

// Muted is secondary text at body size: a pack's Minecraft version, a
// mod's category list.
//
// Secondary tiers use canvas.Text rather than a Label because Fyne's
// Label maps its Importance onto theme colour roles, and the palette's
// muted tier is not one of them. These lines are single-line by nature,
// so losing Label's wrapping costs nothing.
func Muted(text string) *canvas.Text {
	return sized(text, tokens.ColorMuted, tokens.TextBody)
}

// Caption is metadata at the smallest step: a version, a timestamp.
func Caption(text string) *canvas.Text {
	return sized(text, tokens.ColorMuted, tokens.TextCaption)
}

// Dim is the quietest tier, for a path or a hint that should recede.
func Dim(text string) *canvas.Text {
	return sized(text, tokens.ColorDim, tokens.TextCaption)
}

// Status is text in one of the status colours, for a command result or a
// validation message.
func Status(text string, state State) *canvas.Text {
	return sized(text, state.Color(), tokens.TextBody)
}

// Note is quiet explanatory text that may run to several lines.
//
// The single-line helpers above are canvas.Text, which draws a literal
// glyph for a newline rather than breaking the line. Anything with a line
// break, or long enough to need wrapping, has to be a Label instead.
func Note(text string) *widget.Label {
	l := widget.NewLabel(text)
	l.Wrapping = fyne.TextWrapWord
	l.SizeName = fynetheme.SizeNameCaptionText
	l.Importance = widget.LowImportance
	return l
}

// Line is quiet single-line text that is cut off rather than allowed to
// overflow.
//
// The canvas.Text helpers above draw their whole string whatever room
// they were given, which inside a fixed size card means the text runs out
// over its neighbours. A Label is the only thing in Fyne that ellipsises,
// so anything whose length comes from the network uses this.
func Line(text string) *widget.Label {
	l := widget.NewLabel(text)
	l.SizeName = fynetheme.SizeNameCaptionText
	l.Importance = widget.LowImportance
	l.Truncation = fyne.TextTruncateEllipsis
	l.Wrapping = fyne.TextWrapOff
	return l
}

// State selects a status colour.
type State int

// Status states.
const (
	// StateNeutral is ordinary text, reporting no problem.
	StateNeutral State = iota
	// StateError marks a failure.
	StateError
	// StateWarning marks something recoverable.
	StateWarning
	// StateSuccess marks a completed action.
	StateSuccess
)

// Color is the token this state renders in.
func (s State) Color() color.Color {
	switch s {
	case StateError:
		return tokens.ColorError
	case StateWarning:
		return tokens.ColorWarning
	case StateSuccess:
		return tokens.ColorSuccess
	default:
		return tokens.ColorMuted
	}
}

// Hairline is a one pixel divider in the border colour.
func Hairline() *canvas.Rectangle {
	r := canvas.NewRectangle(tokens.ColorBorder)
	r.SetMinSize(fyne.NewSize(tokens.HairlineThickness, tokens.HairlineThickness))
	return r
}

func sized(text string, c color.Color, size float32) *canvas.Text {
	t := canvas.NewText(text, c)
	t.TextSize = size
	return t
}
