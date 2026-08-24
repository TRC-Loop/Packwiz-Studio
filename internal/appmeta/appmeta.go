// Package appmeta reads the app's own name and version at runtime.
//
// Both are declared once with app.SetMetadata at startup, so nothing has
// to be threaded through constructors to put them in a window title.
package appmeta

import "fyne.io/fyne/v2"

// fallbackName is used when metadata was never declared, which happens in
// a build that skipped the declaration.
const fallbackName = "Packwiz Studio"

// Name is the app's display name.
func Name() string {
	if app := fyne.CurrentApp(); app != nil {
		if name := app.Metadata().Name; name != "" {
			return name
		}
	}
	return fallbackName
}

// Version is the app's version, empty when it is not known.
func Version() string {
	if app := fyne.CurrentApp(); app != nil {
		return app.Metadata().Version
	}
	return ""
}

// Label is the app's name with its version, for a window title or an
// about box.
func Label() string {
	if v := Version(); v != "" {
		return Name() + " v" + v
	}
	return Name()
}

// Title builds a window title. A subject in front names what the window
// holds, which is what a taskbar or a window switcher shows first.
func Title(subject string) string {
	if subject == "" {
		return Label()
	}
	return subject + " - " + Label()
}
