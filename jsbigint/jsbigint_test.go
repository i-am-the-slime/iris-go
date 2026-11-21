package jsbigint

import (
	"math/big"
	"testing"
)

func TestFromStringRejectsALoneSign(t *testing.T) {
	parse := func(s string) string {
		return FromString(func(n *big.Int) string { return n.String() }, "Nothing", s)
	}
	for input, want := range map[string]string{"": "0", " -12 ": "-12", "+": "Nothing", "-": "Nothing", " + ": "Nothing"} {
		if got := parse(input); got != want {
			t.Errorf("fromString %q = %s, want %s", input, got, want)
		}
	}
}
