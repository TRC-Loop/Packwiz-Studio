package packwin

import (
	"context"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	fynetheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/PalisadeMC/Packwiz-Studio/internal/changelog"
	"github.com/PalisadeMC/Packwiz-Studio/internal/config"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/widgets"
)

// releaseForm holds the release fields, so publishing can read them all
// without a closure per widget.
type releaseForm struct {
	tag        *widget.Entry
	title      *widget.Entry
	notes      *widget.Entry
	format     *widget.Select
	draft      *widget.Check
	prerelease *widget.Check
	attach     *widget.Check
	images     []string
	imageList  *fyne.Container
}

// renderForm builds the release form in the detail area.
func (a *releasesActivity) renderForm() {
	f := &releaseForm{
		tag:        widget.NewEntry(),
		title:      widget.NewEntry(),
		notes:      widget.NewMultiLineEntry(),
		draft:      widget.NewCheck("Create as a draft", nil),
		prerelease: widget.NewCheck("Mark as a prerelease", nil),
		imageList:  container.NewVBox(),
	}

	f.tag.SetPlaceHolder("v" + orFallback(a.deps.pack.Version, "1.0.0"))
	f.title.SetPlaceHolder("leave empty to use the tag")
	f.notes.Wrapping = fyne.TextWrapWord
	f.notes.SetMinRowsVisible(notesRows)

	prefs := a.deps.prefs()
	f.format = widget.NewSelect(formatLabels, nil)
	f.format.SetSelected(labelForFormat(prefs.ChangelogFormat))

	f.attach = widget.NewCheck(attachLabel(prefs.LastExport), nil)
	f.attach.SetChecked(prefs.LastExport != "")
	if prefs.LastExport == "" {
		f.attach.Disable()
	}

	a.form = f

	generate := widget.NewButtonWithIcon("Generate from git", fynetheme.ViewRefreshIcon(),
		func() { a.generateNotes(f) })

	addImage := widget.NewButtonWithIcon("Attach image", fynetheme.ContentAddIcon(),
		func() { a.pickImage(f) })
	addImage.Importance = widget.LowImportance

	publish := widget.NewButtonWithIcon("Publish release", fynetheme.UploadIcon(),
		func() { a.publish(f) })

	body := container.NewVBox(
		widgets.SubHeading("New release on "+a.host.Kind.Name()),
		widgets.Muted(a.host.Remote.Path()+" at "+a.host.APIBase),
		widgets.VSpace(tokens.SpaceMD),

		widgets.Muted("Tag"),
		f.tag,
		widgets.Dim(a.tagHint()),

		widgets.Muted("Title"),
		f.title,
		widgets.VSpace(tokens.SpaceSM),

		container.NewBorder(nil, nil, widgets.Muted("Changelog"), nil,
			container.NewHBox(f.format, generate)),
		f.notes,
		widgets.VSpace(tokens.SpaceSM),

		f.attach,
		container.NewBorder(nil, nil, widgets.Muted("Images"), nil,
			container.NewHBox(addImage)),
		f.imageList,

		widgets.VSpace(tokens.SpaceSM),
		f.draft,
		f.prerelease,

		widgets.VSpace(tokens.SpaceMD),
		container.NewHBox(publish),
	)

	a.main.Objects = []fyne.CanvasObject{
		container.NewVScroll(widgets.Inset(tokens.SpaceXL, tokens.SpaceLG, body)),
	}
	a.main.Refresh()
}

// notesRows is how tall the changelog field starts, which is enough to
// see a generated list without scrolling.
const notesRows = 10

// tagHint says what the changelog will be compared against.
func (a *releasesActivity) tagHint() string {
	if prev := a.previousTag(); prev != "" {
		return "The changelog compares against " + prev + "."
	}
	return "No previous tag, so a generated changelog lists everything as added."
}

// generateNotes fills the changelog field from the git history, leaving it
// editable.
func (a *releasesActivity) generateNotes(f *releaseForm) {
	format := formatForLabel(f.format.Selected)
	a.deps.setPrefs(func(p *config.Prefs) { p.ChangelogFormat = format })

	prev := a.previousTag()

	go func() {
		ctx := context.Background()

		current, err := changelog.Current(a.deps.pack)
		if err != nil {
			fyne.Do(func() { dialog.ShowError(err, a.deps.win) })
			return
		}

		before := changelog.Snapshot{}
		if prev != "" {
			// A tag whose index cannot be read is treated as an empty
			// pack, which lists everything as added rather than failing.
			if snap, err := changelog.AtRevision(ctx, a.repo, prev,
				a.deps.pack.IndexFile); err == nil {
				before = snap
			}
		}

		text := changelog.Render(changelog.Diff(before, current), format)

		fyne.Do(func() {
			if text == "" {
				text = "No mod changes since " + orFallback(prev, "the start of the pack") + ".\n"
			}
			f.notes.SetText(text)
		})
	}()
}

// pickImage attaches a screenshot or banner to the release notes.
func (a *releasesActivity) pickImage(f *releaseForm) {
	open := dialog.NewFileOpen(func(rc fyne.URIReadCloser, err error) {
		if err != nil || rc == nil {
			return
		}
		path := rc.URI().Path()
		rc.Close()

		f.images = append(f.images, path)
		f.imageList.Add(widgets.Caption(filepath.Base(path)))
		f.imageList.Refresh()
	}, a.deps.win)

	open.SetFilter(storage.NewExtensionFileFilter([]string{".png", ".jpg", ".jpeg", ".webp"}))
	open.Show()
}

// attachLabel describes the export that would be attached.
func attachLabel(lastExport string) string {
	if lastExport == "" {
		return "Attach the exported pack (export one first)"
	}
	return "Attach " + filepath.Base(lastExport)
}

func orFallback(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}
