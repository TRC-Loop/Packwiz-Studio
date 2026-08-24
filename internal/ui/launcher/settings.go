package launcher

import (
	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	fynetheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/TRC-Loop/Packwiz-Studio/internal/config"
	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/tokens"
	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/widgets"
)

// showSettings opens the app settings. These are app wide rather than
// per pack, so they live in a dialog reachable from any window.
func (w *Window) ShowSettings() {
	cfg := w.sess.Cfg.Get()

	binPath := widget.NewEntry()
	binPath.SetText(cfg.PackwizPath)
	binPath.SetPlaceHolder("leave empty to look on PATH")

	browse := widget.NewButtonWithIcon("", fynetheme.FolderOpenIcon(), func() {
		open := dialog.NewFileOpen(func(rc fyne.URIReadCloser, err error) {
			if err != nil || rc == nil {
				return
			}
			binPath.SetText(rc.URI().Path())
			rc.Close()
		}, w.win)
		if dir, err := storage.ListerForURI(storage.NewFileURI(defaultBrowseDir())); err == nil {
			open.SetLocation(dir)
		}
		open.Show()
	})

	giteaURL := widget.NewEntry()
	giteaURL.SetText(cfg.GiteaBaseURL)
	giteaURL.SetPlaceHolder("https://git.example.com")

	useKeyring := widget.NewCheck("Store release tokens in the system keyring", nil)
	useKeyring.SetChecked(cfg.UseKeyring)

	gitEnabled := widget.NewCheck("Enable git and releases", nil)
	gitEnabled.SetChecked(cfg.GitEnabled)

	form := container.NewVBox(
		widgets.SubHeading("packwiz"),
		container.NewBorder(nil, nil, nil, browse, binPath),
		widgets.Dim("Where the packwiz binary lives."),

		widgets.VSpace(tokens.SpaceLG),
		widgets.SubHeading("Git"),
		gitEnabled,
		widgets.Dim("Turn this off to keep the app out of your repository. The git\n"+
			"and releases sections disappear, and no git command is ever run."),

		widgets.VSpace(tokens.SpaceLG),
		widgets.SubHeading("Releases"),
		useKeyring,
		widgets.Dim("Tokens are never written to the config file."),
		widgets.VSpace(tokens.SpaceSM),
		widgets.Muted("Self hosted Gitea or Forgejo API base"),
		giteaURL,
		widgets.Dim("Only needed when your remote host is not a public domain\n"+
			"the app recognises."),
	)

	d := dialog.NewCustomConfirm("Settings", "Save", "Cancel",
		container.NewVScroll(widgets.Inset(tokens.SpaceMD, tokens.SpaceMD, form)),
		func(save bool) {
			if !save {
				return
			}
			w.applySettings(settingsInput{
				packwizPath: binPath.Text,
				giteaURL:    giteaURL.Text,
				useKeyring:  useKeyring.Checked,
				gitEnabled:  gitEnabled.Checked,
			})
		}, w.win)

	d.Resize(fyne.NewSize(tokens.SettingsWidth, tokens.SettingsHeight))
	d.Show()
}

// settingsInput is what the settings dialog collected.
type settingsInput struct {
	packwizPath string
	giteaURL    string
	useKeyring  bool
	gitEnabled  bool
}

// applySettings stores the settings, then re-resolves packwiz so a
// changed path takes effect immediately.
func (w *Window) applySettings(in settingsInput) {
	err := w.sess.Cfg.Update(func(c *config.Config) {
		c.PackwizPath = in.packwizPath
		c.GiteaBaseURL = in.giteaURL
		c.UseKeyring = in.useKeyring
		c.GitEnabled = in.gitEnabled
	})
	if err != nil {
		dialog.ShowError(err, w.win)
		return
	}

	go func() {
		resolveErr := w.sess.ResolvePackwiz(context.Background())
		fyne.Do(func() {
			// A path the user typed by hand can be wrong. Report it, but
			// keep the setting: they may be mid-edit, and silently
			// reverting what they typed would be worse.
			if resolveErr != nil && in.packwizPath != "" {
				dialog.ShowError(resolveErr, w.win)
			}
			w.Refresh()
		})
	}()
}
