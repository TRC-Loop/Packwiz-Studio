package mcmeta

import (
	"context"
	"net/http"
	"strings"
)

// gameVersionAPI lists every Minecraft version Modrinth knows about,
// newest first. Modrinth is used rather than Mojang's own manifest
// because it is the same source the mod search is filtered against, so
// the versions offered here are the ones mods will be found for.
const gameVersionAPI = "https://api.modrinth.com/v2/tag/game_version"

// gameVersion is one entry of the reply.
type gameVersion struct {
	Version string `json:"version"`
	Type    string `json:"version_type"`
}

// GameVersions returns the Minecraft versions, newest first.
//
// Snapshots are included but marked unstable, so the form can offer
// releases by default and reveal the rest on request.
func GameVersions(ctx context.Context, c *http.Client) ([]Version, error) {
	var raw []gameVersion
	if err := getJSON(ctx, c, gameVersionAPI, &raw); err != nil {
		return nil, err
	}

	out := make([]Version, 0, len(raw))
	for _, v := range raw {
		out = append(out, Version{
			ID:     v.Version,
			Stable: strings.EqualFold(v.Type, "release"),
		})
	}
	return out, nil
}
