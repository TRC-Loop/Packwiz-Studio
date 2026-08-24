package packwin

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"sync/atomic"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"

	"github.com/TRC-Loop/Packwiz-Studio/internal/cmdrun"
	"github.com/TRC-Loop/Packwiz-Studio/internal/pack"
	"github.com/TRC-Loop/Packwiz-Studio/internal/packwiz"
	"github.com/TRC-Loop/Packwiz-Studio/internal/studio"
)

// activityDeps is what every activity needs: the session, the window to
// parent dialogs on, the pack, and a way to run a packwiz command without
// two of them overlapping.
type activityDeps struct {
	sess *studio.Session
	win  fyne.Window
	pack pack.Pack

	// onPackChanged is called after a command that altered the pack, so
	// the window can reload its activities and its status bar.
	onPackChanged func()

	// busy guards against a second command starting while one is still
	// running. packwiz rewrites the index, and two writers would leave it
	// inconsistent.
	busy atomic.Bool
}

// client returns a packwiz client for this pack.
func (d *activityDeps) client() *packwiz.Client {
	return d.sess.Client(d.pack.Dir)
}

// errBusy reports a command refused because another is still running.
var errBusy = errors.New("another packwiz command is still running")

// exec runs one packwiz call and turns a refusal into an error.
//
// A non-zero exit is not an error at the runner level, because a packwiz
// command declining to do something is a normal outcome. At this level it
// is: the user asked for something and did not get it. packwiz's own
// wording is passed through, since it explains the refusal better than
// anything the app could infer from an exit code.
func exec(call func() (cmdrun.Result, error)) error {
	res, err := call()
	if err != nil {
		return err
	}
	if res.OK() {
		return nil
	}

	if msg := strings.TrimSpace(res.Output()); msg != "" {
		return errors.New(msg)
	}
	return errors.New("packwiz exited with status " + strconv.Itoa(res.ExitCode))
}

// run performs work off the main goroutine, then reports and reloads.
//
// Every pack mutating action goes through here, so the busy guard, the
// error dialog and the post-command reload live in one place rather than
// being repeated per button.
func (d *activityDeps) run(label string, work func(context.Context) error) {
	if !d.busy.CompareAndSwap(false, true) {
		dialog.ShowError(errBusy, d.win)
		return
	}

	go func() {
		err := work(context.Background())

		fyne.Do(func() {
			d.busy.Store(false)

			if err != nil {
				dialog.ShowError(errors.New(label+" failed: "+err.Error()), d.win)
			}
			if d.onPackChanged != nil {
				d.onPackChanged()
			}
		})
	}()
}

// run on the mods activity defers to the shared runner. The reload the
// window performs afterwards covers this activity too.
func (a *modsActivity) run(label string, work func(context.Context) error) {
	a.deps.run(label, work)
}
