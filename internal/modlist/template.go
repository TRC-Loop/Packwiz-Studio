package modlist

import (
	"strconv"
	"strings"
)

// Placeholder is one token a custom template may use.
type Placeholder struct {
	Token string
	Means string
}

// LinePlaceholders are the tokens that stand for something about a mod.
func LinePlaceholders() []Placeholder {
	return []Placeholder{
		{"{name}", "the mod's name"},
		{"{slug}", "its metafile name, as packwiz calls it"},
		{"{filename}", "the jar it resolves to"},
		{"{side}", "both, client or server"},
		{"{url}", "its Modrinth page, or its download"},
		{"{project}", "the Modrinth project id"},
		{"{version}", "the Modrinth version id"},
		{"{pinned}", "yes when the mod is pinned"},
		{"{optional}", "yes when the mod is optional"},
		{"{index}", "its position in the list, from one"},
	}
}

// PackPlaceholders are the tokens that stand for something about the
// pack. They work in the header, the footer and in a line.
func PackPlaceholders() []Placeholder {
	return []Placeholder{
		{"{pack}", "the pack's name"},
		{"{packversion}", "the pack's version"},
		{"{mcversion}", "the Minecraft version"},
		{"{loader}", "the mod loader"},
		{"{count}", "how many mods are listed"},
		{"{date}", "today's date"},
	}
}

// DefaultTemplate is what the custom format starts as: a markdown
// heading, a bullet per mod and a count, so the placeholders can be seen
// working before anything is typed.
func DefaultTemplate() Spec {
	return Spec{
		Format: Custom,
		Header: "## {pack} {packversion}\n\n{count} mods for {mcversion} {loader}\n",
		Line:   "- {name} ({side})",
		Footer: "",
	}
}

// custom expands the user's own templates.
//
// A line template that is empty renders nothing at all rather than a run
// of blank lines, since an empty box is how a template is left unused.
func custom(meta Meta, entries []Entry, spec Spec) string {
	var b strings.Builder

	if spec.Header != "" {
		b.WriteString(expandPack(spec.Header, meta, len(entries)) + "\n")
	}

	if spec.Line != "" {
		for i, e := range entries {
			line := expandPack(spec.Line, meta, len(entries))
			b.WriteString(expandEntry(line, e, i+1) + "\n")
		}
	}

	if spec.Footer != "" {
		b.WriteString(expandPack(spec.Footer, meta, len(entries)) + "\n")
	}
	return b.String()
}

// expandPack replaces the pack's placeholders.
func expandPack(text string, meta Meta, count int) string {
	return replace(text, map[string]string{
		"{pack}":        meta.Name,
		"{packversion}": meta.Version,
		"{mcversion}":   meta.MCVersion,
		"{loader}":      meta.Loader,
		"{count}":       strconv.Itoa(count),
		"{date}":        meta.Date,
	})
}

// expandEntry replaces one mod's placeholders.
func expandEntry(text string, e Entry, index int) string {
	return replace(text, map[string]string{
		"{name}":     e.Name,
		"{slug}":     e.Slug,
		"{filename}": e.Filename,
		"{side}":     e.Side,
		"{url}":      e.URL,
		"{project}":  e.ProjectID,
		"{version}":  e.VersionID,
		"{pinned}":   yesNo(e.Pinned),
		"{optional}": yesNo(e.Optional),
		"{index}":    strconv.Itoa(index),
	})
}

// replace substitutes every token it is given.
func replace(text string, values map[string]string) string {
	for token, value := range values {
		text = strings.ReplaceAll(text, token, value)
	}
	return text
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
