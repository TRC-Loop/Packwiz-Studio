package packwiz

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// errNoBinary reports an archive with no packwiz binary in it.
var errNoBinary = errors.New("the downloaded build contains no packwiz binary")

// maxBinarySize caps what will be written to disk, so a malformed or
// hostile archive cannot fill the user's disk.
const maxBinarySize = 200 << 20 // 200 MB

// extractBinary pulls the packwiz executable out of an artifact zip and
// writes it to dest. Only the binary is taken: artifacts also carry a
// licence and a readme the app has no use for.
func extractBinary(archive, dest string) error {
	zr, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}
	defer zr.Close()

	for _, f := range zr.File {
		if f.FileInfo().IsDir() || !isBinaryEntry(f.Name) {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}
		err = writeBinary(rc, dest)
		rc.Close()
		return err
	}
	return errNoBinary
}

// isBinaryEntry matches the packwiz executable inside an archive.
//
// Only the base name is compared: an archive may nest the binary in a
// folder. A name containing a path traversal is rejected outright.
func isBinaryEntry(name string) bool {
	if strings.Contains(name, "..") {
		return false
	}
	base := strings.ToLower(path.Base(filepath.ToSlash(name)))
	return base == "packwiz" || base == "packwiz.exe"
}

// writeBinary writes the extracted stream to dest atomically, so a failed
// extraction cannot leave a half-written binary in place of a working one.
func writeBinary(src io.Reader, dest string) error {
	tmp, err := os.CreateTemp(filepath.Dir(dest), ".packwiz-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	written, err := io.Copy(tmp, io.LimitReader(src, maxBinarySize+1))
	if err != nil {
		tmp.Close()
		return err
	}
	if written > maxBinarySize {
		tmp.Close()
		return errors.New("the packwiz binary in the archive is implausibly large")
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o755); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), dest)
}
