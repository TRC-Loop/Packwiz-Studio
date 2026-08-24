package packwiz

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/PalisadeMC/Packwiz-Studio/internal/cmdrun"
	"github.com/PalisadeMC/Packwiz-Studio/internal/logbus"
)

// Method is how a packwiz binary was obtained.
type Method int

// Install methods.
const (
	// MethodDownload fetched the prebuilt binary packwiz's CI produces.
	MethodDownload Method = iota
	// MethodBuild compiled packwiz from source with the Go toolchain.
	MethodBuild
)

// Installer obtains a packwiz binary and puts it where the app can find
// it again.
//
// There are two routes because packwiz's CI builds amd64 only. Where a
// prebuilt binary exists for the machine it is downloaded, which needs no
// toolchain. Where one does not, most notably on arm64, it is compiled
// from source instead, which produces a native binary rather than one
// running through emulation.
type Installer struct {
	// Runner logs the build step, so the output panel shows progress.
	Runner *cmdrun.Runner
	// Bus receives commentary about which route was taken.
	Bus *logbus.Bus
	// Client is the HTTP client for the download. Nil uses a default with
	// a generous timeout.
	Client *http.Client
	// Dest overrides where the binary lands. Empty means the managed path
	// beside the app config.
	Dest string
}

// Plan reports how this machine would be served, so the setup screen can
// say what the install button is going to do before it is pressed.
func Plan(ctx context.Context, client *http.Client) (Method, error) {
	if _, err := LatestArtifact(ctx, client); err == nil {
		return MethodDownload, nil
	}
	if GoAvailable() {
		return MethodBuild, nil
	}
	return MethodBuild, ErrNoRoute
}

// ErrNoRoute reports a machine with neither a prebuilt binary nor a
// toolchain to build one.
var ErrNoRoute = errors.New(
	"packwiz publishes no build for this OS and architecture, and building " +
		"one needs the Go toolchain, which is not on your PATH. Install Go " +
		"from go.dev, or point the app at a packwiz binary you already have")

// GoAvailable reports whether a Go toolchain is present.
func GoAvailable() bool {
	_, err := exec.LookPath("go")
	return err == nil
}

// Install obtains packwiz and verifies the result.
func (in *Installer) Install(ctx context.Context, progress Progress) (Location, error) {
	dest, err := in.dest()
	if err != nil {
		return Location{}, err
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return Location{}, err
	}

	artifact, artifactErr := LatestArtifact(ctx, in.client())

	switch {
	case artifactErr == nil:
		in.notice("downloading " + artifact.Name + " from packwiz's latest build")
		err = in.download(ctx, artifact, dest, progress)
	case GoAvailable():
		in.notice("no prebuilt packwiz for this architecture, building it with Go")
		err = in.build(ctx, dest)
	default:
		return Location{}, ErrNoRoute
	}

	if err != nil {
		return Location{}, err
	}

	loc, err := Verify(ctx, dest)
	if err != nil {
		// Something arrived but it does not work. Remove it rather than
		// leaving a broken binary for Locate to find and fail on at
		// every startup.
		os.Remove(dest)
		return Location{}, err
	}
	return loc, nil
}

// download fetches and unpacks a prebuilt binary.
func (in *Installer) download(ctx context.Context, a Artifact, dest string, progress Progress) error {
	archive, cleanup, err := downloadArtifact(ctx, in.client(), a, progress)
	if err != nil {
		return err
	}
	defer cleanup()

	if err := extractBinary(archive, dest); err != nil {
		return err
	}
	return os.Chmod(dest, 0o755)
}

// build compiles packwiz from source into the app's own folder.
func (in *Installer) build(ctx context.Context, dest string) error {
	goBin, err := exec.LookPath("go")
	if err != nil {
		return ErrNoRoute
	}

	runner := in.Runner
	if runner == nil {
		runner = cmdrun.New(nil)
	}

	// GOBIN puts the binary straight into the app's folder, so installing
	// touches neither the user's GOPATH nor their PATH.
	res, err := runner.Run(ctx, cmdrun.Spec{
		Name: goBin,
		Args: []string{"install", modulePath + "@latest"},
		Env:  append(os.Environ(), "GOBIN="+filepath.Dir(dest)),
	})
	if err != nil {
		return err
	}
	if !res.OK() {
		return buildFailed(res)
	}
	return nil
}

// modulePath is packwiz's Go module, used when building from source.
const modulePath = "github.com/packwiz/packwiz"

func (in *Installer) dest() (string, error) {
	if in.Dest != "" {
		return in.Dest, nil
	}
	return managedPath()
}

func (in *Installer) client() *http.Client {
	if in.Client != nil {
		return in.Client
	}
	return &http.Client{Timeout: 10 * time.Minute}
}

func (in *Installer) notice(text string) {
	if in.Bus != nil {
		in.Bus.Publish(logbus.KindNotice, text)
	}
}
