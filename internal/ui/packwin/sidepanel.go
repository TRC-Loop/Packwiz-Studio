package packwin

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/widgets"
)

// sidePanel is the list pane between the rail and the detail area. Its
// header names the current activity, which is how the icon-only rail
// stays legible once an activity is open.
type sidePanel struct {
	title   *canvas.Text
	body    *fyne.Container
	root    *fyne.Container
	sizer   *canvas.Rectangle
	visible bool
}

func newSidePanel() *sidePanel {
	p := &sidePanel{
		title:   widgets.Caption(""),
		body:    container.NewStack(),
		visible: true,
	}
	p.title.Color = tokens.ColorMuted

	// An invisible rectangle fixes the pane's width. A border layout
	// gives its side child that child's minimum size, so the width comes
	// from here rather than from whatever list is loaded.
	p.sizer = canvas.NewRectangle(nil)
	p.sizer.SetMinSize(fyne.NewSize(tokens.SidePanelWidth, 0))

	bg := canvas.NewRectangle(tokens.ColorSurface)

	header := widgets.Inset(tokens.SpaceMD, tokens.SpaceSM, p.title)
	inner := container.NewBorder(header, nil, nil, nil, p.body)

	p.root = container.NewStack(
		bg,
		p.sizer,
		container.NewBorder(nil, nil, nil, widgets.Hairline(), inner),
	)
	return p
}

// object returns the pane for placement.
func (p *sidePanel) object() fyne.CanvasObject { return p.root }

// set replaces the pane's title and content. A nil content collapses the
// pane, since an activity that wants the whole window has no list.
func (p *sidePanel) set(title string, content fyne.CanvasObject) {
	p.title.Text = title
	p.title.Refresh()

	p.body.Objects = nil
	if content != nil {
		p.body.Objects = []fyne.CanvasObject{content}
	}
	p.body.Refresh()

	if content == nil {
		p.root.Hide()
		return
	}
	if p.visible {
		p.root.Show()
	}
}

// toggle hides or shows the pane, widening the detail area.
func (p *sidePanel) toggle() {
	p.visible = !p.visible
	if p.visible {
		p.root.Show()
	} else {
		p.root.Hide()
	}
}
