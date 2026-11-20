// Package purescript provides runtime support for PureScript compiled to Go.
//
// This runtime is intentionally minimal. As the optimizer improves,
// less code will depend on these helpers.
package purescript

import (
	"reflect"
	"sync"
)

// Any is the universal type for PureScript values.
// As we add more inlining, fewer values will use this type.
type Any = interface{}

// Fn is a curried function type.
type Fn = func(Any) Any

// EffFn is an effect thunk type.
type EffFn = func() Any

// Dict is a PureScript record type.
type Dict = map[string]Any

// Once is used for lazy initialization of top-level values.
type Once = sync.Once

// Uncurried function types for direct calls
type Fn1 = func(Any) Any
type Fn2 = func(Any, Any) Any
type Fn3 = func(Any, Any, Any) Any
type Fn4 = func(Any, Any, Any, Any) Any
type Fn5 = func(Any, Any, Any, Any, Any) Any
type Fn6 = func(Any, Any, Any, Any, Any, Any) Any
type Fn7 = func(Any, Any, Any, Any, Any, Any, Any) Any
type Fn8 = func(Any, Any, Any, Any, Any, Any, Any, Any) Any
type Fn9 = func(Any, Any, Any, Any, Any, Any, Any, Any, Any) Any
type Fn10 = func(Any, Any, Any, Any, Any, Any, Any, Any, Any, Any) Any

// Undefined represents PureScript's undefined value.
const Undefined = "\x00<undefined>\x00"

// Run executes an effect thunk.
// Optimized effect code may produce values directly, so we handle both cases.
func Run(f Any) Any {
	if fn, ok := f.(EffFn); ok {
		return fn()
	}
	// A monomorphized effect is a typed thunk `func() T` or, for Effect Unit,
	// a zero-result `func()`. Invoke either niladic shape reflectively.
	if fv := reflect.ValueOf(f); fv.Kind() == reflect.Func {
		ft := fv.Type()
		if ft.NumIn() == 0 {
			switch ft.NumOut() {
			case 0:
				fv.Call(nil)
				return struct{}{}
			case 1:
				return fv.Call(nil)[0].Interface()
			}
		}
	}
	// Already a value, not a thunk - return as-is
	return f
}

// PsTagged is satisfied by every generated data-type struct (MAR-126): it
// reports the value's active constructor name. Because Go interface
// satisfaction is structural, a boxed constructor can be tag-tested through
// this interface without the test site knowing its concrete Go type.
type PsTagged interface{ PsTag() string }

// IsTag checks if a value is a specific constructor.
// Used for pattern matching on sum types.
func IsTag(v Any, tag string) bool {
	t, ok := v.(PsTagged)
	return ok && t.PsTag() == tag
}

// GetTag returns the constructor tag of a value.
func GetTag(v Any) string {
	if t, ok := v.(PsTagged); ok {
		return t.PsTag()
	}
	return ""
}

// Get retrieves a field from a record.
func Get(dict Dict, key string) Any {
	value, ok := dict[key]
	if !ok {
		panic("Field '" + key + "' not found in record")
	}
	return value
}

// Index retrieves an element from an array.
func Index(arr Any, idx Any) Any {
	a := arr.([]Any)
	i := idx.(int)
	return a[i]
}

// Length returns the length of an array.
func Length(arr Any) Any {
	a := arr.([]Any)
	return len(a)
}

// CopyDict creates a shallow copy of a record.
func CopyDict(dict Any) Any {
	d := dict.(Dict)
	cpy := make(Dict)
	for k, v := range d {
		cpy[k] = v
	}
	return cpy
}

// Fail panics with the given message.
// Used for partial functions and failed pattern matches.
func Fail(msg string) Any {
	panic(msg)
}

//-----------------------------------------------------------------------------
// Uncurried call helpers
//-----------------------------------------------------------------------------

// Call0 calls an any-typed thunk (zero arguments)
func Call0(f Any) Any {
	return f.(func() Any)()
}

// Call1 calls an any-typed function with 1 argument
func Call1(f Any, a1 Any) Any {
	return f.(func(Any) Any)(a1)
}

// Call2 calls an any-typed function with 2 arguments
func Call2(f Any, a1, a2 Any) Any {
	return f.(func(Any, Any) Any)(a1, a2)
}

// Call3 calls an any-typed function with 3 arguments
func Call3(f Any, a1, a2, a3 Any) Any {
	return f.(func(Any, Any, Any) Any)(a1, a2, a3)
}

// Call4 calls an any-typed function with 4 arguments
func Call4(f Any, a1, a2, a3, a4 Any) Any {
	return f.(func(Any, Any, Any, Any) Any)(a1, a2, a3, a4)
}

// Call5 calls an any-typed function with 5 arguments
func Call5(f Any, a1, a2, a3, a4, a5 Any) Any {
	return f.(func(Any, Any, Any, Any, Any) Any)(a1, a2, a3, a4, a5)
}

// Call6 calls an any-typed function with 6 arguments
func Call6(f Any, a1, a2, a3, a4, a5, a6 Any) Any {
	return f.(func(Any, Any, Any, Any, Any, Any) Any)(a1, a2, a3, a4, a5, a6)
}

// Call7 calls an any-typed function with 7 arguments
func Call7(f Any, a1, a2, a3, a4, a5, a6, a7 Any) Any {
	return f.(func(Any, Any, Any, Any, Any, Any, Any) Any)(a1, a2, a3, a4, a5, a6, a7)
}

// Call8 calls an any-typed function with 8 arguments
func Call8(f Any, a1, a2, a3, a4, a5, a6, a7, a8 Any) Any {
	return f.(func(Any, Any, Any, Any, Any, Any, Any, Any) Any)(a1, a2, a3, a4, a5, a6, a7, a8)
}

// Call9 calls an any-typed function with 9 arguments
func Call9(f Any, a1, a2, a3, a4, a5, a6, a7, a8, a9 Any) Any {
	return f.(func(Any, Any, Any, Any, Any, Any, Any, Any, Any) Any)(a1, a2, a3, a4, a5, a6, a7, a8, a9)
}

// Call10 calls an any-typed function with 10 arguments
func Call10(f Any, a1, a2, a3, a4, a5, a6, a7, a8, a9, a10 Any) Any {
	return f.(func(Any, Any, Any, Any, Any, Any, Any, Any, Any, Any) Any)(a1, a2, a3, a4, a5, a6, a7, a8, a9, a10)
}

//-----------------------------------------------------------------------------
// Constructor helpers
//-----------------------------------------------------------------------------

// Constructors are emitted inline as Go value structs in the shared `Types`
// package (MAR-126); the runtime no longer builds them, so the old map-based
// MkCtor* helpers are gone.

//-----------------------------------------------------------------------------
// FFI support
//-----------------------------------------------------------------------------

var foreign = make(map[string]Dict)

// Foreign returns the FFI dictionary for a module.
func Foreign(moduleName string) Dict {
	moduleDict, found := foreign[moduleName]
	if !found {
		moduleDict = Dict{"module": moduleName}
		foreign[moduleName] = moduleDict
	}
	return moduleDict
}

var exitHooks []func()

// OnExit registers work the program's `main` waits for before it returns,
// such as go-support/aff's event loop (node keeps running until no timer or
// fiber is pending). Hooks run in registration order.
func OnExit(hook func()) {
	exitHooks = append(exitHooks, hook)
}

// AwaitBackground runs the OnExit hooks; generated `main` ends with it.
func AwaitBackground() {
	for _, hook := range exitHooks {
		hook()
	}
}
