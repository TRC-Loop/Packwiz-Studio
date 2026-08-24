// Command packwiz-studio is a desktop GUI for authoring, versioning and
// releasing Minecraft modpacks on top of the packwiz CLI.
package main

import (
	"context"

	"fyne.io/fyne/v2/app"

	"github.com/TRC-Loop/Packwiz-Studio/internal/config"
	"github.com/TRC-Loop/Packwiz-Studio/internal/logbus"
	"github.com/TRC-Loop/Packwiz-Studio/internal/studio"
	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/launcher"
	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/theme"
)

func main() {
	declareMetadata()

	a := app.NewWithID(appID)
	a.Settings().SetTheme(theme.New())

	cfg, err := config.Open()
	if err != nil {
		showFatal(a, err)
		return
	}

	sess := studio.New(a, cfg)

	// A missing binary is not fatal: the launcher shows a setup screen
	// for it. The reason still goes to the log, so a broken configured
	// path can be diagnosed rather than looking like a plain absence.
	if err := sess.ResolvePackwiz(context.Background()); err != nil {
		sess.Bus.Publish(logbus.KindNotice, err.Error())
	}

	win := a.NewWindow(appName)
	l := launcher.New(sess, win, appName)
	l.SetOnOpenPack(func(dir string) { openPack(l, dir) })
	l.Refresh()

	win.ShowAndRun()
}
