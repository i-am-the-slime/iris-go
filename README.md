# iris-go

Runtime support for Go programs compiled from PureScript by
[Iris's Go backend](https://github.com/i-am-the-slime/iris-go-backend).
Iris monomorphises a whole PureScript program into typed Go that imports this module,
`github.com/i-am-the-slime/iris-go`: strings with JavaScript's UTF-16 semantics, numbers,
arrays, records, Effect, Aff, Console, Regex, Argonaut and the other foreign implementations.
`examples/` holds the programs Iris's `compiler-backend/go/scripts/run-examples.sh` compiles and
compares against the JavaScript backend.

## Compiling a PureScript program to Go

You need Go 1.27 or later and Spago:

```bash
brew install go spago i-am-the-slime/tap/iris

cd my-spago-project
iris run --go -- some arguments
```

That compiles your program to Go, builds it against this module and runs it with the arguments
after `--`. Rerun an unchanged program and it starts right away. `iris go` builds the same
executable and prints its path. Set `IRIS_GO=/path/to/iris-go` to build against a checkout of
this repository instead of the release.

`--entry Main.main` picks the module and value to run. It can be an `Effect Unit`, or any
value, which the program prints. Programs print exactly what the JavaScript build prints, except for the
differences listed in Iris's
[`documentation/go-output.md`](https://github.com/i-am-the-slime/iris-go-backend/blob/main/documentation/go-output.md).
That document also lists the programs the Go backend rejects.

## Foreign imports

Got a `src/Foo.js`? Put a `src/Foo.go` next to it. Libraries you install work the same way, and
`ffi/` has Go versions of a few popular libraries that don't ship their own yet.

[Foreign imports](https://github.com/i-am-the-slime/iris-go-backend/blob/main/documentation/go-output.md#foreign-imports)
shows an example and how each PureScript type looks in Go.

## Testing the runtime

```bash
go test ./...
```
