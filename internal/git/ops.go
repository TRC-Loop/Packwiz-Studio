package git

import (
	"context"
	"strings"

	"github.com/TRC-Loop/Packwiz-Studio/internal/cmdrun"
)

// Change is one path that differs from HEAD.
type Change struct {
	// Path is relative to the repository root.
	Path string
	// Status is git's two letter porcelain code, for example " M" or "??".
	Status string
	// Staged reports an entry with something in the index.
	Staged bool
}

// Label describes the change in words, for the changed files list.
func (c Change) Label() string {
	switch {
	case strings.HasPrefix(c.Status, "??"):
		return "new"
	case strings.ContainsRune(c.Status, 'D'):
		return "deleted"
	case strings.ContainsRune(c.Status, 'R'):
		return "renamed"
	case strings.ContainsRune(c.Status, 'A'):
		return "added"
	default:
		return "modified"
	}
}

// Changes lists what differs from HEAD. This is a probe, so it is not
// logged.
func (r *Repo) Changes(ctx context.Context) ([]Change, error) {
	res, err := r.probe(ctx, "status", "--porcelain")
	if err != nil {
		return nil, err
	}

	var out []Change
	for _, line := range strings.Split(res.Stdout, "\n") {
		if len(line) < 4 {
			continue
		}

		status := line[:2]
		path := strings.TrimSpace(line[3:])

		// A rename is reported as "old -> new". The new path is the one
		// worth showing.
		if i := strings.Index(path, " -> "); i >= 0 {
			path = path[i+4:]
		}

		out = append(out, Change{
			Path:   path,
			Status: status,
			Staged: status[0] != ' ' && status[0] != '?',
		})
	}
	return out, nil
}

// StageAll stages every change, including new files.
func (r *Repo) StageAll(ctx context.Context) (cmdrun.Result, error) {
	return r.run(ctx, "add", "--all")
}

// Stage stages one path.
func (r *Repo) Stage(ctx context.Context, path string) (cmdrun.Result, error) {
	return r.run(ctx, "add", "--", path)
}

// Unstage removes one path from the index, leaving the working tree
// alone.
func (r *Repo) Unstage(ctx context.Context, path string) (cmdrun.Result, error) {
	return r.run(ctx, "restore", "--staged", "--", path)
}

// Commit records the staged changes.
func (r *Repo) Commit(ctx context.Context, message string) (cmdrun.Result, error) {
	return r.run(ctx, "commit", "-m", message)
}

// Push sends the current branch to origin.
//
// The upstream is set on the first push of a branch, so pushing a new
// branch works without the user having to know about it.
func (r *Repo) Push(ctx context.Context, branch string, setUpstream bool) (cmdrun.Result, error) {
	args := []string{"push"}
	if setUpstream {
		args = append(args, "--set-upstream")
	}
	args = append(args, "origin")
	if branch != "" {
		args = append(args, branch)
	}
	return r.run(ctx, args...)
}

// Pull fetches and merges from origin.
func (r *Repo) Pull(ctx context.Context) (cmdrun.Result, error) {
	return r.run(ctx, "pull")
}

// HasUpstream reports whether the current branch tracks a remote branch,
// which decides whether a push needs to set one.
func (r *Repo) HasUpstream(ctx context.Context) bool {
	res, err := r.probe(ctx, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{upstream}")
	return err == nil && res.OK()
}

// Tags lists the repository's tags, newest first, which is the order a
// changelog wants to compare against.
func (r *Repo) Tags(ctx context.Context) ([]string, error) {
	res, err := r.probe(ctx, "tag", "--sort=-creatordate")
	if err != nil {
		return nil, err
	}

	var out []string
	for _, line := range strings.Split(res.Stdout, "\n") {
		if tag := strings.TrimSpace(line); tag != "" {
			out = append(out, tag)
		}
	}
	return out, nil
}

// Show reads a file's contents at a given revision, which is how a
// changelog sees what the pack looked like at the previous release.
func (r *Repo) Show(ctx context.Context, rev, path string) (string, error) {
	res, err := r.probe(ctx, "show", rev+":"+path)
	if err != nil {
		return "", err
	}
	if !res.OK() {
		return "", errMissingAtRev{rev: rev, path: path}
	}
	return res.Stdout, nil
}

// errMissingAtRev reports a path that did not exist at a revision, which
// for a changelog means the file was added since.
type errMissingAtRev struct {
	rev  string
	path string
}

func (e errMissingAtRev) Error() string {
	return e.path + " does not exist at " + e.rev
}
