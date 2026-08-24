package packwin

import (
	"context"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	fynetheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/PalisadeMC/Packwiz-Studio/internal/cmdrun"
	"github.com/PalisadeMC/Packwiz-Studio/internal/config"
	"github.com/PalisadeMC/Packwiz-Studio/internal/modrinth"
	"github.com/PalisadeMC/Packwiz-Studio/internal/sysopen"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/widgets"
)

// renderResults redraws the result area in the current view mode.
func (a *browseActivity) renderResults() {
	a.results.Objects = nil

	if len(a.hits) == 0 {
		a.results.Refresh()
		a.setMessage("No results. Try a different search, or turn off the pack filter.")
		return
	}

	if a.mode == config.ViewList {
		for _, h := range a.hits {
			a.results.Add(a.listRow(h))
		}
	} else {
		a.results.Add(a.grid())
	}

	a.results.Refresh()

	if len(a.hits) < a.total {
		a.more.SetText("Load more, showing " +
			strconv.Itoa(len(a.hits)) + " of " + strconv.Itoa(a.total))
		a.more.Show()
	} else {
		a.more.Hide()
	}
}

// grid lays the results out as thumbnail cards, which is the default view.
func (a *browseActivity) grid() fyne.CanvasObject {
	cards := make([]fyne.CanvasObject, 0, len(a.hits))
	for _, h := range a.hits {
		cards = append(cards, a.card(h))
	}
	return container.New(layout.NewGridWrapLayout(
		fyne.NewSize(tokens.ModCardWidth, tokens.ModCardHeight)), cards...)
}

// card is one grid entry: icon, title, description, categories, action.
func (a *browseActivity) card(h modrinth.Hit) fyne.CanvasObject {
	title := widgets.Body(h.Title)
	title.Truncation = fyne.TextTruncateEllipsis

	desc := widget.NewLabel(h.Description)
	desc.Wrapping = fyne.TextWrapWord
	desc.SizeName = fynetheme.SizeNameCaptionText

	// The detail line is a truncating label rather than canvas text: its
	// length is whatever Modrinth returned, and a card is a fixed size, so
	// anything longer has to be cut off rather than drawn over the card
	// beside it.
	lines := container.NewVBox(title)
	if detail := a.detailLine(h); detail != "" {
		lines.Add(widgets.Line(detail))
	}

	head := container.NewBorder(nil, nil,
		widgets.RemoteImage(h.IconURL, tokens.IconModCard), nil, lines,
	)

	body := container.NewBorder(
		head, a.cardAction(h), nil, nil,
		container.NewVScroll(desc),
	)

	return widgets.NewPanel(body)
}

// listRow is one compact entry: small icon, title, description, and
// whichever details the settings ask for.
//
// Every line is a Label, so they share a left edge and every one of them
// ellipsises. Canvas text would align just as well but would draw its
// whole string over whatever sits to the right of it.
func (a *browseActivity) listRow(h modrinth.Hit) fyne.CanvasObject {
	title := widgets.Body(h.Title)
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Truncation = fyne.TextTruncateEllipsis

	lines := []fyne.CanvasObject{title, widgets.Line(h.Description)}
	if detail := a.detailLine(h); detail != "" {
		lines = append(lines, widgets.Line(detail))
	}

	row := container.NewBorder(nil, nil,
		widgets.RemoteImage(h.IconURL, tokens.IconInline*2), a.cardAction(h),
		container.NewVBox(lines...),
	)

	return container.NewVBox(widgets.Inset(tokens.SpaceSM, tokens.SpaceXS, row), widgets.Hairline())
}

// detailLine assembles the optional detail line from the settings.
func (a *browseActivity) detailLine(h modrinth.Hit) string {
	prefs := a.deps.sess.Cfg.Get().Browser

	var parts []string
	if prefs.ShowTags {
		if tags := categoryLine(h); tags != "" {
			parts = append(parts, tags)
		}
	}
	if prefs.ShowLicense && h.License != "" {
		parts = append(parts, h.License)
	}
	if prefs.ShowSides {
		parts = append(parts, h.Sides())
	}
	return strings.Join(parts, "  ")
}

// cardAction is the add control, or for a mod the pack already has, a
// note saying so beside a control to take it back out.
func (a *browseActivity) cardAction(h modrinth.Hit) fyne.CanvasObject {
	if installed, ok := a.installed[h.ProjectID]; ok {
		remove := widget.NewButtonWithIcon("", fynetheme.DeleteIcon(), func() {
			confirmRemoveMod(a.deps, installed)
		})
		remove.Importance = widget.LowImportance

		return container.NewHBox(
			widgets.Status("In this pack", widgets.StateSuccess), remove)
	}

	add := widget.NewButtonWithIcon("Add", fynetheme.ContentAddIcon(), func() {
		a.add(h)
	})

	open := widget.NewButtonWithIcon("", fynetheme.ComputerIcon(), func() {
		_ = sysopen.Browse("https://modrinth.com/mod/" + h.Ref())
	})
	open.Importance = widget.LowImportance

	return container.NewHBox(add, open)
}

// add installs a project into the pack.
//
// packwiz is given the project id rather than the search text: it
// resolves an ambiguous search interactively, and the app runs with
// prompts suppressed.
func (a *browseActivity) add(h modrinth.Hit) {
	deps := a.deps.installDependencies()

	a.deps.run("add "+h.Title, func(ctx context.Context) error {
		return exec(func() (cmdrun.Result, error) {
			return a.deps.client().AddModrinth(ctx, h.Ref(), deps)
		})
	})
}

// categoryLine renders a hit's categories, capped so a card's second line
// stays one line.
func categoryLine(h modrinth.Hit) string {
	const maxCategories = 3

	cats := h.Categories
	if len(cats) > maxCategories {
		cats = cats[:maxCategories]
	}
	return strings.Join(cats, ", ")
}
