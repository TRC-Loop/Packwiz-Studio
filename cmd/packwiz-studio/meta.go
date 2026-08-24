package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

const (
	appID   = "com.stellarfoundry.packwizstudio"
	appName = "Packwiz Studio"
	version = "0.1.0"
)

// declareMetadata registers app metadata in code rather than relying on
// FyneApp.toml, which Fyne only reads for development builds.
//
// The fyneDo migration opts into Fyne's threading model: all widget
// mutation happens on the main goroutine via fyne.Do. Declaring it here
// rather than with the migrated_fynedo build tag keeps Fyne's runtime
// thread checks enabled, so a stray off-thread update is reported.
func declareMetadata() {
	app.SetMetadata(fyne.AppMetadata{
		ID:         appID,
		Name:       appName,
		Version:    version,
		Build:      1,
		Migrations: map[string]bool{"fyneDo": true},
	})
}
