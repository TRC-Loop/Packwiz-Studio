package packwin

import (
	"context"
	"errors"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/PalisadeMC/Packwiz-Studio/internal/cmdrun"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/widgets"
)

// AddFromURL adds a mod from a direct download link.
//
// This is the escape hatch for a mod that is not on Modrinth: a jar on
// someone's site, a CI build, anything with a stable URL. packwiz records
// it as an external file, which means it cannot be updated or have its
// dependencies resolved, so the form says as much.
func (w *Window) AddFromURL() {
	name := widget.NewEntry()
	name.SetPlaceHolder("the mod's name")

	url := widget.NewEntry()
	url.SetPlaceHolder("https://example.com/some-mod-1.0.0.jar")

	body := container.NewVBox(
		widgets.Muted("Name"),
		name,
		widgets.Muted("Download URL"),
		url,
		widgets.Note("A direct link to the jar. packwiz stores it as an "+
			"external file, so it cannot check it for updates or pull in its "+
			"dependencies. A mod that is on Modrinth is better added from "+
			"the browser."),
	)

	size := widgets.FitDialog(w.win, tokens.FormWidth, tokens.AddModHeight)

	d := dialog.NewCustomConfirm("Add mod from URL", "Add", "Cancel",
		widgets.Scrollable(size.Width, size.Height,
			widgets.Inset(tokens.SpaceMD, tokens.SpaceMD, body)),
		func(add bool) {
			if add {
				w.runAddURL(name.Text, url.Text)
			}
		}, w.win)

	d.Resize(size)
	d.Show()
}

// runAddURL validates and adds an external file.
func (w *Window) runAddURL(name, url string) {
	name = strings.TrimSpace(name)
	url = strings.TrimSpace(url)

	if name == "" {
		dialog.ShowError(errors.New("enter a name for the mod"), w.win)
		return
	}
	if url == "" {
		dialog.ShowError(errors.New("enter the download URL"), w.win)
		return
	}

	w.deps.run("add "+name, func(ctx context.Context) error {
		return exec(func() (cmdrun.Result, error) {
			return w.deps.client().AddURL(ctx, name, url)
		})
	})
}

// AddFromGitHub adds a mod from a repository's GitHub releases.
func (w *Window) AddFromGitHub() {
	ref := widget.NewEntry()
	ref.SetPlaceHolder("owner/repository, or a full GitHub URL")

	branch := widget.NewEntry()
	branch.SetPlaceHolder("optional, defaults to the repository's own")

	pattern := widget.NewEntry()
	pattern.SetPlaceHolder("optional, for example .*fabric.*\\.jar")

	body := container.NewVBox(
		widgets.Muted("Repository"),
		ref,
		widgets.Muted("Branch"),
		branch,
		widgets.Muted("Asset pattern"),
		pattern,
		widgets.Note("packwiz reads the repository's releases and can keep the "+
			"mod updated from them. A pattern picks one asset when a release "+
			"carries several, which is common for mods that also publish a "+
			"sources or dev jar."),
	)

	size := widgets.FitDialog(w.win, tokens.FormWidth, tokens.AddModHeight)

	d := dialog.NewCustomConfirm("Add mod from GitHub", "Add", "Cancel",
		widgets.Scrollable(size.Width, size.Height,
			widgets.Inset(tokens.SpaceMD, tokens.SpaceMD, body)),
		func(add bool) {
			if add {
				w.runAddGitHub(ref.Text, branch.Text, pattern.Text)
			}
		}, w.win)

	d.Resize(size)
	d.Show()
}

// runAddGitHub validates and adds a GitHub released mod.
func (w *Window) runAddGitHub(ref, branch, pattern string) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		dialog.ShowError(errors.New("enter a repository, as owner/name or a URL"), w.win)
		return
	}

	w.deps.run("add "+ref, func(ctx context.Context) error {
		return exec(func() (cmdrun.Result, error) {
			return w.deps.client().AddGitHub(ctx,
				ref, strings.TrimSpace(branch), strings.TrimSpace(pattern))
		})
	})
}

// addMenu is the set of ways to add a mod, offered where the mod list is
// empty so a new pack has an obvious next step.
func (w *Window) addMenu() fyne.CanvasObject {
	browse := widget.NewButton("Browse Modrinth", w.FocusBrowse)
	browse.Importance = widget.HighImportance

	fromURL := widget.NewButton("From a URL", w.AddFromURL)
	fromGitHub := widget.NewButton("From GitHub", w.AddFromGitHub)

	return container.NewHBox(browse, fromURL, fromGitHub)
}
