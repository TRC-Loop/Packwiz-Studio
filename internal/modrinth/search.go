package modrinth

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
)

// Hit is one search result, as the browser's cards show it.
type Hit struct {
	ProjectID   string   `json:"project_id"`
	Slug        string   `json:"slug"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Categories  []string `json:"categories"`
	IconURL     string   `json:"icon_url"`
	Downloads   int      `json:"downloads"`
	Author      string   `json:"author"`
	ProjectType string   `json:"project_type"`
	Versions    []string `json:"versions"`
	Loaders     []string `json:"loaders"`
	// License is the project's licence identifier, which decides whether
	// a pack may redistribute the mod.
	License string `json:"license"`
	// ClientSide and ServerSide report support per side, each being
	// required, optional or unsupported.
	ClientSide string `json:"client_side"`
	ServerSide string `json:"server_side"`
}

// Sides describes which sides a mod supports, in words.
func (h Hit) Sides() string {
	client := supported(h.ClientSide)
	server := supported(h.ServerSide)

	switch {
	case client && server:
		return "client and server"
	case client:
		return "client only"
	case server:
		return "server only"
	default:
		return "side not stated"
	}
}

// supported reads one of Modrinth's side values.
func supported(value string) bool {
	switch strings.ToLower(value) {
	case "required", "optional":
		return true
	default:
		return false
	}
}

// Ref is the argument packwiz is given to add this project. The project
// id is used rather than the slug because it never changes, while a slug
// can be renamed by the project owner.
func (h Hit) Ref() string {
	if h.ProjectID != "" {
		return h.ProjectID
	}
	return h.Slug
}

// searchResponse is the shape of a search reply.
type searchResponse struct {
	Hits      []Hit `json:"hits"`
	Offset    int   `json:"offset"`
	Limit     int   `json:"limit"`
	TotalHits int   `json:"total_hits"`
}

// Results is a page of search results.
type Results struct {
	Hits  []Hit
	Total int
	// Offset is where this page started, so the browser can ask for the
	// next one.
	Offset int
}

// Query describes a search.
type Query struct {
	// Text is the free text search. An empty query lists popular
	// projects, which is what the browser shows before anything is typed.
	Text string
	// MCVersion and Loader narrow results to what the open pack can
	// actually use. Both are optional.
	MCVersion string
	Loader    string
	// Offset and Limit page the results.
	Offset int
	Limit  int
}

// DefaultLimit is how many results a page holds.
const DefaultLimit = 20

// Search finds projects matching q.
func (c *Client) Search(ctx context.Context, q Query) (Results, error) {
	limit := q.Limit
	if limit <= 0 {
		limit = DefaultLimit
	}

	values := url.Values{}
	if text := strings.TrimSpace(q.Text); text != "" {
		values.Set("query", text)
	}
	values.Set("limit", strconv.Itoa(limit))
	values.Set("offset", strconv.Itoa(q.Offset))
	values.Set("index", "relevance")

	if facets := facets(q); facets != "" {
		values.Set("facets", facets)
	}

	var resp searchResponse
	if err := c.get(ctx, "/search", values, &resp); err != nil {
		return Results{}, err
	}

	return Results{Hits: resp.Hits, Total: resp.TotalHits, Offset: resp.Offset}, nil
}

// facets builds Modrinth's filter expression.
//
// The format is a JSON array of arrays: entries in the same inner array
// are alternatives, and separate inner arrays must all match. Every
// filter here is a separate requirement, so each gets its own array.
func facets(q Query) string {
	var groups [][]string

	// Only mods are offered. A pack can hold resource packs and shaders
	// too, but adding those is a different flow and mixing them into one
	// list would make the results confusing.
	groups = append(groups, []string{"project_type:mod"})

	if v := strings.TrimSpace(q.MCVersion); v != "" {
		groups = append(groups, []string{"versions:" + v})
	}
	if l := strings.TrimSpace(q.Loader); l != "" {
		groups = append(groups, []string{"categories:" + strings.ToLower(l)})
	}

	encoded, err := json.Marshal(groups)
	if err != nil {
		return ""
	}
	return string(encoded)
}
