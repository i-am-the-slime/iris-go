package avar

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"testing"

	"github.com/i-am-the-slime/iris-go/exception"
)

// step is one operation of a scenario recorded by testdata/generate.js.
type step struct {
	Op       string
	Variable int
	ID       int
	Value    int
	Message  string
	Target   int
	Then     []step
}

// tagged stands in for the PureScript data types the constructors build.
type tagged struct {
	tag   string
	value string
}

func (value tagged) String() string {
	if value.value == "" {
		return value.tag
	}
	return value.tag + " " + value.value
}

func left(failure exception.Error) tagged { return tagged{"Left", failure.Message} }
func just(value int) tagged             { return tagged{"Just", fmt.Sprint(value)} }

type interpreter struct {
	variables []AVar[int]
	cancelers map[int]func()
	log       []string
}

func (in *interpreter) perform(s step) {
	callback := func(value tagged) {
		in.log = append(in.log, fmt.Sprintf("%d %s", s.ID, value))
		for _, nested := range s.Then {
			in.perform(nested)
		}
	}
	right := func(value int) tagged { return tagged{"Right", fmt.Sprint(value)} }
	avar := in.variables[s.Variable]
	switch s.Op {
	case "put":
		unit := func(struct{}) tagged { return tagged{tag: "Right"} }
		in.cancelers[s.ID] = PutVar(left, unit, s.Value, avar, callback)
	case "take":
		in.cancelers[s.ID] = TakeVar(left, right, avar, callback)
	case "read":
		in.cancelers[s.ID] = ReadVar(left, right, avar, callback)
	case "cancel":
		if cancel, ok := in.cancelers[s.Target]; ok {
			cancel()
		}
		in.log = append(in.log, fmt.Sprintf("cancel %d", s.Target))
	case "kill":
		KillVar(exception.Error{Message: s.Message}, avar)
		in.log = append(in.log, fmt.Sprintf("kill %d", s.Variable))
	case "tryPut":
		in.log = append(in.log, fmt.Sprintf("tryPut %d %t", s.Variable, TryPutVar(s.Value, avar)))
	case "tryTake":
		in.log = append(in.log, fmt.Sprintf("tryTake %d %s", s.Variable, TryTakeVar(just, tagged{tag: "Nothing"}, avar)))
	case "tryRead":
		in.log = append(in.log, fmt.Sprintf("tryRead %d %s", s.Variable, TryReadVar(just, tagged{tag: "Nothing"}, avar)))
	case "status":
		killed := func(failure exception.Error) tagged { return tagged{"Killed", failure.Message} }
		filled := func(value int) tagged { return tagged{"Filled", fmt.Sprint(value)} }
		in.log = append(in.log, fmt.Sprintf("status %d %s", s.Variable, Status(killed, filled, tagged{tag: "Empty"}, avar)))
	}
}

func TestScenariosMatchAVarJs(t *testing.T) {
	contents, err := os.ReadFile("testdata/scenarios.json")
	if err != nil {
		t.Fatal(err)
	}
	var scenarios []struct {
		Steps []step
		Log   []string
	}
	if err := json.Unmarshal(contents, &scenarios); err != nil {
		t.Fatal(err)
	}
	for index, scenario := range scenarios {
		in := &interpreter{
			variables: []AVar[int]{Empty[int](), NewVar(1000), Empty[int]()},
			cancelers: map[int]func(){},
		}
		for _, s := range scenario.Steps {
			in.perform(s)
		}
		if !slices.Equal(in.log, scenario.Log) {
			t.Errorf("scenario %d:\n go %q\n js %q", index, in.log, scenario.Log)
		}
	}
}

func TestReadsQueuedDuringADrainWaitForTheNextValue(t *testing.T) {
	avar := Empty[int]()
	var log []string
	note := func(label string) func(tagged) {
		return func(value tagged) { log = append(log, label+" "+value.String()) }
	}
	right := func(value int) tagged { return tagged{"Right", fmt.Sprint(value)} }
	ReadVar(left, right, avar, func(value tagged) {
		note("first read")(value)
		ReadVar(left, right, avar, note("second read"))
	})
	TakeVar(left, right, avar, note("take"))
	PutVar(left, func(struct{}) tagged { return tagged{tag: "Right"} }, 1, avar, note("put"))
	want := []string{"first read Right 1", "take Right 1", "put Right"}
	if !slices.Equal(log, want) {
		t.Errorf("log %q, want %q", log, want)
	}
}
