package packwiz

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Progress reports download progress. Total is -1 when the server does
// not declare a content length. It is called from the download goroutine,
// so a UI callback must hop to the main thread with fyne.Do.
type Progress func(done, total int64)

// Installer downloads a packwiz release and puts the binary somewhere the
// app can find it again.
type Installer struct {
	// Client is the HTTP client to use. Nil means a default with a
	// generous timeout for the download itself.
	Client *http.Client
	// Dest overrides the install location. Empty means the managed path
	// beside the app config.
	Dest string
}

// Install fetches the latest release, extracts the binary and verifies it
// runs. It returns the resolved location, so the caller can store the
// path in config straight away.
func (in *Installer) Install(ctx context.Context, progress Progress) (Location, error) {
	client := in.Client
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Minute}
	}

	rel, err := LatestRelease(ctx, client)
	if err != nil {
		return Location{}, err
	}
	asset, err := rel.AssetForHost()
	if err != nil {
		return Location{}, err
	}

	dest := in.Dest
	if dest == "" {
		if dest, err = managedPath(); err != nil {
			return Location{}, err
		}
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return Location{}, err
	}

	archive, cleanup, err := download(ctx, client, asset, progress)
	if err != nil {
		return Location{}, err
	}
	defer cleanup()

	if err := extractBinary(archive, asset.Name, dest); err != nil {
		return Location{}, err
	}
	if err := os.Chmod(dest, 0o755); err != nil {
		return Location{}, err
	}

	loc, err := Verify(ctx, dest)
	if err != nil {
		// The download succeeded but produced something unusable. Remove
		// it rather than leaving a broken binary that Locate would find
		// and fail on at every startup.
		os.Remove(dest)
		return Location{}, err
	}
	if loc.Version == "" {
		loc.Version = rel.Tag
	}
	return loc, nil
}

// download streams an asset to a temporary file, reporting progress. The
// returned cleanup removes the temporary file.
func download(ctx context.Context, client *http.Client, asset Asset, progress Progress) (path string, cleanup func(), err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.URL, nil)
	if err != nil {
		return "", func() {}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", func() {}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", func() {}, fmt.Errorf("downloading %s returned %s", asset.Name, resp.Status)
	}

	tmp, err := os.CreateTemp("", "packwiz-download-*")
	if err != nil {
		return "", func() {}, err
	}
	cleanup = func() {
		tmp.Close()
		os.Remove(tmp.Name())
	}

	total := resp.ContentLength
	if total <= 0 {
		total = asset.Size
	}
	if total <= 0 {
		total = -1
	}

	src := io.Reader(resp.Body)
	if progress != nil {
		src = &progressReader{r: resp.Body, total: total, report: progress}
	}

	if _, err := io.Copy(tmp, src); err != nil {
		cleanup()
		return "", func() {}, err
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return "", func() {}, err
	}
	return tmp.Name(), cleanup, nil
}

// progressReader reports how much of a stream has been consumed. Reports
// are throttled so a fast download does not flood the UI with updates.
type progressReader struct {
	r      io.Reader
	total  int64
	done   int64
	last   time.Time
	report Progress
}

// reportInterval is the minimum gap between progress callbacks.
const reportInterval = 50 * time.Millisecond

func (p *progressReader) Read(b []byte) (int, error) {
	n, err := p.r.Read(b)
	p.done += int64(n)

	now := time.Now()
	atEnd := err != nil
	if atEnd || now.Sub(p.last) >= reportInterval {
		p.last = now
		p.report(p.done, p.total)
	}
	return n, err
}
