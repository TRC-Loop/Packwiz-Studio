package launcher

import (
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/TRC-Loop/Packwiz-Studio/internal/packwiz"
	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/widgets"
)

// setBusy disables the actions while work is running.
func (s *setupScreen) setBusy(busy bool) {
	for _, b := range []*widget.Button{s.primary, s.secondary, s.skip} {
		if busy {
			b.Disable()
		} else {
			b.Enable()
		}
	}
}

// statusText reports a message in the status line, which is also where
// browseForBinary puts its failures.
func (s *setupScreen) statusText(text string, state widgets.State) {
	s.status.SetText(text)
	s.status.Importance = importanceFor(state)
	s.status.Show()
	s.status.Refresh()
}

// importanceFor maps a status state onto a Label importance, which is how
// a Label takes a colour.
func importanceFor(state widgets.State) widget.Importance {
	switch state {
	case widgets.StateError:
		return widget.DangerImportance
	case widgets.StateWarning:
		return widget.WarningImportance
	case widgets.StateSuccess:
		return widget.SuccessImportance
	default:
		return widget.LowImportance
	}
}

// finish handles the end of an install, whichever way it went.
func (s *setupScreen) finish(loc packwiz.Location, err error) {
	s.progress.Hide()
	s.spinner.Stop()
	s.spinner.Hide()
	s.setBusy(false)

	if err != nil {
		s.statusText(err.Error(), widgets.StateError)
		return
	}

	if err := s.win.sess.SetPackwiz(loc); err != nil {
		dialog.ShowError(err, s.win.win)
		return
	}

	// Recording the binary swaps the launcher over to its recents view, so
	// there is nothing more to report here.
	s.win.Refresh()
}
