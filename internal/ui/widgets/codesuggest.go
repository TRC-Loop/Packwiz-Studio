package widgets

import (
	"sort"
	"strings"

	"fyne.io/fyne/v2"
)

// suggest offers what the word being typed could become.
func (e *codeEntry) suggest() {
	if e.menu == nil {
		return
	}

	prefix := wordBefore(e.Text, e.CursorTextOffset())
	if len([]rune(prefix)) < completionMinPrefix {
		e.menu.hide()
		return
	}

	items := candidates(prefix, e.words, e.Text)
	if len(items) == 0 {
		e.menu.hide()
		return
	}

	cell := codeCell()
	origin := codeOrigin(e.code.gutter.width)

	e.menu.show(items, prefix, fyne.NewPos(
		origin.X+cell.Width*float32(e.CursorColumn),
		origin.Y+cell.Height*float32(e.CursorRow+1),
	))
}

// accept takes the highlighted suggestion.
func (e *codeEntry) accept() {
	e.take(e.menu.selected())
}

// take puts one suggestion in place of the word being typed. It is also
// what a click on a row calls.
func (e *codeEntry) take(item string) {
	prefix := e.menu.prefix
	e.menu.hide()

	if item == "" {
		return
	}
	e.replaceBefore(len([]rune(prefix)), item)
}

// wordBefore is the run of word characters ending at an offset. Dots and
// dashes count, because the words worth completing here are keys like
// update.modrinth and hash-format.
func wordBefore(text string, at int) string {
	runes := []rune(text)
	if at > len(runes) {
		at = len(runes)
	}

	start := at
	for start > 0 && isWordRune(runes[start-1]) {
		start--
	}
	return string(runes[start:at])
}

func isWordRune(r rune) bool {
	switch {
	case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		return true
	}
	return r == '_' || r == '.' || r == '-' || r == '[' || r == '$'
}

// candidates are the language's own words plus every word already in the
// file, which between them cover both the keys a format accepts and the
// names this particular file has invented.
func candidates(prefix string, words []string, text string) []string {
	seen := map[string]bool{prefix: true}
	var out []string

	add := func(word string) {
		if seen[word] || !strings.HasPrefix(strings.ToLower(word), strings.ToLower(prefix)) {
			return
		}
		seen[word] = true
		out = append(out, word)
	}

	for _, w := range words {
		add(w)
	}
	for _, w := range strings.FieldsFunc(text, func(r rune) bool { return !isWordRune(r) }) {
		if len([]rune(w)) > completionMinPrefix {
			add(w)
		}
	}

	sort.Strings(out)
	if len(out) > completionLimit {
		out = out[:completionLimit]
	}
	return out
}
