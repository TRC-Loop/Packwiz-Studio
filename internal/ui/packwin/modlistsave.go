package packwin

import (
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"

	"github.com/PalisadeMC/Packwiz-Studio/internal/config"
	"github.com/PalisadeMC/Packwiz-Studio/internal/modlist"
	"github.com/PalisadeMC/Packwiz-Studio/internal/pack"
)

// render is the document the current choices produce.
func (f *modListForm) render() string {
	return modlist.Render(f.meta(), f.entries(), f.spec())
}

// meta describes the pack for the templates.
func (f *modListForm) meta() modlist.Meta {
	p := f.win.pack

	return modlist.Meta{
		Name:      p.Name,
		Version:   p.Version,
		MCVersion: p.MCVersion,
		Loader:    p.Loader,
		Date:      time.Now().Format(time.DateOnly),
	}
}

// entries reads the pack's mods. A pack whose index cannot be read
// renders an empty list rather than an error: the dialog is a preview of
// a document, and the index problem is reported wherever it is acted on.
func (f *modListForm) entries() []modlist.Entry {
	mods, err := f.win.pack.Mods()
	if err != nil {
		return nil
	}

	out := make([]modlist.Entry, 0, len(mods))
	for _, m := range mods {
		out = append(out, modlist.Entry{
			Name:      m.Name,
			Slug:      m.Slug(),
			Filename:  m.Filename,
			Side:      string(m.SideFlag),
			URL:       modListURL(m),
			ProjectID: m.ModrinthID,
			VersionID: m.VersionID,
			Pinned:    m.Pinned,
			Optional:  m.Optional,
		})
	}
	return out
}

// modListURL is where a reader should be sent for a mod: its Modrinth
// page when it has one, and the file it came from otherwise.
func modListURL(m pack.Mod) string {
	if m.ModrinthID != "" {
		return modrinthURL(m.ModrinthID)
	}
	return m.DownloadURL
}

// copy puts the document on the clipboard, which is what a forum post or
// a wiki edit wants.
func (f *modListForm) copy() {
	fyne.CurrentApp().Clipboard().SetContent(f.render())
	f.win.deps.notice("mod list copied to the clipboard")
}

// save asks where the document should go and writes it there.
func (f *modListForm) save() {
	choice := modlist.ByLabel(f.format.Selected)
	content := f.render()

	f.remember()

	d := dialog.NewFileSave(func(w fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, f.win.win)
			return
		}
		if w == nil {
			return
		}
		defer w.Close()

		if _, err := w.Write([]byte(content)); err != nil {
			dialog.ShowError(err, f.win.win)
			return
		}
		f.win.deps.setPrefs(func(p *config.Prefs) {
			p.ModList.Dir = filepath.Dir(w.URI().Path())
		})
		f.win.deps.notice("mod list written to " + w.URI().Path())
	}, f.win.win)

	d.SetFileName(modListName(f.win.pack.Name) + choice.Ext)
	if dir := f.saveDir(); dir != "" {
		if lister, err := storage.ListerForURI(storage.NewFileURI(dir)); err == nil {
			d.SetLocation(lister)
		}
	}
	d.Show()
}

// saveDir is where the picker opens: where the last list went, or the
// pack folder.
func (f *modListForm) saveDir() string {
	if dir := f.win.deps.prefs().ModList.Dir; dir != "" {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}
	return f.win.pack.Dir
}

// remember stores the format and the templates for next time.
func (f *modListForm) remember() {
	spec := f.spec()

	f.win.deps.setPrefs(func(p *config.Prefs) {
		p.ModList.Format = string(spec.Format)
		p.ModList.Header = spec.Header
		p.ModList.Line = spec.Line
		p.ModList.Footer = spec.Footer
	})
}

// modListName is the default file name, which reads as what it is rather
// than as the pack itself.
func modListName(packName string) string {
	name := slug(packName)
	if name == "" {
		name = "modpack"
	}
	return name + "-mods"
}
