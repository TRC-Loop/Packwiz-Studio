package instance

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// CopyFile copies one file from one root to the other, creating whatever
// folders it needs on the way.
func CopyFile(fromRoot, toRoot, rel string) error {
	return copyPath(filepath.Join(fromRoot, rel), filepath.Join(toRoot, rel))
}

// Remove deletes one file from a root, which is how a file the instance
// no longer has is taken out of the pack.
func Remove(root, rel string) error {
	err := os.Remove(filepath.Join(root, rel))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// Import copies whole folders from an instance into a pack.
//
// This is the first step of a pack that was built by playing: the configs
// a mod wrote on first run are in the instance, and the pack has none of
// them yet.
func Import(instanceDir, packDir string, folders []string) (int, error) {
	copied := 0

	for _, folder := range folders {
		src := filepath.Join(instanceDir, folder)
		if !Exists(src) {
			continue
		}

		err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}

			rel, err := filepath.Rel(instanceDir, path)
			if err != nil || skip(rel) {
				return nil
			}
			if err := copyPath(path, filepath.Join(packDir, rel)); err != nil {
				return err
			}
			copied++
			return nil
		})
		if err != nil {
			return copied, err
		}
	}
	return copied, nil
}

// copyPath copies one file, keeping its permissions and replacing
// whatever was there.
func copyPath(from, to string) error {
	if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
		return err
	}

	in, err := os.Open(from)
	if err != nil {
		return err
	}
	defer in.Close()

	mode := os.FileMode(0o644)
	if info, err := in.Stat(); err == nil {
		mode = info.Mode().Perm()
	}

	out, err := os.OpenFile(to, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
