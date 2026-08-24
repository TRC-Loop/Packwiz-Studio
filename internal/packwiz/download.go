package packwiz

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Progress reports download progress. Total is -1 when the size is not
// known. It is called from the download goroutine, so a UI callback must
// hop to the main thread with fyne.Do.
type Progress func(done, total int64)

// downloadArtifact fetches an artifact zip to a temporary file. The
// returned cleanup removes it.
func downloadArtifact(ctx context.Context, client *http.Client, a Artifact,
	progress Progress) (path string, cleanup func(), err error) {

	noop := func() {}

	if client == nil {
		client = &http.Client{Timeout: downloadTimeout}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.DownloadURL(), nil)
	if err != nil {
		return "", noop, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", noop, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", noop, fmt.Errorf("downloading %s returned %s", a.Name, resp.Status)
	}

	tmp, err := os.CreateTemp("", "packwiz-artifact-*.zip")
	if err != nil {
		return "", noop, err
	}
	cleanup = func() {
		tmp.Close()
		os.Remove(tmp.Name())
	}

	total := resp.ContentLength
	if total <= 0 {
		total = a.SizeBytes
	}
	if total <= 0 {
		total = -1
	}

	src := io.Reader(resp.Body)
	if progress != nil {
		src = &progressReader{r: resp.Body, total: total, report: progress}
	}

	written, err := io.Copy(tmp, src)
	if err != nil {
		cleanup()
		return "", noop, err
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return "", noop, err
	}

	// The download came through a proxy rather than from GitHub directly,
	// so the size GitHub reported is checked against what arrived.
	if a.SizeBytes > 0 && written != a.SizeBytes {
		cleanup()
		return "", noop, fmt.Errorf(
			"the downloaded build is %d bytes but GitHub reports %d, so it was not fetched intact",
			written, a.SizeBytes)
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

// downloadTimeout bounds a download when the caller supplies no client.
const downloadTimeout = 10 * time.Minute

func (p *progressReader) Read(b []byte) (int, error) {
	n, err := p.r.Read(b)
	p.done += int64(n)

	now := time.Now()
	if err != nil || now.Sub(p.last) >= reportInterval {
		p.last = now
		p.report(p.done, p.total)
	}
	return n, err
}
