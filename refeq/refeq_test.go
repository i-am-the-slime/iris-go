package refeq

import (
	"math"
	"testing"
)

// The tests pin the runtime-layout facts the helpers rely on: a func value is
// one pointer to its closure, a slice is identified by its data pointer and
// length, and an interface is a type word and a data word.

func double(n int32) int32 { return n * 2 }

type holder struct {
	callback func(int32) int32
}

func adder(step int32) func(int32) int32 {
	return func(n int32) int32 { return n + step }
}

func pass(f func(int32) int32) func(int32) int32 { return f }

func TestAClosureIsItselfAfterBeingPassedOrStored(t *testing.T) {
	closure := adder(1)
	stored := holder{callback: closure}
	slice := []func(int32) int32{closure}
	if !SameFunc(closure, pass(closure)) {
		t.Fatal("a closure passed through a function is not itself")
	}
	if !SameFunc(closure, stored.callback) {
		t.Fatal("a closure stored in a struct is not itself")
	}
	if !SameFunc(slice[0], closure) {
		t.Fatal("a closure stored in a slice is not itself")
	}
}

func TestDistinctClosuresOfOneLiteralAreDifferent(t *testing.T) {
	if SameFunc(adder(1), adder(1)) {
		t.Fatal("two closures of one literal capturing equal values are the same")
	}
	if SameFunc(adder(1), adder(2)) {
		t.Fatal("two closures of one literal capturing different values are the same")
	}
}

func TestATopLevelFunctionIsItself(t *testing.T) {
	first, second := double, double
	if !SameFunc(first, second) {
		t.Fatal("a top-level function is not itself")
	}
	if SameFunc(double, adder(0)) {
		t.Fatal("a top-level function is the same as a closure")
	}
}

func TestASliceIsItselfAndNotACopy(t *testing.T) {
	array := []int32{1, 2, 3}
	alias := array
	if !SameSlice(array, alias) {
		t.Fatal("a slice is not itself")
	}
	if SameSlice(array, append([]int32(nil), array...)) {
		t.Fatal("a copy with the same contents is the same slice")
	}
	if SameSlice(array, array[:2]) {
		t.Fatal("a shorter view of the same array is the same slice")
	}
	if SameSlice(array, array[1:]) {
		t.Fatal("a view at another start is the same slice")
	}
}

func TestAnyTwoEmptySlicesAreTheSame(t *testing.T) {
	literal, made, empty := []int32{}, make([]int32, 0), []int32(nil)
	if !SameSlice(literal, made) || !SameSlice(literal, empty) || !SameSlice(made, empty) {
		t.Fatal("empty slices are told apart")
	}
}

type named []int32

func TestANamedSliceTypeCompares(t *testing.T) {
	values := named{1}
	if !SameSlice(values, values) || SameSlice(values, named{1}) {
		t.Fatal("a named slice type does not compare by identity")
	}
}

func TestAnyComparesPointerShapedValuesByIdentity(t *testing.T) {
	pointer, other := new(int32), new(int32)
	closure := adder(1)
	if !SameAny(pointer, pointer) || SameAny(pointer, other) {
		t.Fatal("pointers at any do not compare by identity")
	}
	if !SameAny(closure, closure) || SameAny(closure, adder(1)) {
		t.Fatal("functions at any do not compare by identity")
	}
	if !SameAny(any(nil), any(nil)) || SameAny(any(nil), pointer) {
		t.Fatal("nil at any is not only itself")
	}
}

func TestAnyComparesScalarsByValue(t *testing.T) {
	if !SameAny(int32(7), int32(7)) || SameAny(int32(7), int32(8)) {
		t.Fatal("integers at any do not compare by value")
	}
	if !SameAny("text", "te"+"xt") {
		t.Fatal("strings at any do not compare by value")
	}
	if SameAny(math.NaN(), math.NaN()) {
		t.Fatal("NaN at any is equal to itself")
	}
}

func TestAnyTellsDynamicTypesApart(t *testing.T) {
	if SameAny(int32(1), int64(1)) {
		t.Fatal("values of different dynamic types are the same")
	}
}

func TestAnyComparesASliceByItsBoxing(t *testing.T) {
	array := []int32{1, 2}
	boxed := any(array)
	if !SameAny(boxed, boxed) {
		t.Fatal("a boxed slice is not itself")
	}
	if SameAny(any(array), any(array)) {
		t.Fatal("the same slice boxed twice is the same")
	}
}

type labelled struct {
	label    string
	callback func(int32) int32
}

func TestAnyComparesAStructWithAFunctionByItsBoxing(t *testing.T) {
	value := labelled{label: "double", callback: double}
	boxed := any(value)
	if !SameAny(boxed, boxed) {
		t.Fatal("a boxed struct is not itself")
	}
	if SameAny(any(value), any(value)) {
		t.Fatal("a struct holding a function boxed twice is the same")
	}
}

func TestAnyComparesAStructOfOneFunctionByItsClosure(t *testing.T) {
	value := holder{callback: adder(1)}
	if !SameAny(any(value), any(value)) {
		t.Fatal("a struct of one function, stored in the interface word, is not itself")
	}
	if SameAny(any(value), any(holder{callback: adder(1)})) {
		t.Fatal("structs of distinct closures are the same")
	}
}
