// Package packwiz locates, installs and drives the packwiz binary. The
// app never reimplements packwiz's TOML, hashing or dependency logic. It
// shells out for every pack operation and parses what comes back.
package packwiz

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// binaryName is the executable to look for on this OS.
var binaryName = func() string {
	if runtime.GOOS == "windows" {
		return "packwiz.exe"
	}
	return "packwiz"
}()

// ErrNotFound reports that no usable packwiz binary could be located.
// The UI responds by offering to browse for one or download it.
var ErrNotFound = errors.New("packwiz binary not found")

// Location is a resolved packwiz binary.
type Location struct {
	// Path is the absolute path to the binary.
	Path string
	// Version is whatever `packwiz --version` reported, trimmed. It may
	// be empty: some builds print nothing useful, which is not a reason
	// to reject the binary.
	Version string
	// FromPATH records that the binary was found on PATH rather than at
	// a configured or installed path, so settings can show where it came
	// from.
	FromPATH bool
}

// Locate resolves a usable binary. A configured path wins, so a user
// override is never silently replaced by something on PATH. The managed
// install location is tried next, then PATH.
//
// A configured path that no longer works is reported as an error rather
// than skipped: silently falling back would hide the fact that the user's
// chosen binary is broken.
func Locate(ctx context.Context, configured string) (Location, error) {
	if configured != "" {
		return Verify(ctx, configured)
	}

	if managed, err := managedPath(); err == nil {
		if loc, err := Verify(ctx, managed); err == nil {
			return loc, nil
		}
	}

	path, err := exec.LookPath(binaryName)
	if err != nil {
		return Location{}, ErrNotFound
	}
	loc, err := Verify(ctx, path)
	if err != nil {
		return Location{}, err
	}
	loc.FromPATH = true
	return loc, nil
}

// Verify checks that path is an executable packwiz and reports its
// version. It is what the settings screen calls after the user browses
// to a binary, so the failure is visible before the path is saved.
func Verify(ctx context.Context, path string) (Location, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return Location{}, err
	}

	info, err := os.Stat(abs)
	if err != nil {
		return Location{}, err
	}
	if info.IsDir() {
		return Location{}, errors.New(abs + " is a directory, not a binary")
	}

	ctx, cancel := context.WithTimeout(ctx, versionTimeout)
	defer cancel()

	out, err := exec.CommandContext(ctx, abs, "--version").CombinedOutput()
	if err != nil {
		return Location{}, &UnusableError{Path: abs, Err: err, Output: string(out)}
	}
	return Location{Path: abs, Version: parseVersion(string(out))}, nil
}

// versionTimeout bounds the version probe. A binary that does not answer
// promptly is treated as unusable rather than hanging startup.
const versionTimeout = 5 * time.Second

// UnusableError reports a path that exists but does not behave like
// packwiz.
type UnusableError struct {
	Path   string
	Err    error
	Output string
}

func (e *UnusableError) Error() string {
	msg := e.Path + " is not a working packwiz binary: " + e.Err.Error()
	if out := strings.TrimSpace(e.Output); out != "" {
		msg += " (" + firstLine(out) + ")"
	}
	return msg
}

func (e *UnusableError) Unwrap() error { return e.Err }

// parseVersion pulls a version out of `packwiz --version` output. The
// format is not guaranteed stable across releases, so anything
// unrecognised is returned as-is rather than rejected.
func parseVersion(out string) string {
	line := firstLine(strings.TrimSpace(out))
	for _, prefix := range []string{"packwiz version ", "packwiz "} {
		if rest, ok := cutPrefixFold(line, prefix); ok {
			return strings.TrimSpace(rest)
		}
	}
	return line
}

func firstLine(s string) string {
	if i := strings.IndexAny(s, "\r\n"); i >= 0 {
		return s[:i]
	}
	return s
}

func cutPrefixFold(s, prefix string) (string, bool) {
	if len(s) >= len(prefix) && strings.EqualFold(s[:len(prefix)], prefix) {
		return s[len(prefix):], true
	}
	return "", false
}

// managedPath is where an app-installed binary lives: alongside the
// config rather than in a cache directory, so the OS will not reclaim a
// binary the app depends on.
func managedPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "PackwizStudio", "bin", binaryName), nil
}
