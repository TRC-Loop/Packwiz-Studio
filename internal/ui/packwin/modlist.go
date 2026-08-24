package packwin

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	fynetheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/PalisadeMC/Packwiz-Studio/internal/config"
	"github.com/PalisadeMC/Packwiz-Studio/internal/modlist"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/widgets"
)

// modListForm is the mod list export dialog's state.
type modListForm struct {
	win *Window

	format *widget.Select
	header *widget.Entry
	line   *widget.Entry
	footer *widget.Entry

	custom  *fyne.Container
	preview *widget.Label
}

// ExportModList writes the pack's mods out as a document.
//
// This is not the same thing as exporting the pack: nobody installs a mod
// list. It is for the page or the post that tells people what is in the
// pack, so the formats are the ones those places want, and a custom
// template covers the ones they do not.
func (w *Window) ExportModList() {
	f := &modListForm{win: w}
	f.build()

	size := widgets.FitDialog(w.win, tokens.ModListWidth, tokens.ModListHeight)

	d := dialog.NewCustomConfirm("Export mod list", "Save", "Cancel",
		f.layout(), func(save bool) {
			if !save {
				return
			}
			f.save()
		}, w.win)

	d.Resize(size)
	d.Show()
}

// build creates the controls and fills them from the remembered choice.
func (f *modListForm) build() {
	prefs := f.win.deps.prefs().ModList
	spec := specFrom(prefs)

	f.header = f.template(spec.Header, "Rendered once, before the list")
	f.line = f.template(spec.Line, "Rendered once per mod")
	f.footer = f.template(spec.Footer, "Rendered once, after the list")

	labels := make([]string, 0, len(modlist.Choices()))
	for _, c := range modlist.Choices() {
		labels = append(labels, c.Label)
	}

	f.preview = widget.NewLabel("")
	f.preview.TextStyle = fyne.TextStyle{Monospace: true}
	f.preview.Wrapping = fyne.TextWrapOff

	f.format = widget.NewSelect(labels, func(string) { f.refresh() })
	f.format.SetSelected(modlist.Find(spec.Format).Label)
}

// template is one of the custom format's boxes.
func (f *modListForm) template(text, hint string) *widget.Entry {
	e := widget.NewMultiLineEntry()
	e.TextStyle = fyne.TextStyle{Monospace: true}
	e.SetMinRowsVisible(templateRows)
	e.SetPlaceHolder(hint)
	e.SetText(text)
	e.OnChanged = func(string) { f.refresh() }
	return e
}

// layout arranges the picker, the templates and the preview.
func (f *modListForm) layout() fyne.CanvasObject {
	f.custom = container.NewVBox(
		widgets.Muted("Header"), f.header,
		widgets.Muted("Line"), f.line,
		widgets.Muted("Footer"), f.footer,
		widgets.Note(placeholderHelp()),
	)

	copyOut := widget.NewButtonWithIcon("Copy", fynetheme.ContentCopyIcon(), f.copy)
	copyOut.Importance = widget.LowImportance

	top := container.NewVBox(
		container.NewBorder(nil, nil, widgets.Muted("Format"), copyOut, f.format),
		f.custom,
		widgets.VSpace(tokens.SpaceSM),
		widgets.Muted("Preview"),
	)

	// The clamp is outside the scroll, not inside it: the scroll is what
	// lets a wide line be read, and the clamp is what stops that line
	// making the dialog wider than the window.
	body := container.NewBorder(top, nil, nil, nil,
		widgets.ClampWidth(tokens.FormWidth, container.NewScroll(f.preview)))

	f.refresh()

	return widgets.Inset(tokens.SpaceMD, tokens.SpaceMD, body)
}

// refresh redraws the preview and shows the template boxes only for the
// format that uses them.
func (f *modListForm) refresh() {
	if f.custom != nil {
		if f.spec().Format == modlist.Custom {
			f.custom.Show()
		} else {
			f.custom.Hide()
		}
	}

	if f.preview != nil {
		f.preview.SetText(f.render())
	}
}

// spec is what the controls currently describe.
func (f *modListForm) spec() modlist.Spec {
	return modlist.Spec{
		Format: modlist.ByLabel(f.format.Selected).Format,
		Header: f.header.Text,
		Line:   f.line.Text,
		Footer: f.footer.Text,
	}
}

// specFrom turns a remembered choice into a spec, filling in the example
// template the first time the dialog is opened.
func specFrom(prefs config.ModListPrefs) modlist.Spec {
	spec := modlist.DefaultTemplate()
	spec.Format = modlist.Find(modlist.Format(prefs.Format)).Format

	if prefs.Header != "" || prefs.Line != "" || prefs.Footer != "" {
		spec.Header, spec.Line, spec.Footer = prefs.Header, prefs.Line, prefs.Footer
	}
	return spec
}

// placeholderHelp lists the tokens a template may use.
func placeholderHelp() string {
	var tokenList []string
	for _, p := range modlist.LinePlaceholders() {
		tokenList = append(tokenList, p.Token)
	}
	for _, p := range modlist.PackPlaceholders() {
		tokenList = append(tokenList, p.Token)
	}

	return "Placeholders: " + strings.Join(tokenList, " ") +
		"\nThe pack ones work in every box, the mod ones only in the line."
}

// templateRows is how tall each template box starts.
const templateRows = 3
