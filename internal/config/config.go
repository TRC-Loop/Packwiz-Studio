// Package config persists app-level settings and the list of known packs
// as JSON in the OS user config directory. It never stores secrets —
// release API tokens live in the OS keyring, see internal/secrets.
package config

import "time"

// ViewMode selects the Modrinth browser's presentation.
type ViewMode string

// Browser view modes.
const (
	ViewGrid ViewMode = "grid"
	ViewList ViewMode = "list"
)

// ChangelogFormat selects how a generated changelog is rendered.
type ChangelogFormat string

// Changelog output formats.
const (
	// FormatFlat is one bullet per change.
	FormatFlat ChangelogFormat = "flat"
	// FormatGrouped is bullets under Added/Updated/Removed headings.
	FormatGrouped ChangelogFormat = "grouped"
	// FormatProse is a single running sentence per change type.
	FormatProse ChangelogFormat = "prose"
)

// Config is the whole persisted state of the app.
type Config struct {
	// PackwizPath is the resolved packwiz binary. Empty means "look on
	// PATH at startup".
	PackwizPath string `json:"packwizPath,omitempty"`
	// GiteaBaseURL is an API base for a self-hosted Gitea or Forgejo
	// instance whose host is not a recognised public domain.
	GiteaBaseURL string `json:"giteaBaseUrl,omitempty"`
	// UseKeyring stores release tokens in the OS keyring. When false the
	// app prompts for a token per release instead of persisting it.
	UseKeyring bool `json:"useKeyring"`
	// Packs is the known-pack list backing the launcher's recents.
	Packs []Pack `json:"packs"`
}

// Pack is one known pack: where it lives plus what the launcher shows.
type Pack struct {
	// Path is the directory containing pack.toml. It is the identity of
	// a pack — comparisons are made on this field.
	Path string `json:"path"`
	// Name, MCVersion and Loader are cached from pack.toml so the
	// launcher can render recents without reading every pack on start.
	Name      string `json:"name,omitempty"`
	MCVersion string `json:"mcVersion,omitempty"`
	Loader    string `json:"loader,omitempty"`
	// LastOpened orders the recents list, newest first.
	LastOpened time.Time `json:"lastOpened,omitempty"`
	// Prefs are this pack's remembered choices.
	Prefs Prefs `json:"prefs"`
}

// Prefs are the per-pack settings the UI remembers between sessions.
type Prefs struct {
	// ExportDir prefills the export dialog. The dialog is always shown;
	// this only supplies its default.
	ExportDir string `json:"exportDir,omitempty"`
	// ViewMode is the browser's last-used presentation.
	ViewMode ViewMode `json:"viewMode,omitempty"`
	// ChangelogFormat is the last-used release changelog rendering.
	ChangelogFormat ChangelogFormat `json:"changelogFormat,omitempty"`
	// LastExport is the most recent .mrpack produced for this pack, so a
	// release can attach it without re-exporting.
	LastExport string `json:"lastExport,omitempty"`
}

// defaults returns a Config for a first run.
func defaults() Config {
	return Config{UseKeyring: true}
}

// withDefaults fills zero-valued fields that need a non-zero default.
func (p Prefs) withDefaults() Prefs {
	if p.ViewMode == "" {
		p.ViewMode = ViewGrid
	}
	if p.ChangelogFormat == "" {
		p.ChangelogFormat = FormatGrouped
	}
	return p
}

// clone returns a deep copy, so callers cannot mutate stored state by
// holding on to a returned Config.
func (c Config) clone() Config {
	out := c
	if c.Packs != nil {
		out.Packs = make([]Pack, len(c.Packs))
		copy(out.Packs, c.Packs)
	}
	return out
}
