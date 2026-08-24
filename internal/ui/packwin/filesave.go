package packwin

import (
	"context"
	"os"
	"path/filepath"

	"github.com/PalisadeMC/Packwiz-Studio/internal/cmdrun"
)

// save writes an edited file and refreshes the pack index.
//
// The index stores a hash of every file it tracks, so an edit that skipped
// the refresh would leave the pack failing its own checks. The refresh is
// therefore part of saving rather than something the user has to remember.
func (a *filesActivity) save(rel, content string) {
	full := filepath.Join(a.deps.pack.Dir, filepath.FromSlash(rel))

	a.deps.run("save "+rel, func(ctx context.Context) error {
		if err := writeFile(full, content); err != nil {
			return err
		}
		return exec(func() (cmdrun.Result, error) {
			return a.deps.client().Refresh(ctx)
		})
	})
}

// writeFile replaces a file's contents, keeping its permissions.
func writeFile(path, content string) error {
	mode := os.FileMode(0o644)
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	}
	return os.WriteFile(path, []byte(content), mode)
}
