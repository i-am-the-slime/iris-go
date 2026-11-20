// Package spec implements the foreign imports that purescript-spec's runner
// needs: Test.Spec.Console, Test.Spec.Runner and Test.Spec.Assertions.String.
package spec

import (
	"os"
	"strings"

	"github.com/i-am-the-slime/iris-go/prim"
)

// Write implements Test.Spec.Console.write, `process.stdout.write(s)`,
// which ignores a failing write.
func Write(text string) {
	_, _ = os.Stdout.WriteString(prim.ToUTF8(text))
}

// Exit implements Test.Spec.Runner.exit, `process.exit(code)`.
func Exit(code int32) {
	os.Exit(int(code))
}

// StartsWith implements Test.Spec.Assertions.String._startsWith: whether
// the string starts with the prefix.
func StartsWith(prefix string, text string) bool {
	return strings.HasPrefix(text, prefix)
}

// EndsWith implements Test.Spec.Assertions.String._endsWith.
func EndsWith(suffix string, text string) bool {
	return strings.HasSuffix(text, suffix)
}
