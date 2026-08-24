package config

// BrowserPrefs is what the Modrinth browser shows about each result.
//
// A result can carry a lot: categories, a licence, which sides it
// supports, download counts. Showing everything makes a dense list
// unreadable, so what appears is a choice, and the default is the two
// things that matter when picking a mod.
type BrowserPrefs struct {
	// ShowTags shows a mod's categories. On by default.
	ShowTags bool `json:"showTags"`
	// ShowLicense shows a mod's licence. On by default, since it decides
	// whether a pack can redistribute the mod at all.
	ShowLicense bool `json:"showLicense"`
	// ShowSides shows which sides a mod supports. Off by default: it
	// matters when building a server pack and is noise otherwise.
	ShowSides bool `json:"showSides"`
}

// defaultBrowserPrefs is what a first run shows.
func defaultBrowserPrefs() BrowserPrefs {
	return BrowserPrefs{ShowTags: true, ShowLicense: true}
}

// Any reports whether the browser shows any detail line at all, so the
// list can skip the row entirely rather than leaving a gap.
func (b BrowserPrefs) Any() bool {
	return b.ShowTags || b.ShowLicense || b.ShowSides
}
