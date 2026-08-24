package launcher

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	"github.com/TRC-Loop/Packwiz-Studio/internal/config"
	"github.com/TRC-Loop/Packwiz-Studio/internal/pack"
	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/tokens"
	"github.com/TRC-Loop/Packwiz-Studio/internal/ui/widgets"
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

// recentRow is one entry: logo, name, versions and path, with a remove
// control that forgets the pack without touching it on disk.
func (w *Window) recentRow(p config.Pack) fyne.CanvasObject {
	details := container.NewVBox(
		widgets.Body(p.Name),
		widgets.Caption(versionLine(p)),
		widgets.Dim(p.Path),
	)

	row := container.NewBorder(
		nil, nil,
		container.NewCenter(widgets.PackLogo(pack.IconPath(p.Path), tokens.IconPackLogo)),
		nil,
		details,
	)

	card := widgets.NewClickable(row, func() { w.openPack(p.Path) })
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
