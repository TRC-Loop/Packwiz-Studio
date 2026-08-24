package packwin

import (
	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	fynetheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/PalisadeMC/Packwiz-Studio/internal/cmdrun"
	"github.com/PalisadeMC/Packwiz-Studio/internal/pack"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/widgets"
)

// EditIgnoreRules edits the pack's two ignore files side by side.
//
// They are edited together because they are easy to confuse and the
// mistake is expensive: .gitignore decides what the repository carries,
// while .packwizignore decides what the exported pack carries, and a
// missing rule in the second one ships the repository's own files into
// somebody's game folder.
func (w *Window) EditIgnoreRules() {
	gitTab, gitText, gitErr := w.ignoreTab(pack.GitIgnoreFile,
		"Kept out of the repository.", pack.DefaultGitIgnore())
	packTab, packText, packErr := w.ignoreTab(pack.PackwizIgnoreFile,
		"Kept out of the pack index, so out of the exported pack.",
		pack.DefaultPackwizIgnore())

	if gitErr != nil {
		dialog.ShowError(gitErr, w.win)
		return
	}
	if packErr != nil {
		dialog.ShowError(packErr, w.win)
		return
	}

	tabs := container.NewAppTabs(gitTab, packTab)
	size := widgets.FitDialog(w.win, tokens.FormWidth, tokens.IgnoreHeight)

	d := dialog.NewCustomConfirm("Ignore rules", "Save", "Cancel", tabs,
		func(save bool) {
			if !save {
				return
			}
			w.saveIgnoreRules(gitText.Text, packText.Text)
		}, w.win)

	d.Resize(size)
	d.Show()
}

// ignoreTab builds one file's tab, returning the entry so the dialog can
// read it back when saving.
func (w *Window) ignoreTab(name, note string, recommended []string) (
	*container.TabItem, *widget.Entry, error) {

	current, err := pack.ReadIgnore(w.pack.Dir, name)
	if err != nil {
		return nil, nil, err
	}

	entry := widget.NewMultiLineEntry()
	entry.TextStyle = fyne.TextStyle{Monospace: true}
	entry.SetMinRowsVisible(ignoreRows)
	entry.SetText(current)
	entry.SetPlaceHolder("One rule per line")

	add := widget.NewButtonWithIcon("Add recommended rules",
		fynetheme.ContentAddIcon(), func() {
			entry.SetText(pack.AddRules(entry.Text, recommended))
		})
	add.Importance = widget.LowImportance

	body := container.NewBorder(
		widgets.Note(note), container.NewHBox(add), nil, nil, entry)

	return container.NewTabItem(name,
		widgets.Inset(tokens.SpaceMD, tokens.SpaceMD, body)), entry, nil
}

// saveIgnoreRules writes both files and refreshes the index, because
// .packwizignore decides what the index holds.
func (w *Window) saveIgnoreRules(gitRules, packRules string) {
	w.deps.run("save ignore rules", func(ctx context.Context) error {
		if err := pack.WriteIgnore(w.pack.Dir, pack.GitIgnoreFile, gitRules); err != nil {
			return err
		}
		if err := pack.WriteIgnore(w.pack.Dir, pack.PackwizIgnoreFile, packRules); err != nil {
			return err
		}
		return exec(func() (cmdrun.Result, error) {
			return w.deps.client().Refresh(ctx)
		})
	})
}

// ignoreRows is how much of an ignore file is visible at once. These
// files are short lists, so this is nearly always all of it.
const ignoreRows = 12
