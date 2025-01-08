// package nonce
package nonce

import (
	"sync"
	"time"
)

// Store is the interface for nonce storage.
type Store interface {
	Check(nonce, endpoint string) bool
	Store(nonce, endpoint string, expiresAt time.Time)
}

// Entry is a nonce entry.
type Entry struct {
	Endpoint  string
	ExpiresAt time.Time
}

// MemoryStore is an in-memory nonce store.
type MemoryStore struct {
	store map[string]Entry
	mu    sync.RWMutex
}

// NewMemoryStore creates a new MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		store: make(map[string]Entry),
	}
}

// Check checks if the nonce exists in the store and is valid.
func (m *MemoryStore) Check(nonce, endpoint string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, exists := m.store[nonce]
	if !exists {
		return false
	}
	if time.Now().After(entry.ExpiresAt) {
		delete(m.store, nonce)
		return false
	}
	return entry.Endpoint == endpoint
}

// Store stores a nonce in the store.
func (m *MemoryStore) Store(nonce, endpoint string, expiresAt time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store[nonce] = Entry{
		Endpoint:  endpoint,
		ExpiresAt: expiresAt,
	}
}
