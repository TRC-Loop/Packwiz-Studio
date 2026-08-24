package packwin

import (
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	fynetheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/PalisadeMC/Packwiz-Studio/internal/tomlhl"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/widgets"
)

// editor views a TOML file with highlighting and edits it as plain text.
//
// The two are separate modes on purpose. Fyne has no editable widget that
// takes a colour per token: Entry can be typed into but is one colour,
// TextGrid takes colours but cannot be typed into. Rather than lose the
// highlighting entirely, viewing uses the grid and editing swaps in an
// entry. That suits this screen, which exists for occasional fixes rather
// than sustained editing.
type editor struct {
	onSave func(rel, content string)

	title   *canvas.Text
	view    *widget.TextGrid
	entry   *widget.Entry
	body    *fyne.Container
	actions *fyne.Container
	root    *fyne.Container

	rel      string
	original string
	editing  bool
}

func newEditor(onSave func(rel, content string)) *editor {
	e := &editor{
		onSave: onSave,
		title:  widgets.Caption(""),
		view:   widget.NewTextGrid(),
		entry:  widget.NewMultiLineEntry(),
		body:   container.NewStack(),
	}

	e.entry.TextStyle = fyne.TextStyle{Monospace: true}
	e.entry.Wrapping = fyne.TextWrapOff

	e.actions = container.NewHBox()
	head := container.NewBorder(nil, nil,
		widgets.Inset(tokens.SpaceMD, tokens.SpaceXS, e.title), e.actions, nil)

	e.root = container.NewBorder(
		container.NewBorder(nil, widgets.Hairline(), nil, nil, head),
		nil, nil, nil,
		e.body,
	)
	return e
}

// object returns the editor for placement.
func (e *editor) object() fyne.CanvasObject { return e.root }

// dirty reports unsaved changes.
func (e *editor) dirty() bool {
	return e.editing && e.entry.Text != e.original
}

// load shows a file, leaving edit mode.
func (e *editor) load(rel, content string) {
	e.rel = rel
	e.original = content
	e.editing = false

	e.title.Text = rel
	e.title.Refresh()

	e.entry.SetText(content)
	e.showView()
}

// showView renders the highlighted read-only grid.
func (e *editor) showView() {
	e.editing = false
	e.view.Rows = highlight(e.original)
	e.view.Refresh()

	e.body.Objects = []fyne.CanvasObject{container.NewScroll(e.view)}
	e.body.Refresh()
	e.setActions()
}

// showEdit swaps in the plain text entry.
func (e *editor) showEdit() {
	e.editing = true
	e.entry.SetText(e.original)

	e.body.Objects = []fyne.CanvasObject{container.NewScroll(e.entry)}
	e.body.Refresh()
	e.setActions()
}

// setActions rebuilds the header buttons for the current mode.
func (e *editor) setActions() {
	e.actions.Objects = nil

	if e.editing {
		save := widget.NewButtonWithIcon("Save", fynetheme.DocumentSaveIcon(), func() {
			e.original = e.entry.Text
			e.onSave(e.rel, e.entry.Text)
			e.showView()
		})
		cancel := widget.NewButton("Cancel", e.showView)
		cancel.Importance = widget.LowImportance

		e.actions.Add(cancel)
		e.actions.Add(save)
	} else {
		edit := widget.NewButtonWithIcon("Edit", fynetheme.DocumentCreateIcon(), e.showEdit)
		e.actions.Add(edit)
	}
	e.actions.Refresh()
}

// highlight turns a file into styled grid rows.
func highlight(content string) []widget.TextGridRow {
	lines := strings.Split(content, "\n")

	rows := make([]widget.TextGridRow, 0, len(lines))
	for _, line := range lines {
		var cells []widget.TextGridCell

		for _, tok := range tomlhl.Line(line) {
			style := tokenStyle(tok.Kind)
			for _, r := range tok.Text {
				cells = append(cells, widget.TextGridCell{Rune: r, Style: style})
			}
		}
		rows = append(rows, widget.TextGridRow{Cells: cells})
	}
	return rows
}

// tokenStyle maps a token kind onto a shade and weight.
//
// Syntax highlighting stays grayscale: the status colours mean state, and
// spending them on syntax would dilute that. Four shades plus bold and
// italic separate the kinds well enough for a file this size.
func tokenStyle(kind tomlhl.Kind) widget.TextGridStyle {
	switch kind {
	case tomlhl.KindComment:
		return &widget.CustomTextGridStyle{
			FGColor:   tokens.ColorDim,
			TextStyle: fyne.TextStyle{Italic: true, Monospace: true},
		}
	case tomlhl.KindTable:
		return &widget.CustomTextGridStyle{
			FGColor:   tokens.ColorStrong,
			TextStyle: fyne.TextStyle{Bold: true, Monospace: true},
		}
	case tomlhl.KindKey:
		return mono(tokens.ColorText)
	case tomlhl.KindString:
		return mono(tokens.ColorMuted)
	case tomlhl.KindNumber, tomlhl.KindBool:
		return &widget.CustomTextGridStyle{
			FGColor:   tokens.ColorMuted,
			TextStyle: fyne.TextStyle{Bold: true, Monospace: true},
		}
	case tomlhl.KindPunct:
		return mono(tokens.ColorDim)
	default:
		return mono(tokens.ColorText)
	}
}

// mono is a plain monospace style in one colour.
func mono(c color.Color) widget.TextGridStyle {
	return &widget.CustomTextGridStyle{
		FGColor:   c,
		TextStyle: fyne.TextStyle{Monospace: true},
	}
}
