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
	w.panes = container.NewBorder(nil, nil, w.side.object(), nil, body)

	// The content and the output drawer share a draggable divider, so the
	// drawer can be sized to whatever is being read. The area is a stack
	// whose contents are swapped when the drawer opens and closes: a split
	// with a hidden child still reserves room for it.
	w.contentArea = container.NewStack(w.panes)

	return container.NewBorder(
		w.head, w.status.object(), w.rail.object(), nil,
		w.contentArea,
	)
}

// setDrawerOpen puts the drawer into the content area, or takes it out.
func (w *Window) setDrawerOpen(open bool) {
	if !open {
		w.contentArea.Objects = []fyne.CanvasObject{w.panes}
		w.contentArea.Refresh()
		return
	}

	split := container.NewVSplit(w.panes, w.drawer.object())
	split.SetOffset(w.drawerOffset())

	w.contentArea.Objects = []fyne.CanvasObject{split}
	w.contentArea.Refresh()
	w.split = split
}

// drawerOffset is how much of the height the content keeps when the
// drawer opens. It is remembered per pack, so a drawer dragged tall stays
// tall the next time it is opened.
func (w *Window) drawerOffset() float64 {
	if saved := w.deps.prefs().DrawerOffset; saved > 0 && saved < 1 {
		return saved
	}
	return defaultDrawerOffset
}

// defaultDrawerOffset leaves the drawer about a quarter of the height.
const defaultDrawerOffset = 0.75

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
