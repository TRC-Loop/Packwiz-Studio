package launcher

import (
	"errors"
	"strings"
)

// errAlreadyAPack reports a folder that already holds a pack, so the new
// pack form refuses rather than letting packwiz init overwrite it.
var errAlreadyAPack = errors.New(
	"that folder already holds a pack.toml, open it instead of creating a new pack")

// initFailed turns packwiz's own complaint into an error for a dialog.
// The output is passed through rather than interpreted: packwiz's
// messages are not a stable format, and its wording is more useful than
// anything the app could infer from it.
func initFailed(output string) error {
	msg := strings.TrimSpace(output)
	if msg == "" {
		return errors.New("packwiz init failed without reporting why, see the log")
	}
	return errors.New("packwiz init failed: " + msg)
}
