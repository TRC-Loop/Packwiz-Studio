package pack

import (
	"os"
	"path/filepath"
	"strings"
)

// Properties are the pack.toml fields a user edits directly.
type Properties struct {
	Name    string
	Author  string
	Version string
}

// SetProperties rewrites the pack's own metadata in pack.toml.
//
// packwiz has no command for this: init writes these fields once and
// nothing updates them afterwards, so the app edits the file. As with the
// side flag, the edit is line based rather than a TOML round trip, so
// comments, key order and anything the app does not understand survive.
//
// pack.toml is not hashed by the index, so no refresh is needed after.
func SetProperties(dir string, p Properties) error {
	path := filepath.Join(dir, FileName)

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	updated := string(data)
	for key, value := range map[string]string{
		"name":    p.Name,
		"author":  p.Author,
		"version": p.Version,
	} {
		updated = setTopLevel(updated, key, value)
	}

	if updated == string(data) {
		return nil
	}

	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(updated), info.Mode().Perm())
}

// setTopLevel assigns a top level string key, adding it when absent.
//
// Only the region before the first table header is considered: a name key
// inside [index] or [versions] belongs to that table, not to the pack.
func setTopLevel(content, key, value string) string {
	lines := strings.Split(content, "\n")
	assignment := key + ` = "` + escapeTOML(value) + `"`

	insertAt := 0
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "[") {
			break
		}
		if tomlKey(trimmed) == key {
			lines[i] = assignment
			return strings.Join(lines, "\n")
		}
		if trimmed != "" {
			insertAt = i + 1
		}
	}

	lines = append(lines, "")
	copy(lines[insertAt+1:], lines[insertAt:])
	lines[insertAt] = assignment

	return strings.Join(lines, "\n")
}

// tomlKey returns the bare key of an assignment, or an empty string when
// the line is not one.
func tomlKey(line string) string {
	if strings.HasPrefix(line, "#") {
		return ""
	}
	name, _, found := strings.Cut(line, "=")
	if !found {
		return ""
	}
	return strings.TrimSpace(name)
}

// escapeTOML escapes a value for a basic TOML string.
func escapeTOML(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	value = strings.ReplaceAll(value, "\n", `\n`)
	return value
}
