package packwiz

import (
	"context"
	"errors"

	"github.com/PalisadeMC/Packwiz-Studio/internal/cmdrun"
)

// Client runs packwiz commands against one pack folder.
//
// Every method blocks until the command finishes and returns the raw
// result, so callers can both display the output and parse it. A non-zero
// exit is reported in Result.ExitCode rather than as an error: a packwiz
// command that refuses to do something is a normal outcome the UI shows,
// not an exceptional one. The error return covers only the cases where
// packwiz could not be run at all.
type Client struct {
	bin    string
	dir    string
	runner *cmdrun.Runner
}

// NewClient returns a Client invoking bin inside the pack folder dir.
func NewClient(bin, dir string, runner *cmdrun.Runner) *Client {
	return &Client{bin: bin, dir: dir, runner: runner}
}

// Dir reports the pack folder this Client operates on.
func (c *Client) Dir() string { return c.dir }

// ErrNoBinary reports a Client built without a resolved packwiz path.
var ErrNoBinary = errors.New("no packwiz binary configured")

// run invokes packwiz with args.
//
// Every invocation passes -y. packwiz prompts on stdin for confirmations
// and for ambiguous searches; a GUI has no terminal to answer with, so an
// un-suppressed prompt would look like a hang. The app therefore only
// ever passes unambiguous arguments (an exact project ID from the
// browser, an exact mod name from the pack) and accepts the defaults.
func (c *Client) run(ctx context.Context, args ...string) (cmdrun.Result, error) {
	if c.bin == "" {
		return cmdrun.Result{ExitCode: -1}, ErrNoBinary
	}
	return c.runner.Run(ctx, cmdrun.Spec{
		Name: c.bin,
		Args: append([]string{"-y"}, args...),
		Dir:  c.dir,
	})
}

// declining is the stdin fed to a command run without -y, answering no to
// whatever it asks. Several lines are supplied because a command may ask
// more than once, and the reader reaching EOF ends the command rather than
// leaving it waiting.
const declining = "n\nn\nn\nn\n"

// runDeclining invokes packwiz without -y, answering no to its prompts.
//
// This is how "do not install dependencies" is expressed: packwiz has no
// flag for it, only a prompt, and -y answers yes to everything. Every
// other call uses -y, so this is reserved for the one case where the
// default answer is not what the user asked for.
func (c *Client) runDeclining(ctx context.Context, args ...string) (cmdrun.Result, error) {
	if c.bin == "" {
		return cmdrun.Result{ExitCode: -1}, ErrNoBinary
	}
	return c.runner.Run(ctx, cmdrun.Spec{
		Name:  c.bin,
		Args:  args,
		Dir:   c.dir,
		Stdin: declining,
	})
}

// runIn invokes packwiz in an explicit directory, for commands that run
// before a pack exists.
func (c *Client) runIn(ctx context.Context, dir string, args ...string) (cmdrun.Result, error) {
	if c.bin == "" {
		return cmdrun.Result{ExitCode: -1}, ErrNoBinary
	}
	return c.runner.Run(ctx, cmdrun.Spec{
		Name: c.bin,
		Args: append([]string{"-y"}, args...),
		Dir:  dir,
	})
}
