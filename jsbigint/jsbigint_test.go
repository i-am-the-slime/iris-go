package jsbigint

import (
	"math/big"
	"testing"
)

func TestFromStringRejectsALoneSign(t *testing.T) {
	parse := func(s string) string {
		return maybe(FromString(s))
	}
	for input, want := range map[string]string{"": "0", " -12 ": "-12", "+": "Nothing", "-": "Nothing", " + ": "Nothing"} {
		if got := parse(input); got != want {
			t.Errorf("fromString %q = %s, want %s", input, got, want)
		}
	}
}

func maybe(n *big.Int, found bool) string {
	if !found {
		return "Nothing"
	}
	return n.String()
}
