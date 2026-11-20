# iris-go

Runtime support for Go programs compiled from PureScript by Iris's Go backend
(`compiler-backend/go` on the `iris-go` branch of
[purescript-analyzer](https://github.com/i-am-the-slime/purescript-analyzer/tree/iris-go)).
Iris monomorphises a whole PureScript program into typed Go that imports this module,
`github.com/i-am-the-slime/iris-go`: strings with JavaScript's UTF-16 semantics, numbers,
arrays, records, Effect, Aff, Console, Regex, Argonaut and the other foreign implementations.
`examples/` holds the programs Iris's `scripts/run-examples.sh` compiles and compares against
the JavaScript backend.

## Compiling a PureScript program to Go

You need Go 1.27 or later and a Rust toolchain.

```bash
git clone -b iris-go https://github.com/i-am-the-slime/purescript-analyzer iris
cargo build --release -p iris-cli --manifest-path iris/Cargo.toml   # builds iris/target/release/iris

cd my-spago-project
mkdir -p out && (cd out && go mod init example.com/myapp)
iris go --entry Main.main out
cd out
go get github.com/i-am-the-slime/iris-go@v0.1.0
go build -o myapp . && ./myapp
```

`--entry` names the module and value to run. It can be an `Effect Unit`, or any value, which
the program prints. Programs print exactly what the JavaScript build prints, except for the
differences listed in Iris's
[`documentation/go-output.md`](https://github.com/i-am-the-slime/purescript-analyzer/blob/iris-go/documentation/go-output.md).
That document also lists the programs the Go backend rejects and the conventions for writing
Go implementations of your own foreign imports.

## Testing the runtime

```bash
go test ./...
```
