// Package mcmeta looks up Minecraft versions and mod loader versions, so
// the new pack form can offer them as lists rather than free text.
//
// Each loader publishes its own version list in its own shape, so there
// is one lookup per loader behind a single entry point. Nothing here is
// required for a pack to be created: every list has a manual fallback,
// because an offline machine or a changed API should not stop someone
// typing a version they already know.
package mcmeta

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// userAgent identifies the app to the APIs it queries.
const userAgent = "PalisadeMC/Packwiz-Studio (github.com/PalisadeMC/Packwiz-Studio)"

// requestTimeout bounds one lookup. These lists are small.
const requestTimeout = 20 * time.Second

// Version is one selectable version.
type Version struct {
	// ID is the version string as the loader's own tooling writes it,
	// which is what packwiz init expects.
	ID string
	// Stable marks a release rather than a beta, snapshot or release
	// candidate. Unstable versions are still offered, just not by default.
	Stable bool
}

// IDs returns the version strings, in order.
func IDs(versions []Version) []string {
	out := make([]string, 0, len(versions))
	for _, v := range versions {
		out = append(out, v.ID)
	}
	return out
}

// Stable filters a list down to its stable entries.
func Stable(versions []Version) []Version {
	out := make([]Version, 0, len(versions))
	for _, v := range versions {
		if v.Stable {
			out = append(out, v)
		}
	}
	return out
}

// client returns the HTTP client to use, defaulting when none is given.
func client(c *http.Client) *http.Client {
	if c != nil {
		return c
	}
	return &http.Client{Timeout: requestTimeout}
}

// getJSON fetches a URL and decodes it into out.
func getJSON(ctx context.Context, c *http.Client, url string, out any) error {
	body, err := get(ctx, c, url)
	if err != nil {
		return err
	}
	defer body.Close()
	return json.NewDecoder(body).Decode(out)
}

// get performs the request, returning the body for the caller to close.
func get(ctx context.Context, c *http.Client, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client(c).Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("%s returned %s", host(url), resp.Status)
	}
	return resp.Body, nil
}

// host names the service in an error, which is more use than the full URL.
func host(url string) string {
	trimmed := strings.TrimPrefix(strings.TrimPrefix(url, "https://"), "http://")
	if i := strings.Index(trimmed, "/"); i > 0 {
		return trimmed[:i]
	}
	return trimmed
}

// unstableMarkers are the substrings that mean a version is not a release.
var unstableMarkers = []string{"-beta", "-alpha", "-rc", "-pre", "snapshot"}

// looksStable reports whether a version string reads as a release.
func looksStable(version string) bool {
	lower := strings.ToLower(version)
	for _, marker := range unstableMarkers {
		if strings.Contains(lower, marker) {
			return false
		}
	}
	return true
}
