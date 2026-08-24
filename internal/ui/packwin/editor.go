package packwin

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	fynetheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

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
	e.entry.OnChanged = func(text string) {
		if text == e.original {
			e.markSaved()
			return
		}
		e.markDirty()
	}

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

// load opens a file for editing.
//
// Selecting a file goes straight into edit mode: this screen exists to
// change something, and making every edit a two step affair is friction
// for no gain. The highlighted read view stays available from the header.
func (e *editor) load(rel, content string) {
	e.rel = rel
	e.original = content

	e.title.Text = rel
	e.title.Refresh()

	e.entry.SetText(content)
	e.showEdit()
}

// save writes the current text through the save handler. It is what both
// the button and the keyboard shortcut call.
func (e *editor) save() {
	if !e.editing || e.rel == "" {
		return
	}

	e.original = e.entry.Text
	e.onSave(e.rel, e.entry.Text)
	e.markSaved()
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
		save := widget.NewButtonWithIcon("Save", fynetheme.DocumentSaveIcon(), e.save)
		save.Importance = widget.HighImportance

		view := widget.NewButtonWithIcon("Highlight",
			fynetheme.VisibilityIcon(), e.showView)
		view.Importance = widget.LowImportance

		e.actions.Add(view)
		e.actions.Add(save)
	} else {
		edit := widget.NewButtonWithIcon("Edit", fynetheme.DocumentCreateIcon(), e.showEdit)
		e.actions.Add(edit)
	}
	e.actions.Refresh()
}

// markSaved restates the title now that there are no pending changes.
func (e *editor) markSaved() {
	e.title.Text = e.rel
	e.title.Color = tokens.ColorMuted
	e.title.Refresh()
}

// markDirty marks the title so unsaved changes are visible.
func (e *editor) markDirty() {
	if e.rel == "" {
		return
	}
	e.title.Text = e.rel + "  (unsaved)"
	e.title.Color = tokens.ColorWarning
	e.title.Refresh()
}
