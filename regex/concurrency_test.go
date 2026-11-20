package regex

import (
	"sync"
	"testing"
)

func TestGoroutinesShareAGlobalRegex(t *testing.T) {
	rx, err := Compile("a", "g")
	if err != nil {
		t.Fatal(err)
	}
	var group sync.WaitGroup
	for range 8 {
		group.Go(func() {
			for range 100 {
				if got := Replace(rx, "b", "banana"); got != "bbnbnb" {
					t.Errorf("Replace = %q, want bbnbnb", got)
				}
			}
		})
	}
	group.Wait()
}

func TestGoroutinesShareAStickyRegexOneOperationAtATime(t *testing.T) {
	rx, err := Compile("a", "y")
	if err != nil {
		t.Fatal(err)
	}
	var group sync.WaitGroup
	for range 8 {
		group.Go(func() {
			for range 100 {
				Replace(rx, "b", "aaaa")
				Test(rx, "aaaa")
			}
		})
	}
	group.Wait()
	if rx.lastIndex < 0 || rx.lastIndex > 4 {
		t.Fatalf("lastIndex = %d, want a position in aaaa", rx.lastIndex)
	}
}
