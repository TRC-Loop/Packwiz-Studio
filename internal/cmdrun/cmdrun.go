// Package cmdrun runs external commands, streaming their output to a
// logbus as it arrives while also accumulating it for the caller to
// parse.
//
// Everything the app does to a pack goes through here: packwiz for pack
// manipulation, git for versioning. Run blocks, so callers invoke it from
// a goroutine and marshal results back onto the UI thread with fyne.Do.
package cmdrun

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/PalisadeMC/Packwiz-Studio/internal/logbus"
)

// Runner executes commands and reports them to a bus.
type Runner struct {
	bus *logbus.Bus
}

// New returns a Runner publishing to bus. A nil bus is allowed: output is
// then only accumulated into the Result.
func New(bus *logbus.Bus) *Runner { return &Runner{bus: bus} }

// Spec describes one invocation.
type Spec struct {
	// Name is the binary: an absolute path, or a name found on PATH.
	Name string
	// Args are passed to the binary unquoted.
	Args []string
	// Dir is the working directory. For pack operations this is the pack
	// folder; packwiz and git both need it set.
	Dir string
	// Env, when non-nil, replaces the process environment entirely.
	Env []string
	// Stdin, when non-empty, is written to the command's standard input.
	// packwiz prompts on stdin for some operations; supplying answers
	// here keeps the GUI from appearing to hang on an invisible prompt.
	Stdin string
}

// String renders the invocation the way the log drawer shows it.
func (s Spec) String() string {
	parts := make([]string, 0, len(s.Args)+1)
	parts = append(parts, shellQuote(s.Name))
	for _, a := range s.Args {
		parts = append(parts, shellQuote(a))
	}
	return strings.Join(parts, " ")
}

// Result is the outcome of a finished command.
type Result struct {
	// Stdout and Stderr hold the full captured streams, for parsing.
	Stdout string
	Stderr string
	// ExitCode is the process exit status. It is -1 when the command
	// could not be started or was killed by a signal.
	ExitCode int
}

// OK reports whether the command exited successfully.
func (r Result) OK() bool { return r.ExitCode == 0 }

// Output returns stdout, falling back to stderr when stdout is empty.
// packwiz is inconsistent about which stream it writes to, so callers
// that just want "what did it say" use this.
func (r Result) Output() string {
	if s := strings.TrimSpace(r.Stdout); s != "" {
		return s
	}
	return strings.TrimSpace(r.Stderr)
}

// Run executes spec to completion. The returned error is non-nil only
// when the command could not be started or the context ended; a non-zero
// exit is reported through Result.ExitCode, not as an error, because a
// failing packwiz run is an expected outcome the UI must display rather
// than an exceptional one.
func (r *Runner) Run(ctx context.Context, spec Spec) (Result, error) {
	cmd := exec.CommandContext(ctx, spec.Name, spec.Args...)
	cmd.Dir = spec.Dir
	cmd.Env = spec.Env
	if spec.Stdin != "" {
		cmd.Stdin = strings.NewReader(spec.Stdin)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return Result{ExitCode: -1}, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return Result{ExitCode: -1}, err
	}

	r.publish(logbus.KindCommand, spec.String())

	if err := cmd.Start(); err != nil {
		r.result("could not start "+spec.Name+": "+err.Error(), true)
		return Result{ExitCode: -1}, err
	}

	var (
		wg             sync.WaitGroup
		outBuf, errBuf strings.Builder
	)
	wg.Add(2)
	go func() { defer wg.Done(); r.stream(stdout, &outBuf, logbus.KindStdout) }()
	go func() { defer wg.Done(); r.stream(stderr, &errBuf, logbus.KindStderr) }()
	wg.Wait()

	waitErr := cmd.Wait()
	res := Result{Stdout: outBuf.String(), Stderr: errBuf.String()}

	var exitErr *exec.ExitError
	switch {
	case waitErr == nil:
		res.ExitCode = 0
		r.result("exit 0", false)
	case errors.As(waitErr, &exitErr):
		res.ExitCode = exitErr.ExitCode()
		r.result("exit "+strconv.Itoa(res.ExitCode), true)
	default:
		res.ExitCode = -1
		r.result(spec.Name+" failed: "+waitErr.Error(), true)
		return res, waitErr
	}

	if ctxErr := ctx.Err(); ctxErr != nil {
		return res, ctxErr
	}
	return res, nil
}

// stream copies a pipe line by line into buf and onto the bus, so output
// appears in the drawer while the command is still running.
func (r *Runner) stream(pipe io.Reader, buf *strings.Builder, kind logbus.Kind) {
	sc := bufio.NewScanner(pipe)
	// packwiz prints progress lines that can exceed bufio's default 64KB
	// limit when a pack has many mods.
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for sc.Scan() {
		line := sc.Text()
		buf.WriteString(line)
		buf.WriteByte('\n')
		r.publish(kind, line)
	}
	if err := sc.Err(); err != nil {
		r.publish(logbus.KindStderr, "output truncated: "+err.Error())
	}
}

func (r *Runner) publish(kind logbus.Kind, text string) {
	if r.bus != nil {
		r.bus.Publish(kind, text)
	}
}

func (r *Runner) result(text string, failed bool) {
	if r.bus != nil {
		r.bus.PublishResult(text, failed)
	}
}

// shellQuote makes an argument readable in the log. This is for display
// only. Arguments are passed to exec directly, never through a shell.
func shellQuote(s string) string {
	if s == "" {
		return `""`
	}
	if !strings.ContainsAny(s, " \t\"'\\$`") {
		return s
	}
	return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
}
