// Package changelog compares a pack against a previous git tag and turns
// the difference into release notes.
package changelog

import (
	"sort"
	"strings"
)

// ChangeKind is what happened to a mod between two revisions.
type ChangeKind int

// Change kinds.
const (
	// Added is a mod that was not in the previous revision.
	Added ChangeKind = iota
	// Removed is a mod that is no longer in the pack.
	Removed
	// Updated is a mod whose version changed.
	Updated
)

// Entry is one line of a changelog.
type Entry struct {
	Kind ChangeKind
	Name string
	// From and To are the old and new versions, set for an update. They
	// hold whatever the metafile recorded, which is a Modrinth version id
	// unless the mod came from elsewhere.
	From string
	To   string
}

// Snapshot is the mods present at one revision, keyed by metafile path.
// The path is the identity rather than the display name, because a
// project can be renamed without becoming a different mod.
type Snapshot map[string]Mod

// Mod is the part of a mod a changelog cares about.
type Mod struct {
	Name    string
	Version string
}

// Diff compares two snapshots and returns the changes, ordered added,
// then updated, then removed, alphabetically within each group.
func Diff(before, after Snapshot) []Entry {
	var added, updated, removed []Entry

	for path, now := range after {
		was, existed := before[path]
		switch {
		case !existed:
			added = append(added, Entry{Kind: Added, Name: now.Name, To: now.Version})
		case was.Version != now.Version:
			updated = append(updated, Entry{
				Kind: Updated,
				Name: now.Name,
				From: was.Version,
				To:   now.Version,
			})
		}
	}

	for path, was := range before {
		if _, still := after[path]; !still {
			removed = append(removed, Entry{Kind: Removed, Name: was.Name, From: was.Version})
		}
	}

	byName(added)
	byName(updated)
	byName(removed)

	out := make([]Entry, 0, len(added)+len(updated)+len(removed))
	out = append(out, added...)
	out = append(out, updated...)
	return append(out, removed...)
}

// byName sorts entries case insensitively, so a changelog reads
// alphabetically regardless of how mods capitalise their names.
func byName(entries []Entry) {
	sort.Slice(entries, func(i, j int) bool {
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})
}

// Group splits entries by kind, which is what the grouped output format
// and the summary line both need.
func Group(entries []Entry) (added, updated, removed []Entry) {
	for _, e := range entries {
		switch e.Kind {
		case Added:
			added = append(added, e)
		case Updated:
			updated = append(updated, e)
		case Removed:
			removed = append(removed, e)
		}
	}
	return added, updated, removed
}
