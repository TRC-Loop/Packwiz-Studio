package packwin

import (
	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	fynetheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/TRC-Loop/Packwiz-Studio/internal/git"
	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/tokens"
	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/widgets"
)

// gitActivity stages, commits, pushes and pulls.
//
// Changed files are the list pane and the commit box is the detail area,
// which matches how the rest of the window is arranged.
type gitActivity struct {
	deps *activityDeps
	repo *git.Repo

	list    *fyne.Container
	side    fyne.CanvasObject
	main    *fyne.Container
	message *widget.Entry

	status  git.Status
	changes []git.Change
}

func newGitActivity(deps *activityDeps, repo *git.Repo) *gitActivity {
	a := &gitActivity{
		deps:    deps,
		repo:    repo,
		list:    container.NewVBox(),
		main:    container.NewStack(),
		message: widget.NewMultiLineEntry(),
	}

	a.message.SetPlaceHolder("Commit message")
	a.message.Wrapping = fyne.TextWrapWord

	a.side = container.NewVScroll(a.list)

	a.Reload()
	return a
}

func (a *gitActivity) ID() string              { return ActivityGit }
func (a *gitActivity) Title() string           { return "Git" }
func (a *gitActivity) Icon() fyne.Resource     { return fynetheme.StorageIcon() }
func (a *gitActivity) Side() fyne.CanvasObject { return a.side }
func (a *gitActivity) Main() fyne.CanvasObject { return a.main }

// Reload re-reads the repository state off the main goroutine, since it
// shells out several times.
func (a *gitActivity) Reload() {
	go func() {
		status := a.repo.Read(context.Background())

		var changes []git.Change
		if status.IsRepo {
			changes, _ = a.repo.Changes(context.Background())
		}

		fyne.Do(func() {
			a.status = status
			a.changes = changes
			a.render()
		})
	}()
}

// render redraws both panes for the current state.
func (a *gitActivity) render() {
	a.renderChanges()

	if !a.status.IsRepo {
		a.renderNoRepo()
		return
	}
	a.renderCommitPane()
}

// renderChanges rebuilds the changed files list.
func (a *gitActivity) renderChanges() {
	a.list.Objects = nil

	if !a.status.IsRepo {
		a.list.Add(widgets.Inset(tokens.SpaceMD, tokens.SpaceSM,
			widgets.Dim("Not a repository")))
		a.list.Refresh()
		return
	}

	if len(a.changes) == 0 {
		a.list.Add(widgets.Inset(tokens.SpaceMD, tokens.SpaceSM,
			widgets.Dim("Nothing changed")))
		a.list.Refresh()
		return
	}

	for _, c := range a.changes {
		a.list.Add(a.changeRow(c))
	}
	a.list.Refresh()
}

// changeRow is one changed path, with a control to stage or unstage it.
func (a *gitActivity) changeRow(c git.Change) fyne.CanvasObject {
	path := c.Path

	label := widgets.Body(shortPath(c.Path))
	label.Truncation = fyne.TextTruncateEllipsis

	state := widgets.Dim(c.Label())

	var toggle *widget.Button
	if c.Staged {
		toggle = widget.NewButtonWithIcon("", fynetheme.ContentRemoveIcon(), func() {
			a.runGit("unstage "+path, func(ctx context.Context) error {
				return execGit(func() (result, error) { return a.repo.Unstage(ctx, path) })
			})
		})
	} else {
		toggle = widget.NewButtonWithIcon("", fynetheme.ContentAddIcon(), func() {
			a.runGit("stage "+path, func(ctx context.Context) error {
				return execGit(func() (result, error) { return a.repo.Stage(ctx, path) })
			})
		})
	}
	toggle.Importance = widget.LowImportance

	row := container.NewBorder(nil, nil, nil,
		container.NewHBox(state, toggle),
		label,
	)
	return widgets.Inset(tokens.SpaceSM, tokens.SpaceXS, row)
}

// renderNoRepo offers to create a repository.
func (a *gitActivity) renderNoRepo() {
	note := widget.NewLabel(
		"This pack folder is not a git repository.\n\n" +
			"Creating one lets the app track changes and publish releases. " +
			"You can also turn the git integration off in Settings if you " +
			"manage the repository elsewhere.")
	note.Wrapping = fyne.TextWrapWord

	create := widget.NewButtonWithIcon("Initialise repository", fynetheme.StorageIcon(), func() {
		a.runGit("git init", func(ctx context.Context) error {
			return execGit(func() (result, error) { return a.repo.Init(ctx) })
		})
	})

	body := container.NewVBox(
		widgets.SubHeading("No repository"),
		note,
		widgets.VSpace(tokens.SpaceMD),
		container.NewHBox(create),
	)

	a.main.Objects = []fyne.CanvasObject{
		widgets.Inset(tokens.SpaceXL, tokens.SpaceLG, body),
	}
	a.main.Refresh()
}

// shortPath trims a path for the narrow list pane, keeping the end, which
// is the part that identifies the file.
func shortPath(path string) string {
	const maxLen = 34

	if len(path) <= maxLen {
		return path
	}
	return "..." + path[len(path)-maxLen+3:]
}
