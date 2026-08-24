package config

// ModListPrefs is the last-used mod list export choice.
//
// The templates are kept even when another format is selected, so
// switching away to check a built-in format and back does not throw away
// a template that took a while to get right.
type ModListPrefs struct {
	// Format is a modlist format value.
	Format string `json:"format,omitempty"`
	// Header, Line and Footer are the custom format's templates.
	Header string `json:"header,omitempty"`
	Line   string `json:"line,omitempty"`
	Footer string `json:"footer,omitempty"`
	// Dir is where the last list was written, prefilling the next save.
	Dir string `json:"dir,omitempty"`
}
