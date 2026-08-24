package packwin

import (
	"context"

	"fyne.io/fyne/v2"

	"github.com/PalisadeMC/Packwiz-Studio/internal/modrinth"
)

// runSearch queries Modrinth from the given offset. An offset of zero
// replaces the results; anything else appends a page.
//
// Each search takes a ticket, and a reply is dropped unless its ticket is
// still the current one. Typing quickly would otherwise let a slow early
// reply land after a fast later one.
func (a *browseActivity) runSearch(offset int) {
	a.searchID++
	ticket := a.searchID
	a.loading = true

	query := modrinth.Query{Text: a.search.Text, Offset: offset}
	if a.onlyPack.Checked {
		query.MCVersion = a.deps.pack.MCVersion
		query.Loader = a.deps.pack.Loader
	}

	if offset == 0 {
		a.setMessage("Searching Modrinth")
	}

	go func() {
		res, err := a.client.Search(context.Background(), query)

		fyne.Do(func() {
			if ticket != a.searchID {
				return
			}
			a.loading = false

			if err != nil {
				a.setMessage(err.Error())
				return
			}

			if offset == 0 {
				a.hits = res.Hits
			} else {
				a.hits = append(a.hits, res.Hits...)
			}
			a.total = res.Total
			a.offset = res.Offset

			a.clearMessage()
			a.renderResults()
		})
	}()
}

// maybeLoadMore fetches the next page when the view nears the end.
//
// The trigger is a margin above the bottom rather than the bottom itself,
// so the next page is on its way before the user runs out of list.
func (a *browseActivity) maybeLoadMore(pos fyne.Position) {
	if a.loading || len(a.hits) == 0 || len(a.hits) >= a.total {
		return
	}

	content := a.scroll.Content.Size().Height
	viewport := a.scroll.Size().Height
	if content <= viewport {
		return
	}

	if pos.Y+viewport >= content-loadMoreMargin {
		a.runSearch(a.offset + modrinth.DefaultLimit)
	}
}

// loadMoreMargin is how far above the end the next page is fetched.
const loadMoreMargin float32 = 320
