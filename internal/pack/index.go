package pack

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
)

// indexFile mirrors the parts of index.toml the app reads.
type indexFile struct {
	Files []indexEntry `toml:"files"`
}

type indexEntry struct {
	File     string `toml:"file"`
	Metafile bool   `toml:"metafile"`
}

// Mods reads the pack's mod list.
//
// The index is the source of truth for what is in a pack, and the entries
// it marks as metafiles are the per-mod TOML files. Anything else in the
// index is a plain included file and is not a mod.
func (p Pack) Mods() ([]Mod, error) {
	entries, err := p.metafiles()
	if err != nil {
		return nil, err
	}

	mods := make([]Mod, 0, len(entries))
	for _, rel := range entries {
		mod, err := loadMod(p.Dir, rel)
		if err != nil {
			// One unreadable metafile should not hide the rest of the
			// pack, so it is reported in place rather than aborting.
			mods = append(mods, Mod{
				Name:     filepath.Base(rel),
				Path:     rel,
				LoadErr:  err,
				SideFlag: SideBoth,
			})
			continue
		}
		mods = append(mods, mod)
	}

	sort.Slice(mods, func(i, j int) bool {
		return strings.ToLower(mods[i].Name) < strings.ToLower(mods[j].Name)
	})
	return mods, nil
}

// metafiles lists the index entries that are per-mod metadata files.
func (p Pack) metafiles() ([]string, error) {
	data, err := os.ReadFile(filepath.Join(p.Dir, p.IndexFile))
	if err != nil {
		return nil, err
	}

	var idx indexFile
	if err := toml.Unmarshal(data, &idx); err != nil {
		return nil, &MalformedError{Path: p.IndexFile, Err: err}
	}

	out := make([]string, 0, len(idx.Files))
	for _, f := range idx.Files {
		if f.Metafile && f.File != "" {
			out = append(out, f.File)
		}
	}
	return out, nil
}

// Files lists the pack's non-mod included files, as the file tree shows
// them alongside the pack's own TOML.
func (p Pack) Files() ([]string, error) {
	data, err := os.ReadFile(filepath.Join(p.Dir, p.IndexFile))
	if err != nil {
		return nil, err
	}

	var idx indexFile
	if err := toml.Unmarshal(data, &idx); err != nil {
		return nil, &MalformedError{Path: p.IndexFile, Err: err}
	}

	out := make([]string, 0, len(idx.Files))
	for _, f := range idx.Files {
		if !f.Metafile && f.File != "" {
			out = append(out, f.File)
		}
	}
	sort.Strings(out)
	return out, nil
}
