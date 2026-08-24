// Package pack reads a packwiz pack's own metadata. It only reads: every
// change to a pack goes through the packwiz CLI, so nothing here writes
// pack.toml or the index.
package pack

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/BurntSushi/toml"
)

// FileName is the file that marks a folder as a packwiz pack.
const FileName = "pack.toml"

// Pack is the metadata in a pack.toml, plus where it was read from.
type Pack struct {
	// Dir is the folder holding pack.toml.
	Dir string
	// Name, Author and Version are the pack's own metadata. All three
	// may be empty: packwiz does not require them to be filled in.
	Name    string
	Author  string
	Version string
	// MCVersion is the Minecraft version from the versions table.
	MCVersion string
	// Loader is the mod loader name, taken from whichever key in the
	// versions table is not minecraft.
	Loader string
	// LoaderVersion is that loader's pinned version.
	LoaderVersion string
	// Format is the pack-format string, for example "packwiz:1.1.0".
	Format string
	// IndexFile is the index's filename relative to Dir.
	IndexFile string
}

// packFile mirrors the parts of pack.toml the app reads.
type packFile struct {
	Name       string            `toml:"name"`
	Author     string            `toml:"author"`
	Version    string            `toml:"version"`
	PackFormat string            `toml:"pack-format"`
	Index      packIndex         `toml:"index"`
	Versions   map[string]string `toml:"versions"`
}

type packIndex struct {
	File string `toml:"file"`
}

// ErrNotAPack reports a folder with no pack.toml in it. The open-pack
// picker turns this into a message naming the folder.
var ErrNotAPack = errors.New("no pack.toml in this folder")

// Load reads the pack.toml in dir.
func Load(dir string) (Pack, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return Pack{}, err
	}

	path := filepath.Join(abs, FileName)
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return Pack{}, ErrNotAPack
	}
	if err != nil {
		return Pack{}, err
	}

	var pf packFile
	if err := toml.Unmarshal(data, &pf); err != nil {
		return Pack{}, &MalformedError{Path: path, Err: err}
	}

	p := Pack{
		Dir:       abs,
		Name:      pf.Name,
		Author:    pf.Author,
		Version:   pf.Version,
		Format:    pf.PackFormat,
		IndexFile: pf.Index.File,
		MCVersion: pf.Versions["minecraft"],
	}
	if p.IndexFile == "" {
		p.IndexFile = "index.toml"
	}
	p.Loader, p.LoaderVersion = loaderFrom(pf.Versions)

	// A pack with no name is common enough in hand-made packs that it is
	// not an error. Fall back to the folder name so the launcher and the
	// pack window have something to show.
	if p.Name == "" {
		p.Name = filepath.Base(abs)
	}
	return p, nil
}

// IsPack reports whether dir holds a pack.toml, without parsing it. The
// picker uses this to reject a folder before doing any real work.
func IsPack(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, FileName))
	return err == nil && !info.IsDir()
}

// MalformedError reports a pack.toml that could not be parsed.
type MalformedError struct {
	Path string
	Err  error
}

func (e *MalformedError) Error() string {
	return e.Path + " could not be read: " + e.Err.Error()
}

func (e *MalformedError) Unwrap() error { return e.Err }

// loaderFrom picks the mod loader out of the versions table. packwiz
// stores minecraft alongside exactly one loader, but the key order is not
// guaranteed, so remaining keys are sorted for a stable answer.
func loaderFrom(versions map[string]string) (name, version string) {
	names := make([]string, 0, len(versions))
	for k := range versions {
		if k != "minecraft" {
			names = append(names, k)
		}
	}
	if len(names) == 0 {
		return "", ""
	}
	sort.Strings(names)
	return names[0], versions[names[0]]
}
