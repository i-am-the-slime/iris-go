// Foreign implementations for module Main. `foreign import data Counter` and
// `foreign import data Labelled` are the named types declared here.
package Main_ffi

type Counter struct{ count int }

type Labelled[A any] struct {
	text  string
	value A
}

func NewCounter(start int) Counter { return Counter{count: start} }

func Bump(counter Counter) Counter { return Counter{count: counter.count + 1} }

func Count(counter Counter) int { return counter.count }

func Label[A any](text string, value A) Labelled[A] { return Labelled[A]{text: text, value: value} }

func LabelOf[A any](labelled Labelled[A]) string { return labelled.text }

func ValueOf[A any](labelled Labelled[A]) A { return labelled.value }
