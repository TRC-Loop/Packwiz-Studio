package packwin

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	fynetheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/PalisadeMC/Packwiz-Studio/internal/pack"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/widgets"
)

// modsActivity lists the pack's installed mods beside a detail pane for
// the selected one.
type modsActivity struct {
	deps *activityDeps

	filter   *widget.Entry
	list     *fyne.Container
	side     fyne.CanvasObject
	main     *fyne.Container
	empty    *widget.Label
	mods     []pack.Mod
	selected string
	rows     map[string]*widgets.Clickable
}

func newModsActivity(deps *activityDeps) *modsActivity {
	a := &modsActivity{
		deps:   deps,
		filter: widget.NewEntry(),
		list:   container.NewVBox(),
		main:   container.NewStack(),
		empty:  widget.NewLabel(""),
		rows:   map[string]*widgets.Clickable{},
	}

	a.filter.SetPlaceHolder("Filter mods")
	a.filter.OnChanged = func(string) { a.renderList() }

	a.empty.Hide()

	a.side = container.NewBorder(
		widgets.Inset(tokens.SpaceSM, tokens.SpaceXS, a.filter),
		nil, nil, nil,
		container.NewVScroll(container.NewVBox(a.list, a.empty)),
	)

	a.Reload()
	return a
}

func (a *modsActivity) ID() string              { return ActivityMods }
func (a *modsActivity) Title() string           { return "Mods" }
func (a *modsActivity) Icon() fyne.Resource     { return fynetheme.ListIcon() }
func (a *modsActivity) Side() fyne.CanvasObject { return a.side }
func (a *modsActivity) Main() fyne.CanvasObject { return a.main }

// Reload re-reads the pack's mod list from disk and redraws both panes.
func (a *modsActivity) Reload() {
	mods, err := a.deps.pack.Mods()
	if err != nil {
		a.mods = nil
		a.renderList()
		a.showMessage("The pack index could not be read: " + err.Error())
		return
	}

	a.mods = mods
	a.renderList()

	// Keep the selection across a reload where possible, so changing a
	// side flag does not throw the user back to an empty detail pane.
	if a.find(a.selected) == nil {
		a.selected = ""
	}
	if a.selected == "" {
		a.showMessage(a.summary())
		return
	}
	a.showDetail(*a.find(a.selected))
}

// renderList rebuilds the side panel rows for the current filter.
func (a *modsActivity) renderList() {
	needle := strings.ToLower(strings.TrimSpace(a.filter.Text))

	a.list.Objects = nil
	a.rows = map[string]*widgets.Clickable{}

	shown := 0
	for _, m := range a.mods {
		if needle != "" && !strings.Contains(strings.ToLower(m.Name), needle) {
			continue
		}
		row := a.row(m)
		a.rows[m.Path] = row
		a.list.Objects = append(a.list.Objects, row)
		shown++
	}
	a.list.Refresh()

	switch {
	case len(a.mods) == 0:
		a.empty.SetText("This pack has no mods yet.")
		a.empty.Show()
	case shown == 0:
		a.empty.SetText("No mod matches that filter.")
		a.empty.Show()
	default:
		a.empty.Hide()
	}
}

// row is one mod in the list.
func (a *modsActivity) row(m pack.Mod) *widgets.Clickable {
	path := m.Path

	label := widgets.Body(m.Name)
	label.Truncation = fyne.TextTruncateEllipsis

	line := container.NewBorder(nil, nil, nil, a.rowBadge(m), label)

	row := widgets.NewClickable(line, func() { a.selectMod(path) })
	row.SetSelected(path == a.selected)
	return row
}

// rowBadge marks a mod that is not a plain both-sides install, so the
// list carries that state without a second line per row.
func (a *modsActivity) rowBadge(m pack.Mod) fyne.CanvasObject {
	switch {
	case m.LoadErr != nil:
		return widgets.Status("unreadable", widgets.StateError)
	case m.SideFlag == pack.SideClient:
		return widgets.Dim("client")
	case m.SideFlag == pack.SideServer:
		return widgets.Dim("server")
	case m.Pinned:
		return widgets.Dim("pinned")
	default:
		return nil
	}
}

// selectMod highlights a row and shows its detail.
func (a *modsActivity) selectMod(path string) {
	a.selected = path
	for p, row := range a.rows {
		row.SetSelected(p == path)
	}

	if m := a.find(path); m != nil {
		a.showDetail(*m)
	}

	// The Mods menu acts on the selection, so it has to be rebuilt now
	// that there is one.
	a.deps.menuChanged()
}

// find looks up a mod by its metafile path.
func (a *modsActivity) find(path string) *pack.Mod {
	for i := range a.mods {
		if a.mods[i].Path == path {
			return &a.mods[i]
		}
	}
	return nil
}

// summary is what the detail pane shows with nothing selected.
func (a *modsActivity) summary() string {
	switch len(a.mods) {
	case 0:
		return "Add mods from the browser to get started."
	case 1:
		return "1 mod installed. Select it to see its details."
	default:
		return itoa(len(a.mods)) + " mods installed. Select one to see its details."
	}
}

// showMessage puts a plain note in the detail pane.
func (a *modsActivity) showMessage(text string) {
	a.main.Objects = []fyne.CanvasObject{
		widgets.Inset(tokens.SpaceXL, tokens.SpaceLG, widgets.Muted(text)),
	}
	a.main.Refresh()
}
