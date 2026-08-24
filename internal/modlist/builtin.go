package modlist

import (
	"encoding/json"
	"strings"
)

// markdownList is one bullet per mod, linked where a link is known.
func markdownList(entries []Entry) string {
	var b strings.Builder

	for _, e := range entries {
		b.WriteString("- " + link(e))
		if note := notes(e); note != "" {
			b.WriteString(" (" + note + ")")
		}
		b.WriteString("\n")
	}
	return b.String()
}

// markdownTable is a column per detail, which is what a wiki page wants.
func markdownTable(entries []Entry) string {
	var b strings.Builder

	b.WriteString("| Mod | Side | Version | File |\n")
	b.WriteString("| --- | --- | --- | --- |\n")

	for _, e := range entries {
		b.WriteString("| " + link(e) +
			" | " + orNone(e.Side) +
			" | " + orNone(e.VersionID) +
			" | " + orNone(e.Filename) + " |\n")
	}
	return b.String()
}

// plain is one name per line.
func plain(entries []Entry) string {
	var b strings.Builder
	for _, e := range entries {
		b.WriteString(e.Name + "\n")
	}
	return b.String()
}

// csv is a spreadsheet. Fields are quoted and inner quotes doubled, which
// is all the escaping the format has.
func csv(entries []Entry) string {
	var b strings.Builder

	b.WriteString("name,slug,side,filename,project,version,url\n")
	for _, e := range entries {
		fields := []string{e.Name, e.Slug, e.Side, e.Filename,
			e.ProjectID, e.VersionID, e.URL}

		for i, f := range fields {
			if i > 0 {
				b.WriteString(",")
			}
			b.WriteString(`"` + strings.ReplaceAll(f, `"`, `""`) + `"`)
		}
		b.WriteString("\n")
	}
	return b.String()
}

// jsonList is the whole list as data, pack details included, for a script
// to read rather than a person.
func jsonList(meta Meta, entries []Entry) string {
	type mod struct {
		Name      string `json:"name"`
		Slug      string `json:"slug"`
		Side      string `json:"side"`
		Filename  string `json:"filename,omitempty"`
		ProjectID string `json:"projectId,omitempty"`
		VersionID string `json:"versionId,omitempty"`
		URL       string `json:"url,omitempty"`
		Pinned    bool   `json:"pinned,omitempty"`
		Optional  bool   `json:"optional,omitempty"`
	}

	doc := struct {
		Pack      string `json:"pack"`
		Version   string `json:"version,omitempty"`
		MCVersion string `json:"minecraft,omitempty"`
		Loader    string `json:"loader,omitempty"`
		Count     int    `json:"count"`
		Mods      []mod  `json:"mods"`
	}{
		Pack:      meta.Name,
		Version:   meta.Version,
		MCVersion: meta.MCVersion,
		Loader:    meta.Loader,
		Count:     len(entries),
		Mods:      make([]mod, 0, len(entries)),
	}

	for _, e := range entries {
		doc.Mods = append(doc.Mods, mod{
			Name: e.Name, Slug: e.Slug, Side: e.Side, Filename: e.Filename,
			ProjectID: e.ProjectID, VersionID: e.VersionID, URL: e.URL,
			Pinned: e.Pinned, Optional: e.Optional,
		})
	}

	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		// The document is made of strings and bools, so this cannot fail.
		// Reporting the error would be noise in a signature nobody else
		// needs to handle.
		return ""
	}
	return string(out) + "\n"
}

// html is a list of links, ready to paste into a page.
func html(entries []Entry) string {
	var b strings.Builder

	b.WriteString("<ul>\n")
	for _, e := range entries {
		name := escapeHTML(e.Name)
		if e.URL != "" {
			name = `<a href="` + escapeHTML(e.URL) + `">` + name + `</a>`
		}
		b.WriteString("  <li>" + name + "</li>\n")
	}
	b.WriteString("</ul>\n")
	return b.String()
}

// bbcode is a forum post list.
func bbcode(entries []Entry) string {
	var b strings.Builder

	b.WriteString("[list]\n")
	for _, e := range entries {
		if e.URL != "" {
			b.WriteString("[*][url=" + e.URL + "]" + e.Name + "[/url]\n")
			continue
		}
		b.WriteString("[*]" + e.Name + "\n")
	}
	b.WriteString("[/list]\n")
	return b.String()
}

// link is a markdown link when there is somewhere to link to.
func link(e Entry) string {
	if e.URL == "" {
		return e.Name
	}
	return "[" + e.Name + "](" + e.URL + ")"
}

// notes describes a mod that is not a plain both-sides install.
func notes(e Entry) string {
	var parts []string
	if e.Side != "" && e.Side != "both" {
		parts = append(parts, e.Side+" only")
	}
	if e.Optional {
		parts = append(parts, "optional")
	}
	return strings.Join(parts, ", ")
}

func orNone(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

// escapeHTML escapes the three characters that would otherwise close a
// tag or an attribute.
func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return strings.ReplaceAll(s, `"`, "&quot;")
}
