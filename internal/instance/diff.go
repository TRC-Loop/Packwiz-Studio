package instance

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// Diff lists every file that differs between a pack and an instance,
// looking only inside the given folders.
//
// Only the named folders are walked. An instance holds saves, logs and
// caches that would swamp the result and must never reach a pack, and
// naming what to look at is a shorter list than naming what to avoid.
func Diff(packDir, instanceDir string, folders []string) ([]Entry, error) {
	packFiles, err := collect(packDir, folders)
	if err != nil {
		return nil, err
	}
	instanceFiles, err := collect(instanceDir, folders)
	if err != nil {
		return nil, err
	}

	var out []Entry

	for rel := range packFiles {
		if _, ok := instanceFiles[rel]; !ok {
			out = append(out, Entry{Rel: rel, State: OnlyInPack})
			continue
		}
		if differs(filepath.Join(packDir, rel), filepath.Join(instanceDir, rel)) {
			out = append(out, Entry{Rel: rel, State: Differs})
		}
	}
	for rel := range instanceFiles {
		if _, ok := packFiles[rel]; !ok {
			out = append(out, Entry{Rel: rel, State: OnlyInInstance})
		}
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].State != out[j].State {
			return out[i].State < out[j].State
		}
		return out[i].Rel < out[j].Rel
	})
	return out, nil
}

// collect lists the files under a root's chosen folders, keyed by their
// path relative to the root and spelled with the platform separator.
func collect(root string, folders []string) (map[string]struct{}, error) {
	found := map[string]struct{}{}

	for _, folder := range folders {
		dir := filepath.Join(root, folder)
		if !Exists(dir) {
			continue
		}

		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				// A folder that cannot be read is skipped rather than
				// failing the comparison: an instance is a live game
				// folder and something in it may be locked.
				return nil
			}
			if d.IsDir() {
				return nil
			}

			rel, err := filepath.Rel(root, path)
			if err != nil || skip(rel) {
				return nil
			}
			found[rel] = struct{}{}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return found, nil
}

// differs reports two files with different contents. A file that cannot
// be read is reported as different, since the safe assumption is that it
// needs looking at.
func differs(a, b string) bool {
	infoA, errA := os.Stat(a)
	infoB, errB := os.Stat(b)
	if errA != nil || errB != nil {
		return true
	}
	if infoA.Size() != infoB.Size() {
		return true
	}

	hashA, errA := hashFile(a)
	hashB, errB := hashFile(b)
	if errA != nil || errB != nil {
		return true
	}
	return hashA != hashB
}
