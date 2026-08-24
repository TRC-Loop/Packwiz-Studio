package packwin

import (
	"context"
	"strconv"

	"github.com/TRC-Loop/Packwiz-Studio/internal/cmdrun"
)

// result is the command outcome type the git helpers pass around.
type result = cmdrun.Result

// execGit turns a git command's refusal into an error, the same way exec
// does for packwiz.
func execGit(call func() (result, error)) error { return exec(call) }

// runGit runs a git action through the shared runner, so it gets the same
// busy guard, error dialog and reload as a packwiz action.
func (a *gitActivity) runGit(label string, work func(context.Context) error) {
	a.deps.run(label, work)
}

// plural renders a count with its noun.
func plural(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return strconv.Itoa(n) + " " + noun + "s"
}
