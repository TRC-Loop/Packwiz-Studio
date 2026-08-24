// Package studio holds the state every window shares: the config store,
// the log bus, the command runner and the resolved packwiz binary.
//
// One Session exists per app launch. Windows read from it rather than
// threading half a dozen dependencies through every constructor.
package studio

import (
	"context"
	"sync"

	"fyne.io/fyne/v2"

	"github.com/TRC-Loop/Packwiz-Studio/internal/cmdrun"
	"github.com/TRC-Loop/Packwiz-Studio/internal/config"
	"github.com/TRC-Loop/Packwiz-Studio/internal/logbus"
	"github.com/TRC-Loop/Packwiz-Studio/internal/packwiz"
)

// Session is the shared state of a running app.
type Session struct {
	// App is the Fyne app, for opening windows and reading preferences.
	App fyne.App
	// Cfg is the persisted config.
	Cfg *config.Store
	// Bus carries command output to the log drawer.
	Bus *logbus.Bus
	// Runner executes packwiz and git.
	Runner *cmdrun.Runner

	mu   sync.RWMutex
	tool packwiz.Location
	// toolListeners are notified whenever the packwiz binary changes, so
	// a status bar or a disabled button can update itself.
	toolListeners []func(packwiz.Location)
	// configListeners are notified when settings are saved. Turning the
	// git integration off has to reach an already open pack window, whose
	// icon rail was built when the setting still said otherwise.
	configListeners []func()
}

// New returns a Session wired to the given config store.
func New(app fyne.App, cfg *config.Store) *Session {
	bus := logbus.New(0)
	return &Session{
		App:    app,
		Cfg:    cfg,
		Bus:    bus,
		Runner: cmdrun.New(bus),
	}
}

// Packwiz returns the currently resolved binary. A zero Location means
// none has been found yet.
func (s *Session) Packwiz() packwiz.Location {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tool
}

// HasPackwiz reports whether a usable binary is available. Screens gate
// their pack actions on this.
func (s *Session) HasPackwiz() bool {
	return s.Packwiz().Path != ""
}

// SetPackwiz records a resolved binary and persists its path, unless the
// binary was found on PATH: storing that would pin the app to today's
// PATH entry and stop it following a later move or upgrade.
func (s *Session) SetPackwiz(loc packwiz.Location) error {
	s.mu.Lock()
	s.tool = loc
	listeners := make([]func(packwiz.Location), len(s.toolListeners))
	copy(listeners, s.toolListeners)
	s.mu.Unlock()

	for _, fn := range listeners {
		fn(loc)
	}

	if loc.FromPATH {
		return nil
	}
	return s.Cfg.Update(func(c *config.Config) { c.PackwizPath = loc.Path })
}

// OnPackwizChange registers fn for every change to the resolved binary.
// It is called on the goroutine that made the change, so a UI listener
// must hop to the main thread with fyne.Do.
func (s *Session) OnPackwizChange(fn func(packwiz.Location)) {
	s.mu.Lock()
	s.toolListeners = append(s.toolListeners, fn)
	s.mu.Unlock()
}

// OnConfigChange registers fn for every settings save. It is called on
// the goroutine that saved, so a UI listener must hop to the main thread
// with fyne.Do.
func (s *Session) OnConfigChange(fn func()) {
	s.mu.Lock()
	s.configListeners = append(s.configListeners, fn)
	s.mu.Unlock()
}

// ConfigChanged announces that settings were saved.
func (s *Session) ConfigChanged() {
	s.mu.RLock()
	listeners := make([]func(), len(s.configListeners))
	copy(listeners, s.configListeners)
	s.mu.RUnlock()

	for _, fn := range listeners {
		fn()
	}
}

// ResolvePackwiz looks for a usable binary and records it. A failure is
// returned rather than surfaced, so the caller decides whether to show a
// setup screen or a settings error.
func (s *Session) ResolvePackwiz(ctx context.Context) error {
	loc, err := packwiz.Locate(ctx, s.Cfg.Get().PackwizPath)
	if err != nil {
		return err
	}
	return s.SetPackwiz(loc)
}

// Client returns a packwiz client for one pack folder.
func (s *Session) Client(dir string) *packwiz.Client {
	return packwiz.NewClient(s.Packwiz().Path, dir, s.Runner)
}

// GitEnabled reports whether the app's git integration is turned on. When
// it is off the app performs no git commands and hides the Git and
// Releases activities entirely.
func (s *Session) GitEnabled() bool {
	return s.Cfg.Get().GitEnabled
}
