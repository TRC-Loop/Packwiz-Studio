package launcher

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	"github.com/TRC-Loop/Packwiz-Studio/internal/studio"
	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/tokens"
	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/widgets"
)

// statusBar is the launcher's foot: one line reporting where packwiz is,
// or that it is missing.
type statusBar struct {
	sess *studio.Session
	text *canvas.Text
	root *fyne.Container
}

func newStatusBar(sess *studio.Session) *statusBar {
	s := &statusBar{sess: sess, text: widgets.Caption("")}

	bg := canvas.NewRectangle(tokens.ColorSurface)
	line := widgets.Inset(tokens.SpaceMD, tokens.SpaceSM, s.text)

	s.root = container.NewStack(bg, container.NewBorder(widgets.Hairline(), nil, nil, nil, line))
	s.refresh()
	return s
}

// object returns the bar for placement in a window.
func (s *statusBar) object() fyne.CanvasObject { return s.root }

// refresh restates the packwiz situation. It must run on the main
// goroutine.
func (s *statusBar) refresh() {
	loc := s.sess.Packwiz()

	if loc.Path == "" {
		s.text.Text = "packwiz not found, set it in Settings"
		s.text.Color = tokens.ColorWarning
	} else {
		s.text.Text = describe(loc.Version) + loc.Path
		s.text.Color = tokens.ColorMuted
	}
	s.text.Refresh()
}

// describe renders the version prefix, which is absent for a binary that
// reports no usable version string.
func describe(version string) string {
	if version == "" {
		return "packwiz at "
	}
	return "packwiz " + version + " at "
}
