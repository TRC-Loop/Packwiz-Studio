package packwin

import (
	"context"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	fynetheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/TRC-Loop/Packwiz-Studio/internal/cmdrun"
	"github.com/TRC-Loop/Packwiz-Studio/internal/config"
	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/tokens"
	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/widgets"
)

// Export asks where the .mrpack should go and writes it there.
//
// The dialog is always shown, even when the pack has a default output
// folder. The default only prefills it: exporting is the step that
// produces a file other people install, so it is never done without the
// destination being confirmed.
func (w *Window) Export() {
	prefs := w.deps.prefs()

	dir := prefs.ExportDir
	if dir == "" {
		dir = w.pack.Dir
	}

	folder := widget.NewEntry()
	folder.SetText(dir)

	name := widget.NewEntry()
	name.SetText(defaultExportName(w.pack.Name, w.pack.Version))

	remember := widget.NewCheck("Use this folder as the default for this pack", nil)
	remember.SetChecked(prefs.ExportDir == "" || prefs.ExportDir == dir)

	browse := widget.NewButtonWithIcon("", fynetheme.FolderOpenIcon(), func() {
		open := dialog.NewFolderOpen(func(list fyne.ListableURI, err error) {
			if err != nil || list == nil {
				return
			}
			folder.SetText(list.Path())
		}, w.win)

		if lister, err := storage.ListerForURI(storage.NewFileURI(folder.Text)); err == nil {
			open.SetLocation(lister)
		}
		open.Show()
	})

	body := container.NewVBox(
		widgets.Muted("Folder"),
		container.NewBorder(nil, nil, nil, browse, folder),
		widgets.VSpace(tokens.SpaceSM),
		widgets.Muted("File name"),
		name,
		widgets.VSpace(tokens.SpaceSM),
		remember,
	)

	d := dialog.NewCustomConfirm("Export Modrinth pack", "Export", "Cancel",
		widgets.Scrollable(tokens.FormWidth, tokens.ExportHeight,
			widgets.Inset(tokens.SpaceMD, tokens.SpaceMD, body)),
		func(export bool) {
			if !export {
				return
			}
			w.runExport(folder.Text, name.Text, remember.Checked)
		}, w.win)

	d.Resize(fyne.NewSize(tokens.FormWidth, tokens.ExportHeight))
	d.Show()
}

// runExport writes the pack and records where it went.
func (w *Window) runExport(dir, name string, remember bool) {
	if strings.TrimSpace(dir) == "" {
		dialog.ShowError(errNoExportFolder, w.win)
		return
	}

	output := filepath.Join(dir, ensureMrpack(name))

	if remember {
		w.deps.setPrefs(func(p *config.Prefs) { p.ExportDir = dir })
	}

	w.deps.run("export", func(ctx context.Context) error {
		if err := exec(func() (cmdrun.Result, error) {
			return w.deps.client().ExportModrinth(ctx, output)
		}); err != nil {
			return err
		}

		// The path is remembered so a release can attach the file without
		// exporting again.
		w.deps.setPrefs(func(p *config.Prefs) { p.LastExport = output })
		return nil
	})
}

// defaultExportName builds a file name from the pack's own metadata,
// which is what a release asset wants to be called.
func defaultExportName(packName, version string) string {
	name := slug(packName)
	if name == "" {
		name = "modpack"
	}
	if v := slug(version); v != "" {
		name += "-" + v
	}
	return name + mrpackExt
}

// mrpackExt is the Modrinth pack extension.
const mrpackExt = ".mrpack"

// ensureMrpack adds the extension when the user removed it.
func ensureMrpack(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "modpack" + mrpackExt
	}
	if strings.EqualFold(filepath.Ext(name), mrpackExt) {
		return name
	}
	return name + mrpackExt
}

// slug makes a string safe for a file name, keeping it readable rather
// than escaping everything.
func slug(s string) string {
	var b strings.Builder

	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '.':
			b.WriteRune(r)
		case r == ' ' || r == '_' || r == '-':
			b.WriteRune('-')
		}
	}
	return strings.Trim(b.String(), "-")
}
