package git

import (
	"context"

	"github.com/TRC-Loop/Packwiz-Studio/internal/cmdrun"
)

// TagExists reports whether a tag is present locally.
func (r *Repo) TagExists(ctx context.Context, tag string) bool {
	res, err := r.probe(ctx, "rev-parse", "--verify", "refs/tags/"+tag)
	return err == nil && res.OK()
}

// CreateTag makes an annotated tag at HEAD.
//
// Annotated rather than lightweight: a release is a marked point in the
// history, and annotated tags carry the author and date that makes it one.
func (r *Repo) CreateTag(ctx context.Context, tag, message string) (cmdrun.Result, error) {
	if message == "" {
		message = tag
	}
	return r.run(ctx, "tag", "--annotate", tag, "--message", message)
}

// PushTag sends one tag to origin.
//
// A release points at a tag on the remote, so the tag has to be pushed
// before the host is asked to create a release for it. Otherwise the host
// either refuses or creates the tag itself at whatever its default branch
// happens to be, which is not necessarily what was released.
func (r *Repo) PushTag(ctx context.Context, tag string) (cmdrun.Result, error) {
	return r.run(ctx, "push", "origin", "refs/tags/"+tag)
}

// RemoteTagExists reports whether origin already has the tag, which
// decides whether pushing it is needed.
func (r *Repo) RemoteTagExists(ctx context.Context, tag string) bool {
	res, err := r.probe(ctx, "ls-remote", "--tags", "origin", "refs/tags/"+tag)
	return err == nil && res.OK() && len(res.Stdout) > 0
}
