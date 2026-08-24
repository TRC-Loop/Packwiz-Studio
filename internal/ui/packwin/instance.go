package packwin

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	fynetheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/PalisadeMC/Packwiz-Studio/internal/instance"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/widgets"
)

// instanceActivity compares the pack against a game instance and copies
// files either way.
//
// The pack folder is not a game folder: a mod writes its config on first
// run, and the edit that matters was made in game. So this is the way
// back, file by file, rather than a folder copied over blind.
type instanceActivity struct {
	deps *activityDeps

	path    *widget.Entry
	folders map[string]*widget.Check
	list    *fyne.Container
	side    fyne.CanvasObject
	main    *fyne.Container

	entries []instance.Entry
	status  *widget.Label
	// compared records that a comparison has been asked for at least once,
	// which is what makes a reload recompare rather than sit idle.
	compared bool
}

func newInstanceActivity(deps *activityDeps) *instanceActivity {
	a := &instanceActivity{
		deps:    deps,
		path:    widget.NewEntry(),
		folders: map[string]*widget.Check{},
		list:    container.NewVBox(),
		main:    container.NewStack(),
		status:  widget.NewLabel(""),
	}

	prefs := deps.prefs().Instance
	a.path.SetPlaceHolder("Path to a .minecraft or instance folder")
	a.path.SetText(prefs.Dir)

	for _, folder := range instance.Folders() {
		check := widget.NewCheck(folder, nil)
		check.SetChecked(wanted(prefs, folder))
		a.folders[folder] = check
	}

	a.status.Wrapping = fyne.TextWrapWord
	a.side = container.NewVScroll(
		widgets.ClampWidth(tokens.SidePanelWidth, a.list))

	a.render()
	return a
}

func (a *instanceActivity) ID() string              { return ActivityInstance }
func (a *instanceActivity) Title() string           { return "Instance" }
func (a *instanceActivity) Icon() fyne.Resource     { return fynetheme.ComputerIcon() }
func (a *instanceActivity) Side() fyne.CanvasObject { return a.side }
func (a *instanceActivity) Main() fyne.CanvasObject { return a.main }

// Reload recompares, but only once a comparison has been asked for.
//
// Comparing reads every file in the chosen folders. Doing that after
// every command in the window, for a pack whose instance was never set
// up, would be work nobody asked for.
func (a *instanceActivity) Reload() {
	if a.compared && instance.Exists(a.path.Text) {
		a.compare()
		return
	}
	a.render()
}

// render redraws the settings pane and the list of differences.
func (a *instanceActivity) render() {
	a.renderList()

	browse := widget.NewButtonWithIcon("", fynetheme.FolderOpenIcon(), a.browse)

	checks := container.NewVBox()
	for _, folder := range instance.Folders() {
		checks.Add(a.folders[folder])
	}

	body := container.NewVBox(
		widgets.SubHeading("Game instance"),
		widgets.Note("Point this at the instance you test the pack in. "+
			"Comparing looks only inside the folders ticked below, so saves, "+
			"logs and caches are never touched."),
		container.NewBorder(nil, nil, nil, browse, a.path),
		widgets.VSpace(tokens.SpaceMD),

		widgets.Muted("Folders"),
		checks,
		widgets.VSpace(tokens.SpaceMD),

		a.actionRow(),
		widgets.VSpace(tokens.SpaceMD),
		a.status,
	)

	a.main.Objects = []fyne.CanvasObject{
		container.NewVScroll(widgets.Inset(tokens.SpaceXL, tokens.SpaceLG, body)),
	}
	a.main.Refresh()
}

// actionRow is compare, import and copy everything back.
func (a *instanceActivity) actionRow() fyne.CanvasObject {
	compare := widget.NewButtonWithIcon("Compare", fynetheme.ViewRefreshIcon(), a.compare)
	compare.Importance = widget.HighImportance

	imports := widget.NewButtonWithIcon("Import from instance",
		fynetheme.MoveDownIcon(), a.confirmImport)

	toPack := widget.NewButtonWithIcon("Copy all changes to the pack",
		fynetheme.ContentCopyIcon(), a.copyAllToPack)
	toPack.Importance = widget.LowImportance
	if len(a.entries) == 0 {
		toPack.Disable()
	}

	return container.NewVBox(
		container.NewHBox(compare, imports),
		container.NewHBox(toPack),
	)
}

// browse picks the instance folder.
func (a *instanceActivity) browse() {
	open := dialog.NewFolderOpen(func(list fyne.ListableURI, err error) {
		if err != nil || list == nil {
			return
		}
		a.path.SetText(list.Path())
		a.remember()
		a.compare()
	}, a.deps.win)

	if a.path.Text != "" {
		if lister, err := storage.ListerForURI(storage.NewFileURI(a.path.Text)); err == nil {
			open.SetLocation(lister)
		}
	}
	open.Show()
}
