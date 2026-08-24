package packwin

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	fynetheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/PalisadeMC/Packwiz-Studio/internal/pack"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/widgets"
)

// filesActivity is the pack folder: a tree on the left and an editor on
// the right.
//
// It shows the whole folder rather than only what the index tracks,
// because the files a pack author adds by hand, configs and KubeJS
// scripts above all, are not in the index until packwiz has been run.
// Hiding them until then would hide the very files being worked on.
type filesActivity struct {
	deps *activityDeps

	tree *widget.Tree
	side fyne.CanvasObject
	main *fyne.Container

	scan     *pack.Tree
	editor   *editor
	selected string
}

func newFilesActivity(deps *activityDeps) *filesActivity {
	a := &filesActivity{
		deps: deps,
		main: container.NewStack(),
	}

	a.editor = newEditor(a.save)
	a.buildTree()

	a.side = container.NewBorder(
		widgets.Inset(tokens.SpaceSM, tokens.SpaceXS, a.toolbar()),
		nil, nil, nil, a.tree)

	a.Reload()
	return a
}

func (a *filesActivity) ID() string              { return ActivityFiles }
func (a *filesActivity) Title() string           { return "Files" }
func (a *filesActivity) Icon() fyne.Resource     { return fynetheme.FolderIcon() }
func (a *filesActivity) Side() fyne.CanvasObject { return a.side }
func (a *filesActivity) Main() fyne.CanvasObject { return a.main }

// toolbar is the file actions. They act on the selection, which for a new
// file or folder means the folder it goes in.
func (a *filesActivity) toolbar() fyne.CanvasObject {
	newFile := widget.NewButtonWithIcon("", fynetheme.ContentAddIcon(), a.promptNewFile)
	newDir := widget.NewButtonWithIcon("", fynetheme.FolderNewIcon(), a.promptNewFolder)
	rename := widget.NewButtonWithIcon("", fynetheme.DocumentCreateIcon(), a.promptRename)
	remove := widget.NewButtonWithIcon("", fynetheme.DeleteIcon(), a.promptDelete)

	for _, b := range []*widget.Button{newFile, newDir, rename, remove} {
		b.Importance = widget.LowImportance
	}

	return container.NewHBox(newFile, newDir, rename, remove)
}

// Reload rescans the folder. The open file stays open when it is still
// there, and the tree keeps whatever was expanded.
func (a *filesActivity) Reload() {
	scan, err := a.deps.pack.ScanTree()
	if err != nil {
		a.showMessage("The pack folder could not be read: " + err.Error())
		return
	}

	a.scan = scan
	a.tree.Refresh()

	if a.selected == "" {
		a.showHint()
		return
	}
	if _, ok := a.scan.Find(a.selected); !ok {
		a.selected = ""
		a.showHint()
		return
	}
	a.open(a.selected)
}

// selectFile opens a file, warning first when the open one is unsaved.
// Selecting a folder only moves the selection, since there is nothing to
// open.
func (a *filesActivity) selectFile(rel string) {
	if a.scan.IsDir(rel) {
		a.selected = rel
		return
	}

	if a.editor.dirty() && rel != a.selected {
		dialog.NewConfirm("Discard changes",
			"Discard the unsaved changes to "+filepath.Base(a.selected)+"?",
			func(discard bool) {
				if discard {
					a.openAndFocus(rel)
				}
			}, a.deps.win).Show()
		return
	}
	a.openAndFocus(rel)
}

// openAndFocus opens a file and puts the caret in it, so a click on a
// file is enough to start typing.
//
// Focus is only taken on an explicit selection. A reload also reopens the
// file, and that must not pull the caret out of whatever the user is
// doing elsewhere in the window.
func (a *filesActivity) openAndFocus(rel string) {
	a.open(rel)
	a.editor.focus(a.deps.win.Canvas())
}

// open loads a file into the editor.
func (a *filesActivity) open(rel string) {
	full := filepath.Join(a.deps.pack.Dir, filepath.FromSlash(rel))

	data, err := os.ReadFile(full)
	if err != nil {
		a.showMessage(err.Error())
		return
	}
	if !isText(data) {
		a.selected = rel
		a.showMessage(rel + " is not a text file, so it is not opened here.")
		return
	}

	a.selected = rel
	a.editor.load(rel, string(data))
	a.main.Objects = []fyne.CanvasObject{a.editor.object()}
	a.main.Refresh()
}

// isText reports content the editor can show. A jar or a texture opened
// as text is a screen full of noise and a saved file that is corrupt, so
// the check is on the way in.
func isText(data []byte) bool {
	limit := min(len(data), 4096)
	for _, b := range data[:limit] {
		if b == 0 {
			return false
		}
	}
	return true
}

// showHint is the empty state.
func (a *filesActivity) showHint() {
	a.showMessage("Select a file to edit it. The tree shows the whole pack " +
		"folder: files that are not in the index yet are dimmed, and saving " +
		"or refreshing puts them in.")
}

func (a *filesActivity) showMessage(text string) {
	a.main.Objects = []fyne.CanvasObject{
		widgets.Inset(tokens.SpaceXL, tokens.SpaceLG, widgets.Note(text)),
	}
	a.main.Refresh()
}
