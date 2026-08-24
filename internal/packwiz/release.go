package packwiz

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"
)

// releaseAPI is the latest-release endpoint for packwiz's own repository.
const releaseAPI = "https://api.github.com/repos/packwiz/packwiz/releases/latest"

// Release is the subset of a GitHub release the installer needs.
type Release struct {
	Tag    string  `json:"tag_name"`
	Assets []Asset `json:"assets"`
}

// Asset is one downloadable file attached to a release.
type Asset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
	Size int64  `json:"size"`
}

// LatestRelease fetches packwiz's most recent release.
func LatestRelease(ctx context.Context, client *http.Client) (Release, error) {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, releaseAPI, nil)
	if err != nil {
		return Release{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return Release{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Release{}, fmt.Errorf("github returned %s for the packwiz release list", resp.Status)
	}

	var rel Release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return Release{}, err
	}
	if len(rel.Assets) == 0 {
		return Release{}, errors.New("the latest packwiz release has no downloadable assets")
	}
	return rel, nil
}

// ErrNoAsset reports that a release carries no build for this platform.
var ErrNoAsset = errors.New("no packwiz build for this OS and architecture")

// AssetFor picks the asset matching the running OS and architecture.
//
// Asset names are matched on keywords rather than a fixed template:
// packwiz's release naming is not a stable contract, so a rename should
// degrade to "ask the user to browse for a binary", not crash.
func (r Release) AssetFor(goos, goarch string) (Asset, error) {
	osWords := osAliases[goos]
	archWords := archAliases[goarch]
	if len(osWords) == 0 || len(archWords) == 0 {
		return Asset{}, ErrNoAsset
	}

	var best Asset
	for _, a := range r.Assets {
		name := strings.ToLower(a.Name)
		if !containsAny(name, osWords) || !containsAny(name, archWords) {
			continue
		}
		if !isArchive(name) {
			continue
		}
		// Prefer the first match, but let a .zip win over a .tar.gz on
		// Windows and lose to it elsewhere. Matching what each platform
		// handles natively is not required, both are supported; this
		// just keeps the choice predictable.
		if best.Name == "" || preferred(name, goos) {
			best = a
		}
	}
	if best.Name == "" {
		return Asset{}, ErrNoAsset
	}
	return best, nil
}

// AssetForHost is AssetFor for the running platform.
func (r Release) AssetForHost() (Asset, error) {
	return r.AssetFor(runtime.GOOS, runtime.GOARCH)
}

var osAliases = map[string][]string{
	"darwin":  {"darwin", "macos", "mac", "osx"},
	"linux":   {"linux"},
	"windows": {"windows", "win"},
}

var archAliases = map[string][]string{
	"amd64": {"amd64", "x86_64", "x64"},
	"arm64": {"arm64", "aarch64"},
	"386":   {"386", "i386", "x86"},
}

func containsAny(s string, words []string) bool {
	for _, w := range words {
		if strings.Contains(s, w) {
			return true
		}
	}
	return false
}

func isArchive(name string) bool {
	return strings.HasSuffix(name, ".zip") ||
		strings.HasSuffix(name, ".tar.gz") ||
		strings.HasSuffix(name, ".tgz")
}

func preferred(name, goos string) bool {
	if goos == "windows" {
		return strings.HasSuffix(name, ".zip")
	}
	return strings.HasSuffix(name, ".tar.gz") || strings.HasSuffix(name, ".tgz")
}
