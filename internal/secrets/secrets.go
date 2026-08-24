// Package secrets stores release API tokens in the OS keyring: Keychain
// on macOS, Secret Service on Linux, Credential Manager on Windows.
//
// Tokens never touch the config file and are never logged. The log bus
// can be told to mask a token as a backstop, which is what Protect on the
// bus is for.
package secrets

import (
	"errors"
	"strings"

	"github.com/zalando/go-keyring"
)

// service is the keyring entry name the app stores tokens under.
const service = "Packwiz Studio"

// ErrNotFound reports a host with no stored token.
var ErrNotFound = errors.New("no token stored for this host")

// ErrUnavailable reports a keyring the OS would not open. On a headless
// Linux session there may be no Secret Service running at all, which is a
// situation to report rather than crash on.
var ErrUnavailable = errors.New("the system keyring is not available")

// Store reads and writes tokens.
type Store struct{}

// New returns a keyring backed store.
func New() *Store { return &Store{} }

// Get reads the token for a host, for example github.com.
func (s *Store) Get(host string) (string, error) {
	host = normalize(host)
	if host == "" {
		return "", ErrNotFound
	}

	token, err := keyring.Get(service, host)
	switch {
	case errors.Is(err, keyring.ErrNotFound):
		return "", ErrNotFound
	case err != nil:
		return "", unavailable(err)
	}
	return token, nil
}

// Set stores the token for a host. An empty token deletes the entry, so
// clearing a field in settings removes the secret rather than storing
// nothing under a live key.
func (s *Store) Set(host, token string) error {
	host = normalize(host)
	if host == "" {
		return ErrNotFound
	}

	if strings.TrimSpace(token) == "" {
		return s.Delete(host)
	}
	if err := keyring.Set(service, host, token); err != nil {
		return unavailable(err)
	}
	return nil
}

// Delete removes a host's token. A host with no token is not an error.
func (s *Store) Delete(host string) error {
	host = normalize(host)
	if host == "" {
		return nil
	}

	err := keyring.Delete(service, host)
	if err == nil || errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return unavailable(err)
}

// Has reports whether a token is stored, without returning it. Settings
// uses this to show that a token exists without reading the secret to
// display it.
func (s *Store) Has(host string) bool {
	token, err := s.Get(host)
	return err == nil && token != ""
}

// normalize lowercases a host so github.com and GitHub.com share one
// entry.
func normalize(host string) string {
	return strings.ToLower(strings.TrimSpace(host))
}

// unavailable wraps a keyring failure, keeping the underlying reason but
// making the common cause clear.
func unavailable(err error) error {
	return errors.Join(ErrUnavailable, err)
}
