package packwiz

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/PalisadeMC/Packwiz-Studio/internal/cmdrun"
)

// Loader identifies a mod loader packwiz can initialise a pack for.
type Loader string

// Supported loaders, spelled as packwiz's --modloader expects them.
const (
	LoaderFabric     Loader = "fabric"
	LoaderForge      Loader = "forge"
	LoaderNeoForge   Loader = "neoforge"
	LoaderQuilt      Loader = "quilt"
	LoaderLiteLoader Loader = "liteloader"
)

// Loaders lists the supported loaders in the order the new-pack form
// offers them.
var Loaders = []Loader{LoaderFabric, LoaderNeoForge, LoaderForge, LoaderQuilt, LoaderLiteLoader}

// Label is the loader's display name for the new-pack form.
func (l Loader) Label() string {
	switch l {
	case LoaderNeoForge:
		return "NeoForge"
	case LoaderLiteLoader:
		return "LiteLoader"
	default:
		return strings.ToUpper(string(l)[:1]) + string(l)[1:]
	}
}

// versionFlag is the flag that sets this loader's version. packwiz uses a
// per-loader flag rather than one shared --loader-version.
func (l Loader) versionFlag() string {
	return "--" + string(l) + "-version"
}

// ParseLoader resolves a loader name, accepting any capitalisation.
func ParseLoader(s string) (Loader, bool) {
	for _, l := range Loaders {
		if strings.EqualFold(string(l), s) {
			return l, true
		}
	}
	return "", false
}

// InitOptions mirrors the prompts of `packwiz init`. The new-pack form
// fills this in and validates it before anything is run.
type InitOptions struct {
	// Dir is the folder the pack is created in. It must exist.
	Dir string
	// Name, Author and Version are the pack's own metadata.
	Name    string
	Author  string
	Version string
	// MCVersion is the Minecraft version, for example "1.20.1".
	MCVersion string
	// Loader is the mod loader.
	Loader Loader
	// LoaderVersion pins the loader. Empty means "use the latest", which
	// passes packwiz's per-loader --*-latest flag.
	LoaderVersion string
	// Reinit recreates pack.toml if the folder already holds a pack.
	Reinit bool
}

// Validate reports why these options cannot be used, so the form can
// refuse before shelling out. The messages are written for display.
func (o InitOptions) Validate() error {
	var problems []string

	if strings.TrimSpace(o.Dir) == "" {
		problems = append(problems, "choose a folder for the pack")
	}
	if strings.TrimSpace(o.Name) == "" {
		problems = append(problems, "enter a pack name")
	}
	if strings.TrimSpace(o.MCVersion) == "" {
		problems = append(problems, "enter a Minecraft version")
	}
	if o.Loader == "" {
		problems = append(problems, "choose a mod loader")
	} else if _, ok := ParseLoader(string(o.Loader)); !ok {
		problems = append(problems, fmt.Sprintf("%q is not a supported mod loader", o.Loader))
	}

	if len(problems) == 0 {
		return nil
	}
	return errors.New(strings.Join(problems, "; "))
}

// args renders the options as packwiz init flags.
func (o InitOptions) args() []string {
	args := []string{"init",
		"--name", strings.TrimSpace(o.Name),
		"--mc-version", strings.TrimSpace(o.MCVersion),
		"--modloader", string(o.Loader),
	}

	// packwiz prompts for an omitted author or version even under -y, so
	// both are always sent, and an empty string is a valid answer.
	args = append(args, "--author", strings.TrimSpace(o.Author))
	args = append(args, "--version", strings.TrimSpace(o.Version))

	if v := strings.TrimSpace(o.LoaderVersion); v != "" {
		args = append(args, o.Loader.versionFlag(), v)
	} else {
		args = append(args, "--"+string(o.Loader)+"-latest")
	}

	if o.Reinit {
		args = append(args, "--reinit")
	}
	return args
}

// Init creates a new pack. The Client's own directory is ignored: a pack
// does not exist yet, so the target folder comes from the options.
func (c *Client) Init(ctx context.Context, o InitOptions) (cmdrun.Result, error) {
	if err := o.Validate(); err != nil {
		return cmdrun.Result{ExitCode: -1}, err
	}
	return c.runIn(ctx, o.Dir, o.args()...)
}
