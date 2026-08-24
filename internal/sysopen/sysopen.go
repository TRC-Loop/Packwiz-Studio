// Package sysopen hands a path or a URL to the desktop environment.
//
// These are the only commands the app runs that are not packwiz or git,
// and they are deliberately not logged: opening a folder is not part of
// the pack's history.
package sysopen

import (
	"errors"
	"os/exec"
	"runtime"
)

// ErrUnsupported reports a platform with no known opener.
var ErrUnsupported = errors.New("no way to open files on this platform")

// Reveal shows a file or folder in the desktop file manager.
func Reveal(path string) error {
	switch runtime.GOOS {
	case "darwin":
		return start("open", path)
	case "windows":
		// explorer exits non-zero even when it succeeds, so its status is
		// deliberately ignored.
		_ = start("explorer", path)
		return nil
	case "linux", "freebsd", "openbsd", "netbsd":
		return start("xdg-open", path)
	default:
		return ErrUnsupported
	}
}

// Browse opens a URL in the default browser.
func Browse(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return start("open", url)
	case "windows":
		return start("rundll32", "url.dll,FileProtocolHandler", url)
	case "linux", "freebsd", "openbsd", "netbsd":
		return start("xdg-open", url)
	default:
		return ErrUnsupported
	}
}

// start launches a command without waiting for it. These openers hand off
// to another process and may outlive the app, so waiting on them would
// block for as long as the file manager stays open.
func start(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()
	return nil
}
