package forge

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// giteaClient publishes releases through the Gitea API, which Forgejo and
// Codeberg also serve.
type giteaClient struct {
	host  Host
	token string
	http  *http.Client
}

func (c *giteaClient) auth() http.Header {
	return http.Header{"Authorization": {"token " + c.token}}
}

func (c *giteaClient) repoURL() string {
	return c.host.APIBase + "/repos/" + c.host.Remote.Path()
}

// giteaRelease is the subset of Gitea's release object used here.
type giteaRelease struct {
	ID      int64  `json:"id"`
	HTMLURL string `json:"html_url"`
}

func (c *giteaClient) CreateRelease(ctx context.Context, r Release) (Published, error) {
	body := map[string]any{
		"tag_name":   r.Tag,
		"name":       r.Title,
		"body":       r.Notes,
		"draft":      r.Draft,
		"prerelease": r.Prerelease,
	}

	var out giteaRelease
	err := request(ctx, c.http, http.MethodPost, c.repoURL()+"/releases",
		c.auth(), body, &out)
	if err != nil {
		return Published{}, err
	}

	return Published{ID: strconv.FormatInt(out.ID, 10), URL: out.HTMLURL}, nil
}

func (c *giteaClient) UploadAsset(ctx context.Context, release Published, a Asset) error {
	endpoint := c.repoURL() + "/releases/" + release.ID +
		"/assets?name=" + url.QueryEscape(a.Name)

	return upload(ctx, c.http, endpoint, c.auth(), "attachment", a.Path, nil)
}
