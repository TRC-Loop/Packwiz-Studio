// Package syntax splits source text into coloured tokens.
//
// These are lexers, not parsers. The editor colours a file while it is
// being typed into, so a half-written line has to come out looking
// reasonable rather than failing: nothing here reports an error, and
// anything unrecognised is returned as plain text.
//
// One generic scanner does the work for every language. A language is a
// small set of rules it consults, which is enough for the file types a
// modpack holds: TOML, JSON, JSON5, YAML, JavaScript, SNBT and the
// properties files Minecraft is full of.
package syntax

import (
	"path/filepath"
	"strings"
)

// Kind is what a token is, which decides how it is drawn.
type Kind int

// Token kinds.
const (
	// KindText is anything unclassified, including whitespace.
	KindText Kind = iota
	// KindComment is a comment, to the end of the line or in a block.
	KindComment
	// KindTable is a section header: a TOML table, an ini section.
	KindTable
	// KindKey is the name on the left of an assignment.
	KindKey
	// KindString is a quoted string.
	KindString
	// KindNumber is a numeric or date literal.
	KindNumber
	// KindBool is a language constant: true, false, null, nil.
	KindBool
	// KindKeyword is a reserved word.
	KindKeyword
	// KindPunct is structural: equals, colons, brackets, braces, commas.
	KindPunct
)

// Token is a run of text of one kind.
type Token struct {
	Text string
	Kind Kind
}

// Lang is one language's rules.
type Lang struct {
	// Name is what the editor calls this language.
	Name string

	// lineComment starts a comment that runs to the end of the line.
	lineComment []string
	// blockOpen and blockClose delimit a comment spanning lines. Empty
	// when the language has none.
	blockOpen  string
	blockClose string
	// quotes are the characters that open a string.
	quotes string
	// assign are the characters that end a key, so that a word or string
	// followed by one is coloured as a key rather than a value.
	assign string
	// tableHeaders colours a line that opens with a bracket as a section
	// header, which is how TOML and ini files are read.
	tableHeaders bool
	// keywords are reserved words, and constants is the subset drawn as
	// language constants.
	keywords  map[string]bool
	constants map[string]bool
	// completions are the words this language offers before anything has
	// been typed, such as the keys a packwiz metafile accepts.
	completions []string
}

// For picks the language for a file path, by extension and then by name.
// An unknown file gets a language that colours strings and comments the
// way most configuration formats spell them, which is more useful than
// leaving it entirely plain.
func For(path string) Lang {
	base := strings.ToLower(filepath.Base(path))

	switch base {
	case "pack.toml":
		return packToml()
	case "index.toml":
		return toml()
	}
	if strings.HasSuffix(base, ".pw.toml") {
		return metafileToml()
	}

	switch strings.TrimPrefix(filepath.Ext(base), ".") {
	case "toml":
		return toml()
	case "json":
		return jsonLang()
	case "json5", "jsonc":
		return json5()
	case "yaml", "yml":
		return yaml()
	case "js", "mjs", "cjs", "ts":
		return javascript()
	case "snbt", "nbt":
		return snbt()
	case "properties", "cfg", "conf", "ini", "toml5":
		return properties()
	default:
		return generic()
	}
}

// Completions are the words this language suggests, sorted by the caller.
func (l Lang) Completions() []string { return l.completions }
