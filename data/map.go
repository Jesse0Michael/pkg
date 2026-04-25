package data

import (
	"iter"
	"sync"
)

// Map is a generic, concurrency-safe map backed by sync.RWMutex.
// Compatible with the maps and iter packages via All, Keys, and Values.
type Map[K comparable, V any] struct {
	mu sync.RWMutex
	m  map[K]V
}

// NewMap returns an initialized Map.
func NewMap[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{m: make(map[K]V)}
}

// Get returns the value for a key, or the zero value if absent.
func (m *Map[K, V]) Get(key K) (V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.m[key]
	return v, ok
}

// Set sets the value for a key.
func (m *Map[K, V]) Set(key K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.m[key] = value
}

// Delete removes a key from the map.
func (m *Map[K, V]) Delete(key K) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.m, key)
}

// Len returns the number of entries in the map.
func (m *Map[K, V]) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.m)
}

// Clear removes all entries from the map.
func (m *Map[K, V]) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	clear(m.m)
}

// All returns an iterator over all key-value pairs.
// The read lock is held for the duration of iteration.
func (m *Map[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		m.mu.RLock()
		defer m.mu.RUnlock()
		for k, v := range m.m {
			if !yield(k, v) {
				return
			}
		}
	}
}

// Keys returns an iterator over all keys.
// The read lock is held for the duration of iteration.
func (m *Map[K, V]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		m.mu.RLock()
		defer m.mu.RUnlock()
		for k := range m.m {
			if !yield(k) {
				return
			}
		}
	}
}

// Values returns an iterator over all values.
// The read lock is held for the duration of iteration.
func (m *Map[K, V]) Values() iter.Seq[V] {
	return func(yield func(V) bool) {
		m.mu.RLock()
		defer m.mu.RUnlock()
		for _, v := range m.m {
			if !yield(v) {
				return
			}
		}
	}
}
