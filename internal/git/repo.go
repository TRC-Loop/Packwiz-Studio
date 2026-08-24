// Package git reads and manipulates the git repository a pack lives in.
//
// Status probes and user actions use different runners. A probe runs
// every time the pack window refreshes, so it stays out of the log
// drawer; an action the user asked for is logged like any other command.
package git

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/PalisadeMC/Packwiz-Studio/internal/cmdrun"
)

// Repo is one pack's repository.
type Repo struct {
	dir string
	// quiet runs status probes with no bus attached, so refreshing the
	// status bar does not fill the log drawer with git plumbing.
	quiet *cmdrun.Runner
	// loud runs what the user asked for, and is logged.
	loud *cmdrun.Runner
}

// New returns a Repo for dir. The runner is used for user actions; status
// probes get their own unlogged runner.
func New(dir string, loud *cmdrun.Runner) *Repo {
	return &Repo{dir: dir, quiet: cmdrun.New(nil), loud: loud}
}

// Dir reports the folder this Repo works in.
func (r *Repo) Dir() string { return r.dir }

// Available reports whether a git binary exists at all. Without one the
// app hides its git features exactly as if they had been turned off.
func Available() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

// IsRepo reports whether the pack folder is inside a work tree.
func (r *Repo) IsRepo(ctx context.Context) bool {
	res, err := r.probe(ctx, "rev-parse", "--is-inside-work-tree")
	return err == nil && res.OK() && strings.TrimSpace(res.Stdout) == "true"
}

// Init creates a repository in the pack folder. This is a user action, so
// it is logged.
func (r *Repo) Init(ctx context.Context) (cmdrun.Result, error) {
	return r.run(ctx, "init")
}

// Clone copies a remote repository into dest.
//
// It runs in dest's parent, because dest itself does not exist yet. Auth
// is left to git: the app assumes an SSH key or a credential helper is
// already set up, as it does for every other remote operation.
func Clone(ctx context.Context, runner *cmdrun.Runner, url, dest string) (cmdrun.Result, error) {
	parent := filepath.Dir(dest)

	return runner.Run(ctx, cmdrun.Spec{
		Name: "git",
		Args: []string{"clone", "--", url, dest},
		Dir:  parent,
	})
}

// probe runs a read-only command without logging it.
func (r *Repo) probe(ctx context.Context, args ...string) (cmdrun.Result, error) {
	return r.quiet.Run(ctx, cmdrun.Spec{Name: "git", Args: args, Dir: r.dir})
}

// run executes a git command and logs it.
func (r *Repo) run(ctx context.Context, args ...string) (cmdrun.Result, error) {
	return r.loud.Run(ctx, cmdrun.Spec{Name: "git", Args: args, Dir: r.dir})
}
