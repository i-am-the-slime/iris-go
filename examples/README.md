# Examples

Small PureScript programs that Iris's Go backend compiles and runs, comparing the
output with the JavaScript backend (`compiler-backend/go/scripts/run-examples.sh`
in purescript-iris).

Each directory is a self-contained spago project. `examples.txt` lists them, one per
line: `name|entry|kind|expected`, where `kind` is `value` (the entry's shown value),
`effect` (an `Effect` run for its output) or `program` (a `main`), and `expected` is
the output it must print. A `go-ffi/` directory holds the Go implementations of the
example's foreign imports.
