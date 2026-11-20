package aff

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/i-am-the-slime/iris-go/exception"
)

type eitherTag int

// eitherInt and eitherUnit mirror generated `Either Error Int` /
// `Either Error Unit` ADT structs.
type eitherInt struct {
	Tag     eitherTag
	LeftF1  exception.Error
	RightF1 int
}

type eitherUnit struct {
	Tag     eitherTag
	LeftF1  exception.Error
	RightF1 struct{}
}

type fiberInt = struct {
	IsSuspended func() bool
	Join        func(func(eitherInt) func()) func() func()
	Kill        func(exception.Error, func(eitherUnit) func()) func() func()
	OnComplete  func(struct {
		Handler func(func(eitherInt) func()) func()
		Rethrow bool
	}) func() func()
	Run func()
}

type trace struct{ lines []string }

func (t *trace) log(line string) struct{ Run func(any) struct{} } {
	return LiftEffectUnit(func() { t.lines = append(t.lines, line) })
}

func (t *trace) String() string { return strings.Join(t.lines, ",") }

func then[A, B any](first struct{ Run func(any) A }, second struct{ Run func(any) B }) struct{ Run func(any) B } {
	return Bind(first, func(A) struct{ Run func(any) B } { return second })
}

func launch[A any](action struct{ Run func(any) A }) {
	MakeFiber[A, eitherInt, eitherUnit](action)().Run()
}

func launchInt(action struct{ Run func(any) int }) fiberInt {
	fiber := MakeFiber[int, eitherInt, eitherUnit](action)()
	fiber.Run()
	return fiber
}

func delay(ms float64) struct{ Run func(any) struct{} } {
	return Delay(ms)
}

func joinFiber(fiber fiberInt) struct{ Run func(any) int } {
	return MakeAff[int](func(k func(eitherInt) func()) func() func(exception.Error) struct{ Run func(any) struct{} } {
		return func() func(exception.Error) struct{ Run func(any) struct{} } {
			fiber.Join(k)()
			return func(exception.Error) struct{ Run func(any) struct{} } { return Pure(struct{}{}) }
		}
	})
}

func expect(t *testing.T, got *trace, want string) {
	t.Helper()
	if got.String() != want {
		t.Fatalf("trace = %s\nwant    %s", got, want)
	}
}

func TestBindRunsInConstantStack(t *testing.T) {
	var loop func(n int) struct{ Run func(any) int }
	loop = func(n int) struct{ Run func(any) int } {
		if n == 0 {
			return Pure(0)
		}
		return Bind(Pure(n), func(int) struct{ Run func(any) int } { return loop(n - 1) })
	}
	result := 0
	launch(Map(func(n int) struct{} { result = n + 1; return struct{}{} }, loop(1_000_000)))
	Await()
	if result != 1 {
		t.Fatalf("result = %d", result)
	}
}

func TestCatchErrorRecoversThrownErrors(t *testing.T) {
	log := &trace{}
	launch(then(
		CatchError(then(log.log("try"), then(ThrowError[struct{}](exception.Error{Message: "boom"}), log.log("unreached"))),
			func(e exception.Error) struct{ Run func(any) struct{} } { return log.log("caught " + e.Message) }),
		log.log("after")))
	Await()
	expect(t, log, "try,caught boom,after")
}

