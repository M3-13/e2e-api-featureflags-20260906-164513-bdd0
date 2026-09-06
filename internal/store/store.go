package store

import (
	"errors"
	"sync"
)

// Flag ist die Datenstruktur für ein Feature-Flag im In-Memory-Store.
type Flag struct {
	Key            string `json:"key"`
	Enabled        bool   `json:"enabled"`
	Description    string `json:"description"`
	RolloutPercent int    `json:"rollout_percent"`
}

var (
	// ErrDuplicate wird von Create zurückgegeben, wenn der key bereits existiert.
	ErrDuplicate = errors.New("flag already exists")
	// ErrLimit wird von Create zurückgegeben, wenn das Flag-Limit erreicht ist.
	ErrLimit = errors.New("flag limit reached")
)

// Store hält Feature-Flags thread-sicher im Speicher.
type Store struct {
	mu       sync.RWMutex
	flags    map[string]Flag
	maxFlags int
}

// NewStore erzeugt einen leeren Store mit einem Maximalwert gleichzeitig
// gespeicherter Flags.
func NewStore(maxFlags int) *Store {
	return &Store{
		flags:    make(map[string]Flag),
		maxFlags: maxFlags,
	}
}

// Create legt ein Flag an. Liefert ErrDuplicate bei vorhandenem key und
// ErrLimit, sobald maxFlags erreicht sind.
func (s *Store) Create(f Flag) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.flags[f.Key]; ok {
		return ErrDuplicate
	}
	if len(s.flags) >= s.maxFlags {
		return ErrLimit
	}
	s.flags[f.Key] = f
	return nil
}

// Get liefert das Flag zum key und ob es existiert.
func (s *Store) Get(key string) (Flag, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	f, ok := s.flags[key]
	return f, ok
}

// List liefert alle gespeicherten Flags als Slice.
func (s *Store) List() []Flag {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Flag, 0, len(s.flags))
	for _, f := range s.flags {
		out = append(out, f)
	}
	return out
}

// Update ersetzt das Flag zum key. Liefert false, wenn der key nicht existiert.
func (s *Store) Update(key string, f Flag) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.flags[key]; !ok {
		return false
	}
	f.Key = key
	s.flags[key] = f
	return true
}

// Delete entfernt das Flag zum key. Liefert false, wenn der key nicht existiert.
func (s *Store) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.flags[key]; !ok {
		return false
	}
	delete(s.flags, key)
	return true
}
