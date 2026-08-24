package packwin

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	fynetheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/TRC-Loop/Packwiz-Studio/internal/git"
	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/tokens"
	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/widgets"
)

// statusBar is the window's foot: git state on the left, packwiz state in
// the middle, the log drawer toggle on the right.
//
// The git half is absent entirely when the git integration is off, rather
// than showing an empty slot.
type statusBar struct {
	gitText     *canvas.Text
	packwizText *canvas.Text
	toggle      *widget.Button
	root        *fyne.Container
}

// newStatusBar builds the bar. onToggleLog opens or closes the drawer.
func newStatusBar(showGit bool, onToggleLog func()) *statusBar {
	s := &statusBar{
		gitText:     widgets.Caption(""),
		packwizText: widgets.Caption(""),
	}

	s.toggle = widget.NewButtonWithIcon("Output", fynetheme.MenuExpandIcon(), onToggleLog)
	s.toggle.Importance = widget.LowImportance

	left := container.NewHBox()
	if showGit {
		left.Add(widgets.Inset(tokens.SpaceMD, tokens.SpaceXS, s.gitText))
	}
	left.Add(widgets.Inset(tokens.SpaceMD, tokens.SpaceXS, s.packwizText))

	bg := canvas.NewRectangle(tokens.ColorSurface)
	line := container.NewBorder(nil, nil, left, s.toggle, nil)

	s.root = container.NewStack(
		bg,
		container.NewBorder(widgets.Hairline(), nil, nil, nil, line),
	)
	return s
}

// object returns the bar for placement.
func (s *statusBar) object() fyne.CanvasObject { return s.root }

// setGit restates the repository. It must run on the main goroutine.
func (s *statusBar) setGit(st git.Status) {
	s.gitText.Text = st.Label()

	switch {
	case !st.IsRepo:
		s.gitText.Color = tokens.ColorDim
	case st.Clean():
		s.gitText.Color = tokens.ColorSuccess
	default:
		s.gitText.Color = tokens.ColorWarning
	}
	s.gitText.Refresh()
}

// setPackwiz restates the tool. It must run on the main goroutine.
func (s *statusBar) setPackwiz(version string, ok bool) {
	if !ok {
		s.packwizText.Text = "packwiz not found"
		s.packwizText.Color = tokens.ColorError
	} else if version == "" {
		s.packwizText.Text = "packwiz ready"
		s.packwizText.Color = tokens.ColorMuted
	} else {
		s.packwizText.Text = "packwiz " + version
		s.packwizText.Color = tokens.ColorMuted
	}
	s.packwizText.Refresh()
}

// setLogOpen points the toggle's arrow the right way.
func (s *statusBar) setLogOpen(open bool) {
	if open {
		s.toggle.SetIcon(fynetheme.MenuDropDownIcon())
	} else {
		s.toggle.SetIcon(fynetheme.MenuExpandIcon())
	}
}
