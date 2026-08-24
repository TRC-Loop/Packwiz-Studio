package packwin

import (
	"fyne.io/fyne/v2/dialog"
)

// releasesActivityOrNil finds the releases activity, which is absent when
// the git integration is off.
func (w *Window) releasesActivityOrNil() *releasesActivity {
	for _, a := range w.items {
		if r, ok := a.(*releasesActivity); ok {
			return r
		}
	}
	return nil
}

// NewRelease opens the release form.
func (w *Window) NewRelease() {
	if w.releasesActivityOrNil() == nil {
		return
	}
	w.Select(ActivityReleases)
}

// GenerateChangelog opens the release form and fills in its changelog,
// which is the same thing the form's own button does.
func (w *Window) GenerateChangelog() {
	r := w.releasesActivityOrNil()
	if r == nil {
		return
	}

	w.Select(ActivityReleases)

	// The form is rebuilt each time the activity renders, so the button
	// on the current form is the one to trigger. Rendering happens after
	// the repository probe finishes, so a form may not exist yet.
	if r.form == nil {
		dialog.ShowInformation("Changelog",
			"Open the Releases section and use Generate from git once the "+
				"repository has been read.", w.win)
		return
	}
	r.generateNotes(r.form)
}

// ForgetReleaseToken clears the stored API token for this pack's host.
func (w *Window) ForgetReleaseToken() {
	r := w.releasesActivityOrNil()
	if r == nil {
		return
	}

	if err := r.ForgetToken(); err != nil {
		dialog.ShowError(err, w.win)
		return
	}
	dialog.ShowInformation("Token cleared",
		"The stored token for "+r.host.Remote.Host+" has been removed.", w.win)
}

// HasRemoteHost reports whether a release could be published at all, so
// the menu can leave the release items out when it could not.
func (w *Window) HasRemoteHost() bool {
	r := w.releasesActivityOrNil()
	return r != nil && r.hostOK
}
