package packwiz

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// errNoBinary reports an archive that did not contain a packwiz binary.
var errNoBinary = errors.New("the downloaded archive contains no packwiz binary")

// maxBinarySize caps what will be written to disk, so a malformed or
// hostile archive cannot fill the user's disk.
const maxBinarySize = 200 << 20 // 200 MB

// extractBinary pulls the packwiz executable out of archive and writes it
// to dest. Only the binary is extracted — release archives also carry a
// licence and readme, which the app has no use for.
func extractBinary(archive, assetName, dest string) error {
	if strings.HasSuffix(strings.ToLower(assetName), ".zip") {
		return extractZip(archive, dest)
	}
	return extractTarGz(archive, dest)
}

func extractZip(archive, dest string) error {
	zr, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}
	defer zr.Close()

	for _, f := range zr.File {
		if !isBinaryEntry(f.Name) {
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

func extractTarGz(archive, dest string) error {
	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return errNoBinary
		}
		if err != nil {
			return err
		}
		if hdr.Typeflag != tar.TypeReg || !isBinaryEntry(hdr.Name) {
			continue
		}
		return writeBinary(tr, dest)
	}
}

// isBinaryEntry matches the packwiz executable inside an archive. Entry
// names are compared on their base only: an archive may nest the binary
// in a versioned folder, and a name with a path traversal in it is
// rejected outright.
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
