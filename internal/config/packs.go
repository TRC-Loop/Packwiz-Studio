package config

import (
	"path/filepath"
	"sort"
	"time"
)

// Packs returns the known packs ordered most recently opened first,
// which is the order the launcher's recents list shows them in.
func (s *Store) Packs() []Pack {
	s.mu.RLock()
	packs := make([]Pack, len(s.cfg.Packs))
	copy(packs, s.cfg.Packs)
	s.mu.RUnlock()

	sort.SliceStable(packs, func(i, j int) bool {
		return packs[i].LastOpened.After(packs[j].LastOpened)
	})
	return packs
}

// Pack looks up one known pack by its directory.
func (s *Store) Pack(path string) (Pack, bool) {
	key := normalize(path)
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.cfg.Packs {
		if normalize(p.Path) == key {
			return p, true
		}
	}
	return Pack{}, false
}

// Touch records a pack as opened now, adding it to the known list if it
// is new and refreshing the cached metadata the launcher displays.
// Existing prefs are preserved.
func (s *Store) Touch(path, name, mcVersion, loader string) error {
	key := normalize(path)
	return s.Update(func(c *Config) {
		for i := range c.Packs {
			if normalize(c.Packs[i].Path) != key {
				continue
			}
			c.Packs[i].Name = name
			c.Packs[i].MCVersion = mcVersion
			c.Packs[i].Loader = loader
			c.Packs[i].LastOpened = time.Now()
			return
		}
		c.Packs = append(c.Packs, Pack{
			Path:       key,
			Name:       name,
			MCVersion:  mcVersion,
			Loader:     loader,
			LastOpened: time.Now(),
			Prefs:      Prefs{}.withDefaults(),
		})
	})
}

// Forget removes a pack from the known list. It does not touch the pack
// on disk. This is "remove from recents", not "delete".
func (s *Store) Forget(path string) error {
	key := normalize(path)
	return s.Update(func(c *Config) {
		kept := c.Packs[:0]
		for _, p := range c.Packs {
			if normalize(p.Path) != key {
				kept = append(kept, p)
			}
		}
		c.Packs = kept
	})
}

// Prefs returns a pack's remembered choices, with defaults filled in for
// a pack that is not yet known.
func (s *Store) Prefs(path string) Prefs {
	p, ok := s.Pack(path)
	if !ok {
		return Prefs{}.withDefaults()
	}
	return p.Prefs.withDefaults()
}

// SetPrefs applies fn to a pack's prefs and persists them. It is a no-op
// for a pack that is not in the known list.
func (s *Store) SetPrefs(path string, fn func(*Prefs)) error {
	key := normalize(path)
	return s.Update(func(c *Config) {
		for i := range c.Packs {
			if normalize(c.Packs[i].Path) != key {
				continue
			}
			prefs := c.Packs[i].Prefs.withDefaults()
			fn(&prefs)
			c.Packs[i].Prefs = prefs
			return
		}
	})
}

// normalize makes pack paths comparable. Paths are cleaned and made
// absolute where possible so the same folder reached two ways resolves to
// one entry. Case is left alone: it matters on Linux, and folding it
// would merge distinct directories there.
func normalize(path string) string {
	if abs, err := filepath.Abs(path); err == nil {
		return filepath.Clean(abs)
	}
	return filepath.Clean(path)
}
