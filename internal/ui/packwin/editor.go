package packwin

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	fynetheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/PalisadeMC/Packwiz-Studio/internal/syntax"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/widgets"
)

// editor is a text editor for one file of the pack.
//
// There is one mode: what you see is what you are typing into, coloured
// as you type. Selecting a file in the list opens it here ready to edit.
type editor struct {
	onSave func(rel, content string)

	title *canvas.Text
	lang  *canvas.Text
	code  *widgets.Code
	root  *fyne.Container

	rel      string
	original string
}

func newEditor(onSave func(rel, content string)) *editor {
	e := &editor{
		onSave: onSave,
		title:  widgets.Caption(""),
		lang:   widgets.Dim(""),
		code:   widgets.NewCode(),
	}

	e.code.OnChanged = func(text string) {
		if text == e.original {
			e.markSaved()
			return
		}
		e.markDirty()
	}

	save := widget.NewButtonWithIcon("Save", fynetheme.DocumentSaveIcon(), e.save)
	save.Importance = widget.HighImportance

	head := container.NewBorder(nil, nil,
		widgets.Inset(tokens.SpaceMD, tokens.SpaceXS, e.title),
		container.NewHBox(e.lang, save), nil)

	e.root = container.NewBorder(
		container.NewBorder(nil, widgets.Hairline(), nil, nil, head),
		nil, nil, nil,
		e.code.Object(),
	)
	return e
}

// object returns the editor for placement.
func (e *editor) object() fyne.CanvasObject { return e.root }

// focus puts the caret in the text.
func (e *editor) focus(c fyne.Canvas) { e.code.Focus(c) }

// dirty reports unsaved changes.
func (e *editor) dirty() bool { return e.code.Text() != e.original }

// load opens a file for editing, coloured and completed as whatever
// language its name says it is.
func (e *editor) load(rel, content string) {
	e.rel = rel
	e.original = content

	lang := syntax.For(rel)
	e.code.SetTokenizer(tokenizerFor(lang))
	e.code.SetWords(lang.Completions())

	e.lang.Text = lang.Name
	e.lang.Refresh()

	e.code.SetText(content)
	e.markSaved()
}

// save writes the current text through the save handler. It is what both
// the button and the keyboard shortcut call.
func (e *editor) save() {
	if e.rel == "" {
		return
	}

	e.original = e.code.Text()
	e.onSave(e.rel, e.original)
	e.markSaved()
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
