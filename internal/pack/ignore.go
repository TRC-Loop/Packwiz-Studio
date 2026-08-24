package pack

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Ignore file names.
const (
	// GitIgnoreFile keeps files out of the repository.
	GitIgnoreFile = ".gitignore"
	// PackwizIgnoreFile keeps files out of the pack index.
	//
	// packwiz refresh indexes every file it finds in the pack folder, so
	// without this the repository's own files end up shipped inside the
	// exported pack.
	PackwizIgnoreFile = ".packwizignore"
)

// ReadIgnore returns an ignore file's contents. A file that does not
// exist reads as empty, which is the same thing as far as the rules go.
func ReadIgnore(dir, name string) (string, error) {
	data, err := os.ReadFile(filepath.Join(dir, name))
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteIgnore replaces an ignore file, or deletes it when the rules are
// emptied out. It always ends the file with a newline, since a rule
// without one is a common way to have the last line quietly ignored by
// other tools.
func WriteIgnore(dir, name, content string) error {
	path := filepath.Join(dir, name)

	body := strings.TrimSpace(content)
	if body == "" {
		err := os.Remove(path)
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}
	return os.WriteFile(path, []byte(body+"\n"), 0o644)
}

// DefaultGitIgnore is what a pack repository usually wants out of git:
// editor and OS clutter, and the exported pack itself, which is a build
// product rebuilt from the metafiles.
func DefaultGitIgnore() []string {
	return []string{
		".DS_Store",
		"Thumbs.db",
		".idea/",
		".vscode/",
		"*.mrpack",
		"dist/",
	}
}

// DefaultPackwizIgnore is what a pack usually wants out of its index: the
// repository's own plumbing and documentation, none of which belongs in
// somebody's game folder.
func DefaultPackwizIgnore() []string {
	return []string{
		".git/**",
		".github/**",
		".gitignore",
		".packwizignore",
		"README.md",
		"LICENSE",
		"*.mrpack",
		".DS_Store",
	}
}

// AddRules appends whichever rules are missing, leaving the existing text
// and its comments alone. Rules already present in any form are left
// where they are rather than moved to the end.
func AddRules(content string, rules []string) string {
	have := map[string]bool{}
	for _, line := range strings.Split(content, "\n") {
		have[strings.TrimSpace(line)] = true
	}

	var add []string
	for _, r := range rules {
		if !have[r] {
			add = append(add, r)
		}
	}
	if len(add) == 0 {
		return content
	}

	body := strings.TrimRight(content, "\n")
	if body != "" {
		body += "\n"
	}
	return body + strings.Join(add, "\n") + "\n"
}
