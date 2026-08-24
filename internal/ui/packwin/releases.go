package packwin

import (
	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	fynetheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/PalisadeMC/Packwiz-Studio/internal/forge"
	"github.com/PalisadeMC/Packwiz-Studio/internal/git"
	"github.com/PalisadeMC/Packwiz-Studio/internal/secrets"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/widgets"
)

// releasesActivity publishes a release to whichever host the remote
// points at.
type releasesActivity struct {
	deps    *activityDeps
	repo    *git.Repo
	secrets *secrets.Store

	list *fyne.Container
	side fyne.CanvasObject
	main *fyne.Container

	status git.Status
	host   forge.Host
	hostOK bool
	tags   []string

	// form is the release form currently on screen, kept so the menu can
	// drive the same fields the buttons do. It is nil until the
	// repository has been read and the form rendered.
	form *releaseForm
}

func newReleasesActivity(deps *activityDeps, repo *git.Repo) *releasesActivity {
	a := &releasesActivity{
		deps:    deps,
		repo:    repo,
		secrets: secrets.New(),
		list:    container.NewVBox(),
		main:    container.NewStack(),
	}

	a.side = container.NewVScroll(a.list)
	a.Reload()
	return a
}

func (a *releasesActivity) ID() string              { return ActivityReleases }
func (a *releasesActivity) Title() string           { return "Releases" }
func (a *releasesActivity) Icon() fyne.Resource     { return fynetheme.UploadIcon() }
func (a *releasesActivity) Side() fyne.CanvasObject { return a.side }
func (a *releasesActivity) Main() fyne.CanvasObject { return a.main }

// Reload re-reads the repository, its tags and which host it points at.
func (a *releasesActivity) Reload() {
	go func() {
		ctx := context.Background()
		status := a.repo.Read(ctx)

		var tags []string
		if status.IsRepo {
			tags, _ = a.repo.Tags(ctx)
		}

		host, ok := resolveHost(status.Remote, a.deps.sess.Cfg.Get().GiteaBaseURL)

		fyne.Do(func() {
			a.status = status
			a.tags = tags
			a.host = host
			a.hostOK = ok
			a.render()

			// The Release menu only exists once a host has been resolved,
			// and that happens here rather than when the window opened.
			a.deps.menuChanged()
		})
	}()
}

// resolveHost works out which API to use from the remote URL.
func resolveHost(remote, configuredBase string) (forge.Host, bool) {
	if remote == "" {
		return forge.Host{}, false
	}

	parsed, err := forge.ParseRemote(remote)
	if err != nil {
		return forge.Host{}, false
	}
	return forge.DetectHost(parsed, configuredBase), true
}

// render redraws both panes.
func (a *releasesActivity) render() {
	a.renderTags()

	switch {
	case !a.status.IsRepo:
		a.message("This pack is not a git repository yet. A release needs one, " +
			"because it points at a tag and is published to a remote.")
	case !a.hostOK:
		a.message("This repository has no remote named origin, so there is " +
			"nowhere to publish a release to.")
	default:
		a.renderForm()
	}
}

// renderTags lists existing tags, which is the pack's release history as
// far as the app can see it.
func (a *releasesActivity) renderTags() {
	a.list.Objects = nil

	if len(a.tags) == 0 {
		a.list.Add(widgets.Inset(tokens.SpaceMD, tokens.SpaceSM,
			widgets.Dim("No tags yet")))
		a.list.Refresh()
		return
	}

	for _, tag := range a.tags {
		a.list.Add(widgets.Inset(tokens.SpaceMD, tokens.SpaceXS, widgets.Body(tag)))
	}
	a.list.Refresh()
}

// message puts a note in the detail area.
func (a *releasesActivity) message(text string) {
	note := widget.NewLabel(text)
	note.Wrapping = fyne.TextWrapWord

	a.main.Objects = []fyne.CanvasObject{
		widgets.Inset(tokens.SpaceXL, tokens.SpaceLG, note),
	}
	a.main.Refresh()
}

// previousTag is the tag a changelog compares against, which is the most
// recent one since tags come back newest first.
func (a *releasesActivity) previousTag() string {
	if len(a.tags) == 0 {
		return ""
	}
	return a.tags[0]
}
