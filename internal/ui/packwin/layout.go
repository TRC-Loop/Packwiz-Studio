package packwin

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	"github.com/PalisadeMC/Packwiz-Studio/internal/pack"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/widgets"
)

// build assembles the window's panes: the title strip on top, the icon
// rail down the left, the list pane beside it, the detail area filling
// the rest, and the log drawer above the status bar.
func (w *Window) build() *fyne.Container {
	bg := canvas.NewRectangle(tokens.ColorBG)
	body := container.NewStack(bg, w.main)

	// The side panel keeps its width while the detail area takes the
	// slack, so widening the window widens the content rather than the
	// list.
	panes := container.NewBorder(nil, nil, w.side.object(), nil, body)

	foot := container.NewVBox(w.drawer.object(), w.status.object())

	return container.NewBorder(
		w.head, foot, w.rail.object(), nil,
		panes,
	)
}

// header names the pack and shows its image.
//
// The whole title block opens the pack's properties. That is where a user
// looks to rename a pack or give it an image, so it is made clickable
// rather than living only in a menu.
func (w *Window) header() fyne.CanvasObject {
	bg := canvas.NewRectangle(tokens.ColorSurface)

	title := container.NewHBox()
	if logo := widgets.PackLogo(pack.IconPath(w.pack.Dir), tokens.IconPackLogo); logo != nil {
		title.Add(logo)
	}
	title.Add(container.NewCenter(widgets.SubHeading(w.pack.Name)))

	clickable := widgets.NewClickable(title, w.EditProperties)

	line := container.NewBorder(nil, nil,
		widgets.Inset(tokens.SpaceSM, tokens.SpaceXS, clickable), nil, nil)

	return container.NewStack(bg,
		container.NewBorder(nil, widgets.Hairline(), nil, nil, line))
}
