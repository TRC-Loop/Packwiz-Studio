package packwin

import (
	"context"
	"errors"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2/dialog"

	"github.com/PalisadeMC/Packwiz-Studio/internal/config"
	"github.com/PalisadeMC/Packwiz-Studio/internal/forge"
	"github.com/PalisadeMC/Packwiz-Studio/internal/logbus"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/widgets"
)

// formatLabels are the changelog formats as the form offers them.
var formatLabels = []string{"Flat list", "Grouped by change", "Prose"}

// labelForFormat and formatForLabel map between the stored value and the
// label, so the config keeps a stable string while the UI can be reworded.
func labelForFormat(f config.ChangelogFormat) string {
	switch f {
	case config.FormatFlat:
		return formatLabels[0]
	case config.FormatProse:
		return formatLabels[2]
	default:
		return formatLabels[1]
	}
}

func formatForLabel(label string) config.ChangelogFormat {
	switch label {
	case formatLabels[0]:
		return config.FormatFlat
	case formatLabels[2]:
		return config.FormatProse
	default:
		return config.FormatGrouped
	}
}

// publish creates the release and attaches its files.
func (a *releasesActivity) publish(f *releaseForm) {
	tag := strings.TrimSpace(f.tag.Text)
	if tag == "" {
		dialog.ShowError(errors.New("enter a tag for this release"), a.deps.win)
		return
	}

	token, err := a.secrets.Get(a.host.TokenKey())
	if err != nil {
		a.askForToken(f, tag)
		return
	}
	a.doPublish(f, tag, token)
}

// doPublish performs the API calls off the main goroutine.
func (a *releasesActivity) doPublish(f *releaseForm, tag, token string) {
	// The token is registered with the log bus so that if it ever appears
	// in command output or an error message it is masked rather than
	// written out.
	a.deps.sess.Bus.Protect(token)

	release := forge.Release{
		Tag:        tag,
		Title:      strings.TrimSpace(f.title.Text),
		Notes:      f.notes.Text,
		Draft:      f.draft.Checked,
		Prerelease: f.prerelease.Checked,
	}

	assets := a.assetsFor(f)

	a.deps.run("publish release", func(ctx context.Context) error {
		client, err := forge.NewClient(a.host, token)
		if err != nil {
			return err
		}

		if err := a.ensureTag(ctx, tag, release.Title); err != nil {
			return err
		}

		a.notice("creating release " + tag + " on " + a.host.Kind.Name())

		published, err := client.CreateRelease(ctx, release)
		if err != nil {
			return err
		}

		for _, asset := range assets {
			a.notice("uploading " + asset.Name)
			if err := client.UploadAsset(ctx, published, asset); err != nil {
				// The release exists at this point, so a failed upload is
				// reported without pretending the whole thing failed.
				return errors.New("the release was created but " + asset.Name +
					" could not be uploaded: " + err.Error())
			}
		}

		a.notice("published " + tag)
		return nil
	})
}

// assetsFor collects the files to attach: the exported pack, then any
// images the user added.
func (a *releasesActivity) assetsFor(f *releaseForm) []forge.Asset {
	var assets []forge.Asset

	if f.attach.Checked {
		if export := a.deps.prefs().LastExport; export != "" {
			assets = append(assets, forge.Asset{
				Name: filepath.Base(export),
				Path: export,
			})
		}
	}

	for _, img := range f.images {
		assets = append(assets, forge.Asset{Name: filepath.Base(img), Path: img})
	}
	return assets
}

// askForToken prompts for an API token, storing it in the keyring when
// the user allows it. Tokens never go into the config file.
func (a *releasesActivity) askForToken(f *releaseForm, tag string) {
	entry := newPasswordEntry()
	remember := newRememberCheck(a.deps.sess.Cfg.Get().UseKeyring)

	body := tokenPrompt(a.host, entry, remember)

	size := widgets.FitDialog(a.deps.win, tokens.FormWidth, tokenPromptHeight)

	d := dialog.NewCustomConfirm("API token for "+a.host.Remote.Host,
		"Continue", "Cancel", body,
		func(ok bool) {
			if !ok {
				return
			}

			token := strings.TrimSpace(entry.Text)
			if token == "" {
				dialog.ShowError(errors.New("enter a token, or cancel"), a.deps.win)
				return
			}

			if remember.Checked {
				if err := a.secrets.Set(a.host.TokenKey(), token); err != nil {
					// Failing to store the token does not stop the
					// release: it just will not be remembered.
					a.notice("could not save the token: " + err.Error())
				}
			}
			a.doPublish(f, tag, token)
		}, a.deps.win)

	d.Resize(size)
	d.Show()
}

// notice writes app commentary to the log drawer, which is where the
// release's progress is visible since the API calls are not commands.
func (a *releasesActivity) notice(text string) {
	a.deps.sess.Bus.Publish(logbus.KindNotice, text)
}

// ForgetToken removes this host's stored token, which is how a wrong or
// expired token gets replaced.
func (a *releasesActivity) ForgetToken() error {
	return a.secrets.Delete(a.host.TokenKey())
}
