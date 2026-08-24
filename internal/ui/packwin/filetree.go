package packwin

import (
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	fynetheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/PalisadeMC/Packwiz-Studio/internal/pack"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/widgets"
)

// buildTree creates the folder tree.
//
// A row is an icon and a name, the icon chosen from the file's kind, so
// the shape of a pack is readable at a glance: which folders hold configs,
// which hold scripts, which entries are mods.
func (a *filesActivity) buildTree() {
	a.tree = widget.NewTree(
		func(id widget.TreeNodeID) []widget.TreeNodeID {
			return a.childIDs(id)
		},
		func(id widget.TreeNodeID) bool {
			return a.scan.IsDir(id)
		},
		func(branch bool) fyne.CanvasObject {
			icon := widget.NewIcon(fynetheme.FileIcon())

			label := widgets.Body("")
			label.Truncation = fyne.TextTruncateEllipsis

			return container.NewBorder(nil, nil, icon, nil, label)
		},
		func(id widget.TreeNodeID, branch bool, node fyne.CanvasObject) {
			a.renderRow(id, branch, node)
		},
	)

	a.tree.OnSelected = func(id widget.TreeNodeID) { a.selectFile(id) }
	a.tree.Root = ""
}

// childIDs lists a folder's contents for the tree.
func (a *filesActivity) childIDs(id widget.TreeNodeID) []widget.TreeNodeID {
	nodes := a.scan.Children(id)

	out := make([]widget.TreeNodeID, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, n.Path)
	}
	return out
}

// renderRow fills in one row of the tree.
func (a *filesActivity) renderRow(id widget.TreeNodeID, branch bool, row fyne.CanvasObject) {
	box, ok := row.(*fyne.Container)
	if !ok || len(box.Objects) < 2 {
		return
	}

	label, _ := box.Objects[0].(*widget.Label)
	icon, _ := box.Objects[1].(*widget.Icon)
	if label == nil {
		return
	}

	node, found := a.scan.Find(id)
	if !found {
		node = pack.Node{Path: id, Name: filepath.Base(id), Dir: branch}
	}

	label.SetText(node.Name)

	// A file the index does not carry is dimmed: it is either new, and a
	// refresh will pick it up, or ignored on purpose. Either way it is not
	// part of the pack yet, and that is worth seeing without asking.
	if node.Dir || node.Indexed {
		label.Importance = widget.MediumImportance
	} else {
		label.Importance = widget.LowImportance
	}
	label.Refresh()

	if icon != nil {
		icon.SetResource(fileIcon(node, a.tree.IsBranchOpen(id)))
	}
}

// fileIcon picks a glyph for a node from Fyne's built-in set.
func fileIcon(node pack.Node, open bool) fyne.Resource {
	if node.Dir {
		if open {
			return fynetheme.FolderOpenIcon()
		}
		return fynetheme.FolderIcon()
	}
	if node.Metafile {
		return fynetheme.DownloadIcon()
	}

	name := strings.ToLower(node.Name)
	switch strings.TrimPrefix(filepath.Ext(name), ".") {
	case "toml", "json", "json5", "jsonc", "yaml", "yml",
		"properties", "cfg", "conf", "ini", "snbt", "nbt":
		return fynetheme.SettingsIcon()
	case "js", "mjs", "cjs", "ts", "zs":
		return fynetheme.ComputerIcon()
	case "md", "txt", "log", "license":
		return fynetheme.FileTextIcon()
	case "png", "jpg", "jpeg", "webp", "gif", "svg":
		return fynetheme.FileImageIcon()
	case "ogg", "wav", "mp3":
		return fynetheme.FileAudioIcon()
	case "jar", "zip", "mrpack", "disabled":
		return fynetheme.FileApplicationIcon()
	default:
		return fynetheme.FileIcon()
	}
}
