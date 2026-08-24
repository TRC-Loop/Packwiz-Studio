package menu

import (
	"fyne.io/fyne/v2"
)

func fileMenu(a Actions) *fyne.Menu {
	items := compact(
		withShortcut(item("New pack", a.NewPack), fyne.KeyN, modPrimary),
		withShortcut(item("Open pack", a.OpenPack), fyne.KeyO, modPrimary),
		recentsSubmenu(a.Recents),
		separator(),
		withShortcut(item("Close window", a.CloseWindow), fyne.KeyW, modPrimary),
		separator(),
		withShortcut(item("Settings", a.Settings), fyne.KeyComma, modPrimary),
	)
	return fyne.NewMenu("File", items...)
}

// recentsSubmenu lists known packs. It is omitted when there are none, so
// the menu never shows an empty submenu.
func recentsSubmenu(recents []Recent) *fyne.MenuItem {
	if len(recents) == 0 {
		return nil
	}

	children := make([]*fyne.MenuItem, 0, len(recents))
	for _, r := range recents {
		children = append(children, fyne.NewMenuItem(r.Label, r.Open))
	}

	parent := fyne.NewMenuItem("Open recent", nil)
	parent.ChildMenu = fyne.NewMenu("", children...)
	return parent
}

func packMenu(a Actions) *fyne.Menu {
	items := compact(
		withShortcut(item("Refresh", a.Refresh), fyne.KeyR, modPrimary),
		withShortcut(item("Export", a.Export), fyne.KeyE, modPrimary),
		withShortcut(item("Export mod list", a.ExportModList), fyne.KeyE, modShift),
		item("Check for updates", a.CheckUpdates),
		separator(),
		item("Ignore rules", a.IgnoreRules),
		withShortcut(item("Properties", a.Properties), fyne.KeyI, modPrimary),
		item("Reveal in file manager", a.RevealInFiles),
	)
	return fyne.NewMenu("Pack", items...)
}

func modsMenu(a Actions) *fyne.Menu {
	items := compact(
		withShortcut(item("Browse Modrinth", a.AddMod), fyne.KeyF, modPrimary),
		item("Add from a URL", a.AddFromURL),
		item("Add from GitHub", a.AddFromGit),
		separator(),
		item("Remove selected mod", a.RemoveMod),
		separator(),
		sideSubmenu(a),
	)
	return fyne.NewMenu("Mods", items...)
}

// sideSubmenu offers the client, server and both side flags for the
// selected mod.
func sideSubmenu(a Actions) *fyne.MenuItem {
	children := compact(
		item("Client only", a.SideClient),
		item("Server only", a.SideServer),
		item("Client and server", a.SideBoth),
	)
	if len(children) == 0 {
		return nil
	}

	parent := fyne.NewMenuItem("Side", nil)
	parent.ChildMenu = fyne.NewMenu("", children...)
	return parent
}

func gitMenu(a Actions) *fyne.Menu {
	items := compact(
		item("Initialise repository", a.GitInit),
		separator(),
		item("Stage all changes", a.GitStage),
		withShortcut(item("Commit", a.GitCommit), fyne.KeyK, modPrimary),
		separator(),
		withShortcut(item("Push", a.GitPush), fyne.KeyP, modShift),
		withShortcut(item("Pull", a.GitPull), fyne.KeyP, modPrimary),
		separator(),
		item("Open remote", a.OpenRemote),
	)
	return fyne.NewMenu("Git", items...)
}

func releaseMenu(a Actions) *fyne.Menu {
	items := compact(
		item("New release", a.NewRelease),
		item("Generate changelog", a.GenerateChangelog),
		separator(),
		item("Clear stored API token", a.ForgetToken),
	)
	return fyne.NewMenu("Release", items...)
}

func viewMenu(a Actions) *fyne.Menu {
	items := compact(
		withShortcut(item("Toggle side panel", a.ToggleSidePanel), fyne.KeyB, modPrimary),
		withShortcut(item("Toggle output", a.ToggleLog), fyne.KeyJ, modPrimary),
		separator(),
		item("Grid view", a.GridView),
		item("List view", a.ListView),
		separator(),
		activitySubmenu(a.Activities),
	)
	return fyne.NewMenu("View", items...)
}

// activitySubmenu mirrors the icon rail, giving each activity a name and
// a number shortcut. The rail is icon only, so this is where its sections
// are discoverable by name.
func activitySubmenu(activities []ActivityItem) *fyne.MenuItem {
	if len(activities) == 0 {
		return nil
	}

	digits := []fyne.KeyName{
		fyne.Key1, fyne.Key2, fyne.Key3, fyne.Key4,
		fyne.Key5, fyne.Key6, fyne.Key7, fyne.Key8, fyne.Key9,
	}

	children := make([]*fyne.MenuItem, 0, len(activities))
	for i, act := range activities {
		it := fyne.NewMenuItem(act.Label, act.Select)
		if i < len(digits) {
			it = withShortcut(it, digits[i], modPrimary)
		}
		children = append(children, it)
	}

	parent := fyne.NewMenuItem("Go to", nil)
	parent.ChildMenu = fyne.NewMenu("", children...)
	return parent
}

func helpMenu(a Actions) *fyne.Menu {
	items := compact(
		item("About Packwiz Studio", a.About),
		item("packwiz version", a.PackwizVersion),
	)
	return fyne.NewMenu("Help", items...)
}
