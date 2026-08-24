package pack

import (
	"errors"
	"image"
	"image/png"
	"os"
	"path/filepath"

	// Decoders for the formats a user is likely to pick. A pack icon is
	// stored as PNG, so anything else is converted rather than refused.
	_ "image/gif"
	_ "image/jpeg"

	_ "golang.org/x/image/webp"
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

// IconName is the filename a logo is stored under. It lives in the pack
// folder so it is committed with the pack and can be attached to a
// release.
const IconName = "icon.png"

// ErrNotAnImage reports a file that could not be read as an image.
var ErrNotAnImage = errors.New(
	"that file is not an image the app can read, try a PNG, JPEG, GIF or WebP")

// SetIcon stores src as the pack's logo, converting it to PNG.
//
// PNG is what Modrinth packs use and what every launcher expects, so the
// file on disk is always a PNG whatever the user picked.
func SetIcon(dir, src string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	decoded, _, err := image.Decode(file)
	if err != nil {
		return ErrNotAnImage
	}

	dest := filepath.Join(dir, IconName)

	// Written to a temporary file first, so a failed conversion cannot
	// destroy the logo that was already there.
	tmp, err := os.CreateTemp(dir, ".icon-*.png")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	if err := png.Encode(tmp, decoded); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		return err
	}

	// A second logo filename would win the lookup, so it is cleared to
	// keep one source of truth.
	clearOtherIcons(dir)

	return os.Rename(tmp.Name(), dest)
}

// clearOtherIcons removes logo files other than the canonical one.
func clearOtherIcons(dir string) {
	for _, name := range iconNames {
		if name == IconName {
			continue
		}
		os.Remove(filepath.Join(dir, name))
	}
}

// RemoveIcon deletes the pack's logo. A pack without one is fine, so a
// missing file is not an error.
func RemoveIcon(dir string) error {
	var firstErr error

	for _, name := range iconNames {
		err := os.Remove(filepath.Join(dir, name))
		if err != nil && !errors.Is(err, os.ErrNotExist) && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// ImageExtensions are the file types the logo picker offers.
var ImageExtensions = []string{".png", ".jpg", ".jpeg", ".gif", ".webp"}
