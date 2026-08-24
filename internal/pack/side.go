package pack

import (
	"os"
	"path/filepath"
	"strings"
)

// SetSide rewrites a mod's side field in its metafile.
//
// This is the one place the app writes a pack file itself. packwiz has no
// command for the side flag, so there is nothing to shell out to. The
// edit is deliberately line based rather than a TOML round trip: parsing
// and re-serialising would drop comments and reorder keys in a file the
// user may also be editing by hand.
//
// The index records a hash per metafile, so the caller must run
// packwiz refresh afterwards.
func SetSide(dir, modPath string, side Side) error {
	full := filepath.Join(dir, filepath.FromSlash(modPath))

	data, err := os.ReadFile(full)
	if err != nil {
		return err
	}

	updated := replaceSide(string(data), side)
	if updated == string(data) {
		return nil
	}

	info, err := os.Stat(full)
	if err != nil {
		return err
	}
	return os.WriteFile(full, []byte(updated), info.Mode().Perm())
}

// replaceSide sets the top level side key, adding it when absent.
//
// Only the region before the first table header is considered: a side key
// inside [download] or [update] would belong to that table, not to the
// mod.
func replaceSide(content string, side Side) string {
	lines := strings.Split(content, "\n")
	assignment := `side = "` + string(side) + `"`

	insertAt := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "[") {
			break
		}
		if key(trimmed) == "side" {
			lines[i] = assignment
			return strings.Join(lines, "\n")
		}
		if key(trimmed) == "name" {
			insertAt = i + 1
		}
	}

	// No side key yet. Put it after the name for readability, or at the
	// very top when there is no name either.
	if insertAt < 0 {
		insertAt = 0
	}
	lines = append(lines, "")
	copy(lines[insertAt+1:], lines[insertAt:])
	lines[insertAt] = assignment

	return strings.Join(lines, "\n")
}

// key returns the bare key of a TOML assignment, or an empty string when
// the line is not one.
func key(line string) string {
	if strings.HasPrefix(line, "#") {
		return ""
	}
	name, _, found := strings.Cut(line, "=")
	if !found {
		return ""
	}
	return strings.TrimSpace(name)
}
