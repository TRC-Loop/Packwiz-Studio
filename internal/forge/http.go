package forge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// defaultHTTP is the client the release APIs use. Uploads can be tens of
// megabytes, so the timeout is generous.
func defaultHTTP() *http.Client {
	return &http.Client{Timeout: 10 * time.Minute}
}

// request sends a JSON request and decodes a JSON response.
//
// The auth header differs per host, so it is passed in rather than being
// decided here.
func request(ctx context.Context, client *http.Client, method, url string,
	auth http.Header, body any, out any) error {

	var payload io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return err
		}
		payload = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, payload)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	maps.Copy(req.Header, auth)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := apiError(resp); err != nil {
		return err
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// upload posts a file as multipart form data, which is how Gitea and
// GitLab take assets.
func upload(ctx context.Context, client *http.Client, url string,
	auth http.Header, field, path string, out any) error {

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	var buf bytes.Buffer
	form := multipart.NewWriter(&buf)

	part, err := form.CreateFormFile(field, filepath.Base(path))
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, file); err != nil {
		return err
	}
	if err := form.Close(); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", form.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	maps.Copy(req.Header, auth)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := apiError(resp); err != nil {
		return err
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// uploadRaw posts a file as a plain body, which is how GitHub takes
// release assets.
func uploadRaw(ctx context.Context, client *http.Client, url string,
	auth http.Header, path string) error {

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, file)
	if err != nil {
		return err
	}
	req.ContentLength = info.Size()
	req.Header.Set("Content-Type", contentType(path))
	req.Header.Set("Accept", "application/json")
	maps.Copy(req.Header, auth)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return apiError(resp)
}

// apiError turns a failure response into an error carrying the host's own
// message, which explains a rejected release better than a status code.
func apiError(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	message := strings.TrimSpace(string(body))

	if message == "" {
		return fmt.Errorf("the host returned %s", resp.Status)
	}
	return fmt.Errorf("the host returned %s: %s", resp.Status, message)
}

// contentType guesses an upload's type from its extension. Hosts store
// this as metadata and do not act on it, so a coarse guess is enough.
func contentType(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".mrpack", ".zip":
		return "application/zip"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}
