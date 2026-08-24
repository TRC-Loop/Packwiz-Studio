package packwin

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	fynetheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/PalisadeMC/Packwiz-Studio/internal/git"
)

// generateRow offers a message written from what changed.
//
// It fills the box rather than committing, and it will not quietly
// overwrite something already typed there: the suggestion is a starting
// point, and the wording stays the user's.
func (a *gitActivity) generateRow() fyne.CanvasObject {
	gen := widget.NewButtonWithIcon("Generate", fynetheme.DocumentCreateIcon(),
		a.generateMessage)
	gen.Importance = widget.LowImportance

	if len(a.changes) == 0 {
		gen.Disable()
	}
	return container.NewHBox(gen)
}

// generateMessage describes the changes and puts that in the message box.
func (a *gitActivity) generateMessage() {
	msg := git.CommitMessage(a.changes)
	if msg == "" {
		dialog.ShowInformation("Generate commit message",
			"There is nothing changed to describe.", a.deps.win)
		return
	}

	if strings.TrimSpace(a.message.Text) != "" {
		dialog.NewConfirm("Replace the message",
			"Replace what is in the message box with:\n\n"+msg,
			func(ok bool) {
				if ok {
					a.message.SetText(msg)
				}
			}, a.deps.win).Show()
		return
	}
	a.message.SetText(msg)
}
