package mcmeta

import (
	"context"
	"net/http"
	"strings"
	"sync"
)

// Loader names, matching what packwiz's --modloader expects.
const (
	LoaderFabric     = "fabric"
	LoaderForge      = "forge"
	LoaderNeoForge   = "neoforge"
	LoaderQuilt      = "quilt"
	LoaderLiteLoader = "liteloader"
)

// ErrNoVersionSource reports a loader with no version list to query.
// LiteLoader is the case: it is long unmaintained and publishes nothing
// to look up, so its version has to be typed in.
type ErrNoVersionSource struct{ Loader string }

func (e ErrNoVersionSource) Error() string {
	return "no version list is published for " + e.Loader + ", type the version you want"
}

// LoaderVersions returns the versions of one loader for one Minecraft
// version, newest first.
func LoaderVersions(ctx context.Context, c *http.Client, loader, mcVersion string) ([]Version, error) {
	if strings.TrimSpace(mcVersion) == "" {
		return nil, nil
	}

	switch strings.ToLower(loader) {
	case LoaderFabric:
		return fabricVersions(ctx, c, fabricLoaderAPI, mcVersion)
	case LoaderQuilt:
		return fabricVersions(ctx, c, quiltLoaderAPI, mcVersion)
	case LoaderForge:
		return forgeVersions(ctx, c, mcVersion)
	case LoaderNeoForge:
		return neoforgeVersionsFor(ctx, c, mcVersion)
	default:
		return nil, ErrNoVersionSource{Loader: loader}
	}
}

// Cache holds version lists for the life of the app.
//
// The Minecraft list is a few hundred entries and never changes while the
// app runs, and reopening the new pack form should not refetch it. Loader
// lists are cached per loader and Minecraft version pair.
type Cache struct {
	client *http.Client

	mu     sync.Mutex
	game   []Version
	loader map[string][]Version
}

// NewCache returns a cache using the given HTTP client, or a default.
func NewCache(c *http.Client) *Cache {
	return &Cache{client: c, loader: map[string][]Version{}}
}

// GameVersions returns the Minecraft versions, fetching them once.
func (c *Cache) GameVersions(ctx context.Context) ([]Version, error) {
	c.mu.Lock()
	cached := c.game
	c.mu.Unlock()

	if cached != nil {
		return cached, nil
	}

	fetched, err := GameVersions(ctx, c.client)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.game = fetched
	c.mu.Unlock()

	return fetched, nil
}

// LoaderVersions returns a loader's versions for a Minecraft version,
// fetching each pair once.
func (c *Cache) LoaderVersions(ctx context.Context, loader, mcVersion string) ([]Version, error) {
	key := loader + "@" + mcVersion

	c.mu.Lock()
	cached, ok := c.loader[key]
	c.mu.Unlock()

	if ok {
		return cached, nil
	}

	fetched, err := LoaderVersions(ctx, c.client, loader, mcVersion)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.loader[key] = fetched
	c.mu.Unlock()

	return fetched, nil
}
