package git

import (
	"path"
	"strconv"
	"strings"
)

// CommitMessage describes a set of changes in one imperative line.
//
// It is built from the paths git reports, not from a diff: in a packwiz
// pack the path is the fact that matters, because one metafile is one
// mod. A new .pw.toml is a mod added, a deleted one is a mod removed, and
// a modified one is that mod moved to another version.
//
// The result is a suggestion. It is put in the message box for the user
// to edit rather than committed on its own.
func CommitMessage(changes []Change) string {
	changes = forMessage(changes)

	var added, removed, updated, other []string
	meta := false

	for _, c := range changes {
		name, isMod := modName(c.Path)
		switch {
		case !isMod && isPackFile(c.Path):
			meta = true
		case !isMod:
			other = append(other, path.Base(c.Path))
		case c.Label() == "new" || c.Label() == "added":
			added = append(added, name)
		case c.Label() == "deleted":
			removed = append(removed, name)
		default:
			updated = append(updated, name)
		}
	}

	var parts []string
	if len(added) > 0 {
		parts = append(parts, "add "+names(added, "mod"))
	}
	if len(removed) > 0 {
		parts = append(parts, "remove "+names(removed, "mod"))
	}
	if len(updated) > 0 {
		parts = append(parts, "update "+names(updated, "mod"))
	}
	if meta && len(parts) == 0 {
		parts = append(parts, "update pack metadata")
	}
	if len(other) > 0 && len(parts) == 0 {
		parts = append(parts, "update "+names(other, "file"))
	}

	if len(parts) == 0 {
		return ""
	}
	return upperFirst(strings.Join(parts, ", "))
}

// forMessage picks which changes the message describes: what is staged
// when anything is, since that is what a commit would record, and
// otherwise everything.
func forMessage(changes []Change) []Change {
	var staged []Change
	for _, c := range changes {
		if c.Staged {
			staged = append(staged, c)
		}
	}
	if len(staged) > 0 {
		return staged
	}
	return changes
}

// modName reports the mod a metafile belongs to.
func modName(p string) (string, bool) {
	base := path.Base(p)
	if !strings.HasSuffix(base, ".pw.toml") {
		return "", false
	}
	return strings.TrimSuffix(base, ".pw.toml"), true
}

// isPackFile reports a change to the pack's own files, which describes
// itself no better than "metadata".
func isPackFile(p string) bool {
	base := path.Base(p)
	return base == "pack.toml" || base == "index.toml"
}

// names lists a few things by name and counts the rest, so a message
// stays one readable line however much changed.
func names(list []string, noun string) string {
	const shown = 3

	if len(list) > shown {
		return strconv.Itoa(len(list)) + " " + noun + "s"
	}
	if len(list) == 1 {
		return list[0]
	}
	return strings.Join(list[:len(list)-1], ", ") + " and " + list[len(list)-1]
}

// upperFirst capitalises the first letter, leaving mod names alone.
func upperFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