func TestBracketReleasesOnFailure(t *testing.T) {
	log := &trace{}
	conditions := func() struct {
		Completed func(struct{}) func(int) struct{ Run func(any) struct{} }
		Failed    func(exception.Error) func(int) struct{ Run func(any) struct{} }
		Killed    func(exception.Error) func(int) struct{ Run func(any) struct{} }
	} {
		return struct {
			Completed func(struct{}) func(int) struct{ Run func(any) struct{} }
			Failed    func(exception.Error) func(int) struct{ Run func(any) struct{} }
			Killed    func(exception.Error) func(int) struct{ Run func(any) struct{} }
		}{
			Completed: func(struct{}) func(int) struct{ Run func(any) struct{} } {
				return func(int) struct{ Run func(any) struct{} } { return log.log("completed") }
			},
			Failed: func(e exception.Error) func(int) struct{ Run func(any) struct{} } {
				return func(int) struct{ Run func(any) struct{} } { return log.log("failed " + e.Message) }
			},
			Killed: func(e exception.Error) func(int) struct{ Run func(any) struct{} } {
				return func(int) struct{ Run func(any) struct{} } { return log.log("killed " + e.Message) }
			},
		}
	}
	failing := GeneralBracket(Pure(1), conditions(), func(int) struct{ Run func(any) struct{} } {
		return ThrowError[struct{}](exception.Error{Message: "use"})
	})
	launch(CatchError(failing, func(exception.Error) struct{ Run func(any) struct{} } { return log.log("handled") }))
	Await()
	expect(t, log, "failed use,handled")

	log.lines = nil
	sleeping := GeneralBracket(Pure(1), conditions(), func(int) struct{ Run func(any) struct{} } {
		return then(delay(1000), log.log("woke"))
	})
	victim := launchInt(Map(func(struct{}) int { return 0 }, sleeping))
	launch(then(delay(5), LiftEffectUnit(func() {
		victim.Kill(exception.Error{Message: "stop"}, func(eitherUnit) func() { return func() { log.lines = append(log.lines, "kill done") } })()
	})))
	start := time.Now()
	Await()
	expect(t, log, "killed stop,kill done")
	if time.Since(start) > 500*time.Millisecond {
		t.Fatalf("killed delay still kept the program alive")
	}
}

func TestForkRunsChildUntilItSuspends(t *testing.T) {
	log := &trace{}
	child := then(log.log("child start"), then(delay(20), Map(func(struct{}) int { return 7 }, log.log("child end"))))
	launch(Bind(Fork[int, eitherInt, eitherUnit](true, child), func(fiber fiberInt) struct{ Run func(any) struct{} } {
		return then(log.log("parent"), Bind(joinFiber(fiber), func(n int) struct{ Run func(any) struct{} } {
			return log.log("joined " + string(rune('0'+n)))
		}))
	}))
	Await()
	expect(t, log, "child start,parent,child end,joined 7")
}

func TestDelaysWakeInDeadlineOrder(t *testing.T) {
	log := &trace{}
	launch(then(delay(60), log.log("slow")))
	launch(then(delay(10), log.log("fast")))
	log.lines = append(log.lines, "sync")
	Await()
	expect(t, log, "sync,fast,slow")
}

func TestParApplyRunsOperandsConcurrently(t *testing.T) {
	sleepThen := func(n int) struct{ Run func(any) int } { return Map(func(struct{}) int { return n }, delay(50)) }
	add := Map(func(a int) func(int) int { return func(b int) int { return a + b } }, sleepThen(1))
	result := 0
	start := time.Now()
	launch(Map(func(n int) struct{} { result = n; return struct{}{} }, ParApply(add, sleepThen(2))))
	Await()
	if result != 3 {
		t.Fatalf("result = %d", result)
	}
	if elapsed := time.Since(start); elapsed > 90*time.Millisecond {
		t.Fatalf("parallel delays took %v", elapsed)
	}
}

func TestParApplyFailureKillsTheOtherOperand(t *testing.T) {
	log := &trace{}
	slow := then(delay(1000), Map(func(struct{}) int { return 1 }, log.log("slow finished")))
	failing := then(delay(5), ThrowError[func(int) int](exception.Error{Message: "fast failure"}))
	launch(CatchError(Map(func(int) struct{} { return struct{}{} }, ParApply(failing, slow)),
		func(e exception.Error) struct{ Run func(any) struct{} } { return log.log("caught " + e.Message) }))
	start := time.Now()
	Await()
	expect(t, log, "caught fast failure")
	if time.Since(start) > 500*time.Millisecond {
		t.Fatalf("the losing operand was not killed")
	}
}

func TestParAltTakesTheFirstSuccess(t *testing.T) {
	value := func(n int, ms float64) struct{ Run func(any) int } {
		return Map(func(struct{}) int { return n }, delay(ms))
	}
	result := 0
	launch(Map(func(n int) struct{} { result = n; return struct{}{} }, ParAlt(value(1, 200), value(2, 5))))
	start := time.Now()
	Await()
	if result != 2 || time.Since(start) > 150*time.Millisecond {
		t.Fatalf("parAlt = %d after %v", result, time.Since(start))
	}
}

