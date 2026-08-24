// Package modrinth talks to the Modrinth API.
//
// The app searches here and adds through packwiz: packwiz has no search
// command, but it owns everything about writing a mod into a pack. So
// this client is read only, and its results feed packwiz a project id.
package modrinth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// baseURL is Modrinth's API root.
const baseURL = "https://api.modrinth.com/v2"

// userAgent identifies the app to Modrinth, which their API guidelines
// ask for so a misbehaving client can be contacted rather than blocked.
const userAgent = "PalisadeMC/Packwiz-Studio (github.com/PalisadeMC/Packwiz-Studio)"

// Client queries the Modrinth API.
type Client struct {
	http *http.Client
}

// New returns a client with a sensible timeout.
func New() *Client {
	return &Client{http: &http.Client{Timeout: 20 * time.Second}}
}

// NewWithHTTP returns a client using a caller supplied HTTP client.
func NewWithHTTP(h *http.Client) *Client {
	if h == nil {
		return New()
	}
	return &Client{http: h}
}

// get fetches path with query and decodes the response into out.
func (c *Client) get(ctx context.Context, path string, query url.Values, out any) error {
	endpoint := baseURL + path
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := statusError(resp); err != nil {
		return err
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// statusError turns a non-success response into an error worth showing.
func statusError(resp *http.Response) error {
	switch {
	case resp.StatusCode == http.StatusOK:
		return nil
	case resp.StatusCode == http.StatusNotFound:
		return ErrNotFound
	case resp.StatusCode == http.StatusTooManyRequests:
		return ErrRateLimited
	default:
		return fmt.Errorf("modrinth returned %s", resp.Status)
	}
}

// Errors the UI distinguishes.
var (
	// ErrNotFound reports a project that does not exist.
	ErrNotFound = errNotFound{}
	// ErrRateLimited reports too many requests, which the browser
	// surfaces as a message rather than an empty result.
	ErrRateLimited = errRateLimited{}
)

type errNotFound struct{}

func (errNotFound) Error() string { return "that project does not exist on Modrinth" }

type errRateLimited struct{}

func (errRateLimited) Error() string {
	return "Modrinth is rate limiting this client, wait a moment and try again"
}
