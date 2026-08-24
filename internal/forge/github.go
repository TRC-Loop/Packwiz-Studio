package forge

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// githubClient publishes releases through the GitHub API.
type githubClient struct {
	host  Host
	token string
	http  *http.Client
}

func (c *githubClient) auth() http.Header {
	return http.Header{
		"Authorization":        {"Bearer " + c.token},
		"X-GitHub-Api-Version": {"2022-11-28"},
	}
}

// repoURL is the API path for this repository.
func (c *githubClient) repoURL() string {
	return c.host.APIBase + "/repos/" + c.host.Remote.Path()
}

// githubRelease is the subset of GitHub's release object used here.
type githubRelease struct {
	ID        int64  `json:"id"`
	HTMLURL   string `json:"html_url"`
	UploadURL string `json:"upload_url"`
}

func (c *githubClient) CreateRelease(ctx context.Context, r Release) (Published, error) {
	body := map[string]any{
		"tag_name":   r.Tag,
		"name":       r.Title,
		"body":       r.Notes,
		"draft":      r.Draft,
		"prerelease": r.Prerelease,
	}

	var out githubRelease
	err := request(ctx, c.http, http.MethodPost, c.repoURL()+"/releases",
		c.auth(), body, &out)
	if err != nil {
		return Published{}, err
	}

	return Published{ID: strconv.FormatInt(out.ID, 10), URL: out.HTMLURL}, nil
}

func (c *githubClient) UploadAsset(ctx context.Context, release Published, a Asset) error {
	// GitHub takes assets on a separate uploads host, and the release
	// object advertises it as a URI template ending in {?name,label}.
	// Building the URL from the release id instead keeps this independent
	// of that template.
	endpoint := c.uploadHost() + "/repos/" + c.host.Remote.Path() +
		"/releases/" + release.ID + "/assets?name=" + url.QueryEscape(a.Name)

	return uploadRaw(ctx, c.http, endpoint, c.auth(), a.Path)
}

// uploadHost is where GitHub takes asset uploads, which is a different
// host from its API on github.com but the same one on Enterprise.
func (c *githubClient) uploadHost() string {
	if c.host.APIBase == "https://api.github.com" {
		return "https://uploads.github.com"
	}
	return strings.TrimSuffix(c.host.APIBase, "/")
}
