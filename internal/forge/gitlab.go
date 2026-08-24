package forge

import (
	"context"
	"net/http"
	"net/url"
)

// gitlabClient publishes releases through the GitLab API.
//
// GitLab differs from the other two in how assets work: a release does
// not hold uploaded files directly. A file is uploaded to the project and
// then linked to the release, so UploadAsset does both steps.
type gitlabClient struct {
	host  Host
	token string
	http  *http.Client
}

func (c *gitlabClient) auth() http.Header {
	return http.Header{"PRIVATE-TOKEN": {c.token}}
}

// projectURL is the API path for this project. GitLab identifies a
// project by its URL encoded path.
func (c *gitlabClient) projectURL() string {
	return c.host.APIBase + "/projects/" + url.PathEscape(c.host.Remote.Path())
}

// gitlabRelease is the subset of GitLab's release object used here.
type gitlabRelease struct {
	TagName string `json:"tag_name"`
	Links   struct {
		Self string `json:"self"`
	} `json:"_links"`
}

// gitlabUpload is the reply to a project file upload.
type gitlabUpload struct {
	URL      string `json:"url"`
	FullPath string `json:"full_path"`
}

func (c *gitlabClient) CreateRelease(ctx context.Context, r Release) (Published, error) {
	body := map[string]any{
		"tag_name":    r.Tag,
		"name":        r.Title,
		"description": r.Notes,
	}

	var out gitlabRelease
	err := request(ctx, c.http, http.MethodPost, c.projectURL()+"/releases",
		c.auth(), body, &out)
	if err != nil {
		return Published{}, err
	}

	// GitLab keys release operations by tag rather than by a numeric id,
	// so the tag is what gets carried forward as the identifier.
	return Published{ID: r.Tag, URL: out.Links.Self}, nil
}

func (c *gitlabClient) UploadAsset(ctx context.Context, release Published, a Asset) error {
	var uploaded gitlabUpload
	err := upload(ctx, c.http, c.projectURL()+"/uploads",
		c.auth(), "file", a.Path, &uploaded)
	if err != nil {
		return err
	}

	link := uploaded.FullPath
	if link == "" {
		link = uploaded.URL
	}
	if link != "" && link[0] == '/' {
		link = "https://" + c.host.Remote.Host + link
	}

	body := map[string]any{"name": a.Name, "url": link}

	return request(ctx, c.http, http.MethodPost,
		c.projectURL()+"/releases/"+url.PathEscape(release.ID)+"/assets/links",
		c.auth(), body, nil)
}
