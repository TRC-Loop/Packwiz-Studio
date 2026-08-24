package packwin

import "errors"

// errNoExportFolder reports an export with nowhere to write to.
var errNoExportFolder = errors.New("choose a folder to export into")

// errNoRepo reports a git action attempted outside a repository.
var errNoRepo = errors.New(
	"this pack folder is not a git repository yet, initialise one first")

// errNothingStaged reports a commit with no staged changes.
var errNothingStaged = errors.New("stage some changes before committing")

// errNoMessage reports a commit with an empty message.
var errNoMessage = errors.New("enter a commit message")

// errNoRemote reports a push or a release with no remote to send to.
var errNoRemote = errors.New("this repository has no remote named origin")
