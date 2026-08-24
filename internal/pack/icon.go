package pack

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// iconNames are the filenames looked for in a pack folder, in the order
// they win. The logo is optional: a pack without one shows a placeholder.
var iconNames = []string{"icon.png", "pack.png"}

// IconPath returns the pack logo's path, or an empty string when the pack
// has none.
func IconPath(dir string) string {
	for _, name := range iconNames {
		path := filepath.Join(dir, name)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path
		}
	}
	return ""
}

// IconName is the filename a logo is stored under when the user sets one.
// It lives in the pack folder so it is committed with the pack and can be
// attached to a release.
const IconName = "icon.png"

// ErrNotPNG reports a logo file that is not a PNG. Only PNG is accepted:
// it is what Modrinth packs use, and converting formats would mean
// pulling in an image encoder for a cosmetic feature.
var ErrNotPNG = errors.New("the pack logo must be a PNG file")

// SetIcon copies src into the pack folder as the pack's logo, replacing
// any existing one.
func SetIcon(dir, src string) error {
	if !strings.EqualFold(filepath.Ext(src), ".png") {
		return ErrNotPNG
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	dest := filepath.Join(dir, IconName)
	if sameFile(src, dest) {
		return nil
	}

	// Write to a temporary file first so a failed copy cannot destroy the
	// logo that was already there.
	tmp, err := os.CreateTemp(dir, ".icon-*.png")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	if _, err := io.Copy(tmp, in); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), dest)
}

// RemoveIcon deletes the pack's logo. A pack without one is fine, so a
// missing file is not an error.
func RemoveIcon(dir string) error {
	err := os.Remove(filepath.Join(dir, IconName))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

// sameFile reports whether two paths are the same file on disk, so
// setting a pack's existing logo as its logo does nothing instead of
// truncating it.
func sameFile(a, b string) bool {
	ai, err := os.Stat(a)
	if err != nil {
		return false
	}
	bi, err := os.Stat(b)
	if err != nil {
		return false
	}
	return os.SameFile(ai, bi)
}
