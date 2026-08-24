package packwin

import (
	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	fynetheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/TRC-Loop/Packwiz-Studio/internal/config"
	"github.com/TRC-Loop/Packwiz-Studio/internal/modrinth"
	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/tokens"
	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/widgets"
)

// browseActivity searches Modrinth and adds results to the pack.
//
// It takes the whole window rather than using the side panel: cards need
// the width, and there is no list-plus-detail relationship to model here.
type browseActivity struct {
	deps   *activityDeps
	client *modrinth.Client

	search   *widget.Entry
	onlyPack *widget.Check
	viewMode *widget.Button
	results  *fyne.Container
	message  *widget.Label
	more     *widget.Button
	main     *fyne.Container

	hits     []modrinth.Hit
	total    int
	offset   int
	mode     config.ViewMode
	searchID int
	// installed maps a project id to the mod already in the pack, so a
	// card can say so instead of offering to add it twice.
	installed map[string]string
}

func newBrowseActivity(deps *activityDeps) *browseActivity {
	a := &browseActivity{
		deps:      deps,
		client:    modrinth.New(),
		search:    widget.NewEntry(),
		results:   container.NewVBox(),
		message:   widget.NewLabel(""),
		main:      container.NewStack(),
		mode:      deps.prefs().ViewMode,
		installed: map[string]string{},
	}

	a.search.SetPlaceHolder("Search Modrinth")
	a.search.OnSubmitted = func(string) { a.runSearch(0) }

	a.onlyPack = widget.NewCheck("Match this pack", func(bool) { a.runSearch(0) })
	a.onlyPack.SetChecked(true)

	a.viewMode = widget.NewButtonWithIcon("", a.modeIcon(), a.toggleMode)
	a.viewMode.Importance = widget.LowImportance

	a.more = widget.NewButton("Load more", func() { a.runSearch(a.offset + modrinth.DefaultLimit) })
	a.more.Hide()

	a.message.Hide()

	a.main.Objects = []fyne.CanvasObject{a.layout()}
	a.reloadInstalled()
	a.runSearch(0)

	return a
}

func (a *browseActivity) ID() string              { return ActivityBrowse }
func (a *browseActivity) Title() string           { return "Browse" }
func (a *browseActivity) Icon() fyne.Resource     { return fynetheme.SearchIcon() }
func (a *browseActivity) Side() fyne.CanvasObject { return nil }
func (a *browseActivity) Main() fyne.CanvasObject { return a.main }

// layout builds the search bar above a scrolling result area.
func (a *browseActivity) layout() fyne.CanvasObject {
	find := widget.NewButtonWithIcon("", fynetheme.SearchIcon(), func() { a.runSearch(0) })

	bar := container.NewBorder(nil, nil, nil,
		container.NewHBox(find, a.onlyPack, a.viewMode),
		a.search,
	)

	body := container.NewVBox(a.message, a.results, widgets.VSpace(tokens.SpaceMD), a.more)

	return container.NewBorder(
		widgets.Inset(tokens.SpaceLG, tokens.SpaceSM, bar),
		nil, nil, nil,
		container.NewVScroll(widgets.Inset(tokens.SpaceLG, tokens.SpaceSM, body)),
	)
}

// Reload re-reads which mods the pack already holds, so cards stay
// accurate after something is added or removed elsewhere.
func (a *browseActivity) Reload() {
	a.reloadInstalled()
	a.renderResults()
}

// reloadInstalled indexes the pack's mods by Modrinth project id.
func (a *browseActivity) reloadInstalled() {
	a.installed = map[string]string{}

	mods, err := a.deps.pack.Mods()
	if err != nil {
		return
	}
	for _, m := range mods {
		if m.ModrinthID != "" {
			a.installed[m.ModrinthID] = m.Name
		}
	}
}

// toggleMode switches between grid and list, remembering the choice for
// this pack.
func (a *browseActivity) toggleMode() {
	if a.mode == config.ViewGrid {
		a.mode = config.ViewList
	} else {
		a.mode = config.ViewGrid
	}

	a.viewMode.SetIcon(a.modeIcon())
	a.deps.setPrefs(func(p *config.Prefs) { p.ViewMode = a.mode })
	a.renderResults()
}

// modeIcon shows what the button will switch to.
func (a *browseActivity) modeIcon() fyne.Resource {
	if a.mode == config.ViewGrid {
		return fynetheme.ListIcon()
	}
	return fynetheme.GridIcon()
}

// runSearch queries Modrinth from the given offset. An offset of zero
// replaces the results; anything else appends a page.
//
// Each search takes a ticket, and a reply is dropped unless its ticket is
// still the current one. Typing quickly would otherwise let a slow early
// reply land after a fast later one.
func (a *browseActivity) runSearch(offset int) {
	a.searchID++
	ticket := a.searchID

	query := modrinth.Query{Text: a.search.Text, Offset: offset}
	if a.onlyPack.Checked {
		query.MCVersion = a.deps.pack.MCVersion
		query.Loader = a.deps.pack.Loader
	}

	a.setMessage("Searching Modrinth")

	go func() {
		res, err := a.client.Search(context.Background(), query)

		fyne.Do(func() {
			if ticket != a.searchID {
				return
			}
			if err != nil {
				a.setMessage(err.Error())
				return
			}

			if offset == 0 {
				a.hits = res.Hits
			} else {
				a.hits = append(a.hits, res.Hits...)
			}
			a.total = res.Total
			a.offset = res.Offset

			a.clearMessage()
			a.renderResults()
		})
	}()
}

// setMessage replaces the result area with a note.
func (a *browseActivity) setMessage(text string) {
	a.message.SetText(text)
	a.message.Show()
	a.more.Hide()
}

func (a *browseActivity) clearMessage() { a.message.Hide() }
