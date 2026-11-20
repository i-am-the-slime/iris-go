package array

import (
	"runtime"
	"sync"
)

// ArrayLength implements Data.Array.length.
func ArrayLength[A any](xs []A) int {
	return len(xs)
}

// parallelMapMinimum is the smallest input worth spreading across goroutines;
// below it the scheduling overhead outweighs any pure work per element.
const parallelMapMinimum = 64

// ParallelMap implements Data.Array.Parallel.parallelMap: the mapped function
// is pure, so contiguous chunks run on separate goroutines and the result
// keeps the input order. A panic (a thrown Error) is re-raised on the
// caller's goroutine after every chunk has stopped; when several elements
// throw, it is the first one's, as in the sequential JavaScript map.
func ParallelMap[A, B any](f func(A) B, xs []A) []B {
	out := make([]B, len(xs))
	workers := runtime.GOMAXPROCS(0)
	if len(xs) < parallelMapMinimum || workers < 2 {
		for i, x := range xs {
			out[i] = f(x)
		}
		return out
	}
	chunk := (len(xs) + workers - 1) / workers
	failures := make([]any, (len(xs)+chunk-1)/chunk)
	var wg sync.WaitGroup
	for start := 0; start < len(xs); start += chunk {
		end := min(start+chunk, len(xs))
		wg.Go(func() {
			defer func() { failures[start/chunk] = recover() }()
			for i := start; i < end; i++ {
				out[i] = f(xs[i])
			}
		})
	}
	wg.Wait()
	for _, failure := range failures {
		if failure != nil {
			panic(failure)
		}
	}
	return out
}
