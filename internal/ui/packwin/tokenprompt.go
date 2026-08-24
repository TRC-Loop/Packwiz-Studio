package packwin

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/TRC-Loop/Packwiz-Studio/internal/forge"
	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/tokens"
	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/widgets"
)

// tokenPromptHeight sizes the token dialog.
const tokenPromptHeight float32 = 300

// newPasswordEntry is the token field. It masks what is typed, since a
// token is a credential and the screen may be shared.
func newPasswordEntry() *widget.Entry {
	entry := widget.NewPasswordEntry()
	entry.SetPlaceHolder("paste an API token")
	return entry
}

// newRememberCheck offers to store the token, defaulting to whatever the
// keyring setting says.
func newRememberCheck(useKeyring bool) *widget.Check {
	check := widget.NewCheck("Remember this token in the system keyring", nil)
	check.SetChecked(useKeyring)
	return check
}

// tokenPrompt explains what the token needs to be able to do, which is
// the part that is easy to get wrong when creating one.
func tokenPrompt(host forge.Host, entry *widget.Entry, remember *widget.Check) fyne.CanvasObject {
	scope := widget.NewLabel(scopeHint(host.Kind))
	scope.Wrapping = fyne.TextWrapWord

	body := container.NewVBox(
		widgets.Muted("Publishing to "+host.Remote.Path()+" needs an API token."),
		scope,
		widgets.VSpace(tokens.SpaceSM),
		entry,
		remember,
		widgets.VSpace(tokens.SpaceXS),
		widgets.Dim("The token is never written to the config file and never logged."),
	)

	return widgets.Inset(tokens.SpaceMD, tokens.SpaceMD, body)
}

// scopeHint names the permission each host's token needs.
func scopeHint(kind forge.Kind) string {
	switch kind {
	case forge.KindGitHub:
		return "A fine grained token needs read and write access to this " +
			"repository's contents. A classic token needs the repo scope."
	case forge.KindGitLab:
		return "A project or personal access token needs the api scope."
	default:
		return "An access token needs write permission for repositories."
	}
}
