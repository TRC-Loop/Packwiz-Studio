package packwin

import (
	"context"

	"github.com/TRC-Loop/Packwiz-Studio/internal/cmdrun"
)

// reloader is an activity that rebuilds itself from disk. Not every
// activity needs to: a static screen has nothing to reload.
type reloader interface {
	Reload()
}

// Reload re-reads the pack and rebuilds every activity that depends on
// it, then restates the status bar. This runs after any command that
// could have changed the pack.
func (w *Window) Reload() {
	for _, a := range w.items {
		if r, ok := a.(reloader); ok {
			r.Reload()
		}
	}

	// The open activity's panes are the ones on screen, so they are
	// reinstalled to pick up the rebuilt content.
	w.selectActivity(w.current)
	w.RefreshStatus()
}

// RefreshIndex runs packwiz refresh, which rebuilds the index after files
// changed underneath it.
func (w *Window) RefreshIndex() {
	w.deps.run("refresh", func(ctx context.Context) error {
		return exec(func() (cmdrun.Result, error) {
			return w.deps.client().Refresh(ctx)
		})
	})
}

// CheckUpdates updates every mod in the pack. This is the manual check:
// the app never runs it on a timer.
func (w *Window) CheckUpdates() {
	w.deps.run("check for updates", func(ctx context.Context) error {
		return exec(func() (cmdrun.Result, error) {
			return w.deps.client().UpdateAll(ctx)
		})
	})
}
