package lazy

import (
	"slices"
	"sync"
	"sync/atomic"
	"testing"
)

func TestEvaluatedForcesToItsValue(t *testing.T) {
	cell := Evaluated(42)
	if got := cell.Force(); got != 42 {
		t.Fatalf("Force() = %d, want 42", got)
	}
}

func TestDeferRunsItsThunkOnce(t *testing.T) {
	runs := 0
	cell := Defer(func() int {
		runs++
		return runs
	})
	first, second := cell.Force(), cell.Force()
	if first != 1 || second != 1 || runs != 1 {
		t.Fatalf("Force() = %d then %d after %d runs, want 1 then 1 after 1 run", first, second, runs)
	}
}

func TestAThunkThatForcesItsOwnCellRunsAgainAndTheOutermostRunIsKeptAsInJavaScript(t *testing.T) {
	runs := 0
	var cell *Lazy[int]
	cell = Defer(func() int {
		runs++
		if runs == 1 {
			return cell.Force() + 10
		}
		return runs
	})
	if got := cell.Force(); got != 12 || runs != 2 {
		t.Fatalf("Force() = %d after %d runs, want the outer run's 12 after 2 runs", got, runs)
	}
	if got := cell.Force(); got != 12 || runs != 2 {
		t.Fatalf("second Force() = %d after %d runs, want 12 after 2 runs", got, runs)
	}
}

func TestAThunkThatThrowsRunsAgainOnEveryForce(t *testing.T) {
	runs := 0
	cell := Defer(func() int {
		runs++
		if runs == 1 {
			panic("first run fails")
		}
		return runs
	})
	func() {
		defer func() { _ = recover() }()
		cell.Force()
	}()
	if got := cell.Force(); got != 2 {
		t.Fatalf("Force() after a throw = %d, want 2", got)
	}
	if got := cell.Force(); got != 3 {
		t.Fatalf("Force() after a throw and a value = %d, want 3", got)
	}
}

func TestGoroutinesForcingOneCellShareOneOfTheirValues(t *testing.T) {
	for range 200 {
		var runs atomic.Int32
		cell := Defer(func() []int {
			return []int{int(runs.Add(1))}
		})
		results := make([][]int, 8)
		var group sync.WaitGroup
		for index := range results {
			group.Go(func() { results[index] = cell.Force() })
		}
		group.Wait()
		published := cell.Force()
		if !slices.ContainsFunc(results, func(result []int) bool { return &result[0] == &published[0] }) {
			t.Fatalf("published %v, which no evaluator returned", published)
		}
		for index, result := range results {
			if result[0] < 1 || int32(result[0]) > runs.Load() {
				t.Fatalf("goroutine %d saw %v after %d runs", index, result, runs.Load())
			}
		}
		if &cell.Force()[0] != &published[0] {
			t.Fatalf("a forced cell returned two different values")
		}
		if cell.thunk != nil {
			t.Fatalf("a forced cell keeps its thunk")
		}
	}
}

func TestAConstantIsComputedByWhatEachForceWithPasses(t *testing.T) {
	cell := Unevaluated[string]()
	runs := 0
	compute := func() string { runs++; return "computed" }
	first, second := cell.ForceWith(compute), cell.ForceWith(compute)
	if first != "computed" || second != "computed" || runs != 1 {
		t.Fatalf("ForceWith = %q then %q after %d runs, want computed twice after 1 run", first, second, runs)
	}
}

func TestGoroutinesForcingOneConstantAgree(t *testing.T) {
	cell := Unevaluated[*int]()
	compute := func() *int { value := 7; return &value }
	results := make([]*int, 8)
	var group sync.WaitGroup
	for index := range results {
		group.Go(func() { results[index] = cell.ForceWith(compute) })
	}
	group.Wait()
	for _, result := range results {
		if *result != 7 {
			t.Fatalf("a goroutine saw %d", *result)
		}
	}
}

func TestAForcedCellDropsItsThunkSoThatAConsumedStreamIsFreedAsInJavaScript(t *testing.T) {
	captured := make([]int, 1000)
	cell := Defer(func() int { return len(captured) })
	if got := cell.Force(); got != 1000 {
		t.Fatalf("Force() = %d, want 1000", got)
	}
	if cell.thunk != nil {
		t.Fatalf("a forced cell keeps its thunk, and with it everything the thunk captured")
	}
}
