package packwin

import (
	"context"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	fynetheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/PalisadeMC/Packwiz-Studio/internal/cmdrun"
	"github.com/PalisadeMC/Packwiz-Studio/internal/pack"
	"github.com/PalisadeMC/Packwiz-Studio/internal/sysopen"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/widgets"
)

// showDetail renders the selected mod in the detail pane.
func (a *modsActivity) showDetail(m pack.Mod) {
	if m.LoadErr != nil {
		a.showMessage(m.Path + " could not be read: " + m.LoadErr.Error())
		return
	}

	body := container.NewVBox(
		widgets.Heading(m.Name),
		widgets.Muted(m.Filename),
		widgets.VSpace(tokens.SpaceMD),

		a.sideRow(m),
		widgets.VSpace(tokens.SpaceMD),

		a.detailRows(m),
		widgets.VSpace(tokens.SpaceLG),

		a.actionRow(m),
	)

	a.main.Objects = []fyne.CanvasObject{
		container.NewVScroll(widgets.Inset(tokens.SpaceXL, tokens.SpaceLG, body)),
	}
	a.main.Refresh()
}

// sideRow is the client and server control.
func (a *modsActivity) sideRow(m pack.Mod) fyne.CanvasObject {
	labels := make([]string, 0, len(pack.Sides))
	for _, s := range pack.Sides {
		labels = append(labels, s.Label())
	}

	sel := widget.NewSelect(labels, nil)
	sel.SetSelected(m.SideFlag.Label())
	sel.OnChanged = func(chosen string) {
		for _, s := range pack.Sides {
			if s.Label() == chosen && s != m.SideFlag {
				a.applySide(m, s)
				return
			}
		}
	}

	return container.NewBorder(nil, nil, widgets.Muted("Installed on"), nil,
		container.NewHBox(sel))
}

// detailRows lists the mod's metadata.
func (a *modsActivity) detailRows(m pack.Mod) fyne.CanvasObject {
	rows := container.NewVBox(
		labelled("Metafile", m.Path),
		labelled("Modrinth project", orDash(m.ModrinthID)),
		labelled("Version", orDash(m.VersionID)),
	)

	if m.Pinned {
		rows.Add(labelled("Updates", "pinned to this version"))
	}
	if m.Optional {
		rows.Add(labelled("Install", "optional"))
	}
	return rows
}

// actionRow is the buttons that act on the mod.
func (a *modsActivity) actionRow(m pack.Mod) fyne.CanvasObject {
	update := widget.NewButtonWithIcon("Update", fynetheme.ViewRefreshIcon(), func() {
		a.run("update "+m.Slug(), func(ctx context.Context) error {
			return exec(func() (cmdrun.Result, error) {
				return a.deps.client().Update(ctx, m.Slug())
			})
		})
	})
	if m.Pinned {
		update.Disable()
	}

	pin := widget.NewButtonWithIcon(pinLabel(m), fynetheme.VisibilityOffIcon(), func() {
		a.run(pinLabel(m), func(ctx context.Context) error {
			return exec(func() (cmdrun.Result, error) {
				client := a.deps.client()
				if m.Pinned {
					return client.Unpin(ctx, m.Slug())
				}
				return client.Pin(ctx, m.Slug())
			})
		})
	})

	remove := widget.NewButtonWithIcon("Remove", fynetheme.DeleteIcon(), func() {
		a.confirmRemove(m)
	})
	remove.Importance = widget.DangerImportance

	row := container.NewHBox(update, pin, remove)

	if m.ModrinthID != "" {
		open := widget.NewButtonWithIcon("Open on Modrinth", fynetheme.ComputerIcon(), func() {
			if err := sysopen.Browse(modrinthURL(m.ModrinthID)); err != nil {
				dialog.ShowError(err, a.deps.win)
			}
		})
		open.Importance = widget.LowImportance
		row.Add(open)
	}

	return row
}

// confirmRemove asks before deleting a mod.
func (a *modsActivity) confirmRemove(m pack.Mod) {
	confirmRemoveMod(a.deps, m)
}

// applySide writes the new side flag and refreshes the index, since the
// index carries a hash of every metafile.
func (a *modsActivity) applySide(m pack.Mod, side pack.Side) {
	a.run("side "+m.Slug(), func(ctx context.Context) error {
		if err := pack.SetSide(a.deps.pack.Dir, m.Path, side); err != nil {
			return err
		}
		return exec(func() (cmdrun.Result, error) {
			return a.deps.client().Refresh(ctx)
		})
	})
}

// labelled is one metadata row: a muted caption and its value.
func labelled(name, value string) fyne.CanvasObject {
	return container.NewBorder(nil, nil, widgets.Muted(name), nil,
		container.NewHBox(widgets.Caption(value)))
}

func orDash(s string) string {
	if s == "" {
		return "not recorded"
	}
	return s
}

func pinLabel(m pack.Mod) string {
	if m.Pinned {
		return "Unpin"
	}
	return "Pin"
}

// modrinthURL is a project's page on Modrinth. The project id works in
// place of the slug, so no extra lookup is needed to open it.
func modrinthURL(id string) string {
	return "https://modrinth.com/mod/" + id
}

func itoa(n int) string { return strconv.Itoa(n) }
