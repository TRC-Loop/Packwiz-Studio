package widgets

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// codeEntry is the editing half of Code: an entry that indents, closes
// brackets and completes words.
//
// Edits are made by rewriting the text and putting the caret back, since
// Fyne's Entry offers no way to insert at the caret. That costs the undo
// history its finer steps, which is a fair trade for the editing this
// screen exists for.
type codeEntry struct {
	widget.Entry

	code  *Code
	words []string
	menu  *completions
}

// pairs are the brackets and quotes that close themselves.
var pairs = map[rune]rune{'{': '}', '[': ']', '(': ')', '"': '"'}

func newCodeEntry(c *Code) *codeEntry {
	e := &codeEntry{code: c}
	e.MultiLine = true
	e.TextStyle = fyne.TextStyle{Monospace: true}

	// Wrapping off and scrolling off together are what make the entry size
	// itself to its content. It then sits inside the editor's own scroll,
	// which the coloured layer shares, so the two cannot drift apart.
	e.Wrapping = fyne.TextWrapOff
	e.Scroll = fyne.ScrollNone

	e.ExtendBaseWidget(e)

	e.OnChanged = c.changed
	e.OnCursorChanged = c.markLine

	return e
}

// TypedKey handles the keys that mean something to an editor before the
// entry gets them.
func (e *codeEntry) TypedKey(key *fyne.KeyEvent) {
	if e.menu != nil && e.menu.open() {
		switch key.Name {
		case fyne.KeyEscape:
			e.menu.hide()
			return
		case fyne.KeyUp:
			e.menu.move(-1)
			return
		case fyne.KeyDown:
			e.menu.move(1)
			return
		case fyne.KeyReturn, fyne.KeyEnter, fyne.KeyTab:
			e.accept()
			return
		}
		e.menu.hide()
	}

	if key.Name == fyne.KeyReturn || key.Name == fyne.KeyEnter {
		if e.newlineKeepsIndent() {
			return
		}
	}
	e.Entry.TypedKey(key)
}

// TypedRune closes brackets and offers completions as words are typed.
func (e *codeEntry) TypedRune(r rune) {
	if closer, ok := pairs[r]; ok {
		e.insert(string(r)+string(closer), 1)
		e.suggest()
		return
	}
	if e.typeOver(r) {
		return
	}

	e.Entry.TypedRune(r)
	e.suggest()
}

// typeOver moves past a closing character the editor added itself rather
// than inserting a second one.
func (e *codeEntry) typeOver(r rune) bool {
	if !strings.ContainsRune("}])\"", r) {
		return false
	}

	runes := []rune(e.Text)
	at := e.CursorTextOffset()
	if at >= len(runes) || runes[at] != r {
		return false
	}

	e.moveTo(at + 1)
	return true
}

// newlineKeepsIndent starts the next line under the current one, one step
// further in after a line that opened a block.
func (e *codeEntry) newlineKeepsIndent() bool {
	lines := strings.Split(e.Text, "\n")
	if e.CursorRow >= len(lines) {
		return false
	}

	line := lines[e.CursorRow]
	indent := indentOf(line)

	if opensBlock(strings.TrimRight(line, " \t")) {
		indent += "  "
	}
	if indent == "" {
		return false
	}

	e.insert("\n"+indent, 0)
	return true
}

// opensBlock reports a line that ends by opening something.
func opensBlock(line string) bool {
	if line == "" {
		return false
	}
	return strings.ContainsRune("{[(:", rune(line[len(line)-1]))
}

// insert puts text in at the caret and leaves the caret back by the given
// number of runes, which is how an inserted closing bracket ends up on
// the far side of it.
func (e *codeEntry) insert(text string, back int) {
	runes := []rune(e.Text)
	at := e.CursorTextOffset()
	if at > len(runes) {
		at = len(runes)
	}

	e.SetText(string(runes[:at]) + text + string(runes[at:]))
	e.moveTo(at + len([]rune(text)) - back)
}

// replaceBefore swaps the run of text just before the caret, which is how
// a completion takes the place of what was typed.
func (e *codeEntry) replaceBefore(count int, text string) {
	runes := []rune(e.Text)
	at := e.CursorTextOffset()
	if at > len(runes) {
		at = len(runes)
	}
	start := max(at-count, 0)

	e.SetText(string(runes[:start]) + text + string(runes[at:]))
	e.moveTo(start + len([]rune(text)))
}

// moveTo puts the caret at a rune offset into the whole text.
func (e *codeEntry) moveTo(offset int) {
	row, col := 0, offset

	for _, line := range strings.Split(e.Text, "\n") {
		length := len([]rune(line))
		if col <= length {
			break
		}
		col -= length + 1
		row++
	}

	e.CursorRow, e.CursorColumn = row, max(col, 0)
	e.Refresh()
}
