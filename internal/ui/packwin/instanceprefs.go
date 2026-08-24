package packwin

import (
	"strings"

	"github.com/PalisadeMC/Packwiz-Studio/internal/config"
	"github.com/PalisadeMC/Packwiz-Studio/internal/instance"
)

// chosen is the folders currently ticked.
func (a *instanceActivity) chosen() []string {
	var out []string
	for _, folder := range instance.Folders() {
		if a.folders[folder].Checked {
			out = append(out, folder)
		}
	}
	return out
}

// remember stores the instance path and folder choice for this pack.
func (a *instanceActivity) remember() {
	dir := a.path.Text
	folders := a.chosen()

	a.deps.setPrefs(func(p *config.Prefs) {
		p.Instance.Dir = dir
		p.Instance.Folders = folders
	})
}

// joinFolders lists folder names for a prompt.
func joinFolders(folders []string) string {
	if len(folders) == 1 {
		return folders[0]
	}
	return strings.Join(folders[:len(folders)-1], ", ") +
		" and " + folders[len(folders)-1]
}

// wanted reports whether a folder was ticked last time. A pack with no
// stored choice starts with the folders a pack nearly always carries.
func wanted(prefs config.InstancePrefs, folder string) bool {
	if len(prefs.Folders) == 0 {
		switch folder {
		case "config", "defaultconfigs", "kubejs", "scripts":
			return true
		}
		return false
	}
	for _, f := range prefs.Folders {
		if f == folder {
			return true
		}
	}
	return false
}
