package mcmeta

import (
	"context"
	"net/http"
	"net/url"
)

// Fabric and Quilt both serve a loader list per Minecraft version in the
// same shape, Quilt's meta service being a fork of Fabric's.
const (
	fabricLoaderAPI = "https://meta.fabricmc.net/v2/versions/loader/"
	quiltLoaderAPI  = "https://meta.quiltmc.org/v3/versions/loader/"
)

// fabricEntry is one loader entry, which nests the version a level down.
type fabricEntry struct {
	Loader struct {
		Version string `json:"version"`
		Stable  bool   `json:"stable"`
	} `json:"loader"`
}

// fabricVersions fetches loader versions for one Minecraft version from a
// Fabric-style meta service.
func fabricVersions(ctx context.Context, c *http.Client, base, mcVersion string) ([]Version, error) {
	var raw []fabricEntry
	if err := getJSON(ctx, c, base+url.PathEscape(mcVersion), &raw); err != nil {
		return nil, err
	}

	out := make([]Version, 0, len(raw))
	for _, e := range raw {
		if e.Loader.Version == "" {
			continue
		}
		// Quilt reports stable false for its beta line but does not always
		// set the field, so the version string is the tie breaker.
		out = append(out, Version{
			ID:     e.Loader.Version,
			Stable: e.Loader.Stable && looksStable(e.Loader.Version),
		})
	}
	return out, nil
}
