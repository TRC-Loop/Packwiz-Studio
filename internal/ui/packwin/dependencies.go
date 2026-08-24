package packwin

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"github.com/PalisadeMC/Packwiz-Studio/internal/config"
)

// dependencyCheck is the control deciding whether adding a mod also pulls
// in the libraries it needs.
//
// packwiz asks about dependencies when it adds a mod, and the app answers
// for it. On means yes, which is the useful default since a mod without
// its libraries will not load. Off is for a pack that manages its
// libraries deliberately.
func (d *activityDeps) dependencyCheck() fyne.CanvasObject {
	check := widget.NewCheck("With dependencies", nil)
	check.SetChecked(d.installDependencies())

	check.OnChanged = func(on bool) {
		if err := d.sess.Cfg.Update(func(c *config.Config) {
			c.InstallDependencies = on
		}); err != nil {
			d.notice("could not save the dependency setting: " + err.Error())
		}
	}
	return check
}

// installDependencies reports the current setting.
func (d *activityDeps) installDependencies() bool {
	return d.sess.Cfg.Get().InstallDependencies
}
