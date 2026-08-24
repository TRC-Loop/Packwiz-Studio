// Package forge works out which code host a repository lives on and
// talks to that host's release API.
package forge

import (
	"errors"
	"net/url"
	"strings"
)

// Remote is a git remote URL broken into the parts an API call needs.
type Remote struct {
	// Host is the bare hostname, for example github.com.
	Host string
	// Owner is the user or organisation.
	Owner string
	// Repo is the repository name, without the .git suffix.
	Repo string
	// Scheme is https for a web URL, or ssh for an SSH remote.
	Scheme string
}

// Path is the owner and repository joined, as most APIs expect it.
func (r Remote) Path() string { return r.Owner + "/" + r.Repo }

// WebURL is the repository's page on its host.
func (r Remote) WebURL() string { return "https://" + r.Host + "/" + r.Path() }

// ErrUnparsable reports a remote URL that could not be broken down.
var ErrUnparsable = errors.New("could not read the git remote URL")

// ParseRemote reads a git remote URL.
//
// Both forms git uses are handled: an https URL, and the scp-like SSH
// form git@host:owner/repo.git, which is not a URL at all and so cannot
// go through net/url.
func ParseRemote(raw string) (Remote, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return Remote{}, ErrUnparsable
	}

	if isSCPForm(raw) {
		return parseSCP(raw)
	}
	return parseURL(raw)
}

// isSCPForm reports the git@host:path notation, which has a colon but no
// scheme separator.
func isSCPForm(raw string) bool {
	if strings.Contains(raw, "://") {
		return false
	}
	colon := strings.Index(raw, ":")
	return colon > 0
}

func parseSCP(raw string) (Remote, error) {
	// Strip any user prefix, which is usually git@.
	if at := strings.LastIndex(raw, "@"); at >= 0 {
		raw = raw[at+1:]
	}

	host, path, found := strings.Cut(raw, ":")
	if !found || host == "" {
		return Remote{}, ErrUnparsable
	}

	owner, repo, err := splitPath(path)
	if err != nil {
		return Remote{}, err
	}
	return Remote{Host: host, Owner: owner, Repo: repo, Scheme: "ssh"}, nil
}

func parseURL(raw string) (Remote, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return Remote{}, ErrUnparsable
	}

	owner, repo, err := splitPath(u.Path)
	if err != nil {
		return Remote{}, err
	}

	scheme := u.Scheme
	if scheme == "" {
		scheme = "https"
	}
	return Remote{Host: u.Hostname(), Owner: owner, Repo: repo, Scheme: scheme}, nil
}

// splitPath pulls the owner and repository out of a remote path. A nested
// path, which self-hosted instances allow, keeps everything before the
// last element as the owner.
func splitPath(path string) (owner, repo string, err error) {
	path = strings.Trim(path, "/")
	path = strings.TrimSuffix(path, ".git")

	slash := strings.LastIndex(path, "/")
	if slash <= 0 || slash == len(path)-1 {
		return "", "", ErrUnparsable
	}
	return path[:slash], path[slash+1:], nil
}
