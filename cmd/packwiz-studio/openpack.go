package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/TRC-Loop/Packwiz-Studio/internal/pack"
	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/launcher"
	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/tokens"
	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/widgets"
)

// openPack replaces the launcher's contents with the pack view, reusing
// the same window.
//
// The pack window itself is not built yet: this is a placeholder standing
// in for it so the launcher's open flow can be exercised end to end. It
// goes away when the pack window shell lands.
func openPack(l *launcher.Window, dir string) {
	win := l.Window()

	p, err := pack.Load(dir)
	if err != nil {
		dialog.ShowError(err, win)
		return
	}

	title := container.NewHBox(
		widgets.PackLogo(pack.IconPath(p.Dir), tokens.IconPackLogo),
		widgets.Heading(p.Name),
	)

	body := container.NewVBox(
		title,
		widgets.Muted(loaderLine(p)),
		widgets.Dim(p.Dir),
		widgets.VSpace(tokens.SpaceLG),
		widgets.Muted("The pack window is not built yet."),
		widgets.VSpace(tokens.SpaceMD),
		container.NewHBox(widget.NewButton("Back to launcher", l.Install)),
	)

	win.SetTitle(appName + ": " + p.Name)
	win.Resize(fyne.NewSize(tokens.PackWindowWidth, tokens.PackWindowHeight))
	win.SetContent(widgets.Inset(tokens.SpaceXL, tokens.SpaceLG, body))
}

// loaderLine describes a pack's Minecraft and loader versions, skipping
// whichever of them the pack.toml left out.
func loaderLine(p pack.Pack) string {
	line := p.MCVersion
	if p.Loader == "" {
		return line
	}
	if line != "" {
		line += " with "
	}
	line += p.Loader
	if p.LoaderVersion != "" {
		line += " " + p.LoaderVersion
	}
	return line
}
