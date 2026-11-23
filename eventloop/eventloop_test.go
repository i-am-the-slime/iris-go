package eventloop

import (
	"sync"
	"testing"
	"time"
)

func TestTurnsRunInParallelWhenNothingSerializesThem(t *testing.T) {
	arrived := make(chan struct{}, 2)
	var both sync.WaitGroup
	for range 2 {
		both.Go(func() {
			Turn(func() {
				arrived <- struct{}{}
				for len(arrived) < 2 {
					time.Sleep(time.Millisecond)
				}
			})
		})
	}
	finished := make(chan struct{})
	go func() {
		both.Wait()
		close(finished)
	}()
	select {
	case <-finished:
	case <-time.After(5 * time.Second):
		t.Fatal("each turn waited for the other")
	}
}

func TestTurnRethrowsOnTheCaller(t *testing.T) {
	defer func() {
		if recovered := recover(); recovered != "boom" {
			t.Fatalf("recovered %v, want boom", recovered)
		}
	}()
	Turn(func() { panic("boom") })
}

func TestAwaitWaitsForHeldEvents(t *testing.T) {
	retire := Hold()
	var retired bool
	var lock sync.Mutex
	go func() {
		time.Sleep(10 * time.Millisecond)
		lock.Lock()
		retired = true
		lock.Unlock()
		retire()
	}()
	Await()
	lock.Lock()
	defer lock.Unlock()
	if !retired {
		t.Fatal("Await returned before the held event was retired")
	}
}
