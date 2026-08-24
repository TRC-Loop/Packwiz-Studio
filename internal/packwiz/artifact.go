package packwiz

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// artifactsAPI lists the builds packwiz's CI produced.
//
// packwiz publishes no releases or tags, so there is no /releases/latest
// to ask: its binaries are GitHub Actions artifacts. Listing them needs no
// authentication, but downloading one through the GitHub API does, which
// is what proxyBase exists for.
const artifactsAPI = "https://api.github.com/repos/packwiz/packwiz/actions/artifacts?per_page=100"

// proxyBase serves Actions artifacts without authentication.
//
// GitHub returns 401 for an unauthenticated artifact download, and asking
// a user for a personal access token just to install a tool is a poor
// trade. nightly.link is the established proxy for this, and the app
// checks the size it returns against the size GitHub reported.
const proxyBase = "https://nightly.link/packwiz/packwiz/actions/artifacts/"

// Artifact is one CI build.
type Artifact struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	SizeBytes int64     `json:"size_in_bytes"`
	Expired   bool      `json:"expired"`
	CreatedAt time.Time `json:"created_at"`
}

// DownloadURL is where the artifact can be fetched without a token.
func (a Artifact) DownloadURL() string {
	return proxyBase + strconv.FormatInt(a.ID, 10) + ".zip"
}

// ErrNoArtifact reports that CI has no build for this OS and architecture.
// packwiz builds amd64 only, so this is the normal outcome on arm64.
var ErrNoArtifact = errors.New("packwiz publishes no build for this OS and architecture")

// artifactList is the shape of the artifacts reply.
type artifactList struct {
	Artifacts []Artifact `json:"artifacts"`
}

// LatestArtifact finds the newest usable build for the running platform.
func LatestArtifact(ctx context.Context, client *http.Client) (Artifact, error) {
	all, err := fetchArtifacts(ctx, client)
	if err != nil {
		return Artifact{}, err
	}
	return pickArtifact(all, runtime.GOOS, runtime.GOARCH)
}

// fetchArtifacts lists what packwiz's CI has produced.
func fetchArtifacts(ctx context.Context, client *http.Client) ([]Artifact, error) {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, artifactsAPI, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github returned %s for the packwiz build list", resp.Status)
	}

	var list artifactList
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, err
	}
	return list.Artifacts, nil
}

// pickArtifact chooses the newest matching, unexpired build.
//
// Artifact names are matched on keywords rather than a fixed pattern.
// They are written for people, not machines, currently along the lines of
// "macOS 64-bit x86", so a rename should degrade to "no build for this
// platform" rather than breaking.
func pickArtifact(all []Artifact, goos, goarch string) (Artifact, error) {
	osWords := artifactOS[goos]
	archWords := artifactArch[goarch]
	if len(osWords) == 0 || len(archWords) == 0 {
		return Artifact{}, ErrNoArtifact
	}

	var best Artifact
	for _, a := range all {
		if a.Expired || a.SizeBytes <= 0 {
			continue
		}

		name := strings.ToLower(a.Name)
		if !containsAny(name, osWords) || !containsAny(name, archWords) {
			continue
		}
		// "64-bit" is treated as amd64, so an arm build must not be
		// picked up by it if one is ever added.
		if goarch != "arm64" && containsAny(name, artifactArch["arm64"]) {
			continue
		}
		if best.ID == 0 || a.CreatedAt.After(best.CreatedAt) {
			best = a
		}
	}

	if best.ID == 0 {
		return Artifact{}, ErrNoArtifact
	}
	return best, nil
}

var artifactOS = map[string][]string{
	"darwin":  {"macos", "darwin", "mac", "osx"},
	"linux":   {"linux"},
	"windows": {"windows", "win"},
}

// artifactArch includes the loose spellings CI uses. "64-bit" alone maps
// to amd64 because that is the only 64-bit build packwiz produces; if an
// arm64 artifact ever appears it will say so explicitly and match first.
var artifactArch = map[string][]string{
	"amd64": {"amd64", "x86_64", "x64", "x86", "64-bit"},
	"arm64": {"arm64", "aarch64"},
	"386":   {"386", "i386", "32-bit"},
}

func containsAny(s string, words []string) bool {
	for _, w := range words {
		if strings.Contains(s, w) {
			return true
		}
	}
	return false
}
