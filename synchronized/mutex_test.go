package synchronized

import (
	"sync"
	"testing"
)

func TestWithMutexSerialisesConcurrentActions(t *testing.T) {
	lock := NewMutex()()
	counter := 0
	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			WithMutexUnit(lock, func() { counter++ })()
		}()
	}
	wg.Wait()
	if counter != 64 {
		t.Fatalf("counter = %d, want 64", counter)
	}
}

func TestWithMutexReleasesOnPanic(t *testing.T) {
	lock := NewMutex()()
	func() {
		defer func() { _ = recover() }()
		WithMutex(lock, func() int { panic("boom") })()
	}()
	if got := WithMutex(lock, func() int { return 2 })(); got != 2 {
		t.Fatalf("relock after panic = %d", got)
	}
}
