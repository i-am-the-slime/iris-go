package aff

import (
	"errors"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/i-am-the-slime/iris-go/direct"
	"github.com/i-am-the-slime/iris-go/eventloop"
	"github.com/i-am-the-slime/iris-go/exception"
)

func TestSuspendingInALoopRunsInConstantStack(t *testing.T) {
	result := 0
	launch(func() struct{} {
		for range 100_000 {
			result += resolvedAtOnce(1)
		}
		return struct{}{}
	})
	eventloop.Await()
	if result != 100_000 {
		t.Fatalf("result = %d", result)
	}
}

func TestCatchErrorRecoversThrownErrors(t *testing.T) {
	log := &trace{}
	launch(func() struct{} {
		direct.CatchError(func() struct{} {
			log.add("try")
			panic(exception.Error{Message: "boom"})
		}, func(e exception.Error) struct{} {
			log.add("caught " + e.Message)
			return struct{}{}
		})
		log.add("after")
		return struct{}{}
	})
	eventloop.Await()
	expect(t, log, "try,caught boom,after")
}

func TestBracketReleasesOnFailure(t *testing.T) {
	log := &trace{}
	launch(func() struct{} {
		return direct.CatchError(func() struct{} {
			return bracket(log, func(int) struct{} { panic(exception.Error{Message: "use"}) })
		}, func(exception.Error) struct{} {
			log.add("handled")
			return struct{}{}
		})
	})
	eventloop.Await()
	expect(t, log, "failed use,handled")
}

func TestBracketReleasesWhenItsFiberIsKilled(t *testing.T) {
	log := &trace{}
	victim := launch(func() struct{} {
		return bracket(log, func(int) struct{} {
			DelayDirect(nil, 1000)
			log.add("woke")
			return struct{}{}
		})
	})
	launch(func() struct{} {
		DelayDirect(nil, 5)
		kill(victim, exception.Error{Message: "stop"}, func() { log.add("kill done") })
		return struct{}{}
	})
	start := time.Now()
	eventloop.Await()
	expect(t, log, "killed stop,kill done")
	if time.Since(start) > 500*time.Millisecond {
		t.Fatalf("killed delay still kept the program alive")
	}
}

func TestForkRunsChildUntilItSuspends(t *testing.T) {
	log := &trace{}
	launch(func() struct{} {
		child := ForkFiber(true, func() int {
			log.add("child start")
			DelayDirect(nil, 20)
			log.add("child end")
			return 7
		})
		log.add("parent")
		log.add("joined " + strconv.Itoa(join(child)))
		return struct{}{}
	})
	eventloop.Await()
	expect(t, log, "child start,parent,child end,joined 7")
}

func TestDelaysWakeInDeadlineOrder(t *testing.T) {
	log := &trace{}
	launch(func() struct{} { DelayDirect(nil, 60); log.add("slow"); return struct{}{} })
	launch(func() struct{} { DelayDirect(nil, 10); log.add("fast"); return struct{}{} })
	log.add("sync")
	eventloop.Await()
	expect(t, log, "sync,fast,slow")
}

func TestParApplyRunsOperandsConcurrently(t *testing.T) {
	result := 0
	start := time.Now()
	launch(func() struct{} {
		add := func() func(int) int {
			a := sleepThen(1, 50)
			return func(b int) int { return a + b }
		}
		result = ParApplyDirect(add, func() int { return sleepThen(2, 50) })
		return struct{}{}
	})
	eventloop.Await()
	if result != 3 {
		t.Fatalf("result = %d", result)
	}
	if elapsed := time.Since(start); elapsed > 90*time.Millisecond {
		t.Fatalf("parallel delays took %v", elapsed)
	}
}

func TestParApplyFailureKillsTheOtherOperand(t *testing.T) {
	log := &trace{}
	slow := func() int {
		DelayDirect(nil, 1000)
		log.add("slow finished")
		return 1
	}
	failing := func() func(int) int {
		DelayDirect(nil, 5)
		panic(exception.Error{Message: "fast failure"})
	}
	launch(func() struct{} {
		return direct.CatchError(func() struct{} {
			ParApplyDirect(failing, slow)
			return struct{}{}
		}, func(e exception.Error) struct{} {
			log.add("caught " + e.Message)
			return struct{}{}
		})
	})
	start := time.Now()
	eventloop.Await()
	expect(t, log, "caught fast failure")
	if time.Since(start) > 500*time.Millisecond {
		t.Fatalf("the losing operand was not killed")
	}
}

func TestParAltTakesTheFirstSuccess(t *testing.T) {
	result := 0
	launch(func() struct{} {
		result = ParAltDirect(func() int { return sleepThen(1, 200) }, func() int { return sleepThen(2, 5) })
		return struct{}{}
	})
	start := time.Now()
	eventloop.Await()
	if result != 2 || time.Since(start) > 150*time.Millisecond {
		t.Fatalf("parAlt = %d after %v", result, time.Since(start))
	}
}

