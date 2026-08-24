package changelog

import (
	"context"
	"path"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/PalisadeMC/Packwiz-Studio/internal/pack"
)

// Reader reads files at a git revision. The git package satisfies this,
// and keeping it an interface here means the changelog does not depend on
// how the revision is fetched.
type Reader interface {
	Show(ctx context.Context, rev, path string) (string, error)
}

// Current builds a snapshot from the pack as it is on disk.
func Current(p pack.Pack) (Snapshot, error) {
	mods, err := p.Mods()
	if err != nil {
		return nil, err
	}

	snap := make(Snapshot, len(mods))
	for _, m := range mods {
		snap[m.Path] = Mod{Name: m.Name, Version: m.VersionID}
	}
	return snap, nil
}

// AtRevision builds a snapshot from what the repository held at rev.
//
// The index at that revision names the metafiles, and each is read back
// out of git rather than off disk. A metafile that will not parse is
// skipped: a changelog is worth producing even when one historical entry
// is unreadable.
func AtRevision(ctx context.Context, r Reader, rev, indexFile string) (Snapshot, error) {
	if indexFile == "" {
		indexFile = "index.toml"
	}

	content, err := r.Show(ctx, rev, indexFile)
	if err != nil {
		return nil, err
	}

	var idx struct {
		Files []struct {
			File     string `toml:"file"`
			Metafile bool   `toml:"metafile"`
		} `toml:"files"`
	}
	if err := toml.Unmarshal([]byte(content), &idx); err != nil {
		return nil, err
	}

	// Metafile paths in the index are relative to the index itself, which
	// is usually at the pack root but need not be.
	base := path.Dir(indexFile)

	snap := Snapshot{}
	for _, f := range idx.Files {
		if !f.Metafile || f.File == "" {
			continue
		}

		full := f.File
		if base != "." && base != "" {
			full = path.Join(base, f.File)
		}

		body, err := r.Show(ctx, rev, full)
		if err != nil {
			continue
		}
		if mod, ok := parseMod(body, f.File); ok {
			snap[f.File] = mod
		}
	}
	return snap, nil
}

// parseMod reads the name and version out of a historical metafile.
func parseMod(content, relPath string) (Mod, bool) {
	var mf struct {
		Name   string `toml:"name"`
		Update struct {
			Modrinth struct {
				Version string `toml:"version"`
			} `toml:"modrinth"`
		} `toml:"update"`
	}
	if err := toml.Unmarshal([]byte(content), &mf); err != nil {
		return Mod{}, false
	}

	name := mf.Name
	if name == "" {
		name = strings.TrimSuffix(path.Base(relPath), ".pw.toml")
	}
	return Mod{Name: name, Version: mf.Update.Modrinth.Version}, true
}
