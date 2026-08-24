package git

import (
	"context"
	"strconv"
	"strings"
)

// Status is a snapshot of the repository, as the status bar shows it.
type Status struct {
	// IsRepo is false when the pack folder is not under git at all. The
	// rest of the fields are then meaningless.
	IsRepo bool
	// Branch is the current branch. It is empty when HEAD is detached.
	Branch string
	// Detached reports a HEAD that is not on a branch.
	Detached bool
	// Changed is how many paths differ from HEAD, staged or not.
	Changed int
	// Remote is the origin URL, empty when there is no origin.
	Remote string
	// Ahead and Behind count commits relative to the upstream branch.
	// Both are zero when there is no upstream to compare against.
	Ahead  int
	Behind int
}

// Clean reports a work tree with nothing to commit.
func (s Status) Clean() bool { return s.Changed == 0 }

// Label renders the branch and change count for the status bar.
func (s Status) Label() string {
	if !s.IsRepo {
		return "not a git repository"
	}

	name := s.Branch
	if s.Detached {
		name = "detached head"
	}
	if name == "" {
		name = "no branch"
	}

	switch s.Changed {
	case 0:
		return name + ", clean"
	case 1:
		return name + ", 1 change"
	default:
		return name + ", " + strconv.Itoa(s.Changed) + " changes"
	}
}

// Read collects the repository state. Every command it runs is a probe,
// so nothing here reaches the log drawer.
func (r *Repo) Read(ctx context.Context) Status {
	var s Status

	if !r.IsRepo(ctx) {
		return s
	}
	s.IsRepo = true

	if res, err := r.probe(ctx, "branch", "--show-current"); err == nil && res.OK() {
		s.Branch = strings.TrimSpace(res.Stdout)
	}
	// An empty branch name on a repository with commits means a detached
	// HEAD. On a brand new repository with no commits it means neither,
	// so the presence of HEAD is what separates the two.
	if s.Branch == "" {
		if res, err := r.probe(ctx, "rev-parse", "--verify", "HEAD"); err == nil && res.OK() {
			s.Detached = true
		}
	}

	if res, err := r.probe(ctx, "status", "--porcelain"); err == nil && res.OK() {
		s.Changed = countLines(res.Stdout)
	}

	if res, err := r.probe(ctx, "remote", "get-url", "origin"); err == nil && res.OK() {
		s.Remote = strings.TrimSpace(res.Stdout)
	}

	s.Ahead, s.Behind = r.tracking(ctx)
	return s
}

// tracking counts commits ahead of and behind the upstream branch. A
// branch with no upstream reports zero for both rather than an error:
// having no upstream is normal, not a failure.
func (r *Repo) tracking(ctx context.Context) (ahead, behind int) {
	res, err := r.probe(ctx, "rev-list", "--left-right", "--count", "@{upstream}...HEAD")
	if err != nil || !res.OK() {
		return 0, 0
	}

	fields := strings.Fields(strings.TrimSpace(res.Stdout))
	if len(fields) != 2 {
		return 0, 0
	}
	behind, _ = strconv.Atoi(fields[0])
	ahead, _ = strconv.Atoi(fields[1])
	return ahead, behind
}

// countLines counts non-empty lines, which for porcelain output is one
// per changed path.
func countLines(out string) int {
	n := 0
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) != "" {
			n++
		}
	}
	return n
}
