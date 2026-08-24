package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"

	"github.com/PalisadeMC/Packwiz-Studio/internal/config"
	"github.com/PalisadeMC/Packwiz-Studio/internal/pack"
	"github.com/PalisadeMC/Packwiz-Studio/internal/sysopen"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/menu"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/packwin"
)

// launcherMenu is the menubar while the window shows the launcher. Pack
// actions are absent because there is no pack open to apply them to.
func (s *shell) launcherMenu() *fyne.MainMenu {
	return menu.Build(menu.Actions{
		NewPack:  s.launcher.ShowNewPack,
		OpenPack: s.launcher.ShowOpenPack,
		Recents:  s.recentItems(),
		Settings: s.launcher.ShowSettings,

		About:          s.showAbout,
		PackwizVersion: s.showPackwizVersion,
	})
}

// packMenu is the menubar while a pack is open. Git and release items are
// left out when the git integration is off, so the menus never offer an
// action the app will not perform.
func (s *shell) packMenu(w *packwin.Window) *fyne.MainMenu {
	a := menu.Actions{
		NewPack:     s.newLauncherWindow,
		OpenPack:    s.launcher.ShowOpenPack,
		Recents:     s.recentItems(),
		CloseWindow: w.Close,
		Settings:    s.launcher.ShowSettings,

		Refresh:       w.RefreshIndex,
		Export:        w.Export,
		CheckUpdates:  w.CheckUpdates,
		Properties:    w.EditProperties,
		RevealInFiles: func() { s.reveal(w.Pack().Dir) },

		AddMod:     w.FocusBrowse,
		AddFromURL: w.AddFromURL,
		AddFromGit: w.AddFromGitHub,

		ToggleSidePanel: w.ToggleSidePanel,
		ToggleLog:       w.ToggleLog,
		GridView:        func() { w.SetViewMode(config.ViewGrid) },
		ListView:        func() { w.SetViewMode(config.ViewList) },
		Activities:      activityItems(w),

		About:          s.showAbout,
		PackwizVersion: s.showPackwizVersion,
	}

	// Git items are absent entirely when the integration is off, so the
	// Git menu disappears rather than sitting there greyed out.
	if w.GitEnabled() {
		a.GitInit = w.GitInit
		a.GitStage = w.GitStageAll
		a.GitCommit = w.GitCommit
		a.GitPush = w.GitPush
		a.GitPull = w.GitPull
		a.OpenRemote = w.OpenRemote
	}

	// A release needs somewhere to publish to, so the Release menu only
	// appears once a remote host has been resolved.
	if w.GitEnabled() && w.HasRemoteHost() {
		a.NewRelease = w.NewRelease
		a.GenerateChangelog = w.GenerateChangelog
		a.ForgetToken = w.ForgetReleaseToken
	}

	// The mod actions apply to the list's selection, so they only appear
	// once something is selected.
	if _, ok := w.SelectedMod(); ok {
		a.RemoveMod = w.RemoveSelectedMod
		a.SideClient = func() { w.SetSelectedSide(pack.SideClient) }
		a.SideServer = func() { w.SetSelectedSide(pack.SideServer) }
		a.SideBoth = func() { w.SetSelectedSide(pack.SideBoth) }
	}

	return menu.Build(a)
}

// recentItems lists known packs for the File menu, opening each one
// beside the current window rather than replacing what is open.
func (s *shell) recentItems() []menu.Recent {
	packs := s.sess.Cfg.Packs()

	items := make([]menu.Recent, 0, len(packs))
	for _, p := range packs {
		dir := p.Path
		items = append(items, menu.Recent{
			Label: p.Name,
			Open:  func() { s.openPackInNewWindow(dir) },
		})
	}
	return items
}

// activityItems mirrors the icon rail into the View menu, which is where
// the rail's icon-only sections get their names and shortcuts.
func activityItems(w *packwin.Window) []menu.ActivityItem {
	acts := w.Activities()

	items := make([]menu.ActivityItem, 0, len(acts))
	for _, a := range acts {
		id := a.ID()
		items = append(items, menu.ActivityItem{
			Label:  a.Title(),
			Select: func() { w.Select(id) },
		})
	}
	return items
}

// reveal opens a folder in the desktop file manager.
func (s *shell) reveal(dir string) {
	if err := sysopen.Reveal(dir); err != nil {
		dialog.ShowError(err, s.win)
	}
}

// showAbout reports what this build is.
func (s *shell) showAbout() {
	dialog.ShowInformation("About Packwiz Studio",
		appName+" "+version+"\n\nA GUI for authoring and releasing Minecraft\nmodpacks with packwiz.",
		s.win)
}

// showPackwizVersion reports the resolved binary, which is the quickest
// answer to why a command behaved unexpectedly.
func (s *shell) showPackwizVersion() {
	loc := s.sess.Packwiz()
	if loc.Path == "" {
		dialog.ShowInformation("packwiz", "No packwiz binary is configured.", s.win)
		return
	}

	body := loc.Path
	if loc.Version != "" {
		body = "packwiz " + loc.Version + "\n\n" + loc.Path
	}
	if loc.FromPATH {
		body += "\n\nFound on PATH."
	}
	dialog.ShowInformation("packwiz", body, s.win)
}
