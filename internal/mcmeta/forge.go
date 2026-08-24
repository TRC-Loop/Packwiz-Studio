package mcmeta

import (
	"context"
	"encoding/xml"
	"net/http"
	"strings"
)

// Forge publishes a maven metadata document; NeoForge has a small JSON
// API over the same maven repository.
const (
	forgeMetadataXML = "https://maven.minecraftforge.net/net/minecraftforge/forge/maven-metadata.xml"
	neoforgeVersions = "https://maven.neoforged.net/api/maven/versions/releases/net/neoforged/neoforge"
)

// forgeMetadata is the part of maven metadata that lists versions.
type forgeMetadata struct {
	Versioning struct {
		Versions struct {
			Version []string `xml:"version"`
		} `xml:"versions"`
	} `xml:"versioning"`
}

// forgeVersions returns Forge builds for one Minecraft version.
//
// Forge names a build "<minecraft>-<forge>", so the Minecraft version is a
// prefix and the part after it is what packwiz wants.
func forgeVersions(ctx context.Context, c *http.Client, mcVersion string) ([]Version, error) {
	body, err := get(ctx, c, forgeMetadataXML)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var doc forgeMetadata
	if err := xml.NewDecoder(body).Decode(&doc); err != nil {
		return nil, err
	}

	prefix := mcVersion + "-"

	var out []Version
	for _, full := range doc.Versioning.Versions.Version {
		if !strings.HasPrefix(full, prefix) {
			continue
		}
		build := strings.TrimPrefix(full, prefix)
		if build == "" {
			continue
		}
		out = append(out, Version{ID: build, Stable: looksStable(build)})
	}

	reverse(out)
	return out, nil
}

// neoforgeReply is the versions API's shape.
type neoforgeReply struct {
	Versions []string `json:"versions"`
}

// neoforgeVersionsFor returns NeoForge builds for one Minecraft version.
//
// NeoForge drops the leading "1." and encodes the Minecraft version in its
// own: Minecraft 1.21.1 gives builds numbered 21.1.x. A Minecraft version
// with no patch, such as 1.21, maps to 21.0.x.
func neoforgeVersionsFor(ctx context.Context, c *http.Client, mcVersion string) ([]Version, error) {
	prefix, ok := neoforgePrefix(mcVersion)
	if !ok {
		return nil, nil
	}

	var reply neoforgeReply
	if err := getJSON(ctx, c, neoforgeVersions, &reply); err != nil {
		return nil, err
	}

	var out []Version
	for _, v := range reply.Versions {
		if strings.HasPrefix(v, prefix) {
			out = append(out, Version{ID: v, Stable: looksStable(v)})
		}
	}

	reverse(out)
	return out, nil
}

// neoforgePrefix turns a Minecraft version into the NeoForge build prefix.
func neoforgePrefix(mcVersion string) (string, bool) {
	parts := strings.Split(strings.TrimSpace(mcVersion), ".")
	if len(parts) < 2 || parts[0] != "1" {
		return "", false
	}

	minor := parts[1]
	patch := "0"
	if len(parts) >= 3 {
		patch = parts[2]
	}
	return minor + "." + patch + ".", true
}

// reverse flips a slice in place, so a maven list that runs oldest first
// is presented newest first like every other list here.
func reverse(versions []Version) {
	for i, j := 0, len(versions)-1; i < j; i, j = i+1, j-1 {
		versions[i], versions[j] = versions[j], versions[i]
	}
}
