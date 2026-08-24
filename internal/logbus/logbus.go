// Package logbus fans raw command output out to subscribers and keeps a
// bounded history so the log drawer can be opened after a command has
// already finished.
//
// The bus is UI-agnostic on purpose. Subscribers are called on whichever
// goroutine published the entry, so a subscriber that touches widgets
// must marshal onto the main goroutine itself with fyne.Do.
package logbus

import (
	"strings"
	"sync"
	"time"
)

// Kind classifies one line in the log.
type Kind int

// Entry kinds.
const (
	// KindCommand is the invocation itself, rendered as "$ packwiz …".
	KindCommand Kind = iota
	// KindStdout is a line the command wrote to stdout.
	KindStdout
	// KindStderr is a line the command wrote to stderr.
	KindStderr
	// KindResult is the outcome: an exit code or a failure to start.
	KindResult
	// KindNotice is the app's own commentary, such as "downloading
	// packwiz 0.20.0" — not command output.
	KindNotice
)

// Entry is one line in the log.
type Entry struct {
	Kind Kind
	Text string
	At   time.Time

	// failed marks a KindResult that reported a non-zero exit or a
	// failure to start. Set through PublishResult.
	failed bool
}

// Failed reports whether the entry represents something going wrong, so
// the drawer can colour it with the error token.
func (e Entry) Failed() bool {
	return e.Kind == KindStderr || (e.Kind == KindResult && e.failed)
}

// Bus distributes entries to subscribers and retains recent history. It
// is safe for concurrent use.
type Bus struct {
	mu       sync.Mutex
	capacity int
	history  []Entry
	subs     map[int]func(Entry)
	nextID   int
	secrets  []string
}

// DefaultCapacity is how many entries the history keeps. A long export
// or a large refresh can emit thousands of lines; older ones are dropped
// rather than growing without bound.
const DefaultCapacity = 5000

// New returns a bus retaining up to capacity entries. A capacity of zero
// or less uses DefaultCapacity.
func New(capacity int) *Bus {
	if capacity <= 0 {
		capacity = DefaultCapacity
	}
	return &Bus{
		capacity: capacity,
		subs:     map[int]func(Entry){},
	}
}

// Publish records an entry and hands it to every subscriber.
func (b *Bus) Publish(kind Kind, text string) {
	b.publish(Entry{Kind: kind, Text: text})
}

// PublishResult records a command outcome. A failed result is styled as
// an error by the drawer.
func (b *Bus) PublishResult(text string, failed bool) {
	b.publish(Entry{Kind: KindResult, Text: text, failed: failed})
}

// publish stamps, redacts and retains an entry, then delivers it.
// Subscribers are called outside the lock, so one may publish from its
// own callback without deadlocking.
func (b *Bus) publish(e Entry) {
	e.At = time.Now()

	b.mu.Lock()
	e.Text = redact(e.Text, b.secrets)

	b.history = append(b.history, e)
	if over := len(b.history) - b.capacity; over > 0 {
		b.history = append(b.history[:0], b.history[over:]...)
	}

	subs := make([]func(Entry), 0, len(b.subs))
	for _, fn := range b.subs {
		subs = append(subs, fn)
	}
	b.mu.Unlock()

	for _, fn := range subs {
		fn(e)
	}
}

// Subscribe registers fn for every entry published from now on. The
// returned function unsubscribes and may be called more than once.
func (b *Bus) Subscribe(fn func(Entry)) (cancel func()) {
	b.mu.Lock()
	id := b.nextID
	b.nextID++
	b.subs[id] = fn
	b.mu.Unlock()

	return func() {
		b.mu.Lock()
		delete(b.subs, id)
		b.mu.Unlock()
	}
}

// History returns the retained entries, oldest first.
func (b *Bus) History() []Entry {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]Entry, len(b.history))
	copy(out, b.history)
	return out
}

// Clear drops the history. Subscribers are not notified — the drawer
// clears its own view when the user asks for it.
func (b *Bus) Clear() {
	b.mu.Lock()
	b.history = nil
	b.mu.Unlock()
}

// Protect registers a string that must never appear in the log. Release
// API tokens go to HTTP clients rather than to commands, so this is a
// backstop: if a token ever reaches an argument or an error message it is
// masked instead of written out.
func (b *Bus) Protect(secret string) {
	if secret == "" {
		return
	}
	b.mu.Lock()
	b.secrets = append(b.secrets, secret)
	b.mu.Unlock()
}

// mask replaces a protected string.
const mask = "••••••••"

func redact(text string, secrets []string) string {
	for _, s := range secrets {
		if s != "" && strings.Contains(text, s) {
			text = strings.ReplaceAll(text, s, mask)
		}
	}
	return text
}