func TestATimedOutBracketOfParallelGoroutineResumptionsReachesTheEndOfItsUse(t *testing.T) {
	for round := range 300 {
		log := &trace{}
		launch(func() struct{} {
			log.add(ParAltDirect(attempted(func() string {
				DelayDirect(nil, 2000)
				panic(exception.Error{Message: "timed out"})
			}), attempted(func() string {
				bracket(log, func(int) struct{} {
					parallelInOrder(append(
						[]func(){func() { log.add("sync") }},
						slices.Repeat([]func(){func() { resolvedOnAnotherGoroutine(1) }}, 20)...))
					resolvedOnAnotherGoroutine(2)
					panic(exception.Error{Message: "end of use"})
				})
				return "passed"
			})))
			return struct{}{}
		})
		eventloop.Await()
		if log.String() != "sync,failed end of use,error end of use" {
			t.Fatalf("round %d: %s", round, log)
		}
	}
}

func TestMakeAffResolvedFromAnotherGoroutine(t *testing.T) {
	result := 0
	launch(func() struct{} {
		result = makeAff(func(k func(outcome)) {
			done := eventloop.Hold()
			go func() {
				time.Sleep(5 * time.Millisecond)
				k(outcome{value: 42})
				done()
			}()
		})
		return struct{}{}
	})
	eventloop.Await()
	if result != 42 {
		t.Fatalf("result = %d", result)
	}
}

func TestMakeAffResolvedFromAGoroutineWhileTheRegistrationReturns(t *testing.T) {
	results := make([]int, 200)
	for i := range results {
		launch(func() struct{} {
			results[i] = makeAff(func(k func(outcome)) {
				done := eventloop.Hold()
				go func() {
					k(outcome{value: i})
					done()
				}()
			})
			return struct{}{}
		})
	}
	eventloop.Await()
	for i, result := range results {
		if result != i {
			t.Fatalf("results[%d] = %d", i, result)
		}
	}
}

func TestMakeAffTakesTheFirstOfResolutionsFromManyGoroutines(t *testing.T) {
	result := -1
	launch(func() struct{} {
		result = makeAff(func(k func(outcome)) {
			done := eventloop.Hold()
			var resolving sync.WaitGroup
			for range 16 {
				resolving.Go(func() { k(outcome{value: 7}) })
			}
			go func() {
				resolving.Wait()
				done()
			}()
		})
		return struct{}{}
	})
	eventloop.Await()
	if result != 7 {
		t.Fatalf("result = %d", result)
	}
}

func TestTurnsFromGoroutinesRunOneAtATime(t *testing.T) {
	count := 0
	launch(func() struct{} {
		makeAff(func(k func(outcome)) {
			done := eventloop.Hold()
			go func() {
				var turns sync.WaitGroup
				for range 32 {
					turns.Go(func() {
						for range 1000 {
							eventloop.Turn(func() { count++ })
						}
					})
				}
				turns.Wait()
				k(outcome{value: 0})
				done()
			}()
		})
		return struct{}{}
	})
	eventloop.Await()
	if count != 32_000 {
		t.Fatalf("count = %d", count)
	}
}

func TestTraverseParArrayRunsElementsConcurrentlyInSourceOrder(t *testing.T) {
	log := &trace{}
	worker := func(n int) int {
		DelayDirect(nil, float64((6-n)*15))
		log.add(strconv.Itoa(n))
		return n * 10
	}
	var result []int
	start := time.Now()
	launch(func() struct{} {
		result = TraverseParArrayDirect(worker, []int{1, 2, 3, 4, 5})
		return struct{}{}
	})
	eventloop.Await()
	expect(t, log, "5,4,3,2,1")
	if !slices.Equal(result, []int{10, 20, 30, 40, 50}) {
		t.Fatalf("result = %v", result)
	}
	if elapsed := time.Since(start); elapsed > 150*time.Millisecond {
		t.Fatalf("parallel traversal took %v", elapsed)
	}
}

func TestLaunchDrainsUntilNoFiberCanRun(t *testing.T) {
	log := &trace{}
	launch(func() int {
		log.add("parent")
		value := join(ForkFiber(true, func() int { return 1 }))
		log.add("joined " + strconv.Itoa(value))
		return value
	})
	log.add("main")
	eventloop.Await()
	expect(t, log, "parent,joined 1,main")
}

func TestForkInsideADrainRunsTheChildUntilItSuspends(t *testing.T) {
	log := &trace{}
	launch(func() int {
		child := ForkFiber(true, func() int {
			log.add("child")
			value := join(ForkFiber(true, func() int { return 2 }))
			log.add("child joined")
			return value
		})
		log.add("parent")
		return join(child)
	})
	log.add("main")
	eventloop.Await()
	expect(t, log, "child,parent,child joined,main")
}

