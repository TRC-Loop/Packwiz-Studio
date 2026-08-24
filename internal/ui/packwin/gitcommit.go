package packwin

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	fynetheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/PalisadeMC/Packwiz-Studio/internal/sysopen"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/widgets"
)

// renderCommitPane is the detail area for a repository: a summary, the
// commit box, and the remote actions.
func (a *gitActivity) renderCommitPane() {
	body := container.NewVBox(
		widgets.SubHeading(a.status.Label()),
		widgets.Muted(a.remoteLine()),
		widgets.VSpace(tokens.SpaceMD),

		a.stageRow(),
		widgets.VSpace(tokens.SpaceMD),

		widgets.Muted("Commit message"),
		a.message,
		widgets.VSpace(tokens.SpaceSM),
		a.commitRow(),

		widgets.VSpace(tokens.SpaceLG),
		widgets.Muted("Remote"),
		a.remoteRow(),
	)

	a.main.Objects = []fyne.CanvasObject{
		container.NewVScroll(widgets.Inset(tokens.SpaceXL, tokens.SpaceLG, body)),
	}
	a.main.Refresh()
}

// remoteLine describes the remote and how far the branch has drifted.
func (a *gitActivity) remoteLine() string {
	if a.status.Remote == "" {
		return "No remote configured"
	}

	line := a.status.Remote
	switch {
	case a.status.Ahead > 0 && a.status.Behind > 0:
		line += ", " + strconv.Itoa(a.status.Ahead) + " ahead and " +
			strconv.Itoa(a.status.Behind) + " behind"
	case a.status.Ahead > 0:
		line += ", " + plural(a.status.Ahead, "commit") + " to push"
	case a.status.Behind > 0:
		line += ", " + plural(a.status.Behind, "commit") + " to pull"
	}
	return line
}

// stageRow offers to stage everything at once.
func (a *gitActivity) stageRow() fyne.CanvasObject {
	stage := widget.NewButtonWithIcon("Stage all changes", fynetheme.ContentAddIcon(), func() {
		a.runGit("stage all", func(ctx context.Context) error {
			return execGit(func() (result, error) { return a.repo.StageAll(ctx) })
		})
	})
	if len(a.changes) == 0 {
		stage.Disable()
	}
	return container.NewHBox(stage)
}

// commitRow is the commit button. Committing fires directly: it is a local
// action and easy to amend, so a confirmation would only be in the way.
func (a *gitActivity) commitRow() fyne.CanvasObject {
	commit := widget.NewButtonWithIcon("Commit", fynetheme.ConfirmIcon(), a.commit)
	if !a.hasStaged() {
		commit.Disable()
	}
	return container.NewHBox(commit)
}

// remoteRow holds push and pull. Push is confirmed first, because it is
// the one action here that other people see.
func (a *gitActivity) remoteRow() fyne.CanvasObject {
	push := widget.NewButtonWithIcon("Push", fynetheme.MoveUpIcon(), a.confirmPush)
	pull := widget.NewButtonWithIcon("Pull", fynetheme.MoveDownIcon(), func() {
		a.runGit("pull", func(ctx context.Context) error {
			return execGit(func() (result, error) { return a.repo.Pull(ctx) })
		})
	})

	row := container.NewHBox(push, pull)

	if a.status.Remote != "" {
		open := widget.NewButtonWithIcon("Open remote", fynetheme.ComputerIcon(), func() {
			url, err := remoteWebURL(a.status.Remote)
			if err != nil {
				dialog.ShowError(err, a.deps.win)
				return
			}
			if err := sysopen.Browse(url); err != nil {
				dialog.ShowError(err, a.deps.win)
			}
		})
		open.Importance = widget.LowImportance
		row.Add(open)
	} else {
		push.Disable()
		pull.Disable()
	}

	return row
}

// commit records the staged changes.
func (a *gitActivity) commit() {
	message := strings.TrimSpace(a.message.Text)
	if message == "" {
		dialog.ShowError(errNoMessage, a.deps.win)
		return
	}
	if !a.hasStaged() {
		dialog.ShowError(errNothingStaged, a.deps.win)
		return
	}

	a.runGit("commit", func(ctx context.Context) error {
		if err := execGit(func() (result, error) { return a.repo.Commit(ctx, message) }); err != nil {
			return err
		}
		fyne.Do(func() { a.message.SetText("") })
		return nil
	})
}

// confirmPush shows what is about to be sent before sending it.
func (a *gitActivity) confirmPush() {
	if a.status.Remote == "" {
		dialog.ShowError(errNoRemote, a.deps.win)
		return
	}

	branch := a.status.Branch
	if branch == "" {
		dialog.ShowError(errors.New("cannot push a detached head, check out a branch first"),
			a.deps.win)
		return
	}

	detail := "Push " + plural(a.status.Ahead, "commit") + " from " + branch +
		"\nto " + a.status.Remote + "?"
	if a.status.Ahead == 0 {
		detail = "Nothing appears to be ahead of the remote.\n\nPush " +
			branch + " to " + a.status.Remote + " anyway?"
	}

	dialog.NewConfirm("Push", detail, func(ok bool) {
		if !ok {
			return
		}
		a.runGit("push", func(ctx context.Context) error {
			setUpstream := !a.repo.HasUpstream(ctx)
			return execGit(func() (result, error) {
				return a.repo.Push(ctx, branch, setUpstream)
			})
		})
	}, a.deps.win).Show()
}

// hasStaged reports whether anything is in the index.
func (a *gitActivity) hasStaged() bool {
	for _, c := range a.changes {
		if c.Staged {
			return true
		}
	}
	return false
}
