// Package instance compares a pack folder against a Minecraft instance
// and copies files between the two.
//
// This is the loop a modpack is actually built in: launch the pack, tweak
// a config or a script in game, then get the changed file back into the
// repository. packwiz has nothing for it, because it only knows about the
// pack folder, so the comparison and the copying live here.
package instance

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// State is how one file differs between the pack and the instance.
type State int

// The differences worth reporting. A file that is the same in both is not
// listed at all.
const (
	// Differs is a file both sides have, with different contents.
	Differs State = iota
	// OnlyInInstance is a file the instance has and the pack does not,
	// which is usually a config a mod wrote on first run.
	OnlyInInstance
	// OnlyInPack is a file the pack ships that the instance has lost, or
	// that the instance has not been launched since.
	OnlyInPack
)

// Label describes a state in the words the list shows.
func (s State) Label() string {
	switch s {
	case OnlyInInstance:
		return "only in the instance"
	case OnlyInPack:
		return "only in the pack"
	default:
		return "different"
	}
}

// Entry is one differing file.
type Entry struct {
	// Rel is the path relative to both folders.
	Rel string
	// State is how it differs.
	State State
}

// Folders are the parts of an instance a pack carries. Mods are absent on
// purpose: those are packwiz's business, and copying jars in by hand is
// how a pack ends up with mods nothing tracks.
func Folders() []string {
	return []string{
		"config",
		"defaultconfigs",
		"kubejs",
		"scripts",
		"resourcepacks",
		"shaderpacks",
		"datapacks",
		"global_packs",
		"patchouli_books",
	}
}

// skipNames are files never compared or copied: editor and OS clutter,
// and the per-instance state that has no business in a pack.
var skipNames = map[string]bool{
	".DS_Store": true, "Thumbs.db": true, "options.txt": true,
	"servers.dat": true, "usercache.json": true,
}

// skipExts are extensions never compared or copied.
var skipExts = map[string]bool{
	".log": true, ".lock": true, ".tmp": true, ".bak": true, ".dat_old": true,
}

// skip reports a path not worth syncing.
func skip(rel string) bool {
	base := filepath.Base(rel)

	if skipNames[base] || strings.HasPrefix(base, ".~") {
		return true
	}
	return skipExts[strings.ToLower(filepath.Ext(base))]
}

// hashFile is the content hash used to tell two files apart. Size alone
// would miss a one character edit, which is exactly the edit this is
// looking for.
func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	sum := sha256.New()
	if _, err := io.Copy(sum, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(sum.Sum(nil)), nil
}

// Exists reports a folder that is there and is a folder.
func Exists(dir string) bool {
	info, err := os.Stat(dir)
	return err == nil && info.IsDir()
}
