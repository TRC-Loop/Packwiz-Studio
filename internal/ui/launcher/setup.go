package launcher

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	fynetheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/TRC-Loop/Packwiz-Studio/internal/packwiz"
	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/tokens"
	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/widgets"
)

// setupScreen is the first-run screen shown until packwiz resolves.
//
// It is laid out as one clear decision rather than a wall of controls:
// a heading saying what is missing, a primary action that is filled in
// once the app knows what it can do on this machine, a secondary action
// for an existing binary, and one status line that carries everything
// from "checking" through progress to failure.
type setupScreen struct {
	win *Window

	primary   *widget.Button
	secondary *widget.Button
	skip      *widget.Button

	detail   *widget.Label
	progress *widget.ProgressBar
	spinner  *widget.ProgressBarInfinite
	status   *widget.Label
}

// buildSetup assembles the setup screen.
func (w *Window) buildSetup() fyne.CanvasObject {
	s := &setupScreen{
		win:      w,
		detail:   widgets.Note("Checking what this machine needs."),
		progress: widget.NewProgressBar(),
		spinner:  widget.NewProgressBarInfinite(),
		status:   widgets.Note(""),
	}

	s.progress.Hide()
	s.spinner.Hide()
	s.spinner.Stop()
	s.status.Hide()

	s.primary = widget.NewButtonWithIcon("Install packwiz",
		fynetheme.DownloadIcon(), s.install)
	s.primary.Importance = widget.HighImportance
	s.primary.Disable()

	s.secondary = widget.NewButtonWithIcon("Locate an existing binary",
		fynetheme.SearchIcon(), func() { s.win.browseForBinary(s.statusText) })

	s.skip = widget.NewButton("Skip for now", func() {
		w.setupDismissed = true
		w.Refresh()
	})
	s.skip.Importance = widget.LowImportance

	go s.describeRoute()

	return s.layout()
}

// layout arranges the screen. The actions sit under the explanation and
// the status area is reserved below them, so nothing jumps around when
// progress appears.
func (s *setupScreen) layout() fyne.CanvasObject {
	heading := widgets.Heading("Set up packwiz")

	intro := widgets.Note("Packwiz Studio does not manage packs itself. It " +
		"drives packwiz, the command line tool that owns your pack's files, " +
		"and packwiz is not on your PATH yet.")

	actions := container.NewVBox(
		container.NewHBox(s.primary, s.secondary),
		s.detail,
	)

	statusArea := container.NewVBox(
		s.progress,
		s.spinner,
		s.status,
	)

	body := container.NewVBox(
		heading,
		intro,
		widgets.VSpace(tokens.SpaceLG),
		actions,
		widgets.VSpace(tokens.SpaceMD),
		widgets.Hairline(),
		widgets.VSpace(tokens.SpaceMD),
		statusArea,
	)

	return container.NewBorder(
		nil,
		widgets.Inset(tokens.SpaceXXL, tokens.SpaceMD, container.NewHBox(s.skip)),
		nil, nil,
		widgets.Inset(tokens.SpaceXXL, tokens.SpaceXL, body),
	)
}

// describeRoute works out how packwiz can be obtained here and says so on
// the primary action, so the button never promises something it cannot do.
func (s *setupScreen) describeRoute() {
	method, err := packwiz.Plan(context.Background(), nil)

	fyne.Do(func() {
		switch {
		case err != nil:
			s.primary.Disable()
			s.detail.SetText(err.Error())

		case method == packwiz.MethodDownload:
			s.primary.SetText("Download packwiz")
			s.primary.Enable()
			s.detail.SetText("Downloads the prebuilt binary from packwiz's " +
				"latest build, into this app's own folder. Nothing else on " +
				"your system is touched.")

		default:
			s.primary.SetText("Build packwiz")
			s.primary.Enable()
			s.detail.SetText("packwiz publishes no prebuilt binary for this " +
				"architecture, so it will be compiled from source with Go. " +
				"That takes a minute and produces a native binary.")
		}
	})
}

// install obtains packwiz and reports progress.
func (s *setupScreen) install() {
	s.setBusy(true)
	s.statusText("Working. The output panel has the detail.", widgets.StateNeutral)

	// A source build reports no percentage, so an indeterminate bar runs
	// until a download reports a size or the work finishes.
	s.spinner.Show()
	s.spinner.Start()

	installer := &packwiz.Installer{
		Runner: s.win.sess.Runner,
		Bus:    s.win.sess.Bus,
	}

	go func() {
		loc, err := installer.Install(context.Background(), func(done, total int64) {
			fyne.Do(func() { s.showProgress(done, total) })
		})
		fyne.Do(func() { s.finish(loc, err) })
	}()
}

// showProgress switches from the indeterminate bar to a real one once a
// download reports its size.
func (s *setupScreen) showProgress(done, total int64) {
	if total <= 0 {
		s.statusText("Downloaded "+megabytes(done), widgets.StateNeutral)
		return
	}

	s.spinner.Stop()
	s.spinner.Hide()
	s.progress.Show()
	s.progress.SetValue(float64(done) / float64(total))

	s.statusText(fmt.Sprintf("Downloaded %s of %s", megabytes(done), megabytes(total)),
		widgets.StateNeutral)
}

// megabytes renders a byte count for a progress line.
func megabytes(n int64) string {
	return fmt.Sprintf("%.1f MB", float64(n)/(1<<20))
}
