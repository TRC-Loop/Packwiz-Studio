package forge

import (
	"context"
	"errors"
)

// Release is a release to publish.
type Release struct {
	// Tag is the git tag the release points at. It must already exist on
	// the remote: the app pushes tags with git rather than asking the API
	// to create them, so the release always matches what was pushed.
	Tag string
	// Title is the release name. An empty title makes hosts fall back to
	// the tag.
	Title string
	// Notes is the changelog body, in markdown.
	Notes string
	// Draft keeps the release unpublished, where the host supports it.
	Draft bool
	// Prerelease marks the release as not production ready.
	Prerelease bool
}

// Published is what a host reports back about a created release.
type Published struct {
	// ID is the host's identifier for the release, needed to attach
	// assets afterwards.
	ID string
	// URL is the release's page, for opening in a browser.
	URL string
}

// Asset is a file to attach to a release.
type Asset struct {
	// Name is the filename the asset is published under.
	Name string
	// Path is the local file to upload.
	Path string
}

// Client publishes releases to one host.
type Client interface {
	// CreateRelease publishes a release and returns what the host made.
	CreateRelease(ctx context.Context, r Release) (Published, error)
	// UploadAsset attaches a file to a published release.
	UploadAsset(ctx context.Context, release Published, a Asset) error
}

// ErrNoToken reports a release attempted without an API token.
var ErrNoToken = errors.New("no API token for this host")

// NewClient returns a client for a host.
func NewClient(host Host, token string) (Client, error) {
	if token == "" {
		return nil, ErrNoToken
	}

	switch host.Kind {
	case KindGitHub:
		return &githubClient{host: host, token: token, http: defaultHTTP()}, nil
	case KindGitLab:
		return &gitlabClient{host: host, token: token, http: defaultHTTP()}, nil
	default:
		return &giteaClient{host: host, token: token, http: defaultHTTP()}, nil
	}
}
