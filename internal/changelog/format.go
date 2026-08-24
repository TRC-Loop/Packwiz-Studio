package changelog

import (
	"strings"

	"github.com/TRC-Loop/Packwiz-Studio/internal/config"
)

// Render turns entries into release notes in the chosen format. The result
// is a starting point: the release form lets the user edit it before
// anything is published.
func Render(entries []Entry, format config.ChangelogFormat) string {
	if len(entries) == 0 {
		return ""
	}

	switch format {
	case config.FormatGrouped:
		return grouped(entries)
	case config.FormatProse:
		return prose(entries)
	default:
		return flat(entries)
	}
}

// flat is one bullet per change.
func flat(entries []Entry) string {
	var b strings.Builder

	for _, e := range entries {
		b.WriteString("- ")
		b.WriteString(line(e))
		b.WriteString("\n")
	}
	return b.String()
}

// grouped puts bullets under Added, Updated and Removed headings. Empty
// groups are left out rather than shown with nothing under them.
func grouped(entries []Entry) string {
	added, updated, removed := Group(entries)

	var b strings.Builder
	section(&b, "Added", added, func(e Entry) string { return withVersion(e.Name, e.To) })
	section(&b, "Updated", updated, func(e Entry) string {
		return e.Name + " " + versionRange(e.From, e.To)
	})
	section(&b, "Removed", removed, func(e Entry) string { return e.Name })

	return strings.TrimRight(b.String(), "\n") + "\n"
}

// section writes one heading and its bullets.
func section(b *strings.Builder, title string, entries []Entry, render func(Entry) string) {
	if len(entries) == 0 {
		return
	}

	b.WriteString("### ")
	b.WriteString(title)
	b.WriteString("\n")

	for _, e := range entries {
		b.WriteString("- ")
		b.WriteString(render(e))
		b.WriteString("\n")
	}
	b.WriteString("\n")
}

// prose is one sentence per change kind.
func prose(entries []Entry) string {
	added, updated, removed := Group(entries)

	var sentences []string
	if len(added) > 0 {
		sentences = append(sentences, "Added "+names(added)+".")
	}
	if len(updated) > 0 {
		parts := make([]string, 0, len(updated))
		for _, e := range updated {
			parts = append(parts, e.Name+" ("+versionRange(e.From, e.To)+")")
		}
		sentences = append(sentences, "Updated "+joinList(parts)+".")
	}
	if len(removed) > 0 {
		sentences = append(sentences, "Removed "+names(removed)+".")
	}
	return strings.Join(sentences, " ") + "\n"
}

// line renders one entry for the flat format.
func line(e Entry) string {
	switch e.Kind {
	case Added:
		return "added " + withVersion(e.Name, e.To)
	case Removed:
		return "removed " + e.Name
	default:
		return "updated " + e.Name + " " + versionRange(e.From, e.To)
	}
}

// withVersion appends a version when there is one to append.
func withVersion(name, version string) string {
	if version == "" {
		return name
	}
	return name + " " + version
}

// versionRange renders an old and new version pair. An unknown side is
// left out rather than printed as an empty string.
func versionRange(from, to string) string {
	switch {
	case from == "" && to == "":
		return "to a new version"
	case from == "":
		return "to " + to
	case to == "":
		return "from " + from
	default:
		return from + " to " + to
	}
}

// names lists entry names for a prose sentence.
func names(entries []Entry) string {
	parts := make([]string, 0, len(entries))
	for _, e := range entries {
		parts = append(parts, e.Name)
	}
	return joinList(parts)
}

// joinList joins with commas and a final and, so a sentence reads.
func joinList(parts []string) string {
	switch len(parts) {
	case 0:
		return ""
	case 1:
		return parts[0]
	case 2:
		return parts[0] + " and " + parts[1]
	default:
		return strings.Join(parts[:len(parts)-1], ", ") + " and " + parts[len(parts)-1]
	}
}
