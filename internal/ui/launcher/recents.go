package launcher

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	"github.com/PalisadeMC/Packwiz-Studio/internal/config"
	"github.com/PalisadeMC/Packwiz-Studio/internal/pack"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/widgets"
)

// buildRecents is the launcher's main screen: known packs on the left,
// actions on the right.
func (w *Window) buildRecents() fyne.CanvasObject {
	left := container.NewBorder(
		widgets.SubHeading("Recent"), nil, nil, nil,
		w.recentsList(),
	)

	body := container.NewBorder(
		widgets.Heading("Packwiz Studio"), nil, nil,
		container.NewVBox(w.actionColumn()...),
		left,
	)

	return widgets.Inset(tokens.SpaceXL, tokens.SpaceLG, body)
}

// recentsList renders the known packs, newest first.
//
// Rows are built eagerly into a scroller rather than driven by a
// widget.List. A recents list holds a handful of entries, and building
// them outright keeps each row a plain container with its own remove
// button, which List's cell reuse would make awkward.
func (w *Window) recentsList() fyne.CanvasObject {
	packs := w.sess.Cfg.Packs()
	if len(packs) == 0 {
		return w.emptyRecents()
	}

	rows := make([]fyne.CanvasObject, 0, len(packs))
	for _, p := range packs {
		rows = append(rows, w.recentRow(p))
	}

	return container.NewVScroll(container.NewVBox(rows...))
}

// emptyRecents is what a first run looks like.
func (w *Window) emptyRecents() fyne.CanvasObject {
	return container.NewVBox(
		widgets.VSpace(tokens.SpaceMD),
		widgets.Muted("No packs yet."),
		widgets.Dim("Create one, or open a folder that holds a pack.toml."),
	)
}

// recentRow is one entry: the pack's image if it has one, its name, its
// versions and its path, with a control that forgets it without touching
// anything on disk.
//
// The text lines are laid out as a fixed stack rather than a box, so the
// three lines sit at even spacing and the row is the same height whether
// or not the pack has an image.
func (w *Window) recentRow(p config.Pack) fyne.CanvasObject {
	// All three lines are canvas text so they share one left edge: a Label
	// would add the theme's inner padding and sit indented from the rest.
	details := container.New(&stackedLines{},
		widgets.Strong(p.Name),
		widgets.Caption(versionLine(p)),
		widgets.Dim(p.Path),
	)

	body := fyne.CanvasObject(details)
	if logo := widgets.PackLogo(pack.IconPath(p.Path), tokens.IconPackLogo); logo != nil {
		body = container.NewBorder(nil, nil,
			container.NewPadded(logo), nil, details)
	}

	card := widgets.NewClickable(body, func() { w.openPack(p.Path) })
	if !w.sess.HasPackwiz() {
		card.OnTapped = nil
	}

	return container.NewBorder(nil, nil, nil, w.forgetButton(p), card)
}

// versionLine is the "1.20.1 with fabric" line under a pack's name. Both
// halves are optional: a pack.toml is not required to name either.
func versionLine(p config.Pack) string {
	switch {
	case p.MCVersion != "" && p.Loader != "":
		return p.MCVersion + " with " + p.Loader
	case p.MCVersion != "":
		return p.MCVersion
	case p.Loader != "":
		return p.Loader
	default:
		return "unknown version"
	}
}