func TestAnUncaughtFiberFailureIsAGoPanicOnALaterTurn(t *testing.T) {
	command := exec.Command(os.Args[0], "-test.run=^$")
	command.Env = append(os.Environ(), "AFF_UNCAUGHT=1")
	var stdout, stderr strings.Builder
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	var exit *exec.ExitError
	if err != nil && !errors.As(err, &exit) {
		t.Fatal(err)
	}
	if stdout.String() != "after launch\n" || !strings.HasPrefix(stderr.String(), "panic: boom\n") || command.ProcessState.ExitCode() != 2 {
		t.Fatalf("stdout %q, stderr %q, status %d", stdout.String(), stderr.String(), command.ProcessState.ExitCode())
	}
}

func TestMain(m *testing.M) {
	if os.Getenv("AFF_UNCAUGHT") != "" {
		launch(func() struct{} { panic(exception.Error{Message: "boom"}) })
		os.Stdout.WriteString("after launch\n")
		eventloop.Await()
		os.Stdout.WriteString("unreached\n")
	}
	os.Exit(m.Run())
}

func BenchmarkForkAndJoin(b *testing.B) {
	for b.Loop() {
		launch(func() int {
			sum := 0
			for n := range 1000 {
				sum += join(ForkFiber(true, func() int { return n }))
			}
			return sum
		})
		eventloop.Await()
	}
}

type trace struct{ lines []string }

func (t *trace) add(line string) { t.lines = append(t.lines, line) }

func (t *trace) String() string { return strings.Join(t.lines, ",") }

func expect(t *testing.T, got *trace, want string) {
	t.Helper()
	if got.String() != want {
		t.Fatalf("trace = %s\nwant    %s", got, want)
	}
}

// launch runs an action as a new fiber, as Effect.Aff.launchAff does.
func launch[A any](action func() A) *Fiber {
	f := NewFiber(action)
	_, _, _, _, run := fields[A](f)
	run()
	return f
}

// join waits for a fiber's result and rethrows its failure, as Effect.Aff.joinFiber does.
func join(f *Fiber) int {
	_, joined, _, _, _ := fields[int](f)
	return makeAff(func(k func(outcome)) { joined(k) })
}

// kill interrupts a fiber, and `done` runs once it has unwound, as Effect.Aff.killFiber does.
func kill(f *Fiber, failure exception.Error, done func()) {
	_, _, killed, _, _ := fields[struct{}](f)
	killed(failure)(func(struct{}) { done() })
}

// makeAff suspends the running fiber until `register`'s callback delivers an outcome, which
// stands for the `Either Error Int` of generated code.
func makeAff(register func(func(outcome))) int {
	return MakeAffDirect(
		func(result outcome) bool { return result.failure != nil },
		func(result outcome) exception.Error { return *result.failure },
		func(result outcome) int { return result.value.(int) },
		func(k func(outcome)) func(exception.Error) {
			register(k)
			return func(exception.Error) {}
		})
}

// resolvedAtOnce is an Aff whose registration calls its callback before it returns.
func resolvedAtOnce(value int) int {
	return makeAff(func(k func(outcome)) { k(outcome{value: value}) })
}

func fields[A any](f *Fiber) (
	func() bool,
	func(func(outcome)) func(),
	func(exception.Error) func(func(struct{})) func(),
	func(struct{}) func(),
	func(),
) {
	return FiberFields(f,
		func(failure exception.Error) outcome { return outcome{failure: &failure} },
		func(value A) outcome { return outcome{value: value} },
		struct{}{},
		func(struct{}) func(func(outcome)) { return func(func(outcome)) {} })
}

// attempted is `try` for an Aff whose result shows as a string.
func attempted(action func() string) func() string {
	return func() string {
		return direct.CatchError(action, func(e exception.Error) string { return "error " + e.Message })
	}
}

// parallelInOrder is `parSequence_` as `foldr (*>)` over ParAff applications.
func parallelInOrder(actions []func()) {
	if len(actions) == 0 {
		return
	}
	ParApplyDirect(func() func(struct{}) struct{} {
		actions[0]()
		return func(struct{}) struct{} { return struct{}{} }
	}, func() struct{} {
		parallelInOrder(actions[1:])
		return struct{}{}
	})
}

// resolvedOnAnotherGoroutine is an Aff that a goroutine of its own resolves after taking a turn,
// as a Go HTTP client resolves its request while the server's handlers take turns.
func resolvedOnAnotherGoroutine(value int) int {
	return makeAff(func(k func(outcome)) {
		done := eventloop.Hold()
		go func() {
			eventloop.Turn(func() {})
			k(outcome{value: value})
			done()
		}()
	})
}

func sleepThen(value int, milliseconds float64) int {
	DelayDirect(nil, milliseconds)
	return value
}

// bracket acquires 1, uses it, and logs the release that runs.
func bracket(log *trace, use func(int) struct{}) struct{} {
	release := func(line string) func(int) {
		return func(int) { log.add(line) }
	}
	return GeneralBracketDirect(
		func() int { return 1 },
		func(struct{}) func(int) { return release("completed") },
		func(e exception.Error) func(int) { return release("failed " + e.Message) },
		func(e exception.Error) func(int) { return release("killed " + e.Message) },
		use)
}
