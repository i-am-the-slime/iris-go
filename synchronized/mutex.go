// Package synchronized implements Effect.Synchronized. A `Mutex` is its lock
// operation, `func() func()` (TypeDocs.mutexGoType): locking returns the
// matching unlock, so the representation needs no named Go type.
package synchronized

import "sync"

// NewMutex implements Effect.Synchronized.newMutex.
func NewMutex() func() func() func() {
	return func() func() func() {
		var mu sync.Mutex
		return func() func() {
			mu.Lock()
			return mu.Unlock
		}
	}
}

// WithMutex implements Effect.Synchronized.withMutex: the action runs while
// the lock is held, and the lock is released even when it throws.
func WithMutex[A any](lock func() func(), action func() A) func() A {
	return func() A {
		unlock := lock()
		defer unlock()
		return action()
	}
}

// WithMutexUnit is WithMutex at `Effect Unit`.
func WithMutexUnit(lock func() func(), action func()) func() {
	return func() {
		unlock := lock()
		defer unlock()
		action()
	}
}