func TestMakeAffResolvedFromAnotherGoroutine(t *testing.T) {
	result := 0
	native := MakeAff[int](func(k func(eitherInt) func()) func() func(exception.Error) struct{ Run func(any) struct{} } {
		return func() func(exception.Error) struct{ Run func(any) struct{} } {
			done := Hold()
			go func() {
				time.Sleep(5 * time.Millisecond)
				k(eitherInt{Tag: 1, RightF1: 42})()
				done()
			}()
			return func(exception.Error) struct{ Run func(any) struct{} } { return Pure(struct{}{}) }
		}
	})
	launch(Map(func(n int) struct{} { result = n; return struct{}{} }, native))
	Await()
	if result != 42 {
		t.Fatalf("result = %d", result)
	}
}

func BenchmarkBindLoop(b *testing.B) {
	var loop func(n int, acc int) struct{ Run func(any) int }
	loop = func(n int, acc int) struct{ Run func(any) int } {
		if n == 0 {
			return Pure(acc)
		}
		return Bind(LiftEffect(func() int { return n }), func(m int) struct{ Run func(any) int } { return loop(n-1, acc+m) })
	}
	for i := 0; i < b.N; i++ {
		launch(loop(1000, 0))
		Await()
	}
}

func TestTraverseParArrayRunsElementsConcurrentlyInSourceOrder(t *testing.T) {
	log := &trace{}
	worker := func(n int) struct{ Run func(any) int } {
		return then(delay(float64((6-n)*15)), Map(func(struct{}) int { return n * 10 }, log.log(strconv.Itoa(n))))
	}
	var result []int
	start := time.Now()
	launch(Map(func(values []int) struct{} { result = values; return struct{}{} }, TraverseParArray(worker, []int{1, 2, 3, 4, 5})))
	Await()
	expect(t, log, "5,4,3,2,1")
	if fmt.Sprint(result) != "[10 20 30 40 50]" {
		t.Fatalf("result = %v", result)
	}
	if elapsed := time.Since(start); elapsed > 150*time.Millisecond {
		t.Fatalf("parallel traversal took %v", elapsed)
	}
}

type directResult struct {
	failure *exception.Error
	value   int
}

func directFiberFields(f *Fiber) (join func(func(directResult)) func(), run func()) {
	_, join, _, _, run = FiberFields(f,
		func(failure exception.Error) directResult { return directResult{failure: &failure} },
		func(value int) directResult { return directResult{value: value} },
		struct{}{},
		func(struct{}) func(func(directResult)) { return nil })
	return join, run
}

func directJoin(f *Fiber) int {
	join, _ := directFiberFields(f)
	return MakeAffDirect(
		func(result directResult) bool { return result.failure != nil },
		func(result directResult) exception.Error { return *result.failure },
		func(result directResult) int { return result.value },
		func(k func(directResult)) func(exception.Error) {
			join(k)
			return func(exception.Error) {}
		})
}

func TestLaunchDrainsUntilNoFiberCanRun(t *testing.T) {
	log := &trace{}
	parent := NewFiber(func() int {
		log.lines = append(log.lines, "parent")
		child := ForkFiber(true, func() int { return 1 })
		value := directJoin(child)
		log.lines = append(log.lines, "joined "+string(rune('0'+value)))
		return value
	})
	_, run := directFiberFields(parent)
	run()
	log.lines = append(log.lines, "main")
	Await()
	expect(t, log, "parent,joined 1,main")
}

func TestForkInsideADrainRunsTheChildUntilItSuspends(t *testing.T) {
	log := &trace{}
	parent := NewFiber(func() int {
		child := ForkFiber(true, func() int {
			log.lines = append(log.lines, "child")
			value := directJoin(ForkFiber(true, func() int { return 2 }))
			log.lines = append(log.lines, "child joined")
			return value
		})
		log.lines = append(log.lines, "parent")
		return directJoin(child)
	})
	_, run := directFiberFields(parent)
	run()
	log.lines = append(log.lines, "main")
	Await()
	expect(t, log, "child,parent,child joined,main")
}
