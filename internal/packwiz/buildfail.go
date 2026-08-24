package packwiz

import (
	"errors"
	"strings"

	"github.com/TRC-Loop/Packwiz-Studio/internal/cmdrun"
)

// buildFailed turns a failed source build into an error carrying Go's own
// message, which explains the failure better than an exit code.
func buildFailed(res cmdrun.Result) error {
	if msg := strings.TrimSpace(res.Output()); msg != "" {
		return errors.New("building packwiz failed: " + firstLines(msg, 4))
	}
	return errors.New("building packwiz failed, see the output panel")
}

// firstLines trims a long build log down to something a dialog can show.
func firstLines(s string, n int) string {
	lines := strings.Split(s, "\n")
	if len(lines) <= n {
		return s
	}
	return strings.Join(lines[:n], "\n") + "\n..."
}
