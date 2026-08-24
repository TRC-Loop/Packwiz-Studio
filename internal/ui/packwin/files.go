package packwin

import (
	"os"
	"path/filepath"
	"slices"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	fynetheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/PalisadeMC/Packwiz-Studio/internal/pack"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/widgets"
)

// filesActivity lists the pack's TOML files and edits one at a time.
//
// This is the escape hatch for the rare manual fix, not the main way to
// work on a pack, so it stays deliberately plain.
type filesActivity struct {
	deps *activityDeps

	list  *fyne.Container
	side  fyne.CanvasObject
	main  *fyne.Container
	rows  map[string]*widgets.Clickable
	paths []string

	editor   *editor
	selected string
}

func newFilesActivity(deps *activityDeps) *filesActivity {
	a := &filesActivity{
		deps: deps,
		list: container.NewVBox(),
		main: container.NewStack(),
		rows: map[string]*widgets.Clickable{},
	}

	a.editor = newEditor(a.save)
	a.side = container.NewVScroll(a.list)

	a.Reload()
	return a
}

func (a *filesActivity) ID() string              { return ActivityFiles }
func (a *filesActivity) Title() string           { return "Files" }
func (a *filesActivity) Icon() fyne.Resource     { return fynetheme.DocumentIcon() }
func (a *filesActivity) Side() fyne.CanvasObject { return a.side }
func (a *filesActivity) Main() fyne.CanvasObject { return a.main }

// Reload rebuilds the file list. The open file stays open when it is
// still there.
func (a *filesActivity) Reload() {
	a.paths = a.collect()
	a.renderList()

	if a.selected == "" {
		a.showHint()
		return
	}
	if !contains(a.paths, a.selected) {
		a.selected = ""
		a.showHint()
		return
	}
	a.open(a.selected)
}

// collect lists the pack's own files plus every file the index tracks.
// The pack file and the index come first because they are the ones a
// manual fix usually targets.
func (a *filesActivity) collect() []string {
	paths := []string{pack.FileName}
	if a.deps.pack.IndexFile != "" {
		paths = append(paths, a.deps.pack.IndexFile)
	}

	var tracked []string
	if mods, err := a.deps.pack.Mods(); err == nil {
		for _, m := range mods {
			tracked = append(tracked, m.Path)
		}
	}
	if files, err := a.deps.pack.Files(); err == nil {
		tracked = append(tracked, files...)
	}

	sort.Strings(tracked)
	return append(paths, tracked...)
}

// renderList rebuilds the side panel.
func (a *filesActivity) renderList() {
	a.list.Objects = nil
	a.rows = map[string]*widgets.Clickable{}

	for _, rel := range a.paths {
		path := rel

		label := widgets.Body(filepath.Base(rel))
		label.Truncation = fyne.TextTruncateEllipsis

		row := widgets.NewClickable(label, func() { a.selectFile(path) })
		row.SetSelected(rel == a.selected)

		a.rows[rel] = row
		a.list.Objects = append(a.list.Objects, row)
	}
	a.list.Refresh()
}

// selectFile opens a file, warning first when the open one is unsaved.
func (a *filesActivity) selectFile(rel string) {
	if a.editor.dirty() && rel != a.selected {
		dialog.NewConfirm("Discard changes",
			"Discard the unsaved changes to "+filepath.Base(a.selected)+"?",
			func(discard bool) {
				if discard {
					a.open(rel)
				}
			}, a.deps.win).Show()
		return
	}
	a.open(rel)
}

// open loads a file into the editor.
func (a *filesActivity) open(rel string) {
	full := filepath.Join(a.deps.pack.Dir, filepath.FromSlash(rel))

	data, err := os.ReadFile(full)
	if err != nil {
		a.showMessage(err.Error())
		return
	}

	a.selected = rel
	for p, row := range a.rows {
		row.SetSelected(p == rel)
	}

	a.editor.load(rel, string(data))
	a.main.Objects = []fyne.CanvasObject{a.editor.object()}
	a.main.Refresh()
}

// showHint is the empty state.
func (a *filesActivity) showHint() {
	a.showMessage("Select a file to view it. Editing here is for the rare " +
		"manual fix, so the pack index is refreshed after every save.")
}

func (a *filesActivity) showMessage(text string) {
	note := widget.NewLabel(text)
	note.Wrapping = fyne.TextWrapWord

	a.main.Objects = []fyne.CanvasObject{
		widgets.Inset(tokens.SpaceXL, tokens.SpaceLG, note),
	}
	a.main.Refresh()
}

func contains(list []string, want string) bool {
	return slices.Contains(list, want)
}
