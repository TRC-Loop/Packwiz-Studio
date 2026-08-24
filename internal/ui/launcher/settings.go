package launcher

import (
	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	fynetheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/PalisadeMC/Packwiz-Studio/internal/config"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/widgets"
)

// ShowSettings opens the app settings.
//
// These are app wide. Anything that belongs to one pack, such as its
// export folder or its remote, is not here: it is read from the pack or
// kept in that pack's own preferences.
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

	gitEnabled := widget.NewCheck("Enable git and releases", nil)
	gitEnabled.SetChecked(cfg.GitEnabled)

	deps := widget.NewCheck("Install dependencies with a mod", nil)
	deps.SetChecked(cfg.InstallDependencies)

	tags := widget.NewCheck("Categories", nil)
	tags.SetChecked(cfg.Browser.ShowTags)

	license := widget.NewCheck("Licence", nil)
	license.SetChecked(cfg.Browser.ShowLicense)

	sides := widget.NewCheck("Which sides it supports", nil)
	sides.SetChecked(cfg.Browser.ShowSides)

	form := container.NewVBox(
		widgets.SubHeading("packwiz"),
		container.NewBorder(nil, nil, nil, browse, binPath),
		widgets.Note("Where the packwiz binary lives. Leave it empty to use "+
			"whichever packwiz is on your PATH."),

		widgets.VSpace(tokens.SpaceLG),
		widgets.SubHeading("Mods"),
		deps,
		widgets.Note("packwiz asks whether to add the libraries a mod needs. "+
			"With this on the answer is yes, which is what a mod usually "+
			"requires to load at all."),
		widgets.VSpace(tokens.SpaceSM),
		widgets.Muted("Show for each mod in the browser"),
		tags,
		license,
		sides,

		widgets.VSpace(tokens.SpaceLG),
		widgets.SubHeading("Git"),
		gitEnabled,
		widgets.Note("Turn this off to keep the app out of your repository. "+
			"The git and releases sections disappear and no git command is "+
			"ever run, so you can manage the repo with another client. "+
			"Release API tokens live in the system keyring, one per host, "+
			"and are never written to the config file."),
	)

	size := widgets.FitDialog(w.win, tokens.SettingsWidth, tokens.SettingsHeight)

	d := dialog.NewCustomConfirm("Settings", "Save", "Cancel",
		widgets.Scrollable(size.Width, size.Height, form),
		func(save bool) {
			if !save {
				return
			}
			w.applySettings(settingsInput{
				packwizPath: binPath.Text,
				gitEnabled:  gitEnabled.Checked,
				installDeps: deps.Checked,
				browserPrefs: config.BrowserPrefs{
					ShowTags:    tags.Checked,
					ShowLicense: license.Checked,
					ShowSides:   sides.Checked,
				},
			})
		}, w.win)

	d.Resize(size)
	d.Show()
}

// settingsInput is what the settings dialog collected.
type settingsInput struct {
	packwizPath  string
	gitEnabled   bool
	installDeps  bool
	browserPrefs config.BrowserPrefs
}

// applySettings stores the settings, then re-resolves packwiz so a
// changed path takes effect immediately.
func (w *Window) applySettings(in settingsInput) {
	err := w.sess.Cfg.Update(func(c *config.Config) {
		c.PackwizPath = in.packwizPath
		c.GitEnabled = in.gitEnabled
		c.InstallDependencies = in.installDeps
		c.Browser = in.browserPrefs
	})
	if err != nil {
		dialog.ShowError(err, w.win)
		return
	}

	// An open pack window built its icon rail from the old settings, so it
	// has to hear about this.
	w.sess.ConfigChanged()

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
