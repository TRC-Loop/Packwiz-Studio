package config

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

// dirName is the app's folder inside the OS user config directory.
const dirName = "PackwizStudio"

// fileName is the config file inside dirName.
const fileName = "config.json"

// Store owns the on-disk config. It is safe for concurrent use: every
// pack window shares one Store.
type Store struct {
	mu   sync.RWMutex
	path string
	cfg  Config
}

// Path reports the config file's location, for display in settings.
func (s *Store) Path() string { return s.path }

// Open loads the config from the OS user config directory, creating a
// default one if none exists. A malformed file is reported rather than
// silently replaced, so a user is never quietly reset.
func Open() (*Store, error) {
	path, err := filePath()
	if err != nil {
		return nil, err
	}
	return openAt(path)
}

func openAt(path string) (*Store, error) {
	s := &Store{path: path, cfg: defaults()}

	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return s, s.save()
	}
	if err != nil {
		return nil, err
	}

	cfg := defaults()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, &MalformedError{Path: path, Err: err}
	}
	s.cfg = cfg
	return s, nil
}

// MalformedError reports a config file that could not be parsed. The UI
// surfaces it with the path so the user can inspect or delete the file.
type MalformedError struct {
	Path string
	Err  error
}

func (e *MalformedError) Error() string {
	return "config file " + e.Path + " is not valid JSON: " + e.Err.Error()
}

func (e *MalformedError) Unwrap() error { return e.Err }

// Get returns a snapshot of the config. Mutating it has no effect on
// stored state; use Update for that.
func (s *Store) Get() Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg.clone()
}

// Update applies fn to the config and persists the result. If saving
// fails the in-memory change is kept, so the UI stays consistent with
// what the user did and the error can be surfaced separately.
func (s *Store) Update(fn func(*Config)) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	fn(&s.cfg)
	return s.save()
}

// save writes the config atomically: a temp file in the same directory
// followed by a rename, so a crash mid-write cannot truncate the config.
// Callers must hold the write lock.
func (s *Store) save() error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s.cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	tmp, err := os.CreateTemp(dir, fileName+".*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), s.path)
}

// filePath resolves the config file location on this OS.
func filePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, dirName, fileName), nil
}
