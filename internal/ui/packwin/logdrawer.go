package packwin

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	fynetheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/PalisadeMC/Packwiz-Studio/internal/logbus"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/tokens"
	"github.com/PalisadeMC/Packwiz-Studio/internal/ui/widgets"
)

// logDrawer shows raw packwiz and git output as it arrives. It sits above
// the status bar and is collapsed by default.
//
// Output is a TextGrid rather than a Label: it is monospaced, which is
// what command output wants, and it takes a colour per cell so a failing
// line can be marked without building a rich text tree per entry.
type logDrawer struct {
	bus    *logbus.Bus
	grid   *widget.TextGrid
	scroll *container.Scroll
	root   *fyne.Container

	cancel   func()
	autoTail bool
}

// newLogDrawer builds the drawer and replays whatever the bus already
// holds, so opening it after a command has run still shows that output.
func newLogDrawer(bus *logbus.Bus) *logDrawer {
	d := &logDrawer{
		bus:      bus,
		grid:     widget.NewTextGrid(),
		autoTail: true,
	}

	d.scroll = container.NewScroll(d.grid)

	bg := canvas.NewRectangle(tokens.ColorElevated)
	body := container.NewStack(bg, d.scroll)

	d.root = container.NewBorder(
		container.NewBorder(widgets.Hairline(), nil, nil, nil, d.header()),
		nil, nil, nil,
		body,
	)
	d.root.Hide()

	for _, e := range bus.History() {
		d.append(e)
	}
	d.cancel = bus.Subscribe(func(e logbus.Entry) {
		fyne.Do(func() { d.append(e) })
	})

	return d
}

// header is the drawer's own strip: a label and a clear control.
func (d *logDrawer) header() fyne.CanvasObject {
	clear := widget.NewButtonWithIcon("", fynetheme.DeleteIcon(), func() {
		d.bus.Clear()
		d.grid.Rows = nil
		d.grid.Refresh()
	})
	clear.Importance = widget.LowImportance

	return container.NewBorder(nil, nil,
		widgets.Inset(tokens.SpaceMD, tokens.SpaceXS, widgets.Caption("Output")),
		clear,
		nil,
	)
}

// object returns the drawer for placement.
func (d *logDrawer) object() fyne.CanvasObject { return d.root }

// visible reports whether the drawer is open.
func (d *logDrawer) visible() bool { return d.root.Visible() }

// toggle opens or closes the drawer.
func (d *logDrawer) toggle() {
	if d.root.Visible() {
		d.root.Hide()
		return
	}
	d.root.Show()
	d.tail()
}

// close hides the drawer without toggling.
func (d *logDrawer) close() { d.root.Hide() }

// detach stops listening to the bus. The pack window calls this when it
// gives up its window, so a closed drawer does not keep receiving output.
func (d *logDrawer) detach() {
	if d.cancel != nil {
		d.cancel()
		d.cancel = nil
	}
}

// append adds one entry as a row. It must run on the main goroutine.
func (d *logDrawer) append(e logbus.Entry) {
	d.grid.Rows = append(d.grid.Rows, row(text(e), style(e)))

	// Trim to the same bound the bus keeps, so a long session does not
	// grow the grid without limit.
	if over := len(d.grid.Rows) - logbus.DefaultCapacity; over > 0 {
		d.grid.Rows = append(d.grid.Rows[:0], d.grid.Rows[over:]...)
	}

	d.grid.Refresh()
	if d.autoTail && d.root.Visible() {
		d.tail()
	}
}

// tail scrolls to the newest output.
func (d *logDrawer) tail() { d.scroll.ScrollToBottom() }

// text renders an entry's line, prefixing a command so it reads like a
// shell transcript.
func text(e logbus.Entry) string {
	if e.Kind == logbus.KindCommand {
		return "$ " + e.Text
	}
	return e.Text
}

// style picks the colour for an entry.
func style(e logbus.Entry) widget.TextGridStyle {
	switch {
	case e.Failed():
		return &widget.CustomTextGridStyle{FGColor: tokens.ColorError}
	case e.Kind == logbus.KindCommand:
		return &widget.CustomTextGridStyle{
			FGColor:   tokens.ColorText,
			TextStyle: fyne.TextStyle{Bold: true},
		}
	case e.Kind == logbus.KindResult:
		return &widget.CustomTextGridStyle{FGColor: tokens.ColorSuccess}
	case e.Kind == logbus.KindNotice:
		return &widget.CustomTextGridStyle{FGColor: tokens.ColorMuted}
	default:
		return &widget.CustomTextGridStyle{FGColor: tokens.ColorText}
	}
}

// row builds a styled grid row from a line of text.
func row(line string, s widget.TextGridStyle) widget.TextGridRow {
	cells := make([]widget.TextGridCell, 0, len(line))
	for _, r := range line {
		cells = append(cells, widget.TextGridCell{Rune: r, Style: s})
	}
	return widget.TextGridRow{Cells: cells}
}
