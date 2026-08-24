package launcher

import (
	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"github.com/PalisadeMC/Packwiz-Studio/internal/mcmeta"
	"github.com/PalisadeMC/Packwiz-Studio/internal/packwiz"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/widgets"
)

// latestLabel is the loader version entry that leaves the choice to
// packwiz, which resolves it at init time.
const latestLabel = "Latest"

// loadingLabel fills a list that has not arrived yet.
const loadingLabel = "Loading"

// buildVersionControls creates the three dependent version controls.
func (f *newPackForm) buildVersionControls() {
	labels := make([]string, 0, len(packwiz.Loaders))
	for _, l := range packwiz.Loaders {
		labels = append(labels, l.Label())
	}

	// Callbacks are attached after every control exists, because setting a
	// selection fires them immediately.
	f.loader = widget.NewSelect(labels, nil)
	f.mcVersion = widget.NewSelect([]string{loadingLabel}, nil)
	f.loaderVersion = widget.NewSelect([]string{latestLabel}, nil)
	f.snapshots = widget.NewCheck("Include snapshots", nil)
	f.loaderManual = widget.NewEntry()

	f.loaderManual.SetPlaceHolder("leave empty for the latest")
	f.loaderManual.Hide()

	f.loader.SetSelectedIndex(0)
	f.mcVersion.SetSelected(loadingLabel)
	f.mcVersion.Disable()
	f.loaderVersion.SetSelected(latestLabel)
	f.loaderVersion.Disable()

	f.loader.OnChanged = func(string) { f.loadLoaderVersions() }
	f.mcVersion.OnChanged = func(string) { f.loadLoaderVersions() }
	f.snapshots.OnChanged = func(bool) { f.fillGameVersions() }
}

// loaderName is the packwiz loader currently selected.
func (f *newPackForm) loaderName() string {
	if i := f.loader.SelectedIndex(); i >= 0 && i < len(packwiz.Loaders) {
		return string(packwiz.Loaders[i])
	}
	return string(packwiz.LoaderFabric)
}

// loadGameVersions fetches the Minecraft version list.
func (f *newPackForm) loadGameVersions() {
	f.say("Loading Minecraft versions", widgets.StateNeutral)

	go func() {
		versions, err := f.win.sess.Meta.GameVersions(context.Background())

		fyne.Do(func() {
			if err != nil {
				f.gameVersionsFailed(err)
				return
			}
			f.games = versions
			f.fillGameVersions()
		})
	}()
}

// gameVersionsFailed falls back to a typed Minecraft version, so an
// offline machine can still create a pack.
func (f *newPackForm) gameVersionsFailed(err error) {
	f.mcVersion.Options = []string{}
	f.mcVersion.ClearSelected()
	f.mcVersion.Disable()
	f.loaderManual.Show()
	f.say("Could not load version lists: "+err.Error(), widgets.StateWarning)
}

// fillGameVersions populates the Minecraft list for the snapshot setting,
// keeping the current selection where it still exists.
func (f *newPackForm) fillGameVersions() {
	shown := f.games
	if !f.snapshots.Checked {
		shown = mcmeta.Stable(f.games)
	}
	if len(shown) == 0 {
		return
	}

	previous := f.mcVersion.Selected

	f.mcVersion.Options = mcmeta.IDs(shown)
	f.mcVersion.Enable()

	// The list is newest first, so the newest version is the default.
	next := shown[0].ID
	if previous != "" && previous != loadingLabel && contains(f.mcVersion.Options, previous) {
		next = previous
	}

	f.mcVersion.SetSelected(next)
	f.mcVersion.Refresh()

	// SetSelected only fires the callback on a change, so a reselection of
	// the same version needs the loader list refreshed explicitly.
	if next == previous {
		f.loadLoaderVersions()
	}
}

// loadLoaderVersions fetches builds for the chosen loader and Minecraft
// version.
func (f *newPackForm) loadLoaderVersions() {
	mc := f.mcVersion.Selected
	if mc == "" || mc == loadingLabel {
		return
	}

	loader := f.loaderName()

	f.loaderVersion.Options = []string{loadingLabel}
	f.loaderVersion.SetSelected(loadingLabel)
	f.loaderVersion.Disable()
	f.loaderManual.Hide()

	go func() {
		versions, err := f.win.sess.Meta.LoaderVersions(context.Background(), loader, mc)

		fyne.Do(func() {
			// A late reply for a loader or version the user has since
			// changed is dropped, so the list always matches the form.
			if loader != f.loaderName() || mc != f.mcVersion.Selected {
				return
			}
			f.fillLoaderVersions(loader, mc, versions, err)
		})
	}()
}

// fillLoaderVersions populates the loader build list, or falls back to a
// typed version when there is nothing to offer.
func (f *newPackForm) fillLoaderVersions(loader, mc string, versions []mcmeta.Version, err error) {
	if err != nil || len(versions) == 0 {
		f.loaderVersion.Options = []string{latestLabel}
		f.loaderVersion.SetSelected(latestLabel)
		f.loaderVersion.Disable()
		f.loaderManual.Show()

		switch {
		case err != nil:
			f.say(err.Error(), widgets.StateWarning)
		default:
			f.say("No "+loader+" build is published for Minecraft "+mc+
				". Choose another version, or type a build below.", widgets.StateWarning)
		}
		return
	}

	// Latest leads the list: it is the right answer most of the time, and
	// it lets packwiz resolve the build itself.
	f.loaderVersion.Options = append([]string{latestLabel}, mcmeta.IDs(versions)...)
	f.loaderVersion.SetSelected(latestLabel)
	f.loaderVersion.Enable()
	f.loaderManual.Hide()

	f.say(plural(len(versions), loader+" build")+" for Minecraft "+mc,
		widgets.StateNeutral)
}

// say reports progress or a problem under the form.
func (f *newPackForm) say(text string, state widgets.State) {
	f.status.SetText(text)
	f.status.Importance = importanceFor(state)
	f.status.Show()
	f.status.Refresh()
}

// contains reports whether a list holds a value.
func contains(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}
