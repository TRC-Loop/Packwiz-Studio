package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/tokens"
	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/widgets"
)

// showFatal reports a problem the app cannot start without, in a window
// of its own. A config file that will not parse is the case this exists
// for: the app refuses to overwrite it, so the user is told where it is
// and left to deal with it.
func showFatal(a fyne.App, err error) {
	win := a.NewWindow(appName)

	message := widget.NewLabel(err.Error())
	message.Wrapping = fyne.TextWrapWord

	body := container.NewVBox(
		widgets.Heading("Packwiz Studio cannot start"),
		message,
		widgets.VSpace(tokens.SpaceMD),
		container.NewHBox(widget.NewButton("Quit", a.Quit)),
	)

	win.SetContent(widgets.Inset(tokens.SpaceXL, tokens.SpaceLG, body))
	win.Resize(fyne.NewSize(tokens.LauncherWidth, tokens.LauncherHeight))
	win.CenterOnScreen()
	win.ShowAndRun()
}
