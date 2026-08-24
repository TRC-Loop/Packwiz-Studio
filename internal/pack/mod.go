package pack

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// Side is which half of a client and server pair a mod belongs to.
type Side string

// Side values, spelled as packwiz writes them.
const (
	SideBoth   Side = "both"
	SideClient Side = "client"
	SideServer Side = "server"
)

// Sides lists the choices the side control offers.
var Sides = []Side{SideBoth, SideClient, SideServer}

// Label is the side's display name.
func (s Side) Label() string {
	switch s {
	case SideClient:
		return "Client only"
	case SideServer:
		return "Server only"
	default:
		return "Client and server"
	}
}

// ParseSide resolves a side written in a metafile. An absent or unknown
// value means both, which is packwiz's own default.
func ParseSide(s string) Side {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case string(SideClient):
		return SideClient
	case string(SideServer):
		return SideServer
	default:
		return SideBoth
	}
}

// Mod is one entry in a pack, read from its .pw.toml metafile.
type Mod struct {
	// Name is the mod's display name.
	Name string
	// Path is the metafile's path relative to the pack folder. It is the
	// mod's identity, and what packwiz commands are given.
	Path string
	// Filename is the jar the metafile resolves to, which is the closest
	// thing to a human readable version.
	Filename string
	// SideFlag is which side the mod is installed on.
	SideFlag Side
	// Pinned reports a mod frozen at its current version.
	Pinned bool
	// Optional reports a mod the pack offers but does not install by
	// default.
	Optional bool
	// ModrinthID and VersionID identify the Modrinth project and the
	// exact version. VersionID is what a changelog diff compares.
	ModrinthID string
	VersionID  string
	// DownloadURL is where the jar comes from, shown in the detail pane.
	DownloadURL string
	// LoadErr records a metafile that could not be parsed. The mod is
	// still listed so the problem is visible rather than silent.
	LoadErr error
}

// Slug is the name packwiz commands use for this mod, which is its
// metafile name without the .pw.toml suffix.
func (m Mod) Slug() string {
	base := filepath.Base(m.Path)
	return strings.TrimSuffix(base, ".pw.toml")
}

// modFile mirrors a .pw.toml.
type modFile struct {
	Name     string       `toml:"name"`
	Filename string       `toml:"filename"`
	Side     string       `toml:"side"`
	Pin      bool         `toml:"pin"`
	Download modDownload  `toml:"download"`
	Update   modUpdate    `toml:"update"`
	Option   modOptionSet `toml:"option"`
}

type modDownload struct {
	URL string `toml:"url"`
}

type modUpdate struct {
	Modrinth modModrinth `toml:"modrinth"`
}

type modModrinth struct {
	ModID   string `toml:"mod-id"`
	Version string `toml:"version"`
}

type modOptionSet struct {
	Optional bool `toml:"optional"`
}

// loadMod reads one metafile, given its path relative to the pack.
func loadMod(dir, rel string) (Mod, error) {
	data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel)))
	if err != nil {
		return Mod{}, err
	}

	var mf modFile
	if err := toml.Unmarshal(data, &mf); err != nil {
		return Mod{}, &MalformedError{Path: rel, Err: err}
	}

	name := mf.Name
	if name == "" {
		name = strings.TrimSuffix(filepath.Base(rel), ".pw.toml")
	}

	return Mod{
		Name:        name,
		Path:        rel,
		Filename:    mf.Filename,
		SideFlag:    ParseSide(mf.Side),
		Pinned:      mf.Pin,
		Optional:    mf.Option.Optional,
		ModrinthID:  mf.Update.Modrinth.ModID,
		VersionID:   mf.Update.Modrinth.Version,
		DownloadURL: mf.Download.URL,
	}, nil
}
