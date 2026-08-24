package packwin

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	fynetheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/PalisadeMC/Packwiz-Studio/internal/pack"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/widgets"
)

// sideSummary reports how the pack's mods are split across client and
// server, with a way to go and change them.
//
// A Modrinth pack is not exported per side. One .mrpack carries every mod
// with a client and server flag on each, and the launcher installing it
// decides what to take. packwiz's export has no side option for that
// reason, so what actually controls the split is each mod's own flag.
// This shows the resulting split and points at where to change it.
func (w *Window) sideSummary() fyne.CanvasObject {
	mods, err := w.pack.Mods()
	if err != nil {
		return widgets.Note("The mod list could not be read: " + err.Error())
	}

	var both, client, server int
	for _, m := range mods {
		switch m.SideFlag {
		case pack.SideClient:
			client++
		case pack.SideServer:
			server++
		default:
			both++
		}
	}

	edit := widget.NewButtonWithIcon("Edit mod sides", fynetheme.ListIcon(), func() {
		w.Select(ActivityMods)
	})
	edit.Importance = widget.LowImportance

	return container.NewVBox(
		widgets.Caption(sideLine(both, client, server)),
		widgets.Note("The export always contains every mod. Each one carries a "+
			"client and server flag, and the launcher installing the pack "+
			"honours it, so the split is set per mod rather than per export."),
		container.NewHBox(edit),
	)
}

// sideLine renders the split as one line.
func sideLine(both, client, server int) string {
	if both+client+server == 0 {
		return "No mods yet"
	}

	line := strconv.Itoa(both) + " on both sides"
	if client > 0 {
		line += ", " + strconv.Itoa(client) + " client only"
	}
	if server > 0 {
		line += ", " + strconv.Itoa(server) + " server only"
	}
	return line
}
