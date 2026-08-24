package packwin

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/widgets"
)

// showEmpty is the detail pane with no mod selected.
//
// It carries the ways to add one, so a new pack has an obvious next step
// rather than an empty panel and a hint to look elsewhere.
func (a *modsActivity) showEmpty() {
	body := container.NewVBox(
		widgets.SubHeading(a.summary()),
		widgets.Note("Mods come from Modrinth, from any direct download link, "+
			"or from a GitHub repository's releases."),
		widgets.VSpace(tokens.SpaceMD),
		a.deps.addMenu(),
	)

	if len(a.mods) > 0 {
		body.Add(widgets.VSpace(tokens.SpaceLG))
		body.Add(widgets.Note("Select a mod on the left to see its version, " +
			"its source and which side it installs on."))
	}

	a.main.Objects = []fyne.CanvasObject{
		widgets.Inset(tokens.SpaceXL, tokens.SpaceLG, body),
	}
	a.main.Refresh()
}

// summary heads the empty pane.
func (a *modsActivity) summary() string {
	switch len(a.mods) {
	case 0:
		return "No mods yet"
	case 1:
		return "1 mod installed"
	default:
		return itoa(len(a.mods)) + " mods installed"
	}
}
