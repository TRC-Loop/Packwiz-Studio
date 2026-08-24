package packwin

import (
	"context"

	"fyne.io/fyne/v2/dialog"

	"github.com/TRC-Loop/Packwiz-Studio/internal/sysopen"
)

// gitActivityOrNil finds the git activity, which is absent when the git
// integration is off.
func (w *Window) gitActivityOrNil() *gitActivity {
	for _, a := range w.items {
		if g, ok := a.(*gitActivity); ok {
			return g
		}
	}
	return nil
}

// GitEnabled reports whether this window offers git features at all, so
// the menubar knows whether to build its Git menu.
func (w *Window) GitEnabled() bool { return w.gitEnabled() }

// GitInit creates a repository in the pack folder.
func (w *Window) GitInit() {
	g := w.gitActivityOrNil()
	if g == nil {
		return
	}
	w.Select(ActivityGit)
	g.runGit("git init", func(ctx context.Context) error {
		return execGit(func() (result, error) { return g.repo.Init(ctx) })
	})
}

// GitStageAll stages every change.
func (w *Window) GitStageAll() {
	g := w.gitActivityOrNil()
	if g == nil {
		return
	}
	if !g.status.IsRepo {
		dialog.ShowError(errNoRepo, w.win)
		return
	}
	g.runGit("stage all", func(ctx context.Context) error {
		return execGit(func() (result, error) { return g.repo.StageAll(ctx) })
	})
}

// GitCommit opens the git activity so the message can be typed. Committing
// needs a message, so the menu item takes the user to the field rather
// than committing something empty.
func (w *Window) GitCommit() {
	if w.gitActivityOrNil() == nil {
		return
	}
	w.Select(ActivityGit)

	if g := w.gitActivityOrNil(); g != nil {
		w.win.Canvas().Focus(g.message)
	}
}

// GitPush pushes the current branch, behind the same confirmation the
// button uses.
func (w *Window) GitPush() {
	g := w.gitActivityOrNil()
	if g == nil {
		return
	}
	if !g.status.IsRepo {
		dialog.ShowError(errNoRepo, w.win)
		return
	}
	g.confirmPush()
}

// GitPull fetches and merges from origin.
func (w *Window) GitPull() {
	g := w.gitActivityOrNil()
	if g == nil {
		return
	}
	if !g.status.IsRepo {
		dialog.ShowError(errNoRepo, w.win)
		return
	}
	g.runGit("pull", func(ctx context.Context) error {
		return execGit(func() (result, error) { return g.repo.Pull(ctx) })
	})
}

// OpenRemote opens the repository's page on its host.
func (w *Window) OpenRemote() {
	g := w.gitActivityOrNil()
	if g == nil {
		return
	}
	if g.status.Remote == "" {
		dialog.ShowError(errNoRemote, w.win)
		return
	}

	url, err := remoteWebURL(g.status.Remote)
	if err != nil {
		dialog.ShowError(err, w.win)
		return
	}
	if err := sysopen.Browse(url); err != nil {
		dialog.ShowError(err, w.win)
	}
}
