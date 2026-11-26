package Main_ffi

import (
	"slices"
	"sync"
	"time"

	"github.com/i-am-the-slime/iris-go/eventloop"
	"github.com/i-am-the-slime/iris-go/exception"
)

func OnTheEventLoop[A any](effect func() A) A {
	var value A
	eventloop.Turn(func() { value = effect() })
	return value
}

func ForEach(values []int, each func(int)) {
	for _, value := range values {
		each(value)
	}
}

func SortWith(compare func(int) func(int) int, values []int) []int {
	sorted := slices.Clone(values)
	slices.SortFunc(sorted, func(a, b int) int { return compare(a)(b) })
	return sorted
}

func FanOut(
	workers int,
	calls int,
	handler func(),
	_ func(exception.Error),
	succeeded func(struct{}),
) func(exception.Error) func(func(exception.Error)) func(func(struct{})) {
	retire := eventloop.Hold()
	go func() {
		var running sync.WaitGroup
		for range workers {
			running.Go(func() {
				for range calls {
					handler()
				}
			})
		}
		running.Wait()
		succeeded(struct{}{})
		retire()
	}()
	return cancel
}

func Later(
	milliseconds float64,
	message string,
	_ func(exception.Error),
	succeeded func(string),
) func(exception.Error) func(func(exception.Error)) func(func(struct{})) {
	retire := eventloop.Hold()
	go func() {
		time.Sleep(time.Duration(milliseconds * float64(time.Millisecond)))
		succeeded(message)
		retire()
	}()
	return cancel
}

func cancel(exception.Error) func(func(exception.Error)) func(func(struct{})) {
	return func(func(exception.Error)) func(func(struct{})) {
		return func(succeeded func(struct{})) { succeeded(struct{}{}) }
	}
}
