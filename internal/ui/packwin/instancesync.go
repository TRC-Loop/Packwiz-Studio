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
	"github.com/PalisadeMC/Packwiz-Studio/internal/instance"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/widgets"
)

// compare rebuilds the list of differences.
//
// The comparison hashes every file in the chosen folders, which for a
// pack full of configs is thousands of reads, so it runs off the main
// goroutine. Nothing else in the window waits for it.
func (a *instanceActivity) compare() {
	dir := a.path.Text
	if !instance.Exists(dir) {
		a.entries = nil
		a.setStatus("That folder is not there.")
		return
	}

	a.remember()
	a.compared = true

	packDir := a.deps.pack.Dir
	folders := a.chosen()

	a.status.SetText("Comparing.")

	go func() {
		entries, err := instance.Diff(packDir, dir, folders)

		fyne.Do(func() {
			if err != nil {
				a.entries = nil
				a.setStatus("The comparison failed: " + err.Error())
				return
			}

			a.entries = entries
			if len(entries) == 0 {
				a.setStatus("The pack and the instance match.")
				return
			}
			a.setStatus(strconv.Itoa(len(entries)) + " files differ.")
		})
	}()
}

// setStatus restates the summary and redraws.
func (a *instanceActivity) setStatus(text string) {
	a.status.SetText(text)
	a.render()
}

// renderList rebuilds the side panel: one row per differing file, with a
// control for each direction.
func (a *instanceActivity) renderList() {
	a.list.Objects = nil

	if len(a.entries) == 0 {
		a.list.Add(widgets.Inset(tokens.SpaceMD, tokens.SpaceSM,
			widgets.Dim("Nothing to sync")))
		a.list.Refresh()
		return
	}

	for _, e := range a.entries {
		a.list.Add(a.row(e))
	}
	a.list.Refresh()
}

// row is one differing file.
func (a *instanceActivity) row(e instance.Entry) fyne.CanvasObject {
	entry := e

	name := widgets.Body(e.Rel)
	name.Truncation = fyne.TextTruncateEllipsis

	toPack := widget.NewButtonWithIcon("", fynetheme.MoveDownIcon(), func() {
		a.copyToPack(entry)
	})
	toInstance := widget.NewButtonWithIcon("", fynetheme.MoveUpIcon(), func() {
		a.copyToInstance(entry)
	})
	toPack.Importance = widget.LowImportance
	toInstance.Importance = widget.LowImportance

	lines := container.NewVBox(name, widgets.Caption(entry.State.Label()))
	row := container.NewBorder(nil, nil, nil,
		container.NewHBox(toPack, toInstance), lines)

	return container.NewVBox(
		widgets.Inset(tokens.SpaceSM, tokens.SpaceXS, row), widgets.Hairline())
}

// copyToPack brings one file into the pack, or takes it out when the
// instance no longer has it.
func (a *instanceActivity) copyToPack(e instance.Entry) {
	dir := a.path.Text

	a.deps.run("copy "+e.Rel+" to the pack", func(ctx context.Context) error {
		if e.State == instance.OnlyInPack {
			if err := instance.Remove(a.deps.pack.Dir, e.Rel); err != nil {
				return err
			}
		} else if err := instance.CopyFile(dir, a.deps.pack.Dir, e.Rel); err != nil {
			return err
		}

		return exec(func() (cmdrun.Result, error) {
			return a.deps.client().Refresh(ctx)
		})
	})
}

// copyToInstance pushes one file out to the instance. The index is not
// refreshed: nothing in the pack changed.
func (a *instanceActivity) copyToInstance(e instance.Entry) {
	dir := a.path.Text

	a.deps.run("copy "+e.Rel+" to the instance", func(context.Context) error {
		if e.State == instance.OnlyInInstance {
			return instance.Remove(dir, e.Rel)
		}
		return instance.CopyFile(a.deps.pack.Dir, dir, e.Rel)
	})
}

// copyAllToPack takes every difference into the pack at once, which is
// the end of a play session where several configs were tweaked.
func (a *instanceActivity) copyAllToPack() {
	dir := a.path.Text
	entries := a.entries

	if len(entries) == 0 {
		return
	}

	dialog.NewConfirm("Copy to the pack",
		"Copy "+strconv.Itoa(len(entries))+" files from the instance into the pack?\n\n"+
			"Files the instance no longer has are deleted from the pack.",
		func(ok bool) {
			if !ok {
				return
			}
			a.deps.run("sync from the instance", func(ctx context.Context) error {
				for _, e := range entries {
					var err error
					if e.State == instance.OnlyInPack {
						err = instance.Remove(a.deps.pack.Dir, e.Rel)
					} else {
						err = instance.CopyFile(dir, a.deps.pack.Dir, e.Rel)
					}
					if err != nil {
						return err
					}
				}
				return exec(func() (cmdrun.Result, error) {
					return a.deps.client().Refresh(ctx)
				})
			})
		}, a.deps.win).Show()
}
