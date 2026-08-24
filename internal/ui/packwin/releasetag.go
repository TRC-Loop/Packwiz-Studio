package packwin

import (
	"context"
)

// ensureTag makes sure the release's tag exists locally and on the remote
// before the host is asked to publish it.
//
// A release is a pointer to a tag. If the tag is not on the remote, the
// host will either refuse the release or create the tag itself at
// whatever its default branch currently is, which is not necessarily what
// was tested and exported. Creating and pushing it here means the release
// always points at the commit the user was working on.
func (a *releasesActivity) ensureTag(ctx context.Context, tag, title string) error {
	if !a.repo.TagExists(ctx, tag) {
		a.notice("creating tag " + tag)

		if err := execGit(func() (result, error) {
			return a.repo.CreateTag(ctx, tag, tagMessage(tag, title))
		}); err != nil {
			return err
		}
	}

	if a.repo.RemoteTagExists(ctx, tag) {
		return nil
	}

	a.notice("pushing tag " + tag)
	return execGit(func() (result, error) { return a.repo.PushTag(ctx, tag) })
}

// tagMessage is the annotation put on a created tag. The release title
// makes a better annotation than the tag name repeated, when there is one.
func tagMessage(tag, title string) string {
	if title != "" {
		return title
	}
	return tag
}
