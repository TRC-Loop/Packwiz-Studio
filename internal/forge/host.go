package forge

import "strings"

// Kind is which API a host speaks.
type Kind int

// Host kinds.
const (
	// KindGitHub is github.com and GitHub Enterprise.
	KindGitHub Kind = iota
	// KindGitLab is gitlab.com and self-hosted GitLab.
	KindGitLab
	// KindGitea covers Gitea and Forgejo, whose release APIs match.
	// Codeberg is a Forgejo instance and lands here.
	KindGitea
)

// Name is the host kind's display name.
func (k Kind) Name() string {
	switch k {
	case KindGitHub:
		return "GitHub"
	case KindGitLab:
		return "GitLab"
	default:
		return "Gitea or Forgejo"
	}
}

// Host is a resolved code host: which API to speak and where.
type Host struct {
	Kind Kind
	// APIBase is the API root, with no trailing slash.
	APIBase string
	// Remote is the repository the host was resolved from.
	Remote Remote
}

// TokenKey is the keyring entry a token for this host is stored under.
// The hostname is used rather than the API base, so one token serves a
// host however its API path happens to be shaped.
func (h Host) TokenKey() string { return h.Remote.Host }

// DetectHost works out which API a remote speaks.
//
// Known public domains are matched by name. Anything else is assumed to
// be Gitea or Forgejo, which is the common case for a self-hosted
// instance and the assumption the app is specified to make. A configured
// base URL overrides the guessed one, for an instance whose API does not
// sit at the host root.
func DetectHost(remote Remote, configuredBase string) Host {
	kind := kindFor(remote.Host)

	base := strings.TrimRight(strings.TrimSpace(configuredBase), "/")
	if base == "" {
		base = defaultAPIBase(kind, remote.Host)
	} else {
		base = appendAPIPath(kind, base)
	}

	return Host{Kind: kind, APIBase: base, Remote: remote}
}

// kindFor maps a hostname onto an API family.
func kindFor(host string) Kind {
	h := strings.ToLower(host)

	switch {
	case h == "github.com" || strings.HasSuffix(h, ".github.com"):
		return KindGitHub
	case h == "gitlab.com" || strings.HasSuffix(h, ".gitlab.com"):
		return KindGitLab
	default:
		// codeberg.org runs Forgejo, and so do most self-hosted
		// instances, which is why this is the fallback rather than an
		// error.
		return KindGitea
	}
}

// defaultAPIBase is where each kind keeps its API.
func defaultAPIBase(kind Kind, host string) string {
	switch kind {
	case KindGitHub:
		if strings.EqualFold(host, "github.com") {
			return "https://api.github.com"
		}
		// GitHub Enterprise serves its API under a path rather than a
		// separate hostname.
		return "https://" + host + "/api/v3"
	case KindGitLab:
		return "https://" + host + "/api/v4"
	default:
		return "https://" + host + "/api/v1"
	}
}

// appendAPIPath adds the API path to a user supplied base when they gave
// just the instance root, which is the likelier thing to paste.
func appendAPIPath(kind Kind, base string) string {
	suffixes := map[Kind]string{
		KindGitHub: "/api/v3",
		KindGitLab: "/api/v4",
		KindGitea:  "/api/v1",
	}
	suffix := suffixes[kind]

	if strings.Contains(base, "/api/") {
		return base
	}
	return base + suffix
}
