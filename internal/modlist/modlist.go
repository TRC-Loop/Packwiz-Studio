// Package modlist renders a pack's mods as a document: a markdown list
// for a repository page, a table for a wiki, CSV or JSON for a script, or
// whatever a custom template with placeholders asks for.
//
// It only formats. Reading the pack and writing the file happen at the
// edges, so the same renderer can serve a preview and a save.
package modlist

import (
	"sort"
	"strings"
)

// Entry is one mod in a list.
type Entry struct {
	// Name is the mod's display name.
	Name string
	// Slug is the metafile name without its suffix, which is what packwiz
	// commands call the mod.
	Slug string
	// Filename is the jar the metafile resolves to.
	Filename string
	// Side is which side the mod installs on: both, client or server.
	Side string
	// URL is the mod's page when it came from Modrinth, and its download
	// otherwise.
	URL string
	// ProjectID and VersionID identify the Modrinth project and version.
	ProjectID string
	VersionID string
	// Pinned and Optional mirror the metafile's flags.
	Pinned   bool
	Optional bool
}

// Meta describes the pack the list came from.
type Meta struct {
	Name      string
	Version   string
	MCVersion string
	Loader    string
	// Date is the day the list was made, already formatted.
	Date string
}

// Format selects a rendering.
type Format string

// The available formats. These are stored in config, so the values are
// stable strings rather than an ordinal.
const (
	// MarkdownList is one bullet per mod, linked where there is a link.
	MarkdownList Format = "markdown-list"
	// MarkdownTable is a table with a column per detail.
	MarkdownTable Format = "markdown-table"
	// Plain is one name per line, with nothing around it.
	Plain Format = "plain"
	// CSV is a spreadsheet, one row per mod.
	CSV Format = "csv"
	// JSON is an array of objects, for a script to read.
	JSON Format = "json"
	// HTML is an unordered list of links.
	HTML Format = "html"
	// BBCode is a forum post list.
	BBCode Format = "bbcode"
	// Custom is the user's own template, expanded per mod.
	Custom Format = "custom"
)

// Choice is a format and the name the UI shows for it.
type Choice struct {
	Format Format
	Label  string
	// Ext is the extension a saved file gets.
	Ext string
}

// Choices lists the formats in the order the picker offers them.
func Choices() []Choice {
	return []Choice{
		{MarkdownList, "Markdown list", ".md"},
		{MarkdownTable, "Markdown table", ".md"},
		{Plain, "Plain text", ".txt"},
		{CSV, "CSV", ".csv"},
		{JSON, "JSON", ".json"},
		{HTML, "HTML list", ".html"},
		{BBCode, "BBCode", ".txt"},
		{Custom, "Custom template", ".txt"},
	}
}

// Find returns a format's choice, falling back to the first one so a
// value read from an older config never leaves the picker empty.
func Find(f Format) Choice {
	for _, c := range Choices() {
		if c.Format == f {
			return c
		}
	}
	return Choices()[0]
}

// ByLabel resolves the name shown in the picker.
func ByLabel(label string) Choice {
	for _, c := range Choices() {
		if c.Label == label {
			return c
		}
	}
	return Choices()[0]
}

// Spec is what to render and how.
type Spec struct {
	Format Format
	// Header, Line and Footer are the custom format's templates. Line is
	// expanded once per mod; the other two once each, and are omitted when
	// empty.
	Header string
	Line   string
	Footer string
}

// Render turns a pack's mods into a document.
func Render(meta Meta, entries []Entry, spec Spec) string {
	sorted := make([]Entry, len(entries))
	copy(sorted, entries)
	sort.Slice(sorted, func(i, j int) bool {
		return strings.ToLower(sorted[i].Name) < strings.ToLower(sorted[j].Name)
	})

	switch spec.Format {
	case MarkdownTable:
		return markdownTable(sorted)
	case Plain:
		return plain(sorted)
	case CSV:
		return csv(sorted)
	case JSON:
		return jsonList(meta, sorted)
	case HTML:
		return html(sorted)
	case BBCode:
		return bbcode(sorted)
	case Custom:
		return custom(meta, sorted, spec)
	default:
		return markdownList(sorted)
	}
}
