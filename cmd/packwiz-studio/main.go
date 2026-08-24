// Command packwiz-studio is a desktop GUI for authoring, versioning and
// releasing Minecraft modpacks on top of the packwiz CLI.
package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"

	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/theme"
	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/tokens"
)

func main() {
	declareMetadata()

	a := app.NewWithID(appID)
	a.Settings().SetTheme(theme.New())

	w := a.NewWindow(appName)
	w.SetContent(widget.NewLabel(appName))
	w.Resize(fyne.NewSize(tokens.LauncherWidth, tokens.LauncherHeight))
	w.CenterOnScreen()
	w.ShowAndRun()
}
